package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type TimeString string

func (t *TimeString) Scan(value interface{}) error {
	if value == nil {
		*t = ""
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = TimeString(v.Format(time.RFC3339))
		return nil
	case string:
		*t = TimeString(v)
		return nil
	case []byte:
		*t = TimeString(string(v))
		return nil
	}
	return fmt.Errorf("cannot scan %T into TimeString", value)
}

func (t TimeString) Value() (driver.Value, error) {
	return string(t), nil
}
