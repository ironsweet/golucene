package util

// util/UnicodeUtil.java

const (
	UNI_SUR_HIGH_START = 0xD800
	UNI_SUR_HIGH_END   = 0xDBFF
	UNI_SUR_LOW_START  = 0xDC00
	UNI_SUR_LOW_END    = 0xDFFF
)

// L345
func IsValidUTF16String(s []rune) bool {
	size := len(s)
	for i, ch := range s {
		if ch >= UNI_SUR_HIGH_START && ch <= UNI_SUR_HIGH_END {
			if i < size-1 {
				i++
				nextCh := s[i]
				if nextCh >= UNI_SUR_LOW_START && nextCh <= UNI_SUR_LOW_END {
					// Valid surrogate pair
				} else {
					// Unmatched high surrogate
					return false
				}
			} else {
				// Unmatched high surrogate
				return false
			}
		} else if ch >= UNI_SUR_LOW_START && ch <= UNI_SUR_LOW_END {
			// Unmatched low surrogate
			return false
		}
	}

	return true
}
