package main

import (
	"log"
	"math/rand"
	"time"
)

// Producer2 is a packet producer that emits packets with random
// selectors and consecuitive ids at a constant rate, with periodic
// bursts. During bursts packets are emmited one every
// FastTick. Packets are emitted on channel Out. Producer2 drops
// packets if pushed back.
type Producer2 struct {
	Out    chan Packet
	n      int          // Selector [0..n)
	id     int          // Packet id
	bEvery int          // Burst every bEvery slow ticks
	bSz    int          // Burst size
	stick  *time.Ticker // Slow ticker (ordinary packets)
	ftick  *time.Ticker // Fast ticker (burst packets)
	quit   chan chan int
}

// NewProducer2 returns a new producer that emits packets with random
// selectors in range [0..n), periodically (one every "every"
// argument). Every "burstEvery" packets it generates a burst of
// "burstSz" packets.
func NewProducer2(n int,
	every time.Duration, burstEvery, burstSz int) *Producer2 {

	p := &Producer2{n: n, id: 0, bEvery: burstEvery, bSz: burstSz}
	p.Out = make(chan Packet)
	p.stick = time.NewTicker(every)
	p.quit = make(chan chan int)
	go p.run()
	return p
}

// FastTick is the delay between packets during bursts.
const FastTick = 1 * time.Millisecond

// run runs as the producer's goroutine.
func (p *Producer2) run() {
	var tick <-chan time.Time
	var out chan<- Packet
	var pck Packet

	tick, out = p.stick.C, nil
	burst, npck := false, p.bEvery
	for {
		select {
		case <-tick:
			if npck == 0 {
				if burst {
					// Burst ended, back to normal tick.
					tick = p.stick.C
					p.ftick.Stop()
					p.ftick = nil
					npck = p.bEvery
					burst = false
				} else {
					// Time for burst.
					p.ftick = time.NewTicker(FastTick)
					tick = p.ftick.C
					npck = p.bSz
					burst = true
				}
				break
			}
			// Generate and emit packet.
			pck = Packet{sel: rand.Intn(p.n), id: p.id}
			p.id++
			out = p.Out
			npck--
		case out <- pck:
			// Packet emitted.
			log.Printf("P : %03d/%d\n", pck.id, pck.sel)
			out = nil
		case r := <-p.quit:
			p.stick.Stop()
			if p.ftick != nil {
				p.ftick.Stop()
			}
			close(p.Out)
			r <- p.id
			return
		}
	}
}

// Stop stops Producer2 and returns the total number of packets
// produced (either emitted or dropped).
func (p *Producer2) Stop() int {
	r := make(chan int)
	p.quit <- r
	close(p.quit)
	return <-r
}
