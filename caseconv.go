package strings2

//go:generate go run ./internal/gencase

import "unicode/utf8"

func toLowerRune(r rune) rune {
	if uint32(r) > caseMaxRune {
		return r
	}
	return r + rune(lowerValues[int(lowerIndex[r>>caseBlockShift])<<caseBlockShift|int(r)&caseBlockMask])
}

func toUpperRune(r rune) rune {
	if uint32(r) > caseMaxRune {
		return r
	}
	return r + rune(upperValues[int(upperIndex[r>>caseBlockShift])<<caseBlockShift|int(r)&caseBlockMask])
}

func foldRune(r rune) rune {
	if uint32(r) > caseMaxRune {
		return r
	}
	return r + rune(foldValues[int(foldIndex[r>>caseBlockShift])<<caseBlockShift|int(r)&caseBlockMask])
}

func isSpaceRune(r rune) bool {
	if uint32(r) <= 0xff {
		switch r {
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xa0:
			return true
		}
		return false
	}
	switch r {
	case 0x1680,
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006,
		0x2007, 0x2008, 0x2009, 0x200a, 0x2028, 0x2029,
		0x202f, 0x205f, 0x3000:
		return true
	}
	return false
}

func toLowerUnicode(s string) string {
	buf := MakeNoZeroCap(0, len(s))
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			buf = append(buf, lowerTable[c])
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		lr := toLowerRune(r)
		if lr == r {
			buf = append(buf, s[i:i+size]...)
		} else {
			buf = utf8.AppendRune(buf, lr)
		}
		i += size
	}
	return unsafeString(buf)
}

func toUpperUnicode(s string) string {
	buf := MakeNoZeroCap(0, len(s))
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			buf = append(buf, upperTable[c])
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		ur := toUpperRune(r)
		if ur == r {
			buf = append(buf, s[i:i+size]...)
		} else {
			buf = utf8.AppendRune(buf, ur)
		}
		i += size
	}
	return unsafeString(buf)
}
