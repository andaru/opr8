package dom

import (
	xml "github.com/andaru/flexml"
)

// NodeIterator is the interface to node iterators.
type NodeIterator interface {
	NodeProvider
	ParentNodeProvider
	SiblingProvider

	// Equal returns true if the iterator is equivalent to the
	// provided iterator, meaning it has the same parent and same
	// current node position.
	Equal(NodeIterator) bool
}

// NewChildIterator returns a new NodeIterator parent's child nodes.
func NewChildIterator(parent Node) NodeIterator {
	return &nodeIterator{parent, nil, func(Node) bool { return true }}
}

// NewChildNamedIterator returns a new NodeIterator over the children of
// parent whose Name is name.
func NewChildNamedIterator(parent Node, name xml.Name) NodeIterator {
	return &nodeIterator{parent, nil, func(n Node) bool { return name == n.Name() }}
}

// NewChildFilteringIterator returns a new filtered NodeIterator over the
// children of parent matching the filter.
func NewChildFilteringIterator(parent Node, filter func(Node) bool) NodeIterator {
	return &nodeIterator{parent, nil, filter}
}

func (it *nodeIterator) Equal(o NodeIterator) bool {
	if ni, ok := o.(*nodeIterator); ok {
		return it.wrap == ni.wrap && it.parent == ni.parent
	}
	return false
}

type nodeIterator struct {
	parent Node
	wrap   Node
	match  func(n Node) bool
}

func (it *nodeIterator) Node() Node {
	if n := it.wrap; n != nil {
		return n
	}
	return nil
}

func (it *nodeIterator) Parent() Node {
	if p := it.parent; p != nil {
		return p
	}
	return nil
}

func (it *nodeIterator) NextSibling() Node {
	if it.wrap == nil {
		it.wrap = it.parent.FirstChild()
	} else {
		it.wrap = it.wrap.NextSibling()
	}
	for cur := it.wrap.nodePtr(); cur != nil; cur = cur.nextSib {
		if it.match(cur) {
			it.wrap = cur
			return cur
		}
	}
	it.wrap = nil
	return nil
}

func (it *nodeIterator) PreviousSibling() Node {
	if it.wrap == nil {
		it.wrap = it.parent.LastChild()
	} else {
		it.wrap = it.wrap.PreviousSibling()
	}
	for cur := it.wrap.nodePtr(); cur.nextSib != nil; cur = cur.prevSib {
		if it.match(cur) {
			it.wrap = cur
			return cur
		}
	}
	it.wrap = nil
	return nil
}
