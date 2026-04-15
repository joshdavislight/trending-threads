// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Post} from '@mattermost/types/posts';

// Shape must stay aligned with Mattermost webapp ActionTypes.SELECT_POST / selectPost(post).
export function rhsSelectPostAction(post: Post, previousRhsState: unknown | undefined, timestamp: number) {
    return {
        type: 'SELECT_POST' as const,
        postId: post.root_id || post.id,
        channelId: post.channel_id,
        previousRhsState,
        timestamp,
    };
}
