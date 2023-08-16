# A load based load balancer

## Introduction
A load balancer acts like an exchange between requesters and workers. A load balancer uses a pool of workers.
__"container/heap"__ is used to manage the pool. In order to use **"container/heap"**, pool _`type`_ (a slice of workers)
needs to satisfy heap.Interface. A load balancer is Request specific in the current implementation using channels.

The data flow looks like a customer files a complaint to a company, the company then assigns an agent to contact the
customer directly to resolve it. Once the issues has been resolved, the agent first notify the customer the outcome
then files a report to the company to notify a resolution. To be an efficient system, the complaint has to be well
formed and understood in the company, by the customer and the agent.

To communicate with requesters and workers, a load balancer needs two channels: a channel of Request and a channel of *Worker.
In current design, `Worker` contains a channel of Request which defines what to do and a channel for notifying the
outcome to the requester. In such implementation, different requester can request different work (even the same type)
and can be notified directly of its outcome. 

In `Request`, it has a task (a specif func, which the whole system understands, represents what to do), a return channel.
It links a Worker and a requester. In this implementation, the load balancer manages computer of `func() int`.
Of course there can be all sorts of computers, which is very hard to abstract. Therefore a load balancer is request
bounded.

A load balancer uses heap to manage `Worker`s by adjusting their loadings and rearrange their positions accordingly. 
From Go document of `container/heap`: A heap is a tree with the property that each node is the minimum-valued node
in its subtree. The minimum element in the tree is the root, at index 0.

## Generalisation of heap in the load balancer
**"container/heap"** expects a slice type to be managed satisfies `sort.Interface` and `Push(x any)` and `Pop() any`.
The `Less` method is element specific: in the case of a pool of `Worker`s, the key field in `Worker` is `pending`:
`Less(i, j int) bool` compares `pending` of elements i and j.
`Swap(i, j int)` can be element specific too: in the case of pool of `Worker`s, `Swap(i, j int)` updates `index`
field after swapping elements i and j. Both `Push(x any)` and `Pop() any` are element specific because of field `index`.

```go
func (h someHeap) Len() int           { return len(h) }
func (h someHeap) Less(i, j int) bool { /* element specific comparison */ }
func (h someHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] /* may have element specific fields to be swapped */ }

// Push and Pop use pointer receivers because they modify the slice's length,
// not just its contents.
func (h *someHeap) Push(x any) {
    // Argument x is of type any which needs to be converted to the underneath type first,
    // it may have element specific fields to be manipulated before appending.
    // In the below example, the underneath type is `int`.
	*h = append(*h, x.(int))
}

func (h *someHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
    // it may have element specific fields to be managed before appending.
    // for example for a PriorityQueue manages structs with a field named `index`:
    // old[n-1] = nil  // avoid memory leak
	// x.index = -1    // for safety
	return x
}

```

There are a few lines of basic lines are the same for all Heap methods but there is no easy and clean way to
abstract and reuse code. It is much easy to copy and modify the template code to suit a specific data structure.

## Load balancer (LB)

### Use a Request channel without LB
```go
	// Just a demonstration if there is no Worker and Balancer, how a Request
	// which was generated from a load generator is processed.
	r := make(chan Request)
	go func(r chan Request) {
            c := make(chan int)
            for i := 0; i < nRequest; i++ {
                r <- Request{fn, c}
            }   
        } (r)

        for req := range r {
            req := <-r
            req.Result <- req.Fn()
        }
```

### Requirements
In current design, 
1. all requesters send their requests through a shared request channel and this channel is monitored by LB;
1. a worker has a channel to receive requests, it has to be a buffered channel. See the question below;
1. when a worker completes a request it sends the result to the requester through requester's response channel;
1. all workers in the pool notify their completion through a shared channel and this channel is monitored
   by LB;
1. LB uses a heap to manage all the workers in the pool to maintain evenly distributed loadings.

nRequester requester gorountines send requests through a shared channel, once a request is sent, it will not 
send more until it has been completed. So for each requester it works in a synchronised way. When dispatch is
blocked, Request channel is blocked in `dispatch` and no requester can send requests.

nWorker Workers goroutines notify LB on completion through a shared complete channel. Once a work has been
received it starts to work on it and will not have capacity to receive another one until it sends result to
the requester (which is always OK because there is alway a requester goroutine ready to receive) and LB.
When dispatch is blocked complete channel is blocked in `dispatch` and no Worker can send.

Waiting for a channel ready for an operation (send/receive) blocks the goroutine the operation is running in.
If all goroutines are blocked, there is a deadlock. A buffered channel can be used like a semaphore, for
instance to limit throughput. The capacity of the channel buffer limits the number of simultaneous calls
to process.

The bottleneck lies on the Request channel of Worker.work because nWorker is smaller than nRequester,
`case req := <-req:` branch will be blocked by waiting for the popped Worker with the smallest `pending`
ready to receive but at the same time that Worker is wanting to send complete channel so, it is blocked
then everything is blocked.


There is no delay in sending Requests from requester, there has to be enough capacity to receive all of
them at once. Then the requesters will pause before sending another round. No significant evidence has been
found when run notification of completion synchronised or concurrent way because once first batch of requests
have been sent, the system is in a stable state: only one a request is done, one new request will be sent.

Maybe oddly, "less is more": using a non-buffered complete channel `comp := make(chan *lb.Worker)` has a
better performance: 38s 35s 39s 27s 33s
vs
`comp := make(chan *lb.Worker, nWorker)`
36s 40s 37s 43s 43s

Maybe too many waiting and orchestration?

## Questions
1. How to deal with too many requests no workers are available? This may never have a perfect answer.
