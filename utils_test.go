package main

import (
	"fmt"
	"testing"
	"time"
)

var testnow = time.Date(2023, time.October, 26, 10, 0, 0, 0, time.Local).Unix()

func testCalcNextServiceTimeHelper(t *testing.T, everyInterval string, expectedTime int64) {
	fmt.Println("testing: ", everyInterval)
	actualTime := calcNextServiceTime(testnow, everyInterval)
	if actualTime != expectedTime {
		t.Errorf("calcNextServiceTime(%s) = %v, expected %v", everyInterval, actualTime, expectedTime)
	}
}

func TestEvery_Thursday_at_8pm(t *testing.T) {
	everyInterval := "every Thursday at 8pm"
	expectedTime := time.Date(2023, time.October, 26, 20, 0, 0, 0, time.Local).Unix() // Same day, 8 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_Monday_at_10am(t *testing.T) {
	everyInterval := "every Monday at 10am"
	expectedTime := time.Date(2023, time.October, 30, 10, 0, 0, 0, time.Local).Unix() // Next Monday, 10 AM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_14th_at_3pm(t *testing.T) {
	everyInterval := "every 14th at 3pm"
	expectedTime := time.Date(2023, time.November, 14, 15, 0, 0, 0, time.Local).Unix() // Next 14th, 3 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_March_20th_at_9pm(t *testing.T) {
	everyInterval := "every March 20th at 9pm"
	expectedTime := time.Date(2024, time.March, 20, 21, 0, 0, 0, time.Local).Unix() // Next March 20th, 9 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestInvalid_format(t *testing.T) {
	everyInterval := "invalid format"
	expectedTime := int64(0)
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_thu_at_8p(t *testing.T) {
	everyInterval := "every thu at 8p"
	expectedTime := time.Date(2023, time.October, 26, 20, 0, 0, 0, time.Local).Unix() // Same day, 8 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_mon_at_10a(t *testing.T) {
	everyInterval := "every mon at 10a"
	expectedTime := time.Date(2023, time.October, 30, 10, 0, 0, 0, time.Local).Unix() // Next Monday, 10 AM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_14_at_3p(t *testing.T) {
	everyInterval := "every 14 at 3p"
	expectedTime := time.Date(2023, time.November, 14, 15, 0, 0, 0, time.Local).Unix() // Next 14th, 3 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_march_20_at_9p(t *testing.T) {
	everyInterval := "every march 20 at 9p"
	expectedTime := time.Date(2024, time.March, 20, 21, 0, 0, 0, time.Local).Unix() // Next March 20th, 9 PM
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_31st_at_12pm(t *testing.T) {
	everyInterval := "every 31st at 12pm"
	expectedTime := time.Date(2023, time.October, 31, 12, 0, 0, 0, time.Local).Unix()
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}

func TestEvery_30th_at_12pm(t *testing.T) {
	everyInterval := "every 30th at 12pm"
	expectedTime := time.Date(2023, time.October, 30, 12, 0, 0, 0, time.Local).Unix()
	testCalcNextServiceTimeHelper(t, everyInterval, expectedTime)
}
