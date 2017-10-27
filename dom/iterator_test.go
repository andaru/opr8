package dom

import (
	"reflect"
	"testing"

	xml "github.com/andaru/flexml"
)

func TestNewChildIterator(t *testing.T) {
	type args struct {
		parent Node
	}
	tests := []struct {
		name string
		args args
		want NodeIterator
	}{
		{name: "constructor", args: args{nil}, want: NewChildIterator(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChildIterator(tt.args.parent); !got.Equal(tt.want) {
				t.Errorf("NewChildIterator() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNewChildNamedIterator(t *testing.T) {
	type args struct {
		parent Node
		name   xml.Name
	}
	tests := []struct {
		name string
		args args
		want NodeIterator
	}{
		{name: "constructor", args: args{nil, xml.Name{Local: "foo"}}, want: NewChildNamedIterator(nil, xml.Name{Local: "foo"})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChildNamedIterator(tt.args.parent, tt.args.name); !got.Equal(tt.want) {
				t.Errorf("NewChildNamedIterator() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNewChildFilteringIterator(t *testing.T) {
	type args struct {
		parent Node
		filter func(Node) bool
	}
	tests := []struct {
		name string
		args args
		want NodeIterator
	}{
		{name: "constructor", args: args{nil, func(Node) bool { return false }}, want: NewChildFilteringIterator(nil, func(Node) bool { return false })},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChildFilteringIterator(tt.args.parent, tt.args.filter); !got.Equal(tt.want) {
				t.Errorf("NewChildFilteringIterator() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_nodeIterator_Node(t *testing.T) {
	type fields struct {
		parent Node
		wrap   Node
		match  func(n Node) bool
	}
	tests := []struct {
		name   string
		fields fields
		want   Node
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := nodeIterator{
				parent: tt.fields.parent,
				wrap:   tt.fields.wrap,
				match:  tt.fields.match,
			}
			if got := it.Node(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeIterator.Node() = %v, want %v", got, tt.want)
			}
		})
	}
}

var testDataNodeIteratorTree = func() Node {
	doc := NewDocument(nil)
	root := CreateElement(xml.StartElement{Name: xml.Name{Local: "root"}})
	if err := doc.AppendChild(root); err != nil {
		panic(err)
	}
	for _, name := range []xml.Name{
		xml.Name{Local: "foo"},
		xml.Name{Local: "bar"},
		xml.Name{Local: "baz"},
		xml.Name{Local: "foo"},
		xml.Name{Local: "qux"},
	} {
		if err := root.AppendChild(CreateElement(xml.StartElement{Name: name})); err != nil {
			panic(err)
		}
	}
	return root
}

func Test_nodeIterator_Parent(t *testing.T) {
	doc := NewDocument(nil)
	root := CreateElement(xml.StartElement{Name: xml.Name{Local: "root"}})
	if err := doc.AppendChild(root); err != nil {
		panic(err)
	}
	for _, name := range []xml.Name{
		xml.Name{Local: "foo"},
		xml.Name{Local: "bar"},
		xml.Name{Local: "baz"},
		xml.Name{Local: "foo"},
		xml.Name{Local: "qux"},
	} {
		if err := root.AppendChild(CreateElement(xml.StartElement{Name: name})); err != nil {
			panic(err)
		}
	}

	type fields struct {
		parent Node
		wrap   Node
		match  func(n Node) bool
	}
	tests := []struct {
		name   string
		fields fields
		want   Node
	}{
		{
			fields: fields{parent: root},
			want:   root,
		},
		{
			fields: fields{parent: root, wrap: root.FirstChild()},
			want:   root,
		},
		{
			fields: fields{parent: root, wrap: root.LastChild()},
			want:   root,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := nodeIterator{
				parent: tt.fields.parent,
				wrap:   tt.fields.wrap,
				match:  tt.fields.match,
			}
			if got := it.Parent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeIterator.Parent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeIterator_NextSibling(t *testing.T) {
	doc := NewDocument(nil)
	root := CreateElement(xml.StartElement{Name: xml.Name{Local: "root"}})
	if err := doc.AppendChild(root); err != nil {
		panic(err)
	}
	for _, name := range []xml.Name{
		xml.Name{Local: "foo"},
		xml.Name{Local: "bar"},
		xml.Name{Local: "baz"},
		xml.Name{Local: "foo"},
		xml.Name{Local: "qux"},
	} {
		if err := root.AppendChild(CreateElement(xml.StartElement{Name: name})); err != nil {
			panic(err)
		}
	}

	type fields struct {
		parent Node
		wrap   Node
		match  func(n Node) bool
	}
	tests := []struct {
		name   string
		fields fields
		want   Node
	}{
		{
			fields: fields{
				parent: root,
				match: func(n Node) bool {
					return true
				},
			},
			want: root.FirstChild(),
		},
		{
			fields: fields{
				wrap:   root.FirstChild(),
				parent: root,
				match: func(n Node) bool {
					return n.Name().Local == "foo"
				},
			},
			want: root.LastChild().PreviousSibling(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := &nodeIterator{
				parent: tt.fields.parent,
				wrap:   tt.fields.wrap,
				match:  tt.fields.match,
			}
			if got := it.NextSibling(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeIterator.NextSibling() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeIterator_PreviousSibling(t *testing.T) {
	type fields struct {
		parent Node
		wrap   Node
		match  func(n Node) bool
	}
	tests := []struct {
		name   string
		fields fields
		want   Node
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := &nodeIterator{
				parent: tt.fields.parent,
				wrap:   tt.fields.wrap,
				match:  tt.fields.match,
			}
			if got := it.PreviousSibling(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeIterator.PreviousSibling() = %v, want %v", got, tt.want)
			}
		})
	}
}
