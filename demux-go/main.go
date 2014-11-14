package main

import "time"

func ProdCons() {
	p := NewProducer(5, 1*time.Second, 0)
	c := NewConsumer1(1, p.Out, 2*time.Second)

	_ = <-time.After(30 * time.Second)
	p.Stop()
	c.Stop()
}

func ProdDemux1Cons(n int) {
	p := NewProducer(n, 1*time.Second, 0)
	d := NewDemux1(n, p.Out, 0)
	c := make([]*Consumer1, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			c[i] = NewConsumer1(i, d.Out[i], 10*time.Second)
		} else {
			c[i] = NewConsumer1(i, d.Out[i], 0)
		}
	}

	_ = <-time.After(30 * time.Second)
	p.Stop()
	d.Stop()
	for i := 0; i < n; i++ {
		c[i].Stop()
	}
}

func main() {
	//	ProdDemux1Cons(2)
	ProdCons()
}
