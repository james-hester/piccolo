package internal

import (
	"fmt"
	"strings"
)

func (tk Tok) String() string {
	if tk.val == "" {
		return tk.ty.String()
	}
	return fmt.Sprintf("%s[%s]", tk.ty.String(), tk.val)
}

func (p *Program) String() string {
	var sb strings.Builder
	sb.WriteString("Program {\n")
	for _, fn := range p.Functions {
		sb.WriteString(fn.String())
		sb.WriteRune('\n')
	}
	sb.WriteRune('}')
	return sb.String()
}

func (f *Function) String() string {
	var sb strings.Builder
	sb.WriteString("Function[")
	sb.WriteString(f.Name)
	sb.WriteString("] {\n")
	for _, stmt := range f.Body {
		sb.WriteString(stmt.String())
		sb.WriteRune('\n')
	}
	sb.WriteRune('}')
	return sb.String()
}

func (a AssignStmt) String() string {
	return "assign"
}

func (r ReturnStmt) String() string {
	return "return"
}

func (i IdentExpr) String() string {
	return i.Name
}

func (n NumExpr) String() string {
	return n.Val
}

func (s IfStmt) String() string {
	return fmt.Sprintf("if %s then %s", s.Cond.String(), s.Then.String())
}

func (e IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]", e.Name, e.Index.String())
}

func (e UnaryExpr) String() string {
	return fmt.Sprintf("%s %s", e.Op.String(), e.Expr.String())
}

func (e BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.Lhs.String(), e.Op.String(), e.Rhs.String())
}

func (e PostfixExpr) String() string {
	return fmt.Sprintf("%s%s", e.Expr.String(), e.Op.String())
}

func (l LabelStmt) String() string {
	return l.Name + ":"
}

type Stmt interface {
	String() string
	isStmt()
	Position() Range
}

type Expr interface {
	String() string
	isExpr()
	Position() Range
}

type Program struct {
	Functions     []Function
	AtBlocks      []AtBlock
	Consts        map[string]int
	Configuration map[string]int
	SFRs          map[string]SFR
	Variables     map[string]Variable
}

type Variable struct {
	Name    string
	Type    string // "i8" for now
	Banked  bool   // true if banked, false if common
	Address int    // Assigned address
	Range   Range
}

type AtBlock struct {
	Address int
	Body    []Stmt
	Range   Range
}

type SFR struct {
	Address int
	Bits    map[string]int
	Range   Range
}

type Function struct {
	Name  string
	Body  []Stmt
	Range Range
}

type CallStmt struct {
	Name  string
	Range Range
}

func (CallStmt) isStmt()           {}
func (s CallStmt) Position() Range { return s.Range }

func (c CallStmt) String() string {
	return fmt.Sprintf("call %s", c.Name)
}

type AssignStmt struct {
	Lhs   Expr
	Op    TTy
	Expr  Expr
	Range Range
}

func (AssignStmt) isStmt()           {}
func (s AssignStmt) Position() Range { return s.Range }

type IfStmt struct {
	Cond  Expr
	Then  Stmt
	Range Range
}

func (IfStmt) isStmt()           {}
func (s IfStmt) Position() Range { return s.Range }

type ReturnStmt struct {
	Range Range
}

func (ReturnStmt) isStmt()           {}
func (s ReturnStmt) Position() Range { return s.Range }

type LabelStmt struct {
	Name  string
	Range Range
}

func (LabelStmt) isStmt()           {}
func (s LabelStmt) Position() Range { return s.Range }

type IdentExpr struct {
	Name  string
	Range Range
}

func (IdentExpr) isExpr()           {}
func (e IdentExpr) Position() Range { return e.Range }

type IndexExpr struct {
	Name  string
	Index Expr
	Range Range
}

func (IndexExpr) isExpr()           {}
func (e IndexExpr) Position() Range { return e.Range }

type NumExpr struct {
	Val   string
	Value int
	Ty    TTy
	Range Range
}

func (NumExpr) isExpr()           {}
func (e NumExpr) Position() Range { return e.Range }

type UnaryExpr struct {
	Op    TTy
	Expr  Expr
	Range Range
}

func (UnaryExpr) isExpr()           {}
func (e UnaryExpr) Position() Range { return e.Range }

type BinaryExpr struct {
	Lhs   Expr
	Op    TTy
	Rhs   Expr
	Range Range
}

func (BinaryExpr) isExpr()           {}
func (e BinaryExpr) Position() Range { return e.Range }

type PostfixExpr struct {
	Expr  Expr
	Op    TTy
	Range Range
}

func (PostfixExpr) isExpr()           {}
func (e PostfixExpr) Position() Range { return e.Range }

func Parse(tokens []Tok) (Program, error) {
	p := newParser(tokens)
	prog := p.parseProgram()
	if len(p.diagnostics) > 0 {
		return prog, p.diagnostics
	}
	return prog, nil
}

type parser struct {
	toks        []Tok
	pos         int
	diagnostics DiagnosticList
}

func newParser(toks []Tok) *parser {
	return &parser{toks: toks}
}

func (p *parser) error(msg string) {
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Code:    ErrSyntax,
		Message: msg,
		Range:   p.current().Range,
	})
}

func (p *parser) current() Tok {
	if p.pos >= len(p.toks) {
		return Tok{ty: EOF}
	}
	return p.toks[p.pos]
}

func (p *parser) peekNext() Tok {
	if p.pos+1 >= len(p.toks) {
		return Tok{ty: EOF}
	}
	return p.toks[p.pos+1]
}

func (p *parser) advance() {
	p.pos++
}

func (p *parser) expect(ty TTy, msg string) (Tok, bool) {
	tok := p.current()
	if tok.ty != ty {
		p.error(msg)
		return Tok{}, false
	}
	p.pos++
	return tok, true
}

func (p *parser) synchronize() {
	for p.current().ty != EOF && p.current().ty != SECTION {
		p.advance()
	}
}

func (p *parser) parseProgram() Program {
	result := Program{
		Functions:     []Function{},
		AtBlocks:      []AtBlock{},
		Consts:        make(map[string]int),
		Configuration: make(map[string]int),
		SFRs:          make(map[string]SFR),
		Variables:     make(map[string]Variable),
	}

	for p.current().ty != EOF {
		if p.current().ty == SECTION {
			p.advance()
			switch p.current().ty {
			case CONSTANTS:
				p.advance()
				p.parseConstants(&result)
			case CONFIGURATION:
				p.advance()
				p.parseConfiguration(&result)
			case DATA:
				p.advance()
				p.parseData(&result)
			case PROGRAM:
				p.advance()
				p.parseFunctions(&result)
			default:
				p.error("unknown section type")
				p.synchronize()
			}
		} else {
			p.error("unexpected token at top level")
			p.synchronize()
		}
	}

	return result
}

func (p *parser) parseConfiguration(prog *Program) {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == IDENT {
			name := p.current().val
			p.advance()
			if _, ok := p.expect(COLON, fmt.Sprintf("expected : after identifier %s", name)); !ok {
				continue
			}

			valExpr, ok := p.parseExpr()
			if !ok {
				continue
			}
			val, ok := valExpr.(NumExpr)
			if !ok {
				p.error(fmt.Sprintf("expected number value for configuration %s", name))
				continue
			}
			prog.Configuration[name] = val.Value
		} else {
			p.error(fmt.Sprintf("unexpected token in configuration section: %s", p.current().String()))
			p.advance()
		}
	}
}

func (p *parser) parseConstants(prog *Program) {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == IDENT {
			p.parseConstant(prog)
		} else {
			p.error(fmt.Sprintf("unexpected token in constants section: %s", p.current().String()))
			p.advance()
		}
	}
}

func (p *parser) parseData(prog *Program) {
	banked := false // Default to common
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == COMMON {
			p.advance()
			if _, ok := p.expect(COLON, "expected : after common"); !ok {
				continue
			}
			banked = false
		} else if p.current().ty == BANKED {
			p.advance()
			if _, ok := p.expect(COLON, "expected : after banked"); !ok {
				continue
			}
			banked = true
		} else if p.current().ty == IDENT {
			nameTok := p.current()
			name := nameTok.val
			p.advance()

			if p.current().ty == I8 {
				p.advance()
				prog.Variables[name] = Variable{
					Name:   name,
					Type:   "i8",
					Banked: banked,
					Range:  nameTok.Range,
				}
			} else {
				p.error(fmt.Sprintf("expected type for variable %s, got %s", name, p.current().String()))
				p.advance()
			}
		} else {
			p.error(fmt.Sprintf("unexpected token in data section: %s", p.current().String()))
			p.advance()
		}
	}
}

func (p *parser) parseFunctions(prog *Program) {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == FN {
			fn, ok := p.parseFunction()
			if !ok {
				// Skip to next function or section
				for p.current().ty != EOF && p.current().ty != SECTION && p.current().ty != FN && p.current().ty != AT {
					p.advance()
				}
				continue
			}
			prog.Functions = append(prog.Functions, fn)
		} else if p.current().ty == AT {
			blk, ok := p.parseAtBlock()
			if !ok {
				// Skip to next function or section
				for p.current().ty != EOF && p.current().ty != SECTION && p.current().ty != FN && p.current().ty != AT {
					p.advance()
				}
				continue
			}
			prog.AtBlocks = append(prog.AtBlocks, blk)
		} else {
			p.error("unexpected token in program section")
			p.advance()
		}
	}
}

func (p *parser) parseAtBlock() (AtBlock, bool) {
	start := p.current().Range.Start
	p.advance() // eat AT
	addrExpr, ok := p.parseExpr()
	if !ok {
		return AtBlock{}, false
	}
	addr, ok := addrExpr.(NumExpr)
	if !ok {
		p.error("expected number address for at block")
		return AtBlock{}, false
	}

	if _, ok := p.expect(BEGIN, "expected begin after at address"); !ok {
		return AtBlock{}, false
	}

	stmts := []Stmt{}
	for p.current().ty != END && p.current().ty != EOF {
		stmt, ok := p.parseStmt()
		if !ok {
			p.advance()
			continue
		}
		stmts = append(stmts, stmt)
	}

	endTok, ok := p.expect(END, "expected end after at block")
	if !ok {
		return AtBlock{}, false
	}

	return AtBlock{Address: addr.Value, Body: stmts, Range: Range{Start: start, End: endTok.Range.End}}, true
}

func (p *parser) parseConstant(prog *Program) bool {
	// Declaration: ident: value [ ... ]
	name := p.current().val
	p.advance()
	if _, ok := p.expect(COLON, fmt.Sprintf("expected : after identifier %s", name)); !ok {
		return false
	}

	valExpr, ok := p.parseExpr()
	if !ok {
		return false
	}
	val, ok := valExpr.(NumExpr)
	if !ok {
		p.error(fmt.Sprintf("expected number value for constant %s", name))
		return false
	}

	if p.current().ty == LBRACK {
		// SFR definition
		p.advance()
		bits := make(map[string]int)
		for p.current().ty != RBRACK {
			bitName, ok := p.expect(IDENT, "expected bit name")
			if !ok {
				return false
			}
			if _, ok := p.expect(COLON, "expected : after bit name"); !ok {
				return false
			}
			bitValExpr, ok := p.parseExpr()
			if !ok {
				return false
			}
			bitVal, ok := bitValExpr.(NumExpr)
			if !ok {
				p.error(fmt.Sprintf("expected number value for bit %s", bitName.val))
				return false
			}
			bits[bitName.val] = bitVal.Value
		}
		p.advance() // ]
		prog.SFRs[name] = SFR{Address: val.Value, Bits: bits}
	} else {
		// Simple constant
		prog.Consts[name] = val.Value
	}
	return true
}

func (p *parser) parseFunction() (Function, bool) {
	// FN IDENT[name] LPAREN RPAREN BEGIN Stmt* END
	// TODO: should probably require functions to have at least one stmt
	result := Function{
		Body: []Stmt{},
	}
	start := p.current().Range.Start

	p.advance() // fn
	name := p.current()
	if name.ty != IDENT {
		p.error(fmt.Sprintf("I can't allow a function to go through life with the name '%v'", name.String()))
		return result, false
	}
	result.Name = name.val
	p.advance()

	if _, ok := p.expect(LPAREN, fmt.Sprintf("a function name must be followed by '(', not '%s'. It's tradition", p.current().String())); !ok {
		return result, false
	}
	if _, ok := p.expect(RPAREN, fmt.Sprintf("we can't go on until you close that parenthesis with ')'. '%s' won't do", p.current().String())); !ok {
		return result, false
	}

	if _, ok := p.expect(BEGIN, fmt.Sprintf("who starts a function with '%v'!? I just sat down!", p.current().String())); !ok {
		return result, false
	}

	for {
		switch p.current().ty {
		case EOF:
			p.error("alas, functions must come to an END")
			return result, false
		case END:
			result.Range = Range{Start: start, End: p.current().Range.End}
			p.advance()
			return result, true
		default:
			stmt, ok := p.parseStmt()
			if !ok {
				// Skip to next statement or end
				// Simple recovery: skip until semicolon (if we had them) or keyword
				// For now, maybe just skip one token? Or skip until a statement start?
				// Let's just skip one token for now to avoid infinite loops if we don't advance
				p.advance()
				continue
			}
			result.Body = append(result.Body, stmt)
		}
	}
}

func (p *parser) parseStmt() (Stmt, bool) {
	switch p.current().ty {
	case IDENT:
		if p.peekNext().ty == COLON {
			name := p.current().val
			start := p.current().Range.Start
			p.advance() // eat IDENT
			end := p.current().Range.End
			p.advance() // eat COLON
			return LabelStmt{Name: name, Range: Range{Start: start, End: end}}, true
		}
		if p.peekNext().ty == LPAREN {
			return p.parseCallStmt()
		}
		return p.parseAssignStmt()
	case RETURN:
		return p.parseReturnStmt()
	case IF:
		return p.parseIfStmt()
	default:
		p.error(fmt.Sprintf("unexpected token %s in statement", p.current().String()))
		return nil, false
	}
}

func (p *parser) parseIfStmt() (Stmt, bool) {
	start := p.current().Range.Start
	p.advance() // IF
	cond, ok := p.parseExpr()
	if !ok {
		return nil, false
	}

	if _, ok := p.expect(THEN, fmt.Sprintf("expected THEN, got %s", p.current().String())); !ok {
		return nil, false
	}

	stmt, ok := p.parseStmt()
	if !ok {
		return nil, false
	}

	return IfStmt{Cond: cond, Then: stmt, Range: Range{Start: start, End: stmt.Position().End}}, true
}

func (p *parser) parseAssignStmt() (Stmt, bool) {
	// IDENT = Expr
	name, ok := p.expect(IDENT, fmt.Sprintf("expected identifier, got %s", p.current().String()))
	if !ok {
		return nil, false
	}

	var lhs Expr = IdentExpr{Name: name.val, Range: name.Range}

	if p.current().ty == LBRACK {
		p.advance()
		idx, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		endTok, ok := p.expect(RBRACK, fmt.Sprintf("expected ], got %s", p.current().String()))
		if !ok {
			return nil, false
		}
		lhs = IndexExpr{Name: name.val, Index: idx, Range: Range{Start: name.Range.Start, End: endTok.Range.End}}
	}

	op := p.current()
	switch op.ty {
	case EQL, ADDEQL, SUBEQL, ANDEQL, OREQL, XOREQL:
		p.advance()
	default:
		p.error(fmt.Sprintf("expected assignment operator, got %s", op.String()))
		return nil, false
	}

	expr, ok := p.parseExpr()
	if !ok {
		return nil, false
	}

	return AssignStmt{
		Lhs:   lhs,
		Op:    op.ty,
		Expr:  expr,
		Range: Range{Start: lhs.Position().Start, End: expr.Position().End},
	}, true
}

func (p *parser) parseCallStmt() (Stmt, bool) {
	nameTok := p.current()
	name := nameTok.val
	p.advance() // eat IDENT
	p.advance() // eat LPAREN

	endTok, ok := p.expect(RPAREN, "expected ) after call arguments")
	if !ok {
		return nil, false
	}
	return CallStmt{Name: name, Range: Range{Start: nameTok.Range.Start, End: endTok.Range.End}}, true
}

func (p *parser) parseReturnStmt() (Stmt, bool) {
	// RETURN
	tok := p.current()
	p.advance()
	return ReturnStmt{Range: tok.Range}, true
}

func (p *parser) parseExpr() (Expr, bool) {
	return p.parseBinaryExpr()
}

func (p *parser) parseBinaryExpr() (Expr, bool) {
	lhs, ok := p.parseUnaryExpr()
	if !ok {
		return nil, false
	}

	for p.current().ty == NEQ {
		opTok := p.current()
		op := opTok.ty
		p.advance()
		rhs, ok := p.parseUnaryExpr()
		if !ok {
			return nil, false
		}
		lhs = BinaryExpr{Lhs: lhs, Op: op, Rhs: rhs, Range: Range{Start: lhs.Position().Start, End: rhs.Position().End}}
	}
	return lhs, true
}

func (p *parser) parseUnaryExpr() (Expr, bool) {
	if p.current().ty == NOT {
		opTok := p.current()
		op := opTok.ty
		p.advance()
		expr, ok := p.parseUnaryExpr()
		if !ok {
			return nil, false
		}
		return UnaryExpr{Op: op, Expr: expr, Range: Range{Start: opTok.Range.Start, End: expr.Position().End}}, true
	}
	return p.parsePostfixExpr()
}

func (p *parser) parsePostfixExpr() (Expr, bool) {
	lhs, ok := p.parsePrimaryExpr()
	if !ok {
		return nil, false
	}

	for {
		switch p.current().ty {
		case INC, DEC:
			opTok := p.current()
			op := opTok.ty
			p.advance()
			lhs = PostfixExpr{Expr: lhs, Op: op, Range: Range{Start: lhs.Position().Start, End: opTok.Range.End}}
		case LBRACK:
			if _, ok := lhs.(IdentExpr); !ok {
				// Indexing only supported on identifiers.
				// If we see a bracket after something else (like a number),
				// it's not part of the expression.
				return lhs, true
			}
			p.advance()
			idx, ok := p.parseExpr()
			if !ok {
				return nil, false
			}
			endTok, ok := p.expect(RBRACK, fmt.Sprintf("expected ], got %s", p.current().String()))
			if !ok {
				return nil, false
			}
			if id, ok := lhs.(IdentExpr); ok {
				lhs = IndexExpr{Name: id.Name, Index: idx, Range: Range{Start: id.Range.Start, End: endTok.Range.End}}
			} else {
				// Should be unreachable due to check above
				p.error("indexing only supported on identifiers")
				return nil, false
			}
		default:
			return lhs, true
		}
	}
}

func (p *parser) parsePrimaryExpr() (Expr, bool) {
	tok := p.current()
	if tok.ty >= NUM_First && tok.ty <= NUM_Last {
		val, err := tok.Number()
		if err != nil {
			p.error(fmt.Sprintf("invalid number %q: %v", tok.val, err))
			return nil, false
		}
		p.advance()
		return NumExpr{Val: tok.val, Value: val, Ty: tok.ty, Range: tok.Range}, true
	}
	switch tok.ty {
	case IDENT:
		p.advance()
		return IdentExpr{Name: tok.val, Range: tok.Range}, true
	case LPAREN:
		p.advance()
		expr, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		_, ok = p.expect(RPAREN, fmt.Sprintf("expected ), got %s", p.current().String()))
		if !ok {
			return nil, false
		}
		return expr, true
	default:
		p.error(fmt.Sprintf("unexpected token %s in expression", p.current().String()))
		return nil, false
	}
}

func isW(name string) bool {
	return strings.ToLower(name) == "w"
}

func getIdent(e Expr) (string, bool) {
	if id, ok := e.(IdentExpr); ok {
		return id.Name, true
	}
	return "", false
}

func getNum(e Expr) (int, bool) {
	if num, ok := e.(NumExpr); ok {
		return num.Value, true
	}
	return 0, false
}
