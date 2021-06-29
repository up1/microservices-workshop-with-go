package main

import (
	"demo/cmd"
	"demo/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/circuitbreaker"
	"github.com/up1/microservices-workshop-with-go/common/messaging"
	"github.com/up1/microservices-workshop-with-go/common/tracing"
)

var appName = "account-service"

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting %v\n", appName)

	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	initializeTracing(cfg)
	mc := initializeMessaging(cfg)
	circuitbreaker.ConfigureHystrix([]string{"account-to-data", "account-to-image", "account-to-quotes"}, mc)

	client := &http.Client{}
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
	circuitbreaker.Client = client
	h := service.NewHandler(mc, client)
	s := service.NewServer(cfg, h)
	s.SetupRoutes()

	handleSigterm(func() {
		circuitbreaker.Deregister(mc)
		mc.Close()
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

func initializeMessaging(cfg *cmd.Config) *messaging.AmqpClient {
	if cfg.AmqpConfig.ServerUrl == "" {
		panic("No 'amqp_server_url' set in configuration, cannot start")
	}

	mc := &messaging.AmqpClient{}
	mc.ConnectToBroker(cfg.AmqpConfig.ServerUrl)
	return mc
}
