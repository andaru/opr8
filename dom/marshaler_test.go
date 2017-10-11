package dom

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"

	xml "github.com/andaru/flexml"
	"github.com/stretchr/testify/assert"
)

func TestNewMarshaler(t *testing.T) {
	type args struct {
		n    Node
		opts []MarshalerOption
	}
	tests := []struct {
		name  string
		args  args
		flags []bitflag
		want  *Marshaler
	}{
		{
			"without flags",
			args{n: newStartElement(xml.StartElement{})},
			nil,
			&Marshaler{Node: newStartElement(xml.StartElement{})},
		},
		{
			"with explicit default namespace marshaling",
			args{
				n:    newStartElement(xml.StartElement{}),
				opts: []MarshalerOption{WithExplicitNS()},
			},
			[]bitflag{marshalExplicitNS},
			&Marshaler{Node: newStartElement(xml.StartElement{}), opts: marshalExplicitNS},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMarshaler(tt.args.n, tt.args.opts...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMarshaler() = %v, want %v", got, tt.want)
			}
			for i, flag := range tt.flags {
				if got.opts.Has(flag) {
					continue
				}
				t.Errorf("expected flag #%d = %d present in %x, but was not",
					i, flag, got.opts)
			}
		})
	}
}

func TestNodeUnmarshalMarshalEquality(t *testing.T) {
	check := assert.New(t)
	for _, tt := range xmlDecoderTestCases {
		t.Run(tt.filename, func(t *testing.T) {
			f, err := os.Open(tt.filename)
			if err != nil {
				t.Fatal(err)
			}

			unmarshalerOpts := []BuilderOption{
				WithTrimPCData(), WithComments(), WithProcInst(), WithDeclaration(),
			}
			doc := NewDocument(context.Background())
			builder := NewBuilder(doc, unmarshalerOpts...)
			unmarshaler := NewUnmarshaler(builder)
			if _, err = unmarshaler.XMLReader().ReadFrom(f); err != nil {
				t.Fatalf("%s: [1] error: %v", tt.filename, err)
			}

			// marshal the decoded tree
			output := &bytes.Buffer{}
			marshaler := NewMarshaler(doc)

			n, err := marshaler.XMLWriter().WriteTo(output)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshaler.WriteTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if int(n) != len(output.String()) {
				t.Errorf("Marshaler.WriteTo() reported %d bytes written but actually wrote %d bytes",
					n, len(output.String()))
			}

			// decode the re-marshaled tree
			doc = NewDocument(context.Background())
			builder = NewBuilder(doc, unmarshalerOpts...)
			unmarshaler = NewUnmarshaler(builder)
			_, err = unmarshaler.XMLReader().ReadFrom(bytes.NewReader(output.Bytes()))
			if err != nil {
				t.Fatalf("%s: [2] error: %v", tt.filename, err)
			}

			// re-marshal the re-decoded tree
			ckoutput := &bytes.Buffer{}
			marshaler = NewMarshaler(doc)
			enc := xml.NewEncoder(ckoutput)
			if err := marshaler.MarshalXML(enc, xml.StartElement{}); (err != nil) != tt.wantErr {
				t.Errorf("Marshaler.WriteTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			// compare the original marshaling output to the output of the
			// re-(un)marshaling
			check.Equal(
				output.String(), ckoutput.String(),
				"first and second responses to xml.Marshal differ.\nfirst:\n%s\nsecond:\n%s\n",
				output.String(), ckoutput.String())

		})
	}
}
