// Modified the code example from https://go.dev/blog/io2013-talk-concurrency

// This is a modified buffer.go which tries to run the buffer in an asynchronous way.
// It also has a
// Also the question does it have any benefit of changing buffer.go to this form?
// Because everything running asynchronously, the communication between them has to be set up correctly.
package main

import (
	"log"
	"time"
)

// size limit of the software buffered chanel
const (
	size  = 2000
	loops = 30000
)

// buffered_channel seems run faster than software_channel when size is very low
// with size = 1, loops = 30000, it is 107.049125ms vs 122.057958ms
// with size = 200, loops = 30000, it is 123.412333ms vs 122.693ms
// with size = 1000, loops = 30000, it is 124.32175ms vs 123.773625ms
// with size = 2000, loops = 30000, it is 119.814791ms, 124.367125ms vs 124.6685ms,122.250709ms
func main() {
	log.Println("Buffered channel")
	t := time.Now()
	bufferedChannel(size)
	log.Println(time.Since(t))
	log.Println("Software channel")
	t = time.Now()
	softwareChannel(size)
	log.Println(time.Since(t))
	// pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
}

func bufferedChannel(cl int) {
	ch := make(chan int, cl)
	done := make(chan struct{})
	go func() {
		c := 0
		// prints until out channel is closed
		for i := range ch {
			log.Println("Background printing", i)
			c++
		}
		log.Println("Out channel is closed, so done with printing goroutine, total calls:", c)
		done <- struct{}{}
	}()
	for i := 1; i <= loops; i++ {
		log.Println("Sending", i)
		ch <- i
	}
	log.Println("Sending is done")
	close(ch)
	<-done
}

func softwareChannel(cl int) {
	// create a software channel with a size, like a buffered channel, with two local channels
	in, out := make(chan int), make(chan int)
	go buffer(in, out, cl)
	done := make(chan struct{})
	go func() {
		c := 0
		// prints until out channel is closed
		for i := range out {
			log.Println("Background printing", i)
			c++
		}
		log.Println("Out channel is closed, so done with printing goroutine, total calls:", c)
		done <- struct{}{}
	}()
	for i := 1; i <= loops; i++ {
		log.Println("Sending", i)
		in <- i
	}
	log.Println("Sending is done")
	close(in)
	<-done
}

// A buffered channel does not block until it is full or empty. built-in functions len and cap can be used to them.
// Question: what are the differences between this buffer and a normal buffered channel when the buffer is full?
// Both block send, allow receive operation to continue. Looks like if there is a size restriction, buffered channel
// is easier - no custom code.
func buffer(in <-chan int, out chan<- int, size int) {
	var buf []int

	for in != nil || len(buf) > 0 {
		var i int
		// freshly define a channel variable but not initialised, so by default it blocks.
		var c chan<- int
		// reestablish communication only when there are things to communicate
		if len(buf) > 0 {
			i = buf[0]
			c = out // enable send case
		}
		// use empty default when there is one case in a select statement
		if len(buf) <= size { // lower than the buffer size, try to read and append
			select {
			case n, ok := <-in:
				if ok {
					// log.Println("Pushing into buffer")
					buf = append(buf, n)
				} else {
					// log.Println("Receiving channels has been closed, prepare for exiting buffer goroutine")
					in = nil // disable receive case
				}
			default:
			}
		} else {
			// log.Printf("Pause for increasing buffer, size = %d, currently len = %d\n", size, len(buf))
		}
		// as long as buf is not empty, read
		select {
		case c <- i:
			// log.Println("Popping out from buffer")
			buf = buf[1:]
		default:
		}

	}
	close(out)
	log.Println("Exiting buffer func")
}
