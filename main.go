package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dialogbox/perftestgo/handler"
	"github.com/julienschmidt/httprouter"
)

var port = 3000
var host = "127.0.0.1"

func main() {
	if len(os.Args) >= 2 {
		port, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) >= 3 {
		host = os.Args[2]
	}

	router := httprouter.New()
	h := handler.NewPerftestHandler()
	router.GET("/ds", h.RawDataSourceHandler)
	// router.GET("/perftest/get/:sample_size", h.GetHandler)

	fmt.Printf("http://%s:%d\n", host, port)

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router)
}
