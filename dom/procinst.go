package dom

import (
	xml "github.com/andaru/flexml"
)

// ProcessingInstruction is out-of-band metadata for an XML implementation, such
// as a reader of the document.
type ProcessingInstruction interface {
	CharacterData
	Target() string
}

type procinst struct {
	xml.ProcInst
}

type procinstNode struct {
	*procinst
	*node
}

func (pi *procinst) nodeType() NodeType { return NodeTypeProcessingInstruction }
func (pi *procinst) Target() string     { return pi.Target() }
func (pi *procinst) Inst() string       { return pi.Inst() }

func newProcInst(pi xml.ProcInst) *node { return &node{value: &procinst{pi.Copy()}} }
