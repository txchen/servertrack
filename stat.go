package main

import (
	"sync"
)

type loadSlot struct {
	StartUnixTime int64   `json:"startTime"`
	Count         uint64  `json:"recordCount"`
	AvgCPU        float32 `json:"avgCPU"`
	AvgRAM        float32 `json:"avgRAM"`
}

func (ls *loadSlot) appendLoad(cpuLoad float32, ramLoad float32, newStartUnixTime int64) {
	if ls.StartUnixTime != newStartUnixTime {
		// current data in slot is old, reset it
		ls.Count = 1
		ls.AvgCPU = cpuLoad
		ls.AvgRAM = ramLoad
		ls.StartUnixTime = newStartUnixTime
	} else {
		// append data to the slot
		ls.AvgCPU = (ls.AvgCPU*float32(ls.Count) + cpuLoad) / float32(ls.Count+1)
		ls.AvgRAM = (ls.AvgRAM*float32(ls.Count) + ramLoad) / float32(ls.Count+1)
		ls.Count++
	}
}

type serverStat struct {
	sync.RWMutex
	minuteData [60]loadSlot
	hourData   [24]loadSlot
}

func (ss *serverStat) addLoad(unixTime int64, cpuLoad float32, ramLoad float32) {
	// get current minute unixTime, e.g 12:22:34 -> 12:22:00
	minuteTime := unixTime / 60 * 60
	// get current minute, e.g 12:22:34 -> 22
	minute := minuteTime / 60 % 60
	minuteSlot := &ss.minuteData[minute]
	minuteSlot.appendLoad(cpuLoad, ramLoad, minuteTime)

	// get current hour unitTime, e.g 12:22:32 -> 12:00:00
	hourTime := unixTime / 3600 * 3600
	// get current hour, e.g 23:11:10 -> 23
	hour := hourTime / 3600 % 24
	hourSlot := &ss.hourData[hour]
	hourSlot.appendLoad(cpuLoad, ramLoad, hourTime)
}

type allserverStats struct {
	sync.RWMutex
	data map[string]*serverStat
}

func newAllserverStats() *allserverStats {
	return &allserverStats{data: make(map[string]*serverStat)}
}

func (as *allserverStats) getserverStat(serverName string) *serverStat {
	as.RLock()
	ss, _ := as.data[serverName]
	as.RUnlock()
	return ss
}

func (as *allserverStats) addserverStat(serverName string) (*serverStat, bool) {
	as.Lock()
	defer as.Unlock()
	ss, ok := as.data[serverName]
	if !ok {
		newSs := serverStat{}
		as.data[serverName] = &newSs
		return &newSs, true
	}
	return ss, false
}

func (as *allserverStats) recordLoad(serverName string, unixTime int64, cpuLoad float32, ramLoad float32) {
	ss := as.getserverStat(serverName)
	if ss == nil {
		ss, _ = as.addserverStat(serverName)
	}
	ss.Lock()
	defer ss.Unlock()
	ss.addLoad(unixTime, cpuLoad, ramLoad)
}

func (as *allserverStats) getLoad(serverName string, unixTime int64) ([]loadSlot, []loadSlot) {
	ss := as.getserverStat(serverName)
	if ss == nil {
		return nil, nil
	}
	ss.RLock()
	defer ss.RUnlock()
	// get current minute unixTime, e.g 12:22:34 -> 12:22:00
	minuteTime := unixTime / 60 * 60
	var minuteResult []loadSlot
	for i := 59; i >= 0; i-- {
		m := minuteTime - int64(i)*60
		mls := ss.minuteData[(m/60)%60]
		if mls.StartUnixTime == m {
			minuteResult = append(minuteResult, mls)
		}
	}
	// get current hour unitTime, e.g 8:10:23 -> 8:00:00
	hourTime := unixTime / 3600 * 3600
	var hourResult []loadSlot
	for i := 23; i >= 0; i-- {
		h := hourTime - int64(i)*3600
		hls := ss.hourData[(h/3600)%24]
		if hls.StartUnixTime == h {
			hourResult = append(hourResult, hls)
		}
	}
	return minuteResult, hourResult
}

func (as *allserverStats) dumpData() map[string]interface{} {
	as.RLock()
	defer as.RUnlock()
	result := make(map[string]interface{})
	for k, v := range as.data {
		v.RLock()
		defer v.RUnlock()
		result[k] = struct {
			HourData   [24]loadSlot
			MinuteData [60]loadSlot
		}{v.hourData, v.minuteData}
	}
	return result
}
