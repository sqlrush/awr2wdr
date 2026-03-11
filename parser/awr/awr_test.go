package awr

import (
	"os"
	"testing"
)

func TestParseAWR(t *testing.T) {
	f, err := os.Open("../../testdata/sample_awr.html")
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
	if data.Source != "oracle" {
		t.Errorf("source = %q, want oracle", data.Source)
	}

	// Instance info
	if data.Instance.DBName != "PRODDB" {
		t.Errorf("DBName = %q, want PRODDB", data.Instance.DBName)
	}
	if data.Instance.InstanceName != "proddb1" {
		t.Errorf("InstanceName = %q, want proddb1", data.Instance.InstanceName)
	}
	if data.Instance.Version != "19.16.0.0.0" {
		t.Errorf("Version = %q, want 19.16.0.0.0", data.Instance.Version)
	}
	if data.Instance.HostName != "ora-prod-01" {
		t.Errorf("HostName = %q, want ora-prod-01", data.Instance.HostName)
	}
	if data.Instance.DBTime != 1245.32 {
		t.Errorf("DBTime = %f, want 1245.32", data.Instance.DBTime)
	}
	if data.Instance.ElapsedTime != 60.0 {
		t.Errorf("ElapsedTime = %f, want 60.0", data.Instance.ElapsedTime)
	}

	// Wait events
	if len(data.WaitEvents) == 0 {
		t.Fatal("WaitEvents should not be empty")
	}
	if data.WaitEvents[0].EventName != "db file sequential read" {
		t.Errorf("first event = %q, want db file sequential read", data.WaitEvents[0].EventName)
	}
	if data.WaitEvents[0].Waits != 2345678 {
		t.Errorf("first event waits = %d, want 2345678", data.WaitEvents[0].Waits)
	}
	if data.WaitEvents[0].WaitClass != "User I/O" {
		t.Errorf("first event class = %q, want User I/O", data.WaitEvents[0].WaitClass)
	}

	// Top SQL
	if len(data.TopSQLs) == 0 {
		t.Fatal("TopSQLs should not be empty")
	}
	if len(data.TopSQLs) != 3 {
		t.Errorf("TopSQLs count = %d, want 3", len(data.TopSQLs))
	}
	sql1 := data.TopSQLs[0]
	if sql1.SQLID != "4r5s6t7u8v" {
		t.Errorf("first SQL ID = %q, want 4r5s6t7u8v", sql1.SQLID)
	}
	if sql1.Executions != 856234 {
		t.Errorf("first SQL executions = %d, want 856234", sql1.Executions)
	}
	if sql1.AvgLogicalRead != 245 {
		t.Errorf("first SQL avg logical read = %d, want 245", sql1.AvgLogicalRead)
	}
	if sql1.AvgPhysicalRead != 18 {
		t.Errorf("first SQL avg physical read = %d, want 18", sql1.AvgPhysicalRead)
	}
}
