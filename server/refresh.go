package main

import (
	"time"
)

// postTriggerDebounceDuration returns how long to wait after the last qualifying post
// before running a post-triggered trending refresh. Non-positive values use 2 seconds.
func postTriggerDebounceDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 2 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

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
	p.stopDebouncedTrendingRefresh()
	if p.refreshTicker != nil {
		p.refreshTicker.Stop()
		p.refreshDone <- true
		p.API.LogInfo("Trending threads refresh ticker stopped")
	}
}

// stopDebouncedTrendingRefresh cancels any pending post-triggered refresh.
func (p *Plugin) stopDebouncedTrendingRefresh() {
	p.refreshDebounceMu.Lock()
	defer p.refreshDebounceMu.Unlock()
	if p.refreshDebounceTimer != nil {
		p.refreshDebounceTimer.Stop()
		p.refreshDebounceTimer = nil
	}
}

// scheduleDebouncedTrendingRefresh schedules a trending rescore shortly after posting activity.
func (p *Plugin) scheduleDebouncedTrendingRefresh() {
	debounce := postTriggerDebounceDuration(p.getConfiguration().PostTriggerDebounceSeconds)

	p.refreshDebounceMu.Lock()
	defer p.refreshDebounceMu.Unlock()
	if p.refreshDebounceTimer != nil {
		p.refreshDebounceTimer.Stop()
	}
	p.refreshDebounceTimer = time.AfterFunc(debounce, func() {
		p.refreshTrendingThreads()
		p.refreshDebounceMu.Lock()
		p.refreshDebounceTimer = nil
		p.refreshDebounceMu.Unlock()
	})
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
