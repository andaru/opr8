package dom

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	xml "github.com/andaru/flexml"
	"github.com/pkg/errors"
)

// TokenDecoder is a token processing provider. It is used by
// implementations in order to validate incoming XML tokens.
type TokenDecoder interface {
	Initialize(...string) error
	Begin(xml.StartElement) error

	StartElement(xml.StartElement) error
	EndElement(xml.EndElement) error
	CharData(xml.CharData) error
	Comment(xml.Comment) error
	ProcInst(xml.ProcInst) error
	Directive(xml.Directive) error

	End(error) error
}

// NewUnmarshaler returns a new JSON/XML DOM unmarshaler, using the provided
// TokenDecoder to react to tokens from the input.
func NewUnmarshaler(td TokenDecoder) *Unmarshaler {
	return &Unmarshaler{TokenDecoder: td}
}

// Builder is a standard DOM document decoder, without a schema.
type Builder struct {
	Node
	opts bitflag
}

// NewBuilder returns a new DOM builder configured with supplied options.
func NewBuilder(root Node, opts ...BuilderOption) *Builder {
	db := &Builder{Node: root}
	for _, opt := range opts {
		opt(db)
	}
	return db
}

// BuilderOption is an option used to configure a newly
// constructed Unmarshaler.
type BuilderOption func(*Builder)

// Unmarshaler is an XML and JSON streaming token decoder providing standard
// unmarshaler interfaces as well as io.ReaderFrom for JSON and XML source data.
type Unmarshaler struct {
	TokenDecoder
	InitializeArgs []string
}

// WithRootNode configures the unmarshaler with the provided root node to
// unmarshal into.
func WithRootNode(n Node) BuilderOption {
	return func(x *Builder) { x.Node = n }
}

// WithDocumentContext configures the unmarshaler with a new document using the
// provided context.
func WithDocumentContext(v context.Context) BuilderOption {
	return func(x *Builder) { x.Node = newDocument(v) }
}

// WithTrimPCData causes the unmarshaler to trim whitespace surrounding decoded
// text elements.
func WithTrimPCData() BuilderOption { return func(x *Builder) { x.opts.Add(parseTrimPCData) } }

// WithComments causes the unmarshaler to include comments in the node tree.
func WithComments() BuilderOption { return func(x *Builder) { x.opts.Add(parseComments) } }

// WithDoctype causes the unmarshaler to include the doctype in the node tree.
func WithDoctype() BuilderOption { return func(x *Builder) { x.opts.Add(parseDoctype) } }

// WithDeclaration causes the unmarshaler to include the <?xml ...?>
// declaration, if present.
func WithDeclaration() BuilderOption { return func(x *Builder) { x.opts.Add(parseDeclaration) } }

// WithProcInst causes the unmarshaler to include any other <?... ...?>
// processing instructions.
func WithProcInst() BuilderOption { return func(x *Builder) { x.opts.Add(parsePI) } }

// WithWhitespacePCData causes the unmarshaler to include text nodes which
// contain only whitespace.
func WithWhitespacePCData() BuilderOption {
	return func(x *Builder) { x.opts.Add(parseWSPCData) }
}

// WithRootFragment causes the unmarshaler to set its root node to a DocumentFragment
// rather than a Document node.
func WithRootFragment() BuilderOption { return func(x *Builder) { x.opts.Add(parseFragment) } }

// JSONReader returns a streaming JSON decoding reader
func (un *Unmarshaler) JSONReader() io.ReaderFrom { return readerJSON{un} }

// XMLReader returns a streaming XML decoder reader
func (un *Unmarshaler) XMLReader() io.ReaderFrom { return readerXML{un} }

// Root returns the root node of unmarshalled data. If no data has been
// unmarshaled, nil will be returned.
func (un Builder) Root() Node {
	for it := un.Node.nodePtr(); it != nil; it = it.parent {
		if it.parent == nil {
			return it
		}
	}
	return nil
}

// Initialize passes configuration key/value pairs from the calling unmarshaler.
func (un Builder) Initialize(...string) error { return nil }

// Begin responds to the beginning of document decoding.
func (un Builder) Begin(se xml.StartElement) error {
	// allow decoding to begin at any element in the token stream,
	// as long as we have a non-nil cursor
	if un.Node == nil {
		return errors.New("error decoding XML: node cursor is nil")
	}
	return nil
}

// StartElement responds to a new start element token.
func (un *Builder) StartElement(se xml.StartElement) error {
	newNode := newStartElement(se)
	newNode.parent = un.Node.nodePtr()
	un.Node = newNode
	return nil
}

// EndElement responds to a new end element token.
func (un *Builder) EndElement(xml.EndElement) error {
	n := un.Node.nodePtr()
	if n.parent == nil {
		return errors.Wrap(ErrHierarchyRequest, "context node has a nil parent")
	} else if err := n.parent.AppendChild(un.Node); err != nil {
		return err
	}
	un.Node = un.Node.Parent()
	return nil
}

// CharData responds to a new text token.
func (un *Builder) CharData(cd xml.CharData) error {
	if err := allowInsertChildErr(un.Node.NodeType(), NodeTypeText); err != nil {
		return err
	} else if len(cd) == 0 && !un.opts.Has(parseWSPCData) {
		return nil
	} else if !un.opts.Has(parseTrimPCData) {
		return un.Node.AppendChild(newText(cd.Copy()))
	} else if trimmed := bytes.TrimSpace(cd.Copy()); un.opts.Has(parseWSPCData) || len(trimmed) > 0 {
		return un.Node.AppendChild(newText(trimmed))
	}
	return nil
}

// Comment responds to a new comment token.
func (un *Builder) Comment(c xml.Comment) error {
	if !un.opts.Has(parseComments) {
		return nil
	} else if err := allowInsertChildErr(un.Node.NodeType(), NodeTypeComment); err != nil {
		return err
	}
	return un.Node.AppendChild(newComment(c))
}

// ProcInst responds to a new processing instruction or declaration.
func (un *Builder) ProcInst(pi xml.ProcInst) error {
	nt := NodeTypeProcessingInstruction
	if pi.Target == "xml" {
		nt = NodeTypeDeclaration
	}
	if err := allowInsertChildErr(un.Node.NodeType(), nt); err != nil {
		return err
	}
	// only store ProcInst/Declarations if enabled
	switch {
	case nt == NodeTypeDeclaration && un.opts.Has(parseDeclaration):
		decl := newDeclaration(pi)
		err := un.Node.AppendChild(decl)
		return err
	case nt == NodeTypeProcessingInstruction && un.opts.Has(parsePI):
		return un.Node.AppendChild(newProcInst(pi))
	}
	return nil
}

// Directive responds to a new XML directive.
func (un *Builder) Directive(xml.Directive) error { return errors.New("TODO") }

// End responds to the end of document processing. The error EOF indicates
// normal completion.
func (un Builder) End(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

// UnmarshalXML performs decoding of a stream of XML tokens from d
// handled by the TokenDecoder.
func (un *Unmarshaler) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	td := un.TokenDecoder
	if err := td.Begin(start); err != nil {
		return err
	}
	for {
		t, err := d.Token()
		if err != nil {
			return td.End(err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			err = td.StartElement(t)
		case xml.EndElement:
			err = td.EndElement(t)
		case xml.Comment:
			err = td.Comment(t)
		case xml.CharData:
			err = td.CharData(t)
		case xml.Directive:
			err = td.Directive(t)
		case xml.ProcInst:
			err = td.ProcInst(t)
		}
		if err != nil {
			return err
		}
	}
}

// UnmarshalJSON decodes JSON input from the byte slice b.
func (un *Unmarshaler) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	td := un.TokenDecoder
	if err := td.Begin(xml.StartElement{}); err != nil {
		return err
	}
	jd := json.NewDecoder(bytes.NewReader(b))
	jd.UseNumber()
	decoder := &JSONDecoder{TokenDecoder: td, Decoder: jd}
	return td.End(decoder.Run())
}

type readerXML struct{ *Unmarshaler }
type readerJSON struct{ *Unmarshaler }

func (un *Unmarshaler) tdInit(args ...string) error {
	if len(un.InitializeArgs) > 0 && len(un.InitializeArgs)%2 == 0 {
		return un.TokenDecoder.Initialize(un.InitializeArgs...)
	}
	return un.TokenDecoder.Initialize(args...)
}

func (rx readerXML) ReadFrom(r io.Reader) (int64, error) {
	if err := rx.tdInit(); err != nil {
		return 0, err
	}
	cr := &countReader{r, 0}
	err := rx.UnmarshalXML(xml.NewDecoder(cr), xml.StartElement{})
	return cr.n, err
}

func (rj readerJSON) ReadFrom(r io.Reader) (n int64, err error) {
	if err := rj.tdInit("name.resolver", "rfc7951"); err != nil {
		return 0, err
	}
	cr := &countReader{r, 0}
	if err := rj.Begin(xml.StartElement{}); err != nil {
		return 0, err
	}
	jd := json.NewDecoder(cr)
	jd.UseNumber()
	decoder := &JSONDecoder{TokenDecoder: rj.TokenDecoder, Decoder: jd}
	decodeErr := decoder.Run()
	return cr.n, rj.End(decodeErr)
}

type countReader struct {
	io.Reader
	n int64
}

func (r *countReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	r.n += int64(n)
	return n, err
}

func newStartElement(elem xml.StartElement) *node {
	n := &node{value: &element{name: elem.Name}}
	for _, a := range elem.Attr {
		appendAttribute(newAttribute(a), n)
	}
	return n
}

type bitflag uint32

func (f bitflag) Has(other bitflag) bool { return f&other != 0 }
func (f *bitflag) Add(other bitflag)     { *f |= other }
func (f *bitflag) Clear(other bitflag)   { *f &= ^other }
func (f *bitflag) Toggle(other bitflag)  { *f ^= other }

const (
	parseDeclaration bitflag = 1 << iota
	parseDoctype
	parsePI
	parseComments
	parseTrimPCData
	parseWSPCData
	parseFragment
)

var (
	_ TokenDecoder  = &Builder{}
	_ io.ReaderFrom = readerJSON{}
	_ io.ReaderFrom = readerXML{}
	_ io.Reader     = &countReader{}
)
