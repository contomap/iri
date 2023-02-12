# Internationalized Resource identifier (IRI) library

This library contains facilities for working with Internationalized Resource
Identifiers (IRI) as specified in RFC 3987.

RFC reference: https://www.ietf.org/rfc/rfc3987.html

Generally speaking, an IRI is a URI that allows international characters;
And a URI is a generic form of the now rather deprecated URL (= address) and URN (= name).

## Origin

The code in this library is forked off from
https://github.com/google/xtoproto/rdf/iri@4cad7286ebfcd65dfec376912eb3b7a03c531c9b

The code from `xtoproto` itself is derived from
https://github.com/golang/go/blob/master/src/net/url/url.go ,
which is why the `LICENSE` file contains extra information.

A separate repository was created to avoid adding a dependency to an unrelated large library,
just for one small package, which should have been on its own to begin with.

Tag `v0.1.0` contains the exact replication of the forked code.

## LICENSE

Apache License 2.0, with notice from the Go language project. See `LICENSE` file.
