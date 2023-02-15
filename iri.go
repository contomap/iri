package iri

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// An IRI (Internationalized Resource Identifier) is a Unicode string [UNICODE]
// that conforms to the syntax defined in RFC 3987.
//
// See https://www.ietf.org/rfc/rfc3987.html
type IRI struct {
	Scheme         string
	ForceAuthority bool // append a double-slash ('//') even if Authority is empty
	Authority      string
	Path           string
	ForceQuery     bool // append a query ('?') even if Query is empty
	Query          string
	ForceFragment  bool // append a fragment ('#') even if Fragment field is empty
	Fragment       string
}

// Parse parses a string into an IRI and checks that it conforms to RFC 3987.
//
// It performs a coarse segmentation based on a regular expression to separate the components,
// and then verifies with detailed regular expressions whether the components are correct.
// Finally, any percent-encoding is verified - yet the returned IRI will have the original percent encoding
// maintained.
// If any of these steps produce an error, this function returns an error and an empty IRI.
func Parse(s string) (IRI, error) {
	match := uriRE.FindStringSubmatch(s)
	if len(match) == 0 {
		return IRI{}, fmt.Errorf("%q is not a valid IRI - does not match regexp %s", s, uriRE)
	}
	scheme := match[uriRESchemeGroup]
	authority := match[uriREAuthorityGroup]
	path := match[uriREPathGroup]
	query := match[uriREQueryGroup]
	fragment := match[uriREFragmentGroup]
	if scheme != "" && !schemeRE.MatchString(scheme) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid scheme %q does not match regexp %s", s, scheme, schemeRE)
	}
	if authority != "" && !iauthorityRE.MatchString(authority) {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid authority %q does not match regexp %s", s, authority, iauthorityRE)
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

	parsed := IRI{
		Scheme:         scheme,
		ForceAuthority: len(match[uriREAuthorityWithSlashSlashGroup]) != 0,
		Authority:      authority,
		Path:           path,
		ForceQuery:     match[uriREQueryWithMarkGroup] != "",
		Query:          query,
		ForceFragment:  match[uriREFragmentWithHashGroup] != "",
		Fragment:       fragment,
	}

	if _, err := NormalizePercentEncoding(parsed); err != nil {
		return IRI{}, fmt.Errorf("%q is not a valid IRI: invalid percent encoding: %w", s, err)
	}

	return parsed, nil
}

// String reassembles the IRI into an IRI string.
// Any components that have been manually set must comply to the format;
// This function performs no further escaping.
func (iri IRI) String() string {
	var result strings.Builder
	if iri.hasScheme() {
		result.WriteString(iri.Scheme)
		result.WriteRune(':')
	}
	if iri.hasAuthority() {
		result.WriteString("//")
		result.WriteString(iri.Authority)
	}
	result.WriteString(iri.Path)
	if iri.hasQuery() {
		result.WriteRune('?')
		result.WriteString(iri.Query)
	}
	if iri.hasFragment() {
		result.WriteRune('#')
		result.WriteString(iri.Fragment)
	}
	return result.String()
}

func (iri IRI) hasScheme() bool    { return iri.Scheme != "" }
func (iri IRI) hasAuthority() bool { return iri.ForceAuthority || iri.Authority != "" }
func (iri IRI) hasQuery() bool     { return iri.ForceQuery || iri.Query != "" }
func (iri IRI) hasFragment() bool  { return iri.ForceFragment || iri.Fragment != "" }

// ResolveReference resolves an IRI reference to an absolute IRI from an absolute
// base IRI u, per RFC 3986 Section 5.2. The IRI reference may be relative or absolute.
func (iri IRI) ResolveReference(other IRI) IRI {
	return resolveReference(iri, other)
}

// NormalizePercentEncoding returns an IRI that replaces any unnecessarily
// percent-escaped characters with unescaped characters.
//
// RFC3987 discusses this normalization procedure in 5.3.2.3:
// https://www.ietf.org/rfc/rfc3987.html#section-5.3.2.3.
func NormalizePercentEncoding(iri IRI) (IRI, error) {
	replaced := iri
	var err error
	replaced.Authority, err = normalizePercentEncoding(iri.Authority)
	if err != nil {
		return IRI{}, err
	}
	replaced.Path, err = normalizePercentEncoding(iri.Path)
	if err != nil {
		return IRI{}, err
	}
	replaced.Query, err = normalizePercentEncoding(iri.Query)
	if err != nil {
		return IRI{}, err
	}
	replaced.Fragment, err = normalizePercentEncoding(iri.Fragment)
	if err != nil {
		return IRI{}, err
	}
	return replaced, nil
}

// normalizePercentEncoding replaces unreserved percent-encoded characters with their equivalent.
//
// Normalization background reading:
// - https://blog.golang.org/normalization
// - https://www.ietf.org/rfc/rfc3987.html#section-5
//   - https://www.ietf.org/rfc/rfc3987.html#section-5.3.2.3 - percent encoding
func normalizePercentEncoding(in string) (string, error) {
	var errs []error
	replaced := pctEncodedCharOneOrMore.ReplaceAllStringFunc(in, func(pctEscaped string) string {
		normalized := ""
		unconsumedOctets := octetsFrom(pctEscaped)
		octetsOffset := 0
		for len(unconsumedOctets) > 0 {
			codePoint, size := utf8.DecodeRune(unconsumedOctets)
			if codePoint == utf8.RuneError {
				errs = append(errs, fmt.Errorf("percent-encoded sequence %q contains invalid UTF-8 code point at start", pctEscaped[octetsOffset*3:]))
				return pctEscaped
			}
			normalized += toUnreservedString(codePoint)
			unconsumedOctets = unconsumedOctets[size:]
			octetsOffset += size
		}
		return normalized
	})
	if len(errs) != 0 {
		return "", errs[0]
	}
	return replaced, nil
}

var (
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
)

func octetsFrom(percentEncoded string) []byte {
	octets := make([]byte, len(percentEncoded)/3)
	for i := 0; i < len(octets); i++ {
		start := i * 3
		digitsStr := strings.ToUpper(percentEncoded[start+1 : start+3])
		octet := hexToByte[digitsStr]
		octets[i] = octet
	}
	return octets
}

func toUnreservedString(r rune) string {
	isUnreserved := iunreservedRE.MatchString(string(r))
	if isUnreserved {
		return string(r)
	}
	var percentEncoded string
	var buf [utf8.UTFMax]byte
	octetCount := utf8.EncodeRune(buf[:], r)
	for i := 0; i < octetCount; i++ {
		percentEncoded += byteToUppercasePercentEncoding[buf[i]]
	}
	return percentEncoded
}
