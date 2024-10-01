package loadbalancer

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// is an instance of the service
// and it's building bloc of the service
// multiple of these build a service that needs load balancing
// DONE: will change how requests are served
// i'm going to add a channel for every instance and the load balancer will work with the channels
type Instance struct {
	Url *url.URL

	Proxy *httputil.ReverseProxy

	WaitingList chan struct {
		res http.ResponseWriter
		r   *http.Request
	}
}

// DONE: we need to add a loop that serves the requests in the waiting list
// and this function will be local to every instance
// maybe send a copy to a local databse for metrics
// DONE: add health check
// when the node fails we return an error
func (i *Instance) serveHTTP() error {

	for r := range i.WaitingList {

		if !i.healthCheck() {
			return errors.New("instance failure")
		}

		i.Proxy.ServeHTTP(r.res, r.r)

	}

	return nil
}

// DONE: localizing health checks to instance level
// need this health check before every request
// returns true if the node is healthy and false if not
func (i *Instance) healthCheck() bool {

	conn, err := net.DialTimeout("tcp", i.Url.Host, 2*time.Second)

	if err != nil {
		log.Printf("Instance with url: %s is not responding and it's removed from the waiting list", i.Url)
		return false
	}

	conn.Close()
	return true
}

// the service that needs load balancing
// contains multiple instances or servers
// TODO: add a port like field where you wait for signals to stop an instance
type Service struct {
	Name           string
	Instances      []*Instance
	InstancesMutex sync.RWMutex
}

// deletes an instance from the service
// it locks before wrting so it's safe to use asynchronously
// TODO: move the requests on this node to other nodes
func (s *Service) DeleteInstance(index int) {

	if index < 0 || index >= len(s.Instances) {
		log.Fatal("Error in instances management")
	}

	s.InstancesMutex.Lock()

	if len(s.Instances) == 1 {
		s.Instances = []*Instance{}
	} else if len(s.Instances)-1 == index {
		s.Instances = s.Instances[:index]
	} else {
		s.Instances = append(s.Instances[:index], s.Instances[index+1:]...)
	}

	s.InstancesMutex.Unlock()
}

// DONE: centralize the serving of requests in the service
// so we can manage failures
func (s *Service) StartService() {

	for index, instance := range s.Instances {

		go func(i *Instance) {
			log.Printf("instance %v is running", index)
			if err := i.serveHTTP(); err != nil {
				// handle instance failure
				// TODO: we need to print the problem and first recheck if the instance is really out
				// then proceed to delete it from the list and copy the request over to serve them in the other nodes

				log.Printf("Instance failure with url: %s", i.Url)
			}
		}(instance)

	}

}
