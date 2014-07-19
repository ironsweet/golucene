package lucene41

import (
	"github.com/balzaczyy/golucene/core/codec/compressing"
)

// lucene41/Lucene41StoredFieldsFormat.java

/*
Lucene 4.1 stored fields format.

Principle

This StoredFieldsFormat compresses blocks of 16KB of documents in
order to improve the compression ratio compared to document-level
compression. It uses the LZ4 compression algorithm, which is fast to
compress and very fast to decompress dta. Although the compression
method that is used focuses more on speed than on compression ratio,
it should provide interesting compression ratios for redundant inputs
(such as log files, HTML or plain text).

File formats

Stored fields are represented by two files:

1. field_data

A fields data file (extension .fdt). This file stores a compact
representation of documents in compressed blocks of 16KB or more.
When writing a segment, documents are appended to an in-memory []byte
buffer. When its size reaches 16KB or more, some metadata about the
documents is flushed to disk, immediately followed by a compressed
representation of the buffer using the [LZ4](http://codec.google.com/p/lz4/)
[compression format](http://fastcompression.blogspot.ru/2011/05/lz4-explained.html)

Here is a more detailed description of the field data fiel format:

- FieldData (.dft) --> <Header>, packedIntsVersion, <Chunk>^ChunkCount
- Header --> CodecHeader
- PackedIntsVersion --> PackedInts.VERSION_CURRENT as a VInt
- ChunkCount is not known in advance and is the number of chunks
nucessary to store all document of the segment
- Chunk --> DocBase, ChunkDocs, DocFieldCounts, DocLengths, <CompressedDoc>
- DocBase --> the ID of the first document of the chunk as a VInt
- ChunkDocs --> the number of the documents in the chunk as a VInt
- DocFieldCounts --> the number of stored fields or every document
in the chunk,  encoded as followed:
  - if hunkDocs=1, the unique value is encoded as a VInt
  - else read VInt (let's call it bitsRequired)
    - if bitsRequired is 0 then all values are equal, and the common
    value is the following VInt
    - else bitsRequired is the number of bits required to store any
    value, and values are stored in a packed array where every value
    is stored on exactly bitsRequired bits
- DocLenghts --> the lengths of all documents in the chunk, encodedwith the same method as DocFieldCounts
- CompressedDocs --> a compressed representation of <Docs> using
the LZ4 compression format
- Docs --> <Doc>^ChunkDocs
- Doc --> <FieldNumAndType, Value>^DocFieldCount
- FieldNumAndType --> a VLong, whose 3 last bits are Type and other
bits are FieldNum
- Type -->
  - 0: Value is string
  - 1: Value is BinaryValue
  - 2: Value is int
  - 3: Value is float32
  - 4: Value is int64
  - 5: Value is float64
  - 6, 7: unused
- FieldNum --> an ID of the field
- Value --> string | BinaryValue | int | float32 | int64 | float64
dpending on Type
- BinaryValue --> ValueLength <Byte>&ValueLength

Notes

- If documents are larger than 16KB then chunks will likely contain
only one document. However, documents can never spread across several
chunks (all fields of a single document are in the same chunk).
- When at least one document in a chunk is large enough so that the
chunk is larger than 32KB, then chunk will actually be compressed in
several LZ4 blocks of 16KB. This allows StoredFieldsVisitors which
are only interested in the first fields of a document to not have to
decompress 10MB of data if the document is 10MB, but only 16KB.
- Given that the original lengths are written in the metadata of the
chunk, the decompressorcan leverage this information to stop decoding
as soon as enough data has been decompressed.
- In case documents are incompressible, CompressedDocs will be less
than 0.5% larger than Docs.

2. field_index

A fields index file (extension .fdx).

- FieldsIndex (.fdx) --> <Header>, <ChunkINdex>
- Header --> CodecHeader
- ChunkIndex: See CompressingStoredFieldsInexWriter

Known limitations

This StoredFieldsFormat does not support individual documents larger
than (2^32 - 2^14) bytes. In case this is a problem, you should use
another format, such as Lucene40StoredFieldsFormat.
*/
type Lucene41StoredFieldsFormat struct {
	*compressing.CompressingStoredFieldsFormat
}

func NewLucene41StoredFieldsFormat() *Lucene41StoredFieldsFormat {
	return &Lucene41StoredFieldsFormat{
		compressing.NewCompressingStoredFieldsFormat("Lucene41StoredFields", "", compressing.COMPRESSION_MODE_FAST, 1<<14),
	}
}
