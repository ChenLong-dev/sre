package null

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// NullInt represents an int that may be null.
// NullInt implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type nullInt struct {
	Int   int
	Valid bool // Valid is true if Int is not NULL
}

// Scan implements the Scanner interface.
func (n *nullInt) Scan(value interface{}) error {
	if value == nil {
		n.Int, n.Valid = 0, false
		return nil
	}
	// use sql.NullInt64 Scan indirectly.
	var i64 Int64
	if err := i64.Scan(value); err != nil {
		return err
	}
	n.Int, n.Valid = int(i64.Int64), true
	return nil
}

// Value implements the driver Valuer interface.
func (n nullInt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return int64(n.Int), nil
}

// Int is an nullable int.
// It does not consider zero values to be null.
// It will decode to null, not zero, if null.
type Int struct {
	nullInt
}

// NewInt creates a new Int
func NewInt(i int, valid bool) Int {
	return Int{
		nullInt: nullInt{
			Int:   i,
			Valid: valid,
		},
	}
}

// IntFrom creates a new Int that will always be valid.
func IntFrom(i int) Int {
	return NewInt(i, true)
}

// IntFromPtr creates a new Int that be null if i is nil.
func IntFromPtr(i *int) Int {
	if i == nil {
		return NewInt(0, false)
	}
	return NewInt(*i, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (i Int) ValueOrZero() int {
	if !i.Valid {
		return 0
	}
	return i.Int
}

func (i Int) ValueOrDefault(d int) int {
	if !i.Valid {
		return d
	}
	return i.Int
}

func (i Int) SetDefault(d int) {
	if !i.Valid {
		i.SetValid(d)
	}
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Int.
// It also supports unmarshalling a NullInt.
func (i *Int) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		// Unmarshal again, directly to int64, to avoid intermediate float64
		err = json.Unmarshal(data, &i.Int)
	case string:
		str := string(x)
		if len(str) == 0 {
			i.Valid = false
			return nil
		}
		i.Int, err = strconv.Atoi(str)
	case map[string]interface{}:
		err = json.Unmarshal(data, &i.nullInt)
	case nil:
		i.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Int", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil
	return err
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Int if the input is a blank or not an integer.
// It will return an error if the input is not an integer, blank, or "null".
func (i *Int) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		i.Valid = false
		return nil
	}
	var err error
	i.Int, err = strconv.Atoi(string(text))
	i.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Int is null.
func (i Int) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.Itoa(i.Int)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Int is null.
func (i Int) MarshalText() ([]byte, error) {
	if !i.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.Itoa(i.Int)), nil
}

// SetValid changes this Int's value and also sets it to be non-null.
func (i *Int) SetValid(n int) {
	i.Int = n
	i.Valid = true
}

// Ptr returns a pointer to this Int's value, or a nil pointer if this Int is null.
func (i Int) Ptr() *int {
	if !i.Valid {
		return nil
	}
	return &i.Int
}

// IsZero returns true for invalid Ints, for future omitempty support (Go 1.4?)
// A non-null Int with a 0 value will not be considered zero.
func (i Int) IsZero() bool {
	return !i.Valid
}
