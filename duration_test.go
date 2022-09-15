package fiqlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryParseISO8601Duration(t *testing.T) {
	var values = []struct {
		input       string
		duration    ISO8601Duration
		errorOutput bool
	}{
		//One year, one month, one day, one hour, one minute, one second, and 100 milliseconds
		{input: "P1Y1M1DT1H1M1.1S", duration: ISO8601Duration{Negative: false, Years: 1, Months: 1, Weeks: 0, Days: 1, Hours: 1, Minutes: 1, Seconds: 1.1}, errorOutput: false},
		{input: "-P1Y1M1DT1H1M1.1S", duration: ISO8601Duration{Negative: true, Years: 1, Months: 1, Weeks: 0, Days: 1, Hours: 1, Minutes: 1, Seconds: 1.1}, errorOutput: false},
		{input: "+P1Y1M1DT1H1M1.1S", duration: ISO8601Duration{Negative: false, Years: 1, Months: 1, Weeks: 0, Days: 1, Hours: 1, Minutes: 1, Seconds: 1.1}, errorOutput: false},
		//Forty days
		{input: "P40D", duration: ISO8601Duration{Days: 40}, errorOutput: false},
		//A year and a day
		{input: "P1Y1D", duration: ISO8601Duration{Years: 1, Days: 1}, errorOutput: false},
		//Three days, four hours and 59 minutes
		{input: "P3DT4H59M", duration: ISO8601Duration{Days: 3, Hours: 4, Minutes: 59}, errorOutput: false},
		//Two and a half hours
		{input: "PT2H30M", duration: ISO8601Duration{Hours: 2, Minutes: 30}, errorOutput: false},
		//One month
		{input: "P1M", duration: ISO8601Duration{Months: 1}, errorOutput: false},
		//One minute
		{input: "PT1M", duration: ISO8601Duration{Minutes: 1}, errorOutput: false},
		//2.1 milliseconds
		{input: "PT0.0021S", duration: ISO8601Duration{Seconds: 0.0021}, errorOutput: false},
		// one week
		{input: "P1W", duration: ISO8601Duration{Weeks: 1}, errorOutput: false},
		// zeros
		{input: "PT0S", duration: ISO8601Duration{}, errorOutput: false},
		{input: "P0D", duration: ISO8601Duration{}, errorOutput: false},
		{input: "", duration: ISO8601Duration{}, errorOutput: false},
		//errors
		{input: "PD", duration: ISO8601Duration{}, errorOutput: true},
		{input: "+1Y1D", duration: ISO8601Duration{}, errorOutput: true},
		{input: "1Y1D", duration: ISO8601Duration{}, errorOutput: true},
		{input: "P1X", duration: ISO8601Duration{}, errorOutput: true},
	}

	for _, v := range values {
		d, err := durationConverter.tryParseISO8601Duration(v.input)
		if !v.errorOutput {
			assert.NoError(t, err)
			if err != nil {
				return
			}
		} else {
			assert.Error(t, err)
		}
		assert.Equal(t, v.duration, d)
	}
}
