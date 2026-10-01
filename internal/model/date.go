package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// Date represents a calendar date without a time component, used for Task.DueDate.
// It marshals to/from "YYYY-MM-DD" in JSON and maps to a SQL DATE column.
type Date struct {
	time.Time
}

// NewDate builds a Date from a time.Time, discarding the time-of-day component.
func NewDate(t time.Time) Date {
	y, m, d := t.Date()
	return Date{time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(dateLayout))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", s, err)
	}
	d.Time = t
	return nil
}

// Value implements driver.Valuer so Date can be written directly by database/sql.
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

// Scan implements sql.Scanner so Date can be read directly by database/sql.
func (d *Date) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan %T into model.Date", value)
	}
}
