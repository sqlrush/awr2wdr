package parser

import (
	"io"

	"awr2wdr/model"
)

// Parser defines the interface for parsing performance reports.
type Parser interface {
	Parse(r io.Reader) (model.ReportData, error)
}
