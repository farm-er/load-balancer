package main

import (
	"flag"
	"log"

	loadbalancer "github.com/farm-er/load-balancer/load_balancer"
)

var (
	port     = flag.String("port", "1234", "port for the load balancer")
	name     = flag.String("name", "", "name of the service")
	strategy = flag.String("strategy", "round-robin", "the algorithm to use default: round-robin")
)

func main() {
	flag.Parse()

	urls := []string{"http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082"}

	loadBalancer, err := loadbalancer.NewLoadBalancer(urls, *strategy, *port, *name)

	if err != nil {
		log.Fatal(err)
	}

	loadBalancer.Start()
}
