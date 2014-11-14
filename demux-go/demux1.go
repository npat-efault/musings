package main

import "fmt"

type Demux1 struct {
	Out  []chan Packet
	in   <-chan Packet
	quit chan struct{}
}

func NewDemux1(n int, in <-chan Packet, buffer int) *Demux1 {
	d := &Demux1{in: in}
	d.Out = make([]chan Packet, n)
	for i := 0; i < n; i++ {
		d.Out[i] = make(chan Packet, buffer)
	}
	d.quit = make(chan struct{})
	go d.run()
	return d
}

func (d *Demux1) run() {
	var in <-chan Packet
	var out chan<- Packet
	var pck Packet

	in = d.in
	for {
		select {
		case pck = <-in:
			if pck.sel < 0 || pck.sel >= len(d.Out) {
				// Drop packet.
				fmt.Printf("D : %03d/%d: Bad selector!\n",
					pck.id, pck.sel)
				break
			}
			// Emit packet.
			in, out = nil, d.Out[pck.sel]
			fmt.Printf("D : %03d/%d: --> O%d\n",
				pck.id, pck.sel, pck.sel)
		case out <- pck:
			// Wait for next packet.
			in, out = d.in, nil
		case <-d.quit:
			return
		}
	}
}

func (d *Demux1) Stop() {
	fmt.Println("Stoping Demux1")
	d.quit <- struct{}{}
	// Prohibit multiple Stop calls
	close(d.quit)
}
