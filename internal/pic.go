package internal

import (
	"fmt"
	"strconv"
)

// PicOp represents a single PIC assembly instruction.
type PicOp interface {
	Assembly() string
	Encode(ctx *AssemblerContext) error
}

func resolveAddr(ctx *AssemblerContext, s string) (int, error) {
	// Try to parse as number
	if val, err := strconv.ParseInt(s, 0, 64); err == nil {
		return int(val), nil
	}
	// Try symbol table
	if addr, ok := ctx.Symbols.GetAddress(s); ok {
		return addr, nil
	}
	return 0, fmt.Errorf("cannot resolve address: %s", s)
}

// W and F are the destination flags for many file register operations.
const (
	DestW = 0
	DestF = 1
)

// LabelOp is a pseudo-op for labels
type LabelOp struct {
	Name string
}

func (op LabelOp) Assembly() string {
	return op.Name + ":"
}

func (op LabelOp) Encode(ctx *AssemblerContext) error {
	// Record label address (current PC)
	// PC is len(ctx.Words)
	ctx.Symbols.SetAddress(op.Name, len(ctx.Words))
	// Reset bank state because we don't know where we came from
	ctx.CurrentBank = -1
	return nil
}

// OrgOp sets the location counter
type OrgOp struct {
	Address int
}

func (op OrgOp) Assembly() string {
	return fmt.Sprintf(" ORG 0x%X", op.Address)
}

func (op OrgOp) Encode(ctx *AssemblerContext) error {
	// Pad with NOPs or change logic to support gaps?
	// For now, let's just pad with NOPs (0x0000) until we reach the address.
	currentPC := len(ctx.Words)
	if op.Address < currentPC {
		return fmt.Errorf("cannot ORG backwards: current %x, target %x", currentPC, op.Address)
	}
	for i := currentPC; i < op.Address; i++ {
		ctx.Emit(0x0000) // NOP
	}
	return nil
}

// CallOp calls a subroutine
type CallOp struct {
	Label string
}

func (op CallOp) Assembly() string {
	return fmt.Sprintf(" CALL %s", op.Label)
}

func (op CallOp) Encode(ctx *AssemblerContext) error {
	// CALL is 10 0kkk kkkk kkkk
	ctx.AddFixup(op.Label, 0x07FF)
	ctx.Emit(0x2000)
	// Reset bank state after call?
	// Usually CALL preserves bank or callee handles it.
	// But when we return, we might be in a different bank if callee changed it and didn't restore.
	// Safe assumption: bank is unknown after CALL.
	ctx.CurrentBank = -1
	return nil
}

// ADDWF f,d
// Add W to F
type Addwf struct {
	F string
	D int
}

func (op Addwf) Assembly() string {
	return fmt.Sprintf("ADDWF %s,%d", op.F, op.D)
}

func (op Addwf) Encode(ctx *AssemblerContext) error {
	// 00 0111 dfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0700 | (uint16(op.D&1) << 7) | (uint16(f) & 0x7F))
	return nil
}

// ADDFSR fsrn, k
// Add literal k to FSRn
type Addfsr struct {
	FSR int
	K   int
}

func (op Addfsr) Assembly() string {
	return fmt.Sprintf("ADDFSR %d,%d", op.FSR, op.K)
}

func (op Addfsr) Encode(ctx *AssemblerContext) error {
	// 11 0001 0nkk kkkk
	k := op.K & 0x3F
	n := op.FSR & 1
	ctx.Emit(0x3100 | (uint16(n) << 6) | uint16(k))
	return nil
}

// ANDLW k
// AND literal with W
type Andlw struct {
	K int
}

func (op Andlw) Assembly() string {
	return fmt.Sprintf("ANDLW %d", op.K)
}

func (op Andlw) Encode(ctx *AssemblerContext) error {
	// 11 1001 kkkk kkkk
	ctx.Emit(0x3900 | (uint16(op.K) & 0xFF))
	return nil
}

// ANDWF f,d
// AND W with F
type Andwf struct {
	F string
	D int
}

func (op Andwf) Assembly() string {
	return fmt.Sprintf("ANDWF %s,%d", op.F, op.D)
}

func (op Andwf) Encode(ctx *AssemblerContext) error {
	// 00 0101 dfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0500 | (uint16(op.D&1) << 7) | (uint16(f) & 0x7F))
	return nil
}

// BCF f,b
// Bit Clear F
type Bcf struct {
	F string
	B int
}

func (op Bcf) Assembly() string {
	return fmt.Sprintf("BCF %s,%d", op.F, op.B)
}

func (op Bcf) Encode(ctx *AssemblerContext) error {
	// 01 00bb bfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x1000 | (uint16(op.B&7) << 7) | (uint16(f) & 0x7F))
	return nil
}

// BSF f,b
// Bit Set F
type Bsf struct {
	F string
	B int
}

func (op Bsf) Assembly() string {
	return fmt.Sprintf("BSF %s,%d", op.F, op.B)
}

func (op Bsf) Encode(ctx *AssemblerContext) error {
	// 01 01bb bfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x1400 | (uint16(op.B&7) << 7) | (uint16(f) & 0x7F))
	return nil
}

// BTFSC f,b
// Bit Test F, Skip if Clear
type Btfsc struct {
	F string
	B int
}

func (op Btfsc) Assembly() string {
	return fmt.Sprintf("BTFSC %s,%d", op.F, op.B)
}

func (op Btfsc) Encode(ctx *AssemblerContext) error {
	// 01 10bb bfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x1800 | (uint16(op.B&7) << 7) | (uint16(f) & 0x7F))
	return nil
}

// BTFSS f,b
// Bit Test F, Skip if Set
type Btfss struct {
	F string
	B int
}

func (op Btfss) Assembly() string {
	return fmt.Sprintf("BTFSS %s,%d", op.F, op.B)
}

func (op Btfss) Encode(ctx *AssemblerContext) error {
	// 01 11bb bfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x1C00 | (uint16(op.B&7) << 7) | (uint16(f) & 0x7F))
	return nil
}

// DECFSZ f,d
// Decrement F, Skip if Zero
type Decfsz struct {
	F string
	D int
}

func (op Decfsz) Assembly() string {
	return fmt.Sprintf("DECFSZ %s,%d", op.F, op.D)
}

func (op Decfsz) Encode(ctx *AssemblerContext) error {
	// 00 1011 dfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0B00 | (uint16(op.D&1) << 7) | (uint16(f) & 0x7F))
	return nil
}

// INCFSZ f,d
// Increment F, Skip if Zero
type Incfsz struct {
	F string
	D int
}

func (op Incfsz) Assembly() string {
	return fmt.Sprintf("INCFSZ %s,%d", op.F, op.D)
}

func (op Incfsz) Encode(ctx *AssemblerContext) error {
	// 00 1111 dfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0F00 | (uint16(op.D&1) << 7) | (uint16(f) & 0x7F))
	return nil
}

// RETURN
// Return from Subroutine
type Return struct{}

func (op Return) Assembly() string {
	return "RETURN"
}

func (op Return) Encode(ctx *AssemblerContext) error {
	// 00 0000 0000 1000
	ctx.Emit(0x0008)
	// Bank unknown after return?
	// No, return pops PC, doesn't change bank.
	// But we are returning to caller.
	// The caller's bank state is what matters.
	// Since we are in a linear flow here, we don't know where we return to.
	// But this instruction itself doesn't change bank.
	// However, execution stops here (or goes to caller).
	// So subsequent instructions (if any, unreachable?) start with unknown bank?
	// Let's say yes.
	ctx.CurrentBank = -1
	return nil
}

// MOVLW k
// Move Literal to W
type Movlw struct {
	K int
}

func (op Movlw) Assembly() string {
	return fmt.Sprintf("MOVLW %d", op.K)
}

func (op Movlw) Encode(ctx *AssemblerContext) error {
	// 11 00xx kkkk kkkk
	// 11 0000 kkkk kkkk (MOVLW)
	ctx.Emit(0x3000 | (uint16(op.K) & 0xFF))
	return nil
}

// MOVF f,d
// Move F
type Movf struct {
	F string
	D int
}

func (op Movf) Assembly() string {
	return fmt.Sprintf("MOVF %s,%d", op.F, op.D)
}

func (op Movf) Encode(ctx *AssemblerContext) error {
	// 00 1000 dfff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0800 | (uint16(op.D&1) << 7) | (uint16(f) & 0x7F))
	return nil
}

// MOVWF f
// Move W to F
type Movwf struct {
	F string
}

func (op Movwf) Assembly() string {
	return fmt.Sprintf("MOVWF %s", op.F)
}

func (op Movwf) Encode(ctx *AssemblerContext) error {
	// 00 0000 1fff ffff
	f, err := resolveAddr(ctx, op.F)
	if err != nil {
		return err
	}
	ctx.EnsureBank(f)
	ctx.Emit(0x0080 | (uint16(f) & 0x7F))
	return nil
}

// GOTO k
// Go to address
type Goto struct {
	Label string
}

func (op Goto) Assembly() string {
	return fmt.Sprintf(" GOTO %s", op.Label)
}

func (op Goto) Encode(ctx *AssemblerContext) error {
	// 10 1kkk kkkk kkkk
	ctx.AddFixup(op.Label, 0x07FF)
	ctx.Emit(0x2800)
	ctx.CurrentBank = -1
	return nil
}
