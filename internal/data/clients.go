package data

import (
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"
)


type Client struct {
	Limiter		*rate.Limiter
	LastSeen	time.Time
}

type ClientMap struct {
	ClientMap		map[string]*Client
	LimitRPS		rate.Limit
	LimitBurst		int
	CleanupInterval	time.Duration
	CleanupTimeout	time.Duration
	Mutex			*sync.Mutex
}

func NewClientMap(cfg Config) *ClientMap {
	return &ClientMap{
		ClientMap:			make(map[string]*Client),
		LimitRPS:			cfg.LimitRPS,
		LimitBurst:			cfg.LimitBurst,
		CleanupInterval:	cfg.ClientCleanupInterval,
		CleanupTimeout:		cfg.ClientCleanupInterval,
		Mutex:				&sync.Mutex{},
	}
}

func (cl ClientMap) GetLimiter(ip string) *rate.Limiter {
	if _, ok := cl.ClientMap[ip]; !ok {
		cl.ClientMap[ip] = &Client{
			Limiter: rate.NewLimiter(
				cl.LimitRPS,
				cl.LimitBurst,
			),
		}
	}
	cl.ClientMap[ip].LastSeen = time.Now()
	return cl.ClientMap[ip].Limiter
}

func (cl ClientMap) Cleanup() {
	for ip, client := range cl.ClientMap {
		if time.Since(client.LastSeen) > cl.CleanupTimeout {
			delete(cl.ClientMap, ip)
		}
	}
}

func (cl ClientMap) IsIpAllowed(ip string) bool {
	cl.Mutex.Lock()
	defer cl.Mutex.Unlock()
	limiter := cl.GetLimiter(ip)
	return limiter.Allow()
}

func (cl ClientMap) CleanCycle() {
	for {
		time.Sleep(cl.CleanupInterval)
		slog.Debug("Running client cleanup")
		cl.Mutex.Lock()
		for ip, client := range cl.ClientMap {
			if time.Since(client.LastSeen) > cl.CleanupTimeout {
				slog.Debug("Removing stale client", "ip", ip)
				delete(cl.ClientMap, ip)
			}
		}
		cl.Mutex.Unlock()
		slog.Debug("Finished client cleanup")
	}
}
