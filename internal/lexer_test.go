package internal

import (
	"testing"
)

const readmeExample string = `section program
fn function-name() begin // function definition, comments
  w = 5 // MOVLW, decimal literal
  return // RETURN
end`

var wantTokens []string = []string{
	"SECTION", "PROGRAM",
	"FN", "IDENT[function-name]", "LPAREN", "RPAREN", "BEGIN",
	"IDENT[w]", "EQL", "NUMDECIMAL[5]",
	"RETURN",
	"END",
	"EOF",
}

func TestLex(t *testing.T) {
	toks, err := Lex(readmeExample)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	for i, tk := range toks {
		if tk.String() != wantTokens[i] {
			t.Errorf("pos %v, want %v, got %v", i, wantTokens[i], tk.String())
		}
	}
	if t.Failed() {
		for _, tk := range toks {
			t.Logf("%q", tk.String())
		}
	}
}

func TestIdentifierScanning(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			"a-b-",
			[]string{"IDENT[a-b]", "MINUS"},
		},
		{
			"a----b--",
			[]string{"IDENT[a----b]", "DEC"},
		},
		{
			"-a-b---",
			[]string{"MINUS", "IDENT[a-b]", "DEC", "MINUS"},
		},
		{
			"--a-b",
			[]string{"DEC", "IDENT[a-b]"},
		},
		{
			"f--",
			[]string{"IDENT[f]", "DEC"},
		},
		{
			"my-var--",
			[]string{"IDENT[my-var]", "DEC"},
		},
		{
			"a--b",
			[]string{"IDENT[a--b]"},
		},
		{
			"---",
			[]string{"DEC", "MINUS"},
		},
		{
			"----",
			[]string{"DEC", "DEC"},
		},
	}

	for _, tc := range tests {
		toks, err := Lex(tc.input)
		if err != nil {
			t.Errorf("Tokens(%q) error: %v", tc.input, err)
			continue
		}

		// Filter out EOF
		var got []string
		for _, tk := range toks {
			if tk.ty != EOF {
				got = append(got, tk.String())
			}
		}

		if len(got) != len(tc.want) {
			t.Errorf("Tokens(%q) length mismatch: want %v, got %v", tc.input, tc.want, got)
			continue
		}

		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("Tokens(%q)[%d]: want %q, got %q", tc.input, i, tc.want[i], got[i])
			}
		}
	}
}
