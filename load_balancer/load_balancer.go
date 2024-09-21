package loadbalancer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	roundrobin "github.com/farm-er/load-balancer/round_robin"
)

type Strategy interface {
	UpdateTotal(value int)
	Next() int
}

// the load balancer instance that will do the work
// for the given service
type LoadBalancer struct {

	// on which port the Load Balancer is running
	Port string

	// the service that the Load Balancer will work on
	Service *Service

	// the algorithm used
	Strategy Strategy
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
	}, nil
}

// checks the instances of a service and removes broken instances
func (l *LoadBalancer) InitialHealthCheck() {

	for i, instance := range l.Service.Instances {

		conn, err := net.DialTimeout("tcp", instance.Url.Host, 5*time.Second)

		if err != nil {

			log.Printf("Instance %v with url: %s is not responding and it's removed from the waiting list", i, instance.Url)

			// remove the instance from the service
			l.Service.DeleteInstance(i)

			// update total
			l.Strategy.UpdateTotal(len(l.Service.Instances))

			continue
		}

		conn.Close()
	}
}

func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, r *http.Request) {

	// do a health before using the service

	healthy := false

	var next int

	for !healthy {

		next = l.Strategy.Next()

		healthy = l.Service.healthCheckInstance(next)

		if !healthy {
			l.Strategy.UpdateTotal(-1)
		}

	}

	l.Service.Instances[next].Proxy.ServeHTTP(res, r)
}

func (l *LoadBalancer) Start() {

	// TODO: add health check before starting the load balancer

	mainServer := http.Server{
		Addr:    ":" + l.Port,
		Handler: l,
	}

	fmt.Println("Load balancer running on port " + l.Port)

	fmt.Println(l.Service.Instances[0].Url)

	if err := mainServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
