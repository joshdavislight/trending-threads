package main

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// MessageHasBeenPosted is invoked after the message has been committed to the database.
func (p *Plugin) MessageHasBeenPosted(_ *plugin.Context, post *model.Post) {
	if post == nil || post.DeleteAt != 0 {
		return
	}
	if post.Type == model.PostTypeEphemeral {
		return
	}
	if strings.HasPrefix(post.Type, model.PostSystemMessagePrefix) {
		return
	}
	p.scheduleDebouncedTrendingRefresh()
}
