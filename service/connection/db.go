package connection

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"

	service_sign "biblebrain-services/service/sign"

	"github.com/go-sql-driver/mysql"
	sqldblogger "github.com/simukti/sqldb-logger"
)

// SlogAdapter struct to adapt the slog to sqldb-logger's Logger interface.
type SlogAdapter struct{}

// Log method to satisfy sqldb-logger's Logger interface.
func (l SlogAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	// Adapt this method according to how slog accepts log messages.
	switch level {
	case sqldblogger.LevelError:
		slog.Error(msg, "data", data)
	case sqldblogger.LevelInfo:
		// Check if the log is from a QueryContext
		if msg == "QueryContext" {
			slog.Debug(msg, "data", data)
		} else {
			slog.Info(msg, "data", data)
		}
	case sqldblogger.LevelDebug:
		slog.Debug(msg, "data", data)
	case sqldblogger.LevelTrace:
		slog.Warn(msg, "data", data)
	default:
		slog.Info("Unhandled log level", "level", level, "msg", msg, "data", data)
		// Add other cases as needed.
	}
}

func GetBibleBrainDB(ctx context.Context) *sql.DB {
	config := getBibleBrainSQLConfig(ctx)
	dsn := config.FormatDSN()
	slog.Debug("MYSQL_CONNECT_STRING", "dsn", dsn)
	databaseInfo := dsn + "?parseTime=true&interpolateParams=true"
	conn, err := sql.Open("mysql", databaseInfo)
	if err != nil {
		log.Panic(err)
	}

	PingDB(conn)

	environment := os.Getenv("environment")
	if environment != "prod" {
		sqlCon := sqldblogger.OpenDriver(
			databaseInfo,
			conn.Driver(),
			SlogAdapter{},
		)

		return sqlCon
	}

	return conn
}

func PingDB(conn *sql.DB) {
	slog.Info("attempting ping")
	err := conn.Ping()
	if err != nil {
		log.Panic(err)
	}

	slog.Info("success pinging datasource")
}

func getBiblebrainDSN(ctx context.Context) string {
	environment := os.Getenv("environment")

	if environment == "local" {
		return os.Getenv("BIBLEBRAIN_DSN")
	}

	ssmClient := service_sign.GetSSMClient(ctx)
	parameterName := os.Getenv("BIBLEBRAIN_DSN_SSM_ID")

	return *service_sign.GetSsmParameter(ctx, ssmClient, parameterName)
}

func getBibleBrainSQLConfig(ctx context.Context) *mysql.Config {
	dsn := getBiblebrainDSN(ctx)

	if len(dsn) < 1 {
		panic("BIBLEBRAIN_DSN not set")
	}

	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic("unable to parse BIBLEBRAIN_DSN ")
	}

	return config
}
