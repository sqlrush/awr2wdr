package wdr

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"awr2wdr/model"

	"github.com/PuerkitoBio/goquery"
)

// Parser parses openGauss WDR HTML reports.
type Parser struct{}

// NewParser creates a new WDR parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a WDR HTML report and extracts performance data.
func (p *Parser) Parse(r io.Reader) (model.ReportData, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return model.ReportData{}, fmt.Errorf("parse HTML: %w", err)
	}

	data := model.ReportData{Source: "opengauss"}

	data.Instance, err = parseInstanceInfo(doc)
	if err != nil {
		return data, fmt.Errorf("parse instance info: %w", err)
	}

	data.WaitEvents, err = parseWaitEvents(doc)
	if err != nil {
		return data, fmt.Errorf("parse wait events: %w", err)
	}

	data.TopSQLs, err = parseTopSQLs(doc)
	if err != nil {
		return data, fmt.Errorf("parse top SQLs: %w", err)
	}

	return data, nil
}

func parseInstanceInfo(doc *goquery.Document) (model.InstanceInfo, error) {
	var info model.InstanceInfo

	// Find the table after "Report Info" heading
	table := findTableAfterHeading(doc, "Report Info")
	if table == nil {
		// Fallback: try "Summary" heading
		table = findTableAfterHeading(doc, "Summary")
	}
	if table == nil {
		return info, nil
	}

	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}
		key := strings.TrimSpace(cells.Eq(0).Text())
		val := strings.TrimSpace(cells.Eq(1).Text())

		switch {
		case strings.Contains(key, "DB Name"):
			info.DBName = val
		case strings.Contains(key, "Instance Name"):
			info.InstanceName = val
		case strings.Contains(key, "Host Name"):
			info.HostName = val
		case strings.Contains(key, "Version"):
			info.Version = strings.TrimPrefix(val, "openGauss ")
		case strings.Contains(key, "Snap Start"):
			info.SnapStart = val
		case strings.Contains(key, "Snap End"):
			info.SnapEnd = val
		case strings.Contains(key, "Elapsed Time"):
			info.ElapsedTime = parseNumber(val)
		case strings.Contains(key, "DB Time"):
			info.DBTime = parseNumber(val)
		case strings.Contains(key, "DB CPU"):
			info.DBCPU = parseNumber(val)
		}
	})

	return info, nil
}

func parseWaitEvents(doc *goquery.Document) ([]model.WaitEvent, error) {
	var events []model.WaitEvent

	table := findTableAfterHeading(doc, "Top Wait Events")
	if table == nil {
		table = findTableAfterHeading(doc, "Wait Events")
	}
	if table == nil {
		return events, nil
	}

	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 5 {
			return
		}

		// WDR wait time is in microseconds, convert to seconds
		totalTimeUs := parseNumber(cells.Eq(3).Text())

		event := model.WaitEvent{
			EventName: strings.TrimSpace(cells.Eq(0).Text()),
			WaitClass: strings.TrimSpace(cells.Eq(1).Text()),
			Waits:     parseInt64(cells.Eq(2).Text()),
			TotalTime: totalTimeUs / 1_000_000, // us to seconds
			PctDBTime: parseNumber(cells.Eq(4).Text()),
		}

		if event.EventName != "" {
			events = append(events, event)
		}
	})

	return events, nil
}

func parseTopSQLs(doc *goquery.Document) ([]model.TopSQL, error) {
	var sqls []model.TopSQL

	table := findTableAfterHeading(doc, "SQL ordered by Executions")
	if table == nil {
		table = findTableAfterHeading(doc, "SQL Statistics")
	}
	if table == nil {
		return sqls, nil
	}

	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 8 {
			return
		}

		executions := parseInt64(cells.Eq(2).Text())
		if executions == 0 {
			return
		}

		// WDR times are total in microseconds
		totalElapsedUs := parseNumber(cells.Eq(3).Text())
		totalCPUUs := parseNumber(cells.Eq(4).Text())
		totalPhysRead := parseInt64(cells.Eq(5).Text())
		totalLogicalRead := parseInt64(cells.Eq(6).Text())
		totalRows := parseInt64(cells.Eq(7).Text())

		sql := model.TopSQL{
			SQLID:           strings.TrimSpace(cells.Eq(0).Text()),
			SQLText:         strings.TrimSpace(cells.Eq(1).Text()),
			Executions:      executions,
			AvgElapsed:      (totalElapsedUs / float64(executions)) / 1000, // us to ms
			AvgCPUTime:      (totalCPUUs / float64(executions)) / 1000,    // us to ms
			AvgPhysicalRead: totalPhysRead / executions,
			AvgLogicalRead:  totalLogicalRead / executions,
			AvgRows:         totalRows / executions,
		}

		sqls = append(sqls, sql)
	})

	return sqls, nil
}

// findTableAfterHeading finds the first <table> that follows a heading containing the given text.
func findTableAfterHeading(doc *goquery.Document, heading string) *goquery.Selection {
	lower := strings.ToLower(heading)
	var found *goquery.Selection

	doc.Find("h1, h2, h3, h4").Each(func(i int, s *goquery.Selection) {
		if found != nil {
			return
		}
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		if strings.Contains(text, lower) {
			// Find the next table sibling
			next := s.NextAll()
			next.Each(func(j int, el *goquery.Selection) {
				if found != nil {
					return
				}
				if goquery.NodeName(el) == "table" {
					found = el
				}
			})
		}
	})

	return found
}

func parseNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f)
}
