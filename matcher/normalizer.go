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

// syntaxMap maps Oracle-specific keywords/functions to PostgreSQL equivalents.
var syntaxMap = map[string]string{
	"sysdate":      "now()",
	"nvl":          "coalesce",
	"decode":       "case",
	"rownum":       "limit",
	"systimestamp": "now()",
}

// NormalizeSQL applies all normalization steps to a SQL string.
func NormalizeSQL(sql string) string {
	s := removeComments(sql)
	s = normalizeWhitespace(s)
	s = normalizeBindVariables(s)
	s = normalizeSyntax(s)
	return s
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

func normalizeSyntax(sql string) string {
	s := sql
	for oracle, pg := range syntaxMap {
		s = replaceWord(s, oracle, pg)
	}
	return s
}

// replaceWord replaces whole-word occurrences (case-insensitive already applied).
func replaceWord(s, old, replacement string) string {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
	return re.ReplaceAllString(s, replacement)
}
