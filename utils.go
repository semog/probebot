package main

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func getTimeStamp() int64 {
	return time.Now().UTC().UnixNano()
}

func intrg_contains(rg []int, item int) bool {
	return slices.Contains(rg, item)
}

// Regular expressions for different formats
var dayOfWeekRegex = regexp.MustCompile(`^every[ \t]+(monday|mon|tuesday|tue|wednesday|wed|thursday|thu|friday|fri|saturday|sat|sunday|sun)([ \t]+at)?[ \t]+(\d{1,2})(am|a|pm|p)$`)
var dayOfMonthRegex = regexp.MustCompile(`^every[ \t]+(\d{1,2})(?:st|nd|rd|th)?([ \t]+at)?[ \t]+(\d{1,2})(am|a|pm|p)$`)
var monthAndDayRegex = regexp.MustCompile(`^every[ \t]+(january|jan|february|feb|march|mar|april|apr|may|june|jun|july|jul|august|aug|september|sep|october|oct|november|nov|december|dec)[ \t]+(\d{1,2})(?:st|nd|rd|th)?([ \t]+at)?[ \t]+(\d{1,2})(am|a|pm|p)$`)

// parseEveryInterval parses the everyInterval string and returns the next scheduled time.
func parseEveryInterval(now time.Time, everyInterval string) (time.Time, error) {
	everyInterval = strings.ToLower(everyInterval)

	// Check for day of the week format
	if matches := dayOfWeekRegex.FindStringSubmatch(everyInterval); matches != nil {
		dayOfWeekStr := matches[1]
		hourStr := matches[3]
		amPm := matches[4]

		dayOfWeek, err := parseDayOfWeek(dayOfWeekStr)
		if err != nil {
			return time.Time{}, err
		}

		hour, err := strconv.Atoi(hourStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid hour format: %v", err)
		}

		hour = convertTo24Hour(hour, amPm)

		return calculateNextDayOfWeek(now, dayOfWeek, hour), nil
	}

	// Check for day of the month format
	if matches := dayOfMonthRegex.FindStringSubmatch(everyInterval); matches != nil {
		dayOfMonthStr := matches[1]
		hourStr := matches[3]
		amPm := matches[4]

		dayOfMonth, err := strconv.Atoi(dayOfMonthStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day of month format: %v", err)
		}
		if dayOfMonth < 1 || dayOfMonth > 31 {
			return time.Time{}, fmt.Errorf("day of month must be between 1 and 31")
		}

		hour, err := strconv.Atoi(hourStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid hour format: %v", err)
		}

		hour = convertTo24Hour(hour, amPm)

		return calculateNextDayOfMonth(now, dayOfMonth, hour), nil
	}

	// Check for month and day format
	if matches := monthAndDayRegex.FindStringSubmatch(everyInterval); matches != nil {
		monthStr := matches[1]
		dayOfMonthStr := matches[2]
		hourStr := matches[4]
		amPm := matches[5]

		month, err := parseMonth(monthStr)
		if err != nil {
			return time.Time{}, err
		}

		dayOfMonth, err := strconv.Atoi(dayOfMonthStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day of month format: %v", err)
		}
		if dayOfMonth < 1 || dayOfMonth > 31 {
			return time.Time{}, fmt.Errorf("day of month must be between 1 and 31")
		}

		hour, err := strconv.Atoi(hourStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid hour format: %v", err)
		}

		hour = convertTo24Hour(hour, amPm)

		return calculateNextMonthAndDay(now, month, dayOfMonth, hour), nil
	}

	return time.Time{}, fmt.Errorf("invalid everyInterval format: %s", everyInterval)
}

// parseDayOfWeek converts a day of the week string to a time.Weekday.
func parseDayOfWeek(dayOfWeekStr string) (time.Weekday, error) {
	var shortDayOfWeekStr string
	if len(dayOfWeekStr) > 3 {
		shortDayOfWeekStr = dayOfWeekStr[:3]
	} else {
		shortDayOfWeekStr = dayOfWeekStr
	}

	switch shortDayOfWeekStr {
	case "mon":
		return time.Monday, nil
	case "tue":
		return time.Tuesday, nil
	case "wed":
		return time.Wednesday, nil
	case "thu":
		return time.Thursday, nil
	case "fri":
		return time.Friday, nil
	case "sat":
		return time.Saturday, nil
	case "sun":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("invalid day of the week: %s", dayOfWeekStr)
	}
}

// parseMonth converts a month string to a time.Month.
func parseMonth(monthStr string) (time.Month, error) {
	var shortMonth string
	if len(monthStr) > 3 {
		shortMonth = monthStr[:3]
	} else {
		shortMonth = monthStr
	}

	switch shortMonth {
	case "jan":
		return time.January, nil
	case "feb":
		return time.February, nil
	case "mar":
		return time.March, nil
	case "apr":
		return time.April, nil
	case "may":
		return time.May, nil
	case "jun":
		return time.June, nil
	case "jul":
		return time.July, nil
	case "aug":
		return time.August, nil
	case "sep":
		return time.September, nil
	case "oct":
		return time.October, nil
	case "nov":
		return time.November, nil
	case "dec":
		return time.December, nil
	default:
		return 0, fmt.Errorf("invalid month: %s", monthStr)
	}
}

// convertTo24Hour converts a 12-hour format hour to 24-hour format.
func convertTo24Hour(hour int, amPm string) int {
	if strings.HasPrefix(amPm, "p") && hour != 12 {
		hour += 12
	}
	if strings.HasPrefix(amPm, "a") && hour == 12 {
		hour = 0
	}
	return hour
}

// calculateNextDayOfWeek calculates the next occurrence of a specific day of the week at a specific hour.
func calculateNextDayOfWeek(now time.Time, dayOfWeek time.Weekday, hour int) time.Time {
	daysUntilNext := int(dayOfWeek - now.Weekday())
	nextTime := now.AddDate(0, 0, daysUntilNext)
	nextTime = time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), hour, 0, 0, 0, now.Location())

	// If the calculated time is in the past, add 7 days
	if nextTime.Before(now) || nextTime.Equal(now) {
		nextTime = nextTime.AddDate(0, 0, 7)
	}
	return nextTime
}

// calculateNextDayOfMonth calculates the next occurrence of a specific day of the month at a specific hour.
func calculateNextDayOfMonth(now time.Time, dayOfMonth int, hour int) time.Time {
	nextTime := time.Date(now.Year(), now.Month(), dayOfMonth, hour, 0, 0, 0, now.Location())

	// If the calculated time is in the past, move to next month
	if nextTime.Before(now) || nextTime.Equal(now) {
		nextTime = time.Date(now.Year(), now.Month()+1, dayOfMonth, hour, 0, 0, 0, now.Location())
		// If the day of month is invalid for the next month, move to the next valid day
		for nextTime.Month() != now.Month()+1 {
			nextTime = time.Date(now.Year(), now.Month()+2, dayOfMonth, hour, 0, 0, 0, now.Location())
		}
	}
	return nextTime
}

// calculateNextMonthAndDay calculates the next occurrence of a specific month and day at a specific hour.
func calculateNextMonthAndDay(now time.Time, month time.Month, dayOfMonth int, hour int) time.Time {
	nextTime := time.Date(now.Year(), month, dayOfMonth, hour, 0, 0, 0, now.Location())

	// If the calculated time is in the past, move to next year
	if nextTime.Before(now) || nextTime.Equal(now) {
		nextTime = time.Date(now.Year()+1, month, dayOfMonth, hour, 0, 0, 0, now.Location())
	}
	return nextTime
}

// calcNextServiceTime calculates the next service time based on the current time and the interval
func calcNextServiceTime(now int64, everyInterval string) int64 {
	t, err := parseEveryInterval(time.Unix(now, 0), everyInterval)
	if err != nil {
		return 0
	}
	return t.Unix()
}
