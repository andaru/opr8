package dom

import "testing"

func TestNodeType_String(t *testing.T) {
	tests := []struct {
		name string
		nt   NodeType
		want string
	}{
		{"invalid", NodeType(-1), "UNKNOWN"},
		{"NodeTypeNull", NodeTypeNull, "NULL"},
		{"NodeTypeAttribute", NodeTypeAttribute, "ATTRIBUTE_NODE"},
		{"NodeTypeElement", NodeTypeElement, "ELEMENT_NODE"},
		{"NodeTypeText", NodeTypeText, "TEXT_NODE"},
		{"NodeTypeCDATASection", NodeTypeCDATASection, "CDATA_SECTION_NODE"},
		{"NodeTypeEntityReference", NodeTypeEntityReference, "ENTITY_REFERENCE_NODE"},
		{"NodeTypeEntity", NodeTypeEntity, "ENTITY_NODE"},
		{"NodeTypeProcessingInstruction", NodeTypeProcessingInstruction, "PROCESSING_INSTRUCTION_NODE"},
		{"NodeTypeComment", NodeTypeComment, "COMMENT_NODE"},
		{"NodeTypeDocument", NodeTypeDocument, "DOCUMENT_NODE"},
		{"NodeTypeDocumentType", NodeTypeDocumentType, "DOCUMENT_TYPE_NODE"},
		{"NodeTypeDocumentFragment", NodeTypeDocumentFragment, "DOCUMENT_FRAGMENT_NODE"},
		{"NodeTypeNotation", NodeTypeNotation, "NOTATION_NODE"},
		{"NodeTypeDeclaration", NodeTypeDeclaration, "PROCESSING_INSTRUCTION_NODE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nt.String(); got != tt.want {
				t.Errorf("NodeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
