package main

import (
	"log"
	"net"
	"regexp"
	"strings"
)

var ( // host specific regexes
	hostNumbersRegex = regexp.MustCompile(`(\.|\-)[0-9]+(\.|\-)`)
	hostLangLocRegex = regexp.MustCompile(`(^|\.)[a-z]{2}-[a-z]{2}(\.|$)`)
	hostLangRegex    = regexp.MustCompile(`(^|\.|\-)(at|ar|au|bd|be|br|ca|ch|cn|de|dk|es|fi|fr|gr|id|in|it|jp|kz|mx|my|nl|no|ph|pl|pt|py|ru|se|sg|tr|uk|us|uy|vn)(\.|\-|:|$)`)
)

func StandardizeHost(host string, useSeeds bool, seeds []string) string {
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
	host = hostLangLocRegex.ReplaceAllString(host, ".xx-xx.")
	host = hostLangLocRegex.ReplaceAllString(host, ".xx-xx.")
	host = hostLangRegex.ReplaceAllString(host, ".xx.")
	host = hostLangRegex.ReplaceAllString(host, ".xx.")

	// treat any numbers as equals (if host != IP)
	if letterRegex.MatchString(host) {
		host = hostNumbersRegex.ReplaceAllString(host, "0") // `(\.|\-)[0-9]+(\.|\-)`
		host = hostNumbersRegex.ReplaceAllString(host, "0") // repeated to catch, for example 1-2-3-4, without lookarounds
	}

	return host
}
