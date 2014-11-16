package main

import "log"

// queue is a queue of packets used internally by Demux2.
type queue []Packet

// NewQueue creatres a new queue with depth of "sz" packets.
func NewQueue(sz int) *queue {
	q := make(queue, 0, sz)
	return &q
}

// Full tests if the queue is full.
func (q *queue) Full() bool {
	return len(*q) == cap(*q)
}

// Put enqueus packet "p".
func (q *queue) Put(p Packet) {
	*q = append(*q, p)
}

// Get dequeues and returns the first packet with selector equal to
// "sel". Returns ok == false if no such packet exists in the queue,
// ok == true otherwise.
func (q *queue) Get(sel int) (p Packet, ok bool) {
	q0 := *q
	for i, p := range q0 {
		if p.sel == sel {
			copy(q0[i:], q0[i+1:])
			// Needed if Packet contains pointers.
			q0[len(q0)-1] = Packet{}
			*q = q0[:len(q0)-1]
			return p, true
		}
	}
	return p, false
}

// Demux2 demultiplexes a sequence of packets received from an input
// channel to several output channels ([]Out) based on each packet's
// selector field. Consumers receiving packets from the
// demultiplexer's output channels must report their availability
// (readiness to receive the next packet) on the "Avail" channel by
// sending their keyed selector value. The demultiplexer uses an
// internal buffer of packets (configurable in depth), and stops when
// the input channel is closed.
type Demux2 struct {
	Out   []chan Packet // Output channels (per consumer).
	Avail chan int      // Avail. reports (from consumers).
	avf   []bool        // Avail. flags (per conumer).
	nbusy int           // # of busy consumers.
	pq    *queue        // Packet queue.
	in    <-chan Packet // Input channel (from producer).
	end   chan int
}

// NewDemux1 creates and returns a new demultiplexer that receives a
// sequence of packets from channel "in", and demultiplexes it to "n"
// output channels. The demultiplexer uses an internal buffer of
// "buffer" packets.
func NewDemux2(n int, in <-chan Packet, buffer int) *Demux2 {
	d := &Demux2{in: in}
	d.Out = make([]chan Packet, n)
	for i := 0; i < n; i++ {
		d.Out[i] = make(chan Packet)
	}
	d.Avail = make(chan int)
	d.avf = make([]bool, n)
	d.nbusy = n
	if buffer < 1 {
		buffer = 1
	}
	d.pq = NewQueue(buffer)
	d.end = make(chan int)
	go d.run()
	return d
}

// run runs as the demultiplexer goroutine.
func (d *Demux2) run() {
	var npck int

	in := d.in
loop:
	for {
		select {
		case pck, ok := <-in:
			if !ok {
				if d.nbusy == 0 {
					// All consumers are idle.  We
					// can exit.
					break loop
				}
				// Delay exit until all consumers become
				// idle.
				in, d.in = nil, nil
				break
			}
			npck++
			if pck.sel < 0 || pck.sel >= len(d.Out) {
				// Drop packet.
				log.Printf("D : %03d/%d: Bad selector!\n",
					pck.id, pck.sel)
				break
			}
			if d.avf[pck.sel] {
				// Emit packet.
				d.avf[pck.sel] = false
				d.nbusy++
				d.Out[pck.sel] <- pck
				log.Printf("D : %03d/%d: --> O%d\n",
					pck.id, pck.sel, pck.sel)
			} else {
				// Enqueue packet.
				d.pq.Put(pck)
				if d.pq.Full() {
					in = nil
				}
			}
		case c := <-d.Avail:
			pck, ok := d.pq.Get(c)
			if ok {
				// Emit packet.
				in = d.in
				d.Out[pck.sel] <- pck
				log.Printf("D : %03d/%d: --> O%d\n",
					pck.id, pck.sel, pck.sel)
			} else {
				// Mark consumer as available.
				d.avf[c] = true
				d.nbusy--
				if d.in == nil && d.nbusy == 0 {
					// Input channel closed, and
					// all consumers idle. We can
					// exit.
					break loop
				}
			}
		}
	}
	for i := range d.Out {
		close(d.Out[i])
	}
	close(d.Avail)
	d.end <- npck
	close(d.end)
}

// Wait waits for Demux2 to stop and returns the total number of
// packets received by it (those forwarded, and those dropped due to
// their sel field being out of range).
func (d *Demux2) Wait() int {
	return <-d.end
}
