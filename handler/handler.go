package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type result struct {
	SampleSize int
	Data       []float64
}

type PerftestHandler struct {
	ApiUrl string
}

func NewPerftestHandler() *PerftestHandler {
	h := new(PerftestHandler)
	h.ApiUrl = "http://127.0.0.1:3000"

	return h
}

func (h *PerftestHandler) average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}

func (h *PerftestHandler) makeData(sample_size int, number_of_copy int) []float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	l := make([]float64, sample_size * number_of_copy)

	for i := 0; i < sample_size; i++ {
		l[i] = r.Float64()
		for j := 0; j < number_of_copy; j++ {
			l[i+ sample_size*j] = l[i]
		}
	}

	return l
}

func (h *PerftestHandler) RawDataSourceHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sample_size := 200
	num_copy := 1
	response_time := 0
	var err error

	v := r.URL.Query().Get("sample_size")
	if v != "" {
		sample_size, err = strconv.Atoi(v)
		if err != nil {
			log.Printf("bad sample_size: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	v = r.URL.Query().Get("num_copy")
	if v != "" {
		num_copy, err = strconv.Atoi(v)
		if err != nil {
			log.Printf("bad num_copy: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	v = r.URL.Query().Get("response_time")
	if v != "" {
		response_time, err = strconv.Atoi(v)
		if err != nil {
			log.Printf("bad response_time: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var timer <-chan time.Time
	if response_time > 0 {
		timer = time.After(time.Duration(response_time) * time.Millisecond)
	}

	data := &result{sample_size*num_copy, h.makeData(sample_size, num_copy)}

	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if response_time > 0 {
		select {
		case <-timer:
		}
	}

	fmt.Fprintf(w, string(b))
}

func (h *PerftestHandler) GetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	resp, err := http.Get(fmt.Sprintf("%s/ds/%d", h.ApiUrl, sample_size))
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

	var data result
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("unable to parse reponse body: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	result := &result{sample_size, []float64{h.average(data.Data)}}
	b, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, string(b))
}
