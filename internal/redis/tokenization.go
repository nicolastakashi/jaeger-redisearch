package redis

import "strings"

const (
	field_tokenization = ",.<>{}[]\"':;!@#$%^&*()-+=~"
)

func Tokenization(value string) string {
	for _, char := range field_tokenization {
		value = strings.Replace(value, string(char), ("\\" + string(char)), -1)
	}
	return value
}

func UnTokenization(value string) string {
	for _, char := range field_tokenization {
		value = strings.Replace(value, ("\\" + string(char)), string(char), -1)
	}
	return value
}
