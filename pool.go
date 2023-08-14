package loadbalancer

type Pool []*Worker

type Worker struct {
	request chan Request // work to do (buffered channel)
	pending int          // count of pending tasks, it decides the order of Worker in they queue
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // index in the heap
}

func NewWorker(req chan Request) Worker {
	return Worker{
		request: req,
	}
}

func (w *Worker) Work(done chan *Worker) {
	for req := range w.request {
		// fmt.Println("Getting a request from pool for requests")
		// req := <-w.request // get a Request from the pool in balancer
		// fmt.Println("The worker with least load has been received. Run the request and pass on the result to request.")
		// send result to requester by the channel defined in Request
		req.Result <- req.Fn() // call fn and send result
		// fmt.Println("Worker has sent result to Request's channel. Next, tell balancer it is done.")
		done <- w // we've finished this request, notify the pool in balancer
		// fmt.Println("Balancer has been notified from a worker.")
	}
}

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	// A Worker with a smaller pending is in the front
	return p[i].pending < p[j].pending
}

func (p Pool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

// Push and Pop use pointer receivers because they modify the slice's length,
// not just its contents.
func (p *Pool) Push(x any) {
	n := len(*p)
	item := x.(*Worker)
	item.index = n
	*p = append(*p, item)
}

func (p *Pool) Pop() any {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*p = old[0 : n-1]
	return item
}
