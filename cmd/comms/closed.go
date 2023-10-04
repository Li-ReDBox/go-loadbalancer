//go:build ignore && OMIT
// +build ignore,OMIT

// When sending data to a closed channel the code will panic. But a closed channel does not block reading from it,
// it just emits zero values of the type of the channel. We can use this behaviour to coordinate one-off event like
// broadcasting quitting: create a channel of struct, pass on to the goroutines need to be notified. Because the channel
// is empty, reading is blocked. By simply close it when the time is up for quitting, the blockage is removed, once
// blocked case can run, for example quit the goroutine.
// This is useful when calling goroutine only cares the quitting of other goroutines which do not return data.

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func say(q chan struct{}) {
	for {
		select {
		// because it is intentionally will never send data to the channel, reading from the channel is blocked until it is close.
		case <-q:
			fmt.Println("    Alright, as you wish, I will quit")
			return
		default:
			time.Sleep(time.Duration(rand.Intn(5)*500) * time.Millisecond)
			fmt.Println("I am saying ...")
		}
	}
}

// STARTMAIN OMIT
func main() {
	// Here we only care about the goroutines created by us are completed, we simpl use Waitgroup to count them
	var wg sync.WaitGroup

	shut := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			say(shut)
			wg.Done()
		}()
	}

	// Notify the say goroutines to stop after 3s.
	time.AfterFunc(3*time.Second, func() {
		fmt.Println("Shut up you talkers")
		// This demonstrates that by simply closing a channel equals to send a zero value to the channel
		// which can be used as a signal
		close(shut)
	})

	// Wait them all finished
	wg.Wait()
	// There should have not goroutine lingering
	panic("Have you all gone?")
}
