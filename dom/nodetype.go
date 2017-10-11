package dom

// NodeType is a code representing the type of the underlying object.
type NodeType int8

const (
	// NodeTypeNull is unused
	NodeTypeNull NodeType = iota
	// NodeTypeElement is an element node
	NodeTypeElement
	// NodeTypeAttribute is an attribute
	NodeTypeAttribute // historical
	// NodeTypeText is a leaf text node
	NodeTypeText
	// NodeTypeCDATASection is a leaf CDATA section node
	NodeTypeCDATASection // historical
	// NodeTypeEntityReference is an entity reference
	NodeTypeEntityReference // historical
	// NodeTypeEntity is an entity node
	NodeTypeEntity // historical
	// NodeTypeProcessingInstruction is a leaf XML processing instruction
	NodeTypeProcessingInstruction
	// NodeTypeComment represents a leaf comment node
	NodeTypeComment
	// NodeTypeDocument is the document root node
	NodeTypeDocument
	// NodeTypeDocumentType is a document type node
	NodeTypeDocumentType
	// NodeTypeDocumentFragment is a document fragment
	NodeTypeDocumentFragment
	// NodeTypeNotation is a notation node
	NodeTypeNotation // historical
	// NodeTypeDeclaration is a document declaration node, i.e.
	// '<?xml version="1.0">'
	NodeTypeDeclaration
)

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeNull:
		return "NULL"
	case NodeTypeElement: // = 1
		return "ELEMENT_NODE"
	case NodeTypeAttribute: // historical
		return "ATTRIBUTE_NODE"
	case NodeTypeText:
		return "TEXT_NODE"
	case NodeTypeCDATASection: // historical
		return "CDATA_SECTION_NODE"
	case NodeTypeEntityReference: // historical
		return "ENTITY_REFERENCE_NODE"
	case NodeTypeEntity: // historical
		return "ENTITY_NODE"
	case NodeTypeProcessingInstruction:
		return "PROCESSING_INSTRUCTION_NODE"
	case NodeTypeComment:
		return "COMMENT_NODE"
	case NodeTypeDocument:
		return "DOCUMENT_NODE"
	case NodeTypeDocumentType:
		return "DOCUMENT_TYPE_NODE"
	case NodeTypeDocumentFragment:
		return "DOCUMENT_FRAGMENT_NODE"
	case NodeTypeNotation: // historical
		return "NOTATION_NODE"
	case NodeTypeDeclaration:
		// declarations are ProcInst nodes
		return "PROCESSING_INSTRUCTION_NODE"
	default:
		return "UNKNOWN"
	}
}
