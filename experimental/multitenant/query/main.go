package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/experimental/multitenant"
)

func main() {
	var (
		listen = flag.String("http.address", ":80", "webserver listen address")
		dynamo = flag.String("dyanmo", "", "URL of DynamoDB instance")
	)
	flag.Parse()

	dynamoDBCollector := multitenant.NewDynamoDBCollector(*dynamo)

	router := mux.NewRouter()
	app.RegisterTopologyRoutes(dynamoDBCollector, router)
	http.Handle("/", router)
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, nil))
	}()
	common.SignalHandlerLoop()
}
