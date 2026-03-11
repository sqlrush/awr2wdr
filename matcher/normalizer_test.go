package matcher

import "testing"

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "collapse spaces and newlines",
			input: "SELECT  a,\n  b\n  FROM   t",
			want:  "select a, b from t",
		},
		{
			name:  "trim leading and trailing",
			input: "  SELECT 1  ",
			want:  "select 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeBindVariables(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "oracle colon bind",
			input: "select * from t where id = :bind1 and name = :b2",
			want:  "select * from t where id = ? and name = ?",
		},
		{
			name:  "postgres dollar bind",
			input: "select * from t where id = $1 and name = $2",
			want:  "select * from t where id = ? and name = ?",
		},
		{
			name:  "mixed numeric literals preserved",
			input: "select * from t where id = 123",
			want:  "select * from t where id = 123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeBindVariables(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "sysdate to now",
			input: "select sysdate from dual",
			want:  "select now() from dual",
		},
		{
			name:  "nvl to coalesce",
			input: "select nvl(a, 0) from t",
			want:  "select coalesce(a, 0) from t",
		},
		{
			name:  "decode to case",
			input: "select decode(status, 1, 'a', 'b') from t",
			want:  "select case(status, 1, 'a', 'b') from t",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSyntax(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSQL(t *testing.T) {
	oracle := "SELECT  NVL(a, 0)\nFROM t WHERE id = :bind1"
	gauss := "select coalesce(a, 0) from t where id = $1"

	oNorm := NormalizeSQL(oracle)
	gNorm := NormalizeSQL(gauss)

	if oNorm != gNorm {
		t.Errorf("normalized SQL should match:\noracle: %q\ngauss:  %q", oNorm, gNorm)
	}
}
