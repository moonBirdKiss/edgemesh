package tunnel

import (
	"errors"
	"k8s.io/klog/v2"
	"strings"
)

type RouteTable struct {
	table map[string]string
}

const (
	TableSize = 10
)

func (r *RouteTable) initTable() {
	klog.Info("[Router]: route table is starting to init")
	// if you want to send traffic to edge-0, you ought to send to edge-1 first
	r.table["edge-0"] = "edge-1,edge-0"
	r.table["edge-1"] = "edge-1"
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		table: make(map[string]string, TableSize),
	}
}

func (r *RouteTable) query(dst string) ([]string, error) {
	path, ok := r.table[dst]
	if !ok {
		klog.Info("[router]: dst ID is not in the route table", dst)
		return nil, errors.New("dst ID is not in the route table")
	}
	ret := strings.Split(path, ",")
	klog.Info("[router]: the route path is: ", ret)
	return ret, nil
}
