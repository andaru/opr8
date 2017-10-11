package dom

import (
	xml "github.com/andaru/flexml"
)

// Comment interface inherits from CharacterData and represents the content of a
// comment, i.e., all the characters between the starting '<!--' and ending
// '-->'. Note that this is the definition of a comment in XML, and, in
// practice, HTML, although some HTML tools may implement the full SGML comment
// structure.
type Comment interface {
	Node
	CharacterData
}

type commentNode struct {
	*comment
	*node
}

type comment struct{ text }

func (c comment) nodeType() NodeType { return NodeTypeComment }

func newComment(c xml.Comment) *node { return &node{value: &comment{text{xml.CharData(c.Copy())}}} }

var _ Comment = &commentNode{}
var _ Comment = commentNode{}
