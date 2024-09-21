package utils

import (
	"encoding/json"
	"strconv"
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
