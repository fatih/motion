// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vim provides an encoder for structured data in Vimscript.
package vim

import (
	"bytes"
	"fmt"
	"reflect"
)

// Marshal returns the Vimscript encoding of v.
func Marshal(x interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshal(&buf, reflect.ValueOf(x)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshal(buf *bytes.Buffer, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Invalid:
		buf.WriteString("null")

	case reflect.Bool:
		fmt.Fprint(buf, v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		fmt.Fprint(buf, v.Interface())

	case reflect.String:
		// TODO(adonovan): check Go's escapes are supported by Vimscript.
		fmt.Fprintf(buf, "%q", v.String())

	case reflect.Ptr:
		return marshal(buf, v.Elem())

	case reflect.Array, reflect.Slice:
		buf.WriteByte('[')
		n := v.Len()
		for i := 0; i < n; i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			if err := marshal(buf, v.Index(i)); err != nil {
				return err
			}
		}
		buf.WriteByte(']')

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("non-string key type in %s", v.Type())
		}
		buf.WriteByte('{')
		for i, k := range v.MapKeys() {
			if i > 0 {
				buf.WriteString(", ")
			}
			if err := marshal(buf, k); err != nil {
				return err
			}
			buf.WriteString(": ")
			if err := marshal(buf, v.MapIndex(k)); err != nil {
				return err
			}
		}
		buf.WriteByte('}')

	case reflect.Struct:
		t := v.Type()
		n := t.NumField()
		buf.WriteByte('{')
		sep := ""
		for i := 0; i < n; i++ {
			sf := t.Field(i)
			if sf.PkgPath != "" { // unexported
				continue
			}

			tag := sf.Tag.Get("vim")
			if tag == "-" {
				continue
			}

			name, options := parseTag(tag)
			if !isValidTag(name) {
				name = ""
			}

			if name == "" {
				name = sf.Name
			}

			fv := v.Field(i)
			if !fv.IsValid() || options == "omitempty" && isEmptyValue(fv) {
				continue
			}

			buf.WriteString(sep)
			sep = ", "

			fmt.Fprintf(buf, "%q: ", name)
			if err := marshal(buf, fv); err != nil {
				return err
			}
		}
		buf.WriteByte('}')

	case reflect.Interface:
		// TODO(adonovan): test with nil
		return marshal(buf, v.Elem())

	case reflect.Complex64, reflect.Complex128,
		reflect.UnsafePointer,
		reflect.Func,
		reflect.Chan:
		return fmt.Errorf("unsupported type: %s", v.Type())
	}

	return nil
}

// from $GOROOT/src/encoding/json/encode.go
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
