package compressing

import (
	"github.com/balzaczyy/golucene/core/util/packed"
)

// codecs/compressing/LZ4.java#compress

const (
	MEMORY_USAGE = 14
)

type DataOutput interface {
	WriteByte(b byte) error
	WriteBytes(buf []byte) error
	WriteInt(i int32) error
	WriteVInt(i int32) error
	WriteString(string) error
}

func hash(i, hashBits int) int {
	assert(hashBits >= 0 && hashBits <= 32)
	return int(uint32(int32(i)*-1640531535) >> uint(32-hashBits))
}

func readInt(buf []byte, i int) int {
	return ((int(buf[i]) & 0xFF) << 24) | ((int(buf[i+1]) & 0xFF) << 16) |
		((int(buf[i+2]) & 0xFF) << 8) | (int(buf[i+3]) & 0xFF)
}

func commonBytes(b1, b2 []byte) int {
	assert(len(b1) > len(b2))
	count, limit := 0, len(b2)
	for count < limit && b1[count] == b2[count] {
		count++
	}
	return count
}

func encodeLen(l int, out DataOutput) error {
	var err error
	for l >= 0xFF && err == nil {
		err = out.WriteByte(byte(0xFF))
		l -= 0xFF
	}
	if err == nil {
		err = out.WriteByte(byte(l))
	}
	return err
}

func encodeLiterals(bytes []byte, token byte, out DataOutput) error {
	err := out.WriteByte(token)
	if err != nil {
		return err
	}

	// encode literal length
	if len(bytes) >= 0x0F {
		err = encodeLen(len(bytes)-0x0F, out)
		if err != nil {
			return err
		}
	}

	// encode literals
	return out.WriteBytes(bytes)
}

func encodeLastLiterals(bytes []byte, out DataOutput) error {
	token := byte(min(len(bytes), 0x0F) << 4)
	return encodeLiterals(bytes, token, out)
}

func encodeSequence(bytes []byte, matchDec, matchLen int, out DataOutput) error {
	literalLen := len(bytes)
	assert(matchLen >= 4)
	// encode token
	token := byte((min(literalLen, 0x0F) << 4) | min(matchLen-4, 0x0F))
	err := encodeLiterals(bytes, token, out)
	if err != nil {
		return err
	}

	// encode match dec
	assert(matchDec > 0 && matchDec < 1<<16)
	err = out.WriteByte(byte(matchDec))
	if err == nil {
		err = out.WriteByte(byte(uint(matchDec) >> 8))
	}
	if err != nil {
		return err
	}

	// encode match len
	if matchLen >= MIN_MATCH+0x0F {
		return encodeLen(matchLen-0x0F-MIN_MATCH, out)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type LZ4HashTable struct {
	hashLog   int
	hashTable packed.Mutable
}

func (h *LZ4HashTable) reset(length int) {
	bitsPerOffset := packed.BitsRequired(int64(length - LAST_LITERALS))
	bitsPerOffsetLog := ceilLog2(bitsPerOffset)
	h.hashLog = MEMORY_USAGE + 3 - bitsPerOffsetLog
	assert(h.hashLog > 0)
	if h.hashTable == nil || h.hashTable.Size() < (1<<uint(h.hashLog)) || h.hashTable.BitsPerValue() < bitsPerOffset {
		h.hashTable = packed.MutableFor(1<<uint(h.hashLog), bitsPerOffset, packed.PackedInts.DEFAULT)
	} else {
		h.hashTable.Clear()
	}
}

// 32 - leadingZero(n-1)
func ceilLog2(n int) int {
	assert(n >= 1)
	if n == 1 {
		return 0
	}
	n--
	ans := 0
	for n > 0 {
		n >>= 1
		ans++
	}
	return ans
}

/*
Compress bytes into out using at most 16KB of memory. ht shouldn't be
shared across threads but can safely be reused.
*/
func LZ4Compress(bytes []byte, out DataOutput, ht *LZ4HashTable) error {
	offset, length := 0, len(bytes)
	base, end := offset, offset+length

	anchor := offset
	offset++

	if length > LAST_LITERALS+MIN_MATCH {
		limit := end - LAST_LITERALS
		matchLimit := limit - MIN_MATCH
		ht.reset(length)
		hashLog := ht.hashLog
		hashTable := ht.hashTable

		for offset <= limit {
			// find a match
			var ref int
			var hasMore = offset < matchLimit
			for hasMore {
				v := readInt(bytes, offset)
				h := hash(v, hashLog)
				ref = base + int(hashTable.Get(h))
				assert(packed.BitsRequired(int64(offset-base)) <= hashTable.BitsPerValue())
				hashTable.Set(h, int64(offset-base))
				if offset-ref < MAX_DISTANCE && readInt(bytes, ref) == v {
					break
				}
				offset++
				hasMore = offset < matchLimit
			}
			if !hasMore {
				break
			}

			// compute match length
			matchLen := MIN_MATCH + commonBytes(
				bytes[ref+MIN_MATCH:limit],
				bytes[offset+MIN_MATCH:limit])

			err := encodeSequence(bytes[anchor:offset], offset-ref, matchLen, out)
			if err != nil {
				return err
			}
			offset += matchLen
			anchor = offset
		}
	}

	// last literals
	literalLen := end - anchor
	assert(literalLen >= LAST_LITERALS || literalLen == length)
	return encodeLastLiterals(bytes[anchor:], out)
}
