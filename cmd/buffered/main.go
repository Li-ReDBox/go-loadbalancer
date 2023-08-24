package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	lb "funmech.com/loadbalancer"
)

var count atomic.Int32

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
	count.Add(1)
	fmt.Println("Done generating requests")
}

func main() {
	nRequester := 8 // this is the maximal pending total: each requester will wait until last request has completed before a new request is sent
	nWorker := 3
	wp := make(lb.Pool, nWorker)

	// Request channel of each Worker is set to the number of requesters or wReqSize like below
	wReqSize := 3 // roundUp(nRequester / nWorker) ==> 8 /3 = 3
	for i := 0; i < nWorker; i++ {
		w := lb.NewWorker(make(chan lb.Request, wReqSize))
		wp[i] = &w
	}

	var wg sync.WaitGroup

	// comp := make(chan *lb.Worker, nWorker)
	comp := make(chan *lb.Worker)
	// set all workers with the same completion notification channel
	for _, w := range wp {
		wg.Add(1)
		go func(w *lb.Worker) {
			defer wg.Done()
			w.Work(comp)
		}(w)
	}

	// Make a request channel for requester to send requests
	r := make(chan lb.Request)

	// Set the Balancer up by passing on request and notification channels
	b := lb.Balancer{}
	// Balance has a timeout of 10s clause to exit when its dispatch is not in deadlock!
	go b.Balance(wp, r, comp)

	start := time.Now()

	// run a few goroutines to generate requests
	for i := 0; i < nRequester; i++ {
		go requester(r, nWorker)
	}

	// Have all requester completed in this demo?
	for count.Load() < int32(nRequester) {
	}
	fmt.Println("Closing the request channel")
	close(r)
	// Wait for all workers have been shutdown
	wg.Wait()

	fmt.Printf("All done in %v\n", time.Since(start))
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
}
