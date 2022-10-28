package redis

import "strings"

var tokenizationReplacer = strings.NewReplacer("-", "\\-", ".", "\\.", ":", "\\:", ";", "\\;")
var unTokenizationReplacer = strings.NewReplacer("\\-", "-", "\\.", ".", "\\:", ":", "\\;", ";")

func Tokenization(s string) string {
	return tokenizationReplacer.Replace(UnTokenization(s))
}

func UnTokenization(s string) string {
	return unTokenizationReplacer.Replace(s)
}
