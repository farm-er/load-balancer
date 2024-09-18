package roundrobin

import (
	"net/http"
	"sync"
)

type server struct {
	port string
}

// a load balancer that implements Round Robin algorithm needs:
// -- a pool of servers that contains the servers we currently have
// -- an index to the last server used or the last server to which we redirected the request
// -- mutex to lock the critical data for concurrency
type roundRobinBalancer struct {
	servers     []server
	last        int
	serversLock sync.RWMutex
	lastMutex   sync.RWMutex
}

func (l *roundRobinBalancer) addServer(server server) {

	l.serversLock.Lock()

	l.servers = append(l.servers, server)

	l.serversLock.Unlock()

}

func (l *roundRobinBalancer) healthCheck() {}

func (l *roundRobinBalancer) serveNext(r *http.Request) {

}
