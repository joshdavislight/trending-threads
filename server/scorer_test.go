package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetThreadsFromChannel_IncludesTeamName(t *testing.T) {
	api := &plugintest.API{}
	p := Plugin{}
	p.SetAPI(api)

	chID := "channel1"
	teamID := "team1"
	postID := "postroot1"
	now := time.Now().UnixMilli()

	post := &model.Post{
		Id:          postID,
		ChannelId:   chID,
		RootId:      "",
		CreateAt:    now,
		Message:     "hello thread",
		ReplyCount:  1,
		LastReplyAt: now,
	}
	postList := model.NewPostList()
	postList.AddOrder(postID)
	postList.AddPost(post)

	cutoff := time.Now().Add(-25 * time.Hour).UnixMilli()

	api.On("GetPostsSince", chID, cutoff).Return(postList, (*model.AppError)(nil))
	api.On("GetPost", postID).Return(post, (*model.AppError)(nil))
	api.On("GetChannel", chID).Return(&model.Channel{
		Id:          chID,
		TeamId:      teamID,
		Name:        "town-square",
		DisplayName: "Town Square",
	}, (*model.AppError)(nil))
	api.On("GetTeam", teamID).Return(&model.Team{
		Id:   teamID,
		Name: "acme-corp",
	}, (*model.AppError)(nil))
	api.On("GetReactions", postID).Return([]*model.Reaction{}, (*model.AppError)(nil))

	cfg := &configuration{
		TimeWindowHours: 24,
		ReplyWeight:     2,
		ReactionWeight:  1,
		DecayRate:       0.1,
	}

	threads, err := p.getThreadsFromChannel(chID, cutoff, cfg)
	require.NoError(t, err)
	require.Len(t, threads, 1)
	assert.Equal(t, "acme-corp", threads[0].TeamName)
	assert.Equal(t, postID, threads[0].PostID)
	assert.Equal(t, chID, threads[0].ChannelID)

	api.AssertExpectations(t)
}

func TestGetThreadsFromChannel_DiscoversThreadViaReplyInWindow(t *testing.T) {
	api := &plugintest.API{}
	p := Plugin{}
	p.SetAPI(api)

	chID := "channel1"
	teamID := "team1"
	rootID := "root1"
	replyID := "reply1"

	base := time.Now().UTC()
	now := base.UnixMilli()
	cutoff := base.Add(-25 * time.Hour).UnixMilli()
	rootCreateAt := base.Add(-720 * time.Hour).UnixMilli()

	reply := &model.Post{
		Id:        replyID,
		ChannelId: chID,
		RootId:    rootID,
		CreateAt:  now,
		Message:   "new reply",
	}
	root := &model.Post{
		Id:          rootID,
		ChannelId:   chID,
		RootId:      "",
		CreateAt:    rootCreateAt,
		Message:     "old thread root",
		ReplyCount:  8,
		LastReplyAt: now,
	}

	postList := model.NewPostList()
	postList.AddOrder(replyID)
	postList.AddPost(reply)

	api.On("GetPostsSince", chID, cutoff).Return(postList, (*model.AppError)(nil))
	api.On("GetPost", rootID).Return(root, (*model.AppError)(nil))
	api.On("GetChannel", chID).Return(&model.Channel{
		Id:          chID,
		TeamId:      teamID,
		Name:        "town-square",
		DisplayName: "Town Square",
	}, (*model.AppError)(nil))
	api.On("GetTeam", teamID).Return(&model.Team{
		Id:   teamID,
		Name: "acme-corp",
	}, (*model.AppError)(nil))
	api.On("GetReactions", rootID).Return([]*model.Reaction{}, (*model.AppError)(nil))

	cfg := &configuration{
		TimeWindowHours: 24,
		ReplyWeight:     2,
		ReactionWeight:  1,
		DecayRate:       0.1,
	}

	threads, err := p.getThreadsFromChannel(chID, cutoff, cfg)
	require.NoError(t, err)
	require.Len(t, threads, 1)
	assert.Equal(t, rootID, threads[0].PostID)
	assert.Equal(t, 8, threads[0].ReplyCount)
	assert.Equal(t, chID, threads[0].ChannelID)

	api.AssertExpectations(t)
}

// When the root post returned by GetPost has LastReplyAt unset (stale), we must still
// score the thread if GetPostsSince returned recent replies in the time window.
func TestGetThreadsFromChannel_IncludesThreadWhenRootLastReplyAtStale(t *testing.T) {
	api := &plugintest.API{}
	p := Plugin{}
	p.SetAPI(api)

	chID := "channel1"
	teamID := "team1"
	rootID := "root1"
	replyID := "reply1"

	base := time.Now().UTC()
	now := base.UnixMilli()
	cutoff := base.Add(-25 * time.Hour).UnixMilli()
	rootCreateAt := base.Add(-720 * time.Hour).UnixMilli()

	reply := &model.Post{
		Id:        replyID,
		ChannelId: chID,
		RootId:    rootID,
		CreateAt:  now,
		UpdateAt:  now,
		Message:   "new reply",
	}
	root := &model.Post{
		Id:          rootID,
		ChannelId:   chID,
		RootId:      "",
		CreateAt:    rootCreateAt,
		UpdateAt:    rootCreateAt,
		Message:     "old thread root",
		ReplyCount:  8,
		LastReplyAt: 0,
	}

	postList := model.NewPostList()
	postList.AddOrder(replyID)
	postList.AddPost(reply)

	api.On("GetPostsSince", chID, cutoff).Return(postList, (*model.AppError)(nil))
	api.On("GetPost", rootID).Return(root, (*model.AppError)(nil))
	api.On("GetChannel", chID).Return(&model.Channel{
		Id:          chID,
		TeamId:      teamID,
		Name:        "town-square",
		DisplayName: "Town Square",
	}, (*model.AppError)(nil))
	api.On("GetTeam", teamID).Return(&model.Team{
		Id:   teamID,
		Name: "acme-corp",
	}, (*model.AppError)(nil))
	api.On("GetReactions", rootID).Return([]*model.Reaction{}, (*model.AppError)(nil))

	cfg := &configuration{
		TimeWindowHours: 24,
		ReplyWeight:     2,
		ReactionWeight:  1,
		DecayRate:       0.1,
	}

	threads, err := p.getThreadsFromChannel(chID, cutoff, cfg)
	require.NoError(t, err)
	require.Len(t, threads, 1)
	assert.Equal(t, rootID, threads[0].PostID)
	assert.Equal(t, 8, threads[0].ReplyCount)
	assert.Equal(t, now, threads[0].LastReplyAt)

	api.AssertExpectations(t)
}
