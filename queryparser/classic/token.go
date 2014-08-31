package classic

type Token struct {
	kind        int
	beginLine   int
	beginColumn int
	endLine     int
	endColumn   int
	image       string
	next        *Token
}

func newToken(ofKind int, image string) *Token {
	return &Token{kind: ofKind, image: image}
}

func (t *Token) String() string {
	return t.image
}
