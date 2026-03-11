package matcher

import (
	"testing"

	"awr2wdr/model"
)

func TestTokenSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		min  float64
		max  float64
	}{
		{
			name: "identical",
			a:    "select a from t where id = ?",
			b:    "select a from t where id = ?",
			min:  1.0,
			max:  1.0,
		},
		{
			name: "minor difference",
			a:    "select a, b from t where id = ?",
			b:    "select a, b from t where id = ? and status = ?",
			min:  0.6,
			max:  1.0,
		},
		{
			name: "completely different",
			a:    "select a from t1",
			b:    "update t2 set x = ?",
			min:  0.0,
			max:  0.3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := TokenSimilarity(tt.a, tt.b)
			if score < tt.min {
				t.Errorf("similarity %f < minimum %f", score, tt.min)
			}
			if score > tt.max {
				t.Errorf("similarity %f > maximum %f", score, tt.max)
			}
		})
	}
}

func TestMatchSQLs(t *testing.T) {
	oracleSQL := []model.TopSQL{
		{SQLID: "aaa", SQLText: "SELECT NVL(a, 0) FROM t WHERE id = :b1", Executions: 1000},
		{SQLID: "bbb", SQLText: "UPDATE t SET x = :b1 WHERE id = :b2", Executions: 500},
	}
	gaussSQL := []model.TopSQL{
		{SQLID: "111", SQLText: "select coalesce(a, 0) from t where id = $1", Executions: 1000},
		{SQLID: "222", SQLText: "update t set x = $1 where id = $2", Executions: 500},
	}

	pairs := MatchSQLs(oracleSQL, gaussSQL, 0.80)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].OracleSQL.SQLID != "aaa" || pairs[0].GaussSQL.SQLID != "111" {
		t.Errorf("first pair mismatch: oracle=%s gauss=%s", pairs[0].OracleSQL.SQLID, pairs[0].GaussSQL.SQLID)
	}
	if pairs[1].OracleSQL.SQLID != "bbb" || pairs[1].GaussSQL.SQLID != "222" {
		t.Errorf("second pair mismatch: oracle=%s gauss=%s", pairs[1].OracleSQL.SQLID, pairs[1].GaussSQL.SQLID)
	}
}

func TestMatchSQLs_NoMatch(t *testing.T) {
	oracleSQL := []model.TopSQL{
		{SQLID: "aaa", SQLText: "SELECT a FROM t1", Executions: 1000},
	}
	gaussSQL := []model.TopSQL{
		{SQLID: "111", SQLText: "DELETE FROM t2 WHERE x = $1", Executions: 500},
	}

	pairs := MatchSQLs(oracleSQL, gaussSQL, 0.80)
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs for unmatched SQL, got %d", len(pairs))
	}
}
