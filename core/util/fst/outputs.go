package fst

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

// util/fst/Outputs.java

/**
 * Represents the outputs for an FST, providing the basic
 * algebra required for building and traversing the FST.
 *
 * <p>Note that any operation that returns NO_OUTPUT must
 * return the same singleton object from {@link
 * #getNoOutput}.</p>
 */
type Outputs interface {
	// Eg subtract("foobar", "foo") -> "bar"
	Subtract(output1, output2 interface{}) interface{}
	/** Eg add("foo", "bar") -> "foobar" */
	Add(prefix interface{}, output interface{}) interface{}
	// Encode an final node output value into a DataOutput. By default
	// this just calls write()
	writeFinalOutput(interface{}, util.DataOutput) error
	/** Decode an output value previously written with {@link
	 *  #write(Object, DataOutput)}. */
	Read(in util.DataInput) (e interface{}, err error)
	// Skip the output; defaults to just calling Read() and discarding the result
	SkipOutput(util.DataInput) error
	/** Decode an output value previously written with {@link
	 *  #writeFinalOutput(Object, DataOutput)}.  By default this
	 *  just calls {@link #read(DataInput)}. */
	ReadFinalOutput(in util.DataInput) (e interface{}, err error)
	// Skip the output previously written with WriteFinalOutput;
	// defaults to just calling ReadFinalOutput and discarding the
	// result.
	SkipFinalOutput(util.DataInput) error
	/** NOTE: this output is compared with == so you must
	 *  ensure that all methods return the single object if
	 *  it's really no output */
	NoOutput() interface{}
	outputToString(interface{}) string
	merge(first, second interface{}) interface{}
}

type iOutputsReader interface {
	Read(in util.DataInput) (e interface{}, err error)
	Write(interface{}, util.DataOutput) error
}

type abstractOutputs struct {
	spi iOutputsReader
}

func (out *abstractOutputs) writeFinalOutput(output interface{}, o util.DataOutput) error {
	return out.spi.Write(output, o)
}

func (out *abstractOutputs) SkipOutput(in util.DataInput) error {
	_, err := out.spi.Read(in)
	return err
}

/* Decode an output value previously written with writeFinalOutput(). By default this just calls read(). */
func (out *abstractOutputs) ReadFinalOutput(in util.DataInput) (e interface{}, err error) {
	return out.spi.Read(in)
}

func (out *abstractOutputs) SkipFinalOutput(in util.DataInput) error {
	return out.SkipOutput(in)
}

func (out *abstractOutputs) merge(first, second interface{}) interface{} {
	panic("not supported yet")
}

// fst/NoOutputs.java

var NO_OUTPUT = newNoOutputs()

/* A nil FST Outputs implementation; use this if you just want to build an FSA. */
type NoOutputs struct {
	*abstractOutputs
}

func newNoOutputs() *NoOutputs {
	ans := &NoOutputs{}
	ans.abstractOutputs = &abstractOutputs{ans}
	return ans
}

func (o *NoOutputs) Subtract(output1, output2 interface{}) interface{} {
	assert(output1 == NO_OUTPUT)
	assert(output2 == NO_OUTPUT)
	return NO_OUTPUT
}

func (o *NoOutputs) Add(prefix, output interface{}) interface{} {
	panic("not implemented yet")
}

func (o *NoOutputs) merge(first, second interface{}) interface{} {
	assert(first == NO_OUTPUT)
	assert(second == NO_OUTPUT)
	return NO_OUTPUT
}

func (o *NoOutputs) Write(prefix interface{}, out util.DataOutput) error {
	return nil
}

func (o *NoOutputs) Read(in util.DataInput) (interface{}, error) {
	return NO_OUTPUT, nil
}

func (o *NoOutputs) NoOutput() interface{} {
	return NO_OUTPUT
}

func (o *NoOutputs) outputToString(output interface{}) string {
	return ""
}

// fst/ByteSequenceOutputs.java

/**
 * An FST {@link Outputs} implementation where each output
 * is a sequence of bytes.
 */
type ByteSequenceOutputs struct {
	*abstractOutputs
}

var noOutputs = make([]byte, 0)
var oneByteSequenceOutputs *ByteSequenceOutputs

func ByteSequenceOutputsSingleton() *ByteSequenceOutputs {
	if oneByteSequenceOutputs == nil {
		oneByteSequenceOutputs = &ByteSequenceOutputs{}
		oneByteSequenceOutputs.abstractOutputs = &abstractOutputs{oneByteSequenceOutputs}
	}
	return oneByteSequenceOutputs
}

func (out *ByteSequenceOutputs) Subtract(output1, output2 interface{}) interface{} {
	panic("not implemented yet")
}

func (out *ByteSequenceOutputs) Add(_prefix interface{}, _output interface{}) interface{} {
	if _prefix == nil || _output == nil {
		panic("assert fail")
	}
	prefix, output := _prefix.([]byte), _output.([]byte)
	// if prefix == noOutputs {
	if len(prefix) == 0 {
		return output
		// } else if output == noOutputs {
	} else if len(output) == 0 {
		return prefix
	} else {
		// if len(prefix) == 0 || len(output) == 0 {
		// 	panic("assert fail")
		// }
		result := make([]byte, len(prefix)+len(output))
		copy(result, prefix)
		copy(result[len(prefix):], output)
		return result
	}
}

func (o *ByteSequenceOutputs) Write(obj interface{}, out util.DataOutput) error {
	assert(obj != nil)
	prefix, ok := obj.(*util.BytesRef)
	assert(ok)
	err := out.WriteVInt(int32(len(prefix.Value)))
	if err == nil {
		err = out.WriteBytes(prefix.Value)
	}
	return err
}

func (out *ByteSequenceOutputs) Read(in util.DataInput) (e interface{}, err error) {
	if length, err := in.ReadVInt(); err == nil {
		fmt.Printf("Length: %v\n", length)
		if length == 0 {
			e = out.NoOutput()
		} else {
			buf := make([]byte, length)
			e = buf
			err = in.ReadBytes(buf)
		}
	} else {
		fmt.Printf("Failed to read length due to %v", err)
	}
	return e, err
}

func (out *ByteSequenceOutputs) NoOutput() interface{} {
	return noOutputs
}

func (out *ByteSequenceOutputs) outputToString(output interface{}) string {
	return string(output.(*util.BytesRef).Value)
}

func (out *ByteSequenceOutputs) String() string {
	return "ByteSequenceOutputs"
}

// util/fst/Util.java

/** Looks up the output for this input, or null if the
 *  input is not accepted */
func GetFSTOutput(fst *FST, input []byte) (output interface{}, err error) {
	if fst.inputType != INPUT_TYPE_BYTE1 {
		panic("assert fail")
	}
	fstReader := fst.BytesReader()
	// TODO: would be nice not to alloc this on every lookup
	arc := fst.FirstArc(&Arc{})

	// Accumulate output as we go
	output = fst.outputs.NoOutput()
	for _, v := range input {
		ret, err := fst.FindTargetArc(int(v), arc, arc, fstReader)
		if ret == nil || err != nil {
			return ret, err
		}
		output = fst.outputs.Add(output, arc.Output)
	}

	if arc.IsFinal() {
		return fst.outputs.Add(output, arc.NextFinalOutput), nil
	} else {
		return nil, nil
	}
}
