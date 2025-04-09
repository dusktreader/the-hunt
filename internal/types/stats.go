package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type respCounter struct {
	cMap	map[string]int64
	lock	sync.RWMutex
}

func newRespCounter() *respCounter {
	return &respCounter{
		cMap: make(map[string]int64),
	}
}

func (rc *respCounter) Incr(statusCode int) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	key := fmt.Sprintf("%s (%d)", http.StatusText(statusCode), statusCode)
	val, ok :=rc.cMap[key]
	if !ok {
		rc.cMap[key] = 1
	} else {
		rc.cMap[key] = val + 1
	}
}

func (rc *respCounter) MarshalJSON() ([]byte, error) {
	rc.lock.RLock()
	defer rc.lock.RUnlock()
	return json.Marshal(rc.cMap)
}

type RequestStats struct {
	RequestCount			int64				`json:"total_requests"`
	ResponseCount 	 		int64				`json:"total_responses"`
	ProcTimeMu	 			int64				`json:"total_proc_time_μs"`

	ResponseCountByStatus	*respCounter		`json:"total_responses_by_status"`
}

func NewRequestStats() *RequestStats {
	return &RequestStats{
		ResponseCountByStatus: newRespCounter(),
	}
}

func (rs *RequestStats) AddResponse(statusCode int) {
	rs.ResponseCountByStatus.Incr(statusCode)
	rs.ResponseCount += 1
}

func (rs *RequestStats) AddRequest() {
	rs.RequestCount += 1
}

func (rs *RequestStats) AddTime(d time.Duration) {
	rs.ProcTimeMu += d.Microseconds()
}
