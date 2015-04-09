package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadSlotAppend(t *testing.T) {
	ls := loadSlot{}
	// append to empty slot
	currentTime := time.Now().UTC().Unix()
	currentMin := currentTime / 60 * 60
	ls.appendLoad(30.0, 60.0, currentMin)
	assert.Equal(t, 30.0, ls.AvgCPU)
	assert.Equal(t, 60.0, ls.AvgRAM)
	assert.Equal(t, uint64(1), ls.Count)
	assert.Equal(t, currentMin, ls.StartUnixTime)
	// now append more
	ls.appendLoad(60.0, 10.0, currentMin)
	assert.Equal(t, 45.0, ls.AvgCPU)
	assert.Equal(t, 35.0, ls.AvgRAM)
	assert.Equal(t, uint64(2), ls.Count)
	assert.Equal(t, currentMin, ls.StartUnixTime)
	// append with new minute, data should be reset
	ls.appendLoad(5.0, 8.0, currentMin+60)
	assert.Equal(t, 5.0, ls.AvgCPU)
	assert.Equal(t, 8.0, ls.AvgRAM)
	assert.Equal(t, uint64(1), ls.Count)
	assert.Equal(t, currentMin+60, ls.StartUnixTime)
}

func TestServerStatAddLoad(t *testing.T) {
	ss := serverStat{}
	// add load, current time
	currentTime := time.Now().UTC().Unix()
	currentMinute := (currentTime / 60) % 60
	currentHour := (currentTime / 3600) % 24
	ss.addLoad(currentTime, 20.0, 2.0)
	ss.addLoad(currentTime, 40.0, 4.0)
	assert.Equal(t, uint64(2), ss.minuteData[currentMinute].Count)
	assert.Equal(t, uint64(2), ss.hourData[currentHour].Count)
	assert.Equal(t, 30.0, ss.minuteData[currentMinute].AvgCPU)
	assert.Equal(t, 30.0, ss.hourData[currentHour].AvgCPU)
	assert.Equal(t, 3.0, ss.minuteData[currentMinute].AvgRAM)
	assert.Equal(t, 3.0, ss.hourData[currentHour].AvgRAM)
	// add load, in same hour, different minute
	nextMinuteTime := currentTime + 60
	if currentMinute == 59 {
		nextMinuteTime -= 120
	}
	nextMinute := (nextMinuteTime / 60) % 60
	ss.addLoad(nextMinuteTime, 60.0, 6.0)
	assert.Equal(t, uint64(2), ss.minuteData[currentMinute].Count)
	assert.Equal(t, uint64(1), ss.minuteData[nextMinute].Count)
	assert.Equal(t, uint64(3), ss.hourData[currentHour].Count)
	assert.Equal(t, 30.0, ss.minuteData[currentMinute].AvgCPU)
	assert.Equal(t, 60.0, ss.minuteData[nextMinute].AvgCPU)
	assert.Equal(t, 40.0, ss.hourData[currentHour].AvgCPU)
	assert.Equal(t, 3.0, ss.minuteData[currentMinute].AvgRAM)
	assert.Equal(t, 6.0, ss.minuteData[nextMinute].AvgRAM)
	assert.Equal(t, 4.0, ss.hourData[currentHour].AvgRAM)
	// add load, in next hour, same minute
	nextHourTime := currentTime + 3600
	nextHour := (nextHourTime / 3600) % 24
	ss.addLoad(nextHourTime, 80.0, 8.0)
	assert.Equal(t, uint64(1), ss.minuteData[currentMinute].Count) // overwritten
	assert.Equal(t, uint64(1), ss.minuteData[nextMinute].Count)
	assert.Equal(t, uint64(3), ss.hourData[currentHour].Count)
	assert.Equal(t, uint64(1), ss.hourData[nextHour].Count)
	assert.Equal(t, 80.0, ss.minuteData[currentMinute].AvgCPU)
	assert.Equal(t, 40.0, ss.hourData[currentHour].AvgCPU)
	assert.Equal(t, 80.0, ss.hourData[nextHour].AvgCPU)
	assert.Equal(t, 8.0, ss.minuteData[currentMinute].AvgRAM)
	assert.Equal(t, 4.0, ss.hourData[currentHour].AvgRAM)
	assert.Equal(t, 8.0, ss.hourData[nextHour].AvgRAM)
}

func TestAllserverStatsAddAndGet(t *testing.T) {
	as := newAllserverStats()
	s1ss, ok := as.addserverStat("server1")
	assert.True(t, ok)
	s1ss2, ok2 := as.addserverStat("server1")
	assert.False(t, ok2)
	assert.Equal(t, s1ss, s1ss2)
	s1ss3 := as.getserverStat("server1")
	assert.Equal(t, s1ss, s1ss3)
	s2ss := as.getserverStat("server2")
	assert.Nil(t, s2ss)
}

func TestAllserverStatsAddConcurrent(t *testing.T) {
	as := newAllserverStats()
	messages := make(chan *serverStat)
	for i := 0; i < 5; i++ {
		go func() {
			nss, _ := as.addserverStat("s1")
			messages <- nss
		}()
	}
	ss := <-messages
	for i := 0; i < 4; i++ {
		newSs := <-messages
		assert.Equal(t, ss, newSs)
	}
}

func TestAllserverStatsRecordAndGet(t *testing.T) {
	as := newAllserverStats()
	currentTime := time.Now().UTC().Unix()
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		j := i
		go func() {
			as.recordLoad("s2", currentTime, 10.0*float32(j), 1.0*float32(j))
			wg.Done()
		}()
	}
	wg.Wait()
	md, hd := as.getLoad("s2", currentTime)
	assert.Equal(t, 1, len(md))
	assert.Equal(t, 20.0, md[0].AvgCPU)
	assert.Equal(t, 2.0, md[0].AvgRAM)
	assert.Equal(t, uint64(5), md[0].Count)
	assert.Equal(t, currentTime/60*60, md[0].StartUnixTime)
	assert.Equal(t, 1, len(hd))
	assert.Equal(t, 20.0, hd[0].AvgCPU)
	assert.Equal(t, 2.0, hd[0].AvgRAM)
	assert.Equal(t, uint64(5), hd[0].Count)
	assert.Equal(t, currentTime/3600*3600, hd[0].StartUnixTime)
	// get non-existing server
	md2, hd2 := as.getLoad("s3", currentTime)
	assert.Nil(t, md2)
	assert.Nil(t, hd2)
}

func BenchmarkRecordLoad(b *testing.B) {
	as := newAllserverStats()
	for n := 0; n < b.N; n++ {
		as.recordLoad("testserver1", time.Now().UTC().Unix(), 15.0, 20.0)
	}
}

func BenchmarkGetLoad(b *testing.B) {
	as := newAllserverStats()
	currentTime := time.Now().UTC().Unix()
	for i := 0; i < 1000; i++ {
		as.recordLoad("testserver2", currentTime+rand.Int63n(86400), rand.Float32()*100, rand.Float32()*100)
	}
	for n := 0; n < b.N; n++ {
		as.getLoad("testserver2", currentTime+int64(n)%86400)
	}
}
