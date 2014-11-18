package elastic

import (
	"testing"
	"time"
)

func TestBlock(t *testing.T) {
	elc := NewElasticT()
	for i := 0; i < 8192; i++ {
		select {
		case elc.S <- T(i):
		case <-time.After(1 * time.Millisecond):
			t.Fatal("Blocked on write:", i)
		}
	}
	close(elc.S)
	for i := 0; i < 8192; i++ {
		select {
		case _, ok := <-elc.R:
			if !ok {
				t.Fatal("Closed @ read:", i)
			}
		case <-time.After(1 * time.Millisecond):
			t.Fatal("Blocked on read:", i)
		}
	}
	if _, ok := <-elc.R; ok {
		t.Fatal("Channel not closed!")
	}
}

func consume(n int, c <-chan T, end chan<- int) {
	var i int
	for i = 0; i < n; i++ {
		q := <-c
		if q != T(i) {
			break
		}
	}
	end <- i
}

func produce(n int, c chan<- T, end chan<- int) {
	var i int
	for i = 0; i < n; i++ {
		c <- T(i)
	}
	close(c)
	end <- i
}

func TestData(t *testing.T) {
	const N = 8192
	elc := NewElasticT()
	endP := make(chan int)
	endC := make(chan int)
	go produce(N, elc.S, endP)
	go consume(N, elc.R, endC)
	r := <-endP
	if r != N {
		t.Fatalf("Producer ret %d != %d", r, N)
	}
	r = <-endC
	if r != N {
		t.Fatalf("Consumer ret %d != %d", r, N)
	}
}

func benchFixed(b *testing.B, buffer int) {
	c := make(chan T, buffer)
	end := make(chan int)
	b.ResetTimer()
	go produce(b.N, c, end)
	go consume(b.N, c, end)
	<-end
	<-end
	b.StopTimer()
}

func BenchmarkFixedNoBuf(b *testing.B) {
	benchFixed(b, 0)
}

func BenchmarkFixed1024(b *testing.B) {
	benchFixed(b, 1024)
}

func benchSeq(b *testing.B, mode ShrinkMode) {
	elc := NewElasticT1(mode)
	end := make(chan int)
	b.ResetTimer()
	go produce(b.N, elc.S, end)
	<-end
	go consume(b.N, elc.R, end)
	<-end
	b.StopTimer()
}

func BenchmarkSeqShrink(b *testing.B) {
	benchSeq(b, Shrink)
}

func BenchmarkSeqShrinkEnpty(b *testing.B) {
	benchSeq(b, ShrinkEmpty)
}

func BenchmarkSeqNoShrink(b *testing.B) {
	benchSeq(b, NoShrink)
}

func benchCon(b *testing.B, mode ShrinkMode) {
	elc := NewElasticT1(mode)
	end := make(chan int)
	b.ResetTimer()
	go produce(b.N, elc.S, end)
	go consume(b.N, elc.R, end)
	<-end
	<-end
	b.StopTimer()
}

func BenchmarkConShrink(b *testing.B) {
	benchCon(b, Shrink)
}

func BenchmarkConShrinkEmpty(b *testing.B) {
	benchCon(b, ShrinkEmpty)
}

func BenchmarkConNoShrink(b *testing.B) {
	benchCon(b, NoShrink)
}
