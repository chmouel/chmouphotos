package photos

import (
	"encoding/json"
	"strings"
	"time"
)

type SimpleJsonDate struct {
	time.Time
}

// Implement Marshaler and Unmarshaler interface
func (j *SimpleJsonDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*j = SimpleJsonDate{t}
	return nil
}

func (j SimpleJsonDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(j)
}

// Maybe a Format function for printing your date
func (j SimpleJsonDate) DayMonthDir() string {
	t := SimpleJsonDate(j)
	return t.Format("2006/01")
}
