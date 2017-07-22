package conf

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type ConfigType int32

const (
	EnvPrefer  ConfigType = 0
	FilePrefer ConfigType = 1
	EnvOnly    ConfigType = 2
	FileOnly   ConfigType = 3
)

type FileDecoder func(string, interface{}, map[string]*ConfigItem) (int, error)
type EnvDecoder func(interface{}, map[string]*ConfigItem) (int, error)

type Manager struct {
	config     interface{}
	t          ConfigType
	file       string
	afterDone  func(interface{})
	fileDecode FileDecoder
	envDecode  EnvDecoder
	fields     map[string]*ConfigItem
	table      string
}

type ConfigOption func(*Manager)

func FileOpt(opt string) ConfigOption {
	return func(m *Manager) {
		m.file = opt
	}
}

func FileDecodeOpt(opt FileDecoder) ConfigOption {
	return func(m *Manager) {
		m.fileDecode = opt
	}
}

func EnvDecodeOpt(opt EnvDecoder) ConfigOption {
	return func(m *Manager) {
		m.envDecode = opt
	}
}

func New(config interface{}, t ConfigType, opts ...ConfigOption) *Manager {
	m := &Manager{
		config: config,
		t:      t,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.afterDone == nil {
		m.afterDone = func(interface{}) {}
	}

	return m
}

var ErrBadConfigType = errors.New("bad config type")
var defaultManager *Manager

func (mgr *Manager) setup() error {
	mgr.fields = make(map[string]*ConfigItem, 100)
	fields := ReadStructEnv(reflect.ValueOf(mgr.config), mgr.fields)
	switch mgr.t {
	case EnvOnly:
		if _, err := mgr.fileDecode(mgr.file, mgr.config, fields); err != nil {
			return err
		}
	case FileOnly:
		if _, err := mgr.envDecode(mgr.config, fields); err != nil {
			return err
		}
	case EnvPrefer:
		if _, err := mgr.fileDecode(mgr.file, mgr.config, fields); err != nil {
			return err
		}
		if _, err := mgr.envDecode(mgr.config, fields); err != nil {
			return err
		}
	case FilePrefer:
		if _, err := mgr.envDecode(mgr.config, fields); err != nil {
			return err
		}
		if _, err := mgr.fileDecode(mgr.file, mgr.config, fields); err != nil {
			return err
		}
	default:
		return ErrBadConfigType
	}

	mgr.afterDone(mgr.config)

	return nil
}

func Cut(src string, n int) string {
	if len(src) <= n+3 {
		return src
	}
	if len(src) < 3 {
		return "..."
	}
	return string(src[:n-3]) + "..."
}

func maxLen(x int, y int) int {
	if x > y {
		return x
	}
	return y
}
func minLen(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func (mgr *Manager) Table() string {
	if mgr.table != "" {
		return mgr.table
	}

	fs := mgr.fields
	result := ReadStructEnv(reflect.ValueOf(mgr.config), nil)
	for key, value := range result {
		tf := fs[key]
		tf.Value = value.Value
	}

	n1, n2, n3, n4, n5 := 20, 20, 20, 20, 20
	for key, value := range fs {
		n1 = maxLen(n1, len(key)+2)
		n2 = maxLen(n2, len(value.Type.String())+2)
		n3 = maxLen(n3, len(value.Tag)+2)
		n4 = maxLen(n4, len(value.Value)+2)
		n5 = maxLen(n5, len(value.Help)+2)
	}
	total := n1 + n2 + n3 + n4 + n5
	format := fmt.Sprintf("\t%%-%ds%%-%ds%%-%ds%%-%ds%%-%ds", n1, n2, n3, n4, n5)
	p := strings.Repeat("=", total) + "\n"

	p += fmt.Sprintf(format+"\n", "Configuration Item", "Data Type", "Tag", "Value", "Help")
	p += strings.Repeat("-", total) + "\n"
	for key, value := range fs {
		p += fmt.Sprintf(format+"\n",
			Cut(key, 100),
			Cut(value.Type.String(), 100),
			Cut(value.Tag, 100),
			Cut(value.Value, 100),
			Cut(value.Help, 100),
		)
	}
	p += strings.Repeat("-", total) + "\n"

	mgr.table = p

	return p
}

// String use for readable logger
func (mgr *Manager) String() string {
	l, inside := FetchEnv(reflect.ValueOf(mgr.config))
	inside["APPLICATION"] = l

	n1, n2, n3, n4, n5 := 0, 0, 0, 0, 0
	ps := []string{}
	for key, items := range inside {
		for _, item := range items {
			n1 = minLen(50, maxLen(n1, len(item.Key)))
			n2 = minLen(20, maxLen(n2, len(item.Type.String())))
			n3 = minLen(20, maxLen(n3, len(item.Tag)))
			n4 = minLen(50, maxLen(n4, len(item.Value)))
			n5 = minLen(100, maxLen(n5, len(item.Help)))
		}

		ps = append(ps, key)
	}

	sort.Strings(ps)

	format := fmt.Sprintf("    %%-%ds%%-%ds%%-%ds%%-%ds%%-%ds", n1+2, n2+2, n3+3, n4+2, n5)
	total := n1 + n2 + n3 + n4 + n5 + 4*2 + 5

	p := []byte{}
	p = append(p, []byte(strings.Repeat("=", total)+"\n")...)
	for _, key := range ps {
		items := inside[key]
		p = append(p, []byte("["+strings.TrimSuffix(strings.ToUpper(key), "CONFIG")+"]\n")...)
		for _, value := range items {
			configItem, exist := mgr.fields[value.Key]
			if exist {
				value.Tag = configItem.Tag
			}
			p = append(p, []byte(fmt.Sprintf(format+"\n",
				Cut(value.Key, n1),
				Cut(value.Type.String(), n2),
				Cut(value.Tag, n3),
				Cut(value.Value, n4),
				Cut(value.Help, n5),
			))...)
		}
	}

	p = append(p, []byte(strings.Repeat("-", total)+"\n")...)

	return string(p)
}

// DumpValues use for structured logger
func (mgr *Manager) DumpValues() map[string]interface{} {
	l, inside := FetchEnv(reflect.ValueOf(mgr.config))
	inside["APPLICATION"] = l
	ps := []string{}
	for key, _ := range inside {
		ps = append(ps, key)
	}

	sort.Strings(ps)

	all := map[string]interface{}{}
	for _, key := range ps {
		items := inside[key]
		for _, value := range items {
			configItem, exist := mgr.fields[value.Key]
			if exist {
				value.Tag = configItem.Tag
			}
			value.Key = strings.ToUpper(key) + "." + value.Key
			if value.Value == "" {
				continue
			}
			all[value.Key] = value.Value
		}
	}

	return all
}

var ErrConfigTypeNotStruct = errors.New("config type not struct")

func Setup(config interface{}, t ConfigType, opts ...ConfigOption) error {
	return SetupToml(config, t, opts...)
}

func Dump() string {
	return defaultManager.String()
}

func DumpValues() map[string]interface{} {
	return defaultManager.DumpValues()
}
