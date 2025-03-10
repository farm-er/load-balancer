package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	loadbalancer "github.com/farm-er/load-balancer/load_balancer"
)

var ()

func main() {

	var port *string = flag.String("port", "1234", "port for the load balancer")
	var name *string = flag.String("name", "load_balancer", "name of the service")
	var strategy *string = flag.String("strategy", "round-robin", "the algorithm to use default: round-robin")
	var urlsFileName *string = flag.String("file", "", "file containing urls pointing to your servers")

	flag.StringVar(port, "p", *port, "alias for --port")
	flag.StringVar(name, "n", *name, "alias for --name")
	flag.StringVar(strategy, "s", *strategy, "alias for --strategy")
	flag.StringVar(urlsFileName, "f", *urlsFileName, "alias for --file")

	flag.Parse()

	urlsFile, err := os.Open(*urlsFileName)

	if err != nil {
		log.Fatal(err)
	}

	urlsScanner := bufio.NewScanner(urlsFile)
	urls := []string{}

	for urlsScanner.Scan() {
		url := urlsScanner.Text()
		// TODO: add check to the urls
		fmt.Println(url)
		urls = append(urls, url)
	}

	urlsFile.Close()

	loadBalancer, err := loadbalancer.NewLoadBalancer(urls, *strategy, *port, *name)

	if err != nil {
		log.Fatal(err)
	}

	loadBalancer.Start()
}
