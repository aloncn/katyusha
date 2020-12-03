package registry

import (
	"encoding/json"
	"github.com/gogf/gf/frame/g"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/resolver"
	"sync"
)

type Watcher struct {
	key       string
	client    *etcd3.Client
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	addresses []resolver.Address
}

func (w *Watcher) Close() {
	w.cancel()
}

func newWatcher(key string, cli *etcd3.Client) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		key:    key,
		client: cli,
		ctx:    ctx,
		cancel: cancel,
	}
	return w
}

func (w *Watcher) GetAllAddresses() []resolver.Address {
	var addresses []resolver.Address
	resp, err := w.client.Get(w.ctx, w.key, etcd3.WithPrefix())
	if err == nil {
		services := extractServices(resp)
		if len(services) > 0 {
			for _, addr := range services {
				v := addr
				addresses = append(addresses, resolver.Address{
					Addr:     v.Address,
					Metadata: &v.Metadata,
				})
			}
		}
	}
	return addresses
}

func (w *Watcher) Watch() chan []resolver.Address {
	out := make(chan []resolver.Address, 10)
	w.wg.Add(1)
	go func() {
		defer func() {
			close(out)
			w.wg.Done()
		}()
		w.addresses = w.GetAllAddresses()
		out <- w.cloneAddresses(w.addresses)

		for watchResponse := range w.client.Watch(w.ctx, w.key, etcd3.WithPrefix()) {
			for _, ev := range watchResponse.Events {
				g.Log().Debugf("watch event: %d, %s", ev.Type, ev.Kv.String())
				switch ev.Type {
				case mvccpb.PUT:
					nodeData := Service{}
					if err := json.Unmarshal(ev.Kv.Value, &nodeData); err != nil {
						g.Log().Error(err)
						continue
					}
					addr := resolver.Address{Addr: nodeData.Address, Metadata: &nodeData.Metadata}
					if w.addAddr(addr) {
						out <- w.cloneAddresses(w.addresses)
					}
				case mvccpb.DELETE:
					nodeData := Service{}
					if err := json.Unmarshal(ev.Kv.Value, &nodeData); err != nil {
						g.Log().Error(err)
						continue
					}
					addr := resolver.Address{Addr: nodeData.Address, Metadata: &nodeData.Metadata}
					if w.removeAddr(addr) {
						out <- w.cloneAddresses(w.addresses)
					}
				}
			}
		}
	}()
	return out
}

func extractServices(resp *etcd3.GetResponse) []Service {
	var services []Service
	if resp == nil || resp.Kvs == nil {
		return services
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			nodeData := Service{}
			if err := json.Unmarshal(v, &nodeData); err != nil {
				g.Log().Errorf("Parse node data error: %v", err)
				continue
			}
			services = append(services, nodeData)
		}
	}
	g.Log().Debugf(`extractServices: %v`, services)
	return services
}

func (w *Watcher) cloneAddresses(in []resolver.Address) []resolver.Address {
	out := make([]resolver.Address, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = in[i]
	}
	return out
}

func (w *Watcher) addAddr(addr resolver.Address) bool {
	// Filter repeated address.
	for _, v := range w.addresses {
		if addr.Addr == v.Addr {
			return false
		}
	}
	w.addresses = append(w.addresses, addr)
	return true
}

func (w *Watcher) removeAddr(addr resolver.Address) bool {
	for i, v := range w.addresses {
		if addr.Addr == v.Addr {
			w.addresses = append(w.addresses[:i], w.addresses[i+1:]...)
			return true
		}
	}
	return false
}