package fiqlparser

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

// ISO8601Duration represents a ISO 8601-2 duration
// the extension 2 allows a sign character (+/-) at
// the beginning of the duration
type ISO8601Duration struct {
	Negative bool
	Years    float64
	Months   float64
	Weeks    float64
	Days     float64
	Hours    float64
	Minutes  float64
	Seconds  float64
}

type iSO8601DurationConverter struct{}

var durationConverter = &iSO8601DurationConverter{}

func (*iSO8601DurationConverter) readNr(input string, pos int) (int, float64, error) {
	var b bytes.Buffer
	for len(input) > pos {
		if input[pos] == '+' {
			pos++
			continue
		}
		if input[pos] == '-' || input[pos] == '.' || unicode.IsNumber(rune(input[pos])) {
			b.WriteByte(input[pos])
			pos++
		} else {
			break
		}
	}
	r, err := strconv.ParseFloat(b.String(), 64)
	return pos, r, err
}

func (i *iSO8601DurationConverter) tryParseISO8601Duration(input string) (ISO8601Duration, error) {
	d := ISO8601Duration{}
	if len(input) == 0 {
		return d, nil
	}
	pos := 0

	if input[0] == '-' {
		pos++
		d.Negative = true
	} else if input[0] == '+' {
		pos++
	}
	if input[pos] != durationPeriod {
		return d, fmt.Errorf("expected P but got `%c`", input[pos])
	}
	pos++
	isTime := false
	for pos < len(input) {
		if input[pos] == durationTime {
			isTime = true
			pos++
		}
		newPos, nr, err := i.readNr(input, pos)
		pos = newPos
		if err != nil {
			return d, err
		}
		mark := input[pos]
		pos++
		switch mark {
		case durationYear:
			d.Years = nr
		case durationMonthOrMinute:
			if isTime {
				d.Minutes = nr
			} else {
				d.Months = nr
			}

		case durationWeek:
			d.Weeks = nr
		case durationDay:
			d.Days = nr
		case durationHour:
			d.Hours = nr
		case durationSecond:
			d.Seconds = nr
		default:
			return d, fmt.Errorf("unexpected token `%c`", mark)
		}
	}

	return d, nil
}

const durationPeriod byte = 'P'
const durationTime byte = 'T'
const durationYear byte = 'Y'
const durationMonthOrMinute byte = 'M'
const durationWeek byte = 'W'
const durationDay byte = 'D'
const durationHour byte = 'H'
const durationSecond byte = 'S'
