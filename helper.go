package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"
)

func Now() time.Time {
	return time.Now().In(timeLocation)
}

func GetPastTimestamp(minsBack int64) time.Time {
	return time.Now().In(timeLocation).Add(time.Minute * time.Duration(minsBack) * -1)
}

// https://github.com/bdwilliams/go-jsonify/blob/master/jsonify/jsonify.go
func RowToMap(rows *sql.Rows, columns []string) map[string]interface{} {
	results := make(map[string]interface{})
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err := rows.Scan(scanArgs...)
	if err != nil {
		log.Panicln("Failed to read row", err.Error())
	}

	for i, value := range values {
		switch value.(type) {
		case nil:
			results[columns[i]] = nil

		case []byte:
			s := string(value.([]byte))
			x, err := strconv.Atoi(s)

			if err != nil {
				results[columns[i]] = s
			} else {
				results[columns[i]] = x
			}

		default:
			results[columns[i]] = value
		}
	}

	return results
}

func MarshalMap(val map[string]interface{}) string {
	b, err := json.Marshal(val)
	if err != nil {
		log.Panicln("Failed to marshal row data", err.Error())
	}

	return strings.TrimSpace(string(b))
}