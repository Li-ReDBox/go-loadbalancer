// go run pool.go load_balancer.go
// package loadbalancer
package main

import (
	"container/heap"
	"fmt"
	"sync/atomic"
	"time"
)

type Request struct {
	fn func() int // The operation to perform: anything takes no arguments and returns an int
	c  chan int   // The channel to return the result.
}

type Worker struct {
	request chan Request // work to do (buffered channel)
	pending int          // count of pending tasks, it decides the order of Worker in they queue
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // index in the heap
}

func (w *Worker) work(done chan *Worker) {
	for req := range w.request {
		// fmt.Println("Getting a request from pool for requests")
		// req := <-w.request // get a Request from the pool in balancer
		// fmt.Println("The worker with least load has been received. Run the request and pass on the result to request.")
		// send result to requester by the channel defined in Request
		req.c <- req.fn() // call fn and send result
		// fmt.Println("Worker has sent result to Request's channel. Next, tell balancer it is done.")
		done <- w // we've finished this request, notify the pool in balancer
		// fmt.Println("Balancer has been notified from a worker.")
	}
}

// The load balancer needs a pool of workers and a single channel to which requesters can report task completion.
type Balancer struct {
	pool Pool
	done chan *Worker
}

func (b *Balancer) print() {
	for i := 0; i < len(b.pool); i++ {
		fmt.Printf("\tworker %d: pending = %d\n", i, b.pool[i].pending)
	}
	fmt.Println()
}

// balance runs load balancing strategy and update the state of the worker pool.
// The balancer waits for new messages on the request and completion channels.
func (b *Balancer) balance(req chan Request) {
	var n atomic.Uint32
	for {
		select {
		case req := <-req: // received a Request...
			n.Add(1)
			fmt.Println("Balancer received request. Start to dispatch ...")
			b.dispatch(req) // ...so send it to a Worker
			b.print()
		case w := <-b.done: // a worker has finished ...
			fmt.Printf("Balancer received the signal of Done.\n\t So far the completed job count: %d\n\n", n.Load())
			b.completed(w) // ...so update its info
		case <-time.After(10 * time.Second):
			return
		}
	}
}

// Send Request to worker
func (b *Balancer) dispatch(req Request) {
	// Grab the least loaded worker...
	w := heap.Pop(&b.pool).(*Worker)
	// ...send it the task.
	w.request <- req
	// One more in its work queue.
	w.pending++
	// Put it into its place on the heap.
	heap.Push(&b.pool, w)
}

// Job is complete; update heap
func (b *Balancer) completed(w *Worker) {
	// One fewer in the queue.
	w.pending--
	// Remove it from heap.
	heap.Remove(&b.pool, w.index)
	// Put it into its place on the heap.
	heap.Push(&b.pool, w)
	// fmt.Printf("Cleanup done, and push the worker %d back to the pool for new requests.\n\n", w.index)

	// this may be replaced by update / heap.Fix
}
