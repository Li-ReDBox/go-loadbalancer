//go:build ignore && OMIT
// +build ignore,OMIT

// This is a modified buffer.go which tries to run the buffer in an asynchronous way. Not sure if it is possible.
// Also the question does it have any benefit of changing buffer.go to this form?
// Because everything running asynchronously, the communication between them has to be set up correctly:
// On September 19, it works to a degree: no blockage, but the last one is discarded.
package main

import (
	"log"
	"sync"
)

func main() {
	// channels are not buffered
	in, out := make(chan int), make(chan int)
	var wg sync.WaitGroup
	// set up "software" buffer
	go buffer(in, out)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range out {
			log.Println("Background printing", i)
		}
		log.Println("Out channel is closed, so done with printing goroutine")
	}()
	for i := 0; i < 10; i++ {
		log.Println("Sending", i)
		in <- i
	}
	log.Println("Sending is done")
	close(in)
	// the simplest way to get everything out is to block main goroutine exit: by sleep or better using `var wg sync.WaitGroup`
	// time.Sleep(time.Second)
	wg.Wait()
	// pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
}

// buffer provides an unbounded buffer between in and out.  buffer
// exits when in is closed and all items in the buffer have been sent
// to out, at which point it closes out.
// Me: this method does not use buffered channels but itself's buffer. The code runs synchronously:
// batch send, no block, close, then print. Because it has unbounded buffer, it could be dangerous.
func buffer(in <-chan int, out chan<- int) {
	var buf []int

	for in != nil || len(buf) > 0 {
		var i int
		// freshly define a channel variable but not initialised, so by default it blocks.
		var c chan<- int
		// var baton <-chan int
		// reestablish communication only when there are things to communicate
		if len(buf) > 0 {
			i = buf[0]
			c = out // enable send case
		}
		// if we want only bounded buffer, how to temporarily block receive from in channel?
		// select either to receive or send. Uses a baton channel?
		// if len(buf) < 3 {
		// 	baton = in // enable receive
		// }
		select {
		case n, ok := <-in:
			// case n, ok := <-baton:
			// checks if the sending channel has been closed
			if ok {
				log.Println("Pushing into buffer")
				buf = append(buf, n)
			} else {
				log.Println("Receiving channels has been closed, prepare for exiting buffer goroutine")
				in = nil // disable receive case
			}
		case c <- i:
			log.Println("Popping out from buffer")
			buf = buf[1:]
		}
	}
	close(out)
	log.Println("Exiting buffer func")
}
