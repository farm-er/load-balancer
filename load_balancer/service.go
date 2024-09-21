package loadbalancer

import (
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// is an instance of the service
// and it's building bloc of the service
// multiple of these build a service that needs load balancing
type Instance struct {
	Url   *url.URL
	Proxy *httputil.ReverseProxy
}

// the service that needs load balancing
// contains multiple instances or servers
type Service struct {
	Name           string
	Instances      []*Instance
	InstancesMutex sync.RWMutex
}

// deletes an instance from the service
// it locks before wrting so it's safe to use asynchronously
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

// checks if an instance is alive
// it returns true if it's alive and false if it's not
func (s *Service) healthCheckInstance(index int) bool {

	conn, err := net.DialTimeout("tcp", s.Instances[index].Url.Host, 2*time.Second)

	if err != nil {

		log.Printf("Instance %v with url: %s is not responding and it's removed from the waiting list", index, s.Instances[index].Url)

		// delete the instance from the list
		s.DeleteInstance(index)

		return false
	}

	conn.Close()

	return true
}
