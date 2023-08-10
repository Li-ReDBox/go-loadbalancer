// This example demonstrates a WorkerPool using the heap interface.
// go test example_workers_test.go pool.go
// package loadbalancer
package main

import (
	"container/heap"
	"fmt"
)

// ExamplePool_Pop creates a Pool with some items, then add a new Worker.
// At the last, pop all of them.
func ExamplePool_Pop() {
	pendings := []int{1, 30, 29, 15, 27}
	wp := make(Pool, len(pendings))

	for i, p := range pendings {
		wp[i] = &Worker{
			pending: p,
			index:   i,
		}
	}

	heap.Init(&wp)

	// Push a new worker with pending of 3
	heap.Push(&wp, &Worker{pending: 3})

	// After Push, the order is not fully established correctly. Why?
	// Check the popped workers' pending - it should be in increase order.
	for wp.Len() > 0 {
		w := heap.Pop(&wp).(*Worker)
		fmt.Printf("%d ", w.pending)
	}

	// Output:
	// 1 3 15 27 29 30
}
