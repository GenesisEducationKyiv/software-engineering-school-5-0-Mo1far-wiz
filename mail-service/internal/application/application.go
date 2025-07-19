package application

import (
	"log"
	"mailer/internal/config"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "pkg/protos/mailer"
)

const shutdownTimeout = 5 * time.Second

type mailer interface {
	SendEmail(stream pb.MailService_SendEmailServer) error
}

type Logger interface {
	ConsoleLogInfo(msg string, fields ...zap.Field)
	ConsoleLogError(msg string, fields ...zap.Field)
	Sync() error
}

type Application struct {
	Config   config.ApplicationConfig
	server   *grpc.Server
	listener net.Listener
	Logger   Logger
	Mailer   pb.MailServiceServer
}

func (a *Application) Initialize() {
	lis, err := net.Listen("tcp", a.Config.Addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", a.Config.Addr, err)
	}

	a.listener = lis

	a.server = grpc.NewServer()
	pb.RegisterMailServiceServer(a.server, a.Mailer)
}

func (a *Application) Run() {
	a.Initialize()
	a.Logger.ConsoleLogInfo("application initialized")

	go func() {
		a.Logger.ConsoleLogInfo("gRPC server listening on", zap.String("addr", a.Config.Addr))
		if err := a.server.Serve(a.listener); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.Logger.ConsoleLogInfo("Shutting down gRPC server...")

	if err := a.Logger.Sync(); err != nil {
		log.Panicf("Logger sync error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.Logger.ConsoleLogInfo("gRPC server stopped cleanly")
	case <-time.After(shutdownTimeout):
		a.Logger.ConsoleLogError("Timeout reached; forcing gRPC stop")
		a.server.Stop()
	}

	a.Logger.ConsoleLogInfo("Server exited properly")
}
