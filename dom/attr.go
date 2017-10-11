package dom

import xml "github.com/andaru/flexml"

// Attr interface represents an attribute in an Element object. Typically the
// allowable values for the attribute are defined in a document type definition
// or external schema.
//
// Attr objects inherit the Node interface; their siblings represent other
// attributes of the same parent element, and their parent is their owning
// element node. Furthermore, Attr nodes may not be immediate children of a
// DocumentFragment. However, they can be associated with Element nodes
// contained within a DocumentFragment. In short, users and implementors of the
// DOM need to be aware that Attr nodes have some things in common with other
// objects inheriting the Node interface, but they also are quite distinct.
//
// The attribute's effective value is determined as follows: if this attribute
// has been explicitly assigned any value, that value is the attribute's
// effective value; otherwise, if there is a declaration for this attribute, and
// that declaration includes a default value, then that default value is the
// attribute's effective value; otherwise, the attribute does not exist on this
// element in the structure model until it has been explicitly added. Note that
// the nodeValue attribute on the Attr instance can also be used to retrieve the
// string version of the attribute's value(s).
type Attr interface {
	Node
}

// AttributeProvider is a Node type which has a collection of attributes, such
// as an Element or Declaration.
type AttributeProvider interface {
	// FirstAttribute returns the first child attribute, or nil if there are no
	// attributes.
	FirstAttribute() Attr
	// LastAttribute returns the last child attribute, or nil if there are no
	// attributes.
	LastAttribute() Attr
	// AppendAttribute appends the provided XML attribute at the end of the
	// attribute list.
	AppendAttribute(xml.Attr) error
	// PrependAttribute inserts the provided XML attribute at the beginning of
	// the attribute list.
	PrependAttribute(xml.Attr) error
	// InsertAttributeAfter appends the provided XML attribute after the
	// provided reference attribute.
	InsertAttributeAfter(xml.Attr, Attr) error
	// InsertAttributeAfter inserts the provided XML attribute before the
	// provided reference attribute.
	InsertAttributeBefore(xml.Attr, Attr) error
}

type attribute struct{ xml.Attr }

func (a attribute) Name() xml.Name     { return a.Attr.Name }
func (a attribute) nodeType() NodeType { return NodeTypeAttribute }
func (a *attribute) SetValue(v string) error {
	a.Value = v
	return nil
}

type attrNode struct {
	*attribute
	*node
}

func (a attrNode) SetValue(v string) error { return a.attribute.SetValue(v) }
func (a attrNode) Name() xml.Name          { return a.attribute.Attr.Name }

func newAttribute(a xml.Attr) *node { return &node{value: &attribute{a}} }

// attrNode and *attrNode must both implement Attr
var _ Attr = attrNode{}
var _ Attr = &attrNode{}
