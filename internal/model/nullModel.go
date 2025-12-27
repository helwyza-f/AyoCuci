package model

import (
	"fmt"
	"strconv"
	"encoding/json"
	"database/sql/driver"
)

// NullString handles nullable string values
type NullString struct {
	Str   string
	Valid bool
}

func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.Str = ""
		ns.Valid = false
		return nil
	}
	switch v := value.(type) {
	case string:
		ns.Str = v
	case []byte:
		ns.Str = string(v) 
	default:
		return fmt.Errorf("NullString: expected string, got %T", value)
	}
	ns.Valid = true
	return nil
}

func (ns NullString) DriverValue() (driver.Value, error) {
	if ns.Valid {
		return ns.Str, nil
	}
	return nil, nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Str) // Just return the string
	}
	return json.Marshal(nil) // Return null if not valid
}

// NullInt handles nullable integer values
type NullInt struct {
	Int   int
	Valid bool
}

func (ni *NullInt) Scan(value interface{}) error {
	if value == nil {
		ni.Int = 0
		ni.Valid = false
		return nil
	}
	switch v := value.(type) {
	case int64:
		ni.Int = int(v)
	case int:
		ni.Int = v
	case []byte:
		intVal, err := strconv.Atoi(string(v))
		if err != nil {
			return fmt.Errorf("NullInt: cannot convert []byte to int: %v", err)
		}
		ni.Int = intVal
	default:
		return fmt.Errorf("NullInt: expected int, got %T", value)
	}
	ni.Valid = true
	return nil
}

func (ni NullInt) DriverValue() (driver.Value, error) {
	if ni.Valid {
		return ni.Int, nil
	}
	return nil, nil
}

func (ni NullInt) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int)
	}
	return json.Marshal(nil)
}

// NullBool handles nullable boolean values
type NullBool struct {
	Bool  bool
	Valid bool
}

func (nb *NullBool) Scan(value interface{}) error {
	if value == nil {
		nb.Bool = false
		nb.Valid = false
		return nil
	}
	switch v := value.(type) {
	case bool:
		nb.Bool = v
	case []byte:
		nb.Bool = string(v) == "1" // MySQL returns "1" for true, "0" for false
	default:
		return fmt.Errorf("NullBool: expected bool, got %T", value)
	}
	nb.Valid = true
	return nil
}

func (nb NullBool) DriverValue() (driver.Value, error) {
	if nb.Valid {
		return nb.Bool, nil
	}
	return nil, nil
}

func (nb NullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(nil)
}

// NullFloat handles nullable float values
type NullFloat struct {
	Float float64
	Valid bool
}

func (nf *NullFloat) Scan(value interface{}) error {
	if value == nil {
		nf.Float = 0.0
		nf.Valid = false
		return nil
	}
	switch v := value.(type) {
	case float64:
		nf.Float = v
	case float32:
		nf.Float = float64(v)
	case []byte:
		floatVal, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return fmt.Errorf("NullFloat: cannot convert []byte to float64: %v", err)
		}
		nf.Float = floatVal
	default:
		return fmt.Errorf("NullFloat: expected float, got %T", value)
	}
	nf.Valid = true
	return nil
}

func (nf NullFloat) DriverValue() (driver.Value, error) {
	if nf.Valid {
		return nf.Float, nil
	}
	return nil, nil
}

func (nf NullFloat) MarshalJSON() ([]byte, error) {
	if nf.Valid {
		return json.Marshal(nf.Float)
	}
	return json.Marshal(nil)
}
