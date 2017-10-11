package dom

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	xml "github.com/andaru/flexml"
)

func ExampleCreateAttribute() {
	attribute := CreateAttribute(xml.Attr{Name: xml.Name{Local: "id"}, Value: "foo"})

	fmt.Println(attribute.NodeType())
	fmt.Println(attribute.Name().Local)
	fmt.Println(attribute.Value())
	// Output:
	// ATTRIBUTE_NODE
	// id
	// foo
}

func ExampleCreateElement() {
	elem := CreateElement(xml.StartElement{Name: xml.Name{Local: "elem", Space: "ns"}})
	doc := NewDocument(context.Background())
	doc.AppendChild(elem)

	fmt.Println(elem.NodeType())
	fmt.Println(elem.Parent().NodeType())
	fmt.Println(elem.Name().Local)
	fmt.Println(elem.Name().Space)
	fmt.Println(elem.FirstChild() == nil)
	// Output:
	// ELEMENT_NODE
	// DOCUMENT_NODE
	// elem
	// ns
	// true
}

func ExampleCreateDocumentFragment() {
	host := CreateElement(xml.StartElement{Name: xml.Name{Local: "foo"}})
	fragment := CreateDocumentFragment(host)

	fmt.Println(fragment.NodeType())
	fmt.Println(fragment.Host().Name().Local)
	// Output:
	// DOCUMENT_FRAGMENT_NODE
	// foo
}

func ExampleNewDocument() {
	doc := NewDocument(context.Background())

	fmt.Println(doc.NodeType())
	fmt.Println(doc.FirstChild() == nil)
	fmt.Println(doc.LastChild() == nil)
	fmt.Println(doc.DocumentElement() == nil)

	elem := CreateElement(xml.StartElement{Name: xml.Name{Local: "elem", Space: "ns"}})
	doc.AppendChild(elem)

	fmt.Println(doc.DocumentElement() == nil)
	fmt.Println(doc.FirstChild() == nil)
	fmt.Println(doc.LastChild() == nil)
	fmt.Println(doc.DocumentElement().NodeType())
	fmt.Println(doc.DocumentElement().Name().Local)
	fmt.Println(doc.DocumentElement().Name().Space)
	fmt.Println(doc.DocumentElement().Name() == doc.FirstChild().Name())
	fmt.Println(doc.DocumentElement().Name() == doc.LastChild().Name())
	// Output:
	// DOCUMENT_NODE
	// true
	// true
	// true
	// false
	// false
	// false
	// ELEMENT_NODE
	// elem
	// ns
	// true
	// true
}

func ExampleCreateText() {
	text := CreateText(xml.CharData("hello"))

	fmt.Println(text.NodeType()) // TEXT_NODE
	fmt.Println(text.Value())    // hello

	text.SetValue("replace")
	fmt.Println(text.Value()) // replace

	text.AppendData("d")
	fmt.Println(text.Value()) // replaced

	text.ReplaceData(0, 6, "foobar")
	fmt.Println(text.Value()) // foobared

	text.DeleteData(1, 2)
	fmt.Println(text.Value()) // fbared
	text.InsertData(1, "u")
	fmt.Println(text.Value()) // fubared
	// Output:
	// TEXT_NODE
	// hello
	// replace
	// replaced
	// foobared
	// fbared
	// fubared
}

func ExampleCreateComment() {
	comment := CreateComment(xml.Comment("initial"))
	fmt.Println(comment.NodeType()) // COMMENT_NODE
	fmt.Println(comment.Value())    // initial
	comment.SetData("value now")
	fmt.Println(comment.Value()) // value now
	comment.SetValue("comment")
	fmt.Println(comment.Value()) // comment
	comment.AppendData("ed")
	fmt.Println(comment.Value()) // commented
	// Output:
	// COMMENT_NODE
	// initial
	// value now
	// comment
	// commented
}

func ExampleUnmarshaler_XMLReader() {
	input := `<foo>bar</foo>`
	reader := strings.NewReader(input)
	fmt.Printf("Input: %s\n", input)

	doc := NewDocument(context.Background())
	builder := NewBuilder(doc)
	un := NewUnmarshaler(builder)

	rn, err := un.XMLReader().ReadFrom(reader)
	if err != nil {
		fmt.Printf("read error: %v\n", err)
	}
	fmt.Printf("input length: %d\n", len(input))
	fmt.Printf("read length: %d\n", rn)

	m := NewMarshaler(doc)
	output := new(bytes.Buffer)
	wn, err := m.XMLWriter().WriteTo(output)
	if err != nil {
		fmt.Printf("write error: %v\n", err)
	}
	fmt.Printf("wrote length: %d\n", wn)
	fmt.Printf("XML: %s", output.String())
	// Output:
	// Input: <foo>bar</foo>
	// input length: 14
	// read length: 14
	// wrote length: 14
	// XML: <foo>bar</foo>
}

func ExampleUnmarshaler_JSONReader() {
	input := `{"foo": "bar"}`
	reader := strings.NewReader(input)
	fmt.Printf("Input: %s\n", input)

	doc := NewDocument(context.Background())
	builder := NewBuilder(doc)
	un := NewUnmarshaler(builder)
	rn, err := un.JSONReader().ReadFrom(reader)
	if err != nil {
		fmt.Printf("read error: %v\n", err)
	}
	fmt.Printf("input length: %d\n", len(input))
	fmt.Printf("read length: %d\n", rn)

	m := NewMarshaler(doc)
	output := new(bytes.Buffer)
	wn, err := m.XMLWriter().WriteTo(output)
	if err != nil {
		fmt.Printf("write error: %v\n", err)
	}
	fmt.Printf("wrote length: %d\n", wn)
	fmt.Printf("XML: %s\n", output.String())
	// Output:
	// Input: {"foo": "bar"}
	// input length: 14
	// read length: 14
	// wrote length: 14
	// XML: <foo>bar</foo>
}
