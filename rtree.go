// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

const (
	rDims       = 2
	rMaxEntries = 16
)

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
