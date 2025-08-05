package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"os"

	"github.com/daniel-munoz/go-namedParameterQuery/namedparameter"
	syslog "github.com/dmachard/go-clientsyslog"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/robfig/cron/v3"
)

var (
	config Config
	timeLocation *time.Location
	cache FileCache
	sysw  *syslog.Writer
	db *sql.DB

	secondParser = cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
)

func main() {
	// Read Config
	readConfigByte, err := os.ReadFile("config.json")
	if err != nil {
		log.Panicln("Failed to open config", err.Error())
	}

	err = json.Unmarshal(readConfigByte, &config)
	if err != nil {
		log.Panicln("Failed to unmarshal config JSON", err.Error())
	}
	log.Println("Config parsed successfully")

	// Init or load cache
	cacheLocation := "cache.json"
	if _, err := os.Stat(cacheLocation); err == nil {
		readCacheByte, err := os.ReadFile(cacheLocation)
		if err != nil {
			log.Panicln("Failed to open cache", err.Error())
		}

		err = json.Unmarshal(readCacheByte, &cache)
		if err != nil {
			log.Panicln("Failed to unmarshal cache JSON", err.Error())
		}

		log.Println("Cache parsed successfully")
	} else {
		cache = FileCache{
			LastChecks: make(map[string]time.Time),
		}
	}

	// Init Syslog
	if strings.Contains(config.Syslog.Prot, "tls") {
		sysw, err = syslog.DialWithTLSCertPath(config.Syslog.Prot, config.Syslog.Host, syslog.LOG_WARNING, "mssql-to-syslog", config.Syslog.CertFile)
	} else {
		sysw, err = syslog.Dial(config.Syslog.Prot, config.Syslog.Host, syslog.LOG_WARNING, "mssql-to-syslog")
	}
	
	if err != nil {
		log.Fatal("Failed to connect to syslog", err)
	}
	defer sysw.Close()

	// Init SQL
	db, err = sql.Open("mssql", config.Database)
	if err != nil {
		log.Fatal("Open connection failed", err.Error())
	}
	defer db.Close()

	// Schedule Cronjobs
	timeLocation, err = time.LoadLocation(config.Timezone)
	if err != nil {
		log.Panicln("Invalid timezone configured", err.Error())
	}

	c := cron.New(cron.WithLocation(timeLocation), cron.WithSeconds())
	for _, job := range config.Jobs {
		go ScheduleJob(c, job)
		log.Println("Scheduled", job.Name)
	}

	// Cron to save the cache
	c.AddFunc("0 * * * * *", func() {
		cacheJson, _ := json.Marshal(cache)
		os.WriteFile(cacheLocation, cacheJson, 0644)
	})
	
	log.Println("Init finished, starting jobs")
	sysw.Info("Init finished, starting jobs")

	c.Start()
	
	time.Sleep(5 * time.Minute)
}

func ScheduleJob(c *cron.Cron, job ConfigJob) {
	_, err := c.AddFunc(job.CronWithLeadingSecond, func() {
		RunJob(job)
	})

	if err != nil {
		log.Fatal("Failed to schedule job", job.Name, err)
	}
}

func RunJob(job ConfigJob) {
	var lastSync time.Time
	if cachedLastSync, ok := cache.LastChecks[job.Name]; ok {
		lastSync = cachedLastSync
	} else {
		lastSync = GetPastTimestamp(job.InitialLastCheckOffsetMin)
	}

	log.Println("Running job", job.Name, "since", lastSync)

	query := namedparameter.NewQuery(job.Query)
	query.SetValue("lastCheck", lastSync)

	// Update before we query to not miss any millisecond
	now := time.Now().In(timeLocation)
	cache.LastChecks[job.Name] = now

	rawRows, err := db.Query(query.GetParsedQuery(), (query.GetParsedParameters())...)
	if err != nil {
		log.Println("Query for", job.Name, "failed: ", err)
		return
	}
	defer rawRows.Close()
	columns, _ := rawRows.Columns()

	cronSchedule, err := secondParser.Parse(job.CronWithLeadingSecond)
	if err != nil {
		log.Fatal("Invalid cron", job.CronWithLeadingSecond, "for", job.Name, "- did you consider the leading non-standard second?")
	}

	// Load all rows to count them
	var rows []map[string]interface{}
	for rawRows.Next() {
		rows = append(rows, RowToMap(rawRows, columns))
	}
	if len(rows) == 0 {
		return
	}

	// Spread out delivery
	nextTrigger := cronSchedule.Next(now)
	msDelay := time.Duration(float64(nextTrigger.Sub(Now())) / float64(len(rows)))

	for _, row := range rows {
		row["meta:transfer-job-name"] = job.Name

		go SendRow(row)
		time.Sleep(msDelay)
	}
}

func SendRow(row map[string]interface{}) {
	if err := sysw.Info(MarshalMap(row)); err != nil {
		log.Println("Failed sending data via syslog", err)
	}
}