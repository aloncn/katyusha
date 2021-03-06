package balancer

import (
	"github.com/gogf/gf/container/gtype"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const BlLeastConnection = "katyusha_balancer_least_connection"

type leastConnectionPickerBuilder struct{}

type leastConnectionPicker struct {
	nodes []*leastConnectionPickerNode
}

type leastConnectionPickerNode struct {
	balancer.SubConn
	inflight *gtype.Int
}

func init() {
	balancer.Register(newLeastConnectionBuilder())
}

// newLeastConnectionBuilder creates a new leastConnection balancer builder.
func newLeastConnectionBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(
		BlLeastConnection,
		&leastConnectionPickerBuilder{},
		base.Config{HealthCheck: true},
	)
}

func (*leastConnectionPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.V2Picker {
	if len(buildInfo.ReadySCs) == 0 {
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}
	var nodes []*leastConnectionPickerNode
	for subConn, _ := range buildInfo.ReadySCs {
		nodes = append(nodes, &leastConnectionPickerNode{
			SubConn:  subConn,
			inflight: gtype.NewInt(),
		})
	}
	return &leastConnectionPicker{
		nodes: nodes,
	}
}

func (p *leastConnectionPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	result := balancer.PickResult{}
	if len(p.nodes) == 0 {
		return result, balancer.ErrNoSubConnAvailable
	}
	var pickedNode *leastConnectionPickerNode
	if len(p.nodes) == 1 {
		pickedNode = p.nodes[0]
	} else {
		for _, node := range p.nodes {
			if pickedNode == nil {
				pickedNode = node
			} else if node.inflight.Val() < pickedNode.inflight.Val() {
				pickedNode = node
			}
		}
	}
	pickedNode.inflight.Add(1)
	result.SubConn = pickedNode
	result.Done = func(info balancer.DoneInfo) {
		pickedNode.inflight.Add(-1)
	}
	return result, nil
}
