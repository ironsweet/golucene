package packed

import (
	"fmt"
)

// util/packed/BulkOperation.java

// Efficient sequential read/write of packed integers.
type BulkOperation interface {
	LongBlockCount() int
	LongValueCount() int
	ByteBlockCount() int
	ByteValueCount() int

	// PackedIntsEncoder
	encodeLongToByte(values []int64, blocks []byte, iterations int)
	encodeLongToLong(values, blocks []int64, iterations int)
	EncodeIntToByte(values []int, blocks []byte, iterations int)

	// PackedIntsDecoder
	decodeLongToLong(blocks, values []int64, iterations int)
	decodeByteToLong(blocks []byte, values []int64, iterations int)
	/*
		For every number of bits per value, there is a minumum number of
		blocks (b) / values (v) you need to write an order to reach the next block
		boundary:
		- 16 bits per value -> b=2, v=1
		- 24 bits per value -> b=3, v=1
		- 50 bits per value -> b=25, v=4
		- 63 bits per value -> b=63, v=8
		- ...

		A bulk read consists in copying iterations*v vlaues that are contained in
		iterations*b blocks into a []int64 (higher values of iterations are likely to
		yield a better throughput) => this requires n * (b + 8v) bytes of memory.

		This method computes iterations as ramBudget / (b + 8v) (since an int64 is
		8 bytes).
	*/
	computeIterations(valueCount, ramBudget int) int
}

var (
	packedBulkOps = []BulkOperation{
		/*[[[gocog
		package main

		import (
			"fmt"
			"io"
			"os"
		)

		const (
			MAX_SPECIALIZED_BITS_PER_VALUE = 24
			HEADER                         = `// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.`
		)

		func isPowerOfTwo(n int) bool {
			return n&(n-1) == 0
		}

		func casts(typ string) (castStart, castEnd string) {
			if typ == "int64" {
				return "", ""
			}
			return fmt.Sprintf("%s(", typ), ")"
		}

		func masks(bits int) (start, end string) {
			if bits == 64 {
				return "", ""
			}
			return "(", fmt.Sprintf(" & %x)", (1<<uint(bits))-1)
		}

		var (
			TYPES = map[int]string{8: "byte", 16: "int16", 32: "int32", 64: "int64"}
			NAMES = map[int]string{8: "Byte", 16: "Short", 32: "Int", 64: "Long"}
		)

		func blockValueCount(bpv, bits int) (blocks, values int) {
			blocks = bpv
			values = blocks * bits / bpv
			for blocks%2 == 0 && values%2 == 0 {
				blocks /= 2
				values /= 2
			}
			assert2(values*bpv == bits*blocks, fmt.Sprintf("%d values, %d blocks, %d bits per value", values, blocks, bpv))
			return blocks, values
		}

		func assert2(ok bool, msg string) {
			if !ok {
				panic(msg)
			}
		}

		func packed64(bpv int, f io.Writer) {
			if bpv == 64 {
				panic("not implemented yet")
			} else {
				p64Decode(bpv, f, 32)
				p64Decode(bpv, f, 64)
			}
		}

		func p64Decode(bpv int, f io.Writer, bits int) {
			_, values := blockValueCount(bpv, 64)
			typ := TYPES[bits]
			castStart, castEnd := casts(typ)
			var mask uint

			fmt.Fprintf(f, "func (op *BulkOperationPacked%d) decodeLongTo%s(blocks []int64, values []%s, iterations int) {\n", bpv, NAMES[bits], typ)
			if bits < bpv {
				fmt.Fprintln(f, "	panic(\"not supported yet\")")
			} else {
				fmt.Fprintln(f, "	blocksOffset, valuesOffset := 0, 0")
				fmt.Fprintf(f, "	for i := 0; i < iterations; i ++ {\n")
				mask = 1<<uint(bpv) - 1

				if isPowerOfTwo(bpv) {
					fmt.Fprintln(f, "		block := blocks[blocksOffset]; blocksOffset++")
					fmt.Fprintf(f, "		for shift := uint(%d); shift >= 0; shift -= %d {\n", 64-bpv, bpv)
					fmt.Fprintf(f, "			values[valuesOffset] = %s(int64(uint64(block) >> shift)) & %d%s; valuesOffset++\n", castStart, mask, castEnd)
					fmt.Fprintln(f, "		}")
				} else {
					for i := 0; i < values; i++ {
						blockOffset := i * bpv / 64
						bitOffset := (i * bpv) % 64
						if bitOffset == 0 {
							// start of block
							fmt.Fprintf(f, "		block%d := blocks[blocksOffset]; blocksOffset++\n", blockOffset)
							fmt.Fprintf(f, "		values[valuesOffset] = %sint64(uint64(block%d) >> %d%s); valuesOffset++\n", castStart, blockOffset, 64-bpv, castEnd)
						} else if bitOffset+bpv == 64 {
							// end of block
							fmt.Fprintf(f, "		values[valuesOffset] = %sblock%d & %d%s; valuesOffset++\n", castStart, blockOffset, mask, castEnd)
						} else if bitOffset+bpv < 64 {
							// middle of block
							fmt.Fprintf(f, "		values[valuesOffset] = %sint64(uint64(block%d) >> %d) & %d%s; valuesOffset++\n", castStart, blockOffset, 64-bitOffset-bpv, mask, castEnd)
						} else {
							// value spans across 2 blocks
							mask1 := int(1<<uint(64-bitOffset)) - 1
							shift1 := bitOffset + bpv - 64
							shift2 := 64 - shift1
							fmt.Fprintf(f, "		block%d := blocks[blocksOffset]; blocksOffset++\n", blockOffset+1)
							fmt.Fprintf(f, "		values[valuesOffset] = %s((block%d & %d) << %d) | (int64(uint64(block%d) >> %d))%s; valuesOffset++\n",
								castStart, blockOffset, mask1, shift1, blockOffset+1, shift2, castEnd)
						}
					}
				}
				fmt.Fprintln(f, "	}")
			}
			fmt.Fprintln(f, "}\n")

			_, byteValues := blockValueCount(bpv, 8)

			fmt.Fprintf(f, "func (op *BulkOperationPacked%d) decodeByteTo%s(blocks []byte, values []%s, iterations int) {\n", bpv, NAMES[bits], typ)
			if bits < bpv {
				fmt.Fprintln(f, "	panic(\"not supported yet\")")
			} else {
				fmt.Fprintln(f, "	blocksOffset, valuesOffset := 0, 0")
				if isPowerOfTwo(bpv) && bpv < 8 {
					fmt.Fprintf(f, "	for j := 0; j < iterations; j ++ {\n")
					fmt.Fprintf(f, "		block := blocks[blocksOffset]\n")
					fmt.Fprintln(f, "		blocksOffset++")
					for shift := 8 - bpv; shift > 0; shift -= bpv {
						fmt.Fprintf(f, "		values[valuesOffset] = %s(byte(uint8(block)) >> %d) & %d\n", typ, shift, mask)
						fmt.Fprintln(f, "		valuesOffset++")
					}
					fmt.Fprintf(f, "		values[valuesOffset] = %s(block & %d)\n", typ, mask)
					fmt.Fprintln(f, "		valuesOffset++")
					fmt.Fprintln(f, "	}")
				} else if bpv == 8 {
					fmt.Fprintln(f, "	for j := 0; j < iterations; j ++ {")
					fmt.Fprintf(f, "		values[valuesOffset] = %s(blocks[blocksOffset]); valuesOffset++; blocksOffset++\n", typ)
					fmt.Fprintln(f, "	}")
				} else if isPowerOfTwo(bpv) && bpv > 8 {
					fmt.Fprintf(f, "	for j := 0; j < iterations; j ++ {\n")
					m := "int32"
					if bits > 32 {
						m = "int64"
					}
					fmt.Fprintf(f, "		values[valuesOffset] =")
					for i, until := 0, bpv/8-1; i < until; i++ {
						fmt.Fprintf(f, " (%s(blocks[blocksOffset+%d]) << %d) |", m, i, bpv-8)
					}
					fmt.Fprintf(f, " %s(blocks[blocksOffset+%d])\n", m, bpv/8-1)
					fmt.Fprintln(f, "		valuesOffset++")
					fmt.Fprintf(f, "		blocksOffset += %d\n", bpv/8)
					fmt.Fprintln(f, "	}")
				} else {
					fmt.Fprintf(f, "	for i := 0; i < iterations; i ++ {\n")
					for i := 0; i < byteValues; i++ {
						byteStart, byteEnd := i*bpv/8, ((i+1)*bpv-1)/8
						bitStart, bitEnd := (i*bpv)%8, ((i+1)*bpv-1)%8
						shift := func(b int) int { return 8*(byteEnd-b-1) + 1 + bitEnd }
						if bitStart == 0 {
							fmt.Fprintf(f, "		byte%d := blocks[blocksOffset]\n", byteStart)
							fmt.Fprintln(f, "		blocksOffset++")
						}
						for b, until := byteStart+1, byteEnd+1; b < until; b++ {
							fmt.Fprintf(f, "		byte%d := blocks[blocksOffset]\n", b)
							fmt.Fprintln(f, "		blocksOffset++")
						}
						fmt.Fprintf(f, "		values[valuesOffset] = %s(", typ)
						if byteStart == byteEnd {
							if bitStart == 0 {
								if bitEnd == 7 {
									fmt.Fprintf(f, " int64(byte%d)", byteStart)
								} else {
									fmt.Fprintf(f, " int64(uint8(byte%d) >> %d)", byteStart, 7-bitEnd)
								}
							} else {
								if bitEnd == 7 {
									fmt.Fprintf(f, " int64(byte%d) & %d", byteStart, 1<<uint(8-bitStart)-1)
								} else {
									fmt.Fprintf(f, " int64(uint8(byte%d) >> %d) & %d", byteStart, 7-bitEnd, 1<<uint(bitEnd-bitStart+1)-1)
								}
							}
						} else {
							if bitStart == 0 {
								fmt.Fprintf(f, "(int64(byte%d) << %d)", byteStart, shift(byteStart))
							} else {
								fmt.Fprintf(f, "(int64(byte%d & %d) << %d)", byteStart, 1<<uint(8-bitStart)-1, shift(byteStart))
							}
							for b, until := byteStart+1, byteEnd; b < until; b++ {
								fmt.Fprintf(f, " | (int64(byte%d) << %d)", b, shift(b))
							}
							if bitEnd == 7 {
								fmt.Fprintf(f, " | int64(byte%d)", byteEnd)
							} else {
								fmt.Fprintf(f, " | int64(uint8(byte%d) >> %d)", byteEnd, 7-bitEnd)
							}
						}
						fmt.Fprintf(f, ")")
						fmt.Fprintln(f, "")
						fmt.Fprintln(f, "		valuesOffset++")
					}
					fmt.Fprintln(f, "	}")
				}
			}
			fmt.Fprintln(f, "}")
		}

		func main() {
			for bpv := 1; bpv <= 64; bpv++ {
				if bpv > MAX_SPECIALIZED_BITS_PER_VALUE {
					fmt.Printf("		newBulkOperationPacked(%d),\n", bpv)
					continue
				}
				f, err := os.Create(fmt.Sprintf("bulkOperation%d.go", bpv))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				fmt.Fprintf(f, "%v\n", HEADER)
				fmt.Fprintf(f, "type BulkOperationPacked%d struct {\n", bpv)
				fmt.Fprintln(f, "	*BulkOperationPacked")
				fmt.Fprintln(f, "}\n")

				fmt.Fprintf(f, "func newBulkOperationPacked%d() BulkOperation {\n", bpv)
				fmt.Fprintf(f, "	return &BulkOperationPacked%d{newBulkOperationPacked(%d)}\n", bpv, bpv)
				fmt.Fprintln(f, "}\n")

				packed64(bpv, f)

				fmt.Printf("		newBulkOperationPacked%d(),\n", bpv)
			}
		}
				gocog]]]*/
		newBulkOperationPacked1(),
		newBulkOperationPacked2(),
		newBulkOperationPacked3(),
		newBulkOperationPacked4(),
		newBulkOperationPacked5(),
		newBulkOperationPacked6(),
		newBulkOperationPacked7(),
		newBulkOperationPacked8(),
		newBulkOperationPacked9(),
		newBulkOperationPacked10(),
		newBulkOperationPacked11(),
		newBulkOperationPacked12(),
		newBulkOperationPacked13(),
		newBulkOperationPacked14(),
		newBulkOperationPacked15(),
		newBulkOperationPacked16(),
		newBulkOperationPacked17(),
		newBulkOperationPacked18(),
		newBulkOperationPacked19(),
		newBulkOperationPacked20(),
		newBulkOperationPacked21(),
		newBulkOperationPacked22(),
		newBulkOperationPacked23(),
		newBulkOperationPacked24(),
		newBulkOperationPacked(25),
		newBulkOperationPacked(26),
		newBulkOperationPacked(27),
		newBulkOperationPacked(28),
		newBulkOperationPacked(29),
		newBulkOperationPacked(30),
		newBulkOperationPacked(31),
		newBulkOperationPacked(32),
		newBulkOperationPacked(33),
		newBulkOperationPacked(34),
		newBulkOperationPacked(35),
		newBulkOperationPacked(36),
		newBulkOperationPacked(37),
		newBulkOperationPacked(38),
		newBulkOperationPacked(39),
		newBulkOperationPacked(40),
		newBulkOperationPacked(41),
		newBulkOperationPacked(42),
		newBulkOperationPacked(43),
		newBulkOperationPacked(44),
		newBulkOperationPacked(45),
		newBulkOperationPacked(46),
		newBulkOperationPacked(47),
		newBulkOperationPacked(48),
		newBulkOperationPacked(49),
		newBulkOperationPacked(50),
		newBulkOperationPacked(51),
		newBulkOperationPacked(52),
		newBulkOperationPacked(53),
		newBulkOperationPacked(54),
		newBulkOperationPacked(55),
		newBulkOperationPacked(56),
		newBulkOperationPacked(57),
		newBulkOperationPacked(58),
		newBulkOperationPacked(59),
		newBulkOperationPacked(60),
		newBulkOperationPacked(61),
		newBulkOperationPacked(62),
		newBulkOperationPacked(63),
		newBulkOperationPacked(64),
		// [[[end]]]
	}

	packedSingleBlockBulkOps = []BulkOperation{
		/*[[[gocog
			package main
			import "fmt"
		  var PACKED_64_SINGLE_BLOCK_BPV = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 16, 21, 32}
			func main() {
				var bpv int = 1
				for _, v := range PACKED_64_SINGLE_BLOCK_BPV {
					for ;bpv < v; bpv++ {
						fmt.Print("		nil,\n")
					}
					fmt.Printf("		newBulkOperationPackedSingleBlock(%v),\n", bpv)
					bpv++
				}
			}
			gocog]]]*/
		newBulkOperationPackedSingleBlock(1),
		newBulkOperationPackedSingleBlock(2),
		newBulkOperationPackedSingleBlock(3),
		newBulkOperationPackedSingleBlock(4),
		newBulkOperationPackedSingleBlock(5),
		newBulkOperationPackedSingleBlock(6),
		newBulkOperationPackedSingleBlock(7),
		newBulkOperationPackedSingleBlock(8),
		newBulkOperationPackedSingleBlock(9),
		newBulkOperationPackedSingleBlock(10),
		nil,
		newBulkOperationPackedSingleBlock(12),
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(16),
		nil,
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(21),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(32),
		// [[[end]]]
	}
)

func newBulkOperation(format PackedFormat, bitsPerValue uint32) BulkOperation {
	// log.Printf("Initializing BulkOperation(%v,%v)", format, bitsPerValue)
	switch int(format) {
	case PACKED:
		assert2(packedBulkOps[bitsPerValue-1] != nil, fmt.Sprintf("bpv=%v", bitsPerValue))
		return packedBulkOps[bitsPerValue-1]
	case PACKED_SINGLE_BLOCK:
		assert2(packedSingleBlockBulkOps[bitsPerValue-1] != nil, fmt.Sprintf("bpv=%v", bitsPerValue))
		return packedSingleBlockBulkOps[bitsPerValue-1]
	}
	panic(fmt.Sprintf("invalid packed format: %v", format))
}

type BulkOperationImpl struct {
	PackedIntsDecoder
}

func newBulkOperationImpl(decoder PackedIntsDecoder) *BulkOperationImpl {
	return &BulkOperationImpl{decoder}
}

func (op *BulkOperationImpl) writeLong(block int64, blocks []byte) int {
	blocksOffset := 0
	for j := 1; j <= 8; j++ {
		blocks[blocksOffset] = byte(uint64(block) >> uint(64-(j<<3)))
		blocksOffset++
	}
	return blocksOffset
}

func (op *BulkOperationImpl) computeIterations(valueCount, ramBudget int) int {
	iterations := ramBudget / (op.ByteBlockCount() + 8*op.ByteValueCount())
	if iterations == 0 {
		// at least 1
		return 1
	} else if (iterations-1)*op.ByteValueCount() >= valueCount {
		// don't allocate for more than the size of the reader
		panic("not implemented yet")
	} else {
		return iterations
	}
}
