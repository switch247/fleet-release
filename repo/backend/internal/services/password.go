package services

import (
	"errors"
	"regexp"
)

var (
	ErrWeakPassword = errors.New("password does not meet complexity requirements")
	upperRe         = regexp.MustCompile(`[A-Z]`)
	lowerRe         = regexp.MustCompile(`[a-z]`)
	digitRe         = regexp.MustCompile(`[0-9]`)
	symbolRe        = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func ValidatePasswordComplexity(password string) error {
	if len(password) < 12 || !upperRe.MatchString(password) || !lowerRe.MatchString(password) || !digitRe.MatchString(password) || !symbolRe.MatchString(password) {
		return ErrWeakPassword
	}
	return nil
}
