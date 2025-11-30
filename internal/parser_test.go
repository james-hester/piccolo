package internal

import (
	"testing"
)

func TestParseReadmeExample(t *testing.T) {
	toks, err := Lex(readmeExample)
	if err != nil {
		t.Fatalf("Tokens: %v", err)
	}

	prog, err := Parse(toks)
	if err != nil {
		t.Fatalf("ParseProgram: %v", err)
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Functions))
	}

	fn := prog.Functions[0]
	if fn.Name != "function-name" {
		t.Errorf("expected fn name %q, got %q", "function-name", fn.Name)
	}

	if len(fn.Body) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(fn.Body))
	}

	// stmt 0: w = 5
	assign, ok := fn.Body[0].(AssignStmt)
	if !ok {
		t.Fatalf("expected first stmt AssignStmt, got %T", fn.Body[0])
	}
	lhsIdent, ok := assign.Lhs.(IdentExpr)
	if !ok {
		t.Fatalf("expected LHS IdentExpr, got %T", assign.Lhs)
	}
	if lhsIdent.Name != "w" {
		t.Errorf("expected assign to w, got %q", lhsIdent.Name)
	}
	numExpr, ok := assign.Expr.(NumExpr)
	if !ok {
		t.Fatalf("expected NumExpr, got %T", assign.Expr)
	}
	if numExpr.Val != "5" {
		t.Errorf("expected RHS '5', got %q", numExpr.Val)
	}
	if numExpr.Value != 5 {
		t.Errorf("expected RHS value 5, got %d", numExpr.Value)
	}

	// stmt 1: return
	if _, ok := fn.Body[1].(ReturnStmt); !ok {
		t.Fatalf("expected second stmt ReturnStmt, got %T", fn.Body[1])
	}
}

func TestParseNumericLiterals(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"$ff", 255},
		{"$FF", 255},
		{"$10", 16},
		{"%10", 2},
		{"%1111_1111", 255},
		{"123", 123},
		{"1_000", 1000},
	}

	for _, tc := range tests {
		toks, err := Lex("section program\nfn f() begin x = " + tc.input + " end")
		if err != nil {
			t.Errorf("Tokens(%q): %v", tc.input, err)
			continue
		}
		prog, err := Parse(toks)
		if err != nil {
			t.Errorf("Parse(%q): %v", tc.input, err)
			continue
		}
		assign := prog.Functions[0].Body[0].(AssignStmt)
		num := assign.Expr.(NumExpr)
		if num.Value != tc.want {
			t.Errorf("Parse(%q): got %d, want %d", tc.input, num.Value, tc.want)
		}
	}
}
