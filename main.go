package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
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
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}

func makeData(sample_size int) []float64 {
	l := make([]float64, sample_size)

	for i := range l {
		l[i] = r.Float64()
	}

	return l
}

func genHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	v := ps.ByName("sample_size")
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

func getHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	v := ps.ByName("sample_size")
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

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/perftest/gen/%d", host, port, sample_size))
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

	result := &Result{sample_size, average(data.Data)}
	b, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, string(b))
}

func main() {
	if len(os.Args) >= 2 {
		port, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) >= 3 {
		host = os.Args[2]
	}

	router := httprouter.New()
	router.GET("/perftest/get/:sample_size", getHandler)
	router.GET("/perftest/gen/:sample_size", genHandler)

	fmt.Printf("http://%s:%d\n", host, port)

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router)
}
