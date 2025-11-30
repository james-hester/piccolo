package internal

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func Compile(prog Program) ([]PicOp, SymbolTable, error) {
	// Allocate variables
	commonAddr := 0x70
	bankedAddr := 0x20

	syms := NewSymbolTable()

	var names []string
	for name := range prog.Variables {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		v := prog.Variables[name]
		if v.Banked {
			v.Address = bankedAddr
			bankedAddr++
		} else {
			v.Address = commonAddr
			commonAddr++
		}
		prog.Variables[name] = v
		syms.SetAddress(name, v.Address)
	}

	c := &asmGen{prog: prog}
	var ops []PicOp

	// Configuration
	// TODO: Emit configuration bits

	// AtBlocks
	for _, blk := range prog.AtBlocks {
		ops = append(ops, OrgOp{Address: blk.Address})
		for _, stmt := range blk.Body {
			compiled, err := c.compileStmt(stmt)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, compiled...)
		}
	}

	for _, fn := range prog.Functions {
		ops = append(ops, LabelOp{Name: fn.Name})
		for _, stmt := range fn.Body {
			compiled, err := c.compileStmt(stmt)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, compiled...)
		}
	}
	return ops, syms, nil
}

type asmGen struct {
	prog Program
}

func (c *asmGen) resolveBit(name string, idx Expr) (int, error) {
	// 1. Try literal number
	if num, ok := idx.(NumExpr); ok {
		return num.Value, nil
	}
	// 2. Try identifier in SFR
	if id, ok := idx.(IdentExpr); ok {
		if sfr, ok := c.prog.SFRs[name]; ok {
			if bit, ok := sfr.Bits[id.Name]; ok {
				return bit, nil
			}
		}
		// Also check global constants? Maybe not for bits.
	}
	return 0, fmt.Errorf("cannot resolve bit index %v for %s", idx, name)
}

func (c *asmGen) resolveAddr(name string) (string, error) {
	// If it's in Consts or SFRs, resolve to address string
	if val, ok := c.prog.Consts[name]; ok {
		return fmt.Sprintf("0x%X", val), nil
	}
	if sfr, ok := c.prog.SFRs[name]; ok {
		return fmt.Sprintf("0x%X", sfr.Address), nil
	}
	if v, ok := c.prog.Variables[name]; ok {
		return fmt.Sprintf("0x%X", v.Address), nil
	}
	// Otherwise return name as is (might be handled by assembler later or is W/FSR)
	return name, nil
}

func (c *asmGen) compileStmt(stmt Stmt) ([]PicOp, error) {
	switch s := stmt.(type) {
	case AssignStmt:
		return c.compileAssign(s)
	case IfStmt:
		return c.compileIf(s)
	case ReturnStmt:
		return []PicOp{Return{}}, nil
	case CallStmt:
		return []PicOp{CallOp{Label: s.Name}}, nil
	case LabelStmt:
		return []PicOp{LabelOp(s)}, nil
	default:
		return nil, fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *asmGen) compileIf(s IfStmt) ([]PicOp, error) {
	cond := s.Cond

	isZero := func(e Expr) bool {
		val, ok := getNum(e)
		return ok && val == 0
	}

	// 1. f[b] -> BTFSC f,b
	if idx, ok := cond.(IndexExpr); ok {
		b, err := c.resolveBit(idx.Name, idx.Index)
		if err == nil {
			f, _ := c.resolveAddr(idx.Name)
			ops := []PicOp{Btfsc{F: f, B: b}}
			bodyOps, err := c.compileStmt(s.Then)
			if err != nil {
				return nil, err
			}
			return append(ops, bodyOps...), nil
		}
	}

	// 2. not f[b] -> BTFSS f,b
	if unary, ok := cond.(UnaryExpr); ok && unary.Op == NOT {
		if idx, ok := unary.Expr.(IndexExpr); ok {
			b, err := c.resolveBit(idx.Name, idx.Index)
			if err == nil {
				f, _ := c.resolveAddr(idx.Name)
				ops := []PicOp{Btfss{F: f, B: b}}
				bodyOps, err := c.compileStmt(s.Then)
				if err != nil {
					return nil, err
				}
				return append(ops, bodyOps...), nil
			}
		}
	}

	// 3. (f--) != 0 -> DECFSZ f,1
	// 4. (f++) != 0 -> INCFSZ f,1
	if bin, ok := cond.(BinaryExpr); ok && bin.Op == NEQ {
		if isZero(bin.Rhs) {
			if post, ok := bin.Lhs.(PostfixExpr); ok {
				if id, ok := post.Expr.(IdentExpr); ok {
					f, _ := c.resolveAddr(id.Name)
					switch post.Op {
					case DEC:
						ops := []PicOp{Decfsz{F: f, D: DestF}}
						bodyOps, err := c.compileStmt(s.Then)
						if err != nil {
							return nil, err
						}
						return append(ops, bodyOps...), nil
					case INC:
						ops := []PicOp{Incfsz{F: f, D: DestF}}
						bodyOps, err := c.compileStmt(s.Then)
						if err != nil {
							return nil, err
						}
						return append(ops, bodyOps...), nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("unsupported if condition: %v", cond)
}

func (c *asmGen) compileAssign(s AssignStmt) ([]PicOp, error) {
	lhsExpr := s.Lhs
	op := s.Op
	rhs := s.Expr

	// Handle f[b] = 0/1
	if idx, ok := lhsExpr.(IndexExpr); ok {
		if op == EQL {
			b, err := c.resolveBit(idx.Name, idx.Index)
			if err == nil {
				f, _ := c.resolveAddr(idx.Name)
				if val, ok := getNum(rhs); ok {
					switch val {
					case 0:
						return []PicOp{Bcf{F: f, B: b}}, nil
					case 1:
						return []PicOp{Bsf{F: f, B: b}}, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("unsupported assignment to index: %v", s)
	}

	// Handle Ident assignments
	lhsName, ok := getIdent(lhsExpr)
	if !ok {
		return nil, fmt.Errorf("LHS must be identifier or index expression")
	}

	switch op {
	case EQL: // =
		if isW(lhsName) {
			// w = k -> MOVLW k
			if k, ok := getNum(rhs); ok {
				return []PicOp{Movlw{K: k}}, nil
			}
			// w = f -> MOVF f,0
			if name, ok := getIdent(rhs); ok {
				f, _ := c.resolveAddr(name)
				return []PicOp{Movf{F: f, D: DestW}}, nil
			}
		} else {
			// f = w -> MOVWF f
			if name, ok := getIdent(rhs); ok && isW(name) {
				f, _ := c.resolveAddr(lhsName)
				return []PicOp{Movwf{F: f}}, nil
			}
			// f = k -> MOVLW k; MOVWF f
			if k, ok := getNum(rhs); ok {
				f, _ := c.resolveAddr(lhsName)
				return []PicOp{
					Movlw{K: k},
					Movwf{F: f},
				}, nil
			}
			// f1 = f2 -> MOVF f2,0; MOVWF f1
			if name, ok := getIdent(rhs); ok {
				f1, _ := c.resolveAddr(lhsName)
				f2, _ := c.resolveAddr(name)
				return []PicOp{
					Movf{F: f2, D: DestW},
					Movwf{F: f1},
				}, nil
			}
		}

	case ADDEQL: // +=
		if isW(lhsName) {
			// w += f -> ADDWF f,0
			if name, ok := getIdent(rhs); ok {
				f, _ := c.resolveAddr(name)
				return []PicOp{Addwf{F: f, D: DestW}}, nil
			}
		} else if strings.HasPrefix(strings.ToLower(lhsName), "fsr") {
			// fsrn += k -> ADDFSR fsrn, k
			fsrStr := strings.ToLower(lhsName)
			fsrNumStr := strings.TrimPrefix(fsrStr, "fsr")
			fsrNum, err := strconv.Atoi(fsrNumStr)
			if err == nil {
				if k, ok := getNum(rhs); ok {
					return []PicOp{Addfsr{FSR: fsrNum, K: k}}, nil
				}
			}
		} else {
			// f += w -> ADDWF f,1
			if name, ok := getIdent(rhs); ok && isW(name) {
				f, _ := c.resolveAddr(lhsName)
				return []PicOp{Addwf{F: f, D: DestF}}, nil
			}
		}

	case SUBEQL: // -=
		if strings.HasPrefix(strings.ToLower(lhsName), "fsr") {
			// fsrn -= k -> ADDFSR fsrn, -k
			fsrStr := strings.ToLower(lhsName)
			fsrNumStr := strings.TrimPrefix(fsrStr, "fsr")
			fsrNum, err := strconv.Atoi(fsrNumStr)
			if err == nil {
				if k, ok := getNum(rhs); ok {
					return []PicOp{Addfsr{FSR: fsrNum, K: -k}}, nil
				}
			}
		}

	case ANDEQL: // &=
		if isW(lhsName) {
			// w &= k -> ANDLW k
			if k, ok := getNum(rhs); ok {
				return []PicOp{Andlw{K: k}}, nil
			}
			// w &= f -> ANDWF f,0
			if name, ok := getIdent(rhs); ok {
				f, _ := c.resolveAddr(name)
				return []PicOp{Andwf{F: f, D: DestW}}, nil
			}
		} else {
			// f &= w -> ANDWF f,1
			if name, ok := getIdent(rhs); ok && isW(name) {
				f, _ := c.resolveAddr(lhsName)
				return []PicOp{Andwf{F: f, D: DestF}}, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot compile assignment: %v %v %v", lhsName, op, rhs)
}
