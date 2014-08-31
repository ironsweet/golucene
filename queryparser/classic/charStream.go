package classic

type CharStream interface {
	readChar() (rune, error)
	endColumn() int
	endLine() int
	backup(int)
	beginToken() (rune, error)
	image() string
}
