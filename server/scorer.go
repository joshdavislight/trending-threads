package main

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ThreadResult represents a single trending thread with its computed score.
type ThreadResult struct {
	PostID        string  `json:"post_id"`
	RootID        string  `json:"root_id"`
	ChannelID     string  `json:"channel_id"`
	ChannelName   string  `json:"channel_name"`
	Message       string  `json:"message"`
	ReplyCount    int     `json:"reply_count"`
	ReactionCount int     `json:"reaction_count"`
	Score         float64 `json:"score"`
	LastReplyAt   int64   `json:"last_reply_at"`
}

// ScoreThreads computes trending scores for threads across the configured scope.
func (p *Plugin) ScoreThreads() ([]ThreadResult, error) {
	config := p.getConfiguration()

	// Validate configuration
	if config.MaxThreads <= 0 {
		return []ThreadResult{}, nil
	}

	// Determine which channels to scan
	channelIDs, err := p.getChannelIDsToScan(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine channels to scan")
	}

	if len(channelIDs) == 0 {
		p.API.LogDebug("No channels to scan for trending threads")
		return []ThreadResult{}, nil
	}

	// Calculate cutoff time based on TimeWindowHours
	cutoffTime := time.Now().Add(-time.Duration(config.TimeWindowHours) * time.Hour).Unix() * 1000

	var allThreads []ThreadResult

	// Scan each channel for threads
	for _, channelID := range channelIDs {
		threads, err := p.getThreadsFromChannel(channelID, cutoffTime, config)
		if err != nil {
			p.API.LogWarn("Failed to get threads from channel", "channel_id", channelID, "error", err.Error())
			continue
		}
		allThreads = append(allThreads, threads...)
	}

	// Sort by score descending
	sort.Slice(allThreads, func(i, j int) bool {
		return allThreads[i].Score > allThreads[j].Score
	})

	// Return top N threads
	if len(allThreads) > config.MaxThreads {
		allThreads = allThreads[:config.MaxThreads]
	}

	return allThreads, nil
}

// getChannelIDsToScan returns the list of channel IDs to scan based on the configuration.
func (p *Plugin) getChannelIDsToScan(config *configuration) ([]string, error) {
	if config.Scope == "channels" {
		// Parse comma-separated channel IDs
		if config.ChannelIDs == "" {
			return []string{}, nil
		}
		channelIDs := strings.Split(config.ChannelIDs, ",")
		// Trim whitespace
		for i := range channelIDs {
			channelIDs[i] = strings.TrimSpace(channelIDs[i])
		}
		return channelIDs, nil
	}

	// Scope is "server" - get all public channels
	// ⚠️ Note: There is no single API call to get all public channels across all teams in v10.11.8.
	// We need to iterate through teams. For now, we'll use GetPublicChannelsForTeam for each team.
	// This may be inefficient for large servers. A future optimization could cache channel lists.

	teams, err := p.API.GetTeams()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get teams")
	}

	var channelIDs []string
	for _, team := range teams {
		channels, appErr := p.API.GetPublicChannelsForTeam(team.Id, 0, 200)
		if appErr != nil {
			p.API.LogWarn("Failed to get channels for team", "team_id", team.Id, "error", appErr.Error())
			continue
		}
		for _, channel := range channels {
			channelIDs = append(channelIDs, channel.Id)
		}
	}

	return channelIDs, nil
}

// getThreadsFromChannel fetches and scores threads from a single channel.
func (p *Plugin) getThreadsFromChannel(channelID string, cutoffTime int64, config *configuration) ([]ThreadResult, error) {
	// Fetch recent posts from the channel
	// ⚠️ Note: GetPostsForChannel returns posts in reverse chronological order (newest first).
	// We use page=0 and perPage=60 to get the most recent 60 posts as a starting point.
	postList, appErr := p.API.GetPostsForChannel(channelID, 0, 60)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get posts for channel")
	}

	// Get channel info for the name
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		p.API.LogWarn("Failed to get channel info", "channel_id", channelID, "error", appErr.Error())
	}
	channelName := ""
	if channel != nil {
		channelName = channel.DisplayName
		if channelName == "" {
			channelName = channel.Name
		}
	}

	var threads []ThreadResult

	// Process each post
	for _, postID := range postList.Order {
		post := postList.Posts[postID]

		// Only consider root posts (threads)
		if post.RootId != "" {
			continue
		}

		// Check if post has activity within the time window
		lastActivityTime := post.CreateAt
		if post.LastReplyAt > 0 {
			lastActivityTime = post.LastReplyAt
		}

		if lastActivityTime < cutoffTime {
			continue
		}

		// Get reaction count
		reactions, appErr := p.API.GetReactions(postID)
		reactionCount := 0
		if appErr == nil {
			reactionCount = len(reactions)
		}

		// Calculate score
		score := p.calculateScore(post.ReplyCount, int64(reactionCount), lastActivityTime, config)

		// Truncate message to 80 characters
		message := post.Message
		if len(message) > 80 {
			message = message[:77] + "..."
		}

		threads = append(threads, ThreadResult{
			PostID:        post.Id,
			RootID:        post.RootId,
			ChannelID:     post.ChannelId,
			ChannelName:   channelName,
			Message:       message,
			ReplyCount:    int(post.ReplyCount),
			ReactionCount: reactionCount,
			Score:         score,
			LastReplyAt:   post.LastReplyAt,
		})
	}

	return threads, nil
}

// calculateScore computes the trending score for a thread using the configured formula.
func (p *Plugin) calculateScore(replyCount int64, reactionCount int64, lastActivityTime int64, config *configuration) float64 {
	// Raw activity score: weighted sum of replies and reactions
	rawActivity := float64(replyCount)*config.ReplyWeight + float64(reactionCount)*config.ReactionWeight

	// Calculate hours since last activity
	hoursInactive := float64(time.Now().Unix()*1000-lastActivityTime) / (1000 * 60 * 60)

	// Recency multiplier: exponential decay
	recencyMultiplier := math.Exp(-config.DecayRate * hoursInactive)

	return rawActivity * recencyMultiplier
}
