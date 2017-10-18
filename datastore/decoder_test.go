package datastore

import (
	"strings"
	"testing"

	"github.com/andaru/flexml"
	"github.com/andaru/opr8/dom"
	"github.com/andaru/opr8/modules"
)

func TestYANGDecoder(t *testing.T) {
	c := modules.NewCollection()
	modules.SetYANGPath("../yang_modules/ietf/RFC/...", "./testdata/")
	if errs := c.ImportAll(); errs != nil {
		for i, err := range errs {
			t.Logf("import error %02d/%02d: %v", i, len(errs), err)
		}
	}
	if errs := c.Process(); errs != nil {
		for _, err := range errs {
			t.Error(err)
		}
		t.Fatal("fatal YANG processing errors")
	}

	for _, tt := range []struct {
		name         string
		xml          string
		json         string
		wantXML      string
		decodeErrors []string
	}{
		{
			name:    "module1 with host-name",
			json:    `{"module1:system":{"host-name": "abc123"}}`,
			xml:     `<system xmlns="urn:mod1"><host-name>abc123</host-name></system>`,
			wantXML: `<system xmlns="urn:mod1"><host-name>abc123</host-name></system>`,
		},

		{
			name:    "module1 with host-name and domain-name-servers",
			json:    `{"module1:system":{"domain-name-servers":["ns1.local","ns2.local"],"host-name":"abc456"}`,
			xml:     `<system xmlns="urn:mod1"><domain-name-servers>ns1.local</domain-name-servers><domain-name-servers>ns2.local</domain-name-servers><host-name>abc456</host-name></system>`,
			wantXML: `<system xmlns="urn:mod1"><domain-name-servers>ns1.local</domain-name-servers><domain-name-servers>ns2.local</domain-name-servers><host-name>abc456</host-name></system>`,
		},

		{
			name:    "interfaces",
			json:    `{"module1:interfaces": {"interface":[{"config":{"interface-name":"Ethernet1"}, "interface-name": "Ethernet1"}]}}`,
			xml:     `<interfaces xmlns="urn:mod1"><interface><config><interface-name>Ethernet1</interface-name></config><interface-name>Ethernet1</interface-name></interface></interfaces>`,
			wantXML: `<interfaces xmlns="urn:mod1"><interface><config><interface-name>Ethernet1</interface-name></config><interface-name>Ethernet1</interface-name></interface></interfaces>`,
		},

		{
			name:    "interfaces with choice and case usage",
			json:    `{"module1:interfaces": {"interface":[{"config":{"interface-name":"Ethernet1", "ethernet-address": "aa:bb:cc:dd:ee:ff"}, "interface-name": "Ethernet1"}]}}`,
			xml:     `<interfaces xmlns="urn:mod1"><interface><config><interface-name>Ethernet1</interface-name><ethernet-address>aa:bb:cc:dd:ee:ff</ethernet-address></config><interface-name>Ethernet1</interface-name></interface></interfaces>`,
			wantXML: `<interfaces xmlns="urn:mod1"><interface><config><interface-name>Ethernet1</interface-name><ethernet-address>aa:bb:cc:dd:ee:ff</ethernet-address></config><interface-name>Ethernet1</interface-name></interface></interfaces>`,
		},

		// partial or complete error cases. in partial errors, some
		// elements are decoded (see wantXML).
		{
			name:    "invalid module namespace, no elements decoded",
			xml:     `<system xmlns="BAD:urn:mod1"><host-name>abc123</host-name></system>`,
			wantXML: ``,
			decodeErrors: []string{
				`unexpected child element <system xmlns="BAD:urn:mod1">`,
				`unexpected child element <host-name xmlns="BAD:urn:mod1">`,
			},
		},
		{
			name:    "invalid JSON module name for host-name, only system decoded",
			wantXML: `<system xmlns="urn:mod1"></system>`,
			json:    `{"module1:system": {"bad:host-name":"foo"}}`,
			decodeErrors: []string{
				`unexpected element <host-name> in unknown module "bad"`,
			},
		},
		{
			name:    "invalid namespace for host-name, only system decoded",
			xml:     `<system xmlns="urn:mod1"><host-name xmlns="foo">abc123</host-name></system>`,
			wantXML: `<system xmlns="urn:mod1"></system>`,
			decodeErrors: []string{
				`unexpected child element <host-name> in namespace "foo" (expected namespace "urn:mod1")`,
			},
		},
		{
			name:    "invalid name for hostname, only system decoded",
			xml:     `<system xmlns="urn:mod1"><hostname>abc123</hostname></system>`,
			wantXML: `<system xmlns="urn:mod1"></system>`,
			decodeErrors: []string{
				`unexpected child element <hostname xmlns="urn:mod1">`,
			},
		},
		{
			name:    "invalid name for hostname, only system decoded",
			json:    `{"module1:system": {"hostname":"foo"}}`,
			wantXML: `<system xmlns="urn:mod1"></system>`,
			decodeErrors: []string{
				`unexpected child element <hostname>`,
			},
		},
	} {
		if tt.xml != "" {
			t.Run("xml:"+tt.name, func(t *testing.T) {
				xmlDoc := dom.NewDocument(nil)
				td := &Decoder{Node: xmlDoc, Modules: c}
				un := dom.NewUnmarshaler(td)
				un.InitializeArgs = []string{"name.resolver", "rfc6020"}
				n, err := un.XMLReader().ReadFrom(strings.NewReader(tt.xml))
				if err != nil {
					t.Fatalf("Unmarshaler.XMLReader().ReadFrom() err = %v, wantErr = false", err)
				}
				if int(n) != len(tt.xml) {
					t.Errorf("Unmarshaler.XMLReader().ReadFrom() reported %d bytes read, want %d",
						n, len(tt.xml))
				}
				if xmlDoc.FirstChild() != td.Root().FirstChild() {
					t.Errorf("xmlDoc.FirstChild() [%#v] is not equal to decoderroot.FirstChild() [%#v]",
						xmlDoc.FirstChild(), td.Root().FirstChild())
				}
				gotErrors := td.DecodingErrors()
				if len(gotErrors) != len(tt.decodeErrors) {
					for i, err := range td.DecodingErrors() {
						t.Logf("error %d: %v", i, err)
					}
					t.Fatalf("got %d decoding errors, want %d", len(gotErrors), len(tt.decodeErrors))
				}
				for i, err := range td.DecodingErrors() {
					want := tt.decodeErrors[i]
					if got := err.Error(); got != want {
						t.Errorf("decoding error %d/%d mismatch:\ngot:\n%s\nwant:\n%s\n",
							i, len(tt.decodeErrors), got, want)
					}
				}
				b, err := flexml.Marshal(dom.NewMarshaler(xmlDoc))
				if err != nil {
					t.Fatalf("xml.Marshal(xmlDoc) error: %v, wantErr false", err)
				} else if string(b) != tt.wantXML {
					t.Errorf("encoded XML did not match input XML, got:\n%s\nwant:\n%s\n", b, tt.wantXML)
				}
			})
		}

		if tt.json != "" {
			t.Run("json:"+tt.name, func(t *testing.T) {
				jsonDoc := dom.NewDocument(nil)
				td := &Decoder{Node: jsonDoc, Modules: c}
				un := dom.NewUnmarshaler(td)
				un.InitializeArgs = []string{"mediatype", "application/yang-data+json"}
				n, err := un.JSONReader().ReadFrom(strings.NewReader(tt.json))
				if err != nil {
					t.Fatalf("(*dom.Unmarshaler).JSONReader().ReadFrom() err = %v, wantErr false", err)
				}
				if int(n) != len(tt.json) {
					t.Errorf("Unmarshaler.JSONReader().ReadFrom() reported %d bytes read, want %d",
						n, len(tt.json))
				}
				if jsonDoc.FirstChild() != td.Root().FirstChild() {
					t.Errorf("jsonDoc.FirstChild() [%#v] is not equal to decoderroot.FirstChild() [%#v]",
						jsonDoc.FirstChild(), td.Root().FirstChild())
				}
				gotErrors := td.DecodingErrors()
				if len(gotErrors) != len(tt.decodeErrors) {
					for i, err := range td.DecodingErrors() {
						t.Logf("error %d: %v", i, err)
					}
					t.Fatalf("got %d decoding errors, want %d", len(gotErrors), len(tt.decodeErrors))
				}
				for i, err := range td.DecodingErrors() {
					want := tt.decodeErrors[i]
					if got := err.Error(); got != want {
						t.Errorf("decoding error %d/%d mismatch:\ngot:\n%s\nwant:\n%s\n",
							i, len(tt.decodeErrors), got, want)
					}
				}
				b, err := flexml.Marshal(dom.NewMarshaler(jsonDoc))
				if err != nil {
					t.Fatalf("xml.Marshal(jsonDoc) error: %v, wantErr = false", err)
				} else if string(b) != tt.wantXML {
					t.Errorf("encoded XML from JSON did not match input XML, got\n%s\nwant:\n%s\n", b, tt.xml)
				}
			})
		}
	}
}
