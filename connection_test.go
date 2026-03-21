package main

import (
	"testing"

	"github.com/cxhello/nacos-proto-validation/client"
)

func TestConnectionHandshake(t *testing.T) {
	c, err := client.NewNacosClient("127.0.0.1:9848")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	t.Logf("Connection established, connectionId=%s", c.ConnectionId())
	if c.ConnectionId() == "" {
		t.Fatal("Expected non-empty connectionId")
	}
}
