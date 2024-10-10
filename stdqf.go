package main

import (
	"regexp"
	"sort"
	"strings"
)

var (
	queryLang5Regex = regexp.MustCompile(`=[a-z]{2}[-/][a-z]{2}&`)
	queryLangRegex  = regexp.MustCompile(`=[a-z]{2}&`)
	ampDivider      = "&"
)

func StandardizeQueryAndFragment(qf string, uniqKeys bool) string {
	qf = strings.ToLower(qf)

	qfSlice := strings.Split(qf, ampDivider)
	sort.Strings(qfSlice)
	qf = strings.Join(qfSlice, ampDivider)
	qf = strings.Trim(qf, ampDivider) + "&" // appending & for regexes

	if uniqKeys {
		qf = regexp.MustCompile(`=.*?&`).ReplaceAllString(qf, "=1&") // treat any values as equals
	}

	qf = uuidRegex.ReplaceAllString(qf, "uuid")          // treat any uuid as equals
	qf = numbersRegex.ReplaceAllString(qf, "0")          // treat any numbers as equals
	qf = queryLang5Regex.ReplaceAllString(qf, "=xx-xx&") // treat xx-xx language notation as equals
	qf = queryLangRegex.ReplaceAllString(qf, "=xx&")     // treat xx language notation as equals

	return qf
}
