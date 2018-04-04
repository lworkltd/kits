package invokeimpl

import (
	"fmt"
	"sync"
	"time"

	"github.com/lworkltd/kits/service/grpcinvoke"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

type serviceNode struct {
	host string
	id   string
}

// ResolveWatcher 服务节点的更新检查器
type resolveWatcher struct {
	nodes     []serviceNode
	discovery grpcinvoke.DiscoveryFunc
	target    string
	quit      chan bool
	done      bool
	mutex     sync.Mutex
	nextCnt   int
}

func newResolveWatcher(target string, discovery grpcinvoke.DiscoveryFunc) *resolveWatcher {
	return &resolveWatcher{
		discovery: discovery,
		target:    target,
		quit:      make(chan bool),
	}
}

// Next 解析被更改时调用
func (rw *resolveWatcher) Next() ([]*naming.Update, error) {
	var updates []*naming.Update
	for {
		hosts, ids, err := rw.discovery(rw.target)
		if err != nil {
			continue
		}

		func() {
			rw.mutex.Lock()
			defer rw.mutex.Unlock()

			updates = delectUpdates(rw.nodes, hosts, ids)
			if len(updates) > 0 {
				for i, host := range hosts {
					rw.nodes = append(rw.nodes, serviceNode{
						host: host,
						id:   ids[i],
					})
				}
			}
		}()

		if len(updates) > 0 {
			break
		}

		select {
		case <-rw.quit:
			return nil, fmt.Errorf("Watcher closed")
		case <-time.After(time.Second * 5):
		}
	}

	return updates, nil
}

// Close 关闭
func (rw *resolveWatcher) Close() {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	rw.done = true
	close(rw.quit)
}

// Resolve 返回名字的更新检查器
func (rw *resolveWatcher) Resolve(target string) (naming.Watcher, error) {
	return rw, nil
}

func delectUpdates(nodes []serviceNode, hosts []string, ids []string) []*naming.Update {
	var updates []*naming.Update
	for i, host := range hosts {
		found := false
		for _, node := range nodes {
			if node.host == host {
				found = true
				continue
			}
		}

		if !found {
			updates = append(updates, &naming.Update{
				Op:       naming.Add,
				Addr:     host,
				Metadata: ids[i],
			})
		}
	}

	for _, node := range nodes {
		found := false
		for _, host := range hosts {
			if node.host == host {
				found = true
				continue
			}
		}
		if !found {
			updates = append(updates, &naming.Update{
				Op:   naming.Delete,
				Addr: node.host,
			})
		}
	}

	return updates
}

func dialGrpcConnByDiscovery(target string, discovery grpcinvoke.DiscoveryFunc) (*grpc.ClientConn, error) {
	if discovery == nil {
		return nil, fmt.Errorf("discovery function missing")
	}

	if target == "" {
		return nil, fmt.Errorf("discovery target mission")
	}

	// 实现将服务名到地址的映射关系
	balancer := grpc.RoundRobin(newResolveWatcher(target, discovery))

	// GRPC内部会使用balancer取获取地址，而balancer会根据ResolveWatcher去初始化和更新服务器的地址
	conn, err := grpc.Dial(
		target,
		grpc.WithInsecure(),
		grpc.WithBalancer(balancer),
	)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func dialGrpcConnByAddr(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		target,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func createAddrDiscovery(addrs ...string) grpcinvoke.DiscoveryFunc {
	return func(string) (hosts []string, ids []string, err error) {
		hosts = addrs
		ids = addrs
		return
	}
}
