package classic

type CharStream interface {
	readChar() (rune, error)
	endColumn() int
	endLine() int
	beginColumn() int
	beginLine() int
	backup(int)
	beginToken() (rune, error)
	image() string
}
