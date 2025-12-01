package strings2

import (
	"math"
	"strings"
	"testing"
)

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
