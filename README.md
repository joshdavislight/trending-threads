# Mattermost Trending Threads Plugin

A Mattermost plugin that surfaces the most active threads in a persistent sidebar section, helping teams discover ongoing conversations.

![Trending Threads](https://img.shields.io/badge/mattermost-10.11.8+-blue.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)

## Features

- **🔥 Trending Sidebar Section**: Always-visible section showing top N most active threads
- **Smart Scoring Algorithm**: Exponential recency decay with configurable weights for replies and reactions
- **Flexible Scope**: Monitor all public channels or specific channels only
- **Configurable Refresh**: Automatic background updates on a configurable interval
- **One-Click Thread Access**: Opens threads natively in Mattermost's thread viewer
- **Theme Integration**: Matches your Mattermost theme automatically

## How It Works

The plugin calculates a trending score for each thread based on:

```
score = (replyCount × replyWeight + reactionCount × reactionWeight) × recencyMultiplier
```

Where:
- **Reply Weight** (default 2.0): Values discussion-heavy threads
- **Reaction Weight** (default 1.0): Values quick engagement
- **Recency Multiplier**: Exponential decay (e^(-decayRate × hoursInactive))
- **Decay Rate** (default 0.1): Controls recency bias (higher = favor recent activity)

## Installation

### Requirements

- Mattermost Server 10.11.8+
- Go 1.25+ (for building from source)
- Node.js 16+ and npm (for building from source)

### From Release

1. Download the latest `.tar.gz` from [Releases](../../releases)
2. Go to **System Console → Plugin Management**
3. Upload the `.tar.gz` file
4. Enable the plugin

### From Source

```bash
git clone https://github.com/YOUR_USERNAME/mattermost-plugin-trending-threads.git
cd mattermost-plugin-trending-threads
make dist

# Upload dist/com.mattermost.trending-threads-*.tar.gz to your Mattermost server
```

## Configuration

All settings are available in **System Console → Plugins → Trending Threads** or in `config.json`:

### Basic Settings

| Setting | Default | Description |
|---------|---------|-------------|
| **Scope** | `server` | `server` = all public channels, `channels` = specific channels only |
| **Channel IDs** | (empty) | Comma-separated channel IDs when Scope = `channels` |
| **Time Window (hours)** | `24` | How far back to look for activity |
| **Max Threads** | `3` | Number of threads to display in sidebar |
| **Refresh Interval (seconds)** | `300` | How often to recalculate trending threads |

### Advanced Settings (Scoring Weights)

| Setting | Default | Description |
|---------|---------|-------------|
| **Reply Weight** | `2.0` | Weight factor for replies in score calculation |
| **Reaction Weight** | `1.0` | Weight factor for reactions in score calculation |
| **Decay Rate** | `0.1` | Exponential decay rate for recency weighting |

## Performance Considerations

### Server Size Guidelines

| Server Size | Recommended Scope | Refresh Interval | Notes |
|-------------|-------------------|------------------|-------|
| Small (<20 channels) | `server` | 300s | Minimal impact |
| Medium (20-100 channels) | `channels` (10-20) | 300-600s | Monitor initially |
| Large (>100 channels) | `channels` (5-10) | 600s | Start conservatively |
| Enterprise (>500 channels) | `channels` (3-5) | 600-900s | Test on staging first |

**Recommendation for large servers:** Start with `channels` scope and monitor performance before expanding.

## Troubleshooting

### Sidebar doesn't appear

- Verify plugin is enabled in System Console
- Check browser console (F12) for JavaScript errors
- Try refreshing the page
- Verify Mattermost version is 10.11.8+

### No threads appear

- Check if threads exist with replies/reactions in the configured time window
- Increase `TimeWindowHours` to 48 or 72
- Check server logs: `tail -f /opt/mattermost/logs/mattermost.log | grep trending`

## Development

### Build

```bash
make
```

### Lint

```bash
make check-style
```

### Test

```bash
make test
```

## Architecture

### Backend (Go)

- **`plugin.go`**: Plugin lifecycle and initialization
- **`scorer.go`**: Thread scoring algorithm
- **`refresh.go`**: Background refresh ticker and cache management
- **`api.go`**: REST API endpoint
- **`configuration.go`**: Configuration loading

### Frontend (React/TypeScript)

- **`index.tsx`**: Plugin registration
- **`components/TrendingSidebar.tsx`**: Main sidebar component
- **`components/TrendingSidebar.css`**: Styling

## License

This project is licensed under the Apache License 2.0.

## Acknowledgments

Built on the [Mattermost Plugin Starter Template](https://github.com/mattermost/mattermost-plugin-starter-template)
