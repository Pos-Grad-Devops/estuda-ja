package timeutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDateTime(t *testing.T) {
	t.Run("date and time", func(t *testing.T) {
		parsed, err := ParseDateTime("29/08/2026 19:00")
		require.NoError(t, err)
		require.Equal(t, 29, parsed.Day())
		require.Equal(t, time.August, parsed.Month())
		require.Equal(t, 2026, parsed.Year())
		require.Equal(t, 19, parsed.Hour())
	})

	t.Run("date only", func(t *testing.T) {
		parsed, err := ParseDateTime("29/08/2026")
		require.NoError(t, err)
		require.Equal(t, FormatDate(parsed), "29/08/2026")
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseDateTime("2026-08-29")
		require.Error(t, err)
	})
}

func TestDateTimeJSON(t *testing.T) {
	value := DateTime{Time: time.Date(2026, 8, 29, 19, 0, 0, 0, location)}
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.Equal(t, `"29/08/2026 19:00"`, string(data))

	var decoded DateTime
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, value.Time.Format(DateTimeLayout), decoded.Time.Format(DateTimeLayout))
}
