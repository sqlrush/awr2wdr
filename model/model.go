package model

// InstanceInfo holds database instance summary.
type InstanceInfo struct {
	DBName       string
	InstanceName string
	Version      string
	HostName     string
	SnapStart    string
	SnapEnd      string
	DBTime       float64 // minutes
	ElapsedTime  float64 // minutes
	DBCPU        float64 // minutes
	QPS          float64 // queries (SQL executions) per second
	TPS          float64 // transactions per second
	AvgDBTime    float64 // DB Time per second (seconds)
	AvgCPUTime   float64 // DB CPU per second (seconds)
}

// WaitEvent holds a single wait event row.
type WaitEvent struct {
	EventName string
	WaitClass string
	Waits     int64
	TotalTime float64 // seconds
	PctDBTime float64 // percentage
}

// TopSQL holds a single SQL entry.
type TopSQL struct {
	SQLID           string
	SQLText         string
	Executions      int64
	AvgElapsed      float64 // milliseconds
	AvgCPUTime      float64 // milliseconds
	AvgLogicalRead  int64
	AvgPhysicalRead int64
	AvgRows         int64
}

// ReportData holds all parsed data from a single report.
type ReportData struct {
	Source     string // "oracle" or "opengauss"
	Instance  InstanceInfo
	WaitEvents []WaitEvent
	TopSQLs    []TopSQL
}

// SQLPair holds a matched pair of SQL entries for comparison.
type SQLPair struct {
	Rank       int
	OracleSQL  TopSQL
	GaussSQL   TopSQL
	Similarity float64
}

// WaitEventPair holds a matched pair of wait events.
type WaitEventPair struct {
	Rank        int
	OracleEvent WaitEvent
	GaussEvent  WaitEvent
	TimeDiffPct float64 // positive = gauss worse, negative = gauss better
}

// ComparisonReport holds the full comparison data for rendering.
type ComparisonReport struct {
	GeneratedAt string
	Oracle      ReportData
	Gauss       ReportData
	WaitPairs   []WaitEventPair
	SQLPairs    []SQLPair
}
