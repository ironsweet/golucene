package fst

import (
	"github.com/balzaczyy/golucene/core/util"
)

type BytesRefFSTEnum struct {
	*FSTEnum
	current *util.BytesRef
	result  *BytesRefFSTEnumIO
	target  *util.BytesRef
}

/* Holds a single input ([]byte) + output pair. */
type BytesRefFSTEnumIO struct {
	Input  []byte
	Output interface{}
}

func NewBytesRefFSTEnum(fst *FST) *BytesRefFSTEnum {
	ans := &BytesRefFSTEnum{
		current: util.NewBytesRef(make([]byte, 10)),
		result:  new(BytesRefFSTEnumIO),
	}
	ans.FSTEnum = newFSTEnum(ans, fst)
	return ans
}

func (e *BytesRefFSTEnum) Next() (*BytesRefFSTEnumIO, error) {
	if err := e.doNext(); err != nil {
		return nil, err
	}
	return e.setResult(), nil
}

func (e *BytesRefFSTEnum) setCurrentLabel(label int) {
	e.current.Value[e.upto] = byte(label)
}

func (e *BytesRefFSTEnum) grow() {
	e.current.Value = util.GrowByteSlice(e.current.Value, e.upto+1)
}

func (e *BytesRefFSTEnum) setResult() *BytesRefFSTEnumIO {
	if e.upto == 0 {
		return nil
	}
	e.result.Input = e.current.Value[1:e.upto]
	e.result.Output = e.output[e.upto]
	return e.result
}
