// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Post} from '@mattermost/types/posts';

import {rhsSelectPostAction} from '../src/actions/rhs_select_post';

describe('rhsSelectPostAction', () => {
    it('uses thread root id for replies and passes channel and rhs state', () => {
        const reply = {
            id: 'reply-1',
            root_id: 'root-1',
            channel_id: 'channel-1',
        } as Post;

        const action = rhsSelectPostAction(reply, 'some_rhs', 42_000);

        expect(action).toEqual({
            type: 'SELECT_POST',
            postId: 'root-1',
            channelId: 'channel-1',
            previousRhsState: 'some_rhs',
            timestamp: 42_000,
        });
    });

    it('uses post id for root messages', () => {
        const root = {
            id: 'root-1',
            root_id: '',
            channel_id: 'channel-1',
        } as Post;

        const action = rhsSelectPostAction(root, undefined, 1);

        expect(action.postId).toBe('root-1');
        expect(action.previousRhsState).toBeUndefined();
    });
});
