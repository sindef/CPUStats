package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Get the current CPU load averages from /proc/loadavg returning a float64 for 1m, 5m and 15m
func GetLoadAvg() (float64, float64, float64, error) {
	raw, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	loadavg := strings.Fields(string(raw))
	if err != nil {
		return 0, 0, 0, err
	}
	return strToFloat(loadavg[0]), strToFloat(loadavg[1]), strToFloat(loadavg[2]), nil

}

func strToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func main() {

	//Define Prometheus Gauges
	var Load1Gauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "load_1m",
			Help: "Load average over the last minute",
		},
	)
	var Load5Gauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "load_5m",
			Help: "Load average over the last 5 minutes",
		},
	)

	var Load15Gauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "load_15m",
			Help: "Load average over the last 15 minutes",
		},
	)

	//Create Prometheus Registry and register the vars
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		Load1Gauge,
		Load5Gauge,
		Load15Gauge,
	)

	//Set the values of the metrics
	go func() {
		for {
			Load1, Load5, Load15, err := GetLoadAvg()
			if err != nil {
				panic(err)
			}
			Load1Gauge.Set(Load1)
			Load5Gauge.Set(Load5)
			Load15Gauge.Set(Load15)

			time.Sleep(time.Second * 5)
		}
	}()

	//Start the server
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil)) //Listen on port 8080
}
