package modules

import (
	"reflect"
	"testing"

	xml "github.com/andaru/flexml"

	"github.com/openconfig/goyang/pkg/yang"
)

func Test_importError_Error(t *testing.T) {
	type fields struct {
		msg  string
		path string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"without path", fields{msg: "foo"}, "foo"},
		{"with path", fields{"foo", "/bar"}, "/bar: foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := importError{
				msg:  tt.fields.msg,
				path: tt.fields.path,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("importError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetYANGPath(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"empty", args{}},
		{"two paths", args{[]string{"/foo", "/bar"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetYANGPath(tt.args.paths...)
			if len(yang.Path) != len(tt.args.paths) {
				t.Errorf("got yang.Path len = %d, want %d", len(yang.Path), len(tt.args.paths))
				return
			}

			for i, path := range yang.Path {
				if tt.args.paths[i] != path {
					t.Errorf("got path %d = %s, want %s", i, path, tt.args.paths[i])
				}
			}
		})
	}
}

func TestCollection_Import(t *testing.T) {
	type args struct {
		moduleName string
	}
	tests := []struct {
		name    string
		paths   []string
		args    args
		wantErr bool
	}{
		{
			name:    "no paths",
			wantErr: true,
		},
		{
			name:    "module name to import is a path",
			paths:   []string{"..."},
			args:    args{"foo.yang"},
			wantErr: true,
		},
		{
			name:    "module name not found",
			paths:   []string{"..."},
			args:    args{"module-name-not-found"},
			wantErr: true,
		},
		{
			name:  "ok import",
			paths: []string{"..."},
			args:  args{"test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection()
			SetYANGPath(tt.paths...)
			if err := c.Import(tt.args.moduleName); (err != nil) != tt.wantErr {
				t.Errorf("Collection.Import() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollection_ImportAll(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  []error
	}{
		{"no paths", nil, nil},
		{"bogus path", []string{"does-not-exist"}, []error{
			importError{"does-not-exist", "lstat does-not-exist: no such file or directory"}}},
		{"ok paths", []string{".", "..."}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCollection()
			SetYANGPath(tt.paths...)
			if got := c.ImportAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collection.ImportAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_Process(t *testing.T) {
	tests := []struct {
		path string
		want []error
	}{
		{".", nil},
		{"path-does-not-exist", nil},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			SetYANGPath(tt.path)
			c := NewCollection()
			if err := c.ImportAll(); err == nil {

			}
			if got := c.Process(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collection.Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_ModuleEntry(t *testing.T) {
	type args struct {
		name string
	}
	c := NewCollection()
	tests := []struct {
		name    string
		args    args
		paths   []string
		wantFn  func() *yang.Entry
		wantErr bool
	}{
		{
			name:    "no paths",
			wantErr: true,
		},
		{
			name:    "no such module",
			paths:   []string{"..."},
			args:    args{"does not exist"},
			wantErr: true,
		},
		{
			name:   "test module",
			paths:  []string{"..."},
			args:   args{"test"},
			wantFn: func() *yang.Entry { return yang.ToEntry(c.ms.Modules["test"]) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetYANGPath(tt.paths...)
			if len(tt.paths) > 0 {
				c.ImportAll()
				c.Process()
			}
			got, err := c.ModuleEntry(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collection.Module() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if f := tt.wantFn; f != nil {
				want := f()
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Collection.Module() = %#v, want %#v", got, want)
				}
			}
		})
	}
}

func TestCollection_RootEntry(t *testing.T) {
	type args struct {
		name xml.Name
	}
	c := NewCollection()
	tests := []struct {
		name    string
		paths   []string
		args    args
		wantFn  func() *yang.Entry
		wantErr bool
	}{
		// failing cases
		{
			name:    "no paths, process not called",
			wantErr: true,
		},
		{
			name:    "incorrect namespace",
			paths:   []string{"testdata"},
			args:    args{xml.Name{Space: "test", Local: "system"}},
			wantErr: true,
		},
		{
			name:    "no such child in module",
			paths:   []string{"testdata"},
			args:    args{xml.Name{Space: "urn:opr8:modules:test:test", Local: "does-not-exist"}},
			wantErr: true,
		},

		// passing cases
		{
			name:   "system container in test module",
			paths:  []string{"testdata"},
			args:   args{xml.Name{Space: "urn:opr8:modules:test:test", Local: "system"}},
			wantFn: func() *yang.Entry { return yang.ToEntry(c.ms.Modules["test"]).Dir["system"] },
		},
	}
	for _, tt := range tests {
		SetYANGPath(tt.paths...)
		if len(tt.paths) > 0 {
			c.ImportAll()
			c.Process()
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.RootEntry(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collection.RootEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFn != nil {
				if want := tt.wantFn(); !reflect.DeepEqual(got, want) {
					t.Errorf("Collection.RootEntry() = %#v, want %#v", got, want)
				}
			}
		})
	}
}

func TestCollection_IterLatest(t *testing.T) {
	type fields struct {
		ms        *yang.Modules
		processed bool
	}
	type args struct {
		f func(*yang.Module) error
	}
	ms := yang.NewModules()
	SetYANGPath(".")

	found := false

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"find latest test module",
			fields{ms: ms},
			args{func(m *yang.Module) error {
				if pfx := m.Prefix; pfx != nil {
					if "test" == pfx.Name {
						if len(m.Revision) > 0 && m.Revision[0].Name == "2017-01-01" {
							found = true
						}
					}
				}
				return nil
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found = false
			c := &Collection{
				ms:        tt.fields.ms,
				processed: tt.fields.processed,
			}
			if errs := c.ImportAll(); len(errs) > 0 {
				t.Fatalf("Collection.ImportAll() %d errors; first error: %s", len(errs), errs[0])
			}

			if err := c.IterLatest(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("Collection.IterLatest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !found {
				t.Error("Collection.IterLatest() did not find wanted module")
			}
		})
	}
}

func Test_expandYANGPath(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"empty", args{}, nil},
		{"current directory", args{[]string{"."}}, []string{"."}},
		{"explicit current and testdata", args{[]string{".", "testdata"}}, []string{".", "testdata"}},
		{"recursive", args{[]string{"..."}}, []string{".", "testdata"}},
		{"current directory recursive", args{[]string{"./..."}}, []string{".", "testdata"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandYANGPath(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandYANGPath() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCollection_ModulesLen(t *testing.T) {
	tests := []struct {
		name          string
		paths         []string
		wantLength    int
		wantLengthRaw int
	}{
		{"no paths", nil, 0, 0},
		{"one path", []string{"testdata"}, 1, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetYANGPath(tt.paths...)
			c := NewCollection()
			c.ImportAll()
			if gotLength := c.ModulesLen(); gotLength != tt.wantLength {
				t.Errorf("Collection.ModulesLen() = %v, want %v", gotLength, tt.wantLength)
			}
			if gotLength := len(c.ms.Modules); gotLength != tt.wantLengthRaw {
				t.Errorf("Collection.len(.ms.Modules) = %v, want %v", gotLength, tt.wantLengthRaw)
			}
		})
	}
}
