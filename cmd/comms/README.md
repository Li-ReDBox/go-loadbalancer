# Concurrency programming

goroutine and channel are the building blocks of concurrency programming.

## Basics
A non-buffered channel, sending or receiving is blocked until the other side is ready. This allows goroutines to synchronize without explicit locks or condition variables. In simple words, a chanel has to be used by at least two goroutines, otherwise, that single goroutine (function) is blocked instantly. When code runs into deadlock situation, check if the both ends of communications (channels) are managed correctly by goroutines. It is the readiness of communication critical to avoid deadlock. A function with blocked channels is not callable. Any attempt to call it will cause panic.

A `nil` channel always block. Sending to a closed channel causes panic.

Channels can be buffered. Sends to a buffered channel block only when the buffer is full. Receives block when there is no data to read: either sender has not sent or the buffer is empty.

The number of values in a buffered channel can be checked by `len`. This could be useful when monitoring buffer is important to the code.

Range, Close and Select
A sender can close a channel to indicate that no more values will be sent. Only sender should close a channel, never the receiver. 
The loop `for i := range c` receives values from the channel repeatedly until it is closed. This is simplest and safest way to consume a channel when a channel will be closed.

A receiver can always read values from a closed channel without blocking: those values are zero values of the channel's element type. Status of a channel when receiving can be checked using `v, ok := <-ch`.
`ok` is `false` if the channel has been closed. 

`select` is an `if ... else ...` equivalent in checking communication operations of channels. `default` is executed when no other cases of communications can be established. Be sure to add `default` case when other cases could be unready at the same time.

Running goroutines can cause a resource leak: goroutines consume memory and runtime resources, and heap references
in goroutine stacks keep data from being garbage collected. Goroutines are not able to be garbage collected; they
must exit on their own. Use [this](https://stackoverflow.com/questions/19094099/how-to-dump-goroutine-stacktraces?rq=1) to have a check:
`	pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)`

## Semaphore and buffered channels
A **semaphore** is a synchronization construct that can be used to provide **mutual exclusion** and
**conditional synchronization**. From another perspective, a semaphore is a shared
object that can be manipulated only by two atomic operations, `P` (`wait`) and `V` (`signal`).

There are two types of semaphores: _Counting Semaphore_ and _Binary Semaphore_. _Counting Semaphore_
can be used for mutual exclusion and conditional synchronization. _Binary Semaphore_
is specially designed for mutual exclusion.

A _binary semaphore_ can start with operational or non-operational. The operations are alternate.
But a _counting semaphore_ never blocks and does not have to alternate.

In Go, an unbuffered or synchronous channel combines communication — the exchange of a value
— with synchronization—guaranteeing that two calculations (goroutines) are in a known state.
In other words, unbuffered channels run concurrency in a synchronous way. The minimal setup
is to have two or more goroutines all share the same channel. Buffered channels are used in
asynchronization operations. Channel is a semaphore: either by counting semaphore (size > 1) or binary semaphore (size = 1).

Choosing the correct buffer size to create correct code to avoid race, deadlock and other problems. In general,
channels should be unbuffered, unless there is a good reason for it to be otherwise.

Choosing a buffer size purely by guessing is fragile: if the guess is wrong, send or receive can be blocked.
Instead, receivers should indicate to the senders that they will stop accepting input. Senders use `select`
to switch between sending or non-blocking even quitting. By closing a channel, because a receive operation
on a closed channel can always proceed immediately, yielding the element type’s zero value. Use a deferred
`close()` to make sure a channel is closed when a `select` statement has a `case <-done: return` branch.