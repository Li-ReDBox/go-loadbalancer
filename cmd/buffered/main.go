package main

import (
	"math/rand"
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
	// each requester only allow to 10 requests to save time
	for i := 0; i < 3; i++ {
		// Kill some time (fake load). Do not flat out.
		// time.Sleep(time.Duration(rand.Int63n(1e3 / int64(nWorker))))
		work <- lb.Request{Fn: workFn, Result: c} // send request, blocks
		<-c                                       // the result of workFn only returns boring 1, so discard by just draining the channel
		// below steps can be potentially added
		// result := <-c              // wait for answer until there is one
		// fmt.Println("Request has been processed, will send to furtherProcess()")
		// furtherProcess(result)
		// fmt.Println("furtherProcess has finished too. Full life cycle of a request is done.")
	}
}

func main() {
	nRequester := 5 // this is the maximal pending total: each requester will wait until last request has completed before a new request is sent
	nWorker := 3
	wp := make(lb.Pool, nWorker)

	// when a non-buffered channel is used here, workers can hold up to 4 requests, then goes into deadlock
	for i := 0; i < nWorker; i++ {
		w := lb.NewWorker(make(chan lb.Request))
		wp[i] = &w
	}

	comp := make(chan *lb.Worker, 1)
	// set all workers with the same completion notification channel ready for receiving Requests
	for _, w := range wp {
		go w.Work(comp)
	}

	// Make a request channel for requester to send requests and run a few goroutines to generate requests
	r := make(chan lb.Request)

	// Set the Balancer up by passing on request and notification channels
	b := lb.Balancer{}
	// Balance has a timeout of 10s clause to exit
	go b.Balance(wp, r, comp)

	for i := 0; i < nRequester; i++ {
		requester(r, nWorker)
	}

}
