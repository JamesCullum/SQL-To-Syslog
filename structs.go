package main

import "time"

type Config struct {
	Timezone string `json:"timezone"`
	Syslog   struct {
		Host string `json:"host"`
		Prot string `json:"prot"`
	} `json:"syslog"`
	Database string      `json:"database"`
	Jobs     []ConfigJob `json:"jobs"`
}

type ConfigJob struct {
	Name                      string `json:"name"`
	CronWithLeadingSecond     string `json:"cronWithLeadingSecond"`
	Query                     string `json:"query"`
	QueryCountColumn          string `json:"queryCountColumn"`
	InitialLastCheckOffsetMin int64  `json:"initialLastCheckOffsetMin"`
}

type FileCache struct {
	LastChecks map[string]time.Time `json:"lastChecks"`
}