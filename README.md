A load balancer works on a pool of workers. "container/heap" is used to manage the pool. In order to use
"container/heap", pool (a slice of workers) needs to satisfy heap.Interface.

A load balancer is Request specific.

From Go document of `container/heap`: A heap is a tree with the property that each node is the minimum-valued node in its subtree.
The minimum element in the tree is the root, at index 0.

The `Less` method is element specific: in the case of a pool of Workers, the key field in Worker is `pending`,
`Less(i, j int) bool` compares `pending` of elements i and j.
`Swap(i, j int)` can be element specific too: in the case of pool of Workers, `Swap(i, j int)` updates `index`
field after swapping elements i and j.
`Len() int` is the same for sortable slices.

Both `Push(x any)` and `Pop() any` are element specific because of field `index`.

## general methods

```go
func (h someHeap) Len() int           { return len(h) }
func (h someHeap) Less(i, j int) bool { /* element specific comparison */ }
func (h someHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] /* may have element specific fields to be swapped */ }

func (h *someHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
    // Argument x is of type any which needs to be converted to the underneath type first,
    // it may have element specific fields to be created before appending.
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

## Load balancer
### Requirements:
1. In current design, all the requests are sent throw one request channel to LB.
1. LB uses a heap to manage all the workers in the pool.
1. A worker has a channel to receive requests, it has to be a buffered channel. See the question below.
1. All workers in the pool of a LB uses the same channel to notify the completion of a work
1. A worker completes a request, and send the result to the requester through requester's response channel.
1. Work loads are evenly distributed.

## Questions
1. pending of each work peaked at 100, which is its buffer size. What is the rule for choosing buffer size?
1. Waiting a channel blocks the goroutine it is running in. If all goroutines blocked, there is a deadlock.
1. How to deal with too many requests no workers are available?
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
        // time.Sleep(1 * time.Second)
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
	/Users/jli753/Documents/mythings/go_loadbalancer/load_balancer.go:77 +0x58
main.(*Balancer).balance(0x1400012a000, 0x14000100180)
	/Users/jli753/Documents/mythings/go_loadbalancer/load_balancer.go:58 +0x1b4
main.main()
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:76 +0x2fc

goroutine 34 [chan send]:
main.(*Worker).work(0x14000114048, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:66 +0x1d4

goroutine 35 [chan send]:
main.(*Worker).work(0x14000114060, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:66 +0x1d4

goroutine 36 [chan send]:
main.(*Worker).work(0x14000114078, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/load_balancer.go:31 +0x64
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:66 +0x1d4

goroutine 37 [chan send]:
main.requester(0x0?, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:72 +0x290

goroutine 38 [chan send]:
main.requester(0x0?, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:72 +0x290

goroutine 39 [chan send]:
main.requester(0x0?, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:27 +0x5c
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:72 +0x290

goroutine 40 [chan receive]:
main.requester(0x0?, 0x0?)
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:28 +0x68
created by main.main
	/Users/jli753/Documents/mythings/go_loadbalancer/main.go:72 +0x290
exit status 2

A buffered channel can be used like a semaphore, for instance to limit throughput. The capacity of the channel buffer limits the number of simultaneous calls to process.