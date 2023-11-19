package tcp

import (
	"context"
	"net"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

type HandlerFunc func(ctx context.Context, conn net.Conn)
