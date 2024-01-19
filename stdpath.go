package main

import (
	"path"
	"regexp"
	"strings"
)

var ( // path specific regexes
	pathLangLocRegex     = regexp.MustCompile(`/[a-z]{2}[-/_][a-z]{2}/`)
	pathLangRegex        = regexp.MustCompile(`/[a-z]{2}/`)
	pathLangHtmlExtRegex = regexp.MustCompile(`/[a-z]{2}\.html/`)
	pathHarshRegex       = regexp.MustCompile(`/([a-z-_]{3}[a-z-_]+)/.`)
)

func StandardizePath(p string, harsh bool) string {
	p = strings.ToLower(p)

	p = path.Clean(p) + "/"                                   // treat with/without trailing slash as equals (appending slash for regexes)
	p = uuidRegex.ReplaceAllString(p, "uuid")                 // treat any uuid as equals
	p = hexRegex.ReplaceAllString(p, "0a9f")                  // treat most hexhashes as equals
	p = numbersRegex.ReplaceAllString(p, "0")                 // treat any numbers as equals
	p = pathLangLocRegex.ReplaceAllString(p, "/xx-xx/")       // treat xx-xx language notation as equals
	p = pathLangRegex.ReplaceAllString(p, "/xx/")             // treat xx language notation as equals
	p = pathLangHtmlExtRegex.ReplaceAllString(p, "/xx.html/") // treat xx language notation as equals

	if harsh {
		// handle 3letters+hyphon/underscore/letter(s)-slugs as equals
		// e.g. treat both /author/ellen & /author/oliver as /author/x
		// treat both /page/qwe as /page/x & /page/qwe/qwe as /page/x/x
		// treat both /jobs/qwe as /jobs/x & /jobs/qwe.html as /jobs/x.x

		// additonal setup for common dupz
		p = strings.Replace(p, "/job/", "/xjobx/", 1)
		p = strings.Replace(p, "/tag/", "/xtagx/", 1)
		p = strings.Replace(p, "/doc/", "/xdocx/", 1)
		p = strings.Replace(p, "/faq/", "/xfaqx/", 1)
		p = strings.Replace(p, "/api/", "/xapix/", 1) // careful
		p = strings.Replace(p, "/c/", "/xxcxx/", 1)   // careful
		p = strings.Replace(p, "/0/0/0/", "/xzerox/", 1)

		if !strings.Contains(p[1:len(p)-1], "/") {
			p = strings.ReplaceAll(p, "-", "/")
			if !strings.Contains(p, "-") {
				p = strings.ReplaceAll(p, "_", "/")
			}
		}

		// harsh dedup
		matches := pathHarshRegex.FindStringSubmatch(p) // `/([a-z-_]{3}[a-z-_]+)/.`
		if len(matches) > 0 {
			pBase := matches[1]
			pArr := strings.SplitN(p, pBase, 2)
			p = pArr[0] + pBase + notSlashRegex.ReplaceAllString(pArr[1], "x")
		}
	}

	return p
}
