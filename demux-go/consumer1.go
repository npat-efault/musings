package main

import (
	"log"
	"time"
)

// Consumer1 is a packet consumer that receives packets from a channel
// and consumes them at a constant rate. It accepts only packets with
// a given selector value. The consumer stops when its input channel
// is closed.
type Consumer1 struct {
	sel   int
	in    <-chan Packet
	delay time.Duration
	end   chan int
}

// NewConsumer1 creates and returns a new consumer that receives
// packets from channel "in", and consumes then at a constant rate of
// one packet every "delay". It accepts only packets with selectors
// equal to "sel".
func NewConsumer1(sel int, in <-chan Packet,
	delay time.Duration) *Consumer1 {

	c := &Consumer1{sel: sel, in: in, delay: delay}
	c.end = make(chan int)
	go c.run()
	return c
}

// run runs as the consumer's goroutine.
func (c *Consumer1) run() {
	var npck int

	for pck := range c.in {
		npck++
		if pck.sel != c.sel {
			log.Printf("C%d: %03d/%d: Bad selector!\n",
				c.sel, pck.id, pck.sel)
			continue
		}
		// Delay for "processing packet".
		<-time.After(c.delay)
		log.Printf("C%d: %03d/%d\n", c.sel, pck.id, pck.sel)
	}
	c.end <- npck
	close(c.end)
}

// Wait waits for the consumer to end and returns the total number of
// packets received by it (either consumed or dropped due to bad
// selectors).
func (c *Consumer1) Wait() int {
	return <-c.end
}
