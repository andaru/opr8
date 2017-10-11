package dom

import (
	"bytes"

	xml "github.com/andaru/flexml"
)

// Declaration is out-of-band metadata for an XML implementation, such
// as a reader of the document.
type Declaration interface {
	CharacterData
	Target() string
}

type declaration struct {
	xml.ProcInst
}

type declarationNode struct {
	*declaration
	*node
}

func (d declaration) nodeType() NodeType { return NodeTypeDeclaration }
func (d declaration) Target() string     { return d.ProcInst.Target }
func (d declaration) Inst() string       { return string(d.ProcInst.Inst) }

func newDeclaration(pi xml.ProcInst) *node {
	decl := &declaration{ProcInst: pi.Copy()}
	n := &node{value: decl}
	pairs := kvPairs(decl.ProcInst.Inst)
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[(i*2)+1]
		appendAttribute(newAttribute(xml.Attr{Name: xml.Name{Local: k}, Value: v}), n)
	}
	return n
}

// kvPairs parses the `param="..."` or `param='...'` value out of the provided
// string, returning a slice of key followed by value pairs for all successfully
// parsed entries.
func kvPairs(input []byte) (kv []string) {
	for _, field := range bytes.Fields(input) {
		idx := bytes.IndexRune(field, '=')
		if idx == -1 {
			continue
		}
		k, v := field[:idx], field[idx+1:]
		if len(v) == 0 || (v[0] != '\'' && v[0] != '"') {
			continue
		} else if idx = bytes.IndexRune(v[1:], rune(v[0])); idx != -1 {
			kv = append(kv, string(k), string(v[1:idx+1]))
		}
	}
	return
}
