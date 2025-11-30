package internal

import (
	"fmt"
	"io"
)

// AssemblerContext holds the state for the assembly process.
type AssemblerContext struct {
	Words       []uint16
	Symbols     SymbolTable
	CurrentBank int
	Fixups      []Fixup
}

// Fixup represents a location in the code that needs to be patched with a label address.
type Fixup struct {
	Index int    // Index in Words
	Label string // Label to resolve
	Mask  uint16 // Mask to apply (usually 0x7FF for GOTO/CALL)
}

// NewAssemblerContext creates a new context.
func NewAssemblerContext(syms SymbolTable) *AssemblerContext {
	if syms == nil {
		syms = NewSymbolTable()
	}
	return &AssemblerContext{
		Symbols:     syms,
		CurrentBank: -1, // Unknown bank
	}
}

// Emit adds a word to the output.
func (ctx *AssemblerContext) Emit(word uint16) {
	ctx.Words = append(ctx.Words, word)
}

// EnsureBank emits a MOVLB instruction if the address is in a different bank.
func (ctx *AssemblerContext) EnsureBank(addr int) {
	// Common RAM (0x70-0x7F) is accessible from any bank on many PICs.
	// We assume standard banking for PIC16F1xxx.
	// Bank size is 128 bytes (0x80).

	// If addr is common (0x70-0x7F), we don't need to switch.
	if addr >= 0x70 && addr <= 0x7F {
		return
	}

	bank := (addr >> 7) & 0x1F // 32 banks max usually
	if ctx.CurrentBank != bank {
		// Emit MOVLB bank
		// MOVLB k: 00 0000 001k kkkk
		// Note: MOVLB opcode is 0x0020 | k
		ctx.Emit(0x0020 | uint16(bank))
		ctx.CurrentBank = bank
	}
}

// AddFixup records a fixup for a label.
func (ctx *AssemblerContext) AddFixup(label string, mask uint16) {
	ctx.Fixups = append(ctx.Fixups, Fixup{
		Index: len(ctx.Words), // The word we are about to emit
		Label: label,
		Mask:  mask,
	})
}

// Assemble converts a list of PicOp into machine code using the context.
func Assemble(ops []PicOp, syms SymbolTable) ([]uint16, error) {
	ctx := NewAssemblerContext(syms)

	// Pass 1: Emit code and collect fixups
	for _, op := range ops {
		if err := op.Encode(ctx); err != nil {
			return nil, err
		}
	}

	// Pass 2: Apply fixups
	for _, fixup := range ctx.Fixups {
		addr, ok := ctx.Symbols.GetAddress(fixup.Label)
		if !ok {
			return nil, fmt.Errorf("undefined label: %s", fixup.Label)
		}
		// Apply fixup
		// We assume the word at Index has 0s where the address goes
		ctx.Words[fixup.Index] |= (uint16(addr) & fixup.Mask)
	}

	return ctx.Words, nil
}

// WriteHex writes the machine code in Intel HEX format.
func WriteHex(w io.Writer, words []uint16) error {
	// Intel HEX format
	// :LLAAAATTDD...DDCC
	// We write 16 bytes (8 words) per line usually.

	addr := 0
	const wordsPerLine = 8

	for i := 0; i < len(words); i += wordsPerLine {
		chunk := words[i:]
		if len(chunk) > wordsPerLine {
			chunk = chunk[:wordsPerLine]
		}

		byteCount := len(chunk) * 2
		recordType := 0 // Data

		// Start code
		fmt.Fprintf(w, ":%02X%04X%02X", byteCount, addr, recordType)

		checksum := byteCount + (addr >> 8) + (addr & 0xFF) + recordType

		for _, word := range chunk {
			// Little endian
			low := word & 0xFF
			high := (word >> 8) & 0xFF
			fmt.Fprintf(w, "%02X%02X", low, high)
			checksum += int(low) + int(high)
		}

		checksum = (^checksum + 1) & 0xFF
		fmt.Fprintf(w, "%02X\n", checksum)

		addr += len(chunk) * 2 // Address is in bytes
	}

	// EOF record
	fmt.Fprintf(w, ":00000001FF\n")
	return nil
}
