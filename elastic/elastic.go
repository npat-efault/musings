// Copyright (c) 2014, Nick Patavalis (npat@efault.net).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can
// be found in the LICENSE file.

// Package elastic demonstrates a simple implementation of elastic
// (growable) buffers for channels. See:
// https://github.com/npat-efault/musings/wiki/Elastic-channels.rest
package elastic

// T is the element-type for the elastic channel.
type T int

const (
	// Fixed buffer sizes at the send and the receive sides of the
	// elastic channel.
	sendBuffer    = 128
	receiveBuffer = 128
	// Maximum allowed queue size (practically unlimited).
	maxQSz = 0x40000000
	// Number of items to receive form the input channel
	// consecutively.
	maxReceive = 1024
)

// ShrinkMode type encodes values that specify when (and if) the
// elastic buffer is shrinked (reduced in size).
type ShrinkMode int

const (
	// ShrinkMode values.
	Shrink      ShrinkMode = iota // Shrink whenever possible.
	ShrinkEmpty                   // Shrink only when empty.
	NoShrink                      // Never shrink.
)

// ElasticT is an elastic channel of T-typed elements.
type ElasticT struct {
	S chan<- T // Send direction.
	R <-chan T // Receive direction.
}

// NewElasticT creates and returns a new elastic channel, using the
// default shrink mode (Shrink).
func NewElasticT() ElasticT {
	return NewElasticT1(Shrink)
}

// NewElasticT1 creates and returns a new elastic channel, using the
// specified shrink mode.
func NewElasticT1(mode ShrinkMode) ElasticT {
	cin := make(chan T, SendBuffer)
	cout := make(chan T, ReceiveBuffer)
	e := ElasticT{S: cin, R: cout}
	go elasticRun1(mode, cout, cin)
	return e
}

// elasticRun runs as the elastic channel goroutine.
func elasticRun(mode ShrinkMode, cout chan<- T, cin <-chan T) {
	var in <-chan T
	var out chan<- T
	var vi, vo T
	var ok bool

	q := newCQT(1, maxQSz)
	in, out = cin, nil
	for {
		select {
		case vi, ok = <-in:
			if !ok {
				if out == nil {
					close(cout)
					return
				}
				in = nil
				break
			}
			if out == nil {
				vo = vi
				out = cout
			} else {
				q.PushBack(vi)
			}
		case out <- vo:
			vo, ok = q.PopFront()
			if !ok {
				if in == nil {
					close(cout)
					return
				}
				out = nil
				if mode == ShrinkEmpty {
					q.Compact(1)
				}
			}
			if mode == Shrink {
				if q.Len() < q.Cap()>>1 {
					q.Compact(1)
				}
			}
		}
	}
}

// elasticRun1 is a replacement (faster) implementation of elasticRun
// that tries to flush the input channel (cin) and the internal queue
// before returning back to the select statement. It takes advantage
// of the fact that select statements seem to have considerable
// overhead over single-channel reveive / select statements.
func elasticRun1(mode ShrinkMode, cout chan<- T, cin <-chan T) {
	var in <-chan T
	var out chan<- T
	var vi, vo T
	var ok bool

	q := newCQT(1, maxQSz)
	in, out = cin, nil
	for {
		select {
		case vi, ok = <-in:
		inLoop:
			for i := 0; i < maxReceive; i++ {
				if !ok {
					if out == nil {
						close(cout)
						return
					}
					in = nil
					break
				}
				if out == nil {
					vo = vi
					out = cout
				} else {
					q.PushBack(vi)
				}
				select {
				case vi, ok = <-in:
				default:
					break inLoop
				}
			}
		case out <- vo:
		outLoop:
			for {
				vo, ok = q.PopFront()
				if !ok {
					if in == nil {
						close(cout)
						return
					}
					out = nil
					if mode == ShrinkEmpty {
						q.Compact(1)
					}
				}
				if mode == Shrink {
					if q.Len() < q.Cap()>>1 {
						q.Compact(1)
					}
				}
				select {
				case out <- vo:
				default:
					break outLoop
				}
			}
		}
	}
}
