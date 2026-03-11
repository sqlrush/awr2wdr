package matcher

import (
	"sort"
	"strings"

	"awr2wdr/model"
)

// TokenSimilarity computes Jaccard similarity on whitespace-delimited tokens.
func TokenSimilarity(a, b string) float64 {
	tokensA := tokenize(a)
	tokensB := tokenize(b)

	if len(tokensA) == 0 && len(tokensB) == 0 {
		return 1.0
	}

	setA := toSet(tokensA)
	setB := toSet(tokensB)

	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 1.0
	}
	return float64(intersection) / float64(union)
}

func tokenize(s string) []string {
	parts := strings.Fields(s)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, "(),;")
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func toSet(tokens []string) map[string]bool {
	s := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		s[t] = true
	}
	return s
}

// MatchSQLs matches Oracle TopSQL to openGauss TopSQL using normalized SQL similarity.
// threshold is the minimum similarity score (0.0-1.0) to consider a match.
// Returns matched pairs sorted by Oracle execution count (descending).
func MatchSQLs(oracleSQLs, gaussSQLs []model.TopSQL, threshold float64) []model.SQLPair {
	oracleNorm := make([]string, len(oracleSQLs))
	for i, s := range oracleSQLs {
		oracleNorm[i] = NormalizeSQL(s.SQLText)
	}
	gaussNorm := make([]string, len(gaussSQLs))
	for i, s := range gaussSQLs {
		gaussNorm[i] = NormalizeSQL(s.SQLText)
	}

	used := make(map[int]bool)
	var pairs []model.SQLPair

	for i, oSQL := range oracleSQLs {
		bestIdx := -1
		bestScore := 0.0

		for j := range gaussSQLs {
			if used[j] {
				continue
			}
			score := TokenSimilarity(oracleNorm[i], gaussNorm[j])
			if score > bestScore && score >= threshold {
				bestScore = score
				bestIdx = j
			}
		}

		if bestIdx >= 0 {
			used[bestIdx] = true
			pairs = append(pairs, model.SQLPair{
				OracleSQL:  oSQL,
				GaussSQL:   gaussSQLs[bestIdx],
				Similarity: bestScore,
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		maxI := pairs[i].OracleSQL.Executions
		if pairs[i].GaussSQL.Executions > maxI {
			maxI = pairs[i].GaussSQL.Executions
		}
		maxJ := pairs[j].OracleSQL.Executions
		if pairs[j].GaussSQL.Executions > maxJ {
			maxJ = pairs[j].GaussSQL.Executions
		}
		return maxI > maxJ
	})

	for i := range pairs {
		pairs[i].Rank = i + 1
	}

	return pairs
}
