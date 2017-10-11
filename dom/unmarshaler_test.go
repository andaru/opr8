package dom

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	xml "github.com/andaru/flexml"
)

var jsonDecoderTestCases = []struct {
	name        string
	json        string
	useNumber   bool
	wantXML     string
	wantJSONErr bool
	wantXMLErr  bool
}{
	{
		name:    "empty array",
		json:    `{"foo": []}`,
		wantXML: ``,
	},
	{
		name:    "empty object",
		json:    `{"foo": {}}`,
		wantXML: `<foo></foo>`,
	},
	{
		name:    "object with a string and an array with string fields",
		json:    `{"f":["o","p","r8"], "b": "bar"}`,
		wantXML: `<f>o</f><f>p</f><f>r8</f><b>bar</b>`,
	},
	{
		name:    "object with string fields",
		json:    `{"foo":"bar","qux":"baz"}`,
		wantXML: `<foo>bar</foo><qux>baz</qux>`,
	},
	{
		name:    "object with number fields",
		json:    `{"foo":11,"qux":22}`,
		wantXML: `<foo>11</foo><qux>22</qux>`,
	},
	{
		name:    "object with number fields in array",
		json:    `{"foo":[11,22]}`,
		wantXML: `<foo>11</foo><foo>22</foo>`,
	},
	{
		name:    "object with boolean fields",
		json:    `{"foo":false,"qux":true}`,
		wantXML: `<foo>false</foo><qux>true</qux>`,
	},
	{
		name:    "object with boolean fields in array",
		json:    `{"foo":[false, true]}`,
		wantXML: `<foo>false</foo><foo>true</foo>`,
	},
	{
		name:    "object with mixed value fields",
		json:    `{"a": -1, "a2": 3.14159,"b": false, "c": "fish & chip's shop", "d": [null, null], "e": null}`,
		wantXML: `<a>-1</a><a2>3.14159</a2><b>false</b><c>fish &amp; chip&#39;s shop</c><d></d><d></d><e></e>`,
	},

	{
		name:    "object containing nested objects with repeated objects",
		json:    `{"bgp":{"neighbors":{"neighbor":[{"config":{"neighbor-address": "172.16.11.201", "as-number": 3840}},{"config":{"neighbor-address":"172.16.11.202", "as-number": 65412}}]}}}`,
		wantXML: `<bgp><neighbors><neighbor><config><neighbor-address>172.16.11.201</neighbor-address><as-number>3840</as-number></config></neighbor><neighbor><config><neighbor-address>172.16.11.202</neighbor-address><as-number>65412</as-number></config></neighbor></neighbors></bgp>`,
	},

	{
		name:    "name prefix separation",
		json:    `{"foo:bar": 1, "foo:baz": 2}`,
		wantXML: `<bar xmlns="foo">1</bar><baz xmlns="foo">2</baz>`,
	},

	{
		name:       "name prefix separation bad",
		json:       `{"foo:bar": 1, "foo:": 2}`,
		wantXML:    `<bar xmlns="foo">1</bar>`,
		wantXMLErr: true,
	},
}

func TestUnmarshaler_UnmarshalJSON(t *testing.T) {
	for _, tt := range jsonDecoderTestCases {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument(context.Background())
			builder := NewBuilder(doc)
			unmarshaler := NewUnmarshaler(builder)
			if err := unmarshaler.UnmarshalJSON([]byte(tt.json)); (err != nil) != tt.wantJSONErr {
				t.Errorf("Unmarshaler.UnmarshalXML() error = %v, wantJSONErr %v", err, tt.wantJSONErr)
			}

			output := &bytes.Buffer{}
			w := NewMarshaler(doc)
			if _, err := w.XMLWriter().WriteTo(output); (err != nil) != tt.wantXMLErr {
				t.Errorf("Marshaler.XMLWriter().WriteTo() error = %v, wantXMLErr %v", err, tt.wantXMLErr)
			}
			if tt.wantXML != output.String() {
				t.Errorf("Want converted XML:\n%s\nGot:\n%s\n", tt.wantXML, output.String())
			}
		})
	}
}

func TestUnmarshaler_JSONReader_ReadFrom(t *testing.T) {
	for _, tt := range jsonDecoderTestCases {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument(context.Background())
			builder := NewBuilder(doc, WithComments(), WithProcInst())
			unmarshaler := NewUnmarshaler(builder)
			n, err := unmarshaler.JSONReader().ReadFrom(strings.NewReader(tt.json))
			if (err != nil) != tt.wantJSONErr {
				t.Errorf("Unmarshaler.JSONReader().ReadFrom() error = %v, wantErr %v", err, tt.wantJSONErr)
			} else if int(n) != len(tt.json) {
				t.Errorf("Unmarshaler.JSONReader().ReadFrom() reported %d bytes read, want %d bytes", n, len(tt.json))
			}

			output := &bytes.Buffer{}
			w := NewMarshaler(doc)
			if _, err := w.XMLWriter().WriteTo(output); (err != nil) != tt.wantXMLErr {
				t.Errorf("Marshaler.XMLWriter().WriteTo() error = %v, wantXMLErr %v", err, tt.wantXMLErr)
			}
			if tt.wantXML != "" && tt.wantXML != output.String() {
				t.Errorf("Want converted XML:\n%s\nGot:\n%s\n", tt.wantXML, output.String())
			}
		})
	}
}

var xmlDecoderTestCases = []struct {
	name     string
	filename string
	wantErr  bool
}{
	{
		filename: "./testdata/utftest_utf8_clean.xml",
	},
	{
		filename: "./testdata/record1.xml",
	},
}

func TestUnmarshaler_UnmarshalXML(t *testing.T) {
	for _, tt := range xmlDecoderTestCases {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument(context.Background())
			builder := NewBuilder(doc, WithDeclaration(), WithComments(), WithProcInst(), WithTrimPCData())
			unmarshaler := NewUnmarshaler(builder)
			f, err := os.Open(tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			d := xml.NewDecoder(f)
			if err = unmarshaler.UnmarshalXML(d, xml.StartElement{}); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshaler.UnmarshalXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshaler_XMLReader_ReadFrom(t *testing.T) {
	for _, tt := range xmlDecoderTestCases {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument(context.Background())
			builder := NewBuilder(doc, WithDeclaration(), WithComments(), WithProcInst(), WithTrimPCData(), WithWhitespacePCData(), WithDoctype())
			unmarshaler := NewUnmarshaler(builder)
			f, err := os.Open(tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			_, err = unmarshaler.XMLReader().ReadFrom(f)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshaler.XMLReader().ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkXMLTreeDecoder(b *testing.B) {
	b.ReportAllocs()

	f, err := os.Open("./testdata/record1.xml")
	if err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, 4096)
	n, err := io.ReadFull(f, buf)
	buf = buf[:n]
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := xml.NewDecoder(bytes.NewReader(buf))
		doc := NewDocument(context.Background())
		builder := NewBuilder(doc, WithComments(), WithDeclaration(), WithDoctype(),
			WithProcInst(), WithTrimPCData(), WithWhitespacePCData())
		unmarshaler := NewUnmarshaler(builder)
		if err := unmarshaler.UnmarshalXML(d, xml.StartElement{}); err != nil {
			b.Fatal(err)
		}
	}
}
