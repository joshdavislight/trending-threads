package main

import (
	"time"
)

// startRefreshTicker starts the background ticker that refreshes trending threads.
func (p *Plugin) startRefreshTicker() {
	config := p.getConfiguration()
	interval := time.Duration(config.RefreshIntervalSeconds) * time.Second

	if interval <= 0 {
		interval = 300 * time.Second // Default to 5 minutes if misconfigured
	}

	p.refreshTicker = time.NewTicker(interval)
	p.refreshDone = make(chan bool)

	// Perform initial refresh
	go p.refreshTrendingThreads()

	// Start background refresh loop
	go func() {
		for {
			select {
			case <-p.refreshTicker.C:
				p.refreshTrendingThreads()
			case <-p.refreshDone:
				return
			}
		}
	}()

	p.API.LogInfo("Trending threads refresh ticker started", "interval_seconds", config.RefreshIntervalSeconds)
}

// stopRefreshTicker stops the background refresh ticker.
func (p *Plugin) stopRefreshTicker() {
	if p.refreshTicker != nil {
		p.refreshTicker.Stop()
		p.refreshDone <- true
		p.API.LogInfo("Trending threads refresh ticker stopped")
	}
}

// refreshTrendingThreads performs a single refresh of the trending threads cache.
func (p *Plugin) refreshTrendingThreads() {
	p.API.LogDebug("Refreshing trending threads")

	threads, err := p.ScoreThreads()
	if err != nil {
		p.API.LogError("Failed to score threads", "error", err.Error())
		return
	}

	// Update the cache
	p.trendingMutex.Lock()
	p.trendingThreads = threads
	p.trendingMutex.Unlock()

	p.API.LogDebug("Trending threads refreshed", "count", len(threads))
}

// getTrendingThreads returns the cached list of trending threads (thread-safe).
func (p *Plugin) getTrendingThreads() []ThreadResult {
	p.trendingMutex.RLock()
	defer p.trendingMutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]ThreadResult, len(p.trendingThreads))
	copy(result, p.trendingThreads)
	return result
}
