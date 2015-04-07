package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type apiResult struct {
	ServerName string     `json:"serverName"`
	MinuteData []loadSlot `json:"minuteData"`
	HourData   []loadSlot `json:"hourData"`
}

var as *allserverStats

func api(w http.ResponseWriter, r *http.Request) {
	serverName := r.URL.Query().Get("servername")
	if serverName == "" {
		http.Error(w, "must specify servername", http.StatusBadRequest)
		return
	}
	serverName = strings.ToLower(serverName)

	switch r.Method {
	case "GET":
		md, hd := as.getLoad(serverName, time.Now().Unix())
		if md == nil {
			http.Error(w, "cannot find data for server:"+serverName, http.StatusNotFound)
			return
		}
		result := apiResult{serverName, md, hd}
		json, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)

	case "POST":
		cpuLoad := r.URL.Query().Get("cpuload")
		ramLoad := r.URL.Query().Get("ramload")
		unixTime := r.URL.Query().Get("unixtime")
		var unixTimeValue int64
		var cpuLoadValue, ramLoadValue float64
		var err error
		if cpuLoad == "" || ramLoad == "" {
			http.Error(w, "must specify cpuload and ramload", http.StatusBadRequest)
			return
		}
		cpuLoadValue, err = strconv.ParseFloat(cpuLoad, 32)
		if err != nil {
			http.Error(w, "cpuload in malformat: "+cpuLoad, http.StatusBadRequest)
			return
		}
		ramLoadValue, err = strconv.ParseFloat(ramLoad, 32)
		if err != nil {
			http.Error(w, "ramload in malformat: "+ramLoad, http.StatusBadRequest)
			return
		}
		if unixTime == "" {
			unixTimeValue = time.Now().Unix()
		}
		as.recordLoad(serverName, unixTimeValue, float32(cpuLoadValue), float32(ramLoadValue))
		w.WriteHeader(200)

	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func dumpData(w http.ResponseWriter, r *http.Request) {
	tmpl := getTemplates()
	if err := tmpl.Execute(w, as.dumpData()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	as = newAllserverStats()
	http.HandleFunc("/", dumpData)
	http.HandleFunc("/load", api)
	fmt.Println("Listening on http://localhost:30000")
	http.ListenAndServe(":30000", nil)
}
