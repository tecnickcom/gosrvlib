package validator

import (
	"context"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	vt "github.com/go-playground/validator/v10"
)

const (
	regexPatternE164NoPlus = `^[1-9][0-9]{7,14}$`
	regexPatternEIN        = `^[0-9]{2}-?[0-9]{7}$`
	regexPatternUSZIPCode  = `^[0-9]{5}(?:-[0-9]{4})?$`
)

var (
	regexE164NoPlus = regexp.MustCompile(regexPatternE164NoPlus)
	regexEIN        = regexp.MustCompile(regexPatternEIN)
	regexUSZIPCode  = regexp.MustCompile(regexPatternUSZIPCode)
)

// CustomValidationTags returns a map of custom tags with validation function.
func CustomValidationTags() map[string]vt.FuncCtx {
	return map[string]vt.FuncCtx{
		"falseif":                  isFalseIf,
		"e164noplus":               isE164NoPlus,
		"ein":                      isEIN,
		"zipcode":                  isUSZIPCode,
		"usstate":                  isUSState,
		"usterritory":              isUSTerritory,
		"datetime_rfc3339":         isDatetimeRFC3339,
		"datetime_rfc3339_relaxed": isDatetimeRFC3339Relaxed,
	}
}

// isE164 checks if the fields value is a valid E.164 phone number format without the leading '+' (e.g.: 123456789012345).
func isE164NoPlus(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()
	return regexE164NoPlus.MatchString(field.String())
}

// isEIN checks if the fields value is a valid EIN US tax code (e.g.: 12-3456789 or 123456789).
func isEIN(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()
	return regexEIN.MatchString(field.String())
}

// isUSZIPCode checks if the fields value is a valid US ZIP code (e.g.: 12345 or 12345-6789).
func isUSZIPCode(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()
	return regexUSZIPCode.MatchString(field.String())
}

// isUSState checks if the fields value is a valid 2-letter US state.
// NOTE: It includes the District of Columbia (DC).
func isUSState(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() == reflect.String {
		switch field.String() {
		case "AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "DC", "FL", "GA", "HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD", "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ", "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC", "SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY":
			return true
		}
	}

	return false
}

// isUSTerritory checks if the fields value is a valid 2-letter US territory (other than the official states and federal district).
func isUSTerritory(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() == reflect.String {
		switch field.String() {
		case "AS", "GU", "MP", "PR", "VI":
			return true
		}
	}

	return false
}

// isFalseIf is a special tag to be used in "OR" combination with another tag.
// It returns false if the specified parameter exist and has the specified value.
// This tag should never be used alone.
// The combined tag will be checked only if this validator returns false.
// Examples:
//
//	"falseif=Country US|usstate" checks if the field is a valid US state only if the Country field is set to "US".
//	"falseif=Country|usstate" checks if the field is a valid US state only if the Country field is set and not empty.
func isFalseIf(_ context.Context, fl vt.FieldLevel) bool {
	param := strings.TrimSpace(fl.Param())
	if param == "" {
		return true
	}

	params := strings.SplitN(param, " ", 3)
	paramField, paramKind, nullable, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), params[0])

	if !found {
		// the field in the param do not exist
		return true
	}

	if len(params) == 1 {
		return hasDefaultValue(paramField, paramKind, nullable)
	}

	return hasNotValue(paramField, paramKind, params[1])
}

// hasDefaultvalue returns true if the field has a default value (nil/zero) or if is unset/invalid.
//

func hasDefaultValue(value reflect.Value, kind reflect.Kind, nullable bool) bool {
	switch kind { //nolint:exhaustive
	case reflect.Invalid:
		return true
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return value.IsNil()
	}

	return (nullable && value.Interface() == nil) || !value.IsValid() || (value.Interface() == reflect.Zero(value.Type()).Interface())
}

// hasNotValue returns true if the field has not the specified value.
//
//nolint:gocyclo,cyclop,exhaustive
func hasNotValue(value reflect.Value, kind reflect.Kind, paramValue string) bool {
	switch kind {
	case reflect.String:
		return value.String() != paramValue
	case reflect.Slice, reflect.Map, reflect.Array:
		p, err := strconv.ParseInt(paramValue, 0, 64)
		return err != nil || int64(value.Len()) != p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, err := strconv.ParseInt(paramValue, 0, 64)
		return err != nil || value.Int() != p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p, err := strconv.ParseUint(paramValue, 0, 64)
		return err != nil || value.Uint() != p
	case reflect.Float32, reflect.Float64:
		p, err := strconv.ParseFloat(paramValue, 64)
		return err != nil || value.Float() != p
	case reflect.Bool:
		p, err := strconv.ParseBool(paramValue)
		return err != nil || value.Bool() != p
	}

	return true
}

// isDatetimeRFC3339 checks if the fields value is a valid RFC3339 date format (e.g.: 2023-10-01T12:00:00Z).
func isDatetimeRFC3339(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.String {
		_, err := time.Parse(time.RFC3339, field.String())
		if err == nil {
			return true
		}
	}

	return false
}

// isRFC3339DatetimeRelaxed checks if the fields value is a
// valid RFC3339 date format or a relaxed format "2006-01-02 15:04:05".
func isDatetimeRFC3339Relaxed(_ context.Context, fl vt.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.String {
		_, err := time.Parse("2006-01-02 15:04:05", field.String())
		if err != nil {
			_, err = time.Parse(time.RFC3339, field.String())
		}

		if err == nil {
			return true
		}
	}

	return false
}
