package lexer

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenTag       // Example: div, h1, box
	TokenSeparator // Char ":"
	TokenDash
	TokenString
	TokenNewline // New Line \n
	TokenIndent  // Hierarchy spaces
)

type Token struct {
	Type    TokenType
	Literal string
}
