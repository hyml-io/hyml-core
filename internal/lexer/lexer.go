package lexer

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	switch l.ch {
	case '"': // <--- If colon is found, read the entire string
		tok.Type = TokenString
		tok.Literal = l.readString()
	case '-':
		tok = Token{Type: TokenDash, Literal: "-"}
	case ':':
		tok = Token{Type: TokenSeparator, Literal: ":"}
	case 0:
		tok = Token{Type: TokenEOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			return Token{Type: TokenTag, Literal: literal}
		}
		tok = Token{Type: TokenError, Literal: string(l.ch)}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	// DO NOT include ':' here
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || '0' <= ch && ch <= '9' || ch == '_' || ch == '-'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
