package client

import (
	"fmt"
	"net"

	pb "github.com/cxhello/nacos-sdk-proto/go"
	commonpb "github.com/cxhello/nacos-sdk-proto/go/common"
	configpb "github.com/cxhello/nacos-sdk-proto/go/config"
	namingpb "github.com/cxhello/nacos-sdk-proto/go/naming"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

var Marshaler = protojson.MarshalOptions{
	EmitDefaultValues: true,
}

var Unmarshaler = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

// type name -> factory function
var typeRegistry = map[string]func() proto.Message{}

func init() {
	// Common
	typeRegistry["ServerCheckResponse"] = func() proto.Message { return &commonpb.ServerCheckResponse{} }
	typeRegistry["SetupAckRequest"] = func() proto.Message { return &commonpb.SetupAckRequest{} }
	typeRegistry["HealthCheckRequest"] = func() proto.Message { return &commonpb.HealthCheckRequest{} }
	typeRegistry["ConnectResetRequest"] = func() proto.Message { return &commonpb.ConnectResetRequest{} }
	typeRegistry["ErrorResponse"] = func() proto.Message { return &commonpb.ErrorResponse{} }
	// Config
	typeRegistry["ConfigQueryResponse"] = func() proto.Message { return &configpb.ConfigQueryResponse{} }
	typeRegistry["ConfigPublishResponse"] = func() proto.Message { return &configpb.ConfigPublishResponse{} }
	typeRegistry["ConfigRemoveResponse"] = func() proto.Message { return &configpb.ConfigRemoveResponse{} }
	typeRegistry["ConfigChangeBatchListenResponse"] = func() proto.Message { return &configpb.ConfigChangeBatchListenResponse{} }
	typeRegistry["ConfigChangeNotifyRequest"] = func() proto.Message { return &configpb.ConfigChangeNotifyRequest{} }
	// Naming
	typeRegistry["InstanceResponse"] = func() proto.Message { return &namingpb.InstanceResponse{} }
	typeRegistry["SubscribeServiceResponse"] = func() proto.Message { return &namingpb.SubscribeServiceResponse{} }
	typeRegistry["QueryServiceResponse"] = func() proto.Message { return &namingpb.QueryServiceResponse{} }
}

func BuildPayload(req proto.Message, typeName string) (*pb.Payload, error) {
	jsonBytes, err := Marshaler.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", typeName, err)
	}
	return &pb.Payload{
		Metadata: &pb.Metadata{
			Type:     typeName,
			ClientIp: localIP(),
			Headers:  map[string]string{},
		},
		Body: &anypb.Any{Value: jsonBytes},
	}, nil
}

func ParsePayload(payload *pb.Payload) (proto.Message, string, error) {
	typeName := payload.GetMetadata().GetType()
	factory, ok := typeRegistry[typeName]
	if !ok {
		return nil, typeName, fmt.Errorf("unknown type: %s", typeName)
	}
	msg := factory()
	if err := Unmarshaler.Unmarshal(payload.GetBody().GetValue(), msg); err != nil {
		return nil, typeName, fmt.Errorf("unmarshal %s: %w", typeName, err)
	}
	return msg, typeName, nil
}

func localIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "127.0.0.1"
}
