package report

import (
	"bytes"
	"strings"
	"testing"

	"awr2wdr/model"
)

func TestGenerate(t *testing.T) {
	data := model.ComparisonReport{
		GeneratedAt: "2026-03-11 20:00:00",
		Oracle: model.ReportData{
			Source: "oracle",
			Instance: model.InstanceInfo{
				DBName: "TESTDB", InstanceName: "testdb1",
				Version: "19.16.0", HostName: "ora-host",
				SnapStart: "2026-03-11 08:00", SnapEnd: "2026-03-11 09:00",
				DBTime: 1245.32, ElapsedTime: 60.0, DBCPU: 523.18,
			},
		},
		Gauss: model.ReportData{
			Source: "opengauss",
			Instance: model.InstanceInfo{
				DBName: "testdb", InstanceName: "gaussdb1",
				Version: "5.0.1", HostName: "gs-host",
				SnapStart: "2026-03-11 08:00", SnapEnd: "2026-03-11 09:00",
				DBTime: 987.56, ElapsedTime: 60.0, DBCPU: 412.33,
			},
		},
		WaitPairs: []model.WaitEventPair{
			{
				Rank: 1,
				OracleEvent: model.WaitEvent{
					EventName: "db file sequential read", WaitClass: "User I/O",
					Waits: 2345678, TotalTime: 8234.5, PctDBTime: 32.4,
				},
				GaussEvent: model.WaitEvent{
					EventName: "DataFileRead", WaitClass: "IO",
					Waits: 1876543, TotalTime: 5123.8, PctDBTime: 21.6,
				},
				TimeDiffPct: -37.7,
			},
		},
		SQLPairs: []model.SQLPair{
			{
				Rank: 1,
				OracleSQL: model.TopSQL{
					SQLID: "abc123", SQLText: "SELECT * FROM users WHERE id = :b1",
					Executions: 856234, AvgElapsed: 12.34, AvgCPUTime: 8.9,
					AvgLogicalRead: 245, AvgPhysicalRead: 18, AvgRows: 1,
				},
				GaussSQL: model.TopSQL{
					SQLID: "78923456", SQLText: "SELECT * FROM users WHERE id = $1",
					Executions: 856012, AvgElapsed: 8.76, AvgCPUTime: 6.2,
					AvgLogicalRead: 198, AvgPhysicalRead: 12, AvgRows: 1,
				},
				Similarity: 0.95,
			},
		},
	}

	var buf bytes.Buffer
	err := Generate(data, &buf)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	html := buf.String()

	checks := []struct {
		name    string
		contain string
	}{
		{"oracle db name", "TESTDB"},
		{"gauss db name", "testdb"},
		{"oracle wait event", "db file sequential read"},
		{"gauss wait event", "DataFileRead"},
		{"oracle sql id", "abc123"},
		{"gauss sql id", "78923456"},
		{"generated time", "2026-03-11 20:00:00"},
		{"html doctype", "<!DOCTYPE html>"},
		{"similarity badge", "95%"},
	}

	for _, c := range checks {
		if !strings.Contains(html, c.contain) {
			t.Errorf("output should contain %s (%q)", c.name, c.contain)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{123, "123"},
		{1234, "1,234"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		got := formatInt(tt.input)
		if got != tt.want {
			t.Errorf("formatInt(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDiffClass(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{-10.0, "better"},
		{10.0, "worse"},
		{0.0, "neutral"},
		{3.0, "neutral"},
		{-3.0, "neutral"},
	}
	for _, tt := range tests {
		got := diffClass(tt.pct)
		if got != tt.want {
			t.Errorf("diffClass(%f) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}
