package loadbalancer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	roundrobin "github.com/farm-er/load-balancer/round_robin"
)

type Strategy interface {
	GetTotal() int
	GetCurrent() int
	UpdateTotal(value int)
	Next() int
}

const (

	// normal mode is for normal load balancer work
	// when there's a minimum of one instance
	// while also trying to recover the other instances
	NORMAL_MODE = iota

	// recovery is used when the load balancer can't find
	// any instance and will just wait the repair of one of the lost ones
	// if the recovery array is empty the Load Balancer will stop
	RECOVERY_MODE
)


// DONE: Create a new type for the combination of request, responseWriter, and finished channel

// This type represent a request in waiting list of instances
// and also in the load balancer's emergency list 
type WaitingRequest struct {

	// Combination of request and responseWriter that we got from the handlers
	res http.ResponseWriter

	r *http.Request

	// a channel to signal to the handler that request processing is finished 
	finished chan bool

}

// the load balancer instance that will do the work
// for the given service
type LoadBalancer struct {

	// on which port the Load Balancer is running
	Port string


	// the algorithm used
	Strategy Strategy

	// mode
	Mode int

	// here the Load Balancer will requests from dead instances
	EmergencyChan chan WaitingRequest 


	// User's servers
	Instances []Instance

	// mutex to write to Instances
	InstancesMutex sync.RWMutex

}

func NewLoadBalancer(urls []string, strategy, port, name string) (*LoadBalancer, error) {

	instances := make([]Instance, 0)

	for _, ur := range urls {

		url, err := url.Parse(ur)
		if err != nil {
			return nil, err
		}

		instances = append(instances, Instance{
			// Proxy to redirect traffic
			Proxy: httputil.NewSingleHostReverseProxy(url),
			// server's URL 
			Url:   url,
			// The instance will listen to this channel for any requests 
			WaitingList: make( chan WaitingRequest, 100),
		})

	}

	var loadStrategy Strategy

	// Will use the strategy based on user's choice
	// All strategies are under the Strategy interface 
	switch strategy {
	case "round-robin":
		loadStrategy = &roundrobin.RoundRobin{
			Current: -1,
			Total:   len(instances),
		}
	default:
		return nil, errors.New("")
	}

	// return the load balancer instance ready to start 
	return &LoadBalancer{
		Port:     port,
		Strategy: loadStrategy,	
		Mode: NORMAL_MODE,
		EmergencyChan: make(chan WaitingRequest, 100),
		Instances: instances,
	}, nil
}

// checks the instances of a service and removes broken instances
func (l *LoadBalancer) InitialHealthCheck() {

	deleted := 0

	for i, instance := range l.Instances {

		conn, err := net.DialTimeout("tcp", instance.Url.Host, 5*time.Second)

		if err != nil {

			log.Printf("Instance %v with url: %s is not responding and it's removed from the waiting list", i, instance.Url)

			// remove the instance from the service
			l.deleteInstance(i - deleted, nil)

			// update total
			l.Strategy.UpdateTotal(len(l.Instances))

			deleted++

			continue
		}

		conn.Close()
	}

	// DONE: if number of instances is zero the Load Balancer will stop

	if l.Strategy.GetTotal() == 0 {
		log.Printf("No instance is responding please check your servers")
		os.Exit(0)
	}

}

func (l *LoadBalancer) Start() {

	// DONE: add health check before starting the load balancer

	l.InitialHealthCheck()

	mainServer := http.Server{
		Addr:    ":" + l.Port,
		Handler: l,
	}

	log.Printf("Starting the service")


	for index, instance := range l.Instances {

		go func(i Instance) {
			log.Printf("instance %v is running", index)
			if req, err := i.redirect(); err != nil {
				// handle instance failure
				// DONE: copy the request over to serve them in the other nodes

				l.deleteInstance( index, req)

				if len(l.Instances) == 0 {
					// DONE: if no instance is alive we're going to change mode to recovery
					l.Mode = RECOVERY_MODE
				}
			}
		}(instance)

	}


	// Start the emergency channel worker here
	go l.relieveEmergencyChannel()

	fmt.Println("Load balancer running on port " + l.Port)

	if err := mainServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func (l *LoadBalancer) SwitchToRecovery() {
	l.Mode = RECOVERY_MODE
	fmt.Printf("there are no instances available the service will switch to recovery mode")

	fmt.Printf("switched to RECOVERY MODE")
}

// DONE: need to change the way we handle requests to use the local channels of every instance
// we will just dump the pair of response writer and request to the channel of the node
// the health check will be done locally in the node
func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, r *http.Request) {

	// moving to the next channel
	next := l.Strategy.Next()

	finished := make(chan bool, 1)

	// dumping the res a r to the instance's local channel
	l.Instances[next].WaitingList <- struct {
		res http.ResponseWriter
		r   *http.Request
		finished chan bool
	}{
		res: res,
		r:   r,
		finished: finished,
	}

	// DONE: create a channel to receive to finishing of request handling

	select {
	case <- finished:
	return
	}

}


func (l *LoadBalancer) deleteInstance( index int, req *WaitingRequest) {

	log.Printf("Removing the instance")

	if index < 0 || index >= len(l.Instances) {
		log.Fatal("Error in instances management")
	}

	// close the waiting list for the dead instance 
	close(l.Instances[index].WaitingList)

	// starting with the uncompleted request first
	if req != nil {
		l.EmergencyChan <- *req
	}
	
	// emptying the waiting list in the emergency channel for the load balancer redistribute
	for item := range l.Instances[index].WaitingList {
		l.EmergencyChan <- item 
	}

	// deleting the instance 
	l.InstancesMutex.Lock()

	if len(l.Instances) == 1 {
		l.Instances = []Instance{}
	} else if len(l.Instances)-1 == index {
		l.Instances = l.Instances[:index]
	} else {
		l.Instances = append(l.Instances[:index], l.Instances[index+1:]...)
	}

	l.InstancesMutex.Unlock()
	
	l.Strategy.UpdateTotal(len(l.Instances))

}






// TODO: add a function to take care of the emergency channel 
// this function will take the strategy into 
func (l *LoadBalancer) relieveEmergencyChannel() {

	for req := range l.EmergencyChan {
		next := l.Strategy.Next()
		// We pass the packet to another waiting list 
		l.Instances[next].WaitingList <- struct {
			res http.ResponseWriter
			r   *http.Request
			finished chan bool
		}{
			res: req.res,
			r:   req.r,
			finished: req.finished,
		}
		log.Printf("Request recovered and passed to instance %s", l.Instances[next].Url)
	}

}





