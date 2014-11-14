package main

import (
	"fmt"
	"time"
)

type Consumer2 struct {
	sel   int
	in    <-chan Packet
	avail chan<- int
	delay time.Duration
	quit  chan struct{}
}

func NewConsumer2(sel int, in <-chan Packet, avail chan<- int,
	delay time.Duration) *Consumer2 {

	c := &Consumer2{sel: sel, in: in, avail: avail, delay: delay}
	c.quit = make(chan struct{})
	go c.run()
	return c
}

func (c *Consumer2) run() {
	var av chan<- int
	var in <-chan Packet
	var tx <-chan time.Time

	// Initially, report availability.
	av, in, tx = c.avail, nil, nil
	for {
		select {
		case av <- c.sel:
			// Get next packet.
			av, in, tx = nil, c.in, nil
		case pck := <-in:
			if pck.sel != c.sel {
				// Drop packet & report availability.
				fmt.Printf("C%d: %03d/%d: Bad selector!\n",
					c.sel, pck.id, pck.sel)
				av, in, tx = c.avail, nil, nil
			} else {
				// Accept packet, delay for processing.
				fmt.Printf("C%d: %03d/%d\n",
					c.sel, pck.id, pck.sel)
				av, in, tx = nil, nil, time.After(c.delay)
			}
		case <-tx:
			// Report availability,
			av, in, tx = c.avail, nil, nil
		case <-c.quit:
			fmt.Printf("Stopped Consumer2:%d\n", c.sel)
			return
		}
	}
}

func (c *Consumer2) Stop() {
	c.quit <- struct{}{}
	// Prohibit multiple Stop calls
	close(c.quit)
}
