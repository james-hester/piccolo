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
}

type Expr interface {
	String() string
	isExpr()
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
}

type AtBlock struct {
	Address int
	Body    []Stmt
}

type SFR struct {
	Address int
	Bits    map[string]int
}

type Function struct {
	Name string
	Body []Stmt
}

type CallStmt struct {
	Name string
}

func (CallStmt) isStmt() {}

func (c CallStmt) String() string {
	return fmt.Sprintf("call %s", c.Name)
}

type AssignStmt struct {
	Lhs  Expr
	Op   TTy
	Expr Expr
}

func (AssignStmt) isStmt() {}

type IfStmt struct {
	Cond Expr
	Then Stmt
}

func (IfStmt) isStmt() {}

type ReturnStmt struct{}

func (ReturnStmt) isStmt() {}

type LabelStmt struct {
	Name string
}

func (LabelStmt) isStmt() {}

type IdentExpr struct {
	Name string
}

func (IdentExpr) isExpr() {}

type IndexExpr struct {
	Name  string
	Index Expr
}

func (IndexExpr) isExpr() {}

type NumExpr struct {
	Val   string
	Value int
	Ty    TTy
}

func (NumExpr) isExpr() {}

type UnaryExpr struct {
	Op   TTy
	Expr Expr
}

func (UnaryExpr) isExpr() {}

type BinaryExpr struct {
	Lhs Expr
	Op  TTy
	Rhs Expr
}

func (BinaryExpr) isExpr() {}

type PostfixExpr struct {
	Expr Expr
	Op   TTy
}

func (PostfixExpr) isExpr() {}

func Parse(tokens []Tok) (Program, error) {
	p := newParser(tokens)
	return p.parseProgram()
}

type parser struct {
	toks []Tok
	pos  int
}

func newParser(toks []Tok) *parser {
	return &parser{toks: toks}
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

func (p *parser) expect(ty TTy, formatter func(Tok) error) (Tok, error) {
	tok := p.current()
	if tok.ty != ty {
		return Tok{}, formatter(tok)
	}
	p.pos++
	return tok, nil
}

func (p *parser) parseProgram() (Program, error) {
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
				if err := p.parseConstants(&result); err != nil {
					return result, err
				}
			case CONFIGURATION:
				p.advance()
				if err := p.parseConfiguration(&result); err != nil {
					return result, err
				}
			case DATA:
				p.advance()
				if err := p.parseData(&result); err != nil {
					return result, err
				}
			case PROGRAM:
				p.advance()
				if err := p.parseFunctions(&result); err != nil {
					return result, err
				}
			default:
				return result, fmt.Errorf("unknown section type: %s", p.current().String())
			}
		} else {
			return result, fmt.Errorf("unexpected token at top level: %v", p.current().String())
		}
	}

	return result, nil
}

func (p *parser) parseConfiguration(prog *Program) error {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == IDENT {
			name := p.current().val
			p.advance()
			if _, err := p.expect(COLON, func(t Tok) error {
				return fmt.Errorf("expected : after identifier %s", name)
			}); err != nil {
				return err
			}

			valExpr, err := p.parseExpr()
			if err != nil {
				return err
			}
			val, ok := valExpr.(NumExpr)
			if !ok {
				return fmt.Errorf("expected number value for configuration %s", name)
			}
			prog.Configuration[name] = val.Value
		} else {
			return fmt.Errorf("unexpected token in configuration section: %s", p.current().String())
		}
	}
	return nil
}

func (p *parser) parseConstants(prog *Program) error {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == IDENT {
			if err := p.parseConstant(prog); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected token in constants section: %s", p.current().String())
		}
	}
	return nil
}

func (p *parser) parseData(prog *Program) error {
	banked := false // Default to common
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == COMMON {
			p.advance()
			if _, err := p.expect(COLON, func(t Tok) error {
				return fmt.Errorf("expected : after common")
			}); err != nil {
				return err
			}
			banked = false
		} else if p.current().ty == BANKED {
			p.advance()
			if _, err := p.expect(COLON, func(t Tok) error {
				return fmt.Errorf("expected : after banked")
			}); err != nil {
				return err
			}
			banked = true
		} else if p.current().ty == IDENT {
			name := p.current().val
			p.advance()

			if p.current().ty == I8 {
				p.advance()
				prog.Variables[name] = Variable{
					Name:   name,
					Type:   "i8",
					Banked: banked,
				}
			} else {
				return fmt.Errorf("expected type for variable %s, got %s", name, p.current().String())
			}
		} else {
			return fmt.Errorf("unexpected token in data section: %s", p.current().String())
		}
	}
	return nil
}

func (p *parser) parseFunctions(prog *Program) error {
	for p.current().ty != EOF && p.current().ty != SECTION {
		if p.current().ty == FN {
			fn, err := p.parseFunction()
			if err != nil {
				return err
			}
			prog.Functions = append(prog.Functions, fn)
		} else if p.current().ty == AT {
			blk, err := p.parseAtBlock()
			if err != nil {
				return err
			}
			prog.AtBlocks = append(prog.AtBlocks, blk)
		} else {
			return fmt.Errorf("unexpected token in program section: %s", p.current().String())
		}
	}
	return nil
}

func (p *parser) parseAtBlock() (AtBlock, error) {
	p.advance() // eat AT
	addrExpr, err := p.parseExpr()
	if err != nil {
		return AtBlock{}, err
	}
	addr, ok := addrExpr.(NumExpr)
	if !ok {
		return AtBlock{}, fmt.Errorf("expected number address for at block")
	}

	if _, err := p.expect(BEGIN, func(t Tok) error {
		return fmt.Errorf("expected begin after at address")
	}); err != nil {
		return AtBlock{}, err
	}

	stmts := []Stmt{}
	for p.current().ty != END && p.current().ty != EOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return AtBlock{}, err
		}
		stmts = append(stmts, stmt)
	}

	if _, err := p.expect(END, func(t Tok) error {
		return fmt.Errorf("expected end after at block")
	}); err != nil {
		return AtBlock{}, err
	}

	return AtBlock{Address: addr.Value, Body: stmts}, nil
}

func (p *parser) parseConstant(prog *Program) error {
	// Declaration: ident: value [ ... ]
	name := p.current().val
	p.advance()
	if _, err := p.expect(COLON, func(t Tok) error {
		return fmt.Errorf("expected : after identifier %s", name)
	}); err != nil {
		return err
	}

	valExpr, err := p.parseExpr()
	if err != nil {
		return err
	}
	val, ok := valExpr.(NumExpr)
	if !ok {
		return fmt.Errorf("expected number value for constant %s", name)
	}

	if p.current().ty == LBRACK {
		// SFR definition
		p.advance()
		bits := make(map[string]int)
		for p.current().ty != RBRACK {
			bitName, err := p.expect(IDENT, func(t Tok) error {
				return fmt.Errorf("expected bit name")
			})
			if err != nil {
				return err
			}
			if _, err := p.expect(COLON, func(t Tok) error {
				return fmt.Errorf("expected : after bit name")
			}); err != nil {
				return err
			}
			bitValExpr, err := p.parseExpr()
			if err != nil {
				return err
			}
			bitVal, ok := bitValExpr.(NumExpr)
			if !ok {
				return fmt.Errorf("expected number value for bit %s", bitName.val)
			}
			bits[bitName.val] = bitVal.Value
		}
		p.advance() // ]
		prog.SFRs[name] = SFR{Address: val.Value, Bits: bits}
	} else {
		// Simple constant
		prog.Consts[name] = val.Value
	}
	return nil
}

func (p *parser) parseFunction() (Function, error) {
	// FN IDENT[name] LPAREN RPAREN BEGIN Stmt* END
	// TODO: should probably require functions to have at least one stmt
	result := Function{
		Body: []Stmt{},
	}

	p.advance() // fn
	name := p.current()
	if name.ty != IDENT {
		//lint:ignore ST1005 "I" must be capitalized
		return result, fmt.Errorf("I can't allow a function to go through life with the name '%v'", name.String())
	}
	result.Name = name.val
	p.advance()

	if _, err := p.expect(LPAREN, func(t Tok) error {
		return fmt.Errorf("a function name must be followed by '(', not '%s'. It's tradition", t.String())
	}); err != nil {
		return result, err
	}
	if _, err := p.expect(RPAREN, func(t Tok) error {
		return fmt.Errorf("we can't go on until you close that parenthesis with ')'. '%s' won't do", t.String())
	}); err != nil {
		return result, err
	}

	if _, err := p.expect(BEGIN, func(t Tok) error {
		//lint:ignore ST1005 silly message
		return fmt.Errorf("who starts a function with '%v'!? I just sat down!", t.String())
	}); err != nil {
		return result, err
	}

	for {
		switch p.current().ty {
		case EOF:
			return result, fmt.Errorf("alas, functions must come to an END")
		case END:
			p.advance()
			return result, nil
		default:
			stmt, err := p.parseStmt()
			if err != nil {
				return result, err
			}
			result.Body = append(result.Body, stmt)
		}
	}
}

func (p *parser) parseStmt() (Stmt, error) {
	switch p.current().ty {
	case IDENT:
		if p.peekNext().ty == COLON {
			name := p.current().val
			p.advance() // eat IDENT
			p.advance() // eat COLON
			return LabelStmt{Name: name}, nil
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
		return nil, fmt.Errorf("unexpected token %s in statement", p.current().String())
	}
}

func (p *parser) parseIfStmt() (Stmt, error) {
	p.advance() // IF
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(THEN, func(t Tok) error {
		return fmt.Errorf("expected THEN, got %s", t.String())
	})
	if err != nil {
		return nil, err
	}

	stmt, err := p.parseStmt()
	if err != nil {
		return nil, err
	}

	return IfStmt{Cond: cond, Then: stmt}, nil
}

func (p *parser) parseAssignStmt() (Stmt, error) {
	// IDENT = Expr
	name, err := p.expect(IDENT, func(t Tok) error {
		return fmt.Errorf("expected identifier, got %s", t.String())
	})
	if err != nil {
		return nil, err
	}

	var lhs Expr = IdentExpr{Name: name.val}

	if p.current().ty == LBRACK {
		p.advance()
		idx, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(RBRACK, func(t Tok) error {
			return fmt.Errorf("expected ], got %s", t.String())
		})
		if err != nil {
			return nil, err
		}
		lhs = IndexExpr{Name: name.val, Index: idx}
	}

	op := p.current()
	switch op.ty {
	case EQL, ADDEQL, SUBEQL, ANDEQL, OREQL, XOREQL:
		p.advance()
	default:
		return nil, fmt.Errorf("expected assignment operator, got %s", op.String())
	}

	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return AssignStmt{
		Lhs:  lhs,
		Op:   op.ty,
		Expr: expr,
	}, nil
}

func (p *parser) parseCallStmt() (Stmt, error) {
	name := p.current().val
	p.advance() // eat IDENT
	p.advance() // eat LPAREN

	if _, err := p.expect(RPAREN, func(t Tok) error {
		return fmt.Errorf("expected ) after call arguments")
	}); err != nil {
		return nil, err
	}
	return CallStmt{Name: name}, nil
}

func (p *parser) parseReturnStmt() (Stmt, error) {
	// RETURN
	p.advance()
	return ReturnStmt{}, nil
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseBinaryExpr()
}

func (p *parser) parseBinaryExpr() (Expr, error) {
	lhs, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for p.current().ty == NEQ {
		op := p.current().ty
		p.advance()
		rhs, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		lhs = BinaryExpr{Lhs: lhs, Op: op, Rhs: rhs}
	}
	return lhs, nil
}

func (p *parser) parseUnaryExpr() (Expr, error) {
	if p.current().ty == NOT {
		op := p.current().ty
		p.advance()
		expr, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return UnaryExpr{Op: op, Expr: expr}, nil
	}
	return p.parsePostfixExpr()
}

func (p *parser) parsePostfixExpr() (Expr, error) {
	lhs, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		switch p.current().ty {
		case INC, DEC:
			op := p.current().ty
			p.advance()
			lhs = PostfixExpr{Expr: lhs, Op: op}
		case LBRACK:
			if _, ok := lhs.(IdentExpr); !ok {
				// Indexing only supported on identifiers.
				// If we see a bracket after something else (like a number),
				// it's not part of the expression.
				return lhs, nil
			}
			p.advance()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(RBRACK, func(t Tok) error {
				return fmt.Errorf("expected ], got %s", t.String())
			})
			if err != nil {
				return nil, err
			}
			if id, ok := lhs.(IdentExpr); ok {
				lhs = IndexExpr{Name: id.Name, Index: idx}
			} else {
				// Should be unreachable due to check above
				return nil, fmt.Errorf("indexing only supported on identifiers")
			}
		default:
			return lhs, nil
		}
	}
}

func (p *parser) parsePrimaryExpr() (Expr, error) {
	tok := p.current()
	if tok.ty >= NUM_First && tok.ty <= NUM_Last {
		p.advance()
		val, err := tok.Number()
		if err != nil {
			return nil, fmt.Errorf("invalid number %q: %w", tok.val, err)
		}
		return NumExpr{Val: tok.val, Value: val, Ty: tok.ty}, nil
	}
	switch tok.ty {
	case IDENT:
		p.advance()
		return IdentExpr{Name: tok.val}, nil
	case LPAREN:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(RPAREN, func(t Tok) error {
			return fmt.Errorf("expected ), got %s", t.String())
		})
		if err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token %s in expression", p.current().String())
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
