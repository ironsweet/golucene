package util

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"time"
)

// util/ThrottledIndexOutput.java

const DEFAULT_MIN_WRITTEN_BYTES = 024

// Intentionally slow IndexOutput for testing.
type ThrottledIndexOutput struct {
	*store.IndexOutputImpl
	bytesPerSecond   int
	delegate         store.IndexOutput
	flushDelayMillis int64
	closeDelayMillis int64
	seekDelayMillis  int64
	pendingBytes     int64
	minBytesWritten  int64
	timeElapsed      int64
	bytes            []byte
}

func (out *ThrottledIndexOutput) NewFromDelegate(output store.IndexOutput) *ThrottledIndexOutput {
	ans := &ThrottledIndexOutput{
		delegate:         output,
		bytesPerSecond:   out.bytesPerSecond,
		flushDelayMillis: out.flushDelayMillis,
		closeDelayMillis: out.closeDelayMillis,
		seekDelayMillis:  out.seekDelayMillis,
		minBytesWritten:  out.minBytesWritten,
		bytes:            make([]byte, 1),
	}
	ans.IndexOutputImpl = store.NewIndexOutput(ans)
	return ans
}

func NewThrottledIndexOutput(bytesPerSecond int, delayInMillis int64, delegate store.IndexOutput) *ThrottledIndexOutput {
	assert(bytesPerSecond > 0)
	ans := &ThrottledIndexOutput{
		delegate:         delegate,
		bytesPerSecond:   bytesPerSecond,
		flushDelayMillis: delayInMillis,
		closeDelayMillis: delayInMillis,
		seekDelayMillis:  delayInMillis,
		minBytesWritten:  DEFAULT_MIN_WRITTEN_BYTES,
		bytes:            make([]byte, 1),
	}
	ans.IndexOutputImpl = store.NewIndexOutput(ans)
	return ans
}

func MBitsToBytes(mBits int) int {
	return mBits * 125000
}

func (out *ThrottledIndexOutput) Close() error {
	<-time.After(time.Duration(out.closeDelayMillis + out.delay(true)))
	return out.delegate.Close()
}

func (out *ThrottledIndexOutput) FilePointer() int64 {
	return out.delegate.FilePointer()
}

func (out *ThrottledIndexOutput) WriteByte(b byte) error {
	out.bytes[0] = b
	return out.WriteBytes(out.bytes)
}

func (out *ThrottledIndexOutput) WriteBytes(buf []byte) error {
	before := time.Now()
	// TODO: sometimes, write only half the bytes, then sleep, then 2nd
	// half, then sleep, so we sometimes interrupt having only written
	// not all bytes
	err := out.delegate.WriteBytes(buf)
	if err != nil {
		return err
	}
	out.timeElapsed += int64(time.Now().Sub(before))
	out.pendingBytes += int64(len(buf))
	<-time.After(time.Duration(out.delay(false)))
	return nil
}

func (out *ThrottledIndexOutput) delay(closing bool) int64 {
	if out.pendingBytes > 0 && (closing || out.pendingBytes > out.minBytesWritten) {
		actualBps := (out.timeElapsed / out.pendingBytes) * 1000000000
		if actualBps > int64(out.bytesPerSecond) {
			expected := out.pendingBytes * 1000 / int64(out.bytesPerSecond)
			delay := expected - (out.timeElapsed / 1000000)
			out.pendingBytes = 0
			out.timeElapsed = 0
			return delay
		}
	}
	return 0
}

func (out *ThrottledIndexOutput) CopyBytes(input util.DataInput, numBytes int64) error {
	return out.delegate.CopyBytes(input, numBytes)
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}
