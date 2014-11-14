package main

import (
	"fmt"
	"time"
)

type Consumer1 struct {
	sel   int
	in    <-chan Packet
	delay time.Duration
	quit  chan struct{}
}

func NewConsumer1(sel int, in <-chan Packet, delay time.Duration) *Consumer1 {
	c := &Consumer1{sel: sel, in: in, delay: delay}
	c.quit = make(chan struct{})
	go c.run()
	return c
}

func (c *Consumer1) run() {
	var in <-chan Packet
	var tx <-chan time.Time

	in = c.in
	for {
		select {
		case pck := <-in:
			if pck.sel != c.sel {
				fmt.Printf("C%d: %03d/%d: Bad selector!\n",
					c.sel, pck.id, pck.sel)
			} else {
				fmt.Printf("C%d: %03d/%d\n",
					c.sel, pck.id, pck.sel)
				// Delay for "processing packet".
				in, tx = nil, time.After(c.delay)
			}
		case <-tx:
			// Get next packet,
			in, tx = c.in, nil
		case <-c.quit:
			return
		}
	}
}

func (c *Consumer1) Stop() {
	fmt.Printf("Stoping Consumer1:%d\n", c.sel)
	c.quit <- struct{}{}
}
