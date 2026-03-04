package utils

import (
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

func GenerateID() int64 {
	return GenerateSnowflakeID()
}

func GenerateJoinCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// CurrentTimestamp returns the current Unix timestamp
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

func IfNil[T any](val *T, defaultVal T) *T {
	if val == nil {
		return &defaultVal
	}
	return val
}

func StripUnsafeFields[T any](ptr *T) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("StripUnsafeFields requires pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("safe") != "true" {
			fv := v.Field(i)
			if fv.CanSet() {
				fv.Set(reflect.Zero(field.Type))
			}
		}
	}
}

// pluralize returns the singular or plural form based on n
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// ExpiryToString converts an expiry in seconds to a human-readable duration string (e.g., '1 hour', '2 hours', '15 minutes', '1 day').
func ExpiryToString(seconds int) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes())
	days := hours / 24

	switch {
	case d < time.Minute:
		return "less than a minute"
	case d < time.Hour:
		min := minutes
		return strconv.Itoa(min) + " " + pluralize(min, "minute", "minutes")
	case d < 24*time.Hour:
		h := hours
		return strconv.Itoa(h) + " " + pluralize(h, "hour", "hours")
	default:
		return strconv.Itoa(days) + " " + pluralize(days, "day", "days")
	}
}

// IfElse returns trueVal if condition is true, else falseVal
func IfElse[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
