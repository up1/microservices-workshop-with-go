package main

import (
	"demo/cmd"
	"demo/db"
	"demo/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/tracing"
)

var appName = "data-service"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	initializeTracing(cfg)
	dbClient := db.NewPostgresClient(cfg)
	s := service.NewServer(cfg, service.NewHandler(dbClient))
	s.SetupRoutes()

	handleSigterm(func() {
		s.Close()
	})
	s.Start()
}

func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}

func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
}
