package internal

import (
	"fmt"
	"strings"
)

type ErrorCode int

const (
	ErrUnknown ErrorCode = iota
	ErrSyntax
	ErrUndefinedSymbol
	ErrType
	ErrInvalidNumber
)

type Position struct {
	Line int
	Col  int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Range struct {
	Start Position
	End   Position
}

type Diagnostic struct {
	Code    ErrorCode
	Message string
	Range   Range
}

func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s", d.Range.Start, d.Message)
}

type DiagnosticList []Diagnostic

func (l DiagnosticList) Error() string {
	if len(l) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, d := range l {
		sb.WriteString(d.Error())
		sb.WriteRune('\n')
	}
	return sb.String()
}

func (l DiagnosticList) HasErrors() bool {
	return len(l) > 0
}
