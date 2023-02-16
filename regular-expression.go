package iri

import (
	"fmt"
	"regexp"
)

// Regular expression const strings, mostly derived from
// https://www.ietf.org/rfc/rfc3987.html#section-2.2.
const (
	hex        = `[0-9A-Fa-f]`
	alphaChars = "[a-zA-Z]" // see https://tools.ietf.org/html/rfc5234 B.1. "ALPHA"
	digitChars = `\d`       // see https://tools.ietf.org/html/rfc5234 B.1. "DIGIT"
	ucschar    = `[` +
		`\xA0-\x{D7FF}` +
		`\x{F900}-\x{FDCF}` +
		`\x{FDF0}-\x{FFEF}` +
		`\x{10000}-\x{1FFFD}` +
		`\x{20000}-\x{2FFFD}` +
		`\x{30000}-\x{3FFFD}` +
		`\x{40000}-\x{4FFFD}` +
		`\x{50000}-\x{5FFFD}` +
		`\x{60000}-\x{6FFFD}` +
		`\x{70000}-\x{7FFFD}` +
		`\x{80000}-\x{8FFFD}` +
		`\x{90000}-\x{9FFFD}` +
		`\x{A0000}-\x{AFFFD}` +
		`\x{B0000}-\x{BFFFD}` +
		`\x{C0000}-\x{CFFFD}` +
		`\x{D0000}-\x{DFFFD}` +
		`\x{E1000}-\x{EFFFD}` +
		`]`
	unreserved  = `(?:` + alphaChars + "|" + digitChars + `|[\-\._~]` + `)`
	iunreserved = `(?:` + alphaChars + "|" + digitChars + `|[\-\._~]|` + ucschar + `)`

	subDelims           = `[!\$\&\'\(\)\*\+\,\;\=]`
	pctEncoded          = `%` + hex + hex
	pctEncodedOneOrMore = `(?:(?:` + pctEncoded + `)+)`

	ipchar = "(?:" + iunreserved + "|" + pctEncoded + "|" + subDelims + `|[\:@])`

	scheme = "(?:" + alphaChars + "(?:" + alphaChars + "|" + digitChars + `|[\+\-\.])*)`

	iauthority = `(?:` + iuserinfo + "@)?" + ihost + `(?:\:` + port + `)?`
	iuserinfo  = `(?:(?:` + iunreserved + `|` + pctEncoded + `|` + subDelims + `|\:)*)`
	port       = `(?:\d*)`
	ihost      = `(?:` + ipLiteral + `|` + ipV4Address + `|` + iregName + `)`
	iregName   = "(?:(?:" + iunreserved + "|" + pctEncoded + "|" + subDelims + ")*)"

	ipath = `(?:` + ipathabempty + // begins with "/" or is empty
		`|` + ipathabsolute + // begins with "/" but not "//"
		`|` + ipathnoscheme + // begins with a non-colon segment
		`|` + ipathrootless + // begins with a segment
		`|` + ipathempty + `)` // zero characters

	ipathabempty  = `(?:(?:\/` + isegment + `)*)`
	ipathabsolute = `(?:\/(?:` + isegmentnz + `(?:\/` + isegment + `)*` + `)?)`
	ipathnoscheme = `(?:` + isegmentnznc + `(?:\/` + isegment + `)*)`
	ipathrootless = `(?:` + isegmentnz + `(?:\/` + isegment + `)*)`
	ipathempty    = `(?:)` // zero characters

	isegment   = `(?:` + ipchar + `*)`
	isegmentnz = `(?:` + ipchar + `+)`
	// Describes a non-zero-length segment without any colon ":".
	isegmentnznc = `(?:` + iunreserved + `|` + pctEncoded + `|` + subDelims + `|` + `[@])`

	iquery = `(?:(?:` + ipchar + `|` + iprivate + `|` + `[\/\?]` + `)*)`

	ifragment = `(?:(?:` + ipchar + `|` + `[\/\?]` + `)*)`

	iprivate = `[` +
		`\x{E000}-\x{F8FF}` +
		`\x{F0000}-\x{FFFFD}` +
		`\x{100000}-\x{10FFFD}` +
		`]`
)

const (
	ipLiteral = `\[(?:` + ipV6Address + `|` + ipVFuture + `)\]`

	ipVFuture = `v` + hex + `\.(?:` + unreserved + `|` + subDelims + `|\:)*`

	// see https://stackoverflow.com/questions/3032593/using-explicitly-numbered-repetition-instead-of-question-mark-star-and-plus
	ipV6Address = `(?:` +
		`(?:(?:` + h16 + `\:){6}` + ls32 + `)` +
		`|(?:\:\:(?:` + h16 + `\:){5}` + ls32 + `)` +
		`|(?:(?:` + h16 + `)?\:\:(?:` + h16 + `\:){4}` + ls32 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,1}` + h16 + `)?\:\:(?:` + h16 + `\:){3}` + ls32 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,2}` + h16 + `)?\:\:(?:` + h16 + `\:){2}` + ls32 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,3}` + h16 + `)?\:\:(?:` + h16 + `\:)` + ls32 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,4}` + h16 + `)?\:\:` + ls32 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,5}` + h16 + `)?\:\:` + h16 + `)` +
		`|(?:(?:(?:` + h16 + `\:){0,6}` + h16 + `)?\:\:)` +
		`)`

	h16         = `(?:` + hex + `{1,4})`
	ls32        = `(?:` + h16 + `\:` + h16 + `|` + ipV4Address + `)`
	ipV4Address = `(?:` + decOctet + `.` + decOctet + `.` + decOctet + `.` + decOctet + `)`

	decOctet = `(?:` +
		`\d` + `|` + // 0-9
		`[1-9]\d` + `|` + // 10-99
		`1\d\d` + `|` + // 100-199
		`2[0-4]\d` + `|` + // 200-249
		`25[0-5]` + // 250-255
		`)`
)

var (
	schemeRE     = mustCompileNamed("schemeRE", "^"+scheme+"$")
	iauthorityRE = mustCompileNamed("iauthorityRE", "^"+iauthority+"$")
	ipathRE      = mustCompileNamed("ipath", "^"+ipath+"$")
	iqueryRE     = mustCompileNamed("iquery", "^"+iquery+"$")
	ifragmentRE  = mustCompileNamed("ifragment", "^"+ifragment+"$")

	pctEncodedCharOneOrMore = mustCompileNamed("pctEncodedOneOrMore", pctEncodedOneOrMore)
	iunreservedRE           = mustCompileNamed("iunreservedRE", "^"+iunreserved+"$")

	// Regular expression from RFC 3986 page 50.
	uriRE                             = mustCompileNamed("uriRE", `^(([^:/?#]+):)?(//([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?`)
	uriRESchemeGroup                  = 2
	uriREAuthorityWithSlashSlashGroup = 3
	uriREAuthorityGroup               = 4
	uriREPathGroup                    = 5
	uriREQueryWithMarkGroup           = 6
	uriREQueryGroup                   = 7
	uriREFragmentGroup                = 9
	uriREFragmentWithHashGroup        = 8
)

func mustCompileNamed(name, expr string) *regexp.Regexp {
	c, err := regexp.Compile(expr)
	if err != nil {
		panic(fmt.Errorf("failed to compile regexp %s: %w", name, err))
	}
	return c
}
