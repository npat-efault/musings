package main

type queue []Packet

func NewQueue(sz int) *queue {
	return &make(queue, 0, sz)
}

func (q *queue) Full() bool {
	return len(*q) == cap(*q)
}

func (q *queue) Put(p Packet) {
	*q = append(*q, p)
}

func (q *queue) Get(sel int) (p Packet, ok bool) {
	q0 := *q
	for i, p := range q0 {
		if p.sel == sel {
			copy(q0[i:], q0[i+i:])
			// Necessary if Packet contains pointers.
			q0[len(q0)-1] = Packet{}
			*q = q0[:len(q0)-1]
			return p, true
		}
	}
	return p, false
}

type Demux2 struct {
	Out   []chan Packet // Output channels (per consumer).
	Avail chan int      // Avail. reports (from consumers).
	avf   []bool        // Avail. flags (per conumer).
	in    <-chan Packet // Input channel (from producer).
	quit  chan struct{}
}

func NewDemux2(n int, in <-chan Packet, buffer int) *Demux2 {
	d := &Demux1{in: in}
	d.Out = make([]chan Packet, n)
	for i := 0; i < n; i++ {
		d.Out[i] = make(chan Packet, buffer)
	}
	d.Avail = make(chan int)
	d.avf = make([]bool, n)
	d.quit = make(chan struct{})
	go d.run()
	return d
}
