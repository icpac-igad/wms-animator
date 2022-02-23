package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gocraft/web"
	"github.com/gorilla/handlers"
	"github.com/icpac-igad/wms-animator/internal/api"
	"github.com/icpac-igad/wms-animator/internal/conf"
	log "github.com/sirupsen/logrus"
)

var server *http.Server
var router *web.Router

func createServers() {
	confServ := conf.Configuration.Server

	bindAddress := fmt.Sprintf("%v:%v", confServ.HttpHost, confServ.HttpPort)

	router = initRouter(confServ.BasePath)

	// writeTimeout is slighlty longer than request timeout to allow writing error response
	timeoutSecRequest := conf.Configuration.Server.WriteTimeoutSec
	timeoutSecWrite := timeoutSecRequest + 1

	// ----  Handler chain  --------
	// set CORS handling according to config
	corsOpt := handlers.AllowedOrigins([]string{conf.Configuration.Server.CORSOrigins})
	corsHandler := handlers.CORS(corsOpt)(router)
	compressHandler := handlers.CompressHandler(corsHandler)

	// Use a TimeoutHandler to ensure a request does not run past the WriteTimeout duration.
	// If timeout expires, service returns 503 and a text message
	timeoutHandler := http.TimeoutHandler(compressHandler,
		time.Duration(timeoutSecRequest)*time.Second,
		api.ErrMsgRequestTimeout)

	// more "production friendly" timeouts
	// https://blog.simon-frey.eu/go-as-in-golang-standard-net-http-config-will-break-your-production/#You_should_at_least_do_this_The_easy_path
	server = &http.Server{
		ReadTimeout:  time.Duration(conf.Configuration.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(timeoutSecWrite) * time.Second,
		Addr:         bindAddress,
		Handler:      timeoutHandler,
	}
}

// Serve starts the web service
func Serve() {
	createServers()

	log.Infof("====  Service: %s  Port: %d ==== \n", conf.Configuration.Metadata.Title, conf.Configuration.Server.HttpPort)

	// start http service
	go func() {
		// ListenAndServe returns http.ErrServerClosed when the server receives
		// a call to Shutdown(). Other errors are unexpected.
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}

	}()

	// wait here for interrupt signal (^C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	// Interrupt signal received:  Start shutting down
	log.Infoln("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	// abort after waiting long enough for service to shutdown gracefully
	// this terminates long-running processes, which otherwise block shutdown
	abortTimeoutSec := conf.Configuration.Server.WriteTimeoutSec + 10
	chanCancelFatal := FatalAfter(abortTimeoutSec, "Timeout on shutdown - aborting.")

	log.Infoln("Service stopped.")

	// cancel the abort since it is not needed
	close(chanCancelFatal)
}
