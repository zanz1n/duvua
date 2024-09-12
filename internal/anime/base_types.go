package anime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/zanz1n/duvua/internal/errors"
)

type StringInt int64

// String implements fmt.Stringer.
func (si *StringInt) String() string {
	return strconv.FormatInt(int64(*si), 10)
}

// UnmarshalJSON implements json.Unmarshaler.
func (si *StringInt) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	switch v := item.(type) {
	case int:
		*si = StringInt(v)
	case float64:
		*si = StringInt(int64(v))
	case string:
		i, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return err

		}
		*si = StringInt(i)

	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (si *StringInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + si.String() + "\""), nil
}

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
