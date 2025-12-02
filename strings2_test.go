package strings2

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NikoMalik/strconv2"
)

var (
	upperStr = strings.Repeat("HELLO-WORLD-123", 50)
	lowerStr = strings.Repeat("hello-world-123", 50)

	// Unicode Latin: Å vs å
	unicode1A = "HÅLL"
	unicode1B = "håll"

	// Unicode Greek: Σ vs σ vs ς (sigma has 3 folds)
	unicode2A = "ΣΕΙΣ"
	unicode2B = "σεις"

	// Unicode Cyrillic: е vs Е
	unicode3A = "Привет"
	unicode3B = "ПРИВЕТ"

	benchSink bool
)

func Benchmark_EqualFoldBytes(b *testing.B) {
	left := upperStr
	right := lowerStr

	b.Run("strings2", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			benchSink = EqualFold(left, right)
		}
	})

	b.Run("std_equalfold", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			benchSink = strings.EqualFold(upperStr, lowerStr)
		}
	})

	b.Run("Unicode1_German_my", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink = EqualFold(unicode1A, unicode1B)
			if !benchSink {
				b.Fatal("have to be true")
			}
		}

	})

	b.Run("Unicode1_German_std", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink = strings.EqualFold(unicode1A, unicode1B)
			if !benchSink {
				b.Fatal("have to be true")
			}
		}

	})

	b.Run("Unicode2_Greek_my", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink = EqualFold(unicode2A, unicode2B)
			if !benchSink {
				b.Fatal("have to be true")
			}
		}
	})

	b.Run("Unicode2_Greek_std", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink = strings.EqualFold(unicode2A, unicode2B)
			if !benchSink {
				b.Fatal("have to be true")
			}
		}

	})
}

func BenchmarkReplaceString(b *testing.B) {
	input := strings.Repeat("abc needle xyz ", 20000)
	needle := "needle"
	repl := "X"

	b.ResetTimer()
	for b.Loop() {
		ReplaceString(input, needle, repl, -1)
	}
}

func BenchmarkStringsReplace(b *testing.B) {
	input := strings.Repeat("abc needle xyz ", 20000)
	needle := "needle"
	repl := "X"

	b.ResetTimer()
	for b.Loop() {
		strings.ReplaceAll(input, needle, repl)
	}
}

func BenchmarkToLowerOriginal(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()
	for b.Loop() {
		strings.ToLower(s)
	}
}

func BenchmarkToLowerOptimized(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()
	for b.Loop() {
		ToLower(s)
	}
}

func BenchmarkToUpperOriginal(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()

	for b.Loop() {
		strings.ToUpper(s)
	}
}

func BenchmarkToUpperOptimized(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()

	for b.Loop() {
		ToUpper(s)
	}
}

func BenchmarkRepeatOriginal(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()

	for b.Loop() {
		strings.Repeat(s, 100)
	}
}

func BenchmarkRepeatOptimized(b *testing.B) {
	s := strings.Repeat("Hello World! ", 10000)
	b.ResetTimer()

	for b.Loop() {
		Repeat(s, 100)
	}
}

func BenchmarkReplaceUnicode(b *testing.B) {
	input := strings.Repeat("你好 😀 世界 😀 ", 5000)
	b.ResetTimer()
	for b.Loop() {
		ReplaceString(input, "😀", "😎", -1)
	}
}

func BenchmarkStringsReplaceUnicode(b *testing.B) {
	input := strings.Repeat("你好 😀 世界 😀 ", 5000)
	b.ResetTimer()
	for b.Loop() {
		strings.ReplaceAll(input, "😀", "😎")
	}
}

func TestReplaceFast(t *testing.T) {
	tests := []struct {
		name string
		s    string
		old  string
		new  string
		n    int
		want string
	}{
		{"english replacement", "hello hello world", "hello", "hi", -1, "hi hi world"},
		{"russian replacement", "привет привет мир", "привет", "здравствуй", -1, "здравствуй здравствуй мир"},
		{"chinese replacement", "你好 你好 世界", "你好", "您好", -1, "您好 您好 世界"},
		{"n = 1", "aaa", "a", "b", 1, "baa"},
		{"n = 2", "aaa", "a", "b", 2, "bba"},
		{"n > occurrences", "aaa", "a", "b", 10, "bbb"},
		{"n = 0", "anything", "old", "new", 0, "anything"},
		{"old == new", "test", "x", "x", -1, "test"},
		{"old empty", "abc", "", "X", -1, "abc"},
		{"overlapping", "aaaa", "aa", "x", -1, "xx"},
		{"overlapping n=1", "aaaa", "aa", "x", 1, "xaa"},
		{"no occurrences", "abcdef", "xyz", "123", -1, "abcdef"},
		{"full replacement", "oldoldold", "old", "new", -1, "newnewnew"},
		{"new longer", "one two three", "two", "два", -1, "one два three"},
		{"new shorter", " hello ", " ", "_", -1, "hello"},
		{"unicode mix", "привет ☀ мир ☀", "☀", "солнце", -1, "привет солнце мир солнце"},
		{"emoji and surrogate", "😀 😃 😄", " ", "", -1, "😀😃😄"},
		{"empty string", "", "anything", "new", -1, ""},
		{"old longer than new", "abcdabcd", "abcd", "x", -1, "xx"},
		{"new contains old", "abc", "b", "bc", -1, "abcc"},
		{"chinese sentence", "我喜欢编程 编程 很有趣", "编程", "coding", -1, "我喜欢coding coding 很有趣"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceString(tt.s, tt.old, tt.new, tt.n)
			expected := strings.Replace(tt.s, tt.old, tt.new, tt.n)
			if got != expected {
				t.Errorf("\nReplaceFast(%q, %q, %q, %d)\ngot : %q\nwant: %q",
					tt.s, tt.old, tt.new, tt.n, got, expected)
			}
		})
	}

}

func TestReplaceFast_Extra(t *testing.T) {
	tests := []struct {
		name string
		s    string
		old  string
		new  string
		n    int
	}{
		// English long
		{"english long", Repeat("hello world ", 1000), "world", "planet", -1},

		// Russian long
		{"russian long", Repeat("привет мир ", 800), "мир", "земля", -1},

		// Chinese long
		{"chinese long", Repeat("你好 世界 ", 600), "世界", "宇宙", -1},

		// Mixed languages
		{"mixed unicode", "hello 你好 привет 😀 world 世界", "world", "земля", -1},
		{"mixed unicode 2", "😀😀😀 hello 😀😀😀", "😀", "😎", -1},

		// Overlapping sequences
		{"overlap unicode", "哈哈哈哈", "哈哈", "嘿", -1},

		// Replace with empty
		{"delete english", "aaa bbb ccc aaa", "aaa", "", -1},
		{"delete russian", "тест тест тест", "тест", "", -1},
		{"delete chinese", "你好你好你好", "你", "", -1},

		// Large random-like string
		{"large random", Repeat("xabcy123", 20000), "abc", "Z", -1},

		// No matches
		{"no match unicode", "мама мыла раму", "爸爸", "父亲", -1},

		// n limiting
		{"n-limit unicode", "你好 你好 你好", "你好", "您", 1},
		{"n-limit russian", "кот кот кот", "кот", "пёс", 2},

		// new contains old (force infinite-loop test)
		{"recursive check", "a", "a", "aa", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Replace(tt.s, tt.old, tt.new, tt.n)
			got := ReplaceString(tt.s, tt.old, tt.new, tt.n)
			if got != expected {
				t.Fatalf("FAIL %s:\n want: %q\n got : %q", tt.name, expected, got)
			}
		})
	}
}

func TestToLowerExtra(t *testing.T) {
	tests := []string{
		"HELLO WORLD",
		"HeLlO WoRlD",
		"ПрИвЕт Мир",
		"你好 世界",
		"😀😃😄HELLO",
		Repeat("ABCXYZ", 2000),
	}

	for _, s := range tests {
		if ToLower(s) != strings.ToLower(s) {
			t.Fatalf("ToLower mismatch for %q", s)
		}
	}
}

func TestToUpperExtra(t *testing.T) {
	tests := []string{
		"hello world",
		"HeLlO WoRlD",
		"Привет мир",
		"你好 世界",
		"😀😃😄hello",
		strings.Repeat("abcxyz", 2000),
	}

	for _, s := range tests {
		if ToUpper(s) != strings.ToUpper(s) {
			t.Fatalf("ToUpper mismatch for %q", s)
		}
	}
}

func TestRepeatExtra(t *testing.T) {
	tests := []struct {
		s     string
		count int
	}{
		{"a", 5000},
		{"привет", 2000},
		{"你好", 3000},
		{"😀", 4000},
		{strings.Repeat("abc", 100), 100},
	}

	for _, tt := range tests {
		want := strings.Repeat(tt.s, tt.count)
		got := Repeat(tt.s, tt.count)
		if want != got {
			t.Fatalf("Repeat mismatch: %d copies of %q", tt.count, tt.s)
		}
	}
}

func TestRepeatPanicOnOverflow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on overflow")
		}
	}()
	Repeat("abc", math.MaxInt)
}

func TestBuilderResetAndKeepCap(t *testing.T) {
	t.Run("clears_len_and_keeps_cap", func(t *testing.T) {
		b := NewBuilder(16)
		b.WriteString("hello")

		oldCap := b.Cap()
		if oldCap != 16 {
			t.Fatalf("unexpected cap: %d", oldCap)
		}

		b.ResetAndKeepCap()

		if b.Len() != 0 {
			t.Fatalf("expected len=0 after ResetAndKeepCap, got %d", b.Len())
		}
		if b.Cap() != oldCap {
			t.Fatalf("expected cap=%d, got %d", oldCap, b.Cap())
		}
	})

	t.Run("memory_is_zeroed", func(t *testing.T) {
		b := NewBuilder(32)
		b.WriteString("secret123")

		b.ResetAndKeepCap()

		raw := b.buf[:cap(b.buf)]
		for i, v := range raw {
			if v != 0 {
				t.Fatalf("memory not zeroed at index %d: %v", i, raw)
			}
		}
	})

	t.Run("write_after_reset_works", func(t *testing.T) {
		b := NewBuilder(8)
		b.WriteString("abc")
		b.ResetAndKeepCap()

		b.WriteString("xyz")

		got := b.String()
		if got != "xyz" {
			t.Fatalf("expected 'xyz', got %q", got)
		}
	})

	t.Run("cap_remains_after_multiple_resets", func(t *testing.T) {
		b := NewBuilder(10)
		for i := 0; i < 5; i++ {
			b.WriteString("test")
			b.ResetAndKeepCap()
			if b.Cap() != 10 {
				t.Fatalf("cap changed after reset #%d: %d", i, b.Cap())
			}
			if b.Len() != 0 {
				t.Fatalf("len != 0 after reset #%d: %d", i, b.Len())
			}
		}
	})
}

func Test_EqualFold(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Expected bool
		S1       string
		S2       string
	}{
		{Expected: true, S1: "/MY/NAME/IS/:PARAM/*", S2: "/my/name/is/:param/*"},
		{Expected: true, S1: "/MY/NAME/IS/:PARAM/*", S2: "/my/name/is/:param/*"},
		{Expected: true, S1: "/MY1/NAME/IS/:PARAM/*", S2: "/MY1/NAME/IS/:PARAM/*"},
		{Expected: false, S1: "/my2/name/is/:param/*", S2: "/my2/name"},
		{Expected: false, S1: "/dddddd", S2: "eeeeee"},
		{Expected: false, S1: "\na", S2: "*A"},
		{Expected: true, S1: "/MY3/NAME/IS/:PARAM/*", S2: "/my3/name/is/:param/*"},
		{Expected: true, S1: "/MY4/NAME/IS/:PARAM/*", S2: "/my4/nAME/IS/:param/*"},
	}

	for _, tc := range testCases {
		got := EqualFold(tc.S1, tc.S2)
		want := strings.EqualFold(tc.S1, tc.S2)
		if want != got {
			t.Fatalf("Equal Fold: mismatch: %s:%s", tc.S1, tc.S2)

		}

	}
}

type myStringer struct{}

func (myStringer) String() string { return "STRINGER_OK" }

func TestToString(t *testing.T) {
	// ---- integers ----
	t.Run("ints", func(t *testing.T) {
		cases := map[string]any{
			"int":    int(42),
			"int8":   int8(-5),
			"int16":  int16(1234),
			"int32":  int32(-999),
			"int64":  int64(1<<40 + 123),
			"uint":   uint(77),
			"uint8":  uint8(255),
			"uint16": uint16(65000),
			"uint32": uint32(1<<31 + 2),
			"uint64": uint64(1<<50 + 7),
		}
		for name, v := range cases {
			got := ToString(v)
			want := fmt.Sprintf("%v", v)
			if got != want {
				t.Fatalf("%s: want=%q got=%q", name, want, got)
			}
		}
	})

	// ---- strings, bytes ----
	t.Run("string and bytes", func(t *testing.T) {
		if ToString("abc") != "abc" {
			t.Fatal("string failed")
		}
		if ToString([]byte("xyz")) != "xyz" {
			t.Fatal("bytes failed")
		}
	})

	// ---- bool ----
	t.Run("bool", func(t *testing.T) {
		if ToString(true) != "true" || ToString(false) != "false" {
			t.Fatal("bool failed")
		}
	})

	// ---- float ----
	t.Run("floats", func(t *testing.T) {
		if ToString(float32(3.14)) != "3.14" {
			t.Fatal("float32 failed")
		}
		if ToString(float64(2.5)) != "2.5" {
			t.Fatal("float64 failed")
		}
	})

	// ---- time ----
	t.Run("time default", func(t *testing.T) {
		tm := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		got := ToString(tm)
		want := "2020-01-02 03:04:05"
		if got != want {
			t.Fatalf("time default: want=%q got=%q", want, got)
		}
	})

	t.Run("time formatted", func(t *testing.T) {
		tm := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		got := ToString(tm, time.RFC3339)
		want := tm.Format(time.RFC3339)
		if got != want {
			t.Fatalf("time formatted: want=%q got=%q", want, got)
		}
	})

	// ---- Stringer ----
	t.Run("stringer", func(t *testing.T) {
		if ToString(myStringer{}) != "STRINGER_OK" {
			t.Fatal("stringer failed")
		}
	})

	// ---- reflect.Value ----
	t.Run("reflect value", func(t *testing.T) {
		v := reflect.ValueOf(123)
		if ToString(v) != "123" {
			t.Fatal("reflect.Value failed")
		}
	})

	// ---- pointer deref ----
	t.Run("pointer", func(t *testing.T) {
		x := 777
		if ToString(&x) != "777" {
			t.Fatal("pointer failed")
		}
	})

	// ---- slice / array ----
	t.Run("slice", func(t *testing.T) {
		got := ToString([]int{1, 2, 3})
		want := "[1 2 3]"
		if got != want {
			t.Fatalf("slice: want=%q got=%q", want, got)
		}

		got = ToString([3]string{"a", "b", "c"})
		want = "[a b c]"
		if got != want {
			t.Fatalf("array: want=%q got=%q", want, got)
		}
	})

	// ---- fallback ----
	t.Run("fallback", func(t *testing.T) {
		type X struct{ A int }
		x := X{A: 7}
		got := ToString(x)
		want := fmt.Sprint(x)
		if got != want {
			t.Fatalf("fallback: want=%q got=%q", want, got)
		}
	})
}

func TestToStringStackEscape(t *testing.T) {
	var s string

	func() {
		s = ToString(123456789)
	}()

	var big [4096]byte
	for i := range big {
		big[i] = byte(i)
	}

	if s != "123456789" {
		t.Fatalf("stack escape broken: got=%q", s)
	}
}

func TestToStringByteAlias(t *testing.T) {
	b := []byte("hello")
	s := ToString(b)

	b[1] = 'a'

	if s != "hallo" {
		t.Fatalf("string alias broken: got=%q", s)
	}
}

func TestToStringSliceSafe(t *testing.T) {
	sl := []int{1, 2, 3}

	s := ToString(sl)
	sl[1] = 999

	if s != "[1 2 3]" {
		t.Fatalf("slice safe failed: got=%q", s)
	}
}

func BenchmarkToStringInt(b *testing.B) {
	x := 123456
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = ToString(x)
	}
}

func BenchmarkStrconv2Int(b *testing.B) {
	x := 123456
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		var buf [strconv2.SAFETY_BUF_SIZE]byte
		n := strconv2.FormatInt6410(buf[:], int64(x))
		_ = unsafeString(buf[:n])
	}
}

func BenchmarkItoaInt(b *testing.B) {
	x := 123456
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = strconv.Itoa(x)
	}
}

func BenchmarkSprintfInt(b *testing.B) {
	x := 123456
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = fmt.Sprintf("%d", x)
	}
}
