package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

/* Default general purpose indexing chain, which handles indexing all types of fields */
type DefaultIndexingChain struct {
	bytesUsed  util.Counter
	docState   *docState
	docWriter  *DocumentsWriterPerThread
	fieldInfos *FieldInfosBuilder

	// Writes postings and term vectors:
	termsHash TermsHash

	storedFieldsWriter StoredFieldsWriter // lazy init
	lastStoredDocId    int

	fieldHash []*PerField
	hashMask  int

	totalFieldCount int
	nextFieldGen    int64

	// Holds fields seen in each document
	fields []*PerField
}

func newDefaultIndexingChain(docWriter *DocumentsWriterPerThread) *DefaultIndexingChain {
	termVectorsWriter := newTermVectorsConsumer(docWriter)
	return &DefaultIndexingChain{
		docWriter:  docWriter,
		fieldInfos: docWriter.fieldInfos,
		docState:   docWriter.docState,
		bytesUsed:  docWriter._bytesUsed,
		termsHash:  newFreqProxTermsWriter(docWriter, termVectorsWriter),
		fieldHash:  make([]*PerField, 2),
		hashMask:   1,
		fields:     make([]*PerField, 1),
	}
}

// TODO: can we remove this lazy-init / make cleaner / do it another way...?
func (c *DefaultIndexingChain) initStoredFieldsWriter() (err error) {
	if c.storedFieldsWriter == nil {
		assert(c != nil)
		assert(c.docWriter != nil)
		assert(c.docWriter.codec != nil)
		assert(c.docWriter.codec.StoredFieldsFormat() != nil)
		c.storedFieldsWriter, err = c.docWriter.codec.StoredFieldsFormat().FieldsWriter(
			c.docWriter.directory, c.docWriter.segmentInfo, store.IO_CONTEXT_DEFAULT)
	}
	return
}

func (c *DefaultIndexingChain) flush(state *SegmentWriteState) (err error) {
	// NOTE: caller (DWPT) handles aborting on any error from this method

	numDocs := state.SegmentInfo.DocCount()
	if err = c.writeNorms(state); err != nil {
		return
	}
	if err = c.writeDocValues(state); err != nil {
		return
	}

	// it's possible all docs hit non-aboritng errors...
	if err = c.initStoredFieldsWriter(); err != nil {
		return
	}
	if err = c.fillStoredFields(numDocs); err != nil {
		return
	}
	if err = c.storedFieldsWriter.Finish(state.FieldInfos, numDocs); err != nil {
		return
	}
	if err = c.storedFieldsWriter.Close(); err != nil {
		return
	}

	fieldsToFlush := make(map[string]TermsHashPerField)
	for _, perField := range c.fieldHash {
		for perField != nil {
			if perField.invertState != nil {
				fieldsToFlush[perField.fieldInfo.Name] = perField.termsHashPerField
			}
			perField = perField.next
		}
	}

	if err = c.termsHash.flush(fieldsToFlush, state); err != nil {
		return
	}

	// important to save after asking consumer to flush so consumer can
	// alter the FieldInfo* if necessary. E.g., FreqProxTermsWriter does
	// this with FieldInfo.storePayload.
	infosWriter := c.docWriter.codec.FieldInfosFormat().FieldInfosWriter()
	return infosWriter(state.Directory, state.SegmentInfo.Name, "", state.FieldInfos, store.IO_CONTEXT_DEFAULT)
}

/* Writes all buffered doc values (called from flush()) */
func (c *DefaultIndexingChain) writeDocValues(state *SegmentWriteState) (err error) {
	docCount := state.SegmentInfo.DocCount()
	var dvConsumer DocValuesConsumer
	var success = false
	if success {
		err = util.Close(dvConsumer)
	} else {
		util.CloseWhileSuppressingError(dvConsumer)
	}

	for _, perField := range c.fieldHash {
		for perField != nil {
			if perField.docValuesWriter != nil {
				if dvConsumer == nil {
					// lazy init
					fmt := state.SegmentInfo.Codec().(Codec).DocValuesFormat()
					if dvConsumer, err = fmt.FieldsConsumer(state); err != nil {
						return
					}
				}

				perField.docValuesWriter.finish(docCount)
				if err = perField.docValuesWriter.flush(state, dvConsumer); err != nil {
					return
				}
				perField.docValuesWriter = nil
			}
			perField = perField.next
		}
	}

	success = true
	return nil
}

/*
Catch up for all docs before us that had no stored fields, or hit
non-aborting errors before writing stored fields.
*/
func (c *DefaultIndexingChain) fillStoredFields(docId int) (err error) {
	for err == nil && c.lastStoredDocId < docId {
		err = c.startStoredFields()
		if err == nil {
			err = c.finishStoredFields()
		}
	}
	return
}

func (c *DefaultIndexingChain) writeNorms(state *SegmentWriteState) (err error) {
	var success = false
	var normsConsumer DocValuesConsumer
	defer func() {
		if success {
			err = util.Close(normsConsumer)
		} else {
			util.CloseWhileSuppressingError(normsConsumer)
		}
	}()

	if state.FieldInfos.HasNorms {
		normsFormat := state.SegmentInfo.Codec().(Codec).NormsFormat()
		assert(normsFormat != nil)
		if normsConsumer, err = normsFormat.NormsConsumer(state); err != nil {
			return
		}

		for _, fi := range state.FieldInfos.Values {
			perField := c.perField(fi.Name)
			assert(perField != nil)

			// we must check the final value of omitNorms for the FieldInfo:
			// it could have changed for this field since the first time we
			// added it.
			if !fi.OmitsNorms() {
				if perField.norms != nil {
					perField.norms.finish(state.SegmentInfo.DocCount())
					if err = perField.norms.flush(state, normsConsumer); err != nil {
						return
					}
					assert(fi.NormType() == DOC_VALUES_TYPE_NUMERIC)
				} else if fi.IsIndexed() {
					assert2(fi.NormType() == 0, "got %v; field=%v", fi.NormType(), fi.Name)
				}
			}
		}
	}
	success = true
	return nil
}

func (c *DefaultIndexingChain) abort() {
	// E.g. close any open files in the stored fields writer:
	if c.storedFieldsWriter != nil {
		c.storedFieldsWriter.Abort() // ignore error
	}

	// E.g. close any open files in the term vectors writer:
	c.termsHash.abort()

	for i, _ := range c.fieldHash {
		c.fieldHash[i] = nil
	}
}

func (c *DefaultIndexingChain) rehash() {
	newHashSize := 2 * len(c.fieldHash)
	assert(newHashSize > len(c.fieldHash))

	newHashArray := make([]*PerField, newHashSize)

	// rehash
	newHashMask := newHashSize - 1
	for _, fp0 := range c.fieldHash {
		for fp0 != nil {
			hashPos2 := util.Hashstr(fp0.fieldInfo.Name) & newHashMask
			fp0.next, newHashArray[hashPos2], fp0 =
				newHashArray[hashPos2], fp0, fp0.next
		}
	}

	c.fieldHash = newHashArray
	c.hashMask = newHashMask
}

/* Calls StoredFieldsWriter.startDocument, aborting the segment if it hits any error. */
func (c *DefaultIndexingChain) startStoredFields() (err error) {
	var success = false
	defer func() {
		if !success {
			c.docWriter.setAborting()
		}
	}()

	if err = c.initStoredFieldsWriter(); err != nil {
		return
	}
	if err = c.storedFieldsWriter.StartDocument(); err != nil {
		return
	}
	success = true

	c.lastStoredDocId++
	return nil
}

/* Calls StoredFieldsWriter.finishDocument(), aborting the segment if it hits any error. */
func (c *DefaultIndexingChain) finishStoredFields() error {
	var success = false
	defer func() {
		if !success {
			c.docWriter.setAborting()
		}
	}()
	if err := c.storedFieldsWriter.FinishDocument(); err != nil {
		return err
	}
	success = true
	return nil
}

func (c *DefaultIndexingChain) processDocument() (err error) {
	// How many indexed field names we've seen (collapses multiple
	// field instances by the same name):
	fieldCount := 0

	fieldGen := c.nextFieldGen
	c.nextFieldGen++

	// NOTE: we need to passes here, in case there are multi-valued
	// fields, because we must process all instances of a given field
	// at once, since the anlayzer is free to reuse TOkenStream across
	// fields (i.e., we cannot have more than one TokenStream running
	// "at once"):

	c.termsHash.startDocument()

	if err = c.fillStoredFields(c.docState.docID); err != nil {
		return
	}
	if err = c.startStoredFields(); err != nil {
		return
	}

	if err = func() error {
		defer func() {
			if !c.docWriter.aborting {
				// Finish each indexed field name seen in the document:
				for _, field := range c.fields[:fieldCount] {
					err = mergeError(err, field.finish())
				}
				err = mergeError(err, c.finishStoredFields())
			}
		}()

		for _, field := range c.docState.doc {
			if fieldCount, err = c.processField(field, fieldGen, fieldCount); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return
	}

	var success = false
	defer func() {
		if !success {
			// Must abort, on the possibility that on-disk term vectors are now corrupt:
			c.docWriter.setAborting()
		}
	}()

	if err = c.termsHash.finishDocument(); err != nil {
		return
	}
	success = true
	return nil
}

func (c *DefaultIndexingChain) processField(field IndexableField,
	fieldGen int64, fieldCount int) (int, error) {

	var fieldName string = field.Name()
	var fieldType IndexableFieldType = field.FieldType()
	var fp *PerField

	// Invert indexed fields:
	if fieldType.Indexed() {

		// if the field omits norms, the boost cannot be indexed.
		if fieldType.OmitNorms() && field.Boost() != 1 {
			panic(fmt.Sprintf(
				"You cannot set an index-time boost: norms are omitted for field '%v'",
				fieldName))
		}

		fp = c.getOrAddField(fieldName, fieldType, true)
		first := fp.fieldGen != fieldGen
		if err := fp.invert(field, first); err != nil {
			return 0, err
		}

		if first {
			c.fields[fieldCount] = fp
			fieldCount++
			fp.fieldGen = fieldGen
		}
	} else {
		panic("not implemented yet")
	}

	// Add stored fields:
	if fieldType.Stored() {
		if fp == nil {
			panic("not implemented yet")
		}
		if fieldType.Stored() {
			if err := func() error {
				var success = false
				defer func() {
					if !success {
						c.docWriter.setAborting()
					}
				}()

				if err := c.storedFieldsWriter.WriteField(fp.fieldInfo, field); err != nil {
					return err
				}
				success = true
				return nil
			}(); err != nil {
				return 0, err
			}
		}
	}

	if dvType := fieldType.DocValueType(); int(dvType) != 0 {
		if fp == nil {
			panic("not implemented yet")
		}
		panic("not implemented yet")
	}

	return fieldCount, nil
}

/*
Returns a previously created PerField, or nil if this field name
wasn't seen yet.
*/
func (c *DefaultIndexingChain) perField(name string) *PerField {
	hashPos := util.Hashstr(name) & c.hashMask
	fp := c.fieldHash[hashPos]
	for fp != nil && fp.fieldInfo.Name != name {
		fp = fp.next
	}
	return fp
}

func (c *DefaultIndexingChain) getOrAddField(name string,
	fieldType IndexableFieldType, invert bool) *PerField {

	// Make sure we have a PerField allocated
	hashPos := util.Hashstr(name) & c.hashMask
	fp := c.fieldHash[hashPos]
	for fp != nil && fp.fieldInfo.Name != name && fp != fp.next {
		fp = fp.next
	}

	if fp == nil {
		// First time we are seeing this field in this segment

		fi := c.fieldInfos.AddOrUpdate(name, fieldType)

		fp = newPerField(c, fi, invert)
		fp.next = c.fieldHash[hashPos]
		c.fieldHash[hashPos] = fp
		c.totalFieldCount++

		// At most 50% load factor:
		if c.totalFieldCount >= len(c.fieldHash)/2 {
			c.rehash()
		}

		if c.totalFieldCount > len(c.fields) {
			newFields := make([]*PerField, util.Oversize(c.totalFieldCount, util.NUM_BYTES_OBJECT_REF))
			copy(newFields, c.fields)
			c.fields = newFields
		}

	} else {
		fp.fieldInfo.Update(fieldType)

		if invert && fp.invertState == nil {
			fp.setInvertState()
		}
	}

	return fp
}

type PerField struct {
	*DefaultIndexingChain // acess at least docState, termsHash.

	fieldInfo  *FieldInfo
	similarity Similarity

	invertState       *FieldInvertState
	termsHashPerField TermsHashPerField

	// non-nil if this field ever had doc values in this segment:
	docValuesWriter DocValuesWriter

	// We use this to know when a PerField is seen for the first time
	// in the current document.
	fieldGen int64

	// Used by the hash table
	next *PerField

	// Lazy init'd:
	norms *NumericDocValuesWriter

	// reused
	tokenStream analysis.TokenStream
}

func newPerField(parent *DefaultIndexingChain,
	fieldInfo *FieldInfo, invert bool) *PerField {

	ans := &PerField{
		DefaultIndexingChain: parent,
		fieldInfo:            fieldInfo,
		similarity:           parent.docState.similarity,
		fieldGen:             -1,
	}
	if invert {
		ans.setInvertState()
	}
	return ans
}

func (f *PerField) setInvertState() {
	f.invertState = newFieldInvertState(f.fieldInfo.Name)
	f.termsHashPerField = f.termsHash.addField(f.invertState, f.fieldInfo)
}

func (f *PerField) finish() error {
	if !f.fieldInfo.OmitsNorms() {
		if f.norms == nil {
			f.fieldInfo.SetNormValueType(DOC_VALUES_TYPE_NUMERIC)
			f.norms = newNumericDocValuesWriter(f.fieldInfo, f.docState.docWriter._bytesUsed, false)
		}
		f.norms.addValue(f.docState.docID, f.similarity.ComputeNorm(f.invertState))
	}
	return f.termsHashPerField.finish()
}

/*
Inverts one field for one document; first is true if this is the
first time we are seeing this field name in this document.
*/
func (f *PerField) invert(field IndexableField, first bool) error {
	if first {
		// first time we're seeing this field (indexed) in this document:
		f.invertState.reset()
	}

	fieldType := field.FieldType()

	analyzed := fieldType.Tokenized() && f.docState.analyzer != nil

	if err := func() (err error) {
		// only bother checking offsets if something will consume them
		// TODO: after we fix analyzers, also check if termVectorOffsets will be indexed.
		checkOffsets := fieldType.IndexOptions() == INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS

		// To assist people in tracking down problems in analysis components,
		// we wish to write the field name to the infostream when we fail.
		// We expect some caller to eventually deal with the real error, so
		// we don't want any error handling, but rather a deferred function
		// that takes note of the problem.
		aborting := false
		succeededInProcessingField := false
		defer func() {
			if err != nil {
				if _, ok := err.(util.MaxBytesLengthExceededError); ok {
					aborting = false
					prefix := make([]byte, 30)
					bigTerm := f.invertState.termAttribute.BytesRef()
					copy(prefix, bigTerm.ToBytes()[:30]) // keep at most 30 characters
					if f.docState.infoStream.IsEnabled("IW") {
						f.docState.infoStream.Message("IW",
							"ERROR: Document contains at least one immense term in field='%v' "+
								"(whose UTF8 encoding is longer than the max length %v), "+
								"all of which were skipped. Please correct the analyzer to not produce such terms. "+
								"The prefix of the first immense term is: '%v...', original message: %v",
							f.fieldInfo.Name, MAX_TERM_LENGTH_UTF8, string(prefix), err)
					}
				}
			}
			if !succeededInProcessingField && aborting {
				f.docState.docWriter.setAborting()
			}

			if !succeededInProcessingField && f.docState.infoStream.IsEnabled("DW") {
				f.docState.infoStream.Message("DW",
					"An error was returned while processing field %v",
					f.fieldInfo.Name)
			}
		}()

		var stream analysis.TokenStream
		stream, err = field.TokenStream(f.docState.analyzer, f.tokenStream)
		if err != nil {
			return err
		}
		defer stream.Close()

		f.tokenStream = stream
		// reset the TokenStream to the first token
		if err = stream.Reset(); err != nil {
			return err
		}

		f.invertState.setAttributeSource(stream.Attributes())

		f.termsHashPerField.start(field, first)

		for {
			var ok bool
			if ok, err = stream.IncrementToken(); err != nil {
				return err
			}
			if !ok {
				break
			}

			// if we hit an error in stream.next below (which is fairly
			// common, e.g. if analyzer chokes on a given document), then
			// it's non-aborting and (above) this one document will be
			// marked as deleted, but still consume a docId

			posIncr := f.invertState.posIncrAttribute.PositionIncrement()
			if f.invertState.position += posIncr; f.invertState.position < f.invertState.lastPosition {
				assert2(posIncr != 0,
					"first position increment must be > 0 (got 0) for field '%v'",
					field.Name)
				panic(fmt.Sprintf(
					"position increments (and gaps) must be >= 0 (got %v) for field '%v'",
					posIncr, field.Name))
			}
			f.invertState.lastPosition = f.invertState.position
			if posIncr == 0 {
				f.invertState.numOverlap++
			}

			if checkOffsets {
				startOffset := f.invertState.offset + f.invertState.offsetAttribute.StartOffset()
				endOffset := f.invertState.offset + f.invertState.offsetAttribute.EndOffset()
				assert2(startOffset >= f.invertState.lastStartOffset && startOffset <= endOffset,
					"startOffset must be non-negative, "+
						"and endOffset must be >= startOffset, "+
						"and offsets must not go backwards "+
						"startOffset=%v,endOffset=%v,lastStartOffset=%v for field '%v'",
					startOffset, endOffset, f.invertState.lastStartOffset, field.Name)
				f.invertState.lastStartOffset = startOffset
			}

			// fmt.Printf("  term=%v\n", f.invertState.termAttribute)

			// if we hit an error in here, we abort all buffered documents
			// since the last flush, on the likelihood that the internal
			// state of the terms hash is now corrupt and should not be
			// flushed to a new segment:
			aborting = true
			if err = f.termsHashPerField.add(); err != nil {
				return err
			}
			aborting = false

			f.invertState.length++
		}

		// trigger streams to perform end-of-stream operations
		if err = stream.End(); err != nil {
			return err
		}

		// TODO: maybe add some safety? then again, it's already checked
		// when we come back arond to the field...
		f.invertState.position += f.invertState.posIncrAttribute.PositionIncrement()
		f.invertState.offset += f.invertState.offsetAttribute.EndOffset()

		// if there is an error coming through, we don't set this to true here:
		succeededInProcessingField = true
		return nil
	}(); err != nil {
		return err
	}

	if analyzed {
		f.invertState.position += f.docState.analyzer.PositionIncrementGap(f.fieldInfo.Name)
		f.invertState.offset += f.docState.analyzer.OffsetGap(f.fieldInfo.Name)
	}

	f.invertState.boost *= field.Boost()
	return nil
}
