package modules

import (
	"os"
	"path/filepath"
	"strings"

	xml "github.com/andaru/flexml"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/pkg/errors"
)

// Package modules provides a YANG module (schema) collection

// Collection is a YANG module collection
type Collection struct {
	ms        *yang.Modules
	processed bool
}

// SetYANGPath sets the YANG import path. Each path in paths is a
// directory to search when importing YANG modules either directly or
// when referenced by other modules during import. This must be
// called prior to NewCollection.
func SetYANGPath(paths ...string) {
	yang.Path = paths
}

// NewCollection returns a new YANG module collection. Prior to
// creating a collection, YANG paths must have been set using
// SetYANGPath.
func NewCollection() *Collection { return &Collection{ms: yang.NewModules()} }

func (c *Collection) Raw() *yang.Modules { return c.ms }

func (c *Collection) GetSchemaNode(schemaNodeID string) (e *yang.Entry) {
	if match := c.IterLatest(func(m *yang.Module) error {
		if mod := yang.ToEntry(m); mod != nil && mod.Find(schemaNodeID) != nil {
			return errors.New("match")
		}
		return nil
	}); match != nil {
		return e
	}
	return nil
}

// Import imports a module by its module name. Process must be called
// before calls to ModuleEntry after this returns.
func (c *Collection) Import(moduleName string) error {
	if len(yang.Path) == 0 {
		return errors.New("no module paths to search for YANG modules, use SetYANGPath")
	}
	if c.ms.Modules[moduleName] != nil {
		return nil
	}
	if strings.HasSuffix(moduleName, ".yang") || strings.Contains(moduleName, string(os.PathSeparator)) {
		return errors.Errorf("received invalid module name %s", moduleName)
	}
	err := c.ms.Read(moduleName)
	if err == nil {
		c.processed = false
	}
	return err
}

func (c *Collection) ReadString(moduleName string, data string) error {
	if c.ms.Modules[moduleName] != nil {
		return nil
	}
	if strings.HasSuffix(moduleName, ".yang") || strings.Contains(moduleName, string(os.PathSeparator)) {
		return errors.Errorf("received invalid module name %s", moduleName)
	}
	err := c.ms.Parse(data, moduleName)
	if err == nil {
		c.processed = false
	}
	return err

}

// ImportAll reads all YANG files found in the YANG path(s), returning
// any import errors. Process must be called before calls to
// ModuleEntry after this returns.
func (c *Collection) ImportAll() []error {
	var errs []error
	for _, root := range expandYANGPath(yang.Path) {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errs = append(errs, importError{path, err.Error()})
				return nil
			}
			if info.Mode().IsRegular() && strings.HasSuffix(path, ".yang") {
				if err := c.ms.Read(path); err != nil {
					errs = append(errs, importError{path, err.Error()})
				} else {
					// clear the processed flag as we've imported a
					// module potentially unforseen
					c.processed = false
				}
			}
			return nil
		})
	}
	return errs
}

// Process processes all modules previous read by Import or ImportAll,
// and must be called before collection accessors, to ensure the
// schema Entry tree including all augmentations is built.
func (c *Collection) Process() []error {
	errs := c.ms.Process()
	c.processed = len(errs) == 0
	return errs
}

// ModulesLen returns the number of unique module names in the
// collection, excluding sub-modules.
func (c *Collection) ModulesLen() (length int) {
	for name := range c.ms.Modules {
		if !strings.Contains(name, "@") {
			length++
		}
	}
	return
}

// ModuleEntry returns the YANG schema node entry for the given YANG module
// name, if it exists in the collection. An error is returned if no
// such module exists in the collection or the collection is not ready
// to be read.
func (c *Collection) ModuleEntry(name string) (*yang.Entry, error) {
	if !c.processed {
		return nil, errors.New("must call Process first")
	}
	if mod := c.ms.Modules[name]; mod != nil {
		return yang.ToEntry(mod), nil
	}
	return nil, errors.New("not found")
}

// RootEntry scans the latest version of the module matching the
// name's Space field for child entries matching the name's Local
// field. If no such module is found, or no such child is found within
// the module, an error is returned.
func (c *Collection) RootEntry(name xml.Name) (*yang.Entry, error) {
	if !c.processed {
		return nil, errors.New("must call Process first")
	}
	var entry *yang.Entry
	if stopped := c.IterLatest(func(mod *yang.Module) error {
		// only search modules with a matching namespace
		if mod.Namespace == nil || mod.Namespace.Name != name.Space {
			return nil
		}
		for local, e := range yang.ToEntry(mod).Dir {
			if name.Local == local {
				entry = e
				return errors.New("stop")
			}
		}
		return nil
	}); stopped != nil {
		return entry, nil
	}

	return nil, errors.New("not found")
}

// IterLatest iterates oves the latest version of all YANG modules in
// the underlying module collection.
func (c *Collection) IterLatest(f func(*yang.Module) error) error {
	for name, mod := range c.ms.Modules {
		// only consider "latest" versions, those without a '@' char.
		if !strings.Contains(name, "@") {
			if err := f(mod); err != nil {
				return err
			}
		}
	}
	return nil
}

func expandYANGPath(paths []string) []string {
	var result []string
	var roots []string
	for _, path := range paths {
		if "..." == filepath.Base(path) {
			roots = append(roots, filepath.Dir(path))
		} else {
			result = append(result, path)
		}
	}
	for _, root := range roots {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || strings.Contains(path, "/.git/") || strings.Contains(path, "/.hg/") || !info.IsDir() || (len(info.Name()) > 0 && info.Name()[0] == '_') {
				return nil
			}
			result = append(result, path)
			return nil
		})
	}
	return result
}

type importError struct {
	path string
	msg  string
}

func (e importError) Error() string {
	if e.path != "" {
		return e.path + ": " + e.msg
	}
	return e.msg
}
