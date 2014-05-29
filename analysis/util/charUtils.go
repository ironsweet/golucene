package util

import (
	"github.com/balzaczyy/golucene/core/util"
	"unicode"
)

// util/CharacterUtils.java

type CharacterUtilsSPI interface{}

/*
Characterutils provides a unified interface to Character-related
operations to implement backwards compatible character operations
based on a version instance.
*/
type CharacterUtils struct {
	CharacterUtilsSPI
}

/* Returns a Characters implementation according to the given Version instance. */
func GetCharacterUtils(matchVersion util.Version) *CharacterUtils {
	return &CharacterUtils{}
}

/* Converts each unicode codepoint to lowerCase via unicode.ToLower(). */
func (cu *CharacterUtils) ToLowerCase(buffer []rune) {
	for i, v := range buffer {
		buffer[i] = unicode.ToLower(v)
	}
}
