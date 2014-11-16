package main

import (
	"log"
	"time"
)

// Consumer1 is a packet consumer that receives packets from a
// channel, consumes them at a constant rate, and reports its
// availability on a second channel.
type Consumer2 struct {
	sel   int
	in    <-chan Packet
	avail chan<- int
	delay time.Duration
	end   chan int
}

// NewConsumer2 creates and returns a new consumer that receives and
// consumes packets from channel "in", at a constant rate of one
// packet every "delay". It accepts only packets with selectors equal
// to "sel". It reports its availability (ready to receive next
// packet) by emiting its selector on channel "avail". The consumer
// stops when when its input channel is closed.
func NewConsumer2(sel int, in <-chan Packet, avail chan<- int,
	delay time.Duration) *Consumer2 {

	c := &Consumer2{sel: sel, in: in, avail: avail, delay: delay}
	c.end = make(chan int)
	go c.run()
	return c
}

// run runs as the consumer's goroutine.
func (c *Consumer2) run() {
	var npck int

	// Initially, report availability.
	c.avail <- c.sel
	for pck := range c.in {
		npck++
		if pck.sel != c.sel {
			// Drop packet, report availability.
			log.Printf("C%d: %03d/%d: Bad selector!\n",
				c.sel, pck.id, pck.sel)
			c.avail <- c.sel
			continue
		}
		// Delay for processing.
		<-time.After(c.delay)
		log.Printf("C%d: %03d/%d\n", c.sel, pck.id, pck.sel)
		// Report availability
		c.avail <- c.sel
	}
	c.end <- npck
	close(c.end)
}

// Wait waits for the consumer to end and returns the total number of
// packets received by it (either consumed or dropped due to bad
// selectors).
func (c *Consumer2) Wait() int {
	return <-c.end
}
