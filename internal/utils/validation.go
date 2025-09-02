package utils

import (
	"fmt"
	"strings"
	"time"
)

func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email не может быть пустым")
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("некорректный формат email")
	}
	return nil
}

func ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("номер телефона не может быть пустым")
	}

	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	if !strings.HasPrefix(cleanPhone, "+7") || len(cleanPhone) != 12 {
		return fmt.Errorf("номер телефона должен быть в формате +7XXXXXXXXXX (12 символов)")
	}

	for i, char := range cleanPhone[2:] {
		if char < '0' || char > '9' {
			return fmt.Errorf("номер телефона может содержать только цифры после +7, недопустимый символ на позиции %d", i+3)
		}
	}

	return nil
}

func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s должно быть больше 0", fieldName)
	}
	return nil
}

func ValidatePositiveFloat(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s должно быть больше 0", fieldName)
	}
	return nil
}

func ValidateStringNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s не может быть пустым", fieldName)
	}
	return nil
}

func ValidateTimeRange(startDate, endDate time.Time) error {
	if startDate.After(endDate) || startDate.Equal(endDate) {
		return fmt.Errorf("дата начала должна быть раньше даты окончания")
	}
	return nil
}

func ValidateFutureDate(date time.Time) error {
	if date.Before(time.Now()) {
		return fmt.Errorf("дата не может быть в прошлом")
	}
	return nil
}

func ValidateDateNotTooFar(date time.Time, maxDays int) error {
	maxDate := time.Now().AddDate(0, 0, maxDays)
	if date.After(maxDate) {
		return fmt.Errorf("дата не может быть более чем на %d дней вперед", maxDays)
	}
	return nil
}

func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s должно быть от %d до %d", fieldName, min, max)
	}
	return nil
}

func ValidateRequiredFields(fields map[string]interface{}) error {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil {
			return fmt.Errorf("поле %s обязательно для заполнения", fieldName)
		}

		if str, ok := fieldValue.(string); ok && strings.TrimSpace(str) == "" {
			return fmt.Errorf("поле %s не может быть пустым", fieldName)
		}

		if num, ok := fieldValue.(int); ok && num == 0 {
			return fmt.Errorf("поле %s должно быть больше 0", fieldName)
		}
	}
	return nil
}

func PreprocessResidentialComplex(complex string) string {
	if complex == "" {
		return complex
	}

	complex = strings.TrimSpace(complex)

	prefixes := []string{"ЖК ", "жк ", "Жк ", "жК "}

	for _, prefix := range prefixes {
		if strings.HasPrefix(complex, prefix) {
			return strings.TrimSpace(complex[len(prefix):])
		}
	}

	return complex
}


func PreprocessResidentialComplexPointer(complex string) *string {
	processed := PreprocessResidentialComplex(complex)
	if processed == "" {
		return nil
	}
	return &processed
}
