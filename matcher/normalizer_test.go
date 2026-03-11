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
			input: "select sysdate",
			want:  "select now()",
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
		{
			name:  "substr to substring",
			input: "select substr(name, 1, 3) from t",
			want:  "select substring(name, 1, 3) from t",
		},
		{
			name:  "instr to strpos",
			input: "select instr(name, 'a') from t",
			want:  "select strpos(name, 'a') from t",
		},
		{
			name:  "listagg to string_agg",
			input: "select listagg(name, ',') from t",
			want:  "select string_agg(name, ',') from t",
		},
		{
			name:  "wm_concat to string_agg",
			input: "select wm_concat(name) from t",
			want:  "select string_agg(name) from t",
		},
		{
			name:  "minus to except",
			input: "select a from t1 minus select a from t2",
			want:  "select a from t1 except select a from t2",
		},
		{
			name:  "to_number to cast",
			input: "select to_number(val) from t",
			want:  "select cast(val) from t",
		},
		{
			name:  "log to ln",
			input: "select log(10, x) from t",
			want:  "select ln(10, x) from t",
		},
		{
			name:  "sys_guid to gen_random_uuid",
			input: "select sys_guid() from t",
			want:  "select gen_random_uuid() from t",
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

func TestNormalizePatterns(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "remove FROM DUAL",
			input: "select sysdate from dual",
			want:  "select sysdate",
		},
		{
			name:  "remove FROM SYS.DUAL",
			input: "select 1 from sys.dual",
			want:  "select 1",
		},
		{
			name:  "remove outer join (+)",
			input: "select * from a, b where a.id = b.id(+)",
			want:  "select * from a, b where a.id = b.id",
		},
		{
			name:  "rownum pagination",
			input: "select * from t where rownum <= 10",
			want:  "select * from t limit 10",
		},
		{
			name:  "rownum with AND",
			input: "select * from t where status = 1 and rownum <= 20",
			want:  "select * from t where status = 1 limit 20",
		},
		{
			name:  "Oracle 12c FETCH FIRST",
			input: "select * from t order by id fetch first 10 rows only",
			want:  "select * from t order by id limit 10",
		},
		{
			name:  "OFFSET ROWS",
			input: "select * from t offset 20 rows fetch next 10 rows only",
			want:  "select * from t offset 20 limit 10",
		},
		{
			name:  "sequence nextval",
			input: "insert into t values (seq_order.nextval, 'a')",
			want:  "insert into t values (nextval('seq_order'), 'a')",
		},
		{
			name:  "sequence currval",
			input: "select seq_order.currval from t",
			want:  "select currval('seq_order') from t",
		},
		{
			name:  "CONNECT BY to recursive",
			input: "select * from t start with id = 1 connect by prior parent_id = id",
			want:  "select * from t recursive id = 1 recursive parent_id = id",
		},
		{
			name:  "BULK COLLECT INTO",
			input: "select name bulk collect into v_names from t",
			want:  "select name into v_names from t",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePatterns(tt.input)
			// Trim extra spaces from replacements
			got = multiSpaceRe.ReplaceAllString(got, " ")
			got = trimStr(got)
			want := multiSpaceRe.ReplaceAllString(tt.want, " ")
			want = trimStr(want)
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func trimStr(s string) string {
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

func TestNormalizeSQL(t *testing.T) {
	tests := []struct {
		name   string
		oracle string
		gauss  string
	}{
		{
			name:   "basic nvl and bind variables",
			oracle: "SELECT  NVL(a, 0)\nFROM t WHERE id = :bind1",
			gauss:  "select coalesce(a, 0) from t where id = $1",
		},
		{
			name:   "pagination rownum vs limit",
			oracle: "SELECT * FROM orders WHERE status = :b1 AND ROWNUM <= 10",
			gauss:  "SELECT * FROM orders WHERE status = $1 LIMIT 10",
		},
		{
			name:   "sysdate from dual vs now()",
			oracle: "SELECT SYSDATE FROM DUAL",
			gauss:  "SELECT now()",
		},
		{
			name:   "substr vs substring",
			oracle: "SELECT SUBSTR(name, 1, 3) FROM users WHERE id = :b1",
			gauss:  "select substring(name, 1, 3) from users where id = $1",
		},
		{
			name:   "listagg vs string_agg",
			oracle: "SELECT dept, LISTAGG(name, ',') FROM emp GROUP BY dept",
			gauss:  "SELECT dept, string_agg(name, ',') FROM emp GROUP BY dept",
		},
		{
			name:   "minus vs except",
			oracle: "SELECT id FROM t1 MINUS SELECT id FROM t2",
			gauss:  "SELECT id FROM t1 EXCEPT SELECT id FROM t2",
		},
		{
			name:   "Oracle 12c fetch first vs limit",
			oracle: "SELECT * FROM t ORDER BY id FETCH FIRST 20 ROWS ONLY",
			gauss:  "SELECT * FROM t ORDER BY id LIMIT 20",
		},
		{
			name:   "sequence nextval syntax",
			oracle: "INSERT INTO orders VALUES (seq_order.NEXTVAL, :b1)",
			gauss:  "INSERT INTO orders VALUES (nextval('seq_order'), $1)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oNorm := NormalizeSQL(tt.oracle)
			gNorm := NormalizeSQL(tt.gauss)
			if oNorm != gNorm {
				t.Errorf("normalized SQL should match:\noracle: %q\ngauss:  %q", oNorm, gNorm)
			}
		})
	}
}
