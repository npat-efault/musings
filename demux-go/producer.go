package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Packet struct {
	sel int
	id  int
}

type Producer struct {
	Out    chan Packet
	n      int
	id     int
	ticker *time.Ticker
	quit   chan struct{}
}

func NewProducer(n int, every time.Duration, buffer int) *Producer {
	p := &Producer{n: n, id: 0}
	p.Out = make(chan Packet, buffer)
	p.ticker = time.NewTicker(every)
	p.quit = make(chan struct{})
	go p.run()
	return p
}

func (p *Producer) run() {
	var tick <-chan time.Time
	var out chan<- Packet
	var pck Packet

	tick = p.ticker.C
	out = nil
	for {
		select {
		case <-tick:
			pck = Packet{sel: rand.Intn(p.n), id: p.id}
			p.id++
			// emit packet
			out = p.Out
			tick = nil
			fmt.Printf("P : %03d/%d\n", pck.id, pck.sel)
		case out <- pck:
			// wait next tick
			out = nil
			tick = p.ticker.C
		case <-p.quit:
			p.ticker.Stop()
			return
		}
	}
}

func (p *Producer) Stop() {
	fmt.Println("Stoping Producer")
	p.quit <- struct{}{}
	// Prohibit multiple Stop calls
	close(p.quit)
}
