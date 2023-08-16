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
If all goroutines are blocked, there is a deadlock.
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
  When 4 requesters request flat out without delay and Workers do not have buffered request channels,
  the system is blocked at 9 dispatched, 3 completed
+
diff --git a/main.go b/main.go
index 7fe0a35..90b61dd 100644
--- a/main.go
+++ b/main.go
@@ -23,7 +23,7 @@ func requester(work chan<- Request, nWorker int) {
        // each requester only allow to 10 requests
        for i := 0; i < 10; i++ {
                // Kill some time (fake load). Do not flat out.
-               time.Sleep(time.Duration(rand.Int63n(1e3 * int64(nWorker))))
+               // time.Sleep(time.Duration(rand.Int63n(1e3 * int64(nWorker))))
                work <- Request{workFn, c} // send request, blocks
                <-c                        // the result of workFn only returns boring 1, so discard by just draining the channel
                // result := <-c              // wait for answer until there is one
@@ -44,13 +44,13 @@ func main() {
        // time.Sleep(1 * time.Second)lb.
        // End of the simple dome

-       nRequester := 8
+       nRequester := 4
        workers := 3
        wp := make(Pool, workers)

        for i := 0; i < workers; i++ {
                wp[i] = &Worker{
-                       request: make(chan Request, nRequester), // this is a buffered channel
+                       request: make(chan Request), // this is a buffered channel
                }
        }

  dispatch: tries to send request to worker's request channel, but it is full, so dispatch is blocked
  all 3 Worker.work: tries to notify balancer a work is done, but the channel is blocked because dispatch is blocked
  all 4 requesters: tries to send or receives requests are blocked by the request channel in the balancer.
  All 8 goroutines are blocked at somewhere.

  Only Worker has higher possibility of blocking because others just read and write at the same time. Once every Worker has a buffered channel, even just one, there will be 3 more worker goroutine crated.

  The deadlock appears in Worker:

fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.(*Balancer).dispatch(0x1400012a000, {0x102ad34f0, 0x14000096000})
	/go_loadbalancer/load_balancer.go:77 +0x58
main.(*Balancer).balance(0x1400012a000, 0x14000100180)
	/go_loadbalancer/load_balancer.go:58 +0x1b4
main.main()
	/go_loadbalancer/main.go:76 +0x2fc

goroutine 34 [chan send]:
main.(*Worker).work(0x14000114048, 0x0?)
	/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/go_loadbalancer/main.go:66 +0x1d4

goroutine 35 [chan send]:
main.(*Worker).work(0x14000114060, 0x0?)
	/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/go_loadbalancer/main.go:66 +0x1d4

goroutine 36 [chan send]:
main.(*Worker).work(0x14000114078, 0x0?)
	/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/go_loadbalancer/main.go:66 +0x1d4

goroutine 37 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/go_loadbalancer/main.go:72 +0x290

goroutine 38 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/go_loadbalancer/main.go:72 +0x290

goroutine 39 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/go_loadbalancer/main.go:72 +0x290

goroutine 40 [chan receive]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/main.go:28 +0x68
created by main.main
	/go_loadbalancer/main.go:72 +0x290
exit status 2

A buffered channel can be used like a semaphore, for instance to limit throughput. The capacity of the channel buffer limits the number of simultaneous calls to process.

```go
	nRequester := 5 // this is the maximal pending total: each requester will wait until last request has completed before a new request is sent
	nWorker := 3

        // request channel is non-buffered
	for i := 0; i < nWorker; i++ {
		w := lb.NewWorker(make(chan lb.Request))
		wp[i] = &w
	}

        comp := make(chan *lb.Worker, nWorker)
        r := make(chan lb.Request)
```

Deadlock error when all 9 goroutines(3 Workers, 5 requesters, 1 Balancer.dispatch) are blocked:
1. funmech.com/loadbalancer.(*Balancer).dispatch cannot send to worker a new request: w.request <- req
2. 2 Worker cannot receive: funmech.com/loadbalancer.(*Worker).Work: for req := range w.request {}
3. 1 Worker cannot send: funmech.com/loadbalancer.(*Worker).Work: done <- w
4. 4 requesters cannot send request:  main.requester: work <- lb.Request{Fn: workFn, Result: c}
5. 1 requester cannot receive: main.requester: <-c
6. only buffered channel is completion channel.

```shell
Balancer received the signal of Done.
	 So far dispatched job count: 7, completed job count: 4

Balancer received request. Start to dispatch ...
	worker 0: pending = 1
	worker 1: pending = 1
	worker 2: pending = 2

Balancer received request. Start to dispatch ...
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
funmech.com/loadbalancer.(*Balancer).dispatch(0x1400000c168, {0x10219b570, 0x1400008e120})
	/go_loadbalancer/load_balancer.go:60 +0x58
funmech.com/loadbalancer.(*Balancer).Balance(0x1400000c168, {0x1400000c030, 0x3, 0x3}, 0x14000068180, 0x140000221e0)
	/go_loadbalancer/load_balancer.go:41 +0x248
main.main()
	/go_loadbalancer/cmd/buffered/main.go:58 +0x210

goroutine 6 [chan receive]:
funmech.com/loadbalancer.(*Worker).Work(0x1400000c048, 0x0?)
	/go_loadbalancer/pool.go:19 +0x70
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:46 +0xe0

goroutine 7 [chan receive]:
funmech.com/loadbalancer.(*Worker).Work(0x1400000c060, 0x0?)
	/go_loadbalancer/pool.go:19 +0x70
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:46 +0xe0

goroutine 8 [chan send]:
funmech.com/loadbalancer.(*Worker).Work(0x1400000c078, 0x0?)
	/go_loadbalancer/pool.go:26 +0x64
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:46 +0xe0

goroutine 9 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:52 +0x188

goroutine 10 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:52 +0x188

goroutine 11 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:52 +0x188

goroutine 12 [chan receive]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:24 +0x68
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:52 +0x188

goroutine 13 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:52 +0x188
exit status 2

```

When non channels is buffered, no completion can be notified, all 9 goroutines are blocked:
```shell
Balancer received request. Start to dispatch ...
	worker 0: pending = 0
	worker 1: pending = 0
	worker 2: pending = 1

Balancer received request. Start to dispatch ...
	worker 0: pending = 0
	worker 1: pending = 1
	worker 2: pending = 1

Balancer received request. Start to dispatch ...
	worker 0: pending = 1
	worker 1: pending = 1
	worker 2: pending = 1

Balancer received request. Start to dispatch ...
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
funmech.com/loadbalancer.(*Balancer).dispatch(0x14000120138, {0x100f3f570, 0x14000096120})
	/go_loadbalancer/load_balancer.go:60 +0x58
funmech.com/loadbalancer.(*Balancer).Balance(0x14000120138, {0x14000120018, 0x3, 0x3}, 0x140001001e0, 0x14000100180)
	/go_loadbalancer/load_balancer.go:41 +0x248
main.main()
	/go_loadbalancer/cmd/buffered/main.go:59 +0x210

goroutine 34 [chan send]:
funmech.com/loadbalancer.(*Worker).Work(0x14000120030, 0x0?)
	/go_loadbalancer/pool.go:26 +0x64
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:47 +0xe0

goroutine 35 [chan send]:
funmech.com/loadbalancer.(*Worker).Work(0x14000120048, 0x0?)
	/go_loadbalancer/pool.go:26 +0x64
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:47 +0xe0

goroutine 36 [chan send]:
funmech.com/loadbalancer.(*Worker).Work(0x14000120060, 0x0?)
	/go_loadbalancer/pool.go:26 +0x64
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:47 +0xe0

goroutine 37 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:53 +0x188

goroutine 38 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:53 +0x188

goroutine 39 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:53 +0x188

goroutine 40 [chan receive]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:24 +0x68
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:53 +0x188

goroutine 41 [chan send]:
main.requester(0x0?, 0x0?)
	/go_loadbalancer/cmd/buffered/main.go:23 +0x5c
created by main.main
	/go_loadbalancer/cmd/buffered/main.go:53 +0x188
exit status 2
```

With completion channel buffed with 1, one completed job can be notified, then all 9 goroutines are deadlocked:
```shell
diff --git a/cmd/buffered/main.go b/cmd/buffered/main.go
index 492e453..e5eb930 100644
--- a/cmd/buffered/main.go
+++ b/cmd/buffered/main.go
@@ -35,12 +35,13 @@ func main() {
        nWorker := 3
        wp := make(lb.Pool, nWorker)
 
+       // when a non-buffered channel is used here, workers can hold up to 4 requests, then goes into deadlock
        for i := 0; i < nWorker; i++ {
-               w := lb.NewWorker(make(chan lb.Request, 1))
+               w := lb.NewWorker(make(chan lb.Request))
                wp[i] = &w
        }
 
-       comp := make(chan *lb.Worker, nWorker)
+       comp := make(chan *lb.Worker, 1)
        // set all workers with the same completion notification channel ready for receiving Requests
        for _, w := range wp {
                go w.Work(comp)

Balancer received the signal of Done.
	 So far dispatched job count: 4, completed job count: 1

Balancer received request. Start to dispatch ...
fatal error: all goroutines are asleep - deadlock!
```