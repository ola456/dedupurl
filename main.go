package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	harsh          = flag.Bool("harsh", false, "harsher deduplication, for example, treats /blog/first & /blog/another as equals")
	ignoreFragment = flag.Bool("ignore-fragment", false, "treat any fragments as equals")
	ignorePath     = flag.Bool("ignore-path", false, "treat any paths as equals")
	ignoreQuery    = flag.Bool("ignore-query", false, "treat any querystrings as equals")
	maxPerHost     = flag.Int("max-per-host", 0, "only return first X per host")
	test           = flag.Bool("test", false, "print uniques and how many entries they capture")
	uniqKeep       = flag.Int("keep", 1, "keep first X dups")
	uniqKeys       = flag.Bool("uniq-keys", false, "treat any value in key/value pairs equally")
	seedsFilepath  = flag.String("seeds", "", "seeds file (e.g. path/to/seeds.txt) with cidr/s (e.g. 192.0.2.0/24) or sub-wildcard domains (e.g. *.example.org)")
)

var ( // universal regexes
	uuidRegex     = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	hexRegex      = regexp.MustCompile(`([0-9]+[a-f]|[a-f]+[0-9])[a-f0-9]{2,}`)
	numbersRegex  = regexp.MustCompile(`[0-9]+`)
	letterRegex   = regexp.MustCompile(`[a-z]`)
	notSlashRegex = regexp.MustCompile(`[^/.]+`)
)

func main() {
	// now := time.Now()

	flag.Parse()
	useSeeds := len(*seedsFilepath) > 0

	var rawInput io.Reader
	filepath := flag.Arg(0)
	if filepath == "" || filepath == "-" {
		// urls via stdin
		rawInput = os.Stdin
	} else {
		// urls via arg
		r, err := os.Open(filepath)
		if err != nil {
			log.Fatalf("Couldn't load input file, error: %v", err)
		}
		rawInput = r
		defer r.Close()
	}

	var seeds []string
	if useSeeds {
		r, err := os.Open(*seedsFilepath)
		if err != nil {
			log.Fatalf("Couldn't load seeds, error: %v", err)
		}
		scanner := bufio.NewScanner(r)
		defer r.Close()
		for scanner.Scan() {
			seeds = append(seeds, strings.TrimSpace(scanner.Text()))
		}

		// sort based on occourence of .
		// so domain in uniques are as specific as possible
		// verified using fmt.Println(strings.Join(seeds[:], "\n"))
		sort.Slice(seeds, func(i, j int) bool {
			return strings.Count(seeds[i], ".") > strings.Count(seeds[j], ".")
		})
	}

	uniques := make(map[string]int)
	hosts := make(map[string]int)

	scanner := bufio.NewScanner(rawInput)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		u, err := url.Parse(line)
		if err != nil {
			escapedLine := strings.ReplaceAll(line, "%", "%25")
			u, err = url.Parse(escapedLine)
			if err != nil {
				log.Fatal(err)
			}
		}

		host := StandardizeHost(u.Host, useSeeds, seeds, *harsh)

		var path string
		if !*ignorePath {
			path = StandardizePath(u.Path, *harsh)
		}

		var query string
		if !*ignoreQuery && len(u.RawQuery) > 0 {
			query = StandardizeQueryAndFragment(u.RawQuery, *uniqKeys)
		}

		var fragment string
		if !*ignoreFragment && len(u.Fragment) > 0 {
			fragment = StandardizeQueryAndFragment(u.Fragment, *uniqKeys)
		}

		unique := host + path + "?" + query + "#" + fragment
		uniques[unique]++

		if uniques[unique] <= *uniqKeep {
			hosts[host]++
			if *maxPerHost != 0 && *maxPerHost < hosts[host] {
				uniques["hostOverfill for "+host]++
			} else if !*test {
				fmt.Println(line)
			}
		}
	}

	if *test {
		for unique, value := range uniques {
			fmt.Println(strconv.Itoa(value) + " - " + unique)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// fmt.Fprintln(os.Stderr, "time elapse:", time.Since(now))
}
