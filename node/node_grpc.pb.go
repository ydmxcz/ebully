// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: node.proto

package node

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Node_HearbeatRpc_FullMethodName    = "/node.Node/HearbeatRpc"
	Node_NodeMessageRpc_FullMethodName = "/node.Node/NodeMessageRpc"
	Node_PeerInfosRpc_FullMethodName   = "/node.Node/PeerInfosRpc"
	Node_PeerIdListRpc_FullMethodName  = "/node.Node/PeerIdListRpc"
	Node_MeetRpc_FullMethodName        = "/node.Node/MeetRpc"
	Node_LeaveRpc_FullMethodName       = "/node.Node/LeaveRpc"
)

// NodeClient is the client API for Node service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeClient interface {
	HearbeatRpc(ctx context.Context, in *NodeInfo, opts ...grpc.CallOption) (*HearbeatResp, error)
	NodeMessageRpc(ctx context.Context, in *MessageReq, opts ...grpc.CallOption) (*MessageResp, error)
	PeerInfosRpc(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodeInfoResp, error)
	PeerIdListRpc(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodeIdListResp, error)
	MeetRpc(ctx context.Context, in *NodeInfo, opts ...grpc.CallOption) (*MeetResp, error)
	LeaveRpc(ctx context.Context, in *LeaveReq, opts ...grpc.CallOption) (*LeaveResp, error)
}

type nodeClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeClient(cc grpc.ClientConnInterface) NodeClient {
	return &nodeClient{cc}
}

func (c *nodeClient) HearbeatRpc(ctx context.Context, in *NodeInfo, opts ...grpc.CallOption) (*HearbeatResp, error) {
	out := new(HearbeatResp)
	err := c.cc.Invoke(ctx, Node_HearbeatRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) NodeMessageRpc(ctx context.Context, in *MessageReq, opts ...grpc.CallOption) (*MessageResp, error) {
	out := new(MessageResp)
	err := c.cc.Invoke(ctx, Node_NodeMessageRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) PeerInfosRpc(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodeInfoResp, error) {
	out := new(NodeInfoResp)
	err := c.cc.Invoke(ctx, Node_PeerInfosRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) PeerIdListRpc(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodeIdListResp, error) {
	out := new(NodeIdListResp)
	err := c.cc.Invoke(ctx, Node_PeerIdListRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) MeetRpc(ctx context.Context, in *NodeInfo, opts ...grpc.CallOption) (*MeetResp, error) {
	out := new(MeetResp)
	err := c.cc.Invoke(ctx, Node_MeetRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) LeaveRpc(ctx context.Context, in *LeaveReq, opts ...grpc.CallOption) (*LeaveResp, error) {
	out := new(LeaveResp)
	err := c.cc.Invoke(ctx, Node_LeaveRpc_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeServer is the server API for Node service.
// All implementations must embed UnimplementedNodeServer
// for forward compatibility
type NodeServer interface {
	HearbeatRpc(context.Context, *NodeInfo) (*HearbeatResp, error)
	NodeMessageRpc(context.Context, *MessageReq) (*MessageResp, error)
	PeerInfosRpc(context.Context, *EmptyMessage) (*NodeInfoResp, error)
	PeerIdListRpc(context.Context, *EmptyMessage) (*NodeIdListResp, error)
	MeetRpc(context.Context, *NodeInfo) (*MeetResp, error)
	LeaveRpc(context.Context, *LeaveReq) (*LeaveResp, error)
	mustEmbedUnimplementedNodeServer()
}

// UnimplementedNodeServer must be embedded to have forward compatible implementations.
type UnimplementedNodeServer struct {
}

func (UnimplementedNodeServer) HearbeatRpc(context.Context, *NodeInfo) (*HearbeatResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HearbeatRpc not implemented")
}
func (UnimplementedNodeServer) NodeMessageRpc(context.Context, *MessageReq) (*MessageResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodeMessageRpc not implemented")
}
func (UnimplementedNodeServer) PeerInfosRpc(context.Context, *EmptyMessage) (*NodeInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeerInfosRpc not implemented")
}
func (UnimplementedNodeServer) PeerIdListRpc(context.Context, *EmptyMessage) (*NodeIdListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PeerIdListRpc not implemented")
}
func (UnimplementedNodeServer) MeetRpc(context.Context, *NodeInfo) (*MeetResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MeetRpc not implemented")
}
func (UnimplementedNodeServer) LeaveRpc(context.Context, *LeaveReq) (*LeaveResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LeaveRpc not implemented")
}
func (UnimplementedNodeServer) mustEmbedUnimplementedNodeServer() {}

// UnsafeNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeServer will
// result in compilation errors.
type UnsafeNodeServer interface {
	mustEmbedUnimplementedNodeServer()
}

func RegisterNodeServer(s grpc.ServiceRegistrar, srv NodeServer) {
	s.RegisterService(&Node_ServiceDesc, srv)
}

func _Node_HearbeatRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).HearbeatRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_HearbeatRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).HearbeatRpc(ctx, req.(*NodeInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_NodeMessageRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).NodeMessageRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_NodeMessageRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).NodeMessageRpc(ctx, req.(*MessageReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_PeerInfosRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).PeerInfosRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_PeerInfosRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).PeerInfosRpc(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_PeerIdListRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).PeerIdListRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_PeerIdListRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).PeerIdListRpc(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_MeetRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).MeetRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_MeetRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).MeetRpc(ctx, req.(*NodeInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_LeaveRpc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaveReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).LeaveRpc(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Node_LeaveRpc_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).LeaveRpc(ctx, req.(*LeaveReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Node_ServiceDesc is the grpc.ServiceDesc for Node service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Node_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "node.Node",
	HandlerType: (*NodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HearbeatRpc",
			Handler:    _Node_HearbeatRpc_Handler,
		},
		{
			MethodName: "NodeMessageRpc",
			Handler:    _Node_NodeMessageRpc_Handler,
		},
		{
			MethodName: "PeerInfosRpc",
			Handler:    _Node_PeerInfosRpc_Handler,
		},
		{
			MethodName: "PeerIdListRpc",
			Handler:    _Node_PeerIdListRpc_Handler,
		},
		{
			MethodName: "MeetRpc",
			Handler:    _Node_MeetRpc_Handler,
		},
		{
			MethodName: "LeaveRpc",
			Handler:    _Node_LeaveRpc_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "node.proto",
}
