package types

type RequestStats struct {
	TotalRequests	int64	`json:"total_requests"`
	TotalResponses	int64	`json:"total_responses"`
	TotalProcTimeMu int64	`json:"total_proc_time_μs"`
}
