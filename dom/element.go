package dom

import xml "github.com/andaru/flexml"

// Prefixer is a provider of an XML prefix relevant to the namespace of the
// object at hand.
type Prefixer interface {
	// Prefix returns the prefix referring to this object's namespace
	Prefix() string
}

// Element nodes are simply known as elements.
//
// Elements have an associated namespace, namespace prefix, local name, custom
// element state, custom element definition, is value. When an element is
// created, all of these values are initialized.
//
// An element’s qualified name is its local name if its namespace prefix is
// null, and its namespace prefix, followed by ":", followed by its local name,
// otherwise.
type Element interface {
	Node
	Prefixer

	AttributeProvider
}

type element struct {
	name   xml.Name
	prefix string
}

func (e element) nodeType() NodeType { return NodeTypeElement }
func (e element) Name() xml.Name     { return e.name }
func (e element) Prefix() string     { return e.prefix }

type elementNode struct {
	*element
	*node
}

func (e elementNode) nodePtr() *node { return e.node }
func (e elementNode) Name() xml.Name { return e.element.name }

// elementNode and *elementNode must both implement Element
var _ Element = &elementNode{}
var _ Element = elementNode{}
