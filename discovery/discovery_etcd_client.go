package discovery

import (
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/text/gstr"
	etcd3 "go.etcd.io/etcd/client/v3"
)

var (
	// etcdClient is the client instance for etcd.
	etcdClient *etcd3.Client
)

// getEtcdClient creates and returns an instance for etcd client.
// It returns the same instance object if it already created one.
func getEtcdClient() (*etcd3.Client, error) {
	if etcdClient != nil {
		return etcdClient, nil
	}
	endpoints := gstr.SplitAndTrim(gcmd.GetWithEnv(EnvKey.Endpoints).String(), ",")
	if len(endpoints) == 0 {
		return nil, gerror.New(`endpoints not found from environment, command-line or configuration file`)
	}
	client, err := etcd3.New(etcd3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	etcdClient = client
	return etcdClient, nil
}
