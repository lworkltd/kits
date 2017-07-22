package svc

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"

	"lwork.com/kits/utils/tags"

	"fmt"
)

type encodeState struct {
	*bytes.Buffer
	buffers      map[string]*bytes.Buffer
	mutilBuffers map[string][]*bytes.Buffer
	scratch      [64]byte
	depth        int16
	inArray      bool
	inMap        bool
	field        *topField
	encoder      string
	key          string
}

func (es *encodeState) check() {
	t := es.field.typ
	if !es.inArray && !es.inMap {
		switch t.Kind() {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			es.error(fmt.Errorf("cannot use type %s as field without specified tags", t.Kind().String()))
		}
	}
}

func jsonEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	b, err := json.Marshal(v.Interface())
	if err != nil {
		e.error(err)
	}
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func Marshal(v interface{}, tag string) (map[string][][]byte, error) {
	es := &encodeState{
		//buffers:      map[string]*bytes.Buffer{},
		mutilBuffers: map[string][]*bytes.Buffer{},
	}
	if err := es.marshal(v, tag); err != nil {
		return nil, err
	}
	bs := map[string][][]byte{}
	// for k, v := range es.buffers {
	// 	bs[k] = [][]byte{v.Bytes()}
	// }

	for k, vs := range es.mutilBuffers {
		lots := [][]byte{}
		fmt.Println("############", k, vs)
		for _, buf := range vs {
			lots = append(lots, buf.Bytes())
		}
		bs[k] = lots
	}

	return bs, nil
}

func (es *encodeState) error(err error) {
	panic(err)
}

func (es *encodeState) marshal(v interface{}, tag string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()

	val := reflect.ValueOf(v)
	topType(tag, val.Type())(es, val, &encOpts{})

	return nil
}

type topEncoder func(e *encodeState, v reflect.Value, opts *encOpts)

func topType(tag string, t reflect.Type) topEncoder {
	fmt.Println("top Type", t.Kind())
	switch t.Kind() {
	case reflect.Struct:
		return newTopStructEncoder(t, tag)
	case reflect.Ptr:
		return newTopPtrEncoder(t, tag)
	case reflect.Map:
		return newTopMapEncoder(t)
	default:
		panic("not support top type")
	}
}

type topStructEncoder struct {
	fields    []topField
	fieldEncs []encoderFunc
}

func (se *topStructEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	for i, f := range se.fields {
		fv := fieldByIndex(v, f.index)
		if !fv.IsValid() || f.omitEmpty && isEmptyValue(fv) {
			continue
		}
		e.field = &f
		e.Buffer = &bytes.Buffer{}
		e.key = f.name
		se.fieldEncs[i](e, fv, opts)
		if e.Buffer != nil {
			bs := e.mutilBuffers[e.key]
			bs = append(bs, e.Buffer)
			e.mutilBuffers[e.key] = bs
		}
	}
}

func newTopStructEncoder(t reflect.Type, tag string) topEncoder {
	fields := cachedTopTypeFields(t, tag)
	se := &topStructEncoder{
		fields:    fields,
		fieldEncs: make([]encoderFunc, len(fields)),
	}

	for i, f := range fields {
		fmt.Println("top struct:", t.Name(), "field", f.name)
		se.fieldEncs[i] = typeEncoder(typeByIndex(t, f.index), tag)
	}

	return se.encode
}

var topFieldCache struct {
	value atomic.Value // map[reflect.Type][]field
	mu    sync.Mutex   // used only by writers
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTopTypeFields(t reflect.Type, tag string) []topField {
	m, _ := topFieldCache.value.Load().(map[reflect.Type][]topField)
	f := m[t]
	if f != nil {
		return f
	}

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = topTypeFields(t, tag)
	if f == nil {
		f = []topField{}
	}

	topFieldCache.mu.Lock()
	m, _ = topFieldCache.value.Load().(map[reflect.Type][]topField)
	newM := make(map[reflect.Type][]topField, len(m)+1)
	for k, v := range m {
		newM[k] = v
	}
	newM[t] = f
	topFieldCache.value.Store(newM)
	topFieldCache.mu.Unlock()
	return f
}

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func topTypeFields(t reflect.Type, tagName string) []topField {
	// Anonymous fields to explore at the current level and the next.
	current := []topField{}
	next := []topField{{typ: t}}

	// Count of queued names for current level and the next.
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []topField

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.PkgPath != "" && !sf.Anonymous { // unexported
					continue
				}
				tag := sf.Tag.Get(tagName)
				if tag == "-" {
					continue
				}
				name, opts := tags.Parse(tag)
				if !isValidTag(name) {
					name = ""
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Only strings, floats, integers, and booleans can be quoted.
				quoted := false
				if opts.Contains("string") {
					switch ft.Kind() {
					case reflect.Bool,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
						reflect.Float32, reflect.Float64,
						reflect.String:
						quoted = true
					}
				}

				jsonize := name == "json"
				if !jsonize && opts.Contains("json") {
					jsonize = true
				}

				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					fields = append(fields, fillTopField(topField{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						quoted:    quoted,
						jsonize:   jsonize,
					}))
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, fillTopField(topField{name: ft.Name(), index: index, typ: ft}))
				}
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byTopIndex(x).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with JSON tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byTopIndex(fields))

	return fields
}

type encoderFunc func(et *encodeState, v reflect.Value, opts *encOpts)

var encoderCache struct {
	sync.RWMutex
	m map[reflect.Type]encoderFunc
}

// func (es *encodeState) reflectValue(v reflect.Value, opts *encOpts) error {
// 	typeEncoder(v.Type(), opts.tag)(es, v, opts)
// 	return nil
// }

func typeEncoder(t reflect.Type, tag string) encoderFunc {
	encoderCache.RLock()
	f := encoderCache.m[t]
	encoderCache.RUnlock()
	if f != nil {
		return f
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	encoderCache.Lock()
	if encoderCache.m == nil {
		encoderCache.m = make(map[reflect.Type]encoderFunc)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	encoderCache.m[t] = func(e *encodeState, v reflect.Value, opts *encOpts) {
		wg.Wait()
		f(e, v, opts)
	}
	encoderCache.Unlock()

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = newTypeEncoder(t, tag, true)
	wg.Done()
	encoderCache.Lock()
	encoderCache.m[t] = f
	encoderCache.Unlock()
	return f
}
func boolEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	if opts.quoted {
		e.WriteByte('"')
	}
	if v.Bool() {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
	if opts.quoted {
		e.WriteByte('"')
	}
}

func intEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	b := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

func uintEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	b := strconv.AppendUint(e.scratch[:0], v.Uint(), 10)
	if opts.quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	f := v.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		e.error(&json.UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, int(bits))})
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := e.scratch[:0]
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, fmt, -1, int(bits))
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	if opts.quoted {
		e.WriteByte('"')
	}

	e.Write(b)
	if opts.quoted {
		e.WriteByte('"')
	}
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func stringEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	if opts.quoted {
		sb, err := json.Marshal(v.String())
		if err != nil {
			e.error(err)
		}
		e.string(string(sb), opts.escapeHTML)
	} else {
		e.string(v.String(), opts.escapeHTML)
	}
}

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      false,
	'=':      true,
	'>':      false,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
var hex = "0123456789abcdef"

// NOTE: keep in sync with stringBytes below.
func (e *encodeState) string(s string, escapeHTML bool) int {
	len0 := e.Len()
	e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				e.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				e.WriteByte('\\')
				e.WriteByte(b)
			case '\n':
				e.WriteByte('\\')
				e.WriteByte('n')
			case '\r':
				e.WriteByte('\\')
				e.WriteByte('r')
			case '\t':
				e.WriteByte('\\')
				e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				e.WriteString(`\u00`)
				e.WriteByte(hex[b>>4])
				e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				e.WriteString(s[start:i])
			}
			e.WriteString(`\u202`)
			e.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.WriteString(s[start:])
	}
	e.WriteByte('"')
	return e.Len() - len0
}

func interfaceEncoder(e *encodeState, v reflect.Value, opts *encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}

	typeEncoder(v.Elem().Type(), opts.tag)(e, v, opts)
}

type topField struct {
	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

	tag        bool
	index      []int
	encodeType string
	omitEmpty  bool
	quoted     bool
	typ        reflect.Type
	jsonize    bool
	comma      bool
}

// A field represents a single field found in a struct.
// type field struct {
// 	name      string
// 	nameBytes []byte                 // []byte(name)
// 	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

// 	tag        bool
// 	index      []int
// 	encodeType string
// 	omitEmpty  bool
// 	quoted     bool
// 	typ        reflect.Type
// 	jsonize    bool
// }

func fillTopField(f topField) topField {
	f.nameBytes = []byte(f.name)
	f.equalFold = foldFunc(f.nameBytes)
	return f
}

// func fillField(f field) field {
// 	f.nameBytes = []byte(f.name)
// 	f.equalFold = foldFunc(f.nameBytes)
// 	return f
// }

// type structEncoder struct {
// 	fields    []field
// 	fieldEncs []encoderFunc
// }

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

// func (se *structEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
// 	e.check()
// }

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

// func newStructEncoder(t reflect.Type, tag string) encoderFunc {
// 	fields := cachedTypeFields(t, tag)
// 	se := &structEncoder{
// 		fields:    fields,
// 		fieldEncs: make([]encoderFunc, len(fields)),
// 	}

// 	for i, f := range fields {
// 		se.fieldEncs[i] = typeEncoder(typeByIndex(t, f.index), tag)
// 	}

// 	return se.encode
// }

// var fieldCache struct {
// 	value atomic.Value // map[reflect.Type][]field
// 	mu    sync.Mutex   // used only by writers
// }

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// JSON tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []topField) (topField, bool) {
	// The fields are sorted in increasing index-length order. The winner
	// must therefore be one with the shortest index length. Drop all
	// longer entries, which is easy: just truncate the slice.
	length := len(fields[0].index)
	tagged := -1 // Index of first tagged field.
	for i, f := range fields {
		if len(f.index) > length {
			fields = fields[:i]
			break
		}
		if f.tag {
			if tagged >= 0 {
				// Multiple tagged fields at the same level: conflict.
				// Return no field.
				return topField{}, false
			}
			tagged = i
		}
	}
	if tagged >= 0 {
		return fields[tagged], true
	}
	// All remaining fields have the same length. If there's more than one,
	// we have a conflict (two fields named "X" at the same level) and we
	// return no field.
	if len(fields) > 1 {
		return topField{}, false
	}
	return fields[0], true
}

// // dominantField looks through the fields, all of which are known to
// // have the same name, to find the single field that dominates the
// // others using Go's embedding rules, modified by the presence of
// // JSON tags. If there are multiple top-level fields, the boolean
// // will be false: This condition is an error in Go and we skip all
// // the fields.
// func dominantField(fields []field) (field, bool) {
// 	// The fields are sorted in increasing index-length order. The winner
// 	// must therefore be one with the shortest index length. Drop all
// 	// longer entries, which is easy: just truncate the slice.
// 	length := len(fields[0].index)
// 	tagged := -1 // Index of first tagged field.
// 	for i, f := range fields {
// 		if len(f.index) > length {
// 			fields = fields[:i]
// 			break
// 		}
// 		if f.tag {
// 			if tagged >= 0 {
// 				// Multiple tagged fields at the same level: conflict.
// 				// Return no field.
// 				return field{}, false
// 			}
// 			tagged = i
// 		}
// 	}
// 	if tagged >= 0 {
// 		return fields[tagged], true
// 	}
// 	// All remaining fields have the same length. If there's more than one,
// 	// we have a conflict (two fields named "X" at the same level) and we
// 	// return no field.
// 	if len(fields) > 1 {
// 		return field{}, false
// 	}
// 	return fields[0], true
// }

// byIndex sorts field by index sequence.
type byTopIndex []topField

func (x byTopIndex) Len() int { return len(x) }

func (x byTopIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byTopIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// type byIndex []topField

// func (x byIndex) Len() int { return len(x) }

// func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// func (x byIndex) Less(i, j int) bool {
// 	for k, xik := range x[i].index {
// 		if k >= len(x[j].index) {
// 			return false
// 		}
// 		if xik != x[j].index[k] {
// 			return xik < x[j].index[k]
// 		}
// 	}
// 	return len(x[i].index) < len(x[j].index)
// }

// // typeFields returns a list of fields that JSON should recognize for the given type.
// // The algorithm is breadth-first search over the set of structs to include - the top struct
// // and then any reachable anonymous structs.
// func typeFields(t reflect.Type, tagName string) []field {
// 	// Anonymous fields to explore at the current level and the next.
// 	current := []field{}
// 	next := []field{{typ: t}}

// 	// Count of queued names for current level and the next.
// 	count := map[reflect.Type]int{}
// 	nextCount := map[reflect.Type]int{}

// 	// Types already visited at an earlier level.
// 	visited := map[reflect.Type]bool{}

// 	// Fields found.
// 	var fields []field

// 	for len(next) > 0 {
// 		current, next = next, current[:0]
// 		count, nextCount = nextCount, map[reflect.Type]int{}

// 		for _, f := range current {
// 			if visited[f.typ] {
// 				continue
// 			}
// 			visited[f.typ] = true

// 			// Scan f.typ for fields to include.
// 			for i := 0; i < f.typ.NumField(); i++ {
// 				sf := f.typ.Field(i)
// 				if sf.PkgPath != "" && !sf.Anonymous { // unexported
// 					continue
// 				}
// 				tag := sf.Tag.Get(tagName)
// 				if tag == "-" {
// 					continue
// 				}
// 				name, opts := tags.Parse(tag)
// 				if !isValidTag(name) {
// 					name = ""
// 				}
// 				index := make([]int, len(f.index)+1)
// 				copy(index, f.index)
// 				index[len(f.index)] = i

// 				ft := sf.Type
// 				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
// 					// Follow pointer.
// 					ft = ft.Elem()
// 				}

// 				// Only strings, floats, integers, and booleans can be quoted.
// 				quoted := false
// 				if opts.Contains("string") {
// 					switch ft.Kind() {
// 					case reflect.Bool,
// 						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
// 						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
// 						reflect.Float32, reflect.Float64,
// 						reflect.String:
// 						quoted = true
// 					}
// 				}

// 				jsonize := name == "json"
// 				if !jsonize && opts.Contains("json") {
// 					jsonize = true
// 				}

// 				// Record found field and index sequence.
// 				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
// 					tagged := name != ""
// 					if name == "" {
// 						name = sf.Name
// 					}
// 					fields = append(fields, fillField(field{
// 						name:      name,
// 						tag:       tagged,
// 						index:     index,
// 						typ:       ft,
// 						omitEmpty: opts.Contains("omitempty"),
// 						quoted:    quoted,
// 						jsonize:   jsonize,
// 					}))
// 					if count[f.typ] > 1 {
// 						// If there were multiple instances, add a second,
// 						// so that the annihilation code will see a duplicate.
// 						// It only cares about the distinction between 1 or 2,
// 						// so don't bother generating any more copies.
// 						fields = append(fields, fields[len(fields)-1])
// 					}
// 					continue
// 				}

// 				// Record new anonymous struct to explore in next round.
// 				nextCount[ft]++
// 				if nextCount[ft] == 1 {
// 					next = append(next, fillField(field{name: ft.Name(), index: index, typ: ft}))
// 				}
// 			}
// 		}
// 	}

// 	sort.Slice(fields, func(i, j int) bool {
// 		x := fields
// 		// sort field by name, breaking ties with depth, then
// 		// breaking ties with "name came from json tag", then
// 		// breaking ties with index sequence.
// 		if x[i].name != x[j].name {
// 			return x[i].name < x[j].name
// 		}
// 		if len(x[i].index) != len(x[j].index) {
// 			return len(x[i].index) < len(x[j].index)
// 		}
// 		if x[i].tag != x[j].tag {
// 			return x[i].tag
// 		}
// 		return byIndex(x).Less(i, j)
// 	})

// 	// Delete all fields that are hidden by the Go rules for embedded fields,
// 	// except that fields with JSON tags are promoted.

// 	// The fields are sorted in primary order of name, secondary order
// 	// of field index length. Loop over names; for each name, delete
// 	// hidden fields by choosing the one dominant field that survives.
// 	out := fields[:0]
// 	for advance, i := 0, 0; i < len(fields); i += advance {
// 		// One iteration per name.
// 		// Find the sequence of fields with the name of this first field.
// 		fi := fields[i]
// 		name := fi.name
// 		for advance = 1; i+advance < len(fields); advance++ {
// 			fj := fields[i+advance]
// 			if fj.name != name {
// 				break
// 			}
// 		}
// 		if advance == 1 { // Only one field with this name
// 			out = append(out, fi)
// 			continue
// 		}
// 		dominant, ok := dominantField(fields[i : i+advance])
// 		if ok {
// 			out = append(out, dominant)
// 		}
// 	}

// 	fields = out
// 	sort.Sort(byIndex(fields))

// 	return fields
// }

// // cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
// func cachedTypeFields(t reflect.Type, tag string) []field {
// 	m, _ := fieldCache.value.Load().(map[reflect.Type][]field)
// 	f := m[t]
// 	if f != nil {
// 		return f
// 	}

// 	// Compute fields without lock.
// 	// Might duplicate effort but won't hold other computations back.
// 	f = typeFields(t, tag)
// 	if f == nil {
// 		f = []field{}
// 	}

// 	fieldCache.mu.Lock()
// 	m, _ = fieldCache.value.Load().(map[reflect.Type][]field)
// 	newM := make(map[reflect.Type][]field, len(m)+1)
// 	for k, v := range m {
// 		newM[k] = v
// 	}
// 	newM[t] = f
// 	fieldCache.value.Store(newM)
// 	fieldCache.mu.Unlock()
// 	return f
// }

func unsupportedTypeEncoder(e *encodeState, v reflect.Value, _ *encOpts) {
	e.error(&json.UnsupportedTypeError{v.Type()})
}

type reflectWithString struct {
	v reflect.Value
	s string
}

func (w *reflectWithString) resolve() error {
	if w.v.Kind() == reflect.String {
		w.s = w.v.String()
		return nil
	}
	if tm, ok := w.v.Interface().(encoding.TextMarshaler); ok {
		buf, err := tm.MarshalText()
		w.s = string(buf)
		return err
	}
	switch w.v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.s = strconv.FormatInt(w.v.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.s = strconv.FormatUint(w.v.Uint(), 10)
		return nil
	}
	panic("unexpected map key type")
}

type topMapEncoder struct {
	elemEnc encoderFunc
}

func (me *topMapEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.WriteByte('{')

	// Extract and sort the keys.
	keys := v.MapKeys()
	sv := make([]reflectWithString, len(keys))
	for i, v := range keys {
		sv[i].v = v
		if err := sv[i].resolve(); err != nil {
			e.error(&json.MarshalerError{v.Type(), err})
		}
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].s < sv[j].s })

	e.inMap = true
	for _, kv := range sv {
		me.elemEnc(e, v.MapIndex(kv.v), opts)
	}
	e.inMap = false
}

// var (
// 	marshalerType     = reflect.TypeOf(new(json.Marshaler)).Elem()
// 	textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
// )

func newTopMapEncoder(t reflect.Type) topEncoder {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		panic("not support map key type")
	}
	me := &topMapEncoder{typeEncoder(t.Elem(), "")}
	return me.encode
}

func encodeByteSlice(e *encodeState, v reflect.Value, _ *encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	s := v.Bytes()
	e.WriteByte('"')
	if len(s) < 1024 {
		// for small buffers, using Encode directly is much faster.
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
		base64.StdEncoding.Encode(dst, s)
		e.Write(dst)
	} else {
		// for large buffers, avoid unnecessary extra temporary
		// buffer space.
		enc := base64.NewEncoder(base64.StdEncoding, e)
		enc.Write(s)
		enc.Close()
	}
	e.WriteByte('"')
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se *sliceEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	se.arrayEnc(e, v, opts)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	enc := &sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *arrayEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	if e.inArray || e.inMap {
		e.error(errors.New("not support array as array element"))
	}
	n := v.Len()

	e.inArray = true
	if opts.onlyComma {
		for i := 0; i < n && opts.onlyComma; i++ {
			if i == 0 {
				e.Buffer = &bytes.Buffer{}
				e.buffers[e.key] = e.Buffer
			}
			if i > 0 {
				e.WriteByte(',')
			}
			ae.elemEnc(e, v.Index(i), opts)
		}

		e.inArray = false
		return
	}

	buffers := make([]*bytes.Buffer, n)
	for i := 0; i < n && !opts.onlyComma; i++ {
		e.Buffer = &bytes.Buffer{}
		ae.elemEnc(e, v.Index(i), opts)
		buffers[i] = e.Buffer
	}
	e.mutilBuffers[e.key] = buffers
	e.Buffer = nil

	e.inArray = false
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := &arrayEncoder{typeEncoder(t.Elem(), "")}
	return enc.encode
}

type topPtrEncoder struct {
	elemEnc topEncoder
}

func (pe *topPtrEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	if v.IsNil() {
		e.WriteString("")
		return
	}

	pe.elemEnc(e, v.Elem(), opts)
}

func newTopPtrEncoder(t reflect.Type, tag string) topEncoder {
	enc := &topPtrEncoder{topType(tag, t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe *ptrEncoder) encode(e *encodeState, v reflect.Value, opts *encOpts) {
	if v.IsNil() {
		e.WriteString("")
		return
	}
	pe.elemEnc(e, v.Elem(), opts)
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := &ptrEncoder{typeEncoder(t.Elem(), "")}
	return enc.encode
}

type encOpts struct {
	// quoted causes primitive fields to be encoded inside JSON strings.
	quoted bool
	// escapeHTML causes '<', '>', and '&' to be escaped in JSON strings.
	escapeHTML bool
	// tag
	tag string
	// max depth
	maxDepth int16
	// array and slice splitor
	onlyComma bool
}

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type, tag string, allowAddr bool) encoderFunc {
	fmt.Println(t.Kind().String())

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return jsonEncoder
	case reflect.Map:
		return jsonEncoder
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}
