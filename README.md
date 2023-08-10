A load balancer works on a pool of workers. "container/heap" is used to manage the pool. In order to use
"container/heap", pool (a slice of workers) needs to satisfy heap.Interface.

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
