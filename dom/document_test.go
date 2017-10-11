package dom

import (
	"context"
	"testing"

	xml "github.com/andaru/flexml"
	"github.com/stretchr/testify/assert"
)

func TestDocumentComplex(t *testing.T) {
	a := assert.New(t)
	_ = a
	type check func(Node) bool
	for _, tt := range []struct {
		name  string
		tree  func() Node
		check []check
	}{
		{
			name: "empty document",
			tree: func() Node {
				root := NewDocument(context.Background())
				return root
			},
			check: []check{
				func(x Node) bool { return a.NotNil(x) },
				func(x Node) bool { return a.Nil(x.FirstChild()) },
				func(x Node) bool { return a.Nil(x.LastChild()) },
			},
		},

		{
			name: "document with root element",
			tree: func() Node {
				root := NewDocument(context.Background())
				rootElem := CreateElement(xml.StartElement{Name: xml.Name{Local: "rootElement"}})
				root.AppendChild(rootElem)
				return root
			},
			check: []check{
				func(x Node) bool { return a.NotNil(x) },
				func(x Node) bool { return a.NotNil(x.FirstChild()) },
				func(x Node) bool { return a.NotNil(x.LastChild()) },
				func(x Node) bool { return a.Equal(x.FirstChild().Name(), xml.Name{Local: "rootElement"}) },
			},
		},

		{
			name: "document with declaration and element",
			tree: func() Node {
				root := NewDocument(context.Background())
				decl := CreateProcessingInstruction(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)})
				root.AppendChild(decl)
				element := CreateElement(xml.StartElement{Name: xml.Name{Local: "element"}})
				root.AppendChild(element)
				return root
			},
			check: []check{
				func(x Node) bool { return a.NotNil(x) },
				func(x Node) bool { return a.NotNil(x.FirstChild()) },
				func(x Node) bool { return a.NotNil(x.LastChild()) },
				func(x Node) bool { return a.NotEqual(x.FirstChild(), x.LastChild()) },
				func(x Node) bool { return a.Equal(x.FirstChild().NodeType(), NodeTypeDeclaration) },
				func(x Node) bool { return a.NotNil(x.FirstChild().NextSibling()) },
				func(x Node) bool { return a.Equal(x.FirstChild().NextSibling().NodeType(), NodeTypeElement) },
				func(x Node) bool { return a.Equal(x.FirstChild().NextSibling().Name(), xml.Name{Local: "element"}) },
				func(x Node) bool { return a.Equal(x.LastChild().Name(), xml.Name{Local: "element"}) },
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.tree()
			for i, ck := range tt.check {
				if ok := ck(root); !ok {
					t.Errorf("check %d failed", i)
				}
			}
		})
	}
}
