package matcher

import (
	"regexp"
	"strings"
)

var (
	multiSpaceRe   = regexp.MustCompile(`\s+`)
	oracleBindRe   = regexp.MustCompile(`:[a-zA-Z_]\w*`)
	pgBindRe       = regexp.MustCompile(`\$\d+`)
	commentLineRe  = regexp.MustCompile(`--[^\n]*`)
	commentBlockRe = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

// syntaxMap maps Oracle-specific keywords/functions to PostgreSQL/openGauss equivalents.
// Grouped by category for clarity.
var syntaxMap = map[string]string{
	// ===== Date/Time Functions =====
	"sysdate":                "now()",
	"systimestamp":           "now()",
	"current_timestamp":      "now()",
	"add_months":             "date_add",
	"months_between":         "age",
	"last_day":               "date_trunc",
	"next_day":               "date_trunc",
	"to_date":                "to_timestamp",
	"to_dsinterval":          "interval",
	"to_yminterval":          "interval",
	"numtodsinterval":        "interval",
	"numtoyminterval":        "interval",
	"extract":                "extract",       // same but for completeness
	"interval":               "interval",      // same

	// ===== NULL Handling =====
	"nvl":                    "coalesce",
	"nvl2":                   "coalesce",      // simplified mapping
	"decode":                 "case",
	"nullif":                 "nullif",        // same

	// ===== String Functions =====
	"substr":                 "substring",
	"instr":                  "strpos",
	"lengthb":                "octet_length",
	"substrb":                "substring",
	"concat":                 "concat",        // same
	"chr":                    "chr",           // same
	"ascii":                  "ascii",         // same
	"initcap":                "initcap",       // same
	"lpad":                   "lpad",          // same
	"rpad":                   "rpad",          // same
	"ltrim":                  "ltrim",         // same
	"rtrim":                  "rtrim",         // same
	"replace":                "replace",       // same
	"translate":              "translate",     // same
	"regexp_substr":          "regexp_matches",
	"regexp_replace":         "regexp_replace", // same
	"regexp_instr":           "regexp_matches",
	"regexp_like":            "similar to",
	"regexp_count":           "regexp_matches",

	// ===== Type Conversion =====
	"to_number":              "cast",
	"to_char":                "to_char",       // same
	"to_clob":                "cast",
	"to_nchar":               "cast",
	"to_nclob":               "cast",
	"hextoraw":               "decode",
	"rawtohex":               "encode",
	"rowidtochar":            "cast",
	"cast":                   "cast",          // same

	// ===== Numeric/Math Functions =====
	"trunc":                  "trunc",         // same, works for both date and number
	"mod":                    "mod",           // same
	"ceil":                   "ceil",          // same
	"floor":                  "floor",         // same
	"round":                  "round",         // same
	"abs":                    "abs",           // same
	"sign":                   "sign",          // same
	"power":                  "power",         // same
	"sqrt":                   "sqrt",          // same
	"log":                    "ln",
	"remainder":              "mod",

	// ===== Aggregate/Analytic Functions =====
	"listagg":                "string_agg",
	"wm_concat":              "string_agg",
	"median":                 "percentile_cont",
	"ratio_to_report":        "sum",           // simplified
	"row_number":             "row_number",    // same
	"rank":                   "rank",          // same
	"dense_rank":             "dense_rank",    // same
	"ntile":                  "ntile",         // same
	"lag":                    "lag",           // same
	"lead":                   "lead",          // same
	"first_value":            "first_value",   // same
	"last_value":             "last_value",    // same

	// ===== Pagination/Row Limiting =====
	"rownum":                 "row_number()",
	"rowid":                  "ctid",

	// ===== Set Operators =====
	"minus":                  "except",

	// ===== Sequence =====
	"currval":                "currval",       // same syntax in openGauss
	"nextval":                "nextval",       // same syntax in openGauss

	// ===== Other Oracle-Specific =====
	"userenv":                "current_setting",
	"sys_context":            "current_setting",
	"sys_guid":               "gen_random_uuid",
	"dbms_random.value":      "random",
	"uid":                    "current_user",
	"user":                   "current_user",
}

// patternReplacements handles structural patterns that need regex-based normalization.
var patternReplacements = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	// "FROM dual" / "FROM sys.dual" → remove (openGauss doesn't need DUAL)
	{regexp.MustCompile(`\bfrom\s+(?:sys\.)?dual\b`), ""},

	// Oracle outer join "(+)" → remove
	{regexp.MustCompile(`\(\+\)`), ""},

	// "WHERE ROWNUM <= N" (sole condition) → "LIMIT N"
	{regexp.MustCompile(`\bwhere\s+rownum\s*<=?\s*(\d+)`), "limit $1"},

	// "AND ROWNUM <= N" → "LIMIT N"
	{regexp.MustCompile(`\band\s+rownum\s*<=?\s*(\d+)`), "limit $1"},

	// Remaining "ROWNUM <= N" → "limit N"
	{regexp.MustCompile(`\brownum\s*<=?\s*(\d+)`), "limit $1"},

	// Oracle sequence "seq_name.nextval" → "nextval('seq_name')"
	{regexp.MustCompile(`(\w+)\.nextval\b`), "nextval('$1')"},
	{regexp.MustCompile(`(\w+)\.currval\b`), "currval('$1')"},

	// "CONNECT BY" hierarchical → "WITH RECURSIVE" (just normalize the keyword)
	{regexp.MustCompile(`\bconnect\s+by\b`), "recursive"},
	{regexp.MustCompile(`\bstart\s+with\b`), "recursive"},
	{regexp.MustCompile(`\bprior\s+`), ""},

	// Oracle "IS TABLE OF" / "BULK COLLECT INTO" → remove PL/SQL noise
	{regexp.MustCompile(`\bbulk\s+collect\s+into\b`), "into"},
	{regexp.MustCompile(`\bforall\b`), ""},

	// "FETCH FIRST N ROWS ONLY" → "LIMIT N" (Oracle 12c+ syntax)
	{regexp.MustCompile(`\bfetch\s+first\s+(\d+)\s+rows?\s+only\b`), "limit $1"},
	{regexp.MustCompile(`\bfetch\s+next\s+(\d+)\s+rows?\s+only\b`), "limit $1"},

	// "OFFSET N ROWS" → "OFFSET N"
	{regexp.MustCompile(`\boffset\s+(\d+)\s+rows?\b`), "offset $1"},

	// Oracle string concatenation using "||" is same in PG, no change needed

	// Remove Oracle hints "/*+ ... */" (already handled by comment removal, but be explicit)
	{regexp.MustCompile(`/\*\+[^*]*\*/`), ""},
}

// NormalizeSQL applies all normalization steps to a SQL string.
func NormalizeSQL(sql string) string {
	s := removeComments(sql)
	s = normalizeWhitespace(s)
	s = normalizeBindVariables(s)
	s = normalizePatterns(s)
	s = normalizeSyntax(s)
	// Final cleanup: collapse any extra spaces introduced by replacements
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func removeComments(sql string) string {
	s := commentBlockRe.ReplaceAllString(sql, " ")
	s = commentLineRe.ReplaceAllString(s, " ")
	return s
}

func normalizeWhitespace(sql string) string {
	s := strings.ToLower(sql)
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func normalizeBindVariables(sql string) string {
	s := oracleBindRe.ReplaceAllString(sql, "?")
	s = pgBindRe.ReplaceAllString(s, "?")
	return s
}

func normalizePatterns(sql string) string {
	s := sql
	for _, pr := range patternReplacements {
		s = pr.pattern.ReplaceAllString(s, pr.replacement)
	}
	return s
}

func normalizeSyntax(sql string) string {
	s := sql
	for oracle, pg := range syntaxMap {
		if oracle == pg {
			continue // skip identical mappings
		}
		s = replaceWord(s, oracle, pg)
	}
	return s
}

// replaceWord replaces whole-word occurrences (case-insensitive already applied).
func replaceWord(s, old, replacement string) string {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
	return re.ReplaceAllString(s, replacement)
}
