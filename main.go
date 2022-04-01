// net_conntrack_dialer_conn_closed_total{dialer_name="alertmanager"} 0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

type label struct {
	Name   			string 
	Cardinality 	int    
}

type config struct {
	// the number of monitored devices, such as being either individual services, hosts, processes, etc.
	DeviceCount int
	// the number of distinct metrics those devices report, e.g.: cpu, mem, errors, ...
	MetricCount int
	// an array of labels that are reported per metric
	Labels []label
}

var conf config = config{ DeviceCount : 2, MetricCount : 2, Labels : []label{{Name : "group", Cardinality : 3}}}
var label_cardinality int = 3

func receiveConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf(string(body))
		err = json.Unmarshal(body, &conf)
		if err != nil {
			fmt.Printf("Config payload wrong")
			log.Println("Config payload is wrong")
			panic(err)
		}
		cardinality := conf.MetricCount * conf.DeviceCount
		label_cardinality = 1
		for l_i := 0; l_i < len(conf.Labels); l_i++ {
			label_cardinality = label_cardinality * conf.Labels[l_i].Cardinality
		}
		log.Println("Received a new service config")
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "{ \"cardinality\" : %d }", cardinality * label_cardinality)
		w.Write(buf.Bytes())
		defer r.Body.Close()
	default:
		fmt.Fprintf(w, "Only POST method is supported.")
	}
	defer r.Body.Close()
}

func handleIcon(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
}

func generateMetric(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	// iterate through the permutations
	for i := 0; i < conf.DeviceCount; i++ {
		device := fmt.Sprintf("device_%d", i)
		for m_i := 0; m_i < conf.MetricCount; m_i++ {
			for c_i := 0; c_i < label_cardinality; c_i++ {
				fmt.Fprintf(&buf,"metric_%d{device=\"%s\"", m_i, device)
				for l_i := 0; l_i < len(conf.Labels); l_i++ {
					fmt.Fprintf(&buf, ",%s=\"%s_%d\"", conf.Labels[l_i].Name, conf.Labels[l_i].Name, c_i % conf.Labels[l_i].Cardinality)
				}
				fmt.Fprintf(&buf, "} %d\n", rand.Intn(10) + 10)
			}
		}
	}
	w.Write(buf.Bytes()) 
	defer r.Body.Close()
}

func main() {
	port := 8080
	if len(os.Args) > 1 {
		arg := os.Args[1]
		fmt.Printf("Start prometheus metric cardinality generator service at port: %s\n", arg)
		i1, err := strconv.Atoi(arg)
		if err == nil {
			port = i1
		}
	} else {
		fmt.Printf("Start prometheus metric cardinality generator at default port: %d\n", port)
	}

	http.HandleFunc("/favicon.ico", handleIcon)
	http.HandleFunc("/metrics", generateMetric)
	http.HandleFunc("/config", receiveConfig)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		panic(err)
	}
}