package awr

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"awr2wdr/model"

	"github.com/PuerkitoBio/goquery"
)

// Parser parses Oracle AWR HTML reports.
type Parser struct{}

// NewParser creates a new AWR parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads an AWR HTML report and extracts performance data.
func (p *Parser) Parse(r io.Reader) (model.ReportData, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return model.ReportData{}, fmt.Errorf("parse HTML: %w", err)
	}

	data := model.ReportData{Source: "oracle"}

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

	// Database instance information table
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		if strings.Contains(strings.ToLower(summary), "database instance information") {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				if j == 0 {
					return // skip header
				}
				cells := row.Find("td")
				if cells.Length() >= 6 {
					info.DBName = strings.TrimSpace(cells.Eq(0).Text())
					info.InstanceName = strings.TrimSpace(cells.Eq(2).Text())
					info.Version = strings.TrimSpace(cells.Eq(5).Text())
				}
			})
		}
	})

	// Host information table
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		if strings.Contains(strings.ToLower(summary), "host information") {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				if j == 0 {
					return
				}
				cells := row.Find("td")
				if cells.Length() >= 1 {
					info.HostName = strings.TrimSpace(cells.Eq(0).Text())
				}
			})
		}
	})

	// Snapshot information table
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		if strings.Contains(strings.ToLower(summary), "snapshot information") {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() < 2 {
					return
				}
				label := strings.TrimSpace(cells.Eq(0).Text())
				switch {
				case strings.HasPrefix(label, "Begin Snap"):
					if cells.Length() >= 3 {
						info.SnapStart = strings.TrimSpace(cells.Eq(2).Text())
					}
				case strings.HasPrefix(label, "End Snap"):
					if cells.Length() >= 3 {
						info.SnapEnd = strings.TrimSpace(cells.Eq(2).Text())
					}
				case strings.HasPrefix(label, "Elapsed"):
					if cells.Length() >= 3 {
						info.ElapsedTime = parseMinutes(cells.Eq(2).Text())
					}
				case strings.HasPrefix(label, "DB Time"):
					if cells.Length() >= 3 {
						info.DBTime = parseMinutes(cells.Eq(2).Text())
					}
				}
			})
		}
	})

	// Load profile for DB CPU, QPS, TPS, AvgDBTime, AvgCPUTime
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		if strings.Contains(strings.ToLower(summary), "load profile") {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() < 2 {
					return
				}
				label := strings.TrimSpace(cells.Eq(0).Text())
				perSec := parseNumber(cells.Eq(1).Text())
				switch {
				case strings.Contains(label, "DB CPU"):
					info.DBCPU = perSec * info.ElapsedTime
					info.AvgCPUTime = perSec
				case strings.Contains(label, "DB Time"):
					info.AvgDBTime = perSec
				case strings.Contains(label, "Executes") || strings.Contains(label, "Execute"):
					info.QPS = perSec
				case strings.Contains(label, "User commits") || strings.Contains(label, "Transactions"):
					info.TPS = perSec
				}
			})
		}
	})

	return info, nil
}

func parseWaitEvents(doc *goquery.Document) ([]model.WaitEvent, error) {
	var events []model.WaitEvent

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		lower := strings.ToLower(summary)
		if !strings.Contains(lower, "top") || !strings.Contains(lower, "wait") {
			return
		}

		s.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 5 {
				return
			}

			event := model.WaitEvent{
				EventName: strings.TrimSpace(cells.Eq(0).Text()),
				Waits:     parseInt64(cells.Eq(1).Text()),
				TotalTime: parseNumber(cells.Eq(2).Text()),
				PctDBTime: parseNumber(cells.Eq(4).Text()),
			}

			if cells.Length() >= 6 {
				event.WaitClass = strings.TrimSpace(cells.Eq(5).Text())
			}

			if event.EventName != "" {
				events = append(events, event)
			}
		})
	})

	return events, nil
}

func parseTopSQLs(doc *goquery.Document) ([]model.TopSQL, error) {
	var sqls []model.TopSQL

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		summary, _ := s.Attr("summary")
		if !strings.Contains(summary, "SQL ordered by Executions") {
			return
		}

		s.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 14 {
				return
			}

			executions := parseInt64(cells.Eq(0).Text())
			if executions == 0 {
				return
			}

			sql := model.TopSQL{
				Executions:      executions,
				AvgRows:         parseInt64(cells.Eq(2).Text()),
				AvgElapsed:      parseNumber(cells.Eq(4).Text()) * 1000, // seconds to ms
				AvgCPUTime:      parseNumber(cells.Eq(6).Text()) * 1000, // seconds to ms
				AvgPhysicalRead: parseInt64(cells.Eq(8).Text()),
				AvgLogicalRead:  parseInt64(cells.Eq(10).Text()),
				SQLID:           strings.TrimSpace(cells.Eq(11).Text()),
				SQLText:         strings.TrimSpace(cells.Eq(13).Text()),
			}

			sqls = append(sqls, sql)
		})
	})

	return sqls, nil
}

// parseMinutes extracts minutes from strings like "1,245.32 (mins)"
func parseMinutes(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "(mins)", "", 1)
	s = strings.Replace(s, "(min)", "", 1)
	s = strings.TrimSpace(s)
	return parseNumber(s)
}

// parseNumber parses a number string, removing commas.
func parseNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseInt64 parses an integer string, removing commas.
func parseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	// Try float first for values like "1.00"
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f)
}
