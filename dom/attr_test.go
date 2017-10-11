package dom

import (
	"reflect"
	"testing"

	xml "github.com/andaru/flexml"
)

func Test_attribute_nodeType(t *testing.T) {
	type fields struct {
		a xml.Attr
	}
	tests := []struct {
		name   string
		fields fields
		want   NodeType
	}{
		{"always NodeTypeAttribute", fields{xml.Attr{Name: xml.Name{Local: "foo"}, Value: "bar"}}, NodeTypeAttribute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := attribute{
				tt.fields.a,
			}
			if got := a.nodeType(); got != tt.want {
				t.Errorf("attribute.NodeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_attribute_Name(t *testing.T) {
	type fields struct {
		a xml.Attr
	}
	tests := []struct {
		name   string
		fields fields
		want   xml.Name
	}{
		{"empty", fields{}, xml.Name{}},
		{"empty", fields{xml.Attr{}}, xml.Name{}},
		{"local only", fields{xml.Attr{Name: xml.Name{Local: "a"}}}, xml.Name{Local: "a"}},
		{"local and xmlns", fields{xml.Attr{Name: xml.Name{Local: "a", Space: "ns"}}}, xml.Name{Local: "a", Space: "ns"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := attribute{
				tt.fields.a,
			}
			if got := a.Name(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("attribute.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_attribute_Value(t *testing.T) {
	type fields struct {
		a xml.Attr
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty", fields{}, ""},
		{"empty", fields{xml.Attr{}}, ""},
		{"empty value", fields{xml.Attr{Name: xml.Name{Local: "a"}}}, ""},
		{"with value", fields{xml.Attr{Name: xml.Name{Local: "a"}, Value: "one"}}, "one"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := attribute{
				tt.fields.a,
			}
			if got := a.Value; got != tt.want {
				t.Errorf("attribute.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newAttribute(t *testing.T) {
	type args struct {
		a xml.Attr
	}
	tests := []struct {
		name string
		args args
		want *node
	}{
		{"empty", args{xml.Attr{}}, newAttribute(xml.Attr{})},
		{"empty value", args{xml.Attr{Name: xml.Name{Local: "a"}}}, newAttribute(xml.Attr{Name: xml.Name{Local: "a"}})},
		{"with value", args{xml.Attr{Name: xml.Name{Local: "a"}, Value: "one"}}, newAttribute(xml.Attr{Name: xml.Name{Local: "a"}, Value: "one"})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newAttribute(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_attribute_SetValue(t *testing.T) {
	type fields struct {
		a xml.Attr
	}
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "set empty",
			fields: fields{xml.Attr{Name: xml.Name{Local: "a"}}},
			args:   args{"foo"},
		},
		{
			name:   "replace",
			fields: fields{xml.Attr{Name: xml.Name{Local: "a"}, Value: "initial"}},
			args:   args{"replaced"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &attribute{
				tt.fields.a,
			}
			if err := a.SetValue(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("attribute.SetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := a.Value; got != tt.args.v {
				t.Errorf("attribute.Value() = %#v, want %#v", got, tt.args.v)
			}
		})
	}
}
