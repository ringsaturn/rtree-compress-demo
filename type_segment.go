// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

// Segment is a two point line
type Segment struct {
	A, B Point
}

// Rect is the outer boundaries of the segment.
func (seg Segment) Rect() Rect {
	var rect Rect
	rect.Min = seg.A
	rect.Max = seg.B
	if rect.Min.X > rect.Max.X {
		rect.Min.X, rect.Max.X = rect.Max.X, rect.Min.X
	}
	if rect.Min.Y > rect.Max.Y {
		rect.Min.Y, rect.Max.Y = rect.Max.Y, rect.Min.Y
	}
	return rect
}
