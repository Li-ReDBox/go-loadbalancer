package main

import (
	"container/heap"
	"math/rand"
	"time"
)

func workFn() int {
	time.Sleep(time.Duration(rand.Intn(9000)) * time.Millisecond)
	// fmt.Println("I did something for the balancer.")
	return 1
}

// func furtherProcess(c int) {
// 	fmt.Println("Demonstrate a further work in a pipeline: printing:", c)
// }

// An artificial but illustrative simulation of a requester, a load generator.
// work is a send-only channel, once set, Balancer can start to dispatch
func requester(work chan<- Request) {
	c := make(chan int) // create a channel for receiving result for a particular requester
	// each requester only allow to 10 requests
	for i := 0; i < 10; i++ {
		// Kill some time (fake load). Do not flat out.
		// time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		work <- Request{workFn, c} // send request, blocks
		<-c
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

	workers := 3
	wp := make(Pool, workers)

	for i := 0; i < workers; i++ {
		wp[i] = &Worker{
			request: make(chan Request, 10), // this is a buffered channel
		}
	}

	heap.Init(&wp)

	b := Balancer{
		wp,
		make(chan *Worker, workers),
	}

	// set all workers to share the same balancer channel
	for _, w := range wp {
		go w.work(b.done)
	}

	// Balancer has only one request channel
	r := make(chan Request)
	for i := 0; i < 3; i++ {
		go requester(r)
	}

	// set up channel, it has to be done through goroutine
	b.balance(r)

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
