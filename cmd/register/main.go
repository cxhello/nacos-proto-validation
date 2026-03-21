package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cxhello/nacos-proto-validation/client"
	namingpb "github.com/cxhello/nacos-sdk-proto/go/naming"
)

func main() {
	c, err := client.NewNacosClient("127.0.0.1:9848")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
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
			Metadata:    map[string]string{"version": "1.0", "source": "proto-validation"},
		},
	}
	resp, err := c.UnaryRequest(req, "InstanceRequest")
	if err != nil {
		log.Fatalf("register: %v", err)
	}
	instResp := resp.(*namingpb.InstanceResponse)
	fmt.Printf("InstanceRegister OK, resultCode=%d, type=%s\n", instResp.ResultCode, instResp.Type)
	fmt.Println("Keeping connection alive for 60s, check Nacos console now...")
	time.Sleep(60 * time.Second)
	fmt.Println("Done, closing connection")
}
