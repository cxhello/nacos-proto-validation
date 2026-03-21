package client

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type NacosClient struct {
	*Connection
}

func NewNacosClient(serverAddr string) (*NacosClient, error) {
	conn, err := NewConnection(serverAddr)
	if err != nil {
		return nil, err
	}
	return &NacosClient{Connection: conn}, nil
}

// UnaryRequest sends a request via Unary RPC and returns the parsed response
func (c *NacosClient) UnaryRequest(req proto.Message, typeName string) (proto.Message, error) {
	payload, err := BuildPayload(req, typeName)
	if err != nil {
		return nil, err
	}
	resp, err := c.requestStub.Request(context.Background(), payload)
	if err != nil {
		return nil, fmt.Errorf("unary rpc: %w", err)
	}
	msg, respType, err := ParsePayload(resp)
	if err != nil {
		return nil, fmt.Errorf("parse response (%s): %w", respType, err)
	}
	// Check for ErrorResponse
	if respType == "ErrorResponse" {
		return nil, fmt.Errorf("server error: %v", msg)
	}
	return msg, nil
}

// BiStreamSend sends a request via BiStream (fire-and-forget)
func (c *NacosClient) BiStreamSend(req proto.Message, typeName string) error {
	payload, err := BuildPayload(req, typeName)
	if err != nil {
		return err
	}
	return c.biStream.Send(payload)
}

// WaitForPush registers a push handler to receive async pushes via BiStream
func (c *NacosClient) WaitForPush(targetType string, ch chan<- proto.Message) {
	c.SetPushHandler(func(typeName string, msg proto.Message) {
		if typeName == targetType {
			ch <- msg
		}
	})
}
