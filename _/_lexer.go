package adb

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type token int

const (
	t_ILLEGAL token = iota
	t_EOF
	t_WS

	t_IDENT

	//t_ASTERISK
	t_COMMA

	t_SELECT
	t_FROM
	t_WHERE

	t_OPERATOR
	t_LT
	t_GT
	t_EQ
	t_NE
)

var opts = []string{"<",">","=","!"}

func toComp(lit string) (string,string,string) {
	for _, op := range opts {
		if n := strings.Index(lit, op); n != -1 {
			return lit[:n], lit[n:n+len(op)], lit[n+len(op):]
		}
	}
	return "","",""
}

type lexer struct {
	r *bufio.Reader
}

func newLexer(r io.Reader) *lexer {
	return &lexer{
		r: bufio.NewReader(r),
	}
}

func (l *lexer) scan() (token, string) {
	ch := l.read()
	if isWhitespace(ch) {
		l.unread()
		return l.scanWhitespace()
	} else if isLetter(ch) {
		l.unread()
		return l.scanIdent()
	}
	switch ch {
	case eof:
		return t_EOF, ""
	//case '*':
	//	return t_ASTERISK, string(ch)
	case ',':
		return t_COMMA, string(ch)
	}
	return t_ILLEGAL, string(ch)
}

func (l *lexer) scanWhitespace() (token, string) {
	var buf bytes.Buffer
	buf.WriteRune(l.read())
	for {
		if ch := l.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			l.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return t_WS, buf.String()
}

func (l *lexer) scanIdent() (token, string) {
	var buf bytes.Buffer
	buf.WriteRune(l.read())
	for {
		if ch := l.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && !isOperator(ch) && ch != '_' {
			l.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	switch strings.ToUpper(buf.String()) {
	case "SELECT":
		return t_SELECT, buf.String()
	case "FROM":
		return t_FROM, buf.String()
	case "WHERE":
		return t_WHERE, buf.String()
	}
	return t_IDENT, buf.String()
}

func (l *lexer) read() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (l *lexer) unread() {
	_ = l.r.UnreadRune()
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func isOperator(ch rune) bool {
	return ch == '=' || ch == '!' || ch == '>' || ch == '<'
}

var eof = rune(0)
