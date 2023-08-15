package main

import (
	"fmt"
	"time"

	lb "funmech.com/loadbalancer"
)

var seed = 0

func fn() int {
	seed++
	return seed
}

// A simple requester and worker communicate through a single Request channel and a single response channel.
// One worker, nRequest requester
func main() {
	nRequest := 10
	r := make(chan lb.Request)

	// The order is critical: before sending, channel has to be ready to receive
	// go func(r chan lb.Request) {
	// 	c := make(chan int)
	// 	for i := 0; i < nRequest; i++ {
	// 		r <- lb.Request{Fn: fn, Result: c}
	// 	}
	// }(r)

	// for req := range r {
	// 	req.Result <- req.Fn()
	// }

	go func(r chan lb.Request) {
		for req := range r {
			req.Result <- req.Fn()
		}
		fmt.Println("All works are done.")
	}(r)

	c := make(chan int)
	for i := 0; i < nRequest; i++ {
		go func() {
			r <- lb.Request{Fn: fn, Result: c}
		}()
	}

	// Retrieve results of all requests sent
	for i := 0; i < nRequest; i++ {
		fmt.Println("Run", i, "has result of", <-c)
	}

	// close the channel to allow the goroutine and wait for it to exit, this is more important there are calls to external
	close(r)
	// wait a bit to allow the print to happen
	// This is not a proper way for waiting because this maybe machine related.
	time.Sleep(100 * time.Microsecond)
}
