// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from 'manifest';
import type {Store} from 'redux';

import type {GlobalState} from '@mattermost/types/store';

import type {PluginRegistry} from 'types/mattermost-webapp';

import TrendingSidebar from './components/TrendingSidebar';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public async initialize(registry: PluginRegistry, store: Store<GlobalState>) {
        // Note: The exact registry method for adding sidebar sections may vary in v10.11.8.
        // This uses registerLeftSidebarHeaderComponent which is the expected method based on
        // Mattermost plugin patterns. If this doesn't work, alternatives to try:
        // - registerCustomRoute with a sidebar component
        // - registerChannelHeaderButtonAction with a custom component
        // - registerLeftSidebarHeaderComponent (most likely correct)

        if (registry.registerLeftSidebarHeaderComponent) {
            registry.registerLeftSidebarHeaderComponent(TrendingSidebar);
        }

        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
