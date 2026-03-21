package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	pb "github.com/cxhello/nacos-sdk-proto/go"
	commonpb "github.com/cxhello/nacos-sdk-proto/go/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type Connection struct {
	conn         *grpc.ClientConn
	requestStub  pb.RequestClient
	biStream     pb.BiRequestStream_RequestBiStreamClient
	connectionId string
	mu           sync.Mutex
	pushHandler  func(typeName string, msg proto.Message)
}

func NewConnection(serverAddr string) (*Connection, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	c := &Connection{
		conn:        conn,
		requestStub: pb.NewRequestClient(conn),
	}

	// Step 1: ServerCheck (Unary RPC)
	if err := c.serverCheck(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("server check: %w", err)
	}

	// Step 2: Open BiStream
	biStreamClient := pb.NewBiRequestStreamClient(conn)
	stream, err := biStreamClient.RequestBiStream(context.Background())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open bistream: %w", err)
	}
	c.biStream = stream

	// Step 3: ConnectionSetup
	if err := c.connectionSetup(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connection setup: %w", err)
	}

	// Step 4: Start BiStream receiver goroutine
	go c.receiveBiStream()

	return c, nil
}

func (c *Connection) ConnectionId() string {
	return c.connectionId
}

func (c *Connection) serverCheck() error {
	req := &commonpb.ServerCheckRequest{}
	payload, _ := BuildPayload(req, "ServerCheckRequest")
	resp, err := c.requestStub.Request(context.Background(), payload)
	if err != nil {
		return err
	}
	msg, _, err := ParsePayload(resp)
	if err != nil {
		return err
	}
	if checkResp, ok := msg.(*commonpb.ServerCheckResponse); ok {
		c.connectionId = checkResp.ConnectionId
		log.Printf("ServerCheck OK, connectionId=%s", c.connectionId)
	}
	return nil
}

func (c *Connection) connectionSetup() error {
	req := &commonpb.ConnectionSetupRequest{
		ClientVersion: "nacos-proto-validation/1.0",
		Labels:        map[string]string{"source": "proto-validation"},
		AbilityTable:  map[string]bool{},
	}
	payload, _ := BuildPayload(req, "ConnectionSetupRequest")
	return c.biStream.Send(payload)
}

func (c *Connection) receiveBiStream() {
	for {
		payload, err := c.biStream.Recv()
		if err != nil {
			log.Printf("BiStream recv error: %v", err)
			return
		}
		msg, typeName, err := ParsePayload(payload)
		if err != nil {
			log.Printf("BiStream parse error (%s): %v", typeName, err)
			// For SetupAckRequest, reply SetupAckResponse even if parse fails
			if typeName == "SetupAckRequest" {
				c.replySetupAck()
			}
			continue
		}

		switch typeName {
		case "SetupAckRequest":
			log.Printf("Received SetupAckRequest, replying SetupAckResponse")
			c.replySetupAck()
		case "ConnectResetRequest":
			log.Printf("Received ConnectResetRequest (not handling, log only)")
		case "HealthCheckRequest":
			log.Printf("Received HealthCheckRequest, ignoring in validation")
		default:
			c.mu.Lock()
			handler := c.pushHandler
			c.mu.Unlock()
			if handler != nil {
				handler(typeName, msg)
			} else {
				log.Printf("Received push: %s", typeName)
			}
		}
	}
}

func (c *Connection) replySetupAck() {
	resp := &commonpb.SetupAckResponse{ResultCode: 200}
	payload, _ := BuildPayload(resp, "SetupAckResponse")
	c.biStream.Send(payload)
}

func (c *Connection) SetPushHandler(handler func(string, proto.Message)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pushHandler = handler
}

func (c *Connection) Close() {
	if c.biStream != nil {
		c.biStream.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
