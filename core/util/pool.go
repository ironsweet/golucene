package util

// util/ByteBlockPool.java

/*
Class that Posting and PostingVector use to write byte streams into
shared fixed-size []byte arrays. The idea is to allocate slices of
increasing lengths. For example, the first slice is 5 bytes, the next
slice is 14, etc. We start by writing our bytes into the first 5
bytes. When we hit the end of the slice, we allocate the next slice
and then write the address of the new slice into the last 4 bytes of
the previous slice (the "forwarding address").

Each slice is filled with 0's initially, and we mark the end with a
non-zero byte. This way the methods that are writing into the slice
don't need to record its length and instead allocate a new slice once
they hit a non-zero byte.
*/

const (
	BYTE_BLOCK_SHIFT = 15
	BYTE_BLOCK_SIZE  = 1 << BYTE_BLOCK_SHIFT
	BYTE_BLOCK_MASK  = BYTE_BLOCK_SIZE - 1
)
