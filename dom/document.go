package dom

import (
	"context"

	"github.com/pkg/errors"
)

// Document interface in the DOM Core provides an interface to the
// list of entities that are defined for the document, and little else
// because the effect of namespaces and the various XML schema efforts
// on DTD representation are not clearly understood as of this
// writing.
type Document interface {
	Node
	// Context returns the Document context for access to document metadata.
	Context() context.Context
	// DocumentElement returns the Document's Element child, or nil if the
	// Document has no Element children.
	DocumentElement() Element
}

// DocumentFragment is a collection of zero or more child nodes.
//
// A fragment is not represented in the DOM node tree. Instead, the children of
// the document fragment appear as children of the fragment's host Element. A
// fragment behaves as a tree-order collection of nodes when referenced in DOM
// operations.
type DocumentFragment interface {
	Node
	// Host returns the Document Fragment's element node.
	Host() Element
	// SetHost sets the Document Fragment's host element.
	SetHost(Element) error
}

// NewDocument returns a new Document using the provided context to access
// document metadata. Pass context.Background() to specify no metadata.
func NewDocument(ctx context.Context) Document { return newDocument(ctx).asDocument() }

type document struct{ ctx context.Context }

type documentNode struct {
	*document
	*node
}

func (d *document) nodeType() NodeType       { return NodeTypeDocument }
func (d *document) Context() context.Context { return d.ctx }

func newDocument(ctx context.Context) *node {
	if ctx != nil {
		return &node{value: &document{ctx: ctx}}
	}
	return &node{value: &document{ctx: context.Background()}}
}

func (d documentNode) DocumentElement() Element {
	for it := d.node.firstChild; it != nil; it = it.nextSib {
		if it.NodeType() == NodeTypeElement {
			return it.asElement()
		}
	}
	return nil
}

type documentFragmentNode struct {
	*documentFragment
	*node
}

type documentFragment struct{}

func (d *documentFragment) nodeType() NodeType { return NodeTypeDocumentFragment }

func (d documentFragmentNode) Host() Element {
	if p := d.node.parent; p != nil && p.NodeType() == NodeTypeElement {
		return p.asElement()
	}
	return nil
}

func (d documentFragmentNode) SetHost(host Element) error {
	if d.node == nil {
		return errors.New("cannot set nil node")
	}
	d.node.parent = host.nodePtr()
	return nil
}

func newDocumentFragment(host *node) *node { return &node{parent: host, value: &documentFragment{}} }

// documentNode and *documentNode must both implement Document
var _ Document = &documentNode{}
var _ Document = documentNode{}

var _ DocumentFragment = &documentFragmentNode{}
var _ DocumentFragment = documentFragmentNode{}
