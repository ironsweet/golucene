package util

import (
	"fmt"
	"reflect"
)

// util.RamUsageEstimator.java

// amd64 system
const (
	NUM_BYTES_CHAR  = 2 // UTF8 uses 1-4 bytes to represent each rune
	NUM_BYTES_SHORT = 2
	NUM_BYTES_INT   = 8
	NUM_BYTES_FLOAT = 4
	NUM_BYTES_LONG  = 8

	/* Number of bytes to represent an object reference */
	NUM_BYTES_OBJECT_REF = 8

	// Number of bytes to represent an object header (no fields, no alignments).
	NUM_BYTES_OBJECT_HEADER = 16

	// Number of bytes to represent an array header (no content, but with alignments).
	NUM_BYTES_ARRAY_HEADER = 24

	// A constant specifying the object alignment boundary inside the
	// JVM. Objects will always take a full multiple of this constant,
	// possibly wasting some space.
	NUM_BYTES_OBJECT_ALIGNMENT = 8
)

/* Aligns an object size to be the next multiple of NUM_BYTES_OBJECT_ALIGNMENT */
func AlignObjectSize(size int64) int64 {
	size += NUM_BYTES_OBJECT_ALIGNMENT - 1
	return size - (size & NUM_BYTES_OBJECT_ALIGNMENT)
}

/* Returns the size in bytes of the object. */
func SizeOf(arr interface{}) int64 {
	if arr == nil {
		return 0
	}
	switch arr.(type) {
	case []int64:
		return AlignObjectSize(NUM_BYTES_ARRAY_HEADER + NUM_BYTES_LONG*int64(len(arr.([]int64))))
	case []int16:
		return AlignObjectSize(NUM_BYTES_ARRAY_HEADER + NUM_BYTES_SHORT*int64(len(arr.([]int16))))
	case []byte:
		return AlignObjectSize(NUM_BYTES_ARRAY_HEADER + int64(len(arr.([]byte))))
	default:
		if t := reflect.TypeOf(arr); t.Kind() == reflect.Slice {
			var size int64
			v := reflect.ValueOf(arr)
			for i := 0; i < v.Len(); i++ {
				size += v.Index(i).Interface().(Accountable).RamBytesUsed()
			}
			return size
		}
		fmt.Println("Unknown type:", reflect.TypeOf(arr))
		panic("not supported yet")
	}
}

/*
Estimates a "shallow" memory usage of the given object. For slices,
this will be the memory taken by slice storage (no subreferences will
be followed). For objects, this will be the memory taken by the fields.
*/
func ShallowSizeOf(obj interface{}) int64 {
	if obj == nil {
		return 0
	}
	clz := reflect.TypeOf(obj)
	if clz.Kind() == reflect.Slice {
		return shallowSizeOfArray(obj)
	}
	return ShallowSizeOfInstance(clz)
}

func ShallowSizeOfInstance(clazz reflect.Type) int64 {
	// TODO later
	fmt.Printf("[TODO] ShallowSizeOfInstance(%v)\n", clazz)
	return 0
}

/* Return shallow size of any array */
func shallowSizeOfArray(arr interface{}) int64 {
	size := int64(NUM_BYTES_ARRAY_HEADER)
	v := reflect.ValueOf(arr)
	if length := v.Len(); length > 0 {
		switch v.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
			reflect.Float64, reflect.Complex64, reflect.Complex128:
			// primitive type
			size += int64(length) * int64(v.Elem().Type().Size())
		default:
			size += int64(length * NUM_BYTES_OBJECT_REF)
		}
	}
	return AlignObjectSize(size)
}
