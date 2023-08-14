package main

import (
	"math/rand"
	"time"

	"funmech.com/loadbalancer"
)

func workFn() int {
	time.Sleep(time.Duration(rand.Intn(9000)) * time.Millisecond)
	// fmt.Println("I did something for the balancer.")
	return 1
}

// An artificial but illustrative simulation of a requester, a load generator.
// work is a send-only channel, once set, Balancer can start to dispatch
func requester(work chan<- loadbalancer.Request, nWorker int) {
	c := make(chan int) // create a channel for receiving result for a particular requester
	// each requester only allow to 10 requests
	for i := 0; i < 10; i++ {
		// Kill some time (fake load). Do not flat out.
		time.Sleep(time.Duration(rand.Int63n(1e3 * int64(nWorker))))
		work <- loadbalancer.Request{Fn: workFn, Result: c} // send request, blocks
		<-c                                                 // the result of workFn only returns boring 1, so discard by just draining the channel
		// result := <-c              // wait for answer until there is one
		// fmt.Println("Request has been processed, will send to furtherProcess()")
		// furtherProcess(result)
		// fmt.Println("furtherProcess has finished too. Full life cycle of a request is done.")
	}
}

func main() {
	// // Just a demonstration if there is no Worker and Balancer, how a Request
	// // which was generated from a load generator is processed.
	// r := make(chan Request)
	// go requester(r)
	// req := <-r
	// req.c <- req.fn()
	// // there is no wait for furtherProcess, so sleep a bit to let furtherProcess to finish
	// time.Sleep(1 * time.Second)
	// End of the simple dome

	nRequester := 8
	nWorker := 3
	wp := make(loadbalancer.Pool, nWorker)

	for i := 0; i < nWorker; i++ {
		w := loadbalancer.NewWorker(make(chan loadbalancer.Request, nRequester))
		wp[i] = &w
	}

	comp := make(chan *loadbalancer.Worker, nWorker)

	// set all workers to share the same completion notification channel
	for _, w := range wp {
		go w.Work(comp)
	}

	// Make a request channel for requester to send requests
	r := make(chan loadbalancer.Request)
	for i := 0; i < nRequester; i++ {
		go requester(r, nWorker)
	}

	// Set the Balancer up
	b := loadbalancer.Balancer{}
	b.Balance(wp, r, comp)

	// // Below creates requests non-stop, very quick resource runs out
	// reaching to LLVM limit of 8128 live goroutines:
	// race: limit on 8128 simultaneously alive goroutines is exceeded, dying
	// go func() {
	// 	for {
	// 		go requester(r)
	// 	}
	// }()
	// boom := time.After(10 * time.Second)
	// <-boom
	// fmt.Println("Too much, going home. BOOM!")
}
