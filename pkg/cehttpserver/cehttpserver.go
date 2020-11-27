package cehttpserver

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alitari/ce-go-template/pkg/cehandler"
)

// CeHTTPServer bla
type CeHTTPServer struct {
	debug           bool
	port            int
	path            string
	method          string
	accept          string
	producerHandler *cehandler.CeProducerHandler
	srv             *http.Server
}

// NewCeHTTPServer bla
func NewCeHTTPServer(port int, path string, method string, debug bool, producerHandler *cehandler.CeProducerHandler) *CeHTTPServer {
	chs := new(CeHTTPServer)
	chs.debug = debug
	chs.producerHandler = producerHandler
	mux := http.NewServeMux()
	mux.HandleFunc(path, chs.ServeHTTP)
	chs.srv = &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: mux}

	go func() {
		if err := chs.srv.ListenAndServe(); err != nil {
			log.Printf("cehttpservertransformer.listenAndServe: %v", err)
		}
	}()

	return chs
}

// ShutDown bla
func (chs *CeHTTPServer) ShutDown() error {
	if err := chs.srv.Shutdown(context.TODO()); err != nil {
		return err
	}
	return nil
}

func (chs *CeHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request: %v", r)
	result := chs.producerHandler.SendCe(*r)
	log.Printf("SendCe result: %s", result.Error())
	io.WriteString(w, "event successfully sent!\n")
}
