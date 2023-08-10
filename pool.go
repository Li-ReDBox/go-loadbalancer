package loadbalancer

type Pool []*Worker

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
