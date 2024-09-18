package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

var (
	port     = flag.Int("port", 8080, "port for the load balancer")
	name     = flag.String("name", "", "name of the service")
	strategy = flag.String("strategy", "round-robin", "the algorithm used")
)

// is an instance of the service
// and it's building bloc of the service
// multiple of these build a service that needs load balancing
type Instance struct {
	Url   string
	Proxy *httputil.ReverseProxy
}

// the service that needs load balancing
// contains multiple instances or servers
type Service struct {
	Name      string
	Instances []*Instance
}

type Strategy interface {
	next() int
}

// the load balancer instance that will do the work
// for the given service
type LoadBalancer struct {
	Port         int
	Service      *Service
	Strategy     *Strategy
	Total        int
	mutexTotal   sync.RWMutex
	Current      int
	mutexCurrent sync.RWMutex
}

func NewLoadBalancer(urls []string) (*LoadBalancer, error) {

	instances := make([]*Instance, 0)

	for _, ur := range urls {

		url, err := url.Parse(ur)
		if err != nil {
			return nil, err
		}

		instances = append(instances, &Instance{
			Proxy: httputil.NewSingleHostReverseProxy(url),
			Url:   ur,
		})

	}

	return &LoadBalancer{
		Port: *port,
		Service: &Service{
			Name:      *name,
			Instances: instances,
		},
	}, nil
}

func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, r *http.Request) {

}

func main() {
	flag.Parse()

	urls := []string{"url1", "url2"}

	loadBalancer, err := NewLoadBalancer(urls)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(loadBalancer)

	mainServer := http.Server{
		Addr:    "",
		Handler: loadBalancer,
	}

	if err = mainServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
