package main

import (
	"fmt"
	"os"
	"os/signal"
	"report/cmd"
	"report/service"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/up1/microservices-workshop-with-go/common/messaging"
	"github.com/up1/microservices-workshop-with-go/common/tracing"
)

var appName = "report-service"

var messagingClient messaging.IMessagingClient

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Println("Starting " + appName + "...")

	cfg := cmd.DefaultConfiguration()
	arg.MustParse(cfg)

	srv := service.NewServer(cfg)
	srv.SetupRoutes()

	initializeTracing(cfg)
	initializeMessaging(cfg)

	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if messagingClient != nil {
			messagingClient.Close()
		}
	})
	srv.Start()
}

func initializeTracing(cfg *cmd.Config) {
	tracing.InitTracing(cfg.ZipkinServerUrl, appName)
}

func onMessage(delivery amqp.Delivery) {
	logrus.Infof("Got a message: %v\n", string(delivery.Body))

	defer tracing.StartTraceFromCarrier(delivery.Headers, "reportService#onMessage").Finish()

	time.Sleep(time.Millisecond * 10)
}

func initializeMessaging(cfg *cmd.Config) {
	if cfg.AmqpConfig.ServerUrl == "" {
		panic("No 'broker_url' set in configuration, cannot start")
	}
	messagingClient = &messaging.AmqpClient{}
	messagingClient.ConnectToBroker(cfg.AmqpConfig.ServerUrl)

	err := messagingClient.SubscribeToQueue("report_queue", appName, onMessage)
	failOnError(err, "Could not start subscribe to vip_queue")

	logrus.Infoln("Successfully initialized messaging")
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

func failOnError(err error, msg string) {
	if err != nil {
		logrus.Errorf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}