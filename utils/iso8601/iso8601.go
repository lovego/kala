package iso8601

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"text/template"
	"time"
)

var (
	// ErrBadFormat is returned when parsing fails
	ErrBadFormat = errors.New("bad format string")

	//nolint:lll
	tmpl = template.Must(template.New("duration").Parse(`P{{if .Years}}{{.Years}}Y{{end}}{{if .Months}}{{.Months}}M{{end}}{{if .Weeks}}{{.Weeks}}W{{end}}{{if .Days}}{{.Days}}D{{end}}{{if .HasTimePart}}T{{end }}{{if .Hours}}{{.Hours}}H{{end}}{{if .Minutes}}{{.Minutes}}M{{end}}{{if .Seconds}}{{.Seconds}}S{{end}}`))

	//nolint:lll
	full = regexp.MustCompile(`P(?:(?P<year>\d+)Y)?(?:(?P<month>\d+)M)?(?:(?P<day>\d+)D)?(?:(?P<time>T)(?:(?P<hour>\d+)H)?(?:(?P<minute>\d+)M)?(?:(?P<second>\d+)S)?)?`)
	week = regexp.MustCompile(`P(?:(?P<week>\d+)W)`)
)

type Duration struct {
	Years   int `json:"years"`
	Months  int `json:"months"`
	Weeks   int `json:"weeks"`
	Days    int `json:"days"`
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

func FromString(dur string) (*Duration, error) {
	var (
		match []string
		re    *regexp.Regexp
	)

	switch {
	case week.MatchString(dur):
		match = week.FindStringSubmatch(dur)
		re = week
	case full.MatchString(dur):
		match = full.FindStringSubmatch(dur)
		re = full
	default:
		return nil, ErrBadFormat
	}

	d := &Duration{}

	timeOccurred := false
	dateUnspecified := true
	weekUnspecified := true
	timeUnspecified := true
	for i, name := range re.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}
		if name == "time" {
			timeOccurred = true
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		switch name {
		case "year":
			d.Years = val
			dateUnspecified = false
		case "month":
			d.Months = val
			dateUnspecified = false
		case "week":
			d.Weeks = val
			weekUnspecified = false
		case "day":
			d.Days = val
			dateUnspecified = false
		case "hour":
			d.Hours = val
			timeUnspecified = false
		case "minute":
			d.Minutes = val
			timeUnspecified = false
		case "second":
			d.Seconds = val
			timeUnspecified = false
		default:
			return nil, fmt.Errorf("unknown field %s", name)
		}
	}

	if (dateUnspecified && weekUnspecified && timeUnspecified) || (timeOccurred && timeUnspecified) {
		return nil, fmt.Errorf("invalid ISO 8601 duration spec %s", dur)
	}

	return d, nil
}

// String prints out the value passed in. It's not strictly according to the
// ISO spec, but it's pretty close. In particular, to completely conform it
// would need to round up to the next largest unit. 61 seconds to 1 minute 1
// second, for example. It would also need to disallow weeks mingling with
// other units.
func (d *Duration) String() string {
	var s bytes.Buffer

	err := tmpl.Execute(&s, d)
	if err != nil {
		panic(err)
	}

	return s.String()
}

func (d *Duration) HasTimePart() bool {
	return d.Hours != 0 || d.Minutes != 0 || d.Seconds != 0
}

func (d *Duration) RelativeTo(t time.Time) time.Duration {
	after := d.Add(t)
	return after.Sub(t)
}

func (d *Duration) Add(t time.Time) time.Time {
	result := addYearsMonths(t, d.Years, d.Months)
	result = result.AddDate(0, 0, d.Days+d.Weeks*7)
	result = result.Add(time.Hour * time.Duration(d.Hours))
	result = result.Add(time.Minute * time.Duration(d.Minutes))
	result = result.Add(time.Second * time.Duration(d.Seconds))
	return result
}

func (d *Duration) IsZero() bool {
	switch {
	case d.Years != 0:
	case d.Months != 0:
	case d.Weeks != 0:
	case d.Days != 0:
	case d.Hours != 0:
	case d.Minutes != 0:
	case d.Seconds != 0:
	default:
		return true
	}
	return false
}

// 增加年和月数，同时处理超出月份天数的情况
func addYearsMonths(t time.Time, addYears int, addMonths int) time.Time {
	// 先增加年月并处理年月溢出
	year := t.Year() + addYears
	month := t.Month() + time.Month(addMonths)
	year += (int(month) - 1) / 12
	month = ((month - 1) % 12) + 1

	// 构造目标日期，暂时用原日期的日
	target := time.Date(year, month, t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// 如果目标日期溢出（变成了下个月或更远），则设为该月最后一天
	if target.Month() != month {
		lastDay := lastDayOfMonth(year, month, t.Location())
		target = time.Date(year, month, lastDay, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	}
	return target
}

// 获取某月的最后一天
func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	var nextMonth time.Time
	if month == 12 {
		nextMonth = time.Date(year+1, 1, 1, 0, 0, 0, 0, loc)
	} else {
		nextMonth = time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	}
	return nextMonth.AddDate(0, 0, -1).Day()
}
