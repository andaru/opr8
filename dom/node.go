package dom

import (
	"fmt"

	xml "github.com/andaru/flexml"
	"github.com/pkg/errors"
)

// Node interface is the primary datatype for the entire Document Object Model.
type Node interface {
	Namer
	Valuer
	ValueSetter
	// ChildValue returns the content value of the first Text child
	// node. Returns the empty string if node has no NodeTypeText
	// child nodes.
	ChildValue() string

	NodeType() NodeType
	// Parent returns the node's parent. This value will be nil if
	// Node represents a Document or is the root of a disconnected
	// subtree.
	Parent() Node

	// OwnerDocument returns the node's owning Document node. This
	// returns nil if the node is "disconnected".
	OwnerDocument() Document

	// FirstChild returns the node's first child node, or nil if there
	// are no children.
	FirstChild() Node
	// LastChild returns the node's last child node, or nil if the
	// node has no children.
	LastChild() Node
	// NextSibling returns the node's next sibling, or nil if the node
	// is the last child of its parent.
	NextSibling() Node
	// PreviousSibling returns the node's previous sibling, or nil if
	// the node is the first child of its parent.
	PreviousSibling() Node

	ChildrenByName(xml.Name) []Node
	ChildByName(xml.Name) Node

	// AppendChild appends the provided Node as a child. Returns an
	// error if adding such a child node to this node would be illegal
	// by the DOM's rules on tree layout.
	AppendChild(Node) error
	// PrependChild prepends the provided Node as a child. Returns an
	// error if adding such a child node to this node would be illegal
	// by the DOM's rules on tree layout.
	PrependChild(Node) error
	// InsertChildAfter inserts the provided child node after ref. If
	// ref is nil, append child instead. Returns an error if adding
	// such a child node to this node would be illegal by the DOM's
	// rules on tree layout.
	InsertChildAfter(child, ref Node) error
	// InsertChildBefore inserts the provided child node before ref.
	// If ref is nil, prepend child instead. Returns an error if
	// adding such a child node to this node would be illegal by the
	// DOM's rules on tree layout.
	InsertChildBefore(child, ref Node) error

	nodePtr
}

// NodeProvider is an interface providing a context Node.
type NodeProvider interface {
	// Node returns the context node.
	//
	// This value typically represents the context node in a document
	// query or a node iterator's current position.
	Node() Node
}

// ParentNodeProvider is an interface providing a context Node's parent Node.
type ParentNodeProvider interface {
	// Parent returns the context node's parent, which may be nil if
	// the context node is a Document (i.e., its NodeType is
	// NodeTypeDocument), or the context node is the root of a subtree
	// not connected to a document.
	Parent() Node
}

// OwnerDocumentProvider is an interface providing a Node's owner document.
type OwnerDocumentProvider interface {
	// OwnerDocument returns the node's owning Document node. This
	// returns nil if the node is "disconnected".
	OwnerDocument() Document
}

// SiblingProvider is an interface providing access to the next and
// previous siblings of a context Node. This supports both Node an
// Attr siblings.
type SiblingProvider interface {
	// NextSibling returns the context node's next sibling, or nil if
	// there are no following siblings.
	NextSibling() Node
	// PreviousSibling returns the context node's previous sibling, or
	// nil if there are no previous siblings.
	PreviousSibling() Node
}

// NodeSet is a collection of Node
type NodeSet []Node

// Namer is an object with a XML name.
//
// Examples of such types in the DOM include Element and Attr nodes.
type Namer interface {
	// Name returns the object's XML name
	Name() xml.Name
}

// Valuer is a node which can report a string value
type Valuer interface {
	Value() string
}

// ValueSetter is a node which permits setting of its value as a string
type ValueSetter interface {
	SetValue(string) error
}

// ValueSetterBool is a node which permits setting of its value as a bool
type ValueSetterBool interface {
	SetValueBool(bool) error
}

// ValueSetterFloat64 is a node which permits setting a float64 value
type ValueSetterFloat64 interface {
	SetValueFloat64(float64) error
}

// ValueSetterInt is a node which permits setting an int value
type ValueSetterInt interface {
	SetValueInt(int) error
}

type nodePtr interface {
	nodePtr() *node
}

type nodeTyper interface {
	nodeType() NodeType
}

type node struct {
	parent           *node
	firstChild       *node
	nextSib, prevSib *node
	firstAttr        *node

	value nodeTyper // must not be nil
}

// CreateAttribute returns a new Attr node using the provided XML attribute.
func CreateAttribute(a xml.Attr) Attr { return newAttribute(a).asAttribute() }

// CreateText returns a new PCDATA text node using the provided data.
func CreateText(cd xml.CharData) Text { return newText(cd.Copy()).asText() }

// CreateElement returns a new Element node using the provided XML StartElement.
func CreateElement(se xml.StartElement) Element { return newStartElement(se).asElement() }

// CreateDocumentFragment returns a new Document Fragment Node with host.
func CreateDocumentFragment(host Element) DocumentFragment {
	return newDocumentFragment(host.(Node).nodePtr()).asFragment()
}

// CreateComment returns a new Comment node using the provided XML comment.
func CreateComment(c xml.Comment) Comment { return newComment(c.Copy()).asComment() }

// CreateProcessingInstruction returns a new Processing Instruction or
// Declaration node (if the ProcInst's target is "xml").
func CreateProcessingInstruction(pi xml.ProcInst) Node {
	if pi.Target == "xml" {
		return newDeclaration(pi.Copy())
	}
	return newProcInst(pi.Copy())
}

// CreateXMLDeclaration returns a Declaration node with the specified encoding.
func CreateXMLDeclaration(encoding string) Node {
	return newDeclaration(
		xml.ProcInst{
			Target: "xml",
			Inst:   []byte(fmt.Sprintf(`version="1.0" encoding="%s"`, encoding))})
}

func (n *node) Value() string {
	switch n.NodeType() {
	case NodeTypeText:
		return string(n.asText().text.value)
	case NodeTypeComment:
		return string(n.asComment().text.value)
	case NodeTypeAttribute:
		return n.asAttribute().Attr.Value
	}
	return ""
}

func (n *node) Format(f fmt.State, c rune) {
	switch c {
	case 'v':
		var as []string
		// gather info
		typ := n.NodeType()
		switch typ {
		case NodeTypeElement:
			as = append(as, "xmlName", fmt.Sprintf("%#v", n.xmlName()))
		case NodeTypeText, NodeTypeComment:
			as = append(as, "value", fmt.Sprintf("%q", n.textValue()))
		}

		// include child node info and attribute info
		if f.Flag('+') {
			if n.parent != nil {
				as = append(as, "Parent", fmt.Sprintf("%v", n.parent))
			}
		}
		if f.Flag('+') || f.Flag('#') {
			if n.firstChild != nil {
				as = append(as, "FirstChild", fmt.Sprintf("%#v", n.firstChild))
			}
			if n.firstAttr != nil {
				as = append(as, "FirstAttribute", fmt.Sprintf("%#v", n.firstAttr))
			}
		}
		// write the format
		f.Write([]byte(fmt.Sprintf("%T{NodeType:%s", n, n.NodeType())))
		if len(as) > 0 {
			for i := 0; i < len(as)/2; i++ {
				f.Write([]byte(", "))
				f.Write([]byte(as[i*2]))
				f.Write([]byte(":"))
				f.Write([]byte(as[(i*2)+1]))
			}
		}
		f.Write([]byte("}"))
	default:
		f.Write([]byte(fmt.Sprintf("%#v", n)))
	}
}

func (n *node) SetValue(value string) error {
	if setter, ok := n.value.(ValueSetter); ok {
		return setter.SetValue(value)
	}
	return errors.Errorf("cannot call SetValue on a %s", n.NodeType())
}

func (n *node) Parent() Node {
	if n.parent != nil {
		return n.parent
	}
	return nil
}

func (n *node) FirstChild() Node {
	if n.firstChild == nil {
		return nil
	}
	return n.firstChild
}
func (n *node) LastChild() Node {
	if n.firstChild == nil {
		return nil
	}
	return n.firstChild.prevSib
}

func (n *node) OwnerDocument() Document {
	for it := n.parent; it != nil; it = it.parent {
		if doc, ok := it.value.(Document); ok {
			return doc
		}
	}
	return nil
}

func (n *node) NodeType() NodeType {
	if v := n.value; v != nil {
		return v.nodeType()
	}
	return NodeTypeNull
}

func (n *node) NextSibling() Node {
	if next := n.nextSib; next != nil {
		return next
	}
	return nil
}

func (n *node) PreviousSibling() Node {
	if n.prevSib.nextSib != nil {
		return n.prevSib
	}
	return nil
}

func (n *node) ChildByName(name xml.Name) Node {
	nodeset := n.ChildrenByName(name)
	if nodeset == nil {
		return nil
	}
	return nodeset[0]
}

func (n *node) ChildrenByName(name xml.Name) (nodeset []Node) {
	iterChildren(n, func(it *node) error {
		if namer, ok := it.value.(Namer); ok && namer.Name() == name {
			nodeset = append(nodeset, it)
		}
		return nil
	})
	return
}

func (n *node) AppendChild(child Node) error {
	if err := allowInsertChildErr(n.NodeType(), child.NodeType()); err != nil {
		return err
	}
	appendNode(child.nodePtr(), n)
	return nil
}

func (n *node) PrependChild(child Node) error {
	if err := allowInsertChildErr(n.NodeType(), child.NodeType()); err != nil {
		return err
	}
	prependNode(child.nodePtr(), n)
	return nil
}

func (n *node) InsertChildAfter(child, after Node) error {
	if err := allowInsertChildErr(n.NodeType(), child.NodeType()); err != nil {
		return err
	} else if after == nil {
		appendNode(child.nodePtr(), n)
		return nil
	} else if after.Parent() != n {
		return ErrHierarchyRequest
	}
	insertNodeAfter(child.nodePtr(), after.nodePtr())
	return nil
}

func (n *node) InsertChildBefore(child, before Node) error {
	if err := allowInsertChildErr(n.NodeType(), child.NodeType()); err != nil {
		return err
	} else if before == nil {
		prependNode(child.nodePtr(), n)
		return nil
	} else if before.Parent() != n {
		return ErrHierarchyRequest
	}
	insertNodeBefore(child.nodePtr(), before.nodePtr())
	return nil
}

func (n *node) defaultNamespace() (owner *node, attrValue string) {
	for it := n; it != nil; it = it.parent {
		if err := iterAttributes(it, func(n *node) error {
			if a := n.asAttribute(); a.Name() == xmlnsDefault {
				owner = it
				attrValue = a.Attr.Value
				return errors.New("matched")
			}
			return nil
		}); err != nil {
			return
		}
	}
	return nil, ""
}

func (n *node) nodePtr() *node                 { return n }
func (n *node) asAttribute() attrNode          { return attrNode{n.value.(*attribute), n} }
func (n *node) asComment() commentNode         { return commentNode{n.value.(*comment), n} }
func (n *node) asDocument() documentNode       { return documentNode{n.value.(*document), n} }
func (n *node) asElement() elementNode         { return elementNode{n.value.(*element), n} }
func (n *node) asText() textNode               { return textNode{n.value.(*text), n} }
func (n *node) asProcInst() procinstNode       { return procinstNode{n.value.(*procinst), n} }
func (n *node) asDeclaration() declarationNode { return declarationNode{n.value.(*declaration), n} }
func (n *node) asFragment() documentFragmentNode {
	return documentFragmentNode{n.value.(*documentFragment), n}
}

func (n *node) ChildValue() string {
	for it := n.firstChild; it != nil; it = it.nextSib {
		switch it.NodeType() {
		case NodeTypeText:
			return it.asText().String()
		}
	}
	return ""
}

func (n *node) Name() xml.Name {
	if namer, ok := n.value.(Namer); ok {
		return namer.Name()
	}
	return xml.Name{}
}

func (n *node) xmlName() xml.Name {
	switch value := n.value.(type) {
	case Namer:
		return value.Name()
	default:
		return xml.Name{}
	}
}

func (n *node) textValue() []byte {
	switch n := n.value.(type) {
	case *text:
		return n.value
	case *comment:
		return n.value
	}
	return nil
}

func (n *node) FirstAttribute() Attr {
	if n.firstAttr == nil {
		return nil
	}
	return n.firstAttr.asAttribute()
}

func (n *node) LastAttribute() Attr {
	if n.firstAttr == nil {
		return nil
	}
	return n.firstAttr.prevSib.asAttribute()
}

func (n *node) Attribute(name xml.Name) Attr {
	for it := n.firstAttr; it != nil; it = it.nextSib {
		if it.value.(*attribute).Attr.Name == name {
			return it.asAttribute()
		}
	}
	return nil
}

func (n *node) InsertAttributeAfter(a xml.Attr, after Attr) error {
	if err := allowInsertAttributeErr(n.NodeType()); err != nil {
		return err
	} else if after == nil {
		appendAttribute(newAttribute(a), n)
		return nil
	} else if after.Parent() != n {
		return ErrHierarchyRequest
	}
	insertAttributeAfter(newAttribute(a), after.(Node).nodePtr(), n)
	return nil
}

func (n *node) InsertAttributeBefore(a xml.Attr, before Attr) error {
	if err := allowInsertAttributeErr(n.NodeType()); err != nil {
		return err
	} else if before == nil {
		prependAttribute(newAttribute(a), n)
		return nil
	} else if before.Parent() != n {
		return ErrHierarchyRequest
	}
	insertAttributeBefore(newAttribute(a), before.(Node).nodePtr(), n)
	return nil
}

func (n *node) AppendAttribute(a xml.Attr) error {
	if err := allowInsertAttributeErr(n.NodeType()); err != nil {
		return err
	}
	appendAttribute(newAttribute(a), n)
	return nil
}

func (n *node) PrependAttribute(a xml.Attr) error {
	if err := allowInsertAttributeErr(n.NodeType()); err != nil {
		return err
	}
	prependAttribute(newAttribute(a), n)
	return nil
}

func appendNode(child, parent *node) {
	child.parent = parent
	if head := parent.firstChild; head != nil {
		tail := head.prevSib
		tail.nextSib = child
		child.prevSib = tail
		head.prevSib = child
	} else {
		parent.firstChild = child
		child.prevSib = child
	}
}

func prependNode(child, parent *node) {
	child.parent = parent
	head := parent.firstChild
	if head != nil {
		child.prevSib = head.prevSib
		head.prevSib = child
	} else {
		child.prevSib = child
	}
	child.nextSib = head
	parent.firstChild = child
}

func insertNodeBefore(child, before *node) {
	parent := before.parent
	child.parent = before
	if before.prevSib != nil {
		before.prevSib.nextSib = child
	} else {
		parent.firstChild.prevSib = child
	}
	child.prevSib = before.prevSib
	child.nextSib = before
	before.prevSib = child
}

func insertNodeAfter(child, after *node) {
	parent := after.parent
	child.parent = parent
	if next := after.nextSib; next != nil {
		next.prevSib = child
	} else {
		parent.firstChild.prevSib = child
	}
	child.nextSib = after.nextSib
	child.prevSib = after
	after.nextSib = child
}

func allowInsertChild(parent, child NodeType) bool {
	if parent == NodeTypeNull || child == NodeTypeNull {
		return false
	} else if parent != NodeTypeDocument && parent != NodeTypeElement {
		return false
	} else if child == NodeTypeDocument || child == 0 {
		return false
	} else if parent != NodeTypeDocument && (child == NodeTypeDocumentType || child == NodeTypeDeclaration) {
		return false
	}
	return true
}

func allowInsertChildErr(parent, child NodeType) error {
	if !allowInsertChild(parent, child) {
		return errors.Wrapf(ErrHierarchyRequest, "parent node type %s may not have a %s child", parent, child)
	}
	return nil
}

func allowInsertAttribute(parent NodeType) bool {
	return parent == NodeTypeElement || parent == NodeTypeDeclaration
}

func allowInsertAttributeErr(parent NodeType) error {
	if !allowInsertAttribute(parent) {
		return errors.Wrapf(ErrHierarchyRequest, "parent node type %s may not have a %s child", parent, NodeTypeAttribute)
	}
	return nil
}

func allowMove(parent, child *node) bool {
	if !allowInsertChild(parent.NodeType(), child.NodeType()) {
		return false
	}

	if parent.OwnerDocument() != child.OwnerDocument() {
		return false
	}

	for cur := parent; cur != nil; cur = cur.parent {
		if cur == child {
			return false
		}
	}

	return true
}

func appendAttribute(attr, parent *node) {
	if head := parent.firstAttr; head != nil {
		tail := head.prevSib
		tail.nextSib = attr
		attr.prevSib = tail
		head.prevSib = attr
	} else {
		parent.firstAttr = attr
		attr.prevSib = attr
	}
}

func prependAttribute(attr, parent *node) {
	head := parent.firstAttr
	if head != nil {
		attr.prevSib = head.prevSib
		head.prevSib = attr
	} else {
		attr.prevSib = attr
	}
	attr.nextSib = head
	parent.firstAttr = attr
}

func insertAttributeAfter(attr, place, parent *node) {
	// attr and place must be an attribute
	if place.value.nodeType() != NodeTypeAttribute || attr.value.nodeType() != NodeTypeAttribute {
		panic(errors.Errorf(
			"want NodeTypeAttribute for both, but got place.nodeType() == %q attr.nodeType() == %q",
			place.value.nodeType(), attr.value.nodeType()))
	}

	if pnext := place.nextSib; pnext != nil {
		pnext.prevSib = attr
	} else {
		parent.firstAttr.prevSib = attr
	}
	attr.nextSib = place.nextSib
	attr.prevSib = place
	place.nextSib = attr
}

func insertAttributeBefore(attr, place, parent *node) {
	// attr and place must be an attribute
	if place.value.nodeType() != NodeTypeAttribute || attr.value.nodeType() != NodeTypeAttribute {
		panic(errors.Errorf(
			"want NodeTypeAttribute for both, but got place.nodeType() == %q attr.nodeType() == %q",
			place.value.nodeType(), attr.value.nodeType()))
	}

	if pprev := place.prevSib; pprev != nil {
		pprev.nextSib = attr
	} else {
		parent.firstAttr = attr
	}
	attr.prevSib = place.prevSib
	attr.nextSib = place
	place.prevSib = attr
}

func removeAttribute(attr, parent *node) {
	// attr must be an attribute
	_ = attr.value.(Attr)

	if parent.firstAttr == nil {
		return
	}

	if next := attr.nextSib; next != nil {
		next.prevSib = attr.prevSib
	} else {
		parent.firstAttr.prevSib = attr.prevSib
	}
	if attr.prevSib.nextSib != nil {
		attr.prevSib.nextSib = attr.nextSib
	} else {
		parent.firstAttr = attr.nextSib
	}
	attr.prevSib = nil
	attr.nextSib = nil
}

func iterAttributes(n *node, fn func(*node) error) error {
	for it := n.firstAttr; it != nil; it = it.nextSib {
		if err := fn(it); err != nil {
			return err
		}
	}
	return nil
}

func iterChildren(n *node, fn func(*node) error) error {
	for it := n.firstChild; it != nil; it = it.nextSib {
		if err := fn(it); err != nil {
			return err
		}
	}
	return nil
}

var _ Node = &node{}
