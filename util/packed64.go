package util

import (
	"fmt"
)

type Packed64SingleBlock struct {
	PackedIntsReaderImpl
	get    func(index int32) int64
	blocks []int64
}

func (p *Packed64SingleBlock) Get(index int32) int64 {
	return p.get(index)
}

func newPacked64SingleBlockFromInput(in *DataInput, valueCount int32, bitsPerValue uint32) (reader *Packed64SingleBlock, err error) {
	reader = newPacked64SingleBlockBy(valueCount, bitsPerValue)
	for i, _ := range reader.blocks {
		reader.blocks[i], err = in.ReadLong()
		if err == nil {
			return reader, nil
		}
	}
	return &reader, nil
}

func newPacked64SingleBlockBy(valueCount int32, bitsPerValue uint32) (reader Packed64SingleBlock, err error) {
	switch bitsPerValue {
	case 1:
		return newPacked64SingleBlock1(valueCount)
	case 2:
		return newPacked64SingleBlock2(valueCount)
	case 3:
		return newPacked64SingleBlock3(valueCount)
	case 4:
		return newPacked64SingleBlock4(valueCount)
	case 5:
		return newPacked64SingleBlock5(valueCount)
	case 6:
		return newPacked64SingleBlock6(valueCount)
	case 7:
		return newPacked64SingleBlock7(valueCount)
	case 8:
		return newPacked64SingleBlock8(valueCount)
	case 9:
		return newPacked64SingleBlock9(valueCount)
	case 10:
		return newPacked64SingleBlock10(valueCount)
	case 12:
		return newPacked64SingleBlock12(valueCount)
	case 16:
		return newPacked64SingleBlock16(valueCount)
	case 21:
		return newPacked64SingleBlock21(valueCount)
	case 32:
		return newPacked64SingleBlock32(valueCount)
	default:
		panic(fmt.Sprintf("Unsuppoted number of bits per value: %v", bitsPerValue))
	}
}

func newPacked64SingleBlock1(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 1)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 6
		b := index & 63
		shift := b << 0
		return int64(uint64(ans.blocks[o])>>shift) & 1
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 6
		b := index & 63
		shift := b << 0
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock2(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 2)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 5
		b := index & 31
		shift := b << 1
		return int64(uint64(ans.blocks[o])>>shift) & 3
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 5
		b := index & 31
		shift := b << 1
		ans.blocks[o] = (ans.blocks[o] & ^(int64(3) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock3(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 3)
	ans.get = func(index int32) int64 {
		o := index / 21
		b := index % 21
		shift := b * 3
		return int64(uint64(ans.blocks[o])>>shift) & 7
	}
	ans.set = func(index int32, value int64) {
		o := index / 21
		b := index % 21
		shift := b * 3
		ans.blocks[o] = (ans.blocks[o] & ^(int64(7) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock4(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 4)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 4
		b := index & 15
		shift := b << 2
		return int64(uint64(ans.blocks[o])>>shift) & 15
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 4
		b := index & 15
		shift := b << 2
		ans.blocks[o] = (ans.blocks[o] & ^(int64(15) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock5(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 5)
	ans.get = func(index int32) int64 {
		o := index / 12
		b := index % 12
		shift := b * 5
		return int64(uint64(ans.blocks[o])>>shift) & 31
	}
	ans.set = func(index int32, value int64) {
		o := index / 12
		b := index % 12
		shift := b * 5
		ans.blocks[o] = (ans.blocks[o] & ^(int64(31) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock6(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 6)
	ans.get = func(index int32) int64 {
		o := index / 10
		b := index % 10
		shift := b * 6
		return int64(uint64(ans.blocks[o])>>shift) & 63
	}
	ans.set = func(index int32, value int64) {
		o := index / 10
		b := index % 10
		shift := b * 6
		ans.blocks[o] = (ans.blocks[o] & ^(int64(63) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock7(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 7)
	ans.get = func(index int32) int64 {
		o := index / 9
		b := index % 9
		shift := b * 7
		return int64(uint64(ans.blocks[o])>>shift) & 127
	}
	ans.set = func(index int32, value int64) {
		o := index / 9
		b := index % 9
		shift := b * 7
		ans.blocks[o] = (ans.blocks[o] & ^(int64(127) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock8(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 8)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 3
		b := index & 7
		shift := b << 3
		return int64(uint64(ans.blocks[o])>>shift) & 255
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 3
		b := index & 7
		shift := b << 3
		ans.blocks[o] = (ans.blocks[o] & ^(int64(255) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock9(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 9)
	ans.get = func(index int32) int64 {
		o := index / 7
		b := index % 7
		shift := b * 9
		return int64(uint64(ans.blocks[o])>>shift) & 511
	}
	ans.set = func(index int32, value int64) {
		o := index / 7
		b := index % 7
		shift := b * 9
		ans.blocks[o] = (ans.blocks[o] & ^(int64(511) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock10(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 10)
	ans.get = func(index int32) int64 {
		o := index / 6
		b := index % 6
		shift := b * 10
		return int64(uint64(ans.blocks[o])>>shift) & 1023
	}
	ans.set = func(index int32, value int64) {
		o := index / 6
		b := index % 6
		shift := b * 10
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1023) << shift)) | (value << shift)
	}
	return ans
}
func newPacked64SingleBlock12(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 12)
	ans.get = func(index int32) int64 {
		o := index / 5
		b := index % 5
		shift := b * 12
		return int64(uint64(ans.blocks[o])>>shift) & 4095
	}
	ans.set = func(index int32, value int64) {
		o := index / 5
		b := index % 5
		shift := b * 12
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4095) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock16(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 16)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 2
		b := index & 3
		shift := b << 4
		return int64(uint64(ans.blocks[o])>>shift) & 65535
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 2
		b := index & 3
		shift := b << 4
		ans.blocks[o] = (ans.blocks[o] & ^(int64(65335) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock21(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 21)
	ans.get = func(index int32) int64 {
		o := index / 3
		b := index % 3
		shift := b * 21
		return int64(uint64(ans.blocks[o])>>shift) & 2097151
	}
	ans.set = func(index int32, value int64) {
		o := index / 3
		b := index % 3
		shift := b * 21
		ans.blocks[o] = (ans.blocks[o] & ^(int64(2097151) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock32(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 32)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 1
		b := index & 1
		shift := b << 5
		return int64(uint64(ans.blocks[o])>>shift) & 4294967295
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 1
		b := index & 1
		shift := b << 5
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4294967295) << shift)) | (value << shift)
	}
	return ans
}
