package main

import (
	"context"
	"fmt"
	"github.com/mkaykisiz/sender"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/joho/godotenv"
	envvars "github.com/mkaykisiz/sender/configs/env-vars"
	"github.com/mkaykisiz/sender/internal/client/messageclient"
	"github.com/mkaykisiz/sender/internal/localization"
	"github.com/mkaykisiz/sender/internal/middlewares"
	"github.com/mkaykisiz/sender/internal/service"
	mongostore "github.com/mkaykisiz/sender/internal/store/mongo"
	redisstore "github.com/mkaykisiz/sender/internal/store/redis"
	httptransport "github.com/mkaykisiz/sender/internal/transport/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	var l log.Logger
	{
		l = log.NewLogfmtLogger(os.Stdout)
		l = log.With(l, "time", log.DefaultTimestampUTC)
	}

	_ = godotenv.Load()

	var ev *envvars.EnvVars
	var err error
	{
		ev, err = envvars.LoadEnvVars()
		if err != nil {
			_ = l.Log("error", err.Error())
			return
		}
	}

	err = localization.InitializeBundle(ev.Localization)
	if err != nil {
		_ = l.Log("error", err.Error())
		return
	}

	var ms mongostore.Store
	{
		ms, err = mongostore.NewStore(ev.Mongo)
		if err != nil {
			_ = l.Log("error", err.Error())
			return
		}
		seedMessages(context.Background(), l, ms, ev.Configs.StartMessageCount)
	}

	var rs redisstore.Store
	{
		rs, err = redisstore.NewStore(ev.Redis)
		if err != nil {
			_ = l.Log("error", err.Error())
			return
		}
	}

	var mc messageclient.MessageClient
	{
		mc = messageclient.NewClient(ev.MessageClient.Url, ev.MessageClient.AuthKey, ev.MessageClient.MaxRetries, ev.MessageClient.RetryDelay*time.Second)
	}

	var w *service.Worker
	{
		w = service.NewWorker(mc, ms, rs, log.With(l, "component", "worker"), int64(ev.Configs.StartMessageCount))
	}

	var s sender.Service
	{
		s = service.NewService(l, ms, rs, ev.Configs, ev.Service.Environment, w)
		s.StartSendMessage(ev.Configs.StartMessageCount, ev.Configs.SendMessageDelay)
	}

	var lm middlewares.Middleware
	{
		lm = middlewares.NewLoggingMiddleware(l)

		s = lm(s)
	}

	var h http.Handler
	{
		h = httptransport.MakeHTTPHandler(log.With(l, "transport", "http"), s)
	}

	var hs *http.Server
	{
		hs = &http.Server{
			Addr:           ev.HTTPServer.Address,
			ReadTimeout:    ev.HTTPServer.ReadTimeout,
			WriteTimeout:   ev.HTTPServer.WriteTimeout,
			IdleTimeout:    ev.HTTPServer.IdleTimeout,
			MaxHeaderBytes: ev.HTTPServer.MaxHeaderBytes,
			Handler:        h,
		}
	}

	// set service health status to true
	sender.HEALTH_STATUS.SetStatus(true)

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		_ = l.Log("msg", "signal received", "signal", <-c)
		errs <- nil
	}()

	go func() {
		_ = l.Log("transport", "http", "address", ev.HTTPServer.Address)

		err = hs.ListenAndServe()
		if err != http.ErrServerClosed {
			errs <- err
		}
	}()

	err = <-errs
	if err != nil {
		_ = l.Log("error", err.Error())
		return
	}

	// give time to new pods to be ready
	time.Sleep(ev.Service.ShutdownSleepDuration)

	// set service health status to false so that k8s will not send traffic to this pod
	sender.HEALTH_STATUS.SetStatus(false)

	// give time to health checker to realize that this pod is unhealthy
	time.Sleep(ev.Service.ShutdownSleepDuration)

	ctx, cf := context.WithTimeout(context.Background(), ev.HTTPServer.ShutdownTimeout)
	defer cf()
	if err := hs.Shutdown(ctx); err != nil {
		_ = l.Log("error", err.Error())
	}

	if err := rs.Close(); err != nil {
		_ = l.Log("error", err.Error())
	}

	if err := ms.Close(); err != nil {
		_ = l.Log("error", err.Error())
	}

	_ = l.Log("shutdown", ev.Service.Name)
}

func seedMessages(ctx context.Context, l log.Logger, ms mongostore.Store, startMessageCount int) {
	count, err := ms.Count(ctx, mongostore.MessageFilter{Status: []string{mongostore.STATUS_PENDING}})
	if err != nil {
		_ = l.Log("method", "seedMessages", "error", err.Error())
		return
	}

	if count < int64(startMessageCount) {
		_ = l.Log("method", "seedMessages", "msg", "seeding messages", "count", count, "target", startMessageCount)

		var messages []sender.MessageTransaction
		for i := 0; i < 20; i++ {
			msg := sender.MessageTransaction{
				ID:        primitive.NewObjectID(),
				Content:   fmt.Sprintf("Auto generated message %d", i),
				Recipient: "+905551111111",
				Status:    mongostore.STATUS_PENDING,
				CreatedAt: time.Now(),
			}
			messages = append(messages, msg)
		}

		if err := ms.InsertMany(ctx, messages); err != nil {
			_ = l.Log("method", "seedMessages", "error", err.Error())
		} else {
			_ = l.Log("method", "seedMessages", "msg", "seeded 20 messages")
		}
	} else {
		_ = l.Log("method", "seedMessages", "msg", "we found enough messages")
	}
}
