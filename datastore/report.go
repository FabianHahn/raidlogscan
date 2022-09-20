package datastore

import (
	"time"
)

type ReportPlayer struct {
	Id     int64
	Name   string
	Class  string
	Server string
	Spec   string
	Role   string
}

type Report struct {
	Title     string
	CreatedAt time.Time
	StartTime time.Time
	EndTime   time.Time
	Zone      string
	Players   []ReportPlayer `datastore:",noindex"`
}
