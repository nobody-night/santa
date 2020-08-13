// MIT License
//
// Copyright (c) 2020 Nobody Night
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package santa

import (
	"math"
	"strconv"
)

// ElementType represents the native data type of an element. The
// Element structure uses the value of this type to determine the
// native data type of the value. For details, please refer to the
/// comment section of the Element structure.
type ElementType uint8

const (
	// TypeInt represents the element's native data type is Int64.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeInt ElementType = iota + 1

	// TypeUint represents the element's native data type is Uint64.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeUint

	// TypeFloat32 represents the element's native data type is Float32.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeFloat32

	// TypeFloat64 represents the element's native data type is Float64.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeFloat64

	// TypeBoolean represents the element's native data type as Bool.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeBoolean

	// TypeString represents the element's native data type as String.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeString

	// TypeBytes represents the element's native data type as Byte slice.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeBytes

	// TypeFields represents the data type of the element as the Field
	// slice. For details, please refer to the comment section of the
	// Element structure and Field structure.
	TypeFields

	// TypeValue represents the native data type of the element is
	// a value type that has implemented the relevant formatter interface.
	// For details, please refer to the comment section of the Element
	// structure.
	TypeValue
)

// Element is a structure that contains a value of native data type.
//
// Each element can contain a converted value of the native data type,
// and restore it to the value of the native data type when in use.
// This is to ensure the consistency of the data type.
//
// For example, the value part of a field can be an element, and each
// member of an array or list can be an element.
//
// It is worth noting that any value that has implemented the relevant
// formatter interface can use element storage. For example, the
// ElementInts data type of the actual native data type []int64
// implements the JSONFormatter interface, so when the FormatJSON
// function of the element is called, the FormatJSON function of the
// value of the ElementInts data type will be automatically called.
//
// This means that applications can easily extend custom element types,
// as long as these types implement the relevant formatter interface.
type Element struct {
	// Type represents the native data type of an element, and its
	// optional options are constants starting with Type... If not
	// provided or the value is 0, it means an invalid element.
	Type ElementType

	// Number represents a number container, and all values of
	// native data types that represent numbers are stored in this
	// container.
	Number int64

	// String represents a string container, and all values of
	// native data types expressing strings are stored in this
	// container.
	String string

	// Interface represents an interface container. The value
	// of any native data type can be stored in this container.
	// The value is stored in the heap memory, but the storage
	// cost is slightly expensive.
	Interface interface { }
}

// FormatJSON formats the value of the element as a JSON string, then
// appends to the given buffer slice, and finally returns the appended
// buffer slice.
func (e Element) FormatJSON(buffer []byte) []byte {
	switch e.Type {
	case TypeInt:
		return strconv.AppendInt(buffer, e.Number, 10)
	case TypeUint:
		return strconv.AppendUint(buffer, uint64(e.Number), 10)
	case TypeFloat32:
		return strconv.AppendFloat(buffer, math.Float64frombits(
			uint64(e.Number)), 'f', -1, 32)
	case TypeFloat64:
		return strconv.AppendFloat(buffer, math.Float64frombits(
			uint64(e.Number)), 'f', -1, 64)
	case TypeBoolean:
		if e.Number > 0 {
			return append(buffer, "true"...)
		}
		return append(buffer, "false"...)
	case TypeString:
		buffer = append(buffer, '"')
		buffer = append(buffer, e.String...)
		return append(buffer, '"')
	case TypeBytes:
		buffer = append(buffer, '"')
		buffer = append(buffer, e.Interface.([]byte)...)
		return append(buffer, '"')
	default:
		element, ok := e.Interface.(JSONFormatter)
		if !ok {
			return append(buffer, "???"...)
		}
		return element.FormatJSON(buffer)
	}
}

// Field is a structure that contains the name and value of a field.
//
// Fields use elements to store the value of a field's native data type.
// A field can represent a composite key-value structure, and the value
// of a field can be a field or an array.
//
// For details on elements, please refer to the comments section of the
// Element structure.
type Field struct {
	// Element represents the field value of the field.
	Element

	// Name represents the unique name of the field.
	Name string
}

// Int returns the value of a field with a given name and a given
// int64 value. For details, see the comments section of the Field
// structure.
func Int(name string, value int64) Field {
	return Field {
		Element: Element {
			Type: TypeInt,
			Number: value,
		},
		Name: name,
	}
}

// Uint returns the value of a field with a given name and a given
// uint64 value. For details, see the comments section of the Field
// structure.
func Uint(name string, value uint64) Field {
	return Field {
		Element: Element {
			Type: TypeUint,
			Number: int64(value),
		},
		Name: name,
	}
}

// Float32 returns the value of a field with a given name and a given
// float32 value. For details, see the comments section of the Field
// structure.
func Float32(name string, value float32) Field {
	return Field {
		Element: Element {
			Type: TypeFloat32,
			Number: int64(math.Float32bits(value)),
		},
		Name: name,
	}
}

// Float64 returns the value of a field with a given name and a given
// float64 value. For details, please refer to the comments section of
// the Field structure.
func Float64(name string, value float64) Field {
	return Field {
		Element: Element {
			Type: TypeFloat64,
			Number: int64(math.Float64bits(value)),
		},
		Name: name,
	}
}

// Boolean returns the value of a field with a given name and a given
// bool value. For details, see the comments section of the Field
// structure.
func Boolean(name string, value bool) Field {
	var number uint8

	if value {
		number = 1
	}

	return Field {
		Element: Element {
			Type: TypeBoolean,
			Number: int64(number),
		},
		Name: name,
	}
}

// String returns the value of a field with a given name and a given
// string value. For details, see the comments section of the Field
// structure.
func String(name string, value string) Field {
	return Field {
		Element: Element {
			Type: TypeString,
			String: value,
		},
		Name: name,
	}
}

// Bytes returns the value of a field with a given name and a given
// []byte value. For details, see the comments section of the Field
// structure.
func Bytes(name string, value []byte) Field {
	return Field {
		Element: Element {
			Type: TypeBytes,
			Interface: value,
		},
		Name: name,
	}
}

// Value returns the value of a field with a given name and a given
// value. The given value must have implemented the relevant formatter
// interface. Please refer to the comments section of the Field
// structure for details.
func Value(name string, value interface { }) Field {
	switch v := value.(type) {
	case int:
		return Int(name, int64(v))
	case int16:
		return Int(name, int64(v))
	case int32:
		return Int(name, int64(v))
	case int64:
		return Int(name, v)
	case uint:
		return Uint(name, uint64(v))
	case uint8:
		return Uint(name, uint64(v))
	case uint16:
		return Uint(name, uint64(v))
	case uint32:
		return Uint(name, uint64(v))
	case uint64:
		return Uint(name, uint64(v))
	case float32:
		return Float32(name, v)
	case float64:
		return Float64(name, v)
	case bool:
		return Boolean(name, v)
	case string:
		return String(name, v)
	case []byte:
		return Bytes(name, v)
	}

	return Field {
		Element: Element {
			Type: TypeValue,
			Interface: value,
		},
		Name: name,
	}
}

// ElementFields represents an element data type whose native data type
// is []Fields. For details, please refer to the comment section of the
// Element structure.
type ElementFields []Field

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementFields) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '{')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = append(buffer, '"')
		buffer = append(buffer, e[index].Name...)
		buffer = append(buffer, "\": "...)
		buffer = e[index].FormatJSON(buffer)

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, '}')
}

// Fields returns the value of a field with a given name and a given
// []Fields value. For details, see the comments section of the Field
// structure.
func Fields(name string, fields ...Field) Field {
	return Value(name, ElementFields(fields))
}

// ElementInts represents an element data type whose native data type
// is []int64. For details, please refer to the comment section of the
// Element structure.
type ElementInts []int64

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementInts) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = strconv.AppendInt(buffer, e[index], 10)

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Ints returns the value of a field with a given name and a given
// []int64 value. For details, see the comments section of the Field
// structure.
func Ints(name string, values []int64) Field {
	return Value(name, ElementInts(values))
}

// ElementUint64s represents an element data type whose native data type
// is []uint64. For details, please refer to the comment section of the
// Element structure.
type ElementUint64s []uint64

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementUint64s) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = strconv.AppendUint(buffer, e[index], 10)

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Uints returns the value of a field with a given name and a given
// []uint64 value. For details, see the comments section of the Field
// structure.
func Uints(name string, values []uint64) Field {
	return Value(name, ElementUint64s(values))
}

// ElementFloat32s represents an element data type whose native data type
// is []float32. For details, please refer to the comment section of the
// Element structure.
type ElementFloat32s []float32

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementFloat32s) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = strconv.AppendFloat(buffer, float64(e[index]),
			'f', -1, 32)

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Float32s returns the value of a field with a given name and a given
// []float32 value. For details, see the comments section of the Field
// structure.
func Float32s(name string, values []float32) Field {
	return Value(name, ElementFloat32s(values))
}

// ElementFloat64s represents an element data type whose native data type
// is []float64. For details, please refer to the comment section of the
// Element structure.
type ElementFloat64s []float64

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementFloat64s) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = strconv.AppendFloat(buffer, float64(e[index]),
			'f', -1, 64)

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Float64s returns the value of a field with a given name and a given
// []float64 value. For details, see the comments section of the Field
// structure.
func Float64s(name string, values []float64) Field {
	return Value(name, ElementFloat64s(values))
}

// ElementBooleans represents an element data type whose native data type
// is []bool. For details, please refer to the comment section of the
// Element structure.
type ElementBooleans []bool

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementBooleans) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = strconv.AppendBool(buffer, e[index])

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Booleans returns the value of a field with a given name and a given
// []bool value. For details, see the comments section of the Field
// structure.
func Booleans(name string, values []bool) Field {
	return Value(name, ElementBooleans(values))
}

// ElementStrings represents an element data type whose native data type
// is []string. For details, please refer to the comment section of the
// Element structure.
type ElementStrings []string

// FormatJSON formats the element as a JSON string, then appends to the
// given buffer slice, and finally returns the appended buffer slice.
func (e ElementStrings) FormatJSON(buffer []byte) []byte {
	buffer = append(buffer, '[')
	last := len(e) - 1

	for index := 0; index < len(e); index++ {
		buffer = append(buffer, '"')
		buffer = append(buffer, e[index]...)
		buffer = append(buffer, '"')

		if index < last {
			buffer = append(buffer, ", "...)
		}
	}

	return append(buffer, ']')
}

// Strings returns the value of a field with a given name and a given
// []string value. For details, see the comments section of the Field
// structure.
func Strings(name string, values []string) Field {
	return Value(name, ElementStrings(values))
}