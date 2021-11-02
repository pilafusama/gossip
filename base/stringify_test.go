package base

// These tests confirm that our various structures stringify correctly.

import (
	"fmt"
	"testing"
)

// Generic test for testing anything with a String() method.
type stringTest struct {
	description string
	input       fmt.Stringer
	expected    string
}

func doTests(tests []stringTest, t *testing.T) {
	passed := 0
	for _, test := range tests {
		if test.input.String() != test.expected {
			t.Errorf("[FAIL] %v: Expected: \"%v\", Got: \"%v\"",
				test.description,
				test.expected,
				test.input.String(),
			)
		} else {
			passed++
		}
	}
	t.Logf("Passed %v/%v tests", passed, len(tests))
}

// Some global ports to use since port is still a pointer.
var port5060 uint16 = 5060
var port6060 uint16 = 6060
var noParams = NewParams()

func TestSipUri(t *testing.T) {
	doTests([]stringTest{
		{"Basic SIP URI",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
			"sip:alice@wonderland.com"},
		{"SIP URI with no user",
			&SipUri{User: NoString{}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
			"sip:wonderland.com"},
		{"SIP URI with password",
			&SipUri{User: String{"alice"}, Password: String{"hunter2"}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
			"sip:alice:hunter2@wonderland.com"},
		{"SIP URI with explicit port 5060",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", Port: &port5060, UriParams: noParams, Headers: noParams},
			"sip:alice@wonderland.com:5060"},
		{"SIP URI with other port",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", Port: &port6060, UriParams: noParams, Headers: noParams},
			"sip:alice@wonderland.com:6060"},
		{"Basic SIPS URI",
			&SipUri{IsEncrypted: true, User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
			"sips:alice@wonderland.com"},
		{"SIP URI with one parameter",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: NewParams().Add("food", String{"cake"}),
				Headers:   noParams},
			"sip:alice@wonderland.com;food=cake"},
		{"SIP URI with one no-value parameter",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: NewParams().Add("something", NoString{}),
				Headers:   noParams},
			"sip:alice@wonderland.com;something"},
		{"SIP URI with three parameters",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: NewParams().Add("food", String{"cake"}).Add("something", NoString{}).Add("drink", String{"tea"}),
				Headers:   noParams},
			"sip:alice@wonderland.com;food=cake;something;drink=tea"},
		{"SIP URI with one header",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: noParams,
				Headers:   NewParams().Add("CakeLocation", String{"Tea Party"})},
			"sip:alice@wonderland.com?CakeLocation=\"Tea Party\""},
		{"SIP URI with three headers",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: noParams,
				Headers:   NewParams().Add("CakeLocation", String{"Tea Party"}).Add("Identity", String{"Mad Hatter"}).Add("OtherHeader", String{"Some value"})},
			"sip:alice@wonderland.com?CakeLocation=\"Tea Party\"&Identity=\"Mad Hatter\"&OtherHeader=\"Some value\""},
		{"SIP URI with parameter and header",
			&SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com",
				UriParams: NewParams().Add("food", String{"cake"}),
				Headers:   NewParams().Add("CakeLocation", String{"Tea Party"})},
			"sip:alice@wonderland.com;food=cake?CakeLocation=\"Tea Party\""},
		{"Wildcard URI", &WildcardUri{}, "*"},
	}, t)
}

func TestHeaders(t *testing.T) {
	doTests([]stringTest{
		// To Headers.
		{"Basic To Header",
			&ToHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"To: <sip:alice@wonderland.com>"},
		{"To Header with display name",
			&ToHeader{DisplayName: String{"Alice Liddell"},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"To: \"Alice Liddell\" <sip:alice@wonderland.com>"},
		{"To Header with parameters",
			&ToHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  NewParams().Add("food", String{"cake"})},
			"To: <sip:alice@wonderland.com>;food=cake"},

		// From Headers.
		{"Basic From Header",
			&FromHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"From: <sip:alice@wonderland.com>"},
		{"From Header with display name",
			&FromHeader{DisplayName: String{"Alice Liddell"},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"From: \"Alice Liddell\" <sip:alice@wonderland.com>"},
		{"From Header with parameters",
			&FromHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  NewParams().Add("food", String{"cake"})},
			"From: <sip:alice@wonderland.com>;food=cake"},

		// Contact Headers
		{"Basic Contact Header",
			&ContactHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"Contact: <sip:alice@wonderland.com>"},
		{"Contact Header with display name",
			&ContactHeader{DisplayName: String{"Alice Liddell"},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  noParams},
			"Contact: \"Alice Liddell\" <sip:alice@wonderland.com>"},
		{"Contact Header with parameters",
			&ContactHeader{DisplayName: NoString{},
				Address: &SipUri{User: String{"alice"}, Password: NoString{}, Host: "wonderland.com", UriParams: noParams, Headers: noParams},
				Params:  NewParams().Add("food", String{"cake"})},
			"Contact: <sip:alice@wonderland.com>;food=cake"},
		{"Contact Header with Wildcard URI",
			&ContactHeader{DisplayName: NoString{}, Address: &WildcardUri{}, Params: noParams},
			"Contact: *"},
		{"Contact Header with display name and Wildcard URI",
			&ContactHeader{DisplayName: String{"Mad Hatter"}, Address: &WildcardUri{}, Params: noParams},
			"Contact: \"Mad Hatter\" *"},
		{"Contact Header with Wildcard URI and parameters",
			&ContactHeader{DisplayName: NoString{}, Address: &WildcardUri{}, Params: NewParams().Add("food", String{"cake"})},
			"Contact: *;food=cake"},

		// Via Headers.
		{"Basic Via Header", ViaHeader{&ViaHop{"SIP", "2.0", "UDP", "wonderland.com", nil, NewParams()}}, "Via: SIP/2.0/UDP wonderland.com"},
		{"Via Header with port", ViaHeader{&ViaHop{"SIP", "2.0", "UDP", "wonderland.com", &port6060, NewParams()}}, "Via: SIP/2.0/UDP wonderland.com:6060"},
		{"Via Header with params", ViaHeader{
			&ViaHop{"SIP", "2.0", "UDP", "wonderland.com", &port6060, NewParams().Add("food", String{"cake"}).Add("delicious", NoString{})}},
			"Via: SIP/2.0/UDP wonderland.com:6060;food=cake;delicious"},
		{"Via Header with 3 simple hops", ViaHeader{
			&ViaHop{"SIP", "2.0", "UDP", "wonderland.com", nil, NewParams()},
			&ViaHop{"SIP", "2.0", "TCP", "looking-glass.net", nil, NewParams()},
			&ViaHop{"SIP", "2.0", "UDP", "oxford.co.uk", nil, NewParams()},
		}, "Via: SIP/2.0/UDP wonderland.com, SIP/2.0/TCP looking-glass.net, SIP/2.0/UDP oxford.co.uk"},
		{"Via Header with 3 complex hops", ViaHeader{
			&ViaHop{"SIP", "2.0", "UDP", "wonderland.com", &port5060, NewParams()},
			&ViaHop{"SIP", "2.0", "TCP", "looking-glass.net", &port6060, NewParams().Add("food", String{"cake"})},
			&ViaHop{"SIP", "2.0", "UDP", "oxford.co.uk", nil, NewParams().Add("delicious", NoString{})},
		}, "Via: SIP/2.0/UDP wonderland.com:5060, SIP/2.0/TCP looking-glass.net:6060;food=cake, SIP/2.0/UDP oxford.co.uk;delicious"},

		// Require Headers.
		{"Require Header (empty)", &RequireHeader{[]string{}}, "Require: "},
		{"Require Header (one option)", &RequireHeader{[]string{"NewFeature1"}}, "Require: NewFeature1"},
		{"Require Header (three options)", &RequireHeader{[]string{"NewFeature1", "FunkyExtension", "UnnecessaryAddition"}}, "Require: NewFeature1, FunkyExtension, UnnecessaryAddition"},

		// Supported Headers.
		{"Supported Header (empty)", &SupportedHeader{[]string{}}, "Supported: "},
		{"Supported Header (one option)", &SupportedHeader{[]string{"NewFeature1"}}, "Supported: NewFeature1"},
		{"Supported Header (three options)", &SupportedHeader{[]string{"NewFeature1", "FunkyExtension", "UnnecessaryAddition"}}, "Supported: NewFeature1, FunkyExtension, UnnecessaryAddition"},

		// Proxy-Require Headers.
		{"Proxy-Require Header (empty)", &ProxyRequireHeader{[]string{}}, "Proxy-Require: "},
		{"Proxy-Require Header (one option)", &ProxyRequireHeader{[]string{"NewFeature1"}}, "Proxy-Require: NewFeature1"},
		{"Proxy-Require Header (three options)", &ProxyRequireHeader{[]string{"NewFeature1", "FunkyExtension", "UnnecessaryAddition"}}, "Proxy-Require: NewFeature1, FunkyExtension, UnnecessaryAddition"},

		// Unsupported Headers.
		{"Unsupported Header (empty)", &UnsupportedHeader{[]string{}}, "Unsupported: "},
		{"Unsupported Header (one option)", &UnsupportedHeader{[]string{"NewFeature1"}}, "Unsupported: NewFeature1"},
		{"Unsupported Header (three options)", &UnsupportedHeader{[]string{"NewFeature1", "FunkyExtension", "UnnecessaryAddition"}}, "Unsupported: NewFeature1, FunkyExtension, UnnecessaryAddition"},

		// Various simple headers.
		{"Call-ID Header", CallId("call-id-1"), "Call-ID: call-id-1"},
		{"CSeq Header", &CSeq{1234, "INVITE"}, "CSeq: 1234 INVITE"},
		{"Max Forwards Header", MaxForwards(70), "Max-Forwards: 70"},
		{"Content Length Header", ContentLength(70), "Content-Length: 70"},
	}, t)
}
