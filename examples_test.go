package iri_test

import (
	"fmt"

	"github.com/contomap/iri"
)

func ExampleParse_https() {
	value, _ := iri.Parse("https://user@example.com/µ/path?q=€#frag1")
	fmt.Printf("%#v", value)
	// Output: iri.IRI{Scheme:"https", ForceAuthority:false, Authority:"user@example.com", Path:"/µ/path", ForceQuery:false, Query:"q=€", ForceFragment:false, Fragment:"frag1"}
}

func ExampleParse_mailto() {
	value, _ := iri.Parse("mailto:user@example.com")
	fmt.Printf("%#v", value)
	// Output: iri.IRI{Scheme:"mailto", ForceAuthority:false, Authority:"", Path:"user@example.com", ForceQuery:false, Query:"", ForceFragment:false, Fragment:""}
}

func ExampleParse_empty() {
	value, _ := iri.Parse("")
	fmt.Printf("%#v", value)
	// Output: iri.IRI{Scheme:"", ForceAuthority:false, Authority:"", Path:"", ForceQuery:false, Query:"", ForceFragment:false, Fragment:""}
}

func ExampleParse_emptyComponents() {
	value, _ := iri.Parse("//?#")
	fmt.Printf("%#v", value)
	// Output: iri.IRI{Scheme:"", ForceAuthority:true, Authority:"", Path:"", ForceQuery:true, Query:"", ForceFragment:true, Fragment:""}
}

func ExampleIRI_String_common() {
	value := iri.IRI{Scheme: "https", ForceAuthority: false, Authority: "user@example.com", Path: "/sub/path", ForceQuery: false, Query: "q=1", ForceFragment: false, Fragment: "frag1"}
	fmt.Printf("%s", value)
	// Output: https://user@example.com/sub/path?q=1#frag1
}

func ExampleIRI_String_empty() {
	var value iri.IRI
	fmt.Printf("'%s'", value)
	// Output: ''
}

func ExampleIRI_String_emptyComponents() {
	value := iri.IRI{ForceAuthority: true, ForceQuery: true, ForceFragment: true}
	fmt.Printf("%s", value)
	// Output: //?#
}

func ExampleNormalizePercentEncoding() {
	value, _ := iri.Parse("https://example.org/dog%20house/%c2%B5")
	result, _ := iri.NormalizePercentEncoding(value)
	fmt.Printf("%s", result.Path)
	// Output: /dog%20house/µ
}

func ExampleIRI_ResolveReference() {
	base, _ := iri.Parse("https://example.com/sub/path/µ?q=1#frag1")
	fmt.Printf("%s\n", base.ResolveReference(iri.IRI{Fragment: "frag2"}))
	fmt.Printf("%s\n", base.ResolveReference(iri.IRI{Path: ".."}))
	fmt.Printf("%s\n", base.ResolveReference(iri.IRI{}))
	// Output:
	// https://example.com/sub/path/µ?q=1#frag2
	// https://example.com/sub/
	// https://example.com/sub/path/µ?q=1
}
