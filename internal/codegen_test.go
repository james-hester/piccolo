package internal

import (
	"testing"
)

func TestCompile(t *testing.T) {
	input := `
section program
fn main() begin
  w += f
  f += w
  fsr0 += 5
  w &= 255
  w &= f
  f &= w
  f[0] = 0
  f[1] = 1
  if f[2] then
    w += f
  if not f[3] then
    f += w
  if (f--) != 0 then
    w &= 15
  if (f++) != 0 then
    w &= 240
  return
end
`
	tokens, err := Lex(input)
	if err != nil {
		t.Fatalf("Tokens failed: %v", err)
	}

	prog, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ops, _, err := Compile(prog)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	expected := []string{
		"main:",
		"ADDWF f,0",
		"ADDWF f,1",
		"ADDFSR 0,5",
		"ANDLW 255",
		"ANDWF f,0",
		"ANDWF f,1",
		"BCF f,0",
		"BSF f,1",
		"BTFSC f,2",
		"ADDWF f,0",
		"BTFSS f,3",
		"ADDWF f,1",
		"DECFSZ f,1",
		"ANDLW 15",
		"INCFSZ f,1",
		"ANDLW 240",
		"RETURN",
	}

	if len(ops) != len(expected) {
		t.Fatalf("Expected %d ops, got %d", len(expected), len(ops))
	}

	for i, op := range ops {
		if op.Assembly() != expected[i] {
			t.Errorf("Op %d: expected %q, got %q", i, expected[i], op.Assembly())
		}
	}
}
