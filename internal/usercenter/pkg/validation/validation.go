package validation

import (
	"regexp"

	"github.com/google/wire"

	"github.com/onexstack/onex/internal/usercenter/store"
	"github.com/onexstack/onex/pkg/api/errno"
)

// Validator is a struct that implements custom validation logic.
type Validator struct {
	// Some complex validation logic may require direct database queries.
	// This is just an example. If validation requires other dependencies
	// like clients, services, resources, etc., they can all be injected here.
	store store.IStore
}

// Use globally precompiled regular expressions to avoid creating and compiling them repeatedly.
var (
	lengthRegex = regexp.MustCompile(`^.{3,20}$`)                                        // Length between 3 and 20 characters
	validRegex  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)                                  // Only letters, numbers, and underscores
	letterRegex = regexp.MustCompile(`[A-Za-z]`)                                         // At least one letter
	numberRegex = regexp.MustCompile(`\d`)                                               // At least one number
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`) // Email format
	phoneRegex  = regexp.MustCompile(`^1[3-9]\d{9}$`)                                    // Chinese phone number
)

// ProviderSet is a Wire provider set that declares dependency injection rules.
var ProviderSet = wire.NewSet(New, wire.Bind(new(any), new(*Validator)))

// New creates a new instance of Validator.
func New(store store.IStore) *Validator {
	return &Validator{store: store}
}

// isValidUsername validates if a username is valid.
func isValidUsername(username string) bool {
	// Validate length
	if !lengthRegex.MatchString(username) {
		return false
	}
	// Validate character legality
	if !validRegex.MatchString(username) {
		return false
	}
	return true
}

// isValidPassword checks whether a password meets complexity requirements.
func isValidPassword(password string) error {
	switch {
	// Check if the new password is empty
	case password == "":
		return errno.ErrorInvalidParameter("password cannot be empty")
	// Check the length requirement of the new password
	case len(password) < 6:
		return errno.ErrorInvalidParameter("password must be at least 6 characters long")
	// Use a regular expression to check if it contains at least one letter
	case !letterRegex.MatchString(password):
		return errno.ErrorInvalidParameter("password must contain at least one letter")
	// Use a regular expression to check if it contains at least one number
	case !numberRegex.MatchString(password):
		return errno.ErrorInvalidParameter("password must contain at least one number")
	}
	return nil
}

// isValidEmail checks whether an email is valid.
func isValidEmail(email string) error {
	// Check if the email is empty
	if email == "" {
		return errno.ErrorInvalidParameter("email cannot be empty")
	}

	// Validate email format using a regular expression
	if !emailRegex.MatchString(email) {
		return errno.ErrorInvalidParameter("invalid email format")
	}

	return nil
}

// isValidPhone checks whether a phone number is valid.
func isValidPhone(phone string) error {
	// Check if the phone number is empty
	if phone == "" {
		return errno.ErrorInvalidParameter("phone cannot be empty")
	}

	// Validate the phone number format (assumed to be a Chinese phone number, 11 digits)
	if !phoneRegex.MatchString(phone) {
		return errno.ErrorInvalidParameter("invalid phone format")
	}

	return nil
}
