package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"awr2wdr/matcher"
	"awr2wdr/model"
	"awr2wdr/parser"
	"awr2wdr/parser/awr"
	"awr2wdr/parser/wdr"
	"awr2wdr/report"
)

func main() {
	awrFile := flag.String("awr", "", "Path to Oracle AWR HTML report")
	wdrFile := flag.String("wdr", "", "Path to openGauss WDR HTML report")
	output := flag.String("o", "comparison_report.html", "Output HTML file path")
	threshold := flag.Float64("threshold", 0.80, "SQL similarity threshold (0.0-1.0)")
	topN := flag.Int("top", 10, "Number of top SQL to compare")
	flag.Parse()

	if *awrFile == "" || *wdrFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: awr2wdr -awr <awr.html> -wdr <wdr.html> [-o output.html] [-threshold 0.80] [-top 10]\n")
		os.Exit(1)
	}

	// Parse AWR
	awrData, err := parseFile(*awrFile, awr.NewParser())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing AWR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed AWR: %s (%s), %d wait events, %d SQLs\n",
		awrData.Instance.DBName, awrData.Instance.Version,
		len(awrData.WaitEvents), len(awrData.TopSQLs))

	// Parse WDR
	wdrData, err := parseFile(*wdrFile, wdr.NewParser())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WDR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed WDR: %s (%s), %d wait events, %d SQLs\n",
		wdrData.Instance.DBName, wdrData.Instance.Version,
		len(wdrData.WaitEvents), len(wdrData.TopSQLs))

	// Limit top SQL
	if len(awrData.TopSQLs) > *topN {
		awrData.TopSQLs = awrData.TopSQLs[:*topN]
	}
	if len(wdrData.TopSQLs) > *topN {
		wdrData.TopSQLs = wdrData.TopSQLs[:*topN]
	}

	// Match SQL
	sqlPairs := matcher.MatchSQLs(awrData.TopSQLs, wdrData.TopSQLs, *threshold)
	fmt.Printf("Matched %d SQL pairs (threshold: %.0f%%)\n", len(sqlPairs), *threshold*100)

	// Build wait event pairs (by rank position)
	waitPairs := buildWaitPairs(awrData.WaitEvents, wdrData.WaitEvents)

	// Build comparison report
	comp := model.ComparisonReport{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Oracle:      awrData,
		Gauss:       wdrData,
		WaitPairs:   waitPairs,
		SQLPairs:    sqlPairs,
	}

	// Generate output
	outFile, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	if err := report.Generate(comp, outFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Report generated: %s\n", *output)
}

func parseFile(path string, p parser.Parser) (model.ReportData, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.ReportData{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	// Read entire file to allow potential re-parsing
	var r io.Reader = f
	return p.Parse(r)
}

func buildWaitPairs(oracleEvents, gaussEvents []model.WaitEvent) []model.WaitEventPair {
	maxLen := len(oracleEvents)
	if len(gaussEvents) > maxLen {
		maxLen = len(gaussEvents)
	}

	pairs := make([]model.WaitEventPair, 0, maxLen)
	for i := 0; i < maxLen; i++ {
		pair := model.WaitEventPair{Rank: i + 1}
		if i < len(oracleEvents) {
			pair.OracleEvent = oracleEvents[i]
		}
		if i < len(gaussEvents) {
			pair.GaussEvent = gaussEvents[i]
		}
		if pair.OracleEvent.TotalTime > 0 {
			pair.TimeDiffPct = ((pair.GaussEvent.TotalTime - pair.OracleEvent.TotalTime) / pair.OracleEvent.TotalTime) * 100
		}
		pairs = append(pairs, pair)
	}
	return pairs
}
