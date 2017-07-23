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

// ProfileParser 配置解析器
// 配置解析器可以按配置文件,环境变量，写死的配置的顺序去加载配置
type ProfileParser interface {
	Parse(reflect.Value) error
}

// ProfileField 每个配置点的信息
type ProfileField struct {
	t         string
	text      string
	route     string
	valueType string
}

// kv 保存在环境变量中配置的键值和数据类型信息
type kv struct {
	k  string
	v  string
	vt string
}

// folder 保存使用环境变量配置的中间元数据
type folder struct {
	kvs   []kv
	fds   []*folder
	route string
	lo    string
}

// parseStatus 是解析过程中的状态记录器
type parseStatus struct {
	meta  map[string]*ProfileField
	route []string
	bytes.Buffer
	root folder
}

// addField 添加一个项目
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

// zoomIn 进入一个配置项目
func (status *parseStatus) zoomIn(n string) {
	status.route = append(status.route, n)
}

// zoomOut 退出当前配置项目，回到父项目
func (status *parseStatus) zoomOut(n string) {
	status.route = status.route[:len(status.route)-1]
}

// routes 当前配置点的路径
func (status *parseStatus) routes(n string) string {
	s := strings.Join(status.route, ".")
	if s == "" {
		return n
	}
	return s + "." + n
}

// profileParserImpl 解析器的具体实现
type profileParserImpl struct {
	f string
}

// Parse 实现服务配置的解析
func (parser *profileParserImpl) Parse(v interface{}) error {
	if err := parseInit(v, &parseStatus{}); err != nil {
		return fmt.Errorf("parse init failed:%v", err)
	}

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

// parseDefault 解析Stub Value入口
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

// parseDefault0 解析默认值实现
func parseDefault0(v reflect.Value, parseStatus *parseStatus) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Interface {
			field = field.Elem()
		}
		// 判断是否实现了接口
		if p, ok := field.Addr().Interface().(Profile); ok {
			p.AfterParse()
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

// parseDefault 解析Stub Value入口
func parseInit(v interface{}, parseStatus *parseStatus) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = errors.New(s)
			}
			panic(r)
		}
	}()

	parseInit0(reflect.ValueOf(v), parseStatus)

	return nil
}

// parseDefault0 解析默认值实现
func parseInit0(v reflect.Value, parseStatus *parseStatus) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Interface {
			field = field.Elem()
		}
		// 判断是否实现了接口
		if p, ok := field.Addr().Interface().(Profile); ok {
			p.BeforeParse()
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
			parseInit0(field, parseStatus)
			parseStatus.zoomOut(tag)
			continue
		}

		parseStatus.addField(&ProfileField{t: "default", route: parseStatus.routes(tag)})
	}
}

func isQuoteField(vt string) bool {
	return vt == "String" || vt == "time.Time" || vt == "time.Duration"
}

// folderToText 使用环境变量的配置生成一个toml文件
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

// parseEnv 解析环境变量入口
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
	_, err = toml.Decode(parseStatus.Buffer.String(), v)
	return err
}

// parseEnv0 解析环境变量实现
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

// 从配置中分析的执行计划数据
// 比如部分配置是使用consul加载，则需要在初始化consul之后才能完善配置
// TODO:然而并没有实现
type Plan struct {
	ConsulKv bool
}

// ParseMeta 配置加载的中间数据，也许打印什么的时候有用
// TODO:然而并没有实现
type ParseMeta struct {
}

// Parse 是解析配置的外部入口
func Parse(f string, v interface{}) (*Plan, *ParseMeta, error) {
	parser := &profileParserImpl{
		f: f,
	}
	if err := parser.Parse(v); err != nil {
		return nil, nil, err
	}

	return &Plan{}, nil, nil
}
