package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"redis/interface/tcp"
	"redis/pkg/logger"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address string        `yaml:"address"`
	MaxConn uint32        `yaml:"max_connect"`
	Timeout time.Duration `yaml:"timeout"`
}

var ClientCounter int

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("tcp server start at %s", cfg.Address))
	ListenAndServe(listener, handler, closeChan)

	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		select {
		case <-closeChan:
			logger.Info("get exit signal")
		case er := <-errCh:
			logger.Info(fmt.Sprintf("accept error: %s", er.Error()))
		}
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitGroup sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Infof("accept occurs temporary error: %v, retry in 5 ms", err)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			errCh <- err
			break
		}

		logger.Info("accept link")
		waitGroup.Add(1)
		ClientCounter++
		go func() {
			defer func() {
				waitGroup.Done()
				ClientCounter--
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitGroup.Wait()
}
