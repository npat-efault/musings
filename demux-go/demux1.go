package main

import "log"

// Demux1 de-multiplexes a sequence of packets received from an input
// channel to several output channels ([]Out) based on each packet's
// selector field. Each output channel has a packet buffer of
// configurable depth. Demux1 stops when the input channel is closed.
type Demux1 struct {
	Out []chan Packet // Output channels
	in  <-chan Packet
	end chan int
}

// NewDemux1 creates and returns a new demultiplexer that receives a
// sequence of packets from channel "in", and demultiplexes it
// (forwards each packet accordingly) to "n" output channels. Each
// output channel has a buffer of "buffer" packets.
func NewDemux1(n int, in <-chan Packet, buffer int) *Demux1 {
	d := &Demux1{in: in}
	d.Out = make([]chan Packet, n)
	for i := 0; i < n; i++ {
		d.Out[i] = make(chan Packet, buffer)
	}
	d.end = make(chan int)
	go d.run()
	return d
}

// run runs as the demultiplexer goroutine.
func (d *Demux1) run() {
	var npck int

	for pck := range d.in {
		npck++
		if pck.sel < 0 || pck.sel >= len(d.Out) {
			// Drop packet.
			log.Printf("D : %03d/%d: Bad selector!\n",
				pck.id, pck.sel)
			continue
		}
		// Emit packet.
		d.Out[pck.sel] <- pck
		log.Printf("D : %03d/%d: --> O%d (%d)\n",
			pck.id, pck.sel, pck.sel, len(d.Out[pck.sel]))
	}
	for i := range d.Out {
		close(d.Out[i])
	}
	d.end <- npck
	close(d.end)
}

// Wait waits for Demux1 to stop and returns the total number of
// packets received by the demultiplexer (those forwarded, and those
// dropped due to their sel field being out of range).
func (d *Demux1) Wait() int {
	return <-d.end
}
