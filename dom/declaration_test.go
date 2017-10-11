package dom

import (
	"reflect"
	"testing"

	xml "github.com/andaru/flexml"
)

func Test_kvPairs(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name   string
		args   args
		wantKv []string
	}{
		{"entirely bogus single", args{[]byte(`za0p`)}, nil},
		{"entirely bogus double", args{[]byte(`za0p p0w`)}, nil},
		{"entirely bogus triple", args{[]byte(`foo = "bar"`)}, nil},
		{"mixed bogus", args{[]byte(`za0p a='a' p="p"0wz! `)}, []string{"a", "a", "p", "p"}},
		{"single pair unquoted nomatch", args{[]byte(`foo=bar`)}, nil},
		{"single pair double quoted", args{[]byte(`foo="bar"`)}, []string{"foo", "bar"}},
		{"single pair single quoted", args{[]byte(`foo='bar'`)}, []string{"foo", "bar"}},
		{"two pair single quoted", args{[]byte(`foo='bar' qux='zap'`)}, []string{"foo", "bar", "qux", "zap"}},
		{"two pair double quoted", args{[]byte(`foo="bar" qux="zap"`)}, []string{"foo", "bar", "qux", "zap"}},
		{"two pair mixed quoted", args{[]byte(`foo='bar' qux="zap"`)}, []string{"foo", "bar", "qux", "zap"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotKv := kvPairs(tt.args.input); !reflect.DeepEqual(gotKv, tt.wantKv) {
				t.Errorf("kvPairs() = %#v, want %#v", gotKv, tt.wantKv)
			}
		})
	}
}

func Test_declaration_nodeType(t *testing.T) {
	type fields struct {
		ProcInst xml.ProcInst
	}
	tests := []struct {
		name   string
		fields fields
		want   NodeType
	}{
		{"always NodeTypeDeclaration", fields{xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0"`)}}, NodeTypeDeclaration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := declaration{
				ProcInst: tt.fields.ProcInst,
			}
			if got := pi.nodeType(); got != tt.want {
				t.Errorf("declaration.nodeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_declaration_Target(t *testing.T) {
	type fields struct {
		ProcInst xml.ProcInst
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"xml", fields{xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0"`)}}, "xml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := declaration{
				ProcInst: tt.fields.ProcInst,
			}
			if got := pi.Target(); got != tt.want {
				t.Errorf("declaration.Target() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_declaration_Inst(t *testing.T) {
	type fields struct {
		ProcInst xml.ProcInst
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"version", fields{xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0"`)}}, `version="1.0"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := declaration{
				ProcInst: tt.fields.ProcInst,
			}
			if got := pi.Inst(); got != tt.want {
				t.Errorf("declaration.Inst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newDeclaration(t *testing.T) {
	type args struct {
		pi xml.ProcInst
	}
	tests := []struct {
		name string
		args args
		want *node
	}{
		{"xml", args{xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0"`)}}, newDeclaration(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0"`)})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDeclaration(tt.args.pi); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDeclaration() = %v, want %v", got, tt.want)
			}
		})
	}
}
