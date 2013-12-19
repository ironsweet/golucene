package automaton

// util/automaton/RegExp.java

/*
Regular Expression extension to Automaton.

Regular expressions are built from the following abstract syntax:

	regexp	::= unionexp
 			|
	unionexp ::= interexp | unionexp 	(union)
	 		| interexp
	interexp ::= concatexp & interexp 	(intersection) 						[OPTIONAL]
 			| concatexp
	concatexp ::= repeatexp concatexp	 (concatenation)
 			| repeatexp
	repeatexp ::= repeatexp ? 			(zero or one occurrence)
 			| repeatexp * 				(zero or more occurrences)
 			| repeatexp + 				(one or more occurrences)
 			| repeatexp {n} 			(n occurrences)
 			| repeatexp {n,} 			(n or more occurrences)
 			| repeatexp {n,m} 			(n to m occurrences, including both)
 			| complexp
	complexp ::= ~ complexp 			(complement) 						[OPTIONAL]
 			| charclassexp
	charclassexp ::= [ charclasses ] 	(character class)
 			| [^ charclasses ] 			(negated character class)
 			| simpleexp
	charclasses ::= charclass charclasses
 			| charclass
	charclass ::= charexp - charexp 	(character range, including end-points)
 			| charexp
	simpleexp ::= charexp
 			| . 						(any single character)
 			| # 						(the empty language) 				[OPTIONAL]
 			| @ 						(any string) 						[OPTIONAL]
			| " <Unicode string without double-quotes>  " (a string)
 			| ( ) 						(the empty string)
 			| ( unionexp ) 				(precedence override)
 			| < <identifier> > 			(named automaton) 					[OPTIONAL]
			| <n-m> 					(numerical interval) 				[OPTIONAL]
	charexp ::= <Unicode character> 	(a single non-reserved character)
 			| \ <Unicode character>  	(a single character)

The productions marked [OPTIONAL] are only allowed if specified by
the syntax flags passed to the RegExp constructor. The reserved
characters used in the (enabled) syntax must be escaped with
backslash (\) or double-quotes ("..."). (In contrast to other regexp
syntaxes, this is required also in character classes.) Be aware that
dash (-) has a special meaning in charclass expressions. An
identifier is a string not containing right angle bracket (>) or dash
(-). Numerical intervals are specified by non-negative decimal
integers and include both end points, and if n and m have the same
number of digits, then the conforming strings must have that length
(i.e. prefixed by 0's).
*/
type RegExp struct {
}

// Constructs new Regexp from a string. Same as RegExp(s, ALL)
func NewRegExp(s string) *RegExp {
	panic("not implemented yet")
}

// Constructs new Automaton from this RegExp. Same as
// ToAutomaton(nil) (empty automaton map).
func (re *RegExp) ToAutomaton() *Automaton {
	panic("not implemented yet")
}
