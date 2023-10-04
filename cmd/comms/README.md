# Concurrency programming

goroutine and channel are the building blocks of concurrency programming.

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