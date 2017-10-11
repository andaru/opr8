package dom

import (
	"io"

	xml "github.com/andaru/flexml"
	"github.com/pkg/errors"
)

// Marshaler is able to emit Node as XML or JSON output.
//
// The type implements the XML and JSON standard marshaler interfaces:
//   (encoding/xml).Marshaler
//   (encoding/json).Marshaler
//
// Specific XML and JSON io.WriterTo encoders are provided by functions:
//   XMLWriter() io.WriterTo
//   JSONWriter() io.WriterTo
type Marshaler struct {
	Node

	opts           bitflag
	emitExplicitNS bool
}

const (
	marshalExplicitNS bitflag = 1 << iota
)

// NewMarshaler returns a marshaler for node, configured with options provided.
func NewMarshaler(node Node, opts ...MarshalerOption) *Marshaler {
	e := &Marshaler{Node: node}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// MarshalerOption is a constructor option for a Node marshaler.
type MarshalerOption func(*Marshaler)

// WithExplicitNS is a marshaler option which causes the XML namespace to be
// emitted with every element, even if the namespace matches the parent element
// namespace.
func WithExplicitNS() MarshalerOption { return func(e *Marshaler) { e.opts.Add(marshalExplicitNS) } }

// XMLWriter returns an XML io.WriterTo for the Node
func (m *Marshaler) XMLWriter() io.WriterTo { return writerXML{m} }

// MarshalXML encodes .Node to the XML encoder.
func (m *Marshaler) MarshalXML(enc *xml.Encoder, se xml.StartElement) error {
	if err := treeOrder(m, enc); err != nil {
		return err
	}
	return enc.Flush()
}

type writerXML struct{ *Marshaler }

func (wx writerXML) WriteTo(w io.Writer) (n int64, err error) {
	cw := &countWriter{w, 0}
	enc := xml.NewEncoder(cw)
	err = treeOrder(wx.Marshaler, enc)
	err2 := enc.Flush()
	if err == nil {
		err = err2
	}
	return cw.n, err
}

type countWriter struct {
	io.Writer
	n int64
}

func (w *countWriter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	w.n += int64(n)
	return
}

var xmlnsDefault = xml.Name{Local: "xmlns"}

type encoderStack struct {
	d []func() (*node, error)
}

func (s *encoderStack) len() int { return len(s.d) }
func (s *encoderStack) push(fn func() (*node, error)) {
	s.d = append(s.d, fn)
}
func (s *encoderStack) pop() (fn func() (*node, error)) {
	fn, s.d = s.d[len(s.d)-1], s.d[:len(s.d)-1]
	return
}

func treeOrder(e *Marshaler, enc *xml.Encoder) error {
	s := &encoderStack{}
	s.push(encodeNodeValueStart(e, enc, e.Node.nodePtr()))
	for s.len() > 0 {
		n, err := s.pop()()
		if err != nil {
			return errors.WithStack(err)
		} else if n == nil {
			continue
		}
		s.push(encodeNodeValueEnd(e, enc, n))
		if n.firstChild == nil {
			continue
		}
		s.push(encodeNodeValueStart(e, enc, n.firstChild.prevSib))
		for it := n.firstChild.prevSib.prevSib; it.nextSib != nil; it = it.prevSib {
			s.push(encodeNodeValueStart(e, enc, it))
		}
	}
	return nil
}

func endElementForNode(n *node, explicitNS bool) xml.EndElement {
	name := n.xmlName()
	if n.parent != nil && !explicitNS && n.parent.xmlName().Space == name.Space {
		name.Space = ""
	}
	return xml.EndElement{Name: name}
}

func encodeNodeValueStart(e *Marshaler, enc *xml.Encoder, n *node) func() (*node, error) {
	return func() (*node, error) {
		var err error
		switch n.NodeType() {
		case NodeTypeElement:
			name := n.xmlName()
			if n.parent != nil && !e.opts.Has(marshalExplicitNS) && n.parent.xmlName().Space == name.Space {
				name.Space = ""
			}
			var attrs []xml.Attr
			for it := n.firstAttr; it != nil; it = it.nextSib {
				this := it.value.(*attribute).Attr
				if this.Name != xmlnsDefault {
					attrs = append(attrs, this)
				}
			}
			err = enc.EncodeToken(xml.StartElement{Name: name, Attr: attrs})
		case NodeTypeComment:
			err = enc.EncodeToken(xml.Comment(n.asComment().text.value))
		case NodeTypeText:
			err = enc.EncodeToken(n.asText().charData())
		case NodeTypeDeclaration:
			err = enc.EncodeToken(n.asDeclaration().declaration.ProcInst)
		case NodeTypeProcessingInstruction:
			err = enc.EncodeToken(n.asProcInst().procinst.ProcInst)
		case NodeTypeDocumentFragment, NodeTypeDocument:
			// these nodes have no value to encode
		default:
			err = errors.Errorf("MarshalXML called on unexpected node type %s", n.NodeType())
		}
		if err != nil {
			return n, errors.WithStack(err)
		}
		return n, nil
	}
}

func encodeNodeValueEnd(e *Marshaler, enc *xml.Encoder, n *node) func() (*node, error) {
	return func() (*node, error) {
		switch n.NodeType() {
		case NodeTypeElement:
			if err := enc.EncodeToken(endElementForNode(n, e.emitExplicitNS)); err != nil {
				return nil, errors.WithStack(err)
			}
		}
		return nil, nil
	}
}
