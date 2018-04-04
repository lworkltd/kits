package invokeimpl

import (
	"testing"

	"github.com/lworkltd/kits/service/grpcinvoke"
)

func TestRoundRobinSelectorSelect(t *testing.T) {
	tests := []struct {
		name     string
		selector *RoundRobinSelector
		want     string
		want1    string
		wantErr  bool
	}{
		{
			name: "my-service",
			selector: NewRoundRobinSelector(func(string) ([]string, []string, error) {
				return []string{"127.0.0.1:34032"}, []string{"my-service-node1"}, nil
			}),
			want:    "127.0.0.1:34032",
			want1:   "my-service-node1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.selector.Select()
			if (err != nil) != tt.wantErr {
				t.Errorf("RoundRobinSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RoundRobinSelector.Select() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("RoundRobinSelector.Select() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUniqueAddrSelectorSelect(t *testing.T) {
	tests := []struct {
		name     string
		selector *UniqueAddrSelector
		want     string
		want1    string
		wantErr  bool
	}{
		{
			name:     "my-service",
			selector: NewUniqueAddrSelector("127.0.0.1:34032"),
			want:     "127.0.0.1:34032",
			want1:    "127.0.0.1:34032",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.selector.Select()
			if (err != nil) != tt.wantErr {
				t.Errorf("UniqueAddrSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UniqueAddrSelector.Select() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UniqueAddrSelector.Select() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewServiceSelector(t *testing.T) {
	type args struct {
		strategy  LbStrategyType
		discovery grpcinvoke.DiscoveryFunc
	}
	tests := []struct {
		name      string
		args      args
		wantHost1 string
		wantId1   string
		wantHost2 string
		wantId2   string
	}{
		{
			name: "create service selector",
			args: args{
				strategy: RoundRobin,
				discovery: func(string) ([]string, []string, error) {
					return []string{"127.0.0.1:34032", "127.0.0.1:34033"}, []string{"my-service-node1", "my-service-node2"}, nil
				},
			},
			wantHost1: "127.0.0.1:34032",
			wantId1:   "my-service-node1",
			wantHost2: "127.0.0.1:34033",
			wantId2:   "my-service-node2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServiceSelector(tt.args.strategy, tt.args.discovery)
			host1, id1, err := got.Select()
			if err != nil {
				t.Errorf("got err %v", err)
				return
			}
			if host1 != tt.wantHost1 || id1 != tt.wantId1 {
				t.Errorf("expect host = %s id = %s got host = %s id = %s", tt.wantHost1, tt.wantId1, host1, id1)
				return
			}

			host2, id2, err := got.Select()
			if err != nil {
				t.Errorf("got err %v", err)
				return
			}

			if host2 != tt.wantHost2 || id2 != tt.wantId2 {
				t.Errorf("expect host2 = %s id = %s got host2 = %s id = %s", tt.wantHost2, tt.wantId2, host2, id2)
				return
			}
		})
	}
}
