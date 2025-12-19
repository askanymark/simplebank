package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

// ValidateString checks if the length of the input string is within the specified minimum and maximum range.
func ValidateString(value string, minLength int, maxLength int) error {
	n := len(value)

	if n < minLength || n > maxLength {
		return fmt.Errorf("length must be between %d and %d", minLength, maxLength)
	}

	return nil
}

// ValidateUsername ensures the username is between 3 and 100 characters and contains only lowercase letters, numbers, and underscores.
func ValidateUsername(value string) error {
	err := ValidateString(value, 3, 100)
	if err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("must contain lowercase letters, numbers and underscores")
	}

	return nil
}

// ValidatePassword ensures the password is between 6 and 100 characters.
func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}

// ValidateEmail ensures the email address is between 6 and 100 characters and is a valid email address.
func ValidateEmail(value string) error {
	if err := ValidateString(value, 6, 100); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

// ValidateFullName ensures the input is between 3 and 100 characters and consists only of letters and spaces.
func ValidateFullName(value string) error {
	err := ValidateString(value, 3, 100)
	if err != nil {
		return err
	}

	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}

	return nil
}

// ValidateEmailId ensures the input is a positive integer.
func ValidateEmailId(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}

	return nil
}

// ValidateSecretCode ensures the input is between 32 and 128 characters.
func ValidateSecretCode(value string) error {
	return ValidateString(value, 32, 128)
}
