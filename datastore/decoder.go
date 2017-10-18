package datastore

import (
	"io"
	"strings"

	xml "github.com/andaru/flexml"
	"github.com/andaru/opr8/dom"
	"github.com/andaru/opr8/modules"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"
)

// Decoder is a YANG data decoder. It may be used with a dom.Unmarshaler to
// read YANG data from JSON or XML sources with streaming support.
//
// Example:
//   modules := modules.NewCollection()
//   /* ... import YANG modules ... */
//   document := dom.NewDocument(nil)
//   yangDecoder := &Decoder{Node: document, Modules: modules}
//   r := dom.NewUnmarshaler(yangDecoder)
//   f, err := os.Open("yang_data.xml")
//   if err != nil { panic(err) }
//   r.XMLReader().ReadFrom(f)
//   // set root to the same node as document
//   root := yangDecoder.Root()
type Decoder struct {
	Node    dom.Node
	Modules *modules.Collection

	schema    *yang.Entry
	stack     yangDecoderStack
	skip      bool
	errors    []error
	childname nameLookup
}

// DecodingErrors returns the YANG schema errors accumulated during
// data decoding.
func (un Decoder) DecodingErrors() []error { return un.errors }

// Root returns the decoder root node
func (un Decoder) Root() dom.Node {
	for n := un.Node; n != nil; n = n.Parent() {
		if n.Parent() == nil {
			return n
		}
	}
	return nil
}

// Initialize passes configuration pairs from the calling unmarshaler.
//
// By default, the decoder uses RFC6020 namespace handling, suitable
// for decoding YANG/XML data; namespaces received are used as is (any
// prefix decoding occurs in the xml.Decoder). This can also be
// explicitly enabled using one of the following calls:
//
//   d := &Decoder{ /*...*/ }
//   d.Initialize("name.resolver", "rfc6020") // is equivalent to..
//   d.Initialize("mediatype", "application/yang-data+xml")
//
// If you instead wish to decode JSON documents, you'll need to enable
// YANG/JSON namespace resolution. In this mode, the provided
// namespace is the YANG module name. The resolver converts this to
// the module namespace, by finding the module in its module
// collection.
//
//   d.Initialize("name.resolver", "rfc7951") // which is the same and..
//   d.Initialize("mediatype", "application/yang-data+json")
func (un *Decoder) Initialize(kv ...string) error {
	if len(kv)%2 != 0 {
		return errors.Errorf("called with odd number of arguments: %v", kv)
	}
	var format string
	var k string
	for i := 0; i < len(kv)/2; i++ {
		k = kv[i*2]
		v := kv[i*2+1]
		switch k {
		case "name.resolver", "mediatype", "format":
			format = strings.ToLower(v)
			break
		}
	}

	// default to XML name resolution (pass-through): JSON users must
	// pass arguments to Initialize to have their namespaces decoded
	// correctly in the resulting DOM document.
	un.childname = rfc6020Lookup

	switch format {
	case "rfc7951", "application/yang-data+json", "json":
		un.childname = rfc7951Lookup
	case "default", "rfc6020", "application/yang-data+xml", "xml":
		un.childname = rfc6020Lookup
	default:
		return errors.Errorf("unsupported '%s' key value: %v", k, format)
	}
	return nil
}

// SetSchema sets the Decoder's YANG schema node.
func (un *Decoder) SetSchema(e *yang.Entry) { un.schema = e }

// Begin responds to the beginning of document decoding.
func (un Decoder) Begin(se xml.StartElement) error {
	// allow decoding to begin at any element in the token stream,
	// as long as we have a non-nil cursor
	if un.Node == nil {
		return errors.New("error decoding YANG data: node cursor is nil")
	}
	return nil
}

// StartElement responds to a new start element token.
func (un *Decoder) StartElement(se xml.StartElement) error {
	oldSchema := un.schema
	oldNode := un.Node

	name, err := un.childname(un.Modules, se.Name)
	if err == nil {
		var newSchema *yang.Entry
		newSchema, err = un.dataChild(name)
		if err == nil {
			if !un.skip {
				se.Name = name
				newNode := dom.CreateElement(se)
				if crit := un.Node.AppendChild(newNode); crit != nil {
					return crit
				}
				un.schema = newSchema
				un.Node = newNode
			}
			un.stack.push(func() {
				un.schema = oldSchema
				un.Node = oldNode
			})
			return nil
		}
	}

	un.errors = append(un.errors, err)
	un.skip = true
	un.stack.push(func() {
		un.schema = oldSchema
		un.Node = oldNode
		un.skip = false
	})
	return nil
}

// EndElement responds to a new end element token.
func (un *Decoder) EndElement(xml.EndElement) error {
	un.stack.pop()()
	return nil
}

// CharData responds to a new text token.
func (un *Decoder) CharData(cd xml.CharData) error {
	if un.skip {
		return nil
	}
	// YANG violation
	if un.schema == nil || un.schema.Kind != yang.LeafEntry {
		un.errors = append(un.errors,
			errors.Wrap(dom.ErrHierarchyRequest, "schema node is not a leaf"))
	}

	// TODO: perform constraint checks against schema
	update := un.stack.pop()
	un.stack.push(update)
	text := dom.CreateText(cd)
	if crit := un.Node.AppendChild(text); crit != nil {
		return crit
	}
	update()
	return nil
}

// Comment responds to a new comment token.
func (un *Decoder) Comment(c xml.Comment) error { return nil }

// ProcInst responds to a new processing instruction or declaration.
func (un *Decoder) ProcInst(pi xml.ProcInst) error { return nil }

// Directive responds to a new XML directive.
func (un *Decoder) Directive(xml.Directive) error { return errors.New("TODO") }

// End responds to the end of document processing. The error EOF indicates
// normal completion.
func (un Decoder) End(err error) error {
	if err == io.EOF {
		if un.Node.Parent() == nil {
			return nil
		}
		return io.ErrUnexpectedEOF
	}
	return err
}

func (un *Decoder) dataChild(n xml.Name) (candidate *yang.Entry, err error) {
	if un.schema == nil {
		candidate, err = un.Modules.RootEntry(n)
		if err != nil {
			return nil, errUnexpectedElementName(n)
		}
	} else {
		candidate = dataChild(un.schema, n)
	}
	if candidate == nil {
		return nil, errUnexpectedElementName(n)
	}

	// validate the candidate child
	if ns := n.Space; ns != "" {
		if want := candidate.Namespace().Name; want != ns {
			return nil, errors.Errorf(`unexpected child element <%s> in namespace %q (expected namespace %q)`, n.Local, ns, want)
		}
	}

	return candidate, nil
}

func errUnexpectedElementName(n xml.Name) error {
	// TODO: contextualize the error; with schema node identifier of YANG node
	// without the suggested child
	if n.Space != "" {
		return errors.Errorf(`unexpected child element <%s xmlns=%q>`, n.Local, n.Space)
	}
	return errors.Errorf("unexpected child element <%s>", n.Local)
}

func isData(e *yang.Entry) bool {
	switch e.Kind {
	case yang.LeafEntry, yang.DirectoryEntry, yang.AnyXMLEntry, yang.AnyDataEntry:
		return true
	}
	return false
}

func dataChild(e *yang.Entry, n xml.Name) *yang.Entry {
	if e == nil {
		return nil
	} else if next, ok := e.Dir[n.Local]; ok && isData(next) {
		return next
	}
	for _, ch := range e.Dir {
		if ch.Kind == yang.ChoiceEntry || ch.Kind == yang.CaseEntry {
			if next, ok := ch.Dir[n.Local]; ok && isData(next) {
				return next
			}
		}
		for _, chch := range ch.Dir {
			if chch.Kind == yang.CaseEntry {
				if next, ok := chch.Dir[n.Local]; ok && isData(next) {
					return next
				}
			}
		}
	}
	return nil
}

type yangDecoderStack struct{ d []func() }

func (s *yangDecoderStack) push(fn func()) { s.d = append(s.d, fn) }
func (s *yangDecoderStack) pop() (fn func()) {
	fn, s.d = s.d[len(s.d)-1], s.d[:len(s.d)-1]
	return
}

type nameLookup func(*modules.Collection, xml.Name) (xml.Name, error)

func rfc7951Lookup(ms *modules.Collection, n xml.Name) (xml.Name, error) {
	// In RFC7951 (YANG/JSON), the "namespace" of n is in fact the module name
	if n.Space == "" {
		return n, nil
	} else if mod, err := ms.ModuleEntry(n.Space); err == nil {
		nn := xml.Name{Local: n.Local, Space: mod.Namespace().Name}
		return nn, nil
	}
	return n, errors.Errorf(`unexpected element <%s> in unknown module %q`, n.Local, n.Space)
}

func rfc6020Lookup(ms *modules.Collection, n xml.Name) (xml.Name, error) { return n, nil }
