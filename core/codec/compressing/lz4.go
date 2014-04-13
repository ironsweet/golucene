package compressing

// codec/compressing/LZ4.java

const (
	MIN_MATCH     = 4       // minimum length of a match
	MAX_DISTANCE  = 1 << 16 // maximum distance of a reference
	LAST_LITERALS = 5       // the last 5 bytes mus tbe encoded as literals
)

type DataInput interface {
	ReadByte() (b byte, err error)
	ReadBytes(buf []byte) error
}

/*
Decompress at least decompressedLen bytes into dest[]. Please
note that dest[] must be large enough to be able to hold all
decompressed data (meaning that you need to know the total
decompressed length)
*/
func LZ4Decompress(compressed DataInput, decompressedLen int, dest []byte) (length int, err error) {
	dOff, destEnd := 0, len(dest)

	for {
		// literals
		var token int
		token, err = asInt(compressed.ReadByte())
		if literalLen := int(uint(token) >> 4); literalLen != 0 {
			if literalLen == 0x0F {
				var b byte = 0xFF
				for b == 0XFF {
					b, err = compressed.ReadByte()
					if err != nil {
						return
					}
					literalLen += int(uint8(b))
				}
			}
			err = compressed.ReadBytes(dest[dOff : dOff+literalLen])
			if err != nil {
				return
			}
			dOff += literalLen
		}

		if dOff >= decompressedLen {
			break
		}

		// matches
		var matchDec int
		matchDec, err = asInt(compressed.ReadByte())
		if err != nil {
			return
		}
		if b, err := asInt(compressed.ReadByte()); err == nil {
			matchDec = matchDec | (b << 8)
		} else {
			return 0, err
		}
		assert(matchDec > 0)

		matchLen := token & 0x0F
		if matchLen == 0x0F {
			var b byte = 0xFF
			for b == 0xFF {
				b, err = compressed.ReadByte()
				if err != nil {
					return
				}
				matchLen += int(b)
			}
		}
		matchLen += MIN_MATCH

		// copying a multiple of 8 bytes can make decompression from 5% to 10% faster
		fastLen := int((int64(matchLen) + 7) & 0xFFFFFFF8)
		if matchDec < matchLen || dOff+fastLen > destEnd {
			// overlap -> naive incremental copy
			for ref, end := dOff-matchDec, dOff+matchLen; dOff < end; {
				dest[dOff] = dest[ref]
				ref++
				dOff++
			}
		} else {
			// no overlap -> arraycopy
			copy(dest[dOff:], dest[dOff-matchDec:dOff-matchDec+fastLen])
			dOff += matchLen
		}

		if dOff >= decompressedLen {
			break
		}
	}

	return dOff, nil
}

func asInt(b byte, err error) (n int, err2 error) {
	return int(b), err
}
