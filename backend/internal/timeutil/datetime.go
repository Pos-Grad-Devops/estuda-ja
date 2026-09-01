package timeutil

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	DateLayout     = "02/01/2006"
	DateTimeLayout = "02/01/2006 15:04"
)

var location = time.Local

func ParseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("data vazia")
	}
	t, err := time.ParseInLocation(DateLayout, value, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("data inválida, use dd/mm/yyyy")
	}
	return t, nil
}

func ParseDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("data vazia")
	}
	if t, err := time.ParseInLocation(DateTimeLayout, value, location); err == nil {
		return t, nil
	}
	t, err := time.ParseInLocation(DateLayout, value, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("data inválida, use dd/mm/yyyy ou dd/mm/yyyy HH:mm")
	}
	return t, nil
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(location).Format(DateLayout)
}

func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	t = t.In(location)
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format(DateLayout)
	}
	return t.Format(DateTimeLayout)
}

// DateTime armazena timestamp no banco e serializa em dd/mm/yyyy HH:mm.
type DateTime struct {
	time.Time
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte(`""`), nil
	}
	formatted := FormatDateTime(d.Time)
	return []byte(`"` + formatted + `"`), nil
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	raw := strings.Trim(string(data), `"`)
	if raw == "" || raw == "null" {
		d.Time = time.Time{}
		return nil
	}
	parsed, err := ParseDateTime(raw)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

func (d DateTime) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

func (d *DateTime) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	default:
		return fmt.Errorf("tipo de data não suportado: %T", value)
	}
}
