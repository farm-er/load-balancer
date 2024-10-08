package roundrobin

import "sync"

type RoundRobin struct {
	Total        int
	Current      int
	mutexTotal   sync.RWMutex
	mutexCurrent sync.RWMutex
}

func (r *RoundRobin) Next() int {

	r.mutexCurrent.Lock()

	r.Current = (r.Current + 1) % r.Total

	r.mutexCurrent.Unlock()

	return r.Current
}

func (r *RoundRobin) UpdateTotal(value int) {

	r.mutexTotal.Lock()

	if value == -1 {
		r.Total -= 1
	} else {
		r.Total = value
	}

	r.mutexTotal.Unlock()

	if r.Current >= r.Total {
		r.mutexCurrent.Lock()

		r.Current = 0

		r.mutexCurrent.Unlock()
	}

}

func (r *RoundRobin) GetTotal() int {
	return r.Total
}

func (r *RoundRobin) GetCurrent() int {
	return r.Current
}
