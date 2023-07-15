// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rtreecompressdemo

import (
	"encoding/binary"
	"math"
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

func appendFloat(dst []byte, num float64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(num))
	return append(dst, buf[:]...)
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
