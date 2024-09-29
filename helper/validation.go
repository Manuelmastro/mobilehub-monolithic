package helper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
)

func validateNameOrInitials(fl validator.FieldLevel) bool {
	fullNameRegex := `^[A-Za-z]+(\s[A-Za-z]+)*$`         // Matches full names like "John Doe"
	initialsRegex := `^[A-Za-z]{1,2}(\.[A-Za-z]{1,2})*$` // Matches initials like "J.D." or "A.B."

	value := fl.Field().String()

	// Compile the regexes
	fullNamePattern := regexp.MustCompile(fullNameRegex)
	initialsPattern := regexp.MustCompile(initialsRegex)

	// Match against either full name or initials
	return fullNamePattern.MatchString(value) || initialsPattern.MatchString(value)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Regex patterns
	hasUpperCase := regexp.MustCompile(`[A-Z]`).MatchString
	hasLowerCase := regexp.MustCompile(`[a-z]`).MatchString
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString
	hasSpecialChar := regexp.MustCompile(`[!@#~$%^&*(),.?":{}|<>]`).MatchString

	// Check if all conditions are met
	return hasUpperCase(password) && hasLowerCase(password) && hasNumber(password) && hasSpecialChar(password)
}

// Function to check if there are no leading or trailing spaces
func noLeadingOrTrailingSpaces(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return !strings.HasPrefix(name, " ") && !strings.HasSuffix(name, " ")
}

// Function to check for repeating spaces
func noRepeatingSpaces(fl validator.FieldLevel) bool {

	name := fl.Field().String()
	repeatingSpacesRegex := regexp.MustCompile(`\s{2,}`)
	return !repeatingSpacesRegex.MatchString(name)
}

func Validate(value interface{}) error {

	validate := validator.New()
	validate.RegisterValidation("nameOrInitials", validateNameOrInitials)
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("no_leading_trailing_spaces", noLeadingOrTrailingSpaces)
	validate.RegisterValidation("no_repeating_spaces", noRepeatingSpaces)
	err := validate.Struct(value)
	if err != nil {

		for _, e := range err.(validator.ValidationErrors) {

			switch e.Tag() {
			case "required":
				return fmt.Errorf("%s is required", e.Field())
			case "email":
				return fmt.Errorf("%s is not a valid email address", e.Field())
			case "numeric":
				return fmt.Errorf("%s shouls contain only digits", e.Field())
			case "len":
				return fmt.Errorf("%s shouls have a length of %s", e.Field(), e.Param())
			case "min":
				return fmt.Errorf("%s shouls have a minimum length of %s", e.Field(), e.Param())
			case "excludesall":
				return fmt.Errorf("%s shouls not contain space", e.Field())
			case "nameOrInitials":
				return fmt.Errorf("%s should be either initials or a regular name", e.Field())
			case "password":
				return fmt.Errorf("%s should contain at least one uppercase letter, one lowercase letter, one digit, and one special character", e.Field())
			case "no_leading_trailing_spaces":
				return fmt.Errorf("%s should not have leading or trailing spaces", e.Field())
			case "no_repeating_spaces":
				return fmt.Errorf("%s  should not have repeating spaces", e.Field())
			case "max":
				return fmt.Errorf("%s exceeds the maximum length", e.Field())
			case "alpha":
				return fmt.Errorf("%s should contain only alphabetic characters", e.Field())
			case "gt":
				return fmt.Errorf("%s must be greater than zero", e.Field())
			default:
				return fmt.Errorf("validation error for field %s", e.Field())
			}
		}
	}
	//return errors.New(strings.Join(errs, ", "))
	return nil
}
