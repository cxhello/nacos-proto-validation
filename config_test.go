package main

import (
	"testing"
	"time"

	"github.com/cxhello/nacos-proto-validation/client"
	configpb "github.com/cxhello/nacos-sdk-proto/go/config"
	"google.golang.org/protobuf/proto"
)

func newClient(t *testing.T) *client.NacosClient {
	t.Helper()
	c, err := client.NewNacosClient("127.0.0.1:9848")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	time.Sleep(2 * time.Second) // wait for handshake
	return c
}

func TestConfigPublish(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	req := &configpb.ConfigPublishRequest{
		DataId:  "proto-test.yaml",
		Group:   "DEFAULT_GROUP",
		Tenant:  "",
		Content: "server:\n  port: 8080",
		AdditionMap: map[string]string{
			"type": "yaml",
		},
	}
	resp, err := c.UnaryRequest(req, "ConfigPublishRequest")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	pubResp := resp.(*configpb.ConfigPublishResponse)
	if pubResp.ResultCode != 200 {
		t.Fatalf("expected 200, got %d: %s", pubResp.ResultCode, pubResp.Message)
	}
	t.Logf("ConfigPublish OK")
}

func TestConfigQuery(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Publish first
	pub := &configpb.ConfigPublishRequest{
		DataId: "proto-test.yaml", Group: "DEFAULT_GROUP",
		Content: "server:\n  port: 9090",
	}
	c.UnaryRequest(pub, "ConfigPublishRequest")
	time.Sleep(1 * time.Second)

	// Query
	req := &configpb.ConfigQueryRequest{
		DataId: "proto-test.yaml",
		Group:  "DEFAULT_GROUP",
	}
	resp, err := c.UnaryRequest(req, "ConfigQueryRequest")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	queryResp := resp.(*configpb.ConfigQueryResponse)
	if queryResp.Content != "server:\n  port: 9090" {
		t.Fatalf("content mismatch: %q", queryResp.Content)
	}
	t.Logf("ConfigQuery OK, content=%q, lastModified=%d", queryResp.Content, queryResp.LastModified)
}

func TestConfigRemove(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Publish first
	pub := &configpb.ConfigPublishRequest{
		DataId: "proto-test-remove.yaml", Group: "DEFAULT_GROUP",
		Content: "to-be-removed",
	}
	c.UnaryRequest(pub, "ConfigPublishRequest")
	time.Sleep(1 * time.Second)

	// Remove
	req := &configpb.ConfigRemoveRequest{
		DataId: "proto-test-remove.yaml",
		Group:  "DEFAULT_GROUP",
	}
	resp, err := c.UnaryRequest(req, "ConfigRemoveRequest")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	removeResp := resp.(*configpb.ConfigRemoveResponse)
	if removeResp.ResultCode != 200 {
		t.Fatalf("expected 200, got %d", removeResp.ResultCode)
	}
	t.Logf("ConfigRemove OK")
}

func TestConfigListen(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	// Publish initial config
	pub := &configpb.ConfigPublishRequest{
		DataId: "proto-test-listen.yaml", Group: "DEFAULT_GROUP",
		Content: "version: 1",
	}
	c.UnaryRequest(pub, "ConfigPublishRequest")
	time.Sleep(1 * time.Second)

	// Send BatchListen request (via Unary RPC)
	listenReq := &configpb.ConfigBatchListenRequest{
		Listen: true,
		ConfigListenContexts: []*configpb.ConfigListenContext{
			{DataId: "proto-test-listen.yaml", Group: "DEFAULT_GROUP", Md5: ""},
		},
	}
	resp, err := c.UnaryRequest(listenReq, "ConfigBatchListenRequest")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	listenResp := resp.(*configpb.ConfigChangeBatchListenResponse)
	t.Logf("BatchListen OK, changedConfigs=%d", len(listenResp.ChangedConfigs))

	// Set up push receiver channel
	pushCh := make(chan proto.Message, 1)
	c.WaitForPush("ConfigChangeNotifyRequest", pushCh)

	// Use another connection to modify config, triggering push
	c2 := newClient(t)
	defer c2.Close()
	pub2 := &configpb.ConfigPublishRequest{
		DataId: "proto-test-listen.yaml", Group: "DEFAULT_GROUP",
		Content: "version: 2",
	}
	c2.UnaryRequest(pub2, "ConfigPublishRequest")

	// Wait for push
	select {
	case msg := <-pushCh:
		notify := msg.(*configpb.ConfigChangeNotifyRequest)
		t.Logf("Received ConfigChangeNotify: dataId=%s, group=%s", notify.DataId, notify.Group)
		if notify.DataId != "proto-test-listen.yaml" {
			t.Fatalf("unexpected dataId: %s", notify.DataId)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for ConfigChangeNotifyRequest push")
	}
}
