// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

import (
	"encoding/binary"
	"math"
)

const (
	rDims       = 2
	rMaxEntries = 16
)

func numBytes(n uint32) byte {
	if n <= 0xFF {
		return 1
	}
	if n <= 0xFFFF {
		return 2
	}
	return 4
}

func appendNum(dst []byte, num uint32, ibytes byte) []byte {
	switch ibytes {
	case 1:
		dst = append(dst, byte(num))
	case 2:
		dst = append(dst, 0, 0)
		binary.LittleEndian.PutUint16(dst[len(dst)-2:], uint16(num))
	default:
		dst = append(dst, 0, 0, 0, 0)
		binary.LittleEndian.PutUint32(dst[len(dst)-4:], uint32(num))
	}
	return dst
}

func readNum(data []byte, ibytes byte) uint32 {
	switch ibytes {
	case 1:
		return uint32(data[0])
	case 2:
		return uint32(binary.LittleEndian.Uint16(data))
	default:
		return binary.LittleEndian.Uint32(data)
	}
}

type RRect struct {
	Data     interface{}
	Min, Max [rDims]float64
}

type RNode struct {
	count int
	rects [rMaxEntries + 1]RRect
}

type RTree struct {
	height   int
	root     RRect
	count    int
	Reinsert []RRect
}

func (r *RRect) expand(b *RRect) {
	for i := 0; i < rDims; i++ {
		if b.Min[i] < r.Min[i] {
			r.Min[i] = b.Min[i]
		}
		if b.Max[i] > r.Max[i] {
			r.Max[i] = b.Max[i]
		}
	}
}

// Insert inserts an item into the RTree
func (tr *RTree) Insert(min, max []float64, value interface{}) {
	var item RRect
	fit(min, max, value, &item)
	tr.insert(&item)
}

func (tr *RTree) insert(item *RRect) {
	if tr.root.Data == nil {
		fit(item.Min[:], item.Max[:], new(RNode), &tr.root)
	}
	grown := tr.root.insert(item, tr.height)
	if grown {
		tr.root.expand(item)
	}
	if tr.root.Data.(*RNode).count == rMaxEntries+1 {
		newRoot := new(RNode)
		tr.root.splitLargestAxisEdgeSnap(&newRoot.rects[1])
		newRoot.rects[0] = tr.root
		newRoot.count = 2
		tr.root.Data = newRoot
		tr.root.recalc()
		tr.height++
	}
	tr.count++
}

func (r *RRect) chooseLeastEnlargement(b *RRect) int {
	j, jenlargement, jarea := -1, 0.0, 0.0
	n := r.Data.(*RNode)
	for i := 0; i < n.count; i++ {
		// force inline
		area := n.rects[i].Max[0] - n.rects[i].Min[0]
		for j := 1; j < rDims; j++ {
			area *= n.rects[i].Max[j] - n.rects[i].Min[j]
		}
		var enlargement float64
		// force inline
		enlargedArea := 1.0
		for j := 0; j < len(n.rects[i].Min); j++ {
			if b.Max[j] > n.rects[i].Max[j] {
				if b.Min[j] < n.rects[i].Min[j] {
					enlargedArea *= b.Max[j] - b.Min[j]
				} else {
					enlargedArea *= b.Max[j] - n.rects[i].Min[j]
				}
			} else {
				if b.Min[j] < n.rects[i].Min[j] {
					enlargedArea *= n.rects[i].Max[j] - b.Min[j]
				} else {
					enlargedArea *= n.rects[i].Max[j] - n.rects[i].Min[j]
				}
			}
		}
		enlargement = enlargedArea - area

		if j == -1 || enlargement < jenlargement {
			j, jenlargement, jarea = i, enlargement, area
		} else if enlargement == jenlargement {
			if area < jarea {
				j, jenlargement, jarea = i, enlargement, area
			}
		}
	}
	return j
}

func (r *RRect) recalc() {
	n := r.Data.(*RNode)
	r.Min = n.rects[0].Min
	r.Max = n.rects[0].Max
	for i := 1; i < n.count; i++ {
		r.expand(&n.rects[i])
	}
}

// contains return struct when b is fully contained inside of n
func (r *RRect) contains(b *RRect) bool {
	for i := 0; i < rDims; i++ {
		if b.Min[i] < r.Min[i] || b.Max[i] > r.Max[i] {
			return false
		}
	}
	return true
}

func (r *RRect) largestAxis() (axis int, size float64) {
	j, jsz := 0, 0.0
	for i := 0; i < rDims; i++ {
		sz := r.Max[i] - r.Min[i]
		if i == 0 || sz > jsz {
			j, jsz = i, sz
		}
	}
	return j, jsz
}

func (r *RRect) splitLargestAxisEdgeSnap(right *RRect) {
	axis, _ := r.largestAxis()
	left := r
	leftNode := left.Data.(*RNode)
	rightNode := new(RNode)
	right.Data = rightNode

	var equals []RRect
	for i := 0; i < leftNode.count; i++ {
		minDist := leftNode.rects[i].Min[axis] - left.Min[axis]
		maxDist := left.Max[axis] - leftNode.rects[i].Max[axis]
		if minDist < maxDist {
			// stay left
		} else {
			if minDist > maxDist {
				// move to right
				rightNode.rects[rightNode.count] = leftNode.rects[i]
				rightNode.count++
			} else {
				// move to equals, at the end of the left array
				equals = append(equals, leftNode.rects[i])
			}
			leftNode.rects[i] = leftNode.rects[leftNode.count-1]
			leftNode.rects[leftNode.count-1].Data = nil
			leftNode.count--
			i--
		}
	}
	for _, b := range equals {
		if leftNode.count < rightNode.count {
			leftNode.rects[leftNode.count] = b
			leftNode.count++
		} else {
			rightNode.rects[rightNode.count] = b
			rightNode.count++
		}
	}
	left.recalc()
	right.recalc()
}

func (r *RRect) insert(item *RRect, height int) (grown bool) {
	n := r.Data.(*RNode)
	if height == 0 {
		n.rects[n.count] = *item
		n.count++
		grown = !r.contains(item)
		return grown
	}
	// choose subtree
	index := r.chooseLeastEnlargement(item)
	child := &n.rects[index]
	grown = child.insert(item, height-1)
	if grown {
		child.expand(item)
		grown = !r.contains(item)
	}
	if child.Data.(*RNode).count == rMaxEntries+1 {
		child.splitLargestAxisEdgeSnap(&n.rects[n.count])
		n.count++
	}
	return grown
}

// fit an external item into a rect type
func fit(min, max []float64, value interface{}, target *RRect) {
	if max == nil {
		max = min
	}
	if len(min) != len(max) {
		panic("min/max dimension mismatch")
	}
	if len(min) != rDims {
		panic("invalid number of dimensions")
	}
	for i := 0; i < rDims; i++ {
		target.Min[i] = min[i]
		target.Max[i] = max[i]
	}
	target.Data = value
}

func (r *RRect) intersects(b *RRect) bool {
	for i := 0; i < rDims; i++ {
		if b.Min[i] > r.Max[i] || b.Max[i] < r.Min[i] {
			return false
		}
	}
	return true
}

func (r *RRect) search(
	target *RRect, height int,
	iter func(min, max []float64, value interface{}) bool,
) bool {
	n := r.Data.(*RNode)
	if height == 0 {
		for i := 0; i < n.count; i++ {
			if target.intersects(&n.rects[i]) {
				if !iter(n.rects[i].Min[:], n.rects[i].Max[:],
					n.rects[i].Data) {
					return false
				}
			}
		}
	} else {
		for i := 0; i < n.count; i++ {
			if target.intersects(&n.rects[i]) {
				if !n.rects[i].search(target, height-1, iter) {
					return false
				}
			}
		}
	}
	return true
}

func (tr *RTree) search(
	target *RRect,
	iter func(min, max []float64, value interface{}) bool,
) {
	if tr.root.Data == nil {
		return
	}
	if target.intersects(&tr.root) {
		tr.root.search(target, tr.height, iter)
	}
}

func (tr *RTree) Search(
	min, max []float64,
	iter func(min, max []float64, value interface{}) bool,
) {
	var target RRect
	fit(min, max, nil, &target)
	tr.search(&target, iter)
}

func appendFloat(dst []byte, num float64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(num))
	return append(dst, buf[:]...)
}

func (tr *RTree) Compress(dst []byte) []byte {
	if tr.root.Data == nil {
		return dst
	}
	dst = append(dst, byte(tr.height))
	return tr.root.Compress(dst, tr.height)
}

func (r *RRect) Compress(dst []byte, height int) []byte {
	n := r.Data.(*RNode)
	dst = appendFloat(dst, r.Min[0])
	dst = appendFloat(dst, r.Min[1])
	dst = appendFloat(dst, r.Max[0])
	dst = appendFloat(dst, r.Max[1])
	dst = append(dst, byte(n.count))
	if height == 0 {
		var ibytes byte = 1
		for i := 0; i < n.count; i++ {
			ibytes2 := numBytes(uint32(n.rects[i].Data.(int)))
			if ibytes2 > ibytes {
				ibytes = ibytes2
			}
		}
		dst = append(dst, ibytes)
		for i := 0; i < n.count; i++ {
			dst = appendNum(dst, uint32(n.rects[i].Data.(int)), ibytes)
		}
		return dst
	}
	mark := make([]int, n.count)
	for i := 0; i < n.count; i++ {
		mark[i] = len(dst)
		dst = append(dst, 0, 0, 0, 0)
	}
	for i := 0; i < n.count; i++ {
		binary.LittleEndian.PutUint32(dst[mark[i]:], uint32(len(dst)))
		dst = n.rects[i].Compress(dst, height-1)
	}
	return dst
}

func RCompressSearch(
	data []byte,
	addr int,
	series *BaseSeries,
	rect Rect,
	iter func(seg Segment, item int) bool,
) bool {
	if int(addr) == len(data) {
		return true
	}
	height := int(data[addr])
	addr++
	return RnCompressSearch(data, addr, series, rect, height, iter)
}

func RnCompressSearch(
	data []byte,
	addr int,
	series *BaseSeries,
	rect Rect,
	height int,
	iter func(seg Segment, item int) bool,
) bool {
	var nrect Rect
	nrect.Min.X = math.Float64frombits(binary.LittleEndian.Uint64(data[addr:]))
	addr += 8
	nrect.Min.Y = math.Float64frombits(binary.LittleEndian.Uint64(data[addr:]))
	addr += 8
	nrect.Max.X = math.Float64frombits(binary.LittleEndian.Uint64(data[addr:]))
	addr += 8
	nrect.Max.Y = math.Float64frombits(binary.LittleEndian.Uint64(data[addr:]))
	addr += 8
	if !rect.IntersectsRect(nrect) {
		return true
	}
	count := int(data[addr])
	addr++
	if height == 0 {
		ibytes := data[addr]
		addr++
		for i := 0; i < count; i++ {
			item := int(readNum(data[addr:], ibytes))
			addr += int(ibytes)
			seg := series.SegmentAt(int(item))
			irect := seg.Rect()
			if irect.IntersectsRect(rect) {
				if !iter(seg, int(item)) {
					return false
				}
			}
		}
		return true
	}
	for i := 0; i < count; i++ {
		naddr := int(binary.LittleEndian.Uint32(data[addr:]))
		addr += 4
		if !RnCompressSearch(data, naddr, series, rect, height-1, iter) {
			return false
		}
	}
	return true
}
