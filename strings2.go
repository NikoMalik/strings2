package strings2

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/NikoMalik/strconv2"
)

// According to static analysis, spaces, dashes, zeros, equals, and tabs
// are the most commonly repeated string literal,
// often used for display on fixed-width terminal windows.
// Pre-declare constants for these for O(1) repetition in the common-case.
const (
	repeatedSpaces = "" +
		"                                                                " +
		"                                                                "
	repeatedDashes = "" +
		"----------------------------------------------------------------" +
		"----------------------------------------------------------------"
	repeatedZeroes = "" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	repeatedEquals = "" +
		"================================================================" +
		"================================================================"
	repeatedTabs = "" +
		"\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t" +
		"\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
)

func ReplaceAll(s, old, new string) string {
	return ReplaceString(s, old, new, -1)
}

func ReplaceString(s, old, new string, n int) string {
	if n == 0 || old == new || len(s) == 0 {
		return s
	}
	if n < 0 {
		n = 1 << 30
	}

	oldLen := len(old)
	if oldLen == 0 {
		return replaceEmptyOld(s, new, n)
	}

	oldb := unsafeBytes(old)
	newb := unsafeBytes(new)
	newLen := len(new)
	delta := newLen - oldLen

	sbRead := unsafeBytes(s)

	var sb []byte // lazy modifiable copy

	pos := 0
	count := 0
	writePos := 0

	if delta <= 0 {
		// In-place:  (delta <=0)
		for count < n {
			idx := findIndex(sbRead, oldb, oldLen, pos)
			if idx == -1 {
				break
			}
			idx += pos

			if sb == nil {
				sb = make([]byte, len(s))
				copy(sb, sbRead)
			}

			segmentLen := idx - pos
			copy(sb[writePos:], sb[pos:pos+segmentLen])
			writePos += segmentLen

			copy(sb[writePos:], newb)
			writePos += newLen

			pos = idx + oldLen
			count++
		}

		if count == 0 {
			return s //zerocalloc
		}

		remaining := len(s) - pos
		copy(sb[writePos:], sbRead[pos:pos+remaining])
		writePos += remaining

		return unsafeString(sb[:writePos])
	} else {
		// delta >0: new buffer, count
		actualReplaces := countLimited(sbRead, oldb, oldLen, n)
		if actualReplaces == 0 {
			return s
		}

		newSize := len(s) + actualReplaces*delta
		result := make([]byte, newSize)

		pos = 0
		currPos := 0
		prevEnd := 0
		for i := 0; i < actualReplaces; i++ {
			idx := findIndex(sbRead, oldb, oldLen, pos)
			idx += pos

			copy(result[currPos:], sbRead[prevEnd:idx])
			currPos += idx - prevEnd

			copy(result[currPos:], newb)
			currPos += newLen

			prevEnd = idx + oldLen
			pos = prevEnd
		}

		copy(result[currPos:], sbRead[prevEnd:])

		return unsafeString(result)
	}
}

// old == ""
func replaceEmptyOld(s, new string, n int) string {
	m := utf8.RuneCountInString(s)
	maxInserts := m + 1
	if n < 0 || n > maxInserts {
		n = maxInserts
	}
	if n == 0 {
		return s
	}

	sbRead := unsafeBytes(s)
	newb := unsafeBytes(new)
	newSize := len(s) + n*len(new)

	result := make([]byte, newSize)

	currPos := 0
	if n > 0 {
		copy(result[currPos:], newb)
		currPos += len(new)
		n--
	}
	start := 0
	for i := 0; i < m && n > 0; i++ {
		_, wid := utf8.DecodeRuneInString(s[start:])
		copy(result[currPos:], sbRead[start:start+wid])
		currPos += wid
		copy(result[currPos:], newb)
		currPos += len(new)
		start += wid
		n--
	}

	copy(result[currPos:], sbRead[start:])
	currPos += len(s) - start

	for ; n > 0; n-- {
		copy(result[currPos:], newb)
		currPos += len(new)
	}

	return unsafeString(result[:currPos])
}

func findIndex(sb []byte, oldb []byte, oldLen int, pos int) int {
	if oldLen == 1 {
		for i := pos; i < len(sb); i++ {
			if sb[i] == oldb[0] {
				return i - pos
			}
		}
		return -1
	}
	return bytes.Index(sb[pos:], oldb)
}

// countLimited:  (non-overlapping)
func countLimited(sb []byte, oldb []byte, oldLen int, n int) int {
	count := 0
	pos := 0
	for count < n {
		idx := findIndex(sb, oldb, oldLen, pos)
		if idx == -1 {
			break
		}
		pos += idx + oldLen
		count++
	}
	return count
}

var lowerTable = func() [256]byte {
	var table [256]byte
	for i := range table {
		if 'A' <= i && i <= 'Z' {
			table[i] = byte(i + ('a' - 'A'))
		} else {
			table[i] = byte(i)
		}
	}
	return table
}()

var upperTable = func() [256]byte {
	var table [256]byte
	for i := range table {
		if 'a' <= i && i <= 'z' {
			table[i] = byte(i - ('a' - 'A'))
		} else {
			table[i] = byte(i)
		}
	}
	return table
}()

func ToLower(s string) string {
	isASCII, hasUpper := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasUpper = hasUpper || ('A' <= c && c <= 'Z')
	}
	if isASCII {
		if !hasUpper {
			return s
		}
		var b = NewBuilder(len(s))
		for i := 0; i < len(s); i++ {
			b.WriteByte(lowerTable[s[i]])
		}
		return b.String()
	}
	return strings.Map(unicode.ToLower, s)
}

func ToUpper(s string) string {
	isASCII, hasLower := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || ('a' <= c && c <= 'z')
	}
	if isASCII {
		if !hasLower {
			return s
		}
		var b = NewBuilder(len(s))
		for i := 0; i < len(s); i++ {
			b.WriteByte(upperTable[s[i]])
		}
		return b.String()
	}
	return strings.Map(unicode.ToUpper, s)
}

// Repeat returns a new string consisting of count copies of the string s.
//
// It panics if count is negative or if the result of (len(s) * count)
// overflows.
func Repeat(s string, count int) string {
	switch count {
	case 0:
		return ""
	case 1:
		return s
	}

	// Since we cannot return an error on overflow,
	// we should panic if the repeat will generate an overflow.
	//
	// See golang.org/issue/16237.
	if count < 0 {
		panic("strings: negative Repeat count")
	}
	hi, lo := bits.Mul(uint(len(s)), uint(count))
	if hi > 0 || lo > uint(math.MaxInt) {
		panic("strings: Repeat output length overflow")
	}
	n := int(lo) // lo = len(s) * count

	if len(s) == 0 {
		return ""
	}

	// Optimize for commonly repeated strings of relatively short length.
	switch s[0] {
	case ' ', '-', '0', '=', '\t':
		switch {
		case n <= len(repeatedSpaces) && strings.HasPrefix(repeatedSpaces, s):
			return repeatedSpaces[:n]
		case n <= len(repeatedDashes) && strings.HasPrefix(repeatedDashes, s):
			return repeatedDashes[:n]
		case n <= len(repeatedZeroes) && strings.HasPrefix(repeatedZeroes, s):
			return repeatedZeroes[:n]
		case n <= len(repeatedEquals) && strings.HasPrefix(repeatedEquals, s):
			return repeatedEquals[:n]
		case n <= len(repeatedTabs) && strings.HasPrefix(repeatedTabs, s):
			return repeatedTabs[:n]
		}
	}

	// Past a certain chunk size it is counterproductive to use
	// larger chunks as the source of the write, as when the source
	// is too large we are basically just thrashing the CPU D-cache.
	// So if the result length is larger than an empirically-found
	// limit (8KB), we stop growing the source string once the limit
	// is reached and keep reusing the same source string - that
	// should therefore be always resident in the L1 cache - until we
	// have completed the construction of the result.
	// This yields significant speedups (up to +100%) in cases where
	// the result length is large (roughly, over L2 cache size).
	const chunkLimit = 8 * 1024
	chunkMax := n
	if n > chunkLimit {
		chunkMax = chunkLimit / len(s) * len(s)
		if chunkMax == 0 {
			chunkMax = len(s)
		}
	}

	var b = NewBuilder(n)
	b.WriteString(s)
	for b.Len() < n {
		chunk := min(n-b.Len(), b.Len(), chunkMax)
		b.WriteString(b.String()[:chunk])
	}
	return b.String()
}

func EqualFold(b, s string) bool {
	if len(b) != len(s) {
		return false
	}

	table := upperTable
	n := len(b)
	i := 0

	// Unroll by 4
	limit := n &^ 3
	for i < limit {
		b0 := b[i+0]
		s0 := s[i+0]
		if b0|s0 >= utf8.RuneSelf {
			goto hasUnicode
		}
		if table[b0] != table[s0] {
			return false
		}

		b1 := b[i+1]
		s1 := s[i+1]
		if b1|s1 >= utf8.RuneSelf {
			goto hasUnicode
		}
		if table[b1] != table[s1] {
			return false
		}

		b2 := b[i+2]
		s2 := s[i+2]
		if b2|s2 >= utf8.RuneSelf {
			goto hasUnicode
		}
		if table[b2] != table[s2] {
			return false
		}

		b3 := b[i+3]
		s3 := s[i+3]
		if b3|s3 >= utf8.RuneSelf {
			goto hasUnicode
		}
		if table[b3] != table[s3] {
			return false
		}

		i += 4
	}

	for i < n {
		bi := b[i]
		si := s[i]
		if bi|si >= utf8.RuneSelf {
			goto hasUnicode
		}
		if table[b[i]] != table[s[i]] {
			return false
		}
		i++
	}
	return true

hasUnicode:
	// Fall back to Unicode-aware path.
	// Trim processed part.
	b = b[i:]
	s = s[i:]

	for len(b) > 0 {
		if len(s) == 0 {
			return false
		}

		var br, sr rune
		var bs, ss int

		// decode b rune
		if b[0] < utf8.RuneSelf {
			br = rune(b[0])
			bs = 1
		} else {
			br, bs = utf8.DecodeRune(unsafeBytes(b))
		}

		// decode s rune
		if s[0] < utf8.RuneSelf {
			sr = rune(s[0])
			ss = 1
		} else {
			sr, ss = utf8.DecodeRune(unsafeBytes(s))
		}

		b = b[bs:]
		s = s[ss:]

		// Fast match
		if br == sr {
			continue
		}

		// Make br < sr
		if sr < br {
			sr, br = br, sr
		}

		// ASCII fast case
		if sr < utf8.RuneSelf {
			if 'A' <= br && br <= 'Z' && sr == br+'a'-'A' {
				continue
			}
			return false
		}

		// unicode.SimpleFold
		r := unicode.SimpleFold(br)
		for r != br && r < sr {
			r = unicode.SimpleFold(r)
		}
		if r == sr {
			continue
		}

		return false
	}

	return len(s) == 0
}

// ToString Change arg to string
func ToString(arg any, timeFormat ...string) string {
	switch v := arg.(type) {
	case int:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(v))
		return unsafeString(buf[:n])
	case int8:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(v))
		return unsafeString(buf[:n])

	case int16:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(v))
		return unsafeString(buf[:n])
	case int32:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(v))
		return unsafeString(buf[:n])
	case int64:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(v))
		return unsafeString(buf[:n])
	case uint:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatUint6410(buf[:], uint64(v))
		return unsafeString(buf[:n])
	case uint8:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatUint6410(buf[:], uint64(v))
		return unsafeString(buf[:n])
	case uint16:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatUint16(buf[:], v)
		return unsafeString(buf[:n])
	case uint32:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatUint6410(buf[:], uint64(v))
		return unsafeString(buf[:n])

	case uint64:
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatUint6410(buf[:], uint64(v))
		return unsafeString(buf[:n])
	case string:
		return v
	case []byte:
		return unsafeString(v)
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		if len(timeFormat) > 0 {
			return v.Format(timeFormat[0])
		}
		return v.Format("2006-01-02 15:04:05")
	case reflect.Value:
		return ToString(v.Interface(), timeFormat...)
	case fmt.Stringer:
		return v.String()
	default:
		rv := reflect.ValueOf(arg)
		if rv.Kind() == reflect.Pointer && !rv.IsNil() {
			return ToString(rv.Elem().Interface(), timeFormat...)
		} else if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			// handle slices
			var buf = NewBuilder(rv.Len())
			buf.WriteString("[") //nolint:errcheck // no need to check error
			for i := 0; i < rv.Len(); i++ {
				if i > 0 {
					buf.WriteString(" ") //nolint:errcheck // no need to check error
				}
				buf.WriteString(ToString(rv.Index(i).Interface())) //nolint:errcheck // no need to check error
			}
			buf.WriteString("]") //nolint:errcheck // no need to check error
			return buf.String()
		}

		return fmt.Sprint(arg)
	}
}
