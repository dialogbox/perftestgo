package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/dialogbox/perftestgo/handler"
	"github.com/julienschmidt/httprouter"
)

var port = 3000
var host = "127.0.0.1"

func main() {
	flag.IntVar(&port, "port", 3000, "Port number")
	flag.StringVar(&host, "bind", "127.0.0.1", "Bind address")
	flag.Parse()

	router := httprouter.New()
	h := handler.NewPerftestHandler()
	router.GET("/ds", h.RawDataSourceHandler)
	// router.GET("/perftest/get/:sample_size", h.GetHandler)

	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("http://%s:%d\n", host, port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router))
}
