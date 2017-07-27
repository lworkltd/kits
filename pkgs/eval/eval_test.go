package eval

import (
	"fmt"
	"github.com/lworkltd/kits/pkgs/jsonize"
	"reflect"
	"testing"
)

func Test_parseDesc(t *testing.T) {
	type args struct {
		desc string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				desc: "ip_of_interface,x,y, z,",
			},
			want:    "ip_of_interface",
			want1:   []string{"x", "y", "z", ""},
			wantErr: false,
		},
		{
			name: "noargs",
			args: args{
				desc: "ip_of_firt_interface",
			},
			want:    "ip_of_firt_interface",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "space noargs",
			args: args{
				desc: "  ip_of_firt_interface",
			},
			want:    "ip_of_firt_interface",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				desc: "",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseDesc(tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDesc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDesc() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseDesc() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSingleArgsExecutor(t *testing.T) {
	type args struct {
		key []string
		f   func(string) (string, bool, error)
	}
	tests := []struct {
		name    string
		value   string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				key: []string{"key"},
				f: func(string) (string, bool, error) {
					return "value", true, nil
				},
			},
			value:   "value",
			wantErr: false,
		},
		{
			name: "lack key",
			args: args{
				key: []string{},
				f: func(string) (string, bool, error) {
					return "value", true, nil
				},
			},
			value:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := SingleArgsExecutor(tt.args.f)(tt.args.key...)
			if tt.wantErr && err == nil {
				t.Errorf("SingleArgsExecutor()  want %v err=%v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(got, tt.value) {
				t.Errorf("SingleArgsExecutor() = %v,  want=%v", got, tt.value)
			}
		})
	}
}

func TestEmptyArgsExecutor(t *testing.T) {
	type args struct {
		f func() (string, bool, error)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				f: func() (string, bool, error) {
					return "value", true, nil
				},
			},
			want:    "value",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := EmptyArgsExecutor(tt.args.f)()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDesc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmptyArgsExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_evalImpl_Eval(t *testing.T) {
	RegisterExecutor("ip_of_interface", SingleArgsExecutor(func(interfaceName string) (string, bool, error) {
		if interfaceName == "eth0" {
			return "127.0.0.1", false, nil
		}
		return "", false, nil
	}))

	RegisterKeyValueExecutor("kv_of_consul", func(key string) (string, bool, error) {
		if key == "key" {
			return "value", true, nil
		}
		return "", false, nil
	})
	type args struct {
		desc string
	}
	tests := []struct {
		name    string
		impl    evalImpl
		args    args
		wantStr string
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				desc: "${ip_of_interface,eth0}:8080",
			},
			wantStr: "127.0.0.1:8080",
			wantErr: false,
		},
		{
			name: "prefix",
			args: args{
				desc: "http://${ip_of_interface,eth0}:8080$%@*@_)(8979723${kv_of_consul, key}",
			},
			wantStr: "http://127.0.0.1:8080$%@*@_)(8979723value",
			wantErr: false,
		},
		{
			name: "none",
			args: args{
				desc: "1233ou21312|{{_*~!",
			},
			wantStr: "1233ou21312|{{_*~!",
			wantErr: false,
		},
		{
			name: "executor_not_found",
			args: args{
				desc: "${executor_not_found,123,123}",
			},
			wantStr: "",
			wantErr: true,
		},
		{
			name: "bad_syntax",
			args: args{
				desc: "${bad_syntax,",
			},
			wantStr: "",
			wantErr: true,
		},
		{
			name: "bad_syntax2",
			args: args{
				desc: "${}",
			},
			wantStr: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, err := tt.impl.Value(tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalImpl.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStr != tt.wantStr {
				t.Errorf("evalImpl.Value() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestValue(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "bad_syntax2",
			args: args{
				s: "12345",
			},
			want:    "12345",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Value(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	RegisterKeyValueExecutor("valueof", func(key string) (string, bool, error) {
		return key, true, nil
	})
	type SubItem struct {
		Slice []string
		Value string
	}
	type Service struct {
		SubItem
		SubItemPtr   *SubItem
		SubItemField SubItem
		Slice1       []string
		String       string
		Interface    interface{}
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			args: args{
				v: &Service{
					SubItem: SubItem{
						Value: "${valueof,123}",
						Slice: []string{"apple", "${valueof,456}"},
					},
					SubItemField: SubItem{
						Value: "${valueof,abcdefg123}",
						Slice: []string{"apple", "${valueof,orange}"},
					},
					SubItemPtr: &SubItem{
						Value: "${valueof,QWERTY}",
						Slice: []string{"apple", "${valueof,orange}"},
					},
					Slice1: []string{"${valueof,ASDFGH}", "${valueof,*()_+}"},
					String: "${valueof,zxcvbn}",
				},
			},
			want: &Service{
				SubItem: SubItem{
					Value: "123",
					Slice: []string{"apple", "456"},
				},
				SubItemField: SubItem{
					Value: "abcdefg123",
					Slice: []string{"apple", "orange"},
				},
				SubItemPtr: &SubItem{
					Value: "${valueof,QWERTY}",
					Slice: []string{"apple", "orange"},
				},
				Slice1: []string{"ASDFGH", "*()_+"},
				String: "zxcvbn",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Complete(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Printf("%s\n", jsonize.V(tt.args.v, true))
		})
	}
}
