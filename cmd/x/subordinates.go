package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// there is a group of subordinates which listen to a channel, once a value is set in the channel,
// all subordinates will stops doing what they were doing but what have already started will not be
// stopped until they are successfully completed. Once every subordinate has reported they
// have done, the controller exit.
// subordinate communicates with the controller through two channels: one for hearing and one
// for acknowledging.
func subordinate(no int, chat chan string, signal <-chan string, ack chan<- string) {
	s := 0
	no++
	for {
		s++
		select {
		case m := <-signal:
			if strings.Contains(m, "ACK") {
				panic("Stop channel has been written out of order:" + m)
			}
			fmt.Printf("  subordinate %d has received signal %q.\n", no, m)
			ack <- fmt.Sprintf("    ACK: subordinate %d has received signal", no)
			return
		default:
			time.Sleep(time.Duration(rand.Intn(9000)) * time.Microsecond)
			fmt.Printf("Subordinate %d's count is %d\n", no, s)
			chat <- fmt.Sprintf("%d's count is %d", no, s)
		}
	}
}

func controller() {
	nSub := 500
	chat := make(chan string, nSub)
	stop := make(chan string) // not buffered, running in synchronous, easy to spot a deadlock
	ack := make(chan string)  // not buffered, running in synchronous, easy to spot a deadlock
	stopped := make(chan struct{})

	for i := 0; i < nSub; i++ {
		go subordinate(i, chat, stop, ack)
	}

	go func() {
		// The whole chatting can only last for 5 seconds, but we are very kind to not shut anyone rashly, so, the slowest sayer will
		// take some to quit, controller wait patiently for everyone
		<-time.After(5 * time.Second)
		for i := 0; i < nSub; i++ {
			// commanding for quitting
			stop <- fmt.Sprintf("commanding %d: please quit", i+1)
			// now we wait for acknowledgement. This is a blocking action
			fmt.Printf("Stopped %d with its message: %s\n", i+1, <-ack)
		}
		stopped <- struct{}{}
	}()

	fmt.Println("Lets hear them")
	for {
		select {
		case saying := <-chat:
			fmt.Println(saying)
		case <-stopped:
			fmt.Println("All have been stopped, time to go home")
			return
		}
	}
}

func main() {
	t := time.Now()
	controller()
	// After speech, that subordinate goroutine has nothing blocks it and it goes to check stop signal immediately,
	// if there is no signal, do speech again (sleep), which is a slow action compares to stop action. This goroutine
	// is blocked. This guarantee speech is completed.
	// The whole thing will not stop immediately, so it will be slightly longer than the allowed time.
	fmt.Println("Chatted:", time.Since(t))

	// if all acknowledged, let's exit too
	fmt.Println("Every subordinate has reported, we are done")
	// pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
}
