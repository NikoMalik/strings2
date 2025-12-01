package strings2

import "unsafe"

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer

func MakeNoZero(l int) []byte {
	return unsafe.Slice((*byte)(mallocgc(uintptr(l), nil, false)), l)
}

func MakeNoZeroCap(l int, c int) []byte {
	return MakeNoZero(c)[:l]
}

// unsafeBytes: (read-only!)
func unsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(p unsafe.Pointer, n uintptr)

// MemclrZero sets memory of slice to zero, assuming T has no heap pointers.
// T MUST NOT contain any references (e.g. pointers, strings, slices, maps, funcs).
func memclr[T any](s []T) {
	if len(s) == 0 {
		return
	}
	size := unsafe.Sizeof(s[0]) * uintptr(len(s))
	ptr := unsafe.Pointer(&s[0])
	memclrNoHeapPointers(ptr, size)
}
