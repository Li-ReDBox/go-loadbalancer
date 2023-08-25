package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

// there a group of subordinates which listen to a channel, once a value is set in the channel,
// all subordinates stops doing what they were doing. Once every subordinate has reported they
// have done, the controller exit.
// subordinate works on two channels, if one is blocked, the other is blocked too, itself is blocked.
func subordinate(no int, chat chan string, signal chan string) {
	s := 0
	for {
		s++
		select {
		case <-signal:
			fmt.Println("subordinate", no+1, "has received signal")
			signal <- fmt.Sprintf("ACK: subordinate %d has received signal", no+1)
			return
		default:
			time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
			fmt.Printf("Subordinate %d's count is %d\n", no+1, s)
			// cannot send to a closed channel, what is the solution?
			chat <- fmt.Sprintf("%d's count is %d", no+1, s)
		}
	}
}

func controller() {
	nSub := 5
	chat := make(chan string, nSub)
	stop := make(chan string) // not buffered, so stop channel is running in synchronous, easy to spot a deadlock
	stopped := make(chan struct{})

	for i := 0; i < nSub; i++ {
		go subordinate(i, chat, stop)
	}

	go func() {
		// The whole chatting can only last for 5 seconds
		<-time.After(5 * time.Second)
		for i := 0; i < nSub; i++ {
			stop <- "please quit"
			// now we can check if there is a response from a subordinate, but stop channel is synchronous channel, there is more a blocker
			fmt.Println(<-stop)
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
	fmt.Println("Chatted:", time.Since(t))

	// if all reported, let's celebrate
	// if main does not wait, channels are buffered, goroutines are left behind
	// Even with waiting time, there still cases some subordinates are blocked at sending to chat channel. Why? chat channel is buffered, it should not happened!!!
	// time.Sleep(5 * time.Second)
	fmt.Println("Every subordinate has reported, we are done")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
}
