package main

import (
	"log"
	"net"
	"regexp"
	"strings"
)

var ( // host specific regexes
	hostNumbersRegex      = regexp.MustCompile(`(\.|\-)[0-9]+(\.|\-)`)
	hostLang5Regex        = regexp.MustCompile(`(^|\.)[a-z]{2}-[a-z]{2}(\.|$)`)
	hostSome2LettersRegex = regexp.MustCompile(`(^|\.|\-)(ar|be|br|cn|de|dk|es|fr|ja|nl|ph|pl|pt|ru|sg|tr|vn|za|zh)(\.|\-|:|$)`)
	hostMany2LettersRegex = regexp.MustCompile(`(^|\.|\-)(ae|al|ar|at|au|be|bg|br|ca|cl|cn|cz|de|dk|es|eu|fr|gr|hk|hu|id|ie|il|in|ja|jp|ke|ko|kz|lk|lt|lv|my|ng|ni|nl|nz|pa|ph|pk|pl|pt|ro|ru|sa|sg|sk|th|tr|ts|tw|ug|uk|uy|vn|za|zh)(\.|\-|:|$)`) // ccTLD && ISO 639-1 (filtered)
)

func StandardizeHost(host string, useSeeds bool, seeds []string, harsh bool) string {
	host = strings.ToLower(host)

	if useSeeds {
		for _, seed := range seeds {
			if strings.Contains(seed, "/") {
				// seed is cidr
				_, ipNet, err := net.ParseCIDR(seed)
				if err != nil {
					log.Fatal(err)
				}

				ip := net.ParseIP(host)
				if ip != nil && ipNet.Contains(ip) {
					host = seed
				}
			} else if strings.HasSuffix(host, seed[1:]) {
				host = seed
				break
			}
		}
	}

	// treat common language notation as equals
	host = hostLang5Regex.ReplaceAllString(host, ".xx-xx.")
	if harsh {
		host = hostMany2LettersRegex.ReplaceAllString(host, ".xx.")
	} else {
		host = hostSome2LettersRegex.ReplaceAllString(host, ".xx.")
	}

	// treat any numbers as equals (if host != IP)
	if letterRegex.MatchString(host) {
		host = hostNumbersRegex.ReplaceAllString(host, "0") // `(\.|\-)[0-9]+(\.|\-)`
		host = hostNumbersRegex.ReplaceAllString(host, "0") // repeated to catch, for example 1-2-3-4, without lookarounds
	}

	return host
}
