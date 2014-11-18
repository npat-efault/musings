// Auto-generated. !! DO NOT EDIT !!

// Copyright (c) 2014, Nick Patavalis (npat@efault.net).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can
// be found in the LICENSE file.

package elastic

// cQT is a circular queue.
//
// It is implemented with a slice and free running indexes. It starts
// with a user specified initial size (which must be a power of 2) and
// grows exponentially (doubles in size), when required, to accomodate
// more elements (up to a user specified maximum size).
//
// Queue operations are *NOT* thread safe.
type cQT struct {
	sz    uint32 /* current queue size */
	maxSz uint32 /* max queue size */
	m     uint32 /* queue mask (sz - 1) */
	s     uint32 /* start index */
	e     uint32 /* end index */
	b     []T    /* buffer */
}

// newCQT creates and returns a new circular queue.
//
// The queue is initially allocated with space for sz elements. It can
// grow, when required, to accomodate up to maxSz elements. Both sz
// and maxSz *must* be powers of 2.
func newCQT(sz, maxSz int) *cQT {
	if sz <= 0 || uint32(sz)&(uint32(sz)-1) != 0 ||
		uint32(maxSz)&(uint32(maxSz)-1) != 0 ||
		maxSz < sz {
		panic("Invalid Q size")
	}
	cq := &cQT{
		sz: uint32(sz), maxSz: uint32(maxSz),
		m: uint32(sz) - 1,
		s: 0, e: 0,
	}
	cq.b = make([]T, sz)
	return cq
}

// Empty tests if the queue is empty.
func (cq *cQT) Empty() bool {
	return cq.s == cq.e
}

// Full tests if the queue is full.
func (cq *cQT) Full() bool {
	return cq.e-cq.s == cq.maxSz
}

// Len returns the number of elements waiting in the queue.
func (cq *cQT) Len() int {
	return int(cq.e - cq.s)
}

// Cap returns the capacity of the queue (# of element slots currently
// allocated).
func (cq *cQT) Cap() int {
	return int(cq.sz)
}

// MaxCap returns the maximum capacity of the queue (max # of element
// allowed).
func (cq *cQT) MaxCap() int {
	return int(cq.maxSz)
}

// PeekFront returns the front (head) element of the queue, without
// removing it. Returns ok == false if the list is empty (unable to
// peek element), ok == true otherwise.
func (cq *cQT) PeekFront() (el T, ok bool) {
	if cq.s == cq.e {
		return el, false
	}
	return cq.b[cq.s&cq.m], true
}

// PeekBack returns the back (tail) element of the queue, without
// removing it. Returns ok == false if the list is empty (unable to
// peek element), ok == true otherwise.
func (cq *cQT) PeekBack() (el T, ok bool) {
	if cq.s == cq.e {
		return el, false
	}
	return cq.b[(cq.e-1)&cq.m], true
}

// PopHead removes the front (head) element from the queue and returns
// it. Returns ok == false if the list was empty (unable to pop
// element), ok == true otherwise.
func (cq *cQT) PopFront() (el T, ok bool) {
	var zero T
	if cq.s == cq.e {
		return zero, false
	}
	el = cq.b[cq.s&cq.m]
	cq.b[cq.s&cq.m] = zero
	cq.s++
	return el, true
}

// PopBack removes the back (tail) element from the queue and returns
// it. Returns ok == false if the list was empty (unable to pop
// elemnt), ok == true otherwise.
func (cq *cQT) PopBack() (el T, ok bool) {
	var zero T
	if cq.s == cq.e {
		return zero, false
	}
	cq.e--
	el = cq.b[cq.e&cq.m]
	cq.b[cq.e&cq.m] = zero
	return el, true
}

// PushBack adds element "el" to the back (tail) of the queue. Returns
// ok == false if the list was full (unable to push element), ok ==
// true otherwise.
func (cq *cQT) PushBack(el T) (ok bool) {
	if cq.e-cq.s == cq.sz {
		if cq.sz == cq.maxSz {
			return false
		}
		cq.resize(cq.sz << 1)
	}
	cq.b[cq.e&cq.m] = el
	cq.e++
	return true
}

// PushFront adds element "e" to the front (head) of the queue. Returns
// ok == false if the list was full (unable to push element), ok ==
// true otherwise.
func (cq *cQT) PushFront(el T) (ok bool) {
	if cq.e-cq.s == cq.sz {
		if cq.sz == cq.maxSz {
			return false
		}
		cq.resize(cq.sz << 1)
	}
	cq.s--
	cq.b[cq.s&cq.m] = el
	return true
}

// roundUp2 rounds v up to the nearest power of 2
func roundUp2(v uint32) uint32 {
	if v == 0 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

// Compact resizes the queue slice (without removing elements from the
// queue) to the smallest possible size, but not smaller than
// sz. Argument sz *must* be a power of 2. In effect, Compact changes
// the current size of the queue slice to the smalest possible size
// nSz that satisfies all three: (1) nSz is a power of 2, (2) nSz >=
// cq.Len(), (3) nSz >= sz. Compact does not affect the capacity
// (maxSz) of the queue.
func (cq *cQT) Compact(sz int) {
	if sz < 0 || uint32(sz) > cq.maxSz || uint32(sz)&(uint32(sz-1)) != 0 {
		panic("Compact Q with invalid size")
	}
	nSz := roundUp2(cq.e - cq.s)
	if nSz < uint32(sz) {
		nSz = uint32(sz)
	}
	if nSz == cq.sz {
		return
	}
	cq.resize(nSz)
}

// resize, resizes the queue to size sz. The caller *must* make sure
// than sz satisfies all three: (1) sz >= cq.Len(), (2) sz is a power
// of 2, (3) sz <= cq.maxSz
func (cq *cQT) resize(sz uint32) {
	b := make([]T, 0, sz)
	si, ei := cq.s&cq.m, cq.e&cq.m
	if si < ei {
		b = append(b, cq.b[si:ei]...)
	} else {
		b = append(b, cq.b[si:]...)
		b = append(b, cq.b[:ei]...)
	}
	cq.b = b[:sz]
	cq.s, cq.e = 0, cq.e-cq.s
	cq.sz = sz
	cq.m = sz - 1
}
