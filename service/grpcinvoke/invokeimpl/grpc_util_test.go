package invokeimpl

import (
	"reflect"
	"testing"

	"github.com/lworkltd/kits/service/grpcinvoke"
	"google.golang.org/grpc/naming"
)

func TestDialGrpcConnByAddr(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				target: "127.0.0.1:8080",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dialGrpcConnByAddr(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("dialGrpcConnByAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if err := got.Close(); err != nil {
					t.Errorf("got.Close() error = %v", err)
				}
			}
		})
	}
}

func TestDialGrpcConnByDiscovery(t *testing.T) {
	type args struct {
		target    string
		discovery grpcinvoke.DiscoveryFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				target: "my-service",
				discovery: func(string) ([]string, []string, error) {
					return []string{}, []string{}, nil
				},
			},
			wantErr: false,
		},

		{
			args: args{
				target: "",
				discovery: func(string) ([]string, []string, error) {
					return []string{}, []string{}, nil
				},
			},
			wantErr: true,
		},
		{
			args: args{
				target:    "my-service",
				discovery: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dialGrpcConnByDiscovery(tt.args.target, tt.args.discovery)
			if (err != nil) != tt.wantErr {
				t.Errorf("dialGrpcConnByDiscovery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if err := got.Close(); err != nil {
					t.Errorf("got.Close() error = %v", err)
					return
				}
			}
		})
	}
}

func countUpdate(updates []*naming.Update) (addCnt, deleteCnt int) {
	for _, u := range updates {
		switch u.Op {
		case naming.Add:
			addCnt++
		case naming.Delete:
			deleteCnt++
		}
	}

	return
}

func TestDelectUpdates(t *testing.T) {
	type args struct {
		nodes []serviceNode
		hosts []string
		ids   []string
	}
	tests := []struct {
		name          string
		args          args
		wantAddCnt    int
		wantDeleteCnt int
	}{
		{
			args: args{
				hosts: []string{"127.0.0.1:8080", "127.0.0.1:8081"},
				ids:   []string{"A", "B"},
			},
			wantAddCnt:    2,
			wantDeleteCnt: 0,
		},
		{
			args: args{
				nodes: []serviceNode{
					{host: "127.0.0.1:8080", id: "A"},
				},
				hosts: []string{"127.0.0.1:8081"},
				ids:   []string{"B"},
			},
			wantAddCnt:    1,
			wantDeleteCnt: 1,
		},
		{
			args: args{
				nodes: []serviceNode{
					{host: "127.0.0.1:8080", id: "A"},
				},
				hosts: []string{"127.0.0.1:8080"},
				ids:   []string{"A"},
			},
			wantAddCnt:    0,
			wantDeleteCnt: 0,
		},

		{
			args: args{
				nodes: []serviceNode{
					{host: "127.0.0.1:8080", id: "A"},
					{host: "127.0.0.1:8081", id: "B"},
				},
				hosts: []string{"127.0.0.1:8080"},
				ids:   []string{"A"},
			},
			wantAddCnt:    0,
			wantDeleteCnt: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := delectUpdates(tt.args.nodes, tt.args.hosts, tt.args.ids)
			adds, deletes := countUpdate(got)
			if adds != tt.wantAddCnt {
				t.Errorf("expect add updates %d got %d", tt.wantAddCnt, adds)
				return
			}
			if deletes != tt.wantDeleteCnt {
				t.Errorf("expect delete updates %d got %d", deletes, tt.wantDeleteCnt)
				return
			}
		})
	}
}

func TestCreateAddrDiscovery(t *testing.T) {
	type args struct {
		addrs []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				addrs: []string{"addr1:8080", "addr2:8080"},
			},
			want: []string{"addr1:8080", "addr2:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := createAddrDiscovery(tt.args.addrs...)
			got, _, _ := d("")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createAddrDiscovery() = %v, want %v", got, tt.want)
			}
		})
	}
}
