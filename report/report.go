package report

import (
	"fmt"
	"html/template"
	"io"
	"math"

	"awr2wdr/model"
)

// Generate renders the comparison report to the writer.
func Generate(data model.ComparisonReport, w io.Writer) error {
	funcMap := template.FuncMap{
		"formatInt":     formatInt,
		"formatFloat":   formatFloat,
		"diffPct":       diffPct,
		"diffPctInt":    diffPctInt,
		"diffClass":     diffClass,
		"formatDiffPct": formatDiffPct,
		"formatPct":     formatPct,
		"abs":           math.Abs,
		"truncateSQL":   truncateSQL,
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(w, data)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}

	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	out := string(result)
	if neg {
		return "-" + out
	}
	return out
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func diffPct(oracle, gauss float64) float64 {
	if oracle == 0 {
		return 0
	}
	return ((gauss - oracle) / oracle) * 100
}

func diffPctInt(oracle, gauss int64) float64 {
	if oracle == 0 {
		return 0
	}
	return (float64(gauss-oracle) / float64(oracle)) * 100
}

func diffClass(pct float64) string {
	if pct < -5 {
		return "better"
	}
	if pct > 5 {
		return "worse"
	}
	return "neutral"
}

func formatDiffPct(pct float64) string {
	if pct > 0 {
		return fmt.Sprintf("+%.1f%%", pct)
	}
	return fmt.Sprintf("%.1f%%", pct)
}

func formatPct(f float64) string {
	return fmt.Sprintf("%.0f%%", f*100)
}

func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}
