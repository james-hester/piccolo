package internal

import (
	"testing"
)

func TestParseDataSection(t *testing.T) {
	input := `
section data
common:
  var1 i8
banked:
  var2 i8
`
	toks, err := Lex(input)
	if err != nil {
		t.Fatalf("Lex failed: %v", err)
	}

	prog, err := Parse(toks)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(prog.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(prog.Variables))
	}

	v1, ok := prog.Variables["var1"]
	if !ok {
		t.Errorf("var1 not found")
	} else {
		if v1.Name != "var1" {
			t.Errorf("var1 name mismatch: %s", v1.Name)
		}
		if v1.Type != "i8" {
			t.Errorf("var1 type mismatch: %s", v1.Type)
		}
		if v1.Banked {
			t.Errorf("var1 should be common (not banked)")
		}
	}

	v2, ok := prog.Variables["var2"]
	if !ok {
		t.Errorf("var2 not found")
	} else {
		if v2.Name != "var2" {
			t.Errorf("var2 name mismatch: %s", v2.Name)
		}
		if v2.Type != "i8" {
			t.Errorf("var2 type mismatch: %s", v2.Type)
		}
		if !v2.Banked {
			t.Errorf("var2 should be banked")
		}
	}
}
