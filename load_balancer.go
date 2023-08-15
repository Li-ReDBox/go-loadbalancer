// go run pool.go load_balancer.go
package loadbalancer

import (
	"container/heap"
	"fmt"
	"time"
)

// Request represents a computation a load balancer support.
// For now, make this a simple data structure, just for passing data
type Request struct {
	Fn     func() int // The operation to perform: anything takes no arguments and returns an int
	Result chan int   // The channel to return the result.
}

// Balancer: a load balancer manages a pool of Workers and a single channel to which Workers can report its completion.
type Balancer struct {
	pool Pool
}

func (b *Balancer) print() {
	for i := 0; i < len(b.pool); i++ {
		fmt.Printf("\tworker %d: pending = %d\n", i, b.pool[i].pending)
	}
	fmt.Println()
}

// Balance runs load balancing strategy and update the state of the worker pool using heap.
// The balancer waits for new messages on the request and completion channels and act accordingly.
func (b *Balancer) Balance(wp Pool, req chan Request, complete chan *Worker) {
	heap.Init(&wp)
	b.pool = wp

	var nN, nC int
	for {
		select {
		case req := <-req: // received a Request...
			nN++
			fmt.Println("Balancer received request. Start to dispatch ...")
			b.dispatch(req) // ...so send it to a Worker
			b.print()
		case w := <-complete: // a worker has finished ...
			nC++
			fmt.Printf("Balancer received the signal of Done.\n\t So far dispatched job count: %d, completed job count: %d\n\n", nN, nC)
			b.completed(w) // ...so update its info
		case <-time.After(10 * time.Second):
			// if anything takes long then the timer's duration, balancer will not wait
			fmt.Println("Maximal waiting time for possible dispatch/completion has elapsed. If the timer was correctly set up, all jobs should have completed.")
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
