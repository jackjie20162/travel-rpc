package server

import "github.com/zeromicro/go-zero/zrpc"

// Register reserves the RPC registration point for generated protobuf services.
// Generated registration code will be wired here after the protobuf toolchain is configured.
func Register(_ *zrpc.RpcServer) {}
