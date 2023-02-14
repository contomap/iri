// Package iri contains facilities for working with Internationalized Resource
// Identifiers as specified in RFC 3987.
//
// RFC reference: https://www.ietf.org/rfc/rfc3987.html
package iri

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// An IRI (Internationalized Resource Identifier) within an RDF graph is a
// Unicode string [UNICODE] that conforms to the syntax defined in RFC 3987
// [RFC3987].
//
// See https://www.w3.org/TR/2014/REC-rdf11-concepts-20140225/#dfn-iri.
type IRI struct {
	Scheme        string
	EmptyAuth     bool // true if the iri is something like `///path`
	ForceUserInfo bool // append an at ('@') even if UserInfo is empty
	UserInfo      string
	Host          string // host including port information
	Path          string
	ForceQuery    bool // append a query ('?') even if Query is empty
	Query         string
	ForceFragment bool // append a fragment ('#') even if Fragment field is empty
	Fragment      string
}

// Parse parses a string into an IRI and checks that it conforms to RFC 3987.
func Parse(s string) (IRI, error) {
	match := uriRE.FindStringSubmatch(s)
	if len(match) == 0 {
		return IRI{}, fmt.Errorf("%q is not a valid IRI - does not match regexp %s", s, uriRE)
	}
	scheme := match[uriRESchemeGroup]
	auth := match[uriREAuthorityGroup]
	path := match[uriREPathGroup]
	query := match[uriREQueryGroup]
	fragment := match[uriREFragmentGroup]
	if scheme != "" && !schemeRE.MatchString(scheme) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid scheme %q does not match regexp %s", s, scheme, schemeRE)
	}
	if auth != "" && !iauthorityRE.MatchString(auth) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid auth %q does not match regexp %s", s, auth, iauthorityRE)
	}
	if path != "" && !ipathRE.MatchString(path) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid path %q does not match regexp %s", s, path, ipathRE)
	}
	if query != "" && !iqueryRE.MatchString(query) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid query %q does not match regexp %s", s, query, iqueryRE)
	}
	if fragment != "" && !ifragmentRE.MatchString(fragment) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid fragment %q does not match regexp %s", s, fragment, ifragmentRE)
	}

	authMatch := iauthorityCaptureRE.FindStringSubmatch(auth)
	var userInfo, host string
	var forceUserInfo bool
	if len(authMatch) != 0 {
		forceUserInfo = authMatch[iauthorityUserInfoWithAtGroup] != ""
		userInfo = authMatch[iauthorityUserInfoGroup]
		host = authMatch[iauthorityHostPortGroup]
	}
	parsed := IRI{
		Scheme:        match[uriRESchemeGroup],
		EmptyAuth:     len(match[uriREAuthorityWithSlashSlashGroup]) != 0 && (userInfo == "" && host == ""),
		ForceUserInfo: forceUserInfo,
		UserInfo:      userInfo,
		Host:          host,
		Path:          match[uriREPathGroup],
		ForceQuery:    match[uriREQueryWithMarkGroup] != "",
		Query:         match[uriREQueryGroup],
		ForceFragment: match[uriREFragmentWithHashGroup] != "",
		Fragment:      match[uriREFragmentGroup],
	}

	if _, err := parsed.normalizePercentEncoding(); err != nil {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid percent encoding: %w", s, err)
	}

	return parsed, nil
}

// Check returns an error if the IRI is invalid.
func (iri IRI) Check() error {
	_, err := Parse(iri.String())
	return err
}

// String reassembles the IRI into a valid IRI string.
func (iri IRI) String() string {
	s := ""
	if iri.Scheme != "" {
		s += iri.Scheme + ":"
	}
	if iri.EmptyAuth || iri.UserInfo != "" || iri.Host != "" {
		s += "//"
	}
	if iri.ForceUserInfo || (iri.UserInfo != "") {
		s += iri.UserInfo + "@"
	}
	if iri.Host != "" {
		s += iri.Host
	}
	if iri.Path != "" {
		s += iri.Path
	}
	if iri.ForceQuery || (iri.Query != "") {
		s += "?" + iri.Query
	}
	if iri.ForceFragment || (iri.Fragment != "") {
		s += "#" + iri.Fragment
	}
	return s
}

// ResolveReference resolves an IRI reference to an absolute IRI from an absolute
// base IRI u, per RFC 3986 Section 5.2. The IRI reference may be relative or
// absolute.
func (iri IRI) ResolveReference(other IRI) IRI {
	return resolveReference(iri, other)
}

// Normalization background reading:
// - https://blog.golang.org/normalization
// - https://www.ietf.org/rfc/rfc3987.html#section-5
//    - https://www.ietf.org/rfc/rfc3987.html#section-5.3.2.3 - percent encoding

// Regular expression const strings, mostly derived from
// https://www.ietf.org/rfc/rfc3987.html#section-2.2.
const (
	hex        = `[0-9A-Fa-f]`
	alphaChars = "[a-zA-Z]" // see https://tools.ietf.org/html/rfc5234 B.1. "ALPHA"
	digitChars = `\d`       // see https://tools.ietf.org/html/rfc5234 B.1. "DIGIT"
	ucschar    = `[\xA0-\x{D7FF}` +
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
		`\x{E1000}-\x{EFFFD}]`
	unreserved  = `(?:` + alphaChars + "|" + digitChars + `|[\-\._~]` + `)`
	iunreserved = `(?:` + alphaChars + "|" + digitChars + `|[\-\._~]|` + ucschar + `)`

	subDelims           = `[!\$\&\'\(\)\*\+\,\;\=]`
	pctEncoded          = `%` + hex + hex
	pctEncodedOneOrMore = `(?:(?:` + pctEncoded + `)+)`

	ipchar = "(?:" + iunreserved + "|" + pctEncoded + "|" + subDelims + `|[\:@])`

	scheme = "(?:" + alphaChars + "(?:" + alphaChars + "|" + digitChars + `|[\+\-\.])*)`

	iauthority                    = `(?:` + iuserinfo + "@)?" + ihost + `(?:\:` + port + `)?`
	iauthorityCapture             = `(?:((` + iuserinfo + ")@)?((?:" + ihost + `)(?:\:(?:` + port + `))?))`
	iauthorityUserInfoWithAtGroup = 1
	iauthorityUserInfoGroup       = 2
	iauthorityHostPortGroup       = 3
	iuserinfo                     = `(?:(?:` + iunreserved + `|` + pctEncoded + `|` + subDelims + `|\:)*)`
	port                          = `(?:\d*)`
	ihost                         = `(?:` + ipLiteral + `|` + ipV4Address + `|` + iregName + `)`
	iregName                      = "(?:(?:" + iunreserved + "|" + pctEncoded + "|" + subDelims + ")*)" // *( iunreserved / pctEncoded / subDelims )

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
	// non-zero-length segment without any colon ":"
	isegmentnznc = `(?:` + iunreserved + `|` + pctEncoded + `|` + subDelims + `|` + `[@])`

	iquery = `(?:(?:` + ipchar + `|` + iprivate + `|` + `\/\?` + `)*)`

	ifragment = `(?:(?:` + ipchar + `|` + `[\/\?]` + `)*)`

	iprivate = `[\x{E000}-\x{F8FF}` + `\x{F0000}-\x{FFFFD}` + `\x{100000}-\x{10FFFD}]`
)

// IP Address related
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

	decOctet = `(?:\d` + `|` + // 0-9
		`[1-9]\d` + `|` + // 10-99
		`1\d\d` + `|` + // 100-199
		`2[0-4]\d` + `|` + // 200-249
		`25[0-5]` + `)` // 250-255
)

var (
	schemeRE            = mustCompileNamed("schemeRE", "^"+scheme+"$")
	iauthorityRE        = mustCompileNamed("iauthorityRE", "^"+iauthority+"$")
	iauthorityCaptureRE = mustCompileNamed("iauthorityCaptureRE", "^"+iauthorityCapture+"$")
	ipathRE             = mustCompileNamed("ipath", "^"+ipath+"$")
	iqueryRE            = mustCompileNamed("iquery", "^"+iquery+"$")
	ifragmentRE         = mustCompileNamed("ifragment", "^"+ifragment+"$")

	percentEncodedChar      = mustCompileNamed("percentEncodedChar", pctEncoded)
	pctEncodedCharOneOrMore = mustCompileNamed("pctEncodedOneOrMore", pctEncodedOneOrMore)
	iunreservedRE           = mustCompileNamed("iunreservedRE", "^"+iunreserved+"$")

	hexToByte = func() map[string]byte {
		m := map[string]byte{}
		for i := 0; i <= 255; i++ {
			m[fmt.Sprintf("%02X", i)] = byte(i)
		}
		return m
	}()
	byteToUppercasePercentEncoding = func() map[byte]string {
		m := map[byte]string{}
		for i := 0; i <= 255; i++ {
			m[byte(i)] = fmt.Sprintf("%%%02X", i)
		}
		return m
	}()

	// re from RFC 3986 page 50.
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

// NormalizePercentEncoding returns an IRI that replaces any unnecessarily
// percent-escaped characters with unescaped characters.
//
// RFC3987 discusses this normalization procedure in 5.3.2.3:
// https://www.ietf.org/rfc/rfc3987.html#section-5.3.2.3.
func (iri IRI) NormalizePercentEncoding() IRI {
	normalized, err := iri.normalizePercentEncoding()
	if err != nil {
		return iri
	}
	return normalized
}

func (iri IRI) normalizePercentEncoding() (IRI, error) {
	var errs []error
	replaced := iri
	// TODO (type-rework) - figure out if only the Path component needs replacing (probably not - think of user names; tests are green however)
	// Find consecutive percent-encoded octets and encode them together.
	replaced.Path = pctEncodedCharOneOrMore.ReplaceAllStringFunc(iri.Path, func(pctEscaped string) string {
		normalized := ""
		unconsumedOctets := octetsFrom(pctEscaped)
		octetsOffset := 0
		for len(unconsumedOctets) > 0 {
			codePoint, size := utf8.DecodeRune(unconsumedOctets)
			if codePoint == utf8.RuneError {
				errs = append(errs, fmt.Errorf("percent-encoded sequence %q  contains invalid UTF-8 code point at start of byte sequence %+v", pctEscaped[octetsOffset*3:], unconsumedOctets))
				return pctEscaped
			}

			if iunreservedRE.MatchString(string(codePoint)) {
				normalized += string(codePoint)
			} else {
				buf := make([]byte, 4)
				codePointOctetCount := utf8.EncodeRune(buf, codePoint)
				for i := 0; i < codePointOctetCount; i++ {
					normalized += byteToUppercasePercentEncoding[buf[i]]
				}
			}
			unconsumedOctets = unconsumedOctets[size:]
			octetsOffset += size
		}

		return normalized
	})
	if len(errs) != 0 {
		return IRI{}, errs[0]
	}
	return replaced, nil
}

func octetsFrom(pctEscaped string) []byte {
	octets := make([]byte, len(pctEscaped)/3)
	for i := 0; i < len(octets); i++ {
		start := i * 3
		digitsStr := strings.ToUpper(pctEscaped[start+1 : start+3])
		octet := hexToByte[digitsStr]
		octets[i] = octet
	}
	return octets
}

func mustCompileNamed(name, expr string) *regexp.Regexp {
	c, err := regexp.Compile(expr)
	if err != nil {
		panic(fmt.Errorf("failed to compile regexp %s: %w", name, err))
	}
	return c
}
