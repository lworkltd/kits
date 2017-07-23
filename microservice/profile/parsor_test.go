package profile

import (
	"os"
	"reflect"
	"testing"
)

type TestDefaultConfig struct {
	TestDefaultItem `toml:"abc"`
	Top             []int32 `toml:"top_level"`
}

type TestDefaultItem struct {
	M       []string
	N       int32
	R       int             `toml:"r_123_xcvf"`
	S       int             `toml:"s"`
	Weather []string        `toml:"weather"`
	Groups  [][]interface{} `toml:"groups"`
}

func (ts *TestDefaultItem) BeforeParse() {
	ts.M = []string{"abc", "123"}
	ts.N = 1
	ts.S = 22
}

func (ts *TestDefaultItem) AfterParse() {
	ts.M = []string{"abc", "123"}
	ts.N = 100
}

func Test_profileParserImpl_Parse(t *testing.T) {
	var tdc TestDefaultConfig
	os.Setenv("abc.n", "321")
	os.Setenv("abc.r_123_xcvf", "1234")
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		parser  *profileParserImpl
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				v: &tdc,
			},
			parser: &profileParserImpl{f: "test.toml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parser.Parse(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("profileParserImpl.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tdc.N != 100 {
				t.Errorf("profileParserImpl.Parse() tdc.N = %v, expect %v", tdc.N, 100)
			}

			if tdc.R != 1234 {
				t.Errorf("profileParserImpl.Parse() tdc.N = %v, expect %v", tdc.R, 1234)
			}
		})
	}
}
func Test_parseDefault(t *testing.T) {
	var tdc TestDefaultConfig
	type args struct {
		v           interface{}
		parseStatus *parseStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				v:           &tdc,
				parseStatus: &parseStatus{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseDefault(tt.args.v, tt.args.parseStatus); (err != nil) != tt.wantErr {
				t.Errorf("parseDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseEnv(t *testing.T) {
	var tdc TestDefaultConfig
	os.Setenv("abc.r_123_xcvf", "1323")
	os.Setenv("top_level", "[1,2,3,4,5,6]")
	os.Setenv("abc.weather", `["spring","winter"]`)
	os.Setenv("abc.groups", `[["spring","winter"],[1,2]]`)
	type args struct {
		v           interface{}
		parseStatus *parseStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				v:           &tdc,
				parseStatus: &parseStatus{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseEnv(tt.args.v, tt.args.parseStatus); (err != nil) != tt.wantErr {
				t.Errorf("parseEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tdc.R != 1323 {
				t.Errorf("parseEnv() tdc.R = %v, expect 1323", tdc.R)
			}
			if equal := reflect.DeepEqual(tdc.Top, []int32{1, 2, 3, 4, 5, 6}); !equal {
				t.Errorf("parseEnv() tdc.Top = %v, expect %v", tdc.Top, []int32{1, 2, 3, 4, 5, 6})
			}
			if equal := reflect.DeepEqual(tdc.Weather, []string{"spring", "winter"}); !equal {
				t.Errorf("parseEnv() tdc.Weather = %v, expect %v", tdc.Weather, []string{"spring", "winter"})
			}
		})
	}
}

func Test_parseInit(t *testing.T) {
	type args struct {
		v           interface{}
		parseStatus *parseStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseInit(tt.args.v, tt.args.parseStatus); (err != nil) != tt.wantErr {
				t.Errorf("parseInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseInit0(t *testing.T) {
	type args struct {
		v           reflect.Value
		parseStatus *parseStatus
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseInit0(tt.args.v, tt.args.parseStatus)
		})
	}
}
