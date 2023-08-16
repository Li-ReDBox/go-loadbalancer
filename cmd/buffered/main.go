package main

import (
	"math/rand"
	"sync"
	"time"

	lb "funmech.com/loadbalancer"
)

func workFn() int {
	time.Sleep(time.Duration(rand.Intn(9)) * time.Second)
	return 1
}

// An artificial but illustrative simulation of a requester, a load generator.
// work is a send-only channel, once set, Balancer can start to dispatch
func requester(work chan<- lb.Request, nWorker int) {
	c := make(chan int) // create a channel for receiving result for a particular requester
	// in this design, the number of requests a requester run is not critical.
	for i := 0; i < 3; i++ {
		work <- lb.Request{Fn: workFn, Result: c} // send request, blocks
		<-c                                       // the result of workFn only returns boring 1, so discard by just draining the channel
	}
}

func main() {
	nRequester := 5 // this is the maximal pending total: each requester will wait until last request has completed before a new request is sent
	nWorker := 3
	wp := make(lb.Pool, nWorker)

	// Request channel of each Worker is set to the number of requesters
	for i := 0; i < nWorker; i++ {
		w := lb.NewWorker(make(chan lb.Request, nRequester))
		wp[i] = &w
	}

	comp := make(chan *lb.Worker, nWorker)
	// set all workers with the same completion notification channel
	for _, w := range wp {
		go w.Work(comp)
	}

	// Make a request channel for requester to send requests
	r := make(chan lb.Request)

	// Set the Balancer up by passing on request and notification channels
	b := lb.Balancer{}
	// Balance has a timeout of 10s clause to exit when its dispatch is not in deadlock!
	go b.Balance(wp, r, comp)

	// run a few goroutines to generate requests
	var wg sync.WaitGroup
	for i := 0; i < nRequester; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			requester(r, nWorker)
		}()
	}
	// Wait for all requests have been completed
	wg.Wait()
}
