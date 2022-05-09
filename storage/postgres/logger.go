package postgres

import (
	"database/sql"
	"fmt"

	"github.com/cmorales95/cloud-native-go-exercises/entities"
)

type TransactionLogger struct {
	events chan<- entities.Event
	errors <-chan error
	db     *sql.DB
}

func NewTransactionLogger(db *sql.DB) (*TransactionLogger, error) {
	logger := TransactionLogger{db: db}
	exists, err := logger.verifyTableExist()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}
	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}
	return &logger, nil
}

func (t *TransactionLogger) WriteDelete(key string) {
	t.events <- entities.Event{EventType: entities.EventDelete, Key: key}
}

func (t *TransactionLogger) WritePut(key, value string) {
	t.events <- entities.Event{
		EventType: entities.EventPut,
		Key:       key,
		Value:     value,
	}
}

func (t *TransactionLogger) Err() <-chan error {
	return t.errors
}

func (t *TransactionLogger) ReadEvents() (<-chan entities.Event, <-chan error) {
	outEvent := make(chan entities.Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := `SELECT sequence, event_type, key, value FROM transactions ORDER BY sequence`
		rows, err := t.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error %w", err)
			return
		}

		defer rows.Close()

		var e entities.Event

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}

			outEvent <- e
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func (t *TransactionLogger) Run() {
	events := make(chan entities.Event, 16)
	t.events = events

	errors := make(chan error, 1)
	t.errors = errors

	go func() {
		query := `INSERT INTO transactions (event_type, key, value) VALUES ($1, $2, $3);`
		for e := range events {
			_, err := t.db.Exec(query, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (t *TransactionLogger) verifyTableExist() (bool, error) {
	var exists bool
	err := t.db.QueryRow(`select exists(
									select from pg_tables
									where schemaname = 'public' and
										  tablename = 'transactions');
	`).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (t *TransactionLogger) createTable() error {
	_, err := t.db.Exec(`CREATE TABLE transactions (
								sequence SERIAL primary key,
								event_type text,
								key text,
								value text,
								created_at timestamp default now()
							);`)
	if err != nil {
		return err
	}
	return nil
}
