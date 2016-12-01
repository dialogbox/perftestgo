package main

import (
	"runtime"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"sort"
)

type Result struct {
	SampleSize int
	Median     float64
}

type Data struct {
	Data []float64
}

var port = 3000
var host = "127.0.0.1"
var mode = "gen"
var api_url = "http://127.0.0.1:3000/perftest/gen"
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func median(xs []float64) float64 {
	sort.Float64s(xs)
	l := len(xs)
	if l == 0 {
		return math.NaN()
	} else if l%2 == 0 {
		return (xs[l/2-1] + xs[l/2+1])/2
	} else {
		return float64(xs[l/2])
	}
}

func makeData(sample_size int) []float64 {
	l := make([]float64, sample_size)

	for i := 0; i < sample_size; i++ {
		l = append(l, r.Float64())
	}

	return l
}

func genHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("sample_size")
	sample_size, err := strconv.Atoi(v)

	if err != nil {
		log.Printf("bad sample_size: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	data := &Data{makeData(sample_size)}

	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, string(b))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("sample_size")
	sample_size := 100
	if v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("unable to parse sample_size: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sample_size = i
		}
	}

	resp, err := http.Get(fmt.Sprintf("%s?sample_size=%d", api_url, sample_size))
	if err != nil {
		log.Printf("unable to parse sample_size: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("unable to read reponse body: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("unable to parse reponse body: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	result := &Result{sample_size, median(data.Data)}
	b, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, string(b))
}

func main() {
	flag.IntVar(&port, "port", 3000, "Port number")
	flag.StringVar(&host, "bind", "127.0.0.1", "Bind address")
	flag.StringVar(&mode, "mode", "gen", "gen | get")
	flag.StringVar(&api_url, "api_url", "http://127.0.0.1:3000/perftest/gen", "gen api url")
	flag.Parse()

	// router := httprouter.New()
	// switch mode {
	// case "gen":
	// 	router.GET("/perftest/gen/:sample_size", genHandler)
	// case "get":
	// 	router.GET("/perftest/get/:sample_size", getHandler)
	// }

	switch mode {
	case "gen":
		http.HandleFunc("/perftest/gen", genHandler)
	case "get":
		http.HandleFunc("/perftest/get", getHandler)
	}

	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("http://%s:%d\n", host, port)

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
