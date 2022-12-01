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

type ReportPlayerAccount struct {
	Name     string
	PlayerId int64
}

type Report struct {
	Title          string
	CreatedAt      time.Time
	StartTime      time.Time
	EndTime        time.Time
	Zone           string
	GuildId        int32
	GuildName      string
	Players        []ReportPlayer        `datastore:",noindex"`
	PlayerAccounts []ReportPlayerAccount `datastore:",noindex"`
	Version        int32
}
