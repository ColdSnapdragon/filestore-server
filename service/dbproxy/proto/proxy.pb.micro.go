// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proxy.proto

package service

import (
	fmt "fmt"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "go-micro.dev/v4/api"
	client "go-micro.dev/v4/client"
	server "go-micro.dev/v4/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for DBProxyService service

func NewDBProxyServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for DBProxyService service

type DBProxyService interface {
	// 请求执行sql动作
	ExecuteAction(ctx context.Context, in *ReqExec, opts ...client.CallOption) (*RespExec, error)
}

type dBProxyService struct {
	c    client.Client
	name string
}

func NewDBProxyService(name string, c client.Client) DBProxyService {
	return &dBProxyService{
		c:    c,
		name: name,
	}
}

func (c *dBProxyService) ExecuteAction(ctx context.Context, in *ReqExec, opts ...client.CallOption) (*RespExec, error) {
	req := c.c.NewRequest(c.name, "DBProxyService.ExecuteAction", in)
	out := new(RespExec)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for DBProxyService service

type DBProxyServiceHandler interface {
	// 请求执行sql动作
	ExecuteAction(context.Context, *ReqExec, *RespExec) error
}

func RegisterDBProxyServiceHandler(s server.Server, hdlr DBProxyServiceHandler, opts ...server.HandlerOption) error {
	type dBProxyService interface {
		ExecuteAction(ctx context.Context, in *ReqExec, out *RespExec) error
	}
	type DBProxyService struct {
		dBProxyService
	}
	h := &dBProxyServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&DBProxyService{h}, opts...))
}

type dBProxyServiceHandler struct {
	DBProxyServiceHandler
}

func (h *dBProxyServiceHandler) ExecuteAction(ctx context.Context, in *ReqExec, out *RespExec) error {
	return h.DBProxyServiceHandler.ExecuteAction(ctx, in, out)
}