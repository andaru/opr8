package dom

import (
	"encoding/json"
	"fmt"
	"strings"

	xml "github.com/andaru/flexml"
	"github.com/pkg/errors"
)

// JSONDecoder is a streaming JSON decoder, emitting function calls to the
// TokenDecoder to populate a DOM document from an JSON source.
type JSONDecoder struct {
	*json.Decoder
	TokenDecoder

	error error
	stack jdStack
}

// Run executes the JSONDecoder, returning critical errors.
func (d *JSONDecoder) Run() error {
	return d.End(runJSDecoder(d, jsDecodeStart))
}

func runJSDecoder(d *JSONDecoder, first jsonDecoderFn) error {
	for state := jsDecodeStart; state != nil; state = state(d) {
	}
	return d.error
}

type jsonDecoderFn func(d *JSONDecoder) jsonDecoderFn

type jdState struct {
	jsonDecoderFn
	key string
}

func xmlNameToTag(n xml.Name) (tag string) {
	if tag = n.Space; tag != "" {
		tag += ":"
	}
	tag += n.Local
	return
}

func jsTagToName(input string) xml.Name {
	if idx := strings.Index(input, ":"); idx > 0 { // && idx < len(input)-1 {
		return xml.Name{Space: input[:idx], Local: input[idx+1:]}
	}
	return xml.Name{Local: input}
}

type jdStack struct{ d []jdState }

func (s *jdStack) push(jds jdState) { s.d = append(s.d, jds) }
func (s *jdStack) pop() (jds jdState) {
	if len(s.d) > 0 {
		jds, s.d = s.d[len(s.d)-1], s.d[:len(s.d)-1]
		return
	}
	return jdState{}
}

func (d *JSONDecoder) bail(err error) jsonDecoderFn {
	d.error = err
	return nil
}

func (d *JSONDecoder) ContextObject(name xml.Name) {
	d.stack.push(jdState{jsDecodeObject, xmlNameToTag(name)})
}

func jsDecodeStart(d *JSONDecoder) jsonDecoderFn {
	t, err := d.Token()
	if err != nil {
		return d.bail(err)
	}
	switch t := t.(type) {
	case json.Delim:
		switch t {
		case '{':
			return jsDecodeObject
		case '[':
			return jsDecodeArray
		}
	}
	return d.bail(errors.Errorf("unexpected JSON token (%T) %v", t, t))
}

func jsDecodeArray(d *JSONDecoder) jsonDecoderFn {
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return d.bail(err)
		}

		var next jsonDecoderFn
		prev := d.stack.pop()
		key := prev.key
		name := jsTagToName(prev.key)
		d.stack.push(prev)
		switch value := t.(type) {
		case json.Delim:
			switch value {
			case '{':
				next = jsDecodeObject
				d.stack.push(jdState{jsDecodeArray, key})
				err = d.StartElement(xml.StartElement{Name: name})
			default:
				err = errors.New("decoding of nested arrays unsupported")
			}
		case string:
			err = d.decode(name, value)
		case bool, float64:
			err = d.decode(name, fmt.Sprintf("%v", value))
		case json.Number:
			err = d.decode(name, value.String())
		case nil:
			err = d.decode(name, "")
		}
		if err != nil {
			return d.bail(err)
		}
		if next != nil {
			return next
		}
	}
	// consume the ending delimiter
	if _, err := d.Token(); err != nil {
		return d.bail(err)
	}
	return d.stack.pop().jsonDecoderFn
}

func (d *JSONDecoder) decode(name xml.Name, value string) error {
	elem := xml.StartElement{Name: name}
	if err := d.StartElement(elem); err != nil {
		return err
	} else if err := d.CharData(xml.CharData(value)); err != nil {
		return err
	}
	return d.EndElement(elem.End())
}

func jsDecodeObject(d *JSONDecoder) jsonDecoderFn {
	for d.More() {
		t1, err := d.Token()
		t2, err2 := d.Token()
		if err != nil {
			return d.bail(err)
		} else if err2 != nil {
			return d.bail(err)
		}

		key, ok := t1.(string)
		if !ok {
			return d.bail(errors.Errorf("bad JSON object key type: %T (%#v)", t1, t1))
		}

		name := jsTagToName(key)
		var next jsonDecoderFn
		switch value := t2.(type) {
		case json.Delim:
			switch value {
			case '[':
				d.stack.push(jdState{jsDecodeObject, key})
				next = jsDecodeArray
			case '{':
				d.stack.push(jdState{jsDecodeObject, key})
				next = jsDecodeObject
				err = d.StartElement(xml.StartElement{Name: name})
			default:
				err = errors.Errorf("unexpected delimiter type: %T (%#v)", t2, t2)
			}
		case string:
			err = d.decode(name, value)
		case bool, float64:
			err = d.decode(name, fmt.Sprintf("%v", value))
		case json.Number:
			err = d.decode(name, value.String())
		case nil:
			err = d.decode(name, "")
		}
		if err != nil {
			return d.bail(err)
		} else if next != nil {
			return next
		}
	}
	// consume the ending delimiter
	if _, err := d.Token(); err != nil {
		return d.bail(err)
	}
	prev := d.stack.pop()
	if prev.key != "" {
		if err := d.EndElement(xml.EndElement{Name: jsTagToName(prev.key)}); err != nil {
			return d.bail(err)
		}
	}
	return prev.jsonDecoderFn
}
