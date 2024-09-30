package loadbalancer

import (
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
// TODO: i need to pass a delete instance function from the service.
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
// TODO: delete the instance
// TODO: manage the requests in the queue
func (i *Instance) serveHTTP() {
	for r := range i.WaitingList {
		i.healthCheck()
		i.Proxy.ServeHTTP(r.res, r.r)
	}
}

// DONE: localizing health checks to instance level
// need this health check before every request
func (i *Instance) healthCheck() {

	conn, err := net.DialTimeout("tcp", i.Url.Host, 2*time.Second)

	if err != nil {
		log.Printf("Instance with url: %s is not responding and it's removed from the waiting list", i.Url)
		return
	}

	conn.Close()
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
