package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"reflect"
)

const MAX_TAG_LENGTH = 50

type ConfigItem struct {
	Help   string
	Type   reflect.Type
	Name   string
	Value  string
	Tag    string
	Module string
	Key    string
}
func ReadStructEnv(v reflect.Value, fields map[string]*ConfigItem) map[string]*ConfigItem {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	module := t.Name()
	if fields == nil {
		fields = make(map[string]*ConfigItem, 100)
	}

	for i := 0; i < t.NumField(); i++ {
		structItem := t.Field(i)
		subt := t.Field(i).Type
		subv := v.Field(i)
		switch subt.Kind() {
		case reflect.Struct:
			ReadStructEnv(subv, fields)
		default:
			tag := structItem.Tag.Get("env")
			help := structItem.Tag.Get("help")
			if tag == "" {
				continue
			}
			if len(tag) > MAX_TAG_LENGTH {
				panic("Env tag too long")
			}
			_, exist := fields[tag]
			if !exist {
				fields[tag] = &ConfigItem{
					Key:    tag,
					Help:   help,
					Type:   subt,
					Name:   subt.Name(),
					Value:  fmt.Sprintf("%v", subv),
					Module: module,
					Tag:    "default",
				}
			} else {
				panic("duplicate fields " + tag)
			}
		}
	}

	return fields
}

func FetchEnv(v reflect.Value) ([]*ConfigItem, map[string][]*ConfigItem) {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	module := t.Name()
	embedded := map[string][]*ConfigItem{}
	l := []*ConfigItem{}
	for i := 0; i < t.NumField(); i++ {
		structItem := t.Field(i)
		subt := t.Field(i).Type
		subv := v.Field(i)
		switch subt.Kind() {
		case reflect.Struct:
			tag := structItem.Tag.Get("embedded")
			if tag == "" {
				tag = subt.Name()

			}
			e, _ := FetchEnv(subv)
			if len(e) > 0 {
				embedded[tag] = e
			}
		default:
			tag := structItem.Tag.Get("env")
			help := structItem.Tag.Get("help")
			if tag == "" {
				continue
			}
			if len(tag) > MAX_TAG_LENGTH {
				panic("Env tag too long")
			}
			l = append(l, &ConfigItem{
				Key:    tag,
				Help:   help,
				Type:   subt,
				Name:   subt.Name(),
				Value:  fmt.Sprintf("%v", subv),
				Module: module,
				Tag:    "default",
			})
		}
	}
	return l, embedded
}

func ReadTomlEnvConfig(config interface{}, fields map[string]*ConfigItem) (int, error) {
	t := reflect.TypeOf(config)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	count := 0
	tomlText := ""
	for key, tf := range fields {
		count++
		v := os.Getenv(key)
		if v == "" {
			continue
		}
		value := ""
		tomlText += key + " = "
		switch tf.Type.Kind() {
		case reflect.String:
			value += "'" + v + "'"
		default:
			value += v
		}
		tomlText += value + "\n"
		tf.Tag = "env"
		fields[key] = tf
	}

	_, err := toml.Decode(tomlText, config)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func ReadTomlFileConfig(file string, config interface{}, fields map[string]*ConfigItem) (int, error) {
	var (
		md  toml.MetaData
		err error
	)

	if md, err = toml.DecodeFile(file, config); err != nil {
		return 0, err
	}
	keys := md.Keys()
	ignores := []string{}
	for _, f := range keys {
		for _, key := range f {
			tf, exist := fields[key]
			if !exist {
				ignores = append(ignores, key)
				continue
			}
			tf.Tag = file
		}
	}

	if len(ignores) > 0 {
		fmt.Printf("Key in %s not used:%v\n", file, ignores)
	}

	return len(md.Keys()), nil
}

func SetupToml(config interface{}, t ConfigType, opts ...ConfigOption) error {
	opts = append(opts, FileDecodeOpt(ReadTomlFileConfig))
	opts = append(opts, EnvDecodeOpt(ReadTomlEnvConfig))
	defaultManager = New(config, t, opts...)
	if err := defaultManager.setup(); err != nil {
		return err
	}
	return nil
}
