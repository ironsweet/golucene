package standard

import (
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"io"
)

// standard/StandardTokenizerImpl.java

/* initial size of the lookahead buffer */
const ZZ_BUFFERSIZE = 255

/* lexical states */
const YYINITIAL = 0

/*
ZZ_LEXSTATE[l] is the state in the DFA for the lexical state l
ZZ_LEXSTATE[l+1] is the state in the DFA for the lexical state l at the beginning of a line
l is of the form l = 2*k, k a non negative integer
*/
var ZZ_LEXSTATE = [2]int{0, 0}

/* Translates characters to character classes */
var ZZ_CMAP_PACKED = []int{
	042, 000, 001, 015, 004, 000, 001, 014, 004, 000, 001, 007, 001, 000, 001, 010, 001, 000, 012, 004,
	001, 006, 001, 007, 005, 000, 032, 001, 004, 000, 001, 011, 001, 000, 032, 001, 057, 000, 001, 001,
	002, 000, 001, 003, 007, 000, 001, 001, 001, 000, 001, 006, 002, 000, 001, 001, 005, 000, 027, 001,
	001, 000, 037, 001, 001, 000, int('\u01ca'), 001, 004, 000, 014, 001, 005, 000, 001, 006, 010, 000, 005, 001,
	007, 000, 001, 001, 001, 000, 001, 001, 021, 000, 0160, 003, 005, 001, 001, 000, 002, 001, 002, 000,
	004, 001, 001, 007, 007, 000, 001, 001, 001, 006, 003, 001, 001, 000, 001, 001, 001, 000, 024, 001,
	001, 000, 0123, 001, 001, 000, 0213, 001, 001, 000, 007, 003, 0236, 001, 011, 000, 046, 001, 002, 000,
	001, 001, 007, 000, 047, 001, 001, 000, 001, 007, 007, 000, 055, 003, 001, 000, 001, 003, 001, 000,
	002, 003, 001, 000, 002, 003, 001, 000, 001, 003, 010, 000, 033, 016, 005, 000, 003, 016, 001, 001,
	001, 006, 013, 000, 005, 003, 007, 000, 002, 007, 002, 000, 013, 003, 001, 000, 001, 003, 003, 000,
	053, 001, 025, 003, 012, 004, 001, 000, 001, 004, 001, 007, 001, 000, 002, 001, 001, 003, 0143, 001,
	001, 000, 001, 001, 010, 003, 001, 000, 006, 003, 002, 001, 002, 003, 001, 000, 004, 003, 002, 001,
	012, 004, 003, 001, 002, 000, 001, 001, 017, 000, 001, 003, 001, 001, 001, 003, 036, 001, 033, 003,
	002, 000, 0131, 001, 013, 003, 001, 001, 016, 000, 012, 004, 041, 001, 011, 003, 002, 001, 002, 000,
	001, 007, 001, 000, 001, 001, 005, 000, 026, 001, 004, 003, 001, 001, 011, 003, 001, 001, 003, 003,
	001, 001, 005, 003, 022, 000, 031, 001, 003, 003, 0104, 000, 001, 001, 001, 000, 013, 001, 067, 000,
	033, 003, 001, 000, 004, 003, 066, 001, 003, 003, 001, 001, 022, 003, 001, 001, 007, 003, 012, 001,
	002, 003, 002, 000, 012, 004, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 003, 003, 001, 000,
	010, 001, 002, 000, 002, 001, 002, 000, 026, 001, 001, 000, 007, 001, 001, 000, 001, 001, 003, 000,
	004, 001, 002, 000, 001, 003, 001, 001, 007, 003, 002, 000, 002, 003, 002, 000, 003, 003, 001, 001,
	010, 000, 001, 003, 004, 000, 002, 001, 001, 000, 003, 001, 002, 003, 002, 000, 012, 004, 002, 001,
	017, 000, 003, 003, 001, 000, 006, 001, 004, 000, 002, 001, 002, 000, 026, 001, 001, 000, 007, 001,
	001, 000, 002, 001, 001, 000, 002, 001, 001, 000, 002, 001, 002, 000, 001, 003, 001, 000, 005, 003,
	004, 000, 002, 003, 002, 000, 003, 003, 003, 000, 001, 003, 007, 000, 004, 001, 001, 000, 001, 001,
	007, 000, 012, 004, 002, 003, 003, 001, 001, 003, 013, 000, 003, 003, 001, 000, 011, 001, 001, 000,
	003, 001, 001, 000, 026, 001, 001, 000, 007, 001, 001, 000, 002, 001, 001, 000, 005, 001, 002, 000,
	001, 003, 001, 001, 010, 003, 001, 000, 003, 003, 001, 000, 003, 003, 002, 000, 001, 001, 017, 000,
	002, 001, 002, 003, 002, 000, 012, 004, 021, 000, 003, 003, 001, 000, 010, 001, 002, 000, 002, 001,
	002, 000, 026, 001, 001, 000, 007, 001, 001, 000, 002, 001, 001, 000, 005, 001, 002, 000, 001, 003,
	001, 001, 007, 003, 002, 000, 002, 003, 002, 000, 003, 003, 010, 000, 002, 003, 004, 000, 002, 001,
	001, 000, 003, 001, 002, 003, 002, 000, 012, 004, 001, 000, 001, 001, 020, 000, 001, 003, 001, 001,
	001, 000, 006, 001, 003, 000, 003, 001, 001, 000, 004, 001, 003, 000, 002, 001, 001, 000, 001, 001,
	001, 000, 002, 001, 003, 000, 002, 001, 003, 000, 003, 001, 003, 000, 014, 001, 004, 000, 005, 003,
	003, 000, 003, 003, 001, 000, 004, 003, 002, 000, 001, 001, 006, 000, 001, 003, 016, 000, 012, 004,
	021, 000, 003, 003, 001, 000, 010, 001, 001, 000, 003, 001, 001, 000, 027, 001, 001, 000, 012, 001,
	001, 000, 005, 001, 003, 000, 001, 001, 007, 003, 001, 000, 003, 003, 001, 000, 004, 003, 007, 000,
	002, 003, 001, 000, 002, 001, 006, 000, 002, 001, 002, 003, 002, 000, 012, 004, 022, 000, 002, 003,
	001, 000, 010, 001, 001, 000, 003, 001, 001, 000, 027, 001, 001, 000, 012, 001, 001, 000, 005, 001,
	002, 000, 001, 003, 001, 001, 007, 003, 001, 000, 003, 003, 001, 000, 004, 003, 007, 000, 002, 003,
	007, 000, 001, 001, 001, 000, 002, 001, 002, 003, 002, 000, 012, 004, 001, 000, 002, 001, 017, 000,
	002, 003, 001, 000, 010, 001, 001, 000, 003, 001, 001, 000, 051, 001, 002, 000, 001, 001, 007, 003,
	001, 000, 003, 003, 001, 000, 004, 003, 001, 001, 010, 000, 001, 003, 010, 000, 002, 001, 002, 003,
	002, 000, 012, 004, 012, 000, 006, 001, 002, 000, 002, 003, 001, 000, 022, 001, 003, 000, 030, 001,
	001, 000, 011, 001, 001, 000, 001, 001, 002, 000, 007, 001, 003, 000, 001, 003, 004, 000, 006, 003,
	001, 000, 001, 003, 001, 000, 010, 003, 022, 000, 002, 003, 015, 000, 060, 020, 001, 021, 002, 020,
	007, 021, 005, 000, 007, 020, 010, 021, 001, 000, 012, 004, 047, 000, 002, 020, 001, 000, 001, 020,
	002, 000, 002, 020, 001, 000, 001, 020, 002, 000, 001, 020, 006, 000, 004, 020, 001, 000, 007, 020,
	001, 000, 003, 020, 001, 000, 001, 020, 001, 000, 001, 020, 002, 000, 002, 020, 001, 000, 004, 020,
	001, 021, 002, 020, 006, 021, 001, 000, 002, 021, 001, 020, 002, 000, 005, 020, 001, 000, 001, 020,
	001, 000, 006, 021, 002, 000, 012, 004, 002, 000, 004, 020, 040, 000, 001, 001, 027, 000, 002, 003,
	006, 000, 012, 004, 013, 000, 001, 003, 001, 000, 001, 003, 001, 000, 001, 003, 004, 000, 002, 003,
	010, 001, 001, 000, 044, 001, 004, 000, 024, 003, 001, 000, 002, 003, 005, 001, 013, 003, 001, 000,
	044, 003, 011, 000, 001, 003, 071, 000, 053, 020, 024, 021, 001, 020, 012, 004, 006, 000, 006, 020,
	004, 021, 004, 020, 003, 021, 001, 020, 003, 021, 002, 020, 007, 021, 003, 020, 004, 021, 015, 020,
	014, 021, 001, 020, 001, 021, 012, 004, 004, 021, 002, 020, 046, 001, 001, 000, 001, 001, 005, 000,
	001, 001, 002, 000, 053, 001, 001, 000, 004, 001, int('\u0100'), 002, 0111, 001, 001, 000, 004, 001, 002, 000,
	007, 001, 001, 000, 001, 001, 001, 000, 004, 001, 002, 000, 051, 001, 001, 000, 004, 001, 002, 000,
	041, 001, 001, 000, 004, 001, 002, 000, 007, 001, 001, 000, 001, 001, 001, 000, 004, 001, 002, 000,
	017, 001, 001, 000, 071, 001, 001, 000, 004, 001, 002, 000, 0103, 001, 002, 000, 003, 003, 040, 000,
	020, 001, 020, 000, 0125, 001, 014, 000, int('\u026c'), 001, 002, 000, 021, 001, 001, 000, 032, 001, 005, 000,
	0113, 001, 003, 000, 003, 001, 017, 000, 015, 001, 001, 000, 004, 001, 003, 003, 013, 000, 022, 001,
	003, 003, 013, 000, 022, 001, 002, 003, 014, 000, 015, 001, 001, 000, 003, 001, 001, 000, 002, 003,
	014, 000, 064, 020, 040, 021, 003, 000, 001, 020, 004, 000, 001, 020, 001, 021, 002, 000, 012, 004,
	041, 000, 004, 003, 001, 000, 012, 004, 006, 000, 0130, 001, 010, 000, 051, 001, 001, 003, 001, 001,
	005, 000, 0106, 001, 012, 000, 035, 001, 003, 000, 014, 003, 004, 000, 014, 003, 012, 000, 012, 004,
	036, 020, 002, 000, 005, 020, 013, 000, 054, 020, 004, 000, 021, 021, 007, 020, 002, 021, 006, 000,
	012, 004, 001, 020, 003, 000, 002, 020, 040, 000, 027, 001, 005, 003, 004, 000, 065, 020, 012, 021,
	001, 000, 035, 021, 002, 000, 001, 003, 012, 004, 006, 000, 012, 004, 006, 000, 016, 020, 0122, 000,
	005, 003, 057, 001, 021, 003, 007, 001, 004, 000, 012, 004, 021, 000, 011, 003, 014, 000, 003, 003,
	036, 001, 015, 003, 002, 001, 012, 004, 054, 001, 016, 003, 014, 000, 044, 001, 024, 003, 010, 000,
	012, 004, 003, 000, 003, 001, 012, 004, 044, 001, 0122, 000, 003, 003, 001, 000, 025, 003, 004, 001,
	001, 003, 004, 001, 003, 003, 002, 001, 011, 000, 0300, 001, 047, 003, 025, 000, 004, 003, int('\u0116'), 001,
	002, 000, 006, 001, 002, 000, 046, 001, 002, 000, 006, 001, 002, 000, 010, 001, 001, 000, 001, 001,
	001, 000, 001, 001, 001, 000, 001, 001, 001, 000, 037, 001, 002, 000, 065, 001, 001, 000, 007, 001,
	001, 000, 001, 001, 003, 000, 003, 001, 001, 000, 007, 001, 003, 000, 004, 001, 002, 000, 006, 001,
	004, 000, 015, 001, 005, 000, 003, 001, 001, 000, 007, 001, 017, 000, 004, 003, 010, 000, 002, 010,
	012, 000, 001, 010, 002, 000, 001, 006, 002, 000, 005, 003, 020, 000, 002, 011, 003, 000, 001, 007,
	017, 000, 001, 011, 013, 000, 005, 003, 001, 000, 012, 003, 001, 000, 001, 001, 015, 000, 001, 001,
	020, 000, 015, 001, 063, 000, 041, 003, 021, 000, 001, 001, 004, 000, 001, 001, 002, 000, 012, 001,
	001, 000, 001, 001, 003, 000, 005, 001, 006, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001,
	001, 000, 004, 001, 001, 000, 013, 001, 002, 000, 004, 001, 005, 000, 005, 001, 004, 000, 001, 001,
	021, 000, 051, 001, int('\u032d'), 000, 064, 001, int('\u0716'), 000, 057, 001, 001, 000, 057, 001, 001, 000, 0205, 001,
	006, 000, 004, 001, 003, 003, 002, 001, 014, 000, 046, 001, 001, 000, 001, 001, 005, 000, 001, 001,
	002, 000, 070, 001, 007, 000, 001, 001, 017, 000, 001, 003, 027, 001, 011, 000, 007, 001, 001, 000,
	007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000, 007, 001, 001, 000,
	007, 001, 001, 000, 007, 001, 001, 000, 040, 003, 057, 000, 001, 001, 0120, 000, 032, 012, 001, 000,
	0131, 012, 014, 000, 0326, 012, 057, 000, 001, 001, 001, 000, 001, 012, 031, 000, 011, 012, 006, 003,
	001, 000, 005, 005, 002, 000, 003, 012, 001, 001, 001, 001, 004, 000, 0126, 013, 002, 000, 002, 003,
	002, 005, 003, 013, 0133, 005, 001, 000, 004, 005, 005, 000, 051, 001, 003, 000, 0136, 002, 021, 000,
	033, 001, 065, 000, 020, 005, 0320, 000, 057, 005, 001, 000, 0130, 005, 0250, 000, int('\u19b6'), 012, 0112, 000,
	int('\u51cd'), 012, 063, 000, int('\u048d'), 001, 0103, 000, 056, 001, 002, 000, int('\u010d'), 001, 003, 000, 020, 001, 012, 004,
	002, 001, 024, 000, 057, 001, 004, 003, 001, 000, 012, 003, 001, 000, 031, 001, 007, 000, 001, 003,
	0120, 001, 002, 003, 045, 000, 011, 001, 002, 000, 0147, 001, 002, 000, 004, 001, 001, 000, 004, 001,
	014, 000, 013, 001, 0115, 000, 012, 001, 001, 003, 003, 001, 001, 003, 004, 001, 001, 003, 027, 001,
	005, 003, 030, 000, 064, 001, 014, 000, 002, 003, 062, 001, 021, 003, 013, 000, 012, 004, 006, 000,
	022, 003, 006, 001, 003, 000, 001, 001, 004, 000, 012, 004, 034, 001, 010, 003, 002, 000, 027, 001,
	015, 003, 014, 000, 035, 002, 003, 000, 004, 003, 057, 001, 016, 003, 016, 000, 001, 001, 012, 004,
	046, 000, 051, 001, 016, 003, 011, 000, 003, 001, 001, 003, 010, 001, 002, 003, 002, 000, 012, 004,
	006, 000, 033, 020, 001, 021, 004, 000, 060, 020, 001, 021, 001, 020, 003, 021, 002, 020, 002, 021,
	005, 020, 002, 021, 001, 020, 001, 021, 001, 020, 030, 000, 005, 020, 013, 001, 005, 003, 002, 000,
	003, 001, 002, 003, 012, 000, 006, 001, 002, 000, 006, 001, 002, 000, 006, 001, 011, 000, 007, 001,
	001, 000, 007, 001, 0221, 000, 043, 001, 010, 003, 001, 000, 002, 003, 002, 000, 012, 004, 006, 000,
	int('\u2ba4'), 002, 014, 000, 027, 002, 004, 000, 061, 002, int('\u2104'), 000, int('\u016e'), 012, 002, 000, 0152, 012, 046, 000,
	007, 001, 014, 000, 005, 001, 005, 000, 001, 016, 001, 003, 012, 016, 001, 000, 015, 016, 001, 000,
	005, 016, 001, 000, 001, 016, 001, 000, 002, 016, 001, 000, 002, 016, 001, 000, 012, 016, 0142, 001,
	041, 000, int('\u016b'), 001, 022, 000, 0100, 001, 002, 000, 066, 001, 050, 000, 014, 001, 004, 000, 020, 003,
	001, 007, 002, 000, 001, 006, 001, 007, 013, 000, 007, 003, 014, 000, 002, 011, 030, 000, 003, 011,
	001, 007, 001, 000, 001, 010, 001, 000, 001, 007, 001, 006, 032, 000, 005, 001, 001, 000, 0207, 001,
	002, 000, 001, 003, 007, 000, 001, 010, 004, 000, 001, 007, 001, 000, 001, 010, 001, 000, 012, 004,
	001, 006, 001, 007, 005, 000, 032, 001, 004, 000, 001, 011, 001, 000, 032, 001, 013, 000, 070, 005,
	002, 003, 037, 002, 003, 000, 006, 002, 002, 000, 006, 002, 002, 000, 006, 002, 002, 000, 003, 002,
	034, 000, 003, 003, 004, 000, 014, 001, 001, 000, 032, 001, 001, 000, 023, 001, 001, 000, 002, 001,
	001, 000, 017, 001, 002, 000, 016, 001, 042, 000, 0173, 001, 0105, 000, 065, 001, 0210, 000, 001, 003,
	0202, 000, 035, 001, 003, 000, 061, 001, 057, 000, 037, 001, 021, 000, 033, 001, 065, 000, 036, 001,
	002, 000, 044, 001, 004, 000, 010, 001, 001, 000, 005, 001, 052, 000, 0236, 001, 002, 000, 012, 004,
	int('\u0356'), 000, 006, 001, 002, 000, 001, 001, 001, 000, 054, 001, 001, 000, 002, 001, 003, 000, 001, 001,
	002, 000, 027, 001, 0252, 000, 026, 001, 012, 000, 032, 001, 0106, 000, 070, 001, 006, 000, 002, 001,
	0100, 000, 001, 001, 003, 003, 001, 000, 002, 003, 005, 000, 004, 003, 004, 001, 001, 000, 003, 001,
	001, 000, 033, 001, 004, 000, 003, 003, 004, 000, 001, 003, 040, 000, 035, 001, 0203, 000, 066, 001,
	012, 000, 026, 001, 012, 000, 023, 001, 0215, 000, 0111, 001, int('\u03b7'), 000, 003, 003, 065, 001, 017, 003,
	037, 000, 012, 004, 020, 000, 003, 003, 055, 001, 013, 003, 002, 000, 001, 003, 022, 000, 031, 001,
	007, 000, 012, 004, 006, 000, 003, 003, 044, 001, 016, 003, 001, 000, 012, 004, 0100, 000, 003, 003,
	060, 001, 016, 003, 004, 001, 013, 000, 012, 004, int('\u04a6'), 000, 053, 001, 015, 003, 010, 000, 012, 004,
	int('\u0936'), 000, int('\u036f'), 001, 0221, 000, 0143, 001, int('\u0b9d'), 000, int('\u042f'), 001, int('\u33d1'), 000, int('\u0239'), 001, int('\u04c7'), 000, 0105, 001,
	013, 000, 001, 001, 056, 003, 020, 000, 004, 003, 015, 001, int('\u4060'), 000, 001, 005, 001, 013, int('\u2163'), 000,
	005, 003, 003, 000, 026, 003, 002, 000, 007, 003, 036, 000, 004, 003, 0224, 000, 003, 003, int('\u01bb'), 000,
	0125, 001, 001, 000, 0107, 001, 001, 000, 002, 001, 002, 000, 001, 001, 002, 000, 002, 001, 002, 000,
	004, 001, 001, 000, 014, 001, 001, 000, 001, 001, 001, 000, 007, 001, 001, 000, 0101, 001, 001, 000,
	004, 001, 002, 000, 010, 001, 001, 000, 007, 001, 001, 000, 034, 001, 001, 000, 004, 001, 001, 000,
	005, 001, 001, 000, 001, 001, 003, 000, 007, 001, 001, 000, int('\u0154'), 001, 002, 000, 031, 001, 001, 000,
	031, 001, 001, 000, 037, 001, 001, 000, 031, 001, 001, 000, 037, 001, 001, 000, 031, 001, 001, 000,
	037, 001, 001, 000, 031, 001, 001, 000, 037, 001, 001, 000, 031, 001, 001, 000, 010, 001, 002, 000,
	062, 004, int('\u1600'), 000, 004, 001, 001, 000, 033, 001, 001, 000, 002, 001, 001, 000, 001, 001, 002, 000,
	001, 001, 001, 000, 012, 001, 001, 000, 004, 001, 001, 000, 001, 001, 001, 000, 001, 001, 006, 000,
	001, 001, 004, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000, 003, 001, 001, 000,
	002, 001, 001, 000, 001, 001, 002, 000, 001, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000,
	001, 001, 001, 000, 001, 001, 001, 000, 002, 001, 001, 000, 001, 001, 002, 000, 004, 001, 001, 000,
	007, 001, 001, 000, 004, 001, 001, 000, 004, 001, 001, 000, 001, 001, 001, 000, 012, 001, 001, 000,
	021, 001, 005, 000, 003, 001, 001, 000, 005, 001, 001, 000, 021, 001, int('\u032a'), 000, 032, 017, 001, 013,
	int('\u0dff'), 000, int('\ua6d7'), 012, 051, 000, int('\u1035'), 012, 013, 000, 0336, 012, int('\u3fe2'), 000, int('\u021e'), 012, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\uffff'), 000, int('\u05ee'), 000,
	001, 003, 036, 000, 0140, 003, 0200, 000, 0360, 003, int('\uffff'), 000, int('\uffff'), 000, int('\ufe12'), 0,
}

/* Translates characters to character classes */
var ZZ_CMAP = zzUnpackCMap(ZZ_CMAP_PACKED)

/* Unpacks the comressed character translation table */
func zzUnpackCMap(packed []int) []rune {
	m := make([]rune, 0x110000)
	j := 0 // index in unpacked array
	var count int
	assert(len(packed) == 2836)
	for i, value := range packed {
		if i%2 == 0 {
			count = value
		} else {
			m[j] = rune(value)
			j++
			count--
			for count > 0 {
				m[j] = rune(value)
				j++
				count--
			}
		}
	}
	return m
}

/* Translates DFA states to action switch labels. */
var ZZ_ACTION = zzUnpackAction([]int{
	001, 000, 001, 001, 001, 002, 001, 003, 001, 004, 001, 005, 001, 001, 001, 006,
	001, 007, 001, 002, 001, 001, 001, 010, 001, 002, 001, 000, 001, 002, 001, 000,
	001, 004, 001, 000, 002, 002, 002, 000, 001, 001, 001, 0,
})

func zzUnpackAction(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
	}
	return m
}

/* Translates a state to a row index in the transition table */
var ZZ_ROWMAP = zzUnpackRowMap([]int{
	000, 000, 000, 022, 000, 044, 000, 066, 000, 0110, 000, 0132, 000, 0154, 000, 176,
	000, 0220, 000, 0242, 000, 0264, 000, 0306, 000, 0330, 000, 0352, 000, 0374, 000, int('\u010e'),
	000, int('\u0120'), 000, 0154, 000, int('\u0132'), 000, int('\u0144'), 000, int('\u0156'), 000, 0264, 000, int('\u0168'), 000, int('\u017a'),
})

func zzUnpackRowMap(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		high, low := packed[i]<<16, packed[i+1]
		m[j] = high | low
		j++
	}
	return m
}

/* The transition table of the DFA */
var ZZ_TRANS = zzUnpackTrans([]int{
	001, 002, 001, 003, 001, 004, 001, 002, 001, 005, 001, 006, 003, 002, 001, 007,
	001, 010, 001, 011, 002, 002, 001, 012, 001, 013, 002, 014, 023, 000, 003, 003,
	001, 015, 001, 000, 001, 016, 001, 000, 001, 016, 001, 017, 002, 000, 001, 016,
	001, 000, 001, 012, 002, 000, 001, 003, 001, 000, 001, 003, 002, 004, 001, 015,
	001, 000, 001, 016, 001, 000, 001, 016, 001, 017, 002, 000, 001, 016, 001, 000,
	001, 012, 002, 000, 001, 004, 001, 000, 002, 003, 002, 005, 002, 000, 002, 020,
	001, 021, 002, 000, 001, 020, 001, 000, 001, 012, 002, 000, 001, 005, 003, 000,
	001, 006, 001, 000, 001, 006, 003, 000, 001, 017, 007, 000, 001, 006, 001, 000,
	002, 003, 001, 022, 001, 005, 001, 023, 003, 000, 001, 022, 004, 000, 001, 012,
	002, 000, 001, 022, 003, 000, 001, 010, 015, 000, 001, 010, 003, 000, 001, 011,
	015, 000, 001, 011, 001, 000, 002, 003, 001, 012, 001, 015, 001, 000, 001, 016,
	001, 000, 001, 016, 001, 017, 002, 000, 001, 024, 001, 025, 001, 012, 002, 000,
	001, 012, 003, 000, 001, 026, 013, 000, 001, 027, 001, 000, 001, 026, 003, 000,
	001, 014, 014, 000, 002, 014, 001, 000, 002, 003, 002, 015, 002, 000, 002, 030,
	001, 017, 002, 000, 001, 030, 001, 000, 001, 012, 002, 000, 001, 015, 001, 000,
	002, 003, 001, 016, 012, 000, 001, 003, 002, 000, 001, 016, 001, 000, 002, 003,
	001, 017, 001, 015, 001, 023, 003, 000, 001, 017, 004, 000, 001, 012, 002, 000,
	001, 017, 003, 000, 001, 020, 001, 005, 014, 000, 001, 020, 001, 000, 002, 003,
	001, 021, 001, 005, 001, 023, 003, 000, 001, 021, 004, 000, 001, 012, 002, 000,
	001, 021, 003, 000, 001, 023, 001, 000, 001, 023, 003, 000, 001, 017, 007, 000,
	001, 023, 001, 000, 002, 003, 001, 024, 001, 015, 004, 000, 001, 017, 004, 000,
	001, 012, 002, 000, 001, 024, 003, 000, 001, 025, 012, 000, 001, 024, 002, 000,
	001, 025, 003, 000, 001, 027, 013, 000, 001, 027, 001, 000, 001, 027, 003, 000,
	001, 030, 001, 015, 014, 000, 001, 030,
})

func zzUnpackTrans(packed []int) []int {
	m := make([]int, 396)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]-1
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
	}
	return m
}

/* error codes */
const (
	ZZ_UNKNOWN_ERROR = 0
	ZZ_NO_MATCH      = 1
)

/* error messages for the codes above */
var ZZ_ERROR_MSG = [3]string{
	"Unkown internal scanner error",
	"Error: could not match input",
	"Error: pushback value was too large",
}

/* ZZ_ATTRIBUTE[aState] contains the attributes of state aState */
var ZZ_ATTRIBUTE = zzUnpackAttribute([]int{
	001, 000, 001, 011, 013, 001, 001, 000, 001, 001, 001, 000, 001, 001, 001, 000,
	002, 001, 002, 000, 001, 001, 001, 0,
})

func zzUnpackAttribute(packed []int) []int {
	m := make([]int, 24)
	j := 0
	for i, l := 0, len(packed); i < l; i += 2 {
		count, value := packed[i], packed[i+1]
		m[j] = value
		j++
		count--
		for count > 0 {
			m[j] = value
			j++
			count--
		}
		i += 2
	}
	return m
}

const (
	WORD_TYPE             = ALPHANUM
	NUMERIC_TYPE          = NUM
	SOUTH_EAST_ASIAN_TYPE = SOUTHEAST_ASIAN
	IDEOGRAPHIC_TYPE      = IDEOGRAPHIC
	HIRAGANA_TYPE         = HIRAGANA
	KATAKANA_TYPE         = KATAKANA
	HANGUL_TYPE           = HANGUL
)

/*
This class implements Word Break rules from the Unicode Text
Segmentation algorithm, as specified in Unicode Standard Annex #29.

Tokens produced are of the following types:

	- <ALPHANUM>: A sequence of alphabetic and numeric characters
	- <NUM>: A number
	- <SOUTHEAST_ASIAN>: A sequence of characters from South and Southeast Asian languages, including Thai, Lao, Myanmar, and Khmer
	- IDEOGRAPHIC>: A single CJKV ideographic character
	- <HIRAGANA>: A single hiragana character

Technically it should auto generated by JFlex but there is no GoFlex
yet. So it's a line-by-line port.
*/
type StandardTokenizerImpl struct {
	// the input device
	zzReader io.RuneReader

	// the current state of the DFA
	zzState int

	// the current lexical state
	zzLexicalState int

	// this buffer contains the current text to be matched and is the
	// source of yytext() string
	zzBuffer []rune

	// the text position at the last accepting state
	zzMarkedPos int

	// the current text position in the buffer
	zzCurrentPos int

	// startRead marks the beginning of the yytext() string in the buffer
	zzStartRead int

	// endRead marks the last character in the buffer, that has been read from input
	zzEndRead int

	// number of newlines encountered up to the start of the matched text
	yyline int

	// the number of characters up to the start of the matched text
	_yychar int

	// the number of characters from the last newline up to the start of the matched text
	yycolumn int

	// zzAtBOL == true <=> the scanner is currently at the beginning of a line
	zzAtBOL bool

	// zzAtEOF == true <=> the scanner is at the EOF
	zzAtEOF bool

	// denotes if the user-EOF-code has already been executed
	zzEOFDone bool

	// The number of occupied positions in zzBuffer beyond zzEndRead.
	// When a lead/high surrogate has been read from the input stream
	// into the final zzBuffer position, this will have a value of 1;
	// otherwise, it will have a value of 0.
	zzFinalHighSurrogate int
}

func newStandardTokenizerImpl(in io.RuneReader) *StandardTokenizerImpl {
	return &StandardTokenizerImpl{
		zzReader:       in,
		zzLexicalState: YYINITIAL,
		zzBuffer:       make([]rune, ZZ_BUFFERSIZE),
		zzAtBOL:        true,
	}
}

func (t *StandardTokenizerImpl) yychar() int {
	return t._yychar
}

/* Fills CharTermAttribute with the current token text. */
func (t *StandardTokenizerImpl) text(tt CharTermAttribute) {
	tt.CopyBuffer(t.zzBuffer[t.zzStartRead:t.zzMarkedPos])
}

/* Refills the input buffer. */
func (t *StandardTokenizerImpl) zzRefill() (bool, error) {
	// first: make room (if you can)
	if t.zzStartRead > 0 {
		t.zzEndRead += t.zzFinalHighSurrogate
		t.zzFinalHighSurrogate = 0
		copy(t.zzBuffer, t.zzBuffer[t.zzStartRead:t.zzEndRead])

		// translate stored positions
		t.zzEndRead -= t.zzStartRead
		t.zzCurrentPos -= t.zzStartRead
		t.zzMarkedPos -= t.zzStartRead
		t.zzStartRead = 0
	}

	// fill the buffer with new input
	var requested = len(t.zzBuffer) - t.zzEndRead - t.zzFinalHighSurrogate
	var totalRead = 0
	var err error
	for totalRead < requested && err != io.EOF {
		var numRead int
		if numRead, err = readRunes(t.zzReader.(io.RuneReader),
			t.zzBuffer[t.zzEndRead+totalRead:]); err != nil && err != io.EOF {
			return false, err
		}
		totalRead += numRead
	}

	if totalRead > 0 {
		t.zzEndRead += totalRead
		if totalRead == requested { // possibly more input available
			panic("niy")
		}
		return false, nil
	}

	assert(totalRead == 0 && err == io.EOF)
	return true, nil
}

func readRunes(r io.RuneReader, buffer []rune) (int, error) {
	for i, _ := range buffer {
		ch, _, err := r.ReadRune()
		if err != nil {
			return i, err
		}
		buffer[i] = ch
	}
	return len(buffer), nil
}

/*
Resets the scanner to read from a new input stream.
Does not close the old reader.

All internal variables are reset, the old input stream
cannot be reused (internal buffer is discarded and lost).
Lexical state is set to ZZ_INITIAL.

Internal scan buffer is resized down to its initial length, if it has grown.
*/
func (t *StandardTokenizerImpl) yyreset(reader io.RuneReader) {
	t.zzReader = reader
	t.zzAtBOL = true
	t.zzAtEOF = false
	t.zzEOFDone = false
	t.zzEndRead, t.zzStartRead = 0, 0
	t.zzCurrentPos, t.zzMarkedPos = 0, 0
	t.zzFinalHighSurrogate = 0
	t.yyline, t._yychar, t.yycolumn = 0, 0, 0
	t.zzLexicalState = YYINITIAL
	if len(t.zzBuffer) > ZZ_BUFFERSIZE {
		t.zzBuffer = make([]rune, ZZ_BUFFERSIZE)
	}
}

/* Returns the length of the matched text region. */
func (t *StandardTokenizerImpl) yylength() int {
	return t.zzMarkedPos - t.zzStartRead
}

/*
Reports an error that occurred while scanning.

In a wellcormed scanner (no or only correct usage of yypushack(int)
and a match-all fallback rule) this method will only be called with
things that "can't possibly happen". If thismethod is called,
something is seriously wrong (e.g. a JFlex bug producing a faulty
scanner etc.).

Usual syntax/scanner level error handling should be done in error
fallback rules.
*/
func (t *StandardTokenizerImpl) zzScanError(errorCode int) {
	var msg string
	if errorCode >= 0 && errorCode < len(ZZ_ERROR_MSG) {
		msg = ZZ_ERROR_MSG[errorCode]
	} else {
		msg = ZZ_ERROR_MSG[ZZ_UNKNOWN_ERROR]
	}
	panic(msg)
}

/*
Resumes scanning until the next regular expression is matched, the
end of input is encountered or an I/O-Error occurs.
*/
func (t *StandardTokenizerImpl) nextToken() (int, error) {
	var zzInput, zzAction int

	// cached fields:
	var zzCurrentPosL, zzMarkedPosL int
	zzEndReadL := t.zzEndRead
	zzBufferL := t.zzBuffer
	zzCMapL := ZZ_CMAP

	zzTransL := ZZ_TRANS
	zzRowMapL := ZZ_ROWMAP
	zzAttrL := ZZ_ATTRIBUTE

	for {
		zzMarkedPosL = t.zzMarkedPos

		t._yychar += zzMarkedPosL - t.zzStartRead

		zzAction = -1

		zzCurrentPosL = zzMarkedPosL
		t.zzCurrentPos = zzMarkedPosL
		t.zzStartRead = zzMarkedPosL

		t.zzState = ZZ_LEXSTATE[t.zzLexicalState]

		// set up zzAction for empty match case:
		if zzAttributes := zzAttrL[t.zzState]; (zzAttributes & 1) == 1 {
			zzAction = t.zzState
		}

		for {
			if zzCurrentPosL < zzEndReadL {
				zzInput = int(zzBufferL[zzCurrentPosL])
				zzCurrentPosL++
			} else if t.zzAtEOF {
				zzInput = YYEOF
				break
			} else {
				// store back cached positions
				t.zzCurrentPos = zzCurrentPosL
				t.zzMarkedPos = zzMarkedPosL
				eof, err := t.zzRefill()
				if err != nil {
					return 0, err
				}
				// get translated positions and possibly new buffer
				zzCurrentPosL = t.zzCurrentPos
				zzMarkedPosL = t.zzMarkedPos
				zzBufferL = t.zzBuffer
				zzEndReadL = t.zzEndRead
				if eof {
					zzInput = YYEOF
					break
				} else {
					zzInput = int(zzBufferL[zzCurrentPosL])
					zzCurrentPosL++
				}
			}
			zzNext := zzTransL[zzRowMapL[t.zzState]+int(zzCMapL[zzInput])]
			if zzNext == -1 {
				break
			}
			t.zzState = zzNext

			if zzAttributes := zzAttrL[t.zzState]; (zzAttributes & 1) == 1 {
				zzAction = t.zzState
				zzMarkedPosL = zzCurrentPosL
				if (zzAttributes & 8) == 8 {
					break
				}
			}
		}

		// store back cached position
		t.zzMarkedPos = zzMarkedPosL

		cond := zzAction
		if zzAction >= 0 {
			cond = ZZ_ACTION[zzAction]
		}
		switch cond {
		case 1, 9, 10, 11, 12, 13, 14, 15, 16:
			// break so we don't hit fall-through warning:
			// not numeric, word, ideographic, hiragana, or SE Asian -- ignore it.
		case 2:
			return WORD_TYPE, nil
		case 3:
			return HANGUL_TYPE, nil
		case 4:
			return NUMERIC_TYPE, nil
		case 5:
			return KATAKANA_TYPE, nil
		case 6:
			return IDEOGRAPHIC_TYPE, nil
		case 7:
			return HIRAGANA_TYPE, nil
		case 8:
			return SOUTH_EAST_ASIAN_TYPE, nil
		default:
			if zzInput == YYEOF && t.zzStartRead == t.zzCurrentPos {
				t.zzAtEOF = true
				return YYEOF, nil
			} else {
				t.zzScanError(ZZ_NO_MATCH)
				panic("should not be here")
			}
		}
	}
}
