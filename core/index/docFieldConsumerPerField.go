package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

type DocFieldConsumerPerField interface {
	// Processes all occurrences of a single field
	processFields([]model.IndexableField, int) error
	abort()
	fieldInfo() *model.FieldInfo
}

// index/DocInverterPerField.java

type DocInverterPerField struct {
	_fieldInfo  *model.FieldInfo
	consumer    InvertedDocConsumerPerField
	endConsumer InvertedDocEndConsumerPerField
	docState    *docState
	fieldState  *FieldInvertState
}

func newDocInverterPerField(parent *DocInverter, fieldInfo *model.FieldInfo) *DocInverterPerField {
	ans := &DocInverterPerField{
		_fieldInfo: fieldInfo,
		docState:   parent.docState,
		fieldState: newFieldInvertState(fieldInfo.Name),
	}
	ans.consumer = parent.consumer.addField(ans, fieldInfo)
	ans.endConsumer = parent.endConsumer.addField(ans, fieldInfo)
	return ans
}

func (dipf *DocInverterPerField) abort() {
	defer dipf.endConsumer.abort()
	dipf.consumer.abort()
}

func (di *DocInverterPerField) processFields(fields []model.IndexableField, count int) error {
	di.fieldState.reset()

	doInvert, err := di.consumer.start(fields, count)
	if err != nil {
		return err
	}

	for i, field := range fields[:count] {
		fieldType := field.FieldType()

		// TODO FI: this should be "genericized" to querying consumer if
		// it wants to see this particular field tokenized.
		if fieldType.Indexed() && doInvert {
			analyzed := fieldType.Tokenized() && di.docState.analyzer != nil

			// if the field omits norms, the boost cannot be indexed.
			assert2(!fieldType.OmitNorms() || field.Boost() == 1.0,
				"You cannot set an index-time boost: norms are omitted for field '%v'",
				field.Name())

			// only bother checking offsets if something will consume them.
			// TODO: after we fix analyzers, also check if termVectorOffsets will be indexed.
			checkOffsets := fieldType.IndexOptions() == model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
			lastStartOffset := 0

			if i > 0 && analyzed {
				panic("not implemented yet")
				// di.fieldState.position += di.docState.analyzer.PositionIncrementGap(di.fieldInfo().Name())
			}

			if stream, err := field.TokenStream(di.docState.analyzer); err == nil {
				// reset the TokenStream to the first token
				if err = stream.Reset(); err == nil {
					err = func() (err error) {
						var success2 = false
						defer func() {
							if !success2 {
								util.CloseWhileSuppressingError(stream)
							} else {
								err = stream.Close()
							}
						}()

						var hasMoreTokens bool
						hasMoreTokens, err = stream.IncrementToken()
						if err != nil {
							return err
						}

						di.fieldState.attributeSource = stream.Attributes()

						fmt.Println("DEBUGa", checkOffsets, lastStartOffset, hasMoreTokens)

						panic("not implemented yet")
					}()
				}
			}
			if err != nil {
				return err
			}
		}

		// LUCENE-2387: don't hang onto the field, so GC can recliam
		fields[i] = nil
	}

	err = di.consumer.finish()
	if err == nil {
		err = di.endConsumer.finish()
	}
	return err
}

func (dipf *DocInverterPerField) fieldInfo() *model.FieldInfo {
	return dipf._fieldInfo
}
