package transactionlogger

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	. "github.com/go-jet/jet/v2/postgres"
	table "github.com/mrityunjaygr8/cloud-native-go/kvapi/db/gen/kvapi/public/table"

	"github.com/mrityunjaygr8/cloud-native-go/kvapi/db/gen/kvapi/public/model"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	log.Println("WritePut")
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}
func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}
func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}
func (l *PostgresTransactionLogger) Close() {
	log.Println("In close method of PostgresTransactionLogger")

	for len(l.events) > 0 {
		log.Println(fmt.Sprintf("Items still in channel: %d", len(l.events)))
		time.Sleep(time.Millisecond * 250)
	}

	l.db.Close()
	log.Println("Exiting PostgresTransactionLogger")
	os.Exit(0)
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {

			if e.EventType == EventDelete {
				stmt := table.Events.DELETE().WHERE(table.Events.Key.EQ(String(e.Key)))
				_, err := stmt.Exec(l.db)
				if err != nil {
					errors <- err
					return
				}

			} else {
				event := model.Events{
					Key:       e.Key,
					Value:     e.Value,
					EventType: int32(e.EventType),
				}
				stmt := table.Events.INSERT(table.Events.MutableColumns).MODEL(event).ON_CONFLICT(table.Events.Key).DO_UPDATE(
					SET(
						table.Events.Value.SET(String(event.Value)),
					),
				)

				_, err := stmt.Exec(l.db)

				if err != nil {
					errors <- err
					return
				}
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)
	go func() {

		defer close(outError)
		defer close(outEvent)

		stmt := table.Events.SELECT(table.Events.AllColumns).ORDER_BY(table.Events.Sequence.ASC())
		var dest []struct {
			model.Events
		}

		err := stmt.Query(l.db, &dest)
		if err != nil {
			outError <- fmt.Errorf("input parse error: %w", err)
			return
		}

		for _, event := range dest {
			outEvent <- Event{Key: event.Key, Value: event.Value, EventType: EventType(event.EventType)}
		}

	}()

	return outEvent, outError
}

type PostgresConfig struct {
	Host     string
	Password string
	Username string
	DbName   string
	Port     int
	SslMode  string
}

func NewPostgresTransactionLogger() (TransactionLogger, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	type Config struct {
		DbName   string `mapstructure:"KVAPI_DB_NAME"`
		Password string `mapstructure:"KVAPI_DB_PASS"`
		Username string `mapstructure:"KVAPI_DB_USER"`
		Port     int    `mapstructure:"KVAPI_DB_PORT"`
		SslMode  string `mapstructure:"KVAPI_DB_SSL"`
		Host     string `mapstructure:"KVAPI_DB_HOST"`
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
		log.Fatal(fmt.Errorf("An error occurred when reading configuration from .env"))
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal(fmt.Errorf("An error occurred when reading configuration from .env"))
	}
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s port=%d sslmode=%s",
		config.Host, config.DbName, config.Username, config.Password, config.Port, config.SslMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	err = db.Ping() // Test the database connection
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}
	return &PostgresTransactionLogger{db: db}, nil
}
