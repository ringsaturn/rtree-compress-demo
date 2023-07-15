// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

import "encoding/binary"

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
			// Data is an original series' segment index:
			// https://github.com/tidwall/geojson/blob/v1.4.3/geometry/series.go#L316-L318
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
