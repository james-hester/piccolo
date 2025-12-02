//go:generate stringer -type=TTy
package internal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Token type
// This is a signed byte (int8) for the silly reason that
// stringer generates "if i < 0 || idx >= len(...)"
// which shows a warning if TTy is uint8
type TTy int8

const (
	UNKNOWN TTy = iota
	EOF

	// Operators
	EQL    // =
	NEQ    // !=
	INC    // ++
	DEC    // --
	ANDEQL // &=
	OREQL  // |=
	XOREQL // ^=
	ADDEQL // +=
	SUBEQL // -=
	MINUS  // -

	// Punctuation
	LBRACK // [
	RBRACK // ]
	LPAREN // (
	RPAREN // )
	COLON  // :

	// Keywords
	FN
	BEGIN
	END
	RETURN
	IF
	THEN
	NOT
	SECTION
	CONSTANTS
	DATA
	PROGRAM
	CONFIGURATION
	BANKED
	COMMON
	I8
	AT

	// Names and literals
	IDENT

	NUM_First
	NUMDECIMAL
	NUMHEX
	NUMBINARY
	NUM_Last
)

type Tok struct {
	ty    TTy
	val   string
	Range Range
}

func (tok *Tok) Number() (int, error) {
	// Remove underscores
	clean := strings.ReplaceAll(tok.val, "_", "")

	var base int
	switch tok.ty {
	case NUMHEX:
		base = 16
	case NUMBINARY:
		base = 2
	case NUMDECIMAL:
		base = 10
	default:
		return 0, fmt.Errorf("invalid number type: %v", tok.ty)
	}

	val, err := strconv.ParseInt(clean, base, 32)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

var keywords = map[string]TTy{
	"fn":            FN,
	"begin":         BEGIN,
	"end":           END,
	"return":        RETURN,
	"if":            IF,
	"then":          THEN,
	"not":           NOT,
	"section":       SECTION,
	"constants":     CONSTANTS,
	"data":          DATA,
	"program":       PROGRAM,
	"configuration": CONFIGURATION,
	"banked":        BANKED,
	"common":        COMMON,
	"i8":            I8,
	"at":            AT,
}

func Lex(text string) ([]Tok, error) {
	result := []Tok{}
	var diagnostics DiagnosticList
	l := newLexer(text)

	for {
		l.skipWhitespace()
		l.startTok()
		if l.peek() == 0 {
			break
		}

		// Comments start with // and run to the end of the line.
		if l.peek() == '/' && l.peekNext() == '/' {
			l.skipLineComment()
			continue
		}

		if op, ok := l.tryTwoCharOp(); ok {
			result = append(result, l.finishTok(op))
			continue
		}

		switch l.peek() {
		case '$':
			result = append(result, l.finishTokVal(NUMHEX, l.scanHex()))
			continue
		case '%':
			result = append(result, l.finishTokVal(NUMBINARY, l.scanBinary()))
			continue
		case '=':
			l.advance()
			result = append(result, l.finishTok(EQL))
			continue
		case '[':
			l.advance()
			result = append(result, l.finishTok(LBRACK))
			continue
		case ']':
			l.advance()
			result = append(result, l.finishTok(RBRACK))
			continue
		case '(':
			l.advance()
			result = append(result, l.finishTok(LPAREN))
			continue
		case ')':
			l.advance()
			result = append(result, l.finishTok(RPAREN))
			continue
		case ':':
			l.advance()
			result = append(result, l.finishTok(COLON))
			continue
		case '-':
			l.advance()
			result = append(result, l.finishTok(MINUS))
			continue
		}

		if unicode.IsDigit(l.peek()) {
			result = append(result, l.finishTokVal(NUMDECIMAL, l.scanDecimal()))
			continue
		}

		if isIdentStart(l.peek()) {
			ident := l.scanIdent()
			if kw, ok := keywords[strings.ToLower(ident)]; ok {
				result = append(result, l.finishTok(kw))
			} else {
				result = append(result, l.finishTokVal(IDENT, ident))
			}
			continue
		}

		// Error case
		diagnostics = append(diagnostics, Diagnostic{
			Code:    ErrSyntax,
			Message: fmt.Sprintf("unexpected character %q", l.peek()),
			Range:   l.currentRange(),
		})
		l.advance() // Skip the bad char
	}

	l.startTok()
	result = append(result, l.finishTok(EOF))

	if len(diagnostics) > 0 {
		return result, diagnostics
	}
	return result, nil
}

// lexer is a minimal rune-based scanner for Piccolo source.
type lexer struct {
	src   []rune
	pos   int
	line  int
	col   int
	start Position
}

func newLexer(input string) *lexer {
	return &lexer{
		src:  []rune(input),
		line: 1,
		col:  1,
	}
}

func (l *lexer) startTok() {
	l.start = l.position()
}

func (l *lexer) finishTok(ty TTy) Tok {
	return Tok{
		ty:    ty,
		Range: l.currentRange(),
	}
}

func (l *lexer) finishTokVal(ty TTy, val string) Tok {
	return Tok{
		ty:    ty,
		val:   val,
		Range: l.currentRange(),
	}
}

func (l *lexer) currentRange() Range {
	return Range{Start: l.start, End: l.position()}
}

func (l *lexer) position() Position {
	return Position{Line: l.line, Col: l.col}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) peekNext() rune {
	if l.pos+1 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+1]
}

func (l *lexer) advance() rune {
	r := l.peek()
	if r == 0 {
		return 0
	}
	l.pos++
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) skipWhitespace() {
	for {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			l.advance()
			continue
		}
		return
	}
}

func (l *lexer) skipLineComment() {
	for r := l.peek(); r != 0 && r != '\n'; r = l.peek() {
		l.advance()
	}
}

func (l *lexer) scanIdent() string {
	var sb strings.Builder
	for isIdentPart(l.peek()) {
		sb.WriteRune(l.advance())
	}
	s := sb.String()
	// Backtrack trailing hyphens
	trim := strings.TrimRight(s, "-")
	diff := len(s) - len(trim)
	l.pos -= diff
	l.col -= diff
	return trim
}

func (l *lexer) tryTwoCharOp() (TTy, bool) {
	switch {
	case l.peek() == '!' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return NEQ, true
	case l.peek() == '+' && l.peekNext() == '+':
		l.advance()
		l.advance()
		return INC, true
	case l.peek() == '-' && l.peekNext() == '-':
		l.advance()
		l.advance()
		return DEC, true
	case l.peek() == '&' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return ANDEQL, true
	case l.peek() == '|' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return OREQL, true
	case l.peek() == '^' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return XOREQL, true
	case l.peek() == '+' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return ADDEQL, true
	case l.peek() == '-' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return SUBEQL, true
	default:
		return UNKNOWN, false
	}
}

func (l *lexer) scanHex() string {
	var sb strings.Builder
	l.advance() // $
	for isHexDigit(l.peek()) || l.peek() == '_' {
		sb.WriteRune(l.advance())
	}
	return sb.String()
}

func (l *lexer) scanBinary() string {
	var sb strings.Builder
	l.advance() // %
	for l.peek() == '0' || l.peek() == '1' || l.peek() == '_' {
		sb.WriteRune(l.advance())
	}
	return sb.String()
}

func (l *lexer) scanDecimal() string {
	var sb strings.Builder
	for unicode.IsDigit(l.peek()) || l.peek() == '_' {
		sb.WriteRune(l.advance())
	}
	return sb.String()
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r)
}

func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
