package profile

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/lvhuat/kits/pkgs/tags"
)

type ProfileParser interface {
	Parse(reflect.Value) error
}

type ProfileField struct {
	t         string
	text      string
	route     string
	valueType string
}

type kv struct {
	k  string
	v  string
	vt string
}

type folder struct {
	kvs   []kv
	fds   []*folder
	route string
	lo    string
}

type parseStatus struct {
	meta  map[string]*ProfileField
	route []string
	bytes.Buffer
	root folder
}

func (status *parseStatus) addField(l *ProfileField) {
	if status.meta == nil {
		status.meta = make(map[string]*ProfileField, 30)
	}
	_, exist := status.meta[l.route]
	if exist {
		panic(l.route + " duplicate defined")
	}
	status.meta[l.route] = l
}

func (status *parseStatus) zoomIn(n string) {
	status.route = append(status.route, n)
}

func (status *parseStatus) routes(n string) string {
	s := strings.Join(status.route, ".")
	if s == "" {
		return n
	}
	return s + "." + n
}

func (status *parseStatus) zoomOut(n string) {
	status.route = status.route[:len(status.route)-1]
}

type profileParserImpl struct {
	f string
}

func (parser *profileParserImpl) Parse(v interface{}) error {
	_, err := toml.DecodeFile(parser.f, v)
	if err != nil {
		return fmt.Errorf("parse %s failed:%v", parser.f, err)
	}

	if err := parseEnv(v, &parseStatus{}); err != nil {
		return fmt.Errorf("parse env failed:%v", err)
	}

	return parseDefault(v, &parseStatus{})
}

var (
	profileType = reflect.TypeOf(new(Profile)).Elem()
)

func parseDefault(v interface{}, parseStatus *parseStatus) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = errors.New(s)
			}
			panic(r)
		}
	}()

	parseDefault0(reflect.ValueOf(v), parseStatus)

	return nil
}

func parseDefault0(v reflect.Value, parseStatus *parseStatus) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Interface {
			field = field.Elem()
		}
		if p, ok := field.Addr().Interface().(Profile); ok {
			p.Init()
		}
		fieldTag := v.Type().Field(i).Tag
		tagName := fieldTag.Get("toml")
		if tagName == "" {
			tagName = v.Type().Field(i).Name
		}
		tag, _ := tags.Parse(tagName)
		if tag == "-" {
			continue
		}
		if field.Kind() == reflect.Struct {
			parseStatus.zoomIn(tag)
			parseDefault0(field, parseStatus)
			parseStatus.zoomOut(tag)
			continue
		}

		parseStatus.addField(&ProfileField{t: "default", route: parseStatus.routes(tag)})
	}
}

func isQuoteField(vt string) bool {
	return vt == "String" || vt == "time.Date" || vt == "time.Duration"
}

func folderToText(buffer *bytes.Buffer, fd *folder) {
	if fd.lo != "" {
		buffer.WriteString(fmt.Sprintf("[%s]\n", fd.lo))
	}

	for _, kvt := range fd.kvs {
		if isQuoteField(kvt.vt) {
			buffer.WriteString(fmt.Sprintf("%s=\"%s\"\n", kvt.k, kvt.v))
			continue
		}
		buffer.WriteString(fmt.Sprintf("%s=%s\n", kvt.k, kvt.v))
	}

	for _, subFd := range fd.fds {
		folderToText(buffer, subFd)
	}
}

func parseEnv(v interface{}, parseStatus *parseStatus) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = errors.New(s)
			}
			panic(r)
		}
	}()
	parseEnv0(reflect.ValueOf(v), parseStatus, &parseStatus.root)
	folderToText(&parseStatus.Buffer, &parseStatus.root)
	fmt.Printf(parseStatus.Buffer.String())
	_, err = toml.Decode(parseStatus.Buffer.String(), v)
	return err
}

func parseEnv0(v reflect.Value, parseStatus *parseStatus, fd *folder) (err error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Interface {
			field = field.Elem()
		}

		fieldTag := v.Type().Field(i).Tag
		tagName := fieldTag.Get("toml")
		if tagName == "" {
			tagName = v.Type().Field(i).Name
		}
		tag, _ := tags.Parse(tagName)
		if tag == "-" {
			continue
		}
		if field.Kind() == reflect.Struct {
			subFolder := &folder{
				route: parseStatus.routes(tag),
				lo:    tag,
			}
			fd.fds = append(fd.fds, subFolder)

			parseStatus.zoomIn(tag)
			parseEnv0(field, parseStatus, subFolder)
			parseStatus.zoomOut(tag)
			continue
		}
		key := tag
		value, exist := syscall.Getenv(parseStatus.routes(tag))
		if !exist {
			continue
		}
		fd.kvs = append(fd.kvs, kv{
			k:  key,
			v:  value,
			vt: field.Type().String(),
		})
	}

	return nil
}

type Plan struct {
	ConsulKv bool
}

type ParseMeta struct {
}

func Parse(f string, v interface{}) (*Plan, *ParseMeta, error) {
	parser := &profileParserImpl{
		f: f,
	}
	if err := parser.Parse(v); err != nil {
		return nil, nil, err
	}

	return &Plan{}, nil, nil
}
