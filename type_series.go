// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

// // Series is just a series of points with utilities for efficiently accessing
// // segments from rectangle queries, making stuff like point-in-polygon lookups
// // very quick.
// type Series interface {
// 	Rect() Rect
// 	Empty() bool
// 	Convex() bool
// 	Clockwise() bool
// 	NumPoints() int
// 	NumSegments() int
// 	PointAt(index int) Point
// 	SegmentAt(index int) Segment
// 	Search(rect Rect, iter func(seg Segment, index int) bool)
// 	Index() interface{}
// 	Valid() bool
// }

// BaseSeries is a concrete type containing all that is needed to make a Series.
type BaseSeries struct {
	points []Point // original points
}

func (series *BaseSeries) SegmentAt(index int) Segment {
	var seg Segment
	seg.A = series.points[index]
	if index == len(series.points)-1 {
		seg.B = series.points[0]
	} else {
		seg.B = series.points[index+1]
	}
	return seg
}
