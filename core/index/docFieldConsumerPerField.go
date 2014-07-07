package index

// import (
// 	ta "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
// 	"github.com/balzaczyy/golucene/core/index/model"
// 	"github.com/balzaczyy/golucene/core/util"
// )

// type DocFieldConsumerPerField interface {
// 	// Processes all occurrences of a single field
// 	processFields([]model.IndexableField, int) error
// 	abort()
// 	fieldInfo() *model.FieldInfo
// }

// index/DocInverterPerField.java

// type DocInverterPerField struct {
// 	_fieldInfo  *model.FieldInfo
// 	consumer    InvertedDocConsumerPerField
// 	endConsumer InvertedDocEndConsumerPerField
// 	docState    *docState
// 	fieldState  *FieldInvertState
// }

// func newDocInverterPerField(parent *DocInverter, fieldInfo *model.FieldInfo) *DocInverterPerField {
// 	ans := &DocInverterPerField{
// 		_fieldInfo: fieldInfo,
// 		docState:   parent.docState,
// 		fieldState: newFieldInvertState(fieldInfo.Name),
// 	}
// 	ans.consumer = parent.consumer.addField(ans, fieldInfo)
// 	ans.endConsumer = parent.endConsumer.addField(ans, fieldInfo)
// 	return ans
// }

// func (dipf *DocInverterPerField) abort() {
// 	defer dipf.endConsumer.abort()
// 	dipf.consumer.abort()
// }

// func (di *DocInverterPerField) processFields(fields []model.IndexableField, count int) error {
// 	di.fieldState.reset()

// 	doInvert, err := di.consumer.start(fields, count)
// 	if err != nil {
// 		return err
// 	}

// 	for i, field := range fields[:count] {
// 		fieldType := field.FieldType()

// 		// TODO FI: this should be "genericized" to querying consumer if
// 		// it wants to see this particular field tokenized.
// 		if fieldType.Indexed() && doInvert {
// 			analyzed := fieldType.Tokenized() && di.docState.analyzer != nil

// 			// if the field omits norms, the boost cannot be indexed.
// 			assert2(!fieldType.OmitNorms() || field.Boost() == 1.0,
// 				"You cannot set an index-time boost: norms are omitted for field '%v'",
// 				field.Name())

// 			// only bother checking offsets if something will consume them.
// 			// TODO: after we fix analyzers, also check if termVectorOffsets will be indexed.
// 			checkOffsets := fieldType.IndexOptions() == model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
// 			lastStartOffset := 0

// 			if i > 0 && analyzed {
// 				panic("not implemented yet")
// 				// di.fieldState.position += di.docState.analyzer.PositionIncrementGap(di.fieldInfo().Name())
// 			}

// 			if stream, err := field.TokenStream(di.docState.analyzer); err == nil {
// 				// reset the TokenStream to the first token
// 				if err = stream.Reset(); err == nil {
// 					err = func() (err error) {
// 						var success2 = false
// 						defer func() {
// 							if !success2 {
// 								util.CloseWhileSuppressingError(stream)
// 							} else {
// 								err = stream.Close()
// 							}
// 						}()

// 						var hasMoreTokens bool
// 						hasMoreTokens, err = stream.IncrementToken()
// 						if err != nil {
// 							return err
// 						}

// 						atts := stream.Attributes()
// 						di.fieldState.attributeSource = atts

// 						offsetAttribute := atts.Add("OffsetAttribute").(ta.OffsetAttribute)
// 						posIncrAttribute := atts.Add("PositionIncrementAttribute").(ta.PositionIncrementAttribute)

// 						if hasMoreTokens {
// 							di.consumer.startField(field)

// 							for {
// 								// If we hit an error in stream.next below (which is
// 								// fairy common, eg if analyer chokes on a given
// 								// document), then it's non-aborting and (above) this
// 								// one document will be marked as deleted, but still
// 								// consume a docID

// 								posIncr := posIncrAttribute.PositionIncrement()
// 								assert2(posIncr >= 0,
// 									"position increment must be >=0 (got %v) for field '%v'",
// 									posIncr, field.Name())
// 								assert2(di.fieldState.position != 0 || posIncr != 0,
// 									"first position increment must be > 0 (got 0) for field '%v'",
// 									field.Name())
// 								position := di.fieldState.position + posIncr
// 								if position > 0 {
// 									// NOTE: confusing: this "mirrors" the position++ we do below
// 									position--
// 								} else {
// 									assert2(position >= 0, "position overflow for field '%v'", field.Name())
// 								}

// 								// position is legal, we can safely place it in fieldState now.
// 								// not sure if anything will use fieldState after non-aborting exc...
// 								di.fieldState.position = position

// 								if posIncr == 0 {
// 									di.fieldState.numOverlap++
// 								}

// 								if checkOffsets {
// 									startOffset := di.fieldState.offset + offsetAttribute.StartOffset()
// 									endOffset := di.fieldState.offset + offsetAttribute.EndOffset()
// 									assert2(startOffset >= 0 && startOffset <= endOffset,
// 										"startOffset must be non-negative, and endOffset must be >= startOffset, startOffset=%v,endOffset=%v for field '%v'",
// 										startOffset, endOffset, field.Name())
// 									assert2(startOffset >= lastStartOffset,
// 										"offsets must not go backwards startOffset=%v is < lastStartOffset=%v for field '%v'",
// 										startOffset, lastStartOffset, field.Name())
// 									lastStartOffset = startOffset
// 								}

// 								if err = func() error {
// 									var success = false
// 									defer func() {
// 										if !success {
// 											di.docState.docWriter.setAborting()
// 										}
// 									}()
// 									// If we hit an error here, we abort all buffered
// 									// documents since the last flush, on the
// 									// likelihood that the internal state of the
// 									// consumer is now corrupt and should not be
// 									// flushed to a new segment:
// 									if err := di.consumer.add(); err != nil {
// 										return err
// 									}
// 									success = true
// 									return nil
// 								}(); err != nil {
// 									return err
// 								}

// 								di.fieldState.length++
// 								di.fieldState.position++
// 								ok, err := stream.IncrementToken()
// 								if err != nil {
// 									return err
// 								}
// 								if !ok {
// 									break
// 								}
// 							}
// 						}
// 						// trigger stream to perform end-of-stream operations
// 						err = stream.End()
// 						if err != nil {
// 							return err
// 						}
// 						// TODO: maybe add some safety? then again, it's alread
// 						// checked when we come back around to the field...
// 						di.fieldState.position += posIncrAttribute.PositionIncrement()
// 						di.fieldState.offset += offsetAttribute.EndOffset()
// 						success2 = true
// 						return nil
// 					}()
// 				}
// 			}
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// LUCENE-2387: don't hang onto the field, so GC can recliam
// 		fields[i] = nil
// 	}

// 	err = di.consumer.finish()
// 	if err == nil {
// 		err = di.endConsumer.finish()
// 	}
// 	return err
// }

// func (dipf *DocInverterPerField) fieldInfo() *model.FieldInfo {
// 	return dipf._fieldInfo
// }
