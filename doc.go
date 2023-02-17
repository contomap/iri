// Package iri provides facilities for working with Internationalized Resource Identifiers as specified in RFC 3987.
//
// RFC reference: https://www.ietf.org/rfc/rfc3987.html
//
// Although conceptually an IRI is meant to be a generalized concept of a URL,
// type IRI and its functions cannot be used as a drop-in replacement of
// "net/url.URL". The standard Go implementation handles many corner cases and
// "real life" behaviour of existing systems.
// The implementation of this package is inspired by "net/url", yet follows
// more strictly the RFC specifications.
package iri
