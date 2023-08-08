package loadbalancer

import "fmt"

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p *Pool) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	// fmt.Println("popped out", n-1, "its index was", item.index)
	item.index = -1 // for safety
	*p = old[0 : n-1]
	return item
}

func (p *Pool) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	// fmt.Println("Pushing in", x.(*Worker).index)
	n := len(*p)
	item := x.(*Worker)
	item.index = n
	*p = append(*p, item)
}

func (p Pool) Swap(i, j int) {
	// fmt.Println("Swapping", i, j)
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

func (p Pool) Check() {
	for _, v := range p {
		fmt.Println(v.index, " has loading of", v.pending)
	}
}
