package main

import (
	"log"
	"math/rand"
	"time"
)

// Producer1 is a packet producer that emits packets with consecutive
// ids and random selectors at a constant rate. Packets are emitted on
// channel Out. Producer1 drops packets if pushed back.
type Producer1 struct {
	Out  chan Packet
	n    int          // Selector [0..n)
	id   int          // Packet id
	tick *time.Ticker // Packet ticker
	quit chan chan int
}

// NewProducer1 creates and returns a new producer that emits packets
// with random selectors in range [0..n) periodically (one every
// "every" argument).
func NewProducer1(n int, every time.Duration) *Producer1 {
	p := &Producer1{n: n, id: 0}
	p.Out = make(chan Packet)
	p.tick = time.NewTicker(every)
	p.quit = make(chan chan int)
	go p.run()
	return p
}

// run runs as the producer goroutine.
func (p *Producer1) run() {
	var out chan<- Packet
	var pck Packet

	// Initially, no packet to emit
	out = nil
	for {
		select {
		case <-p.tick.C:
			// Generate and emit packet.
			pck = Packet{sel: rand.Intn(p.n), id: p.id}
			p.id++
			out = p.Out
		case out <- pck:
			// Packet emitted.
			log.Printf("P : %03d/%d\n", pck.id, pck.sel)
			out = nil
		case r := <-p.quit:
			p.tick.Stop()
			close(p.Out)
			r <- p.id
			return
		}
	}
}

// Stop stops Producer1 and returns the total number of packets
// produced (either emitted or dropped).
func (p *Producer1) Stop() int {
	r := make(chan int)
	p.quit <- r
	close(p.quit)
	return <-r
}
