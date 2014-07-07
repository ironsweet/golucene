package index

// import (
// 	"github.com/balzaczyy/golucene/core/index/model"
// 	"github.com/balzaczyy/golucene/core/util"
// )

// type InvertedDocEndConsumer interface {
// 	flush(map[string]InvertedDocEndConsumerPerField, *model.SegmentWriteState) error
// 	abort()
// 	addField(*DocInverterPerField, *model.FieldInfo) InvertedDocEndConsumerPerField
// 	startDocument()
// 	finishDocument()
// }

// /*
// Writes norms. Each thread X field accumlates the norms for the
// doc/fields it saw, then the flush method below merges all of these
//  together into a single _X.nrm file.
// */
// type NormsConsumer struct {
// }

// func (nc *NormsConsumer) abort() {}

// func (nc *NormsConsumer) flush(fieldsToFlush map[string]InvertedDocEndConsumerPerField,
// 	state *model.SegmentWriteState) (err error) {
// 	var success = false
// 	var normsConsumer DocValuesConsumer
// 	defer func() {
// 		if success {
// 			err = mergeError(err, util.Close(normsConsumer))
// 		} else {
// 			util.CloseWhileSuppressingError(normsConsumer)
// 		}
// 	}()

// 	if state.FieldInfos.HasNorms {
// 		normsFormat := state.SegmentInfo.Codec().(Codec).NormsFormat()
// 		assert(normsFormat != nil)
// 		normsConsumer, err = normsFormat.NormsConsumer(state)
// 		if err != nil {
// 			return err
// 		}

// 		for _, fi := range state.FieldInfos.Values {
// 			toWrite := fieldsToFlush[fi.Name].(*NormsConsumerPerField)
// 			// we must check the final value of omitNorms for the fieldinfo,
// 			// it could have changed for this field since the first time we
// 			// added it.
// 			if !fi.OmitsNorms() {
// 				if toWrite != nil && !toWrite.isEmpty() {
// 					err = toWrite.flush(state, normsConsumer)
// 					if err != nil {
// 						return err
// 					}
// 					assert(fi.NormType() == model.DOC_VALUES_TYPE_NUMERIC)
// 				} else if fi.IsIndexed() {
// 					assertn(int(fi.NormType()) == 0, "got %v; field=%v", fi.NormType(), fi.Name)
// 				}
// 			}
// 		}
// 	}
// 	success = true
// 	return nil
// }

// func (nc *NormsConsumer) finishDocument() {}

// func (nc *NormsConsumer) startDocument() {}

// func (nc *NormsConsumer) addField(docInverterPerField *DocInverterPerField,
// 	fieldInfo *model.FieldInfo) InvertedDocEndConsumerPerField {
// 	return newNormsConsumerPerField(docInverterPerField, fieldInfo, nc)
// }
