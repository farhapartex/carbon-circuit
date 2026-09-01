package database

import (
	"database/sql/driver"
	"fmt"
)

type JSONDocument []byte

func (d JSONDocument) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	return string(d), nil
}

func (d *JSONDocument) Scan(source any) error {
	switch value := source.(type) {
	case nil:
		*d = nil
	case []byte:
		*d = append((*d)[:0], value...)
	case string:
		*d = JSONDocument(value)
	default:
		return fmt.Errorf("cannot scan %T into a json document", source)
	}
	return nil
}

func (d JSONDocument) Bytes() []byte { return d }
