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

// the load balancer instance that will do the work
// for the given service
type LoadBalancer struct {

	// on which port the Load Balancer is running
	Port string

	// the service that the Load Balancer will work on
	Service *Service

	// the algorithm used
	Strategy Strategy

	// mode
	Mode int
}

func NewLoadBalancer(urls []string, strategy, port, name string) (*LoadBalancer, error) {

	instances := make([]*Instance, 0)

	for _, ur := range urls {

		url, err := url.Parse(ur)
		if err != nil {
			return nil, err
		}

		instances = append(instances, &Instance{
			Proxy: httputil.NewSingleHostReverseProxy(url),
			Url:   url,
			WaitingList: make(chan struct {
				res http.ResponseWriter
				r   *http.Request
			}, 100),
		})

	}

	var loadStrategy Strategy

	switch strategy {
	case "round-robin":
		loadStrategy = &roundrobin.RoundRobin{
			Current: -1,
			Total:   len(instances),
		}
	default:
		return nil, errors.New("")
	}

	return &LoadBalancer{
		Port:     port,
		Strategy: loadStrategy,
		Service: &Service{
			Name:      name,
			Instances: instances,
		},
		Mode: NORMAL_MODE,
	}, nil
}

// checks the instances of a service and removes broken instances
func (l *LoadBalancer) InitialHealthCheck() {

	deleted := 0

	for i, instance := range l.Service.Instances {

		conn, err := net.DialTimeout("tcp", instance.Url.Host, 5*time.Second)

		if err != nil {

			log.Printf("Instance %v with url: %s is not responding and it's removed from the waiting list", i, instance.Url)

			// remove the instance from the service
			l.Service.DeleteInstance(i - deleted)

			// update total
			l.Strategy.UpdateTotal(len(l.Service.Instances))

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

	l.Service.StartService()

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

// TODO: need to change the way we handle requests to use the local channels of every instance
// we will just dump the pair of response writer and request to the channel of the node
// the health check will be done locally in the node
func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, r *http.Request) {

	// moving to the next channel
	next := l.Strategy.Next()

	log.Printf("dumping request in %v", next)

	// dumping the res a r to the instance's local channel
	l.Service.Instances[next].WaitingList <- struct {
		res http.ResponseWriter
		r   *http.Request
	}{
		res: res,
		r:   r,
	}
}
