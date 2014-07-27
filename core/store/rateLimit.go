package store

import (
	"math"
)

// store/RateLimiter.java

/*
Abstract base class to rate limit IO. Typically implementations are
shared across multiple IndexInputs or IndexOutputs (for example those
involved all merging). Those IndexInputs and IndexOutputs would call
pause() whenever they have read or written more than
minPauseCheckBytes() bytes.
*/
type RateLimiter interface {
	// Sets an updated mb per second rate limit.
	SetMbPerSec(mbPerSec float64)
	// The current mb per second rate limit.
	MbPerSec() float64
	/*
		Pause, if necessary, to keep the instantaneous IO rate at or below
		the target.

		Note: the implementation is thread-safe
	*/
	Pause(bytes int64) int64
}

// Simple class to rate limit IO
// Ian: volatile is not supported
type SimpleRateLimiter struct {
	mbPerSec           float64 // volatile
	minPauseCheckBytes int64   // volatile
	lastNS             int64
}

// mbPerSec is the MB/sec max IO rate
func newSimpleRateLimiter(mbPerSec float64) *SimpleRateLimiter {
	ans := &SimpleRateLimiter{}
	ans.SetMbPerSec(mbPerSec)
	return ans
}

func (srl *SimpleRateLimiter) SetMbPerSec(mbPerSec float64) {
	srl.mbPerSec = mbPerSec
	panic("not implemented yet")
}

func (srl *SimpleRateLimiter) MbPerSec() float64 {
	return srl.mbPerSec
}

/*
Pause, if necessary, to keep the instantaneous IO rate at or below
the target. Be sure to only call this method when bytes >
minPauseCheckBytes(), otherwise it will pause way too long!
*/
func (srl *SimpleRateLimiter) Pause(bytes int64) int64 {
	panic("not implemented yet")
	// if bytes == 1 {
	// 	return 0
	// }

	// // TODO: this is purely instantaneous rate; maybe we
	// // should also offer decayed recent history one?
	// srl.lastNS += int64(float64(bytes) * srl.nsPerByte)
	// targetNS := srl.lastNS
	// startNS := time.Now().UnixNano()
	// curNS := startNS
	// if srl.lastNS < curNS {
	// 	srl.lastNS = curNS
	// }

	// // While loop because sleep doesn't always sleep enough:
	// for pauseNS := targetNS - curNS; pauseNS > 0; pauseNS = targetNS - curNS {
	// 	time.Sleep(time.Duration(pauseNS * int64(time.Nanosecond)))
	// 	curNS = time.Now().UnixNano()
	// }
	// return curNS - startNS
}

// store/RateLimitedDirectoryWrapper.java

// A Directory wrapper that allows IndexOutput rate limiting using
// IO context specific rate limiters.
type RateLimitedDirectoryWrapper struct {
	Directory
	// we need to be volatile here to make sure we see all the values
	// that are set / modified concurrently
	// Ian: volatile is not supported
	contextRateLimiters []RateLimiter // volatile
	isOpen              bool
}

func NewRateLimitedDirectoryWrapper(wrapped Directory) *RateLimitedDirectoryWrapper {
	panic("not implemented yet")
	// return &RateLimitedDirectoryWrapper{
	// 	Directory:           wrapped,
	// 	contextRateLimiters: make([]RateLimiter, 4), // TODO magic number
	// 	isOpen:              true,
	// }
}

func (w *RateLimitedDirectoryWrapper) CreateOutput(name string, ctx IOContext) (IndexOutput, error) {
	w.EnsureOpen()
	output, err := w.Directory.CreateOutput(name, ctx)
	if err == nil {
		if limiter := w.rateLimiter(ctx.context); limiter != nil {
			output = newRateLimitedIndexOutput(limiter, output)
		}
	}
	return output, err
}

// func (w *RateLimitedDirectoryWrapper) Close() error {
// 	w.isOpen = false
// 	return w.Directory.Close()
// }

// func (w *RateLimitedDirectoryWrapper) String() string {
// 	return fmt.Sprintf("RateLimitedDirectoryWrapper(%v)", w.Directory)
// }

func (w *RateLimitedDirectoryWrapper) rateLimiter(ctx IOContextType) RateLimiter {
	assert(int(ctx) != 0)
	return w.contextRateLimiters[int(ctx)-1]
}

/*
Sets the maximum (approx) MB/sec allowed by all write IO performed by
IndexOutput created with the given context. Pass non-positve value to
have no limit.

NOTE: For already created IndexOutput instances there is no guarantee
this new rate will apply to them; it will only be guaranteed to apply
for new created IndexOutput instances.

NOTE: this is an optional operation and might not be respected by all
Directory implementations. Currently only buffered Directory
implementations use rate-limiting.
*/
func (w *RateLimitedDirectoryWrapper) SetMaxWriteMBPerSec(mbPerSec float64, context int) {
	if !w.isOpen {
		panic("this Directory is closed")
	}
	if context == 0 {
		panic("Context must not be nil")
	}
	ord := context - 1
	limiter := w.contextRateLimiters[ord]
	if mbPerSec <= 0 {
		if limiter != nil {
			limiter.SetMbPerSec(math.MaxFloat64)
			w.contextRateLimiters[ord] = nil
			// atomic.StorePointer(&(w.contextRateLimiters[ord]), nil)
		}
	} else if limiter != nil {
		limiter.SetMbPerSec(mbPerSec)
		// atomic.StorePointer(&(w.contextRateLimiters[ord]), limiter) // cross the mem barrier again
	} else {
		w.contextRateLimiters[ord] = newSimpleRateLimiter(mbPerSec)
		// atomic.StorePointer(&(w.contextRateLimiters[ord]), newSimpleRateLimiter(mbPerSec))
	}
}

/*
Sets the rate limiter to be used to limit (approx) MB/sec allowed by
all IO performed with the given context. Pass non-positive to have no
limit.

Passing an instance of rate limiter compared to settng it using
setMaxWriteMBPersec() allows to use the same limiter instance across
several directories globally limiting IO across them.
*/
func (w *RateLimitedDirectoryWrapper) setRateLimiter(mergeWriteRateLimiter RateLimiter, context int) {
	panic("not implemented yet")
}

func (w *RateLimitedDirectoryWrapper) MaxWriteMBPerSec(context int) {
	panic("not implemented yet")
}

// store/RateLimitedIndexOutput.java

type flushBuffer interface {
	FlushBuffer(buf []byte) error
}

/* A rate limiting IndexOutput */
type RateLimitedIndexOutput struct {
	*IndexOutputImpl
	delegate    IndexOutput
	rateLimiter RateLimiter
}

func newRateLimitedIndexOutput(rateLimiter RateLimiter, delegate IndexOutput) *RateLimitedIndexOutput {
	panic("not implemented yet")
	// ans := &RateLimitedIndexOutput{}
	// ans.BufferedIndexOutput = NewBufferedIndexOutput(DEFAULT_BUFFER_SIZE, ans)
	// ans.delegate = delegate
	// ans.rateLimiter = rateLimiter
	// return ans
}

func (out *RateLimitedIndexOutput) Close() error {
	return out.delegate.Close()
}

func (out *RateLimitedIndexOutput) FilePointer() int64 {
	panic("not implementd yet")
}

func (out *RateLimitedIndexOutput) Checksum() int64 {
	return out.delegate.Checksum()
}

func (out *RateLimitedIndexOutput) WriteByte(b byte) error {
	panic("not implemented yet")
}

func (out *RateLimitedIndexOutput) WriteBytes(p []byte) error {
	panic("not implemented yet")
}

// func (out *RateLimitedIndexOutput) FlushBuffer(buf []byte) error {
// 	out.rateLimiter.Pause(int64(len(buf)))
// 	if v, ok := out.delegate.(flushBuffer); ok {
// 		return v.FlushBuffer(buf)
// 	}
// 	panic("double check if flushBuffer interface is satisfied")
// 	return out.delegate.WriteBytes(buf)
// }

// func (out *RateLimitedIndexOutput) Length() (int64, error) {
// 	return out.delegate.Length()
// }
