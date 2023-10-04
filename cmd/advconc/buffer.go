//go:build ignore && OMIT
// +build ignore,OMIT

// Modified the code example from https://go.dev/blog/io2013-talk-concurrency

package main

import (
	"fmt"
)

// main runs synchronously:
// batch send define number of data, close sending channel, drain receiving channel, then main goroutine exits.
func main() {
	in, out := make(chan int), make(chan int)
	go buffer(in, out)
	for i := 0; i < 10; i++ {
		fmt.Println("Sending")
		in <- i
	}
	close(in)
	fmt.Println("Sending is done.")
	for i := range out {
		fmt.Println(i)
	}
}

// buffer provides an unbounded buffer between in and out.  buffer
// exits when the in channel is closed and all items in the buffer have been sent
// to out, at which point it closes out.
// The buffer acts as a normal channel. Because it is unbounded, so it never blocks like which happens when a buffered channel reaches its size limit.
// That's been said, it is not possible to define an unbounded buffered channel out of box.
// Both receiving and sending channels it manages use close to signal completion.
// Because it has unbounded buffer, it could be dangerous if buf grows quicker than shrinks as with any unbound data structures.
// It can be blocked when both in and out channels are blocked.
func buffer(in <-chan int, out chan<- int) {
	var buf []int
	// if in channel is not disabled (it will be disabled internally when it is closed outside), or there are data to be processed, run the loop.
	for in != nil || len(buf) > 0 {
		var i int
		var c chan<- int
		// if there are data in the queue, connect a local channel to the receiver's of this buffer.
		if len(buf) > 0 {
			i = buf[0]
			c = out // enable send case
		}
		// read from in channel or send to out channel when anyone is ready
		select {
		case n, ok := <-in:
			if ok {
				buf = append(buf, n)
			} else {
				in = nil // disable receive case
			}
		case c <- i: // send to receiver, and pop one out of top
			buf = buf[1:]
		}
	}
	close(out)
}
