// demux-go is a program demonstrating two simple demultiplexer
// implementations. See:
// https://github.com/npat-efault/musings/wiki/A-demultiplexer-in-Go
package main

import (
	"log"
	"time"
)

type conf struct {
	nSel        int
	pEvery      time.Duration
	pBurstEvery int
	pBurstSz    int
	cDelay      time.Duration
	dBuffer     int
	runfor      time.Duration
}

// Producer1 --> Consumer1
func Prod1Cons1(cf conf) {
	p := NewProducer1(cf.nSel, cf.pEvery)
	c := NewConsumer1(0, p.Out, cf.cDelay)

	<-time.After(cf.runfor)
	npro := p.Stop()
	ncon := c.Wait()
	log.Println("Generated", npro, "packets")
	log.Println("Consumer received", ncon, "packets")
}

// Producer2 --> Demux1 --> [cf.nSel * Consumer1]
func Prod2Demux1Cons1(cf conf) {
	p := NewProducer2(cf.nSel, cf.pEvery, cf.pBurstEvery, cf.pBurstSz)
	d := NewDemux1(cf.nSel, p.Out, cf.dBuffer)
	c := make([]*Consumer1, cf.nSel)
	for i := 0; i < cf.nSel; i++ {
		c[i] = NewConsumer1(i, d.Out[i], cf.cDelay)
	}

	<-time.After(cf.runfor)
	npro := p.Stop()
	ndem := d.Wait()
	log.Println("Generated", npro, "packets")
	log.Println("Demux1 received", ndem, "packets")
	for i := range c {
		ncon := c[i].Wait()
		log.Printf("Consumer1 %d received %d packets", i, ncon)
	}
}

// Producer2 --> Demux2 --> [cf.nSel * Consumer2]
func Prod2Demux2Cons2(cf conf) {
	p := NewProducer2(cf.nSel, cf.pEvery, cf.pBurstEvery, cf.pBurstSz)
	d := NewDemux2(cf.nSel, p.Out, cf.dBuffer)
	c := make([]*Consumer2, cf.nSel)
	for i := 0; i < cf.nSel; i++ {
		c[i] = NewConsumer2(i, d.Out[i], d.Avail, cf.cDelay)
	}

	<-time.After(cf.runfor)
	npro := p.Stop()
	ndem := d.Wait()
	log.Println("Generated", npro, "packets")
	log.Println("Demux2 received", ndem, "packets")
	for i := range c {
		ncon := c[i].Wait()
		log.Printf("Consumer2 %d received %d packets", i, ncon)
	}
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	cf := conf{
		nSel:        2,
		pEvery:      1 * time.Second,
		pBurstEvery: 10,
		pBurstSz:    10,
		cDelay:      1 * time.Second,
		dBuffer:     1,
		runfor:      30 * time.Second,
	}
	//Prod1Cons1(cf)
	//Prod2Demux1Cons1(cf)
	Prod2Demux2Cons2(cf)
}
