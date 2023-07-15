// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

import (
	"encoding/binary"
	"math"
)

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
