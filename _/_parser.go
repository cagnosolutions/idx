package adb

import (
	"fmt"
	"strings"
)

type comp struct {
	fld, opt, val string
}

type statement struct {
	store  string
	comps []comp
}

type parser struct {
	l   *lexer
	buf struct {
		tok token
		lit string
		n   int
	}
}

func Q(s string) (string, []comp) {
	stmt, err := parse(s)
	if err != nil {
		panic(err)
	}
	return stmt.store, stmt.comps
}

func parse(qry string) (*statement, error) {
	p := &parser{
		l: newLexer(strings.NewReader(qry)),
	}
	stmt := &statement{}
	if tok, lit := p.scanIgnoreWhitespace(); tok != t_SELECT {
		return nil, fmt.Errorf("found %q, expected SELECT", lit)
	}
	if tok, lit := p.scanIgnoreWhitespace(); tok != t_FROM {
		return nil, fmt.Errorf("found %q, expected FROM", lit)
	}
	tok, lit := p.scanIgnoreWhitespace()
	if tok != t_IDENT {
		return nil, fmt.Errorf("found %q, expected store name", lit)
	}
	stmt.store = lit
	if tok, _ := p.scanIgnoreWhitespace(); tok != t_WHERE {
		stmt.comps = append(stmt.comps, comp{fld:"*"})
		return stmt, nil
	}
	for {
		tok, lit := p.scanIgnoreWhitespace()
		if tok != t_IDENT && tok < t_OPERATOR {
			return nil, fmt.Errorf("found %q, expected field set", lit)
		}
		fld, opt, val := toComp(lit)
		stmt.comps = append(stmt.comps, comp{fld, opt, val})
		if tok, _ := p.scanIgnoreWhitespace(); tok != t_COMMA {
			p.unscan()
			break
		}
	}
	return stmt, nil
}

func (p *parser) scan() (token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit := p.l.scan()
	p.buf.tok, p.buf.lit = tok, lit
	return tok, lit
}

func (p *parser) scanIgnoreWhitespace() (token, string) {
	tok, lit := p.scan()
	if tok == t_WS {
		tok, lit = p.scan()
	}
	return tok, lit
}

func (p *parser) unscan() {
	p.buf.n = 1
}
