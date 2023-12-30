# Internationalized Resource identifiers (IRI) library for Go

[![Go version of Go module](https://img.shields.io/github/go-mod/go-version/contomap/iri.svg)](https://github.com/contomap/iri)
[![GoDoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/contomap/iri)
[![GoReportCard](https://goreportcard.com/badge/github.com/contomap/iri)](https://goreportcard.com/report/github.com/contomap/iri)
[![License](https://img.shields.io/github/license/contomap/iri.svg)](https://github.com/contomap/iri/blob/main/LICENSE)

**This repository is archived. Ultimately, there was no need for this library.
Furthermore, the realization is also that "Resource Identifier" could be an "unlimited"
data structure, and separate serialization rules then govern whether it becomes an IRI, URI, URL, or whatever flavor (RFC) wanted.
This would be a further (incompatible) rework of the API.**

This library provides facilities for working with Internationalized Resource
Identifiers (IRI) as specified in RFC 3987.

RFC reference: https://www.ietf.org/rfc/rfc3987.html

Generally speaking, an IRI is a URI that allows international characters;
And a URI is a generic form of the now rather deprecated URL (= address) and URN (= name).

Although conceptually an IRI is meant to be a generalized concept of a URL,
type `iri.IRI` and its functions cannot be used as a drop-in replacement of
`net/url.URL`. The standard Go implementation handles many corner cases and
"real life" behaviour of existing systems.

## Examples

```go
func ExampleParse_https() {
	value, _ := iri.Parse("https://user@example.com/µ/path?q=€#frag1")
	fmt.Printf("%#v", value)
	// Output: iri.IRI{Scheme:"https", ForceAuthority:false, Authority:"user@example.com", Path:"/µ/path", ForceQuery:false, Query:"q=€", ForceFragment:false, Fragment:"frag1"}
}

func ExampleIRI_String_common() {
	value := iri.IRI{Scheme: "https", ForceAuthority: false, Authority: "user@example.com", Path: "/sub/path", ForceQuery: false, Query: "q=1", ForceFragment: false, Fragment: "frag1"}
	fmt.Printf("%s", value)
	// Output: https://user@example.com/sub/path?q=1#frag1
}
```

## Origin

The code in this library is forked off from
[https://github.com/google/xtoproto/rdf/iri @4cad7286ebfcd65dfec376912eb3b7a03c531c9b](https://github.com/google/xtoproto/tree/4cad7286ebfcd65dfec376912eb3b7a03c531c9b/rdf/iri)

The code from `xtoproto` itself is derived from
https://github.com/golang/go/blob/master/src/net/url/url.go ,
which is why the `LICENSE` file contains extra information.

A separate repository was created to avoid adding a dependency to an unrelated large library,
just for one small package, which should have been on its own to begin with.

Tag `v0.1.0` contains the exact replication of the forked code.

## LICENSE

Apache License 2.0, with notice from the Go language project. See `LICENSE` file.
