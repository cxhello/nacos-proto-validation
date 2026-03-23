package main

import (
	"encoding/json"
	"fmt"

	configpb "github.com/cxhello/nacos-sdk-proto/go/config"
	namingpb "github.com/cxhello/nacos-sdk-proto/go/naming"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	pjEmitAll = protojson.MarshalOptions{EmitDefaultValues: true}
	pjDefault = protojson.MarshalOptions{}
)

func compare(label string, msg proto.Message) {
	fmt.Printf("--- %s ---\n", label)
	j, _ := json.Marshal(msg)
	p1, _ := pjEmitAll.Marshal(msg)
	p2, _ := pjDefault.Marshal(msg)
	fmt.Printf("  encoding/json:               %s\n", j)
	fmt.Printf("  protojson(EmitDefault=true):  %s\n", p1)
	fmt.Printf("  protojson(EmitDefault=false): %s\n", p2)
	fmt.Println()
}

func main() {
	// Case 1: Normal values
	compare("Normal values", &configpb.ConfigPublishRequest{
		RequestId: "req-001",
		DataId:    "app.yaml",
		Group:     "DEFAULT_GROUP",
		Content:   "hello",
		AdditionMap: map[string]string{"type": "yaml"},
	})

	// Case 2: All zero values
	compare("All zero values", &configpb.ConfigQueryRequest{})

	// Case 3: ephemeral=false (persistent instance)
	compare("ephemeral=false (persistent)", &namingpb.InstanceRequest{
		Namespace:   "public",
		ServiceName: "svc",
		GroupName:   "DEFAULT_GROUP",
		Type:        "registerInstance",
		Instance: &namingpb.Instance{
			Ip:        "192.168.1.1",
			Port:      8080,
			Weight:    1.0,
			Healthy:   true,
			Enabled:   true,
			Ephemeral: false,
		},
	})

	// Case 4: Empty map vs nil map
	compare("Empty map ({})", &configpb.ConfigPublishRequest{
		DataId:      "app.yaml",
		Group:       "DEFAULT_GROUP",
		Content:     "test",
		AdditionMap: map[string]string{},
	})

	compare("Nil map (not set)", &configpb.ConfigPublishRequest{
		DataId:  "app.yaml",
		Group:   "DEFAULT_GROUP",
		Content: "test",
	})

	// Case 5: int32 zero value (port=0)
	compare("Port=0 (zero int)", &namingpb.Instance{
		Ip:        "192.168.1.1",
		Port:      0,
		Weight:    1.0,
		Healthy:   true,
		Ephemeral: true,
	})

	// Case 6: weight=0.0 (zero double)
	compare("Weight=0.0 (zero double)", &namingpb.Instance{
		Ip:        "192.168.1.1",
		Port:      8080,
		Weight:    0.0,
		Ephemeral: true,
	})
}
