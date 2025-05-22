package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type MetricsUpdater interface {
	UpdateGauge(string, float64)
	UpdateCounter(string, int64)
}

func (ms *MemStorage) UpdateGauge(name string, value float64) {
	ms.Gauge[name] = value
}

func (ms *MemStorage) UpdateCounter(name string, delta int64) {
	ms.Counter[name] += delta
}

var ms = MemStorage{
	Gauge:   make(map[string]float64),
	Counter: make(map[string]int64),
}

func saveMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type not supported", http.StatusUnsupportedMediaType)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	if parts[2] == "" {
		http.Error(w, "empty metric name", http.StatusNotFound)
		return
	}

	var us MetricsUpdater = &ms

	switch parts[1] {
	case typeGauge:
		value, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			http.Error(w, "invalid metric delta", http.StatusBadRequest)
			return
		}
		us.UpdateGauge(parts[2], value)
	case typeCounter:
		delta, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			http.Error(w, "invalid metric delta", http.StatusBadRequest)
			return
		}
		us.UpdateCounter(parts[2], delta)
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", saveMetric)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
