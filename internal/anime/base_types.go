package anime

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zanz1n/duvua/internal/errors"
)

type Date struct {
	year       uint16
	month, day uint8
}

var (
	_ json.Marshaler   = &Date{}
	_ json.Unmarshaler = &Date{}
	_ fmt.Stringer     = &Date{}
)

func (d *Date) Year() uint16 {
	return d.year
}

func (d *Date) Month() uint8 {
	return d.month
}

func (d *Date) Day() uint8 {
	return d.day
}

func (d *Date) StringPtBr() string {
	return fmt.Sprintf("%d/%d/%d", d.day, d.month, d.year)
}

// String implements fmt.Stringer.
func (d *Date) String() string {
	return fmt.Sprintf("%d-%d-%d", d.year, d.month, d.day)
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(b []byte) error {
	_, err := fmt.Fscanf(
		bytes.NewReader(b),
		"\"%d-%d-%d\"",
		&d.year, &d.month, &d.day,
	)
	if err != nil {
		return err
	}

	if d.day > 32 || d.day < 1 {
		return errors.Unexpectedf(
			"cannot parse date `%s`: invalid day `%d`",
			string(b), d.day,
		)
	} else if d.month > 12 || d.month < 1 {
		return errors.Unexpectedf(
			"cannot parse date `%s`: invalid month `%d`",
			string(b), d.month,
		)
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (d *Date) MarshalJSON() ([]byte, error) {
	return []byte("\"" + d.String() + "\""), nil
}
