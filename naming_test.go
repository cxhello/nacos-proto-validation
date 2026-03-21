package main

import (
	"testing"
	"time"

	namingpb "github.com/cxhello/nacos-sdk-proto/go/naming"
)

func TestInstanceRegister(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	req := &namingpb.InstanceRequest{
		Namespace:   "public",
		ServiceName: "proto-test-service",
		GroupName:   "DEFAULT_GROUP",
		Type:        "registerInstance",
		Instance: &namingpb.Instance{
			Ip:          "192.168.1.100",
			Port:        8080,
			Weight:      1.0,
			Healthy:     true,
			Enabled:     true,
			Ephemeral:   true,
			ClusterName: "DEFAULT",
			ServiceName: "proto-test-service",
			Metadata:    map[string]string{"version": "1.0", "env": "test"},
		},
	}
	resp, err := c.UnaryRequest(req, "InstanceRequest")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	instResp := resp.(*namingpb.InstanceResponse)
	if instResp.ResultCode != 200 {
		t.Fatalf("expected 200, got %d: %s", instResp.ResultCode, instResp.Message)
	}
	t.Logf("InstanceRegister OK, type=%s", instResp.Type)
}

func TestServiceQuery(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Register first
	regReq := &namingpb.InstanceRequest{
		Namespace: "public", ServiceName: "proto-test-query", GroupName: "DEFAULT_GROUP",
		Type: "registerInstance",
		Instance: &namingpb.Instance{
			Ip: "192.168.1.101", Port: 8081, Weight: 1.0,
			Healthy: true, Enabled: true, Ephemeral: true, ClusterName: "DEFAULT",
		},
	}
	c.UnaryRequest(regReq, "InstanceRequest")
	time.Sleep(1 * time.Second)

	// Query
	req := &namingpb.ServiceQueryRequest{
		Namespace:   "public",
		ServiceName: "proto-test-query",
		GroupName:   "DEFAULT_GROUP",
		Cluster:     "DEFAULT",
	}
	resp, err := c.UnaryRequest(req, "ServiceQueryRequest")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	queryResp := resp.(*namingpb.QueryServiceResponse)
	if queryResp.ServiceInfo == nil {
		t.Fatal("expected non-nil serviceInfo")
	}
	t.Logf("ServiceQuery OK, hosts=%d", len(queryResp.ServiceInfo.Hosts))
}

func TestInstanceDeregister(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Register first
	regReq := &namingpb.InstanceRequest{
		Namespace: "public", ServiceName: "proto-test-dereg", GroupName: "DEFAULT_GROUP",
		Type: "registerInstance",
		Instance: &namingpb.Instance{
			Ip: "192.168.1.102", Port: 8082, Weight: 1.0,
			Healthy: true, Enabled: true, Ephemeral: true, ClusterName: "DEFAULT",
		},
	}
	c.UnaryRequest(regReq, "InstanceRequest")
	time.Sleep(1 * time.Second)

	// Deregister
	deregReq := &namingpb.InstanceRequest{
		Namespace: "public", ServiceName: "proto-test-dereg", GroupName: "DEFAULT_GROUP",
		Type: "deregisterInstance",
		Instance: &namingpb.Instance{
			Ip: "192.168.1.102", Port: 8082, Ephemeral: true, ClusterName: "DEFAULT",
		},
	}
	resp, err := c.UnaryRequest(deregReq, "InstanceRequest")
	if err != nil {
		t.Fatalf("deregister: %v", err)
	}
	instResp := resp.(*namingpb.InstanceResponse)
	if instResp.ResultCode != 200 {
		t.Fatalf("expected 200, got %d", instResp.ResultCode)
	}
	t.Logf("InstanceDeregister OK")
}

func TestSubscribe(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Register an instance first
	regReq := &namingpb.InstanceRequest{
		Namespace: "public", ServiceName: "proto-test-sub", GroupName: "DEFAULT_GROUP",
		Type: "registerInstance",
		Instance: &namingpb.Instance{
			Ip: "192.168.1.103", Port: 8083, Weight: 1.0,
			Healthy: true, Enabled: true, Ephemeral: true, ClusterName: "DEFAULT",
		},
	}
	c.UnaryRequest(regReq, "InstanceRequest")
	time.Sleep(1 * time.Second)

	// Subscribe
	subReq := &namingpb.SubscribeServiceRequest{
		Namespace:   "public",
		ServiceName: "proto-test-sub",
		GroupName:   "DEFAULT_GROUP",
		Subscribe:   true,
		Clusters:    "",
	}
	resp, err := c.UnaryRequest(subReq, "SubscribeServiceRequest")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	subResp := resp.(*namingpb.SubscribeServiceResponse)
	if subResp.ServiceInfo == nil {
		t.Fatal("expected non-nil serviceInfo")
	}
	t.Logf("Subscribe OK, hosts=%d", len(subResp.ServiceInfo.Hosts))
	if len(subResp.ServiceInfo.Hosts) == 0 {
		t.Fatal("expected at least 1 host")
	}
}
