package utils

import (
	"time"
)

var KazakhstanTZ = time.FixedZone("Asia/Almaty", 5*60*60)

func GetCurrentTimeUTC() time.Time {
	return time.Now().UTC()
}

func ConvertInputToUTC(inputTime time.Time) time.Time {
	if inputTime.Location() == time.UTC || inputTime.Location().String() == "UTC" {
		localTime := time.Date(
			inputTime.Year(), inputTime.Month(), inputTime.Day(),
			inputTime.Hour(), inputTime.Minute(), inputTime.Second(),
			inputTime.Nanosecond(), KazakhstanTZ,
		)
		return localTime.UTC()
	}

	return inputTime.UTC()
}

func ConvertOutputFromUTC(utcTime time.Time) time.Time {
	return utcTime.In(KazakhstanTZ)
}

func ParseUserInput(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	var parsedTime time.Time
	var err error

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return time.Time{}, err
	}

	return ConvertInputToUTC(parsedTime), nil
}

func FormatForUser(utcTime time.Time) string {
	localTime := ConvertOutputFromUTC(utcTime)
	return localTime.Format("2006-01-02T15:04:05")
}

func FormatForUserPtr(utcTime *time.Time) *string {
	if utcTime == nil {
		return nil
	}
	formatted := FormatForUser(*utcTime)
	return &formatted
}

func IsTimeInRange(checkTime, startTime, endTime time.Time, gracePeriod time.Duration) bool {
	return checkTime.After(startTime) && checkTime.Before(endTime.Add(gracePeriod))
}

const DefaultGracePeriod = 30 * time.Minute

const (
	DaytimeStart = 10
	DaytimeEnd   = 22
)

const (
	RentalDuration3Hours  = 3
	RentalDuration6Hours  = 6
	RentalDuration12Hours = 12
	RentalDuration24Hours = 24
)

const (
	Discount6Hours  = 60
	Discount12Hours = 70
)

func IsDaytimeHour(t time.Time) bool {
	localTime := t.In(KazakhstanTZ)
	hour := localTime.Hour()
	return hour >= DaytimeStart && hour < DaytimeEnd
}

func GetAvailableRentalDurations(startTime time.Time, supportsHourly, supportsDaily bool) []int {
	var durations []int

	if supportsDaily {
		localTime := startTime.In(KazakhstanTZ)
		now := GetCurrentTimeUTC().In(KazakhstanTZ)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, KazakhstanTZ)
		selectedDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, KazakhstanTZ)

		if selectedDay.After(today) || (selectedDay.Equal(today) && now.Hour() < 23) {
			durations = append(durations, 24)
		}
	}

	if supportsHourly {
		localTime := startTime.In(KazakhstanTZ)
		now := GetCurrentTimeUTC().In(KazakhstanTZ)

		var effectiveStartTime time.Time
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, KazakhstanTZ)
		selectedDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, KazakhstanTZ)

		if selectedDay.Equal(today) {
			if now.After(localTime) {
				effectiveStartTime = now
			} else {
				effectiveStartTime = localTime
			}
		} else {
			effectiveStartTime = localTime
		}

		startHour := effectiveStartTime.Hour()

		if startHour >= DaytimeStart && startHour < DaytimeEnd {
			possibleDurations := []int{RentalDuration3Hours, RentalDuration6Hours, RentalDuration12Hours}

			for _, duration := range possibleDurations {
				endTime := effectiveStartTime.Add(time.Duration(duration) * time.Hour)
				endHour := endTime.Hour()
				endMinute := endTime.Minute()

				if endHour < DaytimeEnd || (endHour == DaytimeEnd && endMinute == 0) {
					durations = append(durations, duration)
				}
			}
		}
	}

	return durations
}

func ValidateRentalTime(startTime time.Time, duration int, supportsHourly, supportsDaily bool) bool {
	availableDurations := GetAvailableRentalDurations(startTime, supportsHourly, supportsDaily)

	for _, availableDuration := range availableDurations {
		if duration == availableDuration {
			return true
		}
	}

	return false
}

func GetRentalTimeInfo(startTime time.Time) map[string]interface{} {
	localTime := startTime.In(KazakhstanTZ)
	isDaytime := IsDaytimeHour(startTime)

	return map[string]interface{}{
		"local_time":       localTime.Format("2006-01-02 15:04:05"),
		"is_daytime":       isDaytime,
		"hourly_available": isDaytime,
		"daily_available":  true,
		"time_restriction": "Почасовая аренда доступна только с 10:00 до 22:00 (местное время)",
	}
}

func CalculateHourlyPrice(hourlyPrice, duration int) int {
	basePrice := hourlyPrice * duration

	switch duration {
	case RentalDuration6Hours:
		return basePrice * (100 - Discount6Hours) / 100
	case RentalDuration12Hours:
		return basePrice * (100 - Discount12Hours) / 100
	default:
		return basePrice
	}
}
