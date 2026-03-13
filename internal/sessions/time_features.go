package sessions

import "time"

const analyticsTimeZone = "Europe/Zurich"

type temporalFeatures struct {
	EventDate        string
	EventWeekday     int
	EventStartTime   string
	EventStartHour   int
	EventStartMinute int
	EventTimeKnown   bool
}

func deriveTemporalFeatures(startRaw string, fallbackDateUnix int64) temporalFeatures {
	loc := loadAnalyticsLocation()
	fallback := time.Unix(fallbackDateUnix, 0).In(loc)

	parsed, ok := parseEventStart(startRaw, loc)
	if !ok {
		return temporalFeatures{
			EventDate:        fallback.Format("2006-01-02"),
			EventWeekday:     isoWeekday(fallback),
			EventStartTime:   "",
			EventStartHour:   0,
			EventStartMinute: 0,
			EventTimeKnown:   false,
		}
	}

	return temporalFeatures{
		EventDate:        parsed.Format("2006-01-02"),
		EventWeekday:     isoWeekday(parsed),
		EventStartTime:   parsed.Format("15:04"),
		EventStartHour:   parsed.Hour(),
		EventStartMinute: parsed.Hour()*60 + parsed.Minute(),
		EventTimeKnown:   true,
	}
}

func loadAnalyticsLocation() *time.Location {
	loc, err := time.LoadLocation(analyticsTimeZone)
	if err != nil {
		return time.Local
	}
	return loc
}

func parseEventStart(value string, loc *time.Location) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	layoutsWithZone := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04-0700",
	}
	for _, layout := range layoutsWithZone {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.In(loc), true
		}
	}

	layoutsLocal := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
	}
	for _, layout := range layoutsLocal {
		if parsed, err := time.ParseInLocation(layout, value, loc); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func isoWeekday(t time.Time) int {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}
