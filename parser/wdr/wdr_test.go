package wdr

import (
	"math"
	"os"
	"testing"
)

func TestParseWDR(t *testing.T) {
	f, err := os.Open("../../testdata/sample_wdr.html")
	if err != nil {
		t.Fatalf("open test data: %v", err)
	}
	defer f.Close()

	p := NewParser()
	data, err := p.Parse(f)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Source
	if data.Source != "opengauss" {
		t.Errorf("source = %q, want opengauss", data.Source)
	}

	// Instance info
	if data.Instance.DBName != "proddb" {
		t.Errorf("DBName = %q, want proddb", data.Instance.DBName)
	}
	if data.Instance.InstanceName != "gaussdb_prod" {
		t.Errorf("InstanceName = %q, want gaussdb_prod", data.Instance.InstanceName)
	}
	if data.Instance.Version != "5.0.1" {
		t.Errorf("Version = %q, want 5.0.1", data.Instance.Version)
	}
	if data.Instance.HostName != "gs-prod-01" {
		t.Errorf("HostName = %q, want gs-prod-01", data.Instance.HostName)
	}
	if data.Instance.DBTime != 987.56 {
		t.Errorf("DBTime = %f, want 987.56", data.Instance.DBTime)
	}
	if data.Instance.ElapsedTime != 60.0 {
		t.Errorf("ElapsedTime = %f, want 60.0", data.Instance.ElapsedTime)
	}
	if data.Instance.DBCPU != 412.33 {
		t.Errorf("DBCPU = %f, want 412.33", data.Instance.DBCPU)
	}

	// Wait events
	if len(data.WaitEvents) != 5 {
		t.Fatalf("WaitEvents count = %d, want 5", len(data.WaitEvents))
	}
	if data.WaitEvents[0].EventName != "DataFileRead" {
		t.Errorf("first event = %q, want DataFileRead", data.WaitEvents[0].EventName)
	}
	if data.WaitEvents[0].Waits != 1876543 {
		t.Errorf("first event waits = %d, want 1876543", data.WaitEvents[0].Waits)
	}
	// 5123800000 us = 5123.8 s
	if math.Abs(data.WaitEvents[0].TotalTime-5123.8) > 0.01 {
		t.Errorf("first event total time = %f, want 5123.8", data.WaitEvents[0].TotalTime)
	}

	// Top SQL
	if len(data.TopSQLs) != 3 {
		t.Fatalf("TopSQLs count = %d, want 3", len(data.TopSQLs))
	}
	sql1 := data.TopSQLs[0]
	if sql1.SQLID != "78923456" {
		t.Errorf("first SQL ID = %q, want 78923456", sql1.SQLID)
	}
	if sql1.Executions != 856012 {
		t.Errorf("first SQL executions = %d, want 856012", sql1.Executions)
	}
	// avg logical read = 169490376 / 856012 = 198
	if sql1.AvgLogicalRead != 198 {
		t.Errorf("first SQL avg logical read = %d, want 198", sql1.AvgLogicalRead)
	}
}
