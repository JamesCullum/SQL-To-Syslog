# SQL 2 Syslog

Some applications contain access logs or other relevant data in an SQL database. This lightweight and performant tool queries the database with a cron syntax and forwards the data to your favorite logging tool or SIEM via Syslog protocol.

# Features

- Single executable, compatible with every OS

- Docker / K8s

- Lightweight & Fast

- Auto-resume after failure with checkpoints

- Deliver messages evenly over allocated timeframe

As database backend, currently only Microsoft SQL is supported. Other databases can be easily added - contributions are more than welcome.

## Config

The configuration with an example is contained in the config.json. If you want to connect multiple syslog servers or use multiple credentials, you'll need to use multiple instances.

- Database: Connection string to compatible database, e.g. [MSSQL](https://github.com/microsoft/go-mssqldb)

- Name: This is just used for error messages and caching

- cronWithLeadingSecond: Typical cron format, but with a leading second to allow intervals smaller than one minute, [see more](https://pkg.go.dev/github.com/robfig/cron/v3#section-readme)

- Query: If you didn't specify a database in the connection string, remember to use the `USE` command to select a database. Use `:lastCheck` to get the time of the last sync. If the datetime in your table does not contain a timezone, make sure to cast that column to a matching timezone.

- initialLastCheckOffsetMin: On the first run, how many minutes in the past should `:lastCheck` be?

# Sponsors

[![CRIF GmbH](https://www.crif.de/media/2447/crif_tagline.jpg)](https://careers.crif.com/search/?optionsFacetsDD_country=DE)
