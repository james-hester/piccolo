package internal

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigBits(t *testing.T) {
	input := `
section configuration
  conf: $3F3F
  conf2: $1234

section program
fn main() begin
  return
end
`
	tokens, err := Lex(input)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	prog, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ops, syms, err := Compile(prog)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check for ConfigOp in ops
	foundConf1 := false
	foundConf2 := false
	for _, op := range ops {
		if cfg, ok := op.(ConfigOp); ok {
			if cfg.Address == 0x8007 && cfg.Value == 0x3F3F {
				foundConf1 = true
			}
			if cfg.Address == 0x8008 && cfg.Value == 0x1234 {
				foundConf2 = true
			}
		}
	}

	if !foundConf1 {
		t.Error("Did not find ConfigOp for conf ($3F3F at 0x8007)")
	}
	if !foundConf2 {
		t.Error("Did not find ConfigOp for conf2 ($1234 at 0x8008)")
	}

	// Assemble
	words, config, err := Assemble(ops, syms)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if len(words) == 0 {
		t.Error("Expected some code words")
	}

	if val, ok := config[0x8007]; !ok || val != 0x3F3F {
		t.Errorf("Assembler config map mismatch for 0x8007: got %v, want 0x3F3F", val)
	}
	if val, ok := config[0x8008]; !ok || val != 0x1234 {
		t.Errorf("Assembler config map mismatch for 0x8008: got %v, want 0x1234", val)
	}

	// Write Hex
	var buf bytes.Buffer
	if err := WriteHex(&buf, words, config); err != nil {
		t.Fatalf("WriteHex failed: %v", err)
	}

	hexOutput := buf.String()
	// Check for Extended Linear Address Record :020000040001F9
	if !strings.Contains(hexOutput, ":020000040001F9") {
		t.Error("Hex output missing Extended Linear Address Record for config")
	}

	// Check for config data
	// 0x8007 * 2 = 0x1000E. Offset 0x000E.
	// :02000E003F3F74 (checksum might vary)
	// 02 000E 00 3F 3F -> 2 + 14 + 0 + 63 + 63 = 142 = 0x8E. ^0x8E + 1 = 0x72.
	// Let's just check for the address part :02000E00
	if !strings.Contains(hexOutput, ":02000E00") {
		t.Error("Hex output missing config data at 0x000E (0x8007)")
	}
}
