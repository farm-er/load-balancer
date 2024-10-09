package loadbalancer

import (
	"errors"
	"log"
	"net"
	"net/http/httputil"
	"net/url"
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

	// here are the waiting requests that this instance will serve 
	WaitingList chan WaitingRequest
}


// DONE: we need to add a loop that serves the requests in the waiting list
// and this function will be local to every instance
// maybe send a copy to a local databse for metrics
// DONE: add health check
// when the node fails we return an error
func (i *Instance) redirect() ( *WaitingRequest, error) {

	for r := range i.WaitingList {

		log.Printf("proccessing request")
		
		// Health check performed before every request 
		if !i.healthCheck() {
			// we will return the error and the request will get treated
			// i need also to return the request that was not successfu
			return &r ,errors.New("instance failure")
		}

		i.Proxy.ServeHTTP(r.res, r.r)

		r.finished <- true

	}

	return nil, nil
}

// DONE: localizing health checks to instance level
// need this health check before every request
// returns true if the node is healthy and false if not
func (i *Instance) healthCheck() bool {

	conn, err := net.DialTimeout("tcp", i.Url.Host, 2*time.Second)

	if err != nil {
		log.Printf("Instance with url: %s is not responding", i.Url)
		return false
	}

	conn.Close()
	return true
}

