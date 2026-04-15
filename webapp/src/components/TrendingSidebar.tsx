import manifest from 'manifest';
import React, {useCallback, useEffect, useState} from 'react';

import './TrendingSidebar.css';

interface ThreadResult {
    post_id: string;
    root_id: string;
    channel_id: string;
    channel_name: string;
    team_name: string;
    message: string;
    reply_count: number;
    reaction_count: number;
    score: number;
    last_reply_at: number;
}

interface Props {

    // Note: The exact props required may vary based on the registry method used.
    // Commonly available: theme, currentUserId, currentTeamId
    onOpenThread?: (postId: string) => void | Promise<unknown>;
    selectPost?: (postId: string) => void;
}

const panelHiddenStorageKey = `${manifest.id}_trending_panel_hidden`;

function readPanelHidden(): boolean {
    if (typeof window === 'undefined' || !window.localStorage) {
        return false;
    }
    return window.localStorage.getItem(panelHiddenStorageKey) === '1';
}

const TrendingSidebar: React.FC<Props> = ({onOpenThread, selectPost}) => {
    const [threads, setThreads] = useState<ThreadResult[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [panelHidden, setPanelHidden] = useState<boolean>(readPanelHidden);

    const setPanelHiddenPersisted = useCallback((hidden: boolean) => {
        setPanelHidden(hidden);
        if (typeof window !== 'undefined' && window.localStorage) {
            if (hidden) {
                window.localStorage.setItem(panelHiddenStorageKey, '1');
            } else {
                window.localStorage.removeItem(panelHiddenStorageKey);
            }
        }
    }, []);

    const fetchTrendingThreads = async () => {
        try {
            const response = await fetch(`/plugins/${manifest.id}/api/v1/trending`, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`Failed to fetch trending threads: ${response.statusText}`);
            }

            const data: ThreadResult[] = await response.json();
            setThreads(data);
            setError(null);
        } catch (err) {
            setError('Failed to load trending threads');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTrendingThreads();

        // TODO: Read RefreshIntervalSeconds from plugin config if exposed to webapp
        const refreshInterval = 300 * 1000; // 5 minutes default
        const intervalId = setInterval(fetchTrendingThreads, refreshInterval);

        return () => clearInterval(intervalId);
    }, []);

    const handleThreadClick = async (thread: ThreadResult) => {
        if (onOpenThread) {
            try {
                const result = await onOpenThread(thread.post_id);
                if (result && typeof result === 'object' && 'data' in result && result.data === true) {
                    return;
                }
            } catch {
                // Fall through to legacy open paths.
            }
        }
        if (selectPost) {
            selectPost(thread.post_id);
        } else if (thread.team_name) {
            window.location.href = `/${thread.team_name}/pl/${thread.post_id}`;
        }
    };

    if (panelHidden) {
        return (
            <div className='trending-threads-sidebar trending-threads-sidebar--collapsed'>
                <button
                    type='button'
                    className='trending-threads-collapsed-bar'
                    onClick={() => setPanelHiddenPersisted(false)}
                    aria-expanded={false}
                    title='Show trending threads'
                >
                    <span
                        className='trending-icon'
                        aria-hidden={true}
                    >{'🔥'}</span>
                    <span className='trending-collapsed-label'>{'Trending'}</span>
                    <span
                        className='trending-toggle-chevron'
                        aria-hidden={true}
                    >{'▸'}</span>
                </button>
            </div>
        );
    }

    return (
        <div className='trending-threads-sidebar'>
            <div className='trending-threads-header'>
                <span
                    className='trending-icon'
                    aria-hidden={true}
                >{'🔥'}</span>
                <span className='trending-title'>{'Trending'}</span>
                <button
                    type='button'
                    className='trending-threads-toggle'
                    onClick={() => setPanelHiddenPersisted(true)}
                    aria-expanded={true}
                    title='Hide trending threads'
                >
                    <span
                        className='trending-toggle-chevron'
                        aria-hidden={true}
                    >{'▾'}</span>
                </button>
            </div>
            {loading && (
                <div className='trending-loading'>{'Loading...'}</div>
            )}
            {!loading && error && (
                <div className='trending-error'>{error}</div>
            )}
            {!loading && !error && threads.length === 0 && (
                <div className='trending-empty'>{'No trending threads yet'}</div>
            )}
            {!loading && !error && threads.length > 0 && (
                <div className='trending-threads-list'>
                    {threads.map((thread) => (
                        <div
                            key={thread.post_id}
                            className='trending-thread-item'
                            onClick={() => handleThreadClick(thread)}
                            role='button'
                            tabIndex={0}
                            onKeyPress={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    handleThreadClick(thread);
                                }
                            }}
                        >
                            <div className='thread-message'>{thread.message}</div>
                            <div className='thread-meta'>
                                <span className='thread-channel'>{thread.channel_name}</span>
                                <span className='thread-stats'>
                                    {thread.reply_count > 0 && (
                                        <span className='stat-item'>{'💬 '}{thread.reply_count}</span>
                                    )}
                                    {thread.reaction_count > 0 && (
                                        <span className='stat-item'>{'❤️ '}{thread.reaction_count}</span>
                                    )}
                                </span>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default TrendingSidebar;
