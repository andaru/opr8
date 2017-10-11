package dom

import (
	xml "github.com/andaru/flexml"
)

// CharacterData interface extends Node with a set of attributes and methods for
// accessing character data in the DOM. For clarity this set is defined here
// rather than on each object that uses these attributes and methods. No DOM
// objects correspond directly to CharacterData, though Text and others do
// inherit the interface from it. All offsets in this interface start from 0.
type CharacterData interface {
	// Empty returns true if the node's character data is empty
	Empty() bool
	// Data returns the node's character data
	Data() string
	// SetData sets the node's character data, replacing any existing data
	SetData(arg string) error
	// AppendData appends the provided argument to the node's character data
	AppendData(arg string) error
	// InsertData inserts the provided argument at offset in the character data.
	InsertData(offset int, arg string) error
	// DeleteData deletes count runes of character data at offset. If offset
	// plus count exceeds the node's character data length, ErrIndexSize is
	// returned.
	DeleteData(offset, count int) error
	// ReplaceData replaces count runes of character data. If offset plus count
	// exceeds the node's character data length or arg is longer than count,
	// ErrIndexSize is returned.
	ReplaceData(offset, count int, arg string) error
}

// Text interface inherits from CharacterData and represents the textual content
// (termed character data in XML) of an Element or Attr. If there is no markup
// inside an element's content, the text is contained in a single object
// implementing the Text interface that is the only child of the element. If
// there is markup, it is parsed into the information items (elements, comments,
// etc.)  and Text nodes that form the list of children of the element.
//
// When a document is first made available via the DOM, there is only one Text
// node for each block of text. Users may create adjacent Text nodes that
// represent the contents of a given element without any intervening markup, but
// should be aware that there is no way to represent the separations between
// these nodes in XML or HTML, so they will not (in general) persist between DOM
// editing sessions. The normalize() method on Node merges any such adjacent
// Text objects into a single node for each block of text.
type Text interface {
	Node
	CharacterData

	// TODO: implement
	// SplitText(offset int) (Text, error)
}

type text struct {
	value []byte
}

type textNode struct {
	*text
	*node
}

func (t text) String() string               { return string(t.value) }
func (t text) nodeType() NodeType           { return NodeTypeText }
func (t text) Empty() bool                  { return len(t.value) == 0 }
func (t text) Data() string                 { return string(t.value) }
func (t text) charData() xml.CharData       { return xml.CharData(t.value).Copy() }
func (t *text) SetValue(value string) error { return t.SetData(value) }

func (t *text) SetData(arg string) error {
	t.value = t.value[:]
	t.value = []byte(arg)
	return nil
}

func (t *text) DeleteData(offset, count int) error {
	if count < 0 || offset+count > len(t.value) {
		return ErrIndexSize
	}
	t.value = append(t.value[:offset], t.value[offset+count:]...)
	return nil
}

func (t *text) ReplaceData(offset, count int, arg string) error {
	if count < 0 || offset+count > len(t.value) || len(arg) < count {
		return ErrIndexSize
	}
	copy(t.value[offset:], arg[:count])
	return nil
}

func (t *text) AppendData(arg string) error {
	t.value = append(t.value, []byte(arg)...)
	return nil
}

func (t *text) InsertData(offset int, arg string) error {
	if offset >= 0 && offset < len(t.value) {
		t.value = append(t.value[:offset], append([]byte(arg), t.value[offset:]...)...)
		return nil
	}
	return ErrIndexSize
}

func (t textNode) SetValue(v string) error { return t.text.SetData(v) }

func newText(cd xml.CharData) *node { return &node{value: &text{cd}} }

// textNode and *textNode must both implement Text
var _ Text = &textNode{}
var _ Text = textNode{}
