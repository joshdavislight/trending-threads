// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Post} from '@mattermost/types/posts';
import type {GlobalState} from '@mattermost/types/store';

import {getChannel as fetchChannel} from 'mattermost-redux/actions/channels';
import {getPost as fetchPost} from 'mattermost-redux/actions/posts';
import {getChannel as getChannelFromState} from 'mattermost-redux/selectors/entities/channels';
import {getPost as getPostFromState} from 'mattermost-redux/selectors/entities/posts';

import {rhsSelectPostAction} from './rhs_select_post';

type StateWithRhs = GlobalState & {
    views?: {
        rhs?: {
            rhsState?: unknown;
        };
    };
};

type OpenThreadResult = {data: boolean};

export function openTrendingThread(postId: string) {
    return async (dispatch: (action: unknown) => unknown | Promise<unknown>, getState: () => GlobalState): Promise<OpenThreadResult> => {
        let post = getPostFromState(getState(), postId) as Post | undefined;

        if (!post) {
            const res = (await dispatch(fetchPost(postId, false, false))) as {data?: Post; error?: unknown};
            if (res.error) {
                return {data: false};
            }
            post = res.data;
        }

        if (!post || post.delete_at !== 0) {
            return {data: false};
        }

        if (!getChannelFromState(getState(), post.channel_id)) {
            const chRes = (await dispatch(fetchChannel(post.channel_id))) as {data?: unknown; error?: unknown};
            if (chRes.error) {
                return {data: false};
            }
        }

        const st = getState() as StateWithRhs;
        const previousRhsState = st.views?.rhs?.rhsState;

        dispatch(rhsSelectPostAction(post, previousRhsState, Date.now()));

        return {data: true};
    };
}
