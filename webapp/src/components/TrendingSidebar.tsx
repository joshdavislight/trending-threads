import manifest from 'manifest';
import React, {useEffect, useState} from 'react';

import './TrendingSidebar.css';

interface ThreadResult {
    post_id: string;
    root_id: string;
    channel_id: string;
    channel_name: string;
    message: string;
    reply_count: number;
    reaction_count: number;
    score: number;
    last_reply_at: number;
}

interface Props {

    // Note: The exact props required may vary based on the registry method used.
    // Commonly available: theme, currentUserId, currentTeamId
    selectPost?: (postId: string) => void;
}

const TrendingSidebar: React.FC<Props> = ({selectPost}) => {
    const [threads, setThreads] = useState<ThreadResult[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

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
        // Initial fetch
        fetchTrendingThreads();

        // Set up periodic refresh
        // TODO: Read RefreshIntervalSeconds from plugin config if exposed to webapp
        const refreshInterval = 300 * 1000; // 5 minutes default
        const intervalId = setInterval(fetchTrendingThreads, refreshInterval);

        return () => clearInterval(intervalId);
    }, []);

    const handleThreadClick = (thread: ThreadResult) => {
        // Note: The exact method to open a thread in the native thread viewer may vary in v10.11.8.
        // Common approaches:
        // 1. Use selectPost action from props (if available)
        // 2. Dispatch a Redux action to open the thread panel
        // 3. Navigate to the post URL

        if (selectPost) {
            selectPost(thread.post_id);
        } else {
            // Fallback: Navigate to the post URL
            // This should open the post in the center channel and potentially the thread viewer
            window.location.href = `/${thread.channel_id}/pl/${thread.post_id}`;
        }
    };

    if (loading) {
        return (
            <div className='trending-threads-sidebar'>
                <div className='trending-threads-header'>
                    <span className='trending-icon'>{'🔥'}</span>
                    <span className='trending-title'>{'Trending'}</span>
                </div>
                <div className='trending-loading'>{'Loading...'}</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className='trending-threads-sidebar'>
                <div className='trending-threads-header'>
                    <span className='trending-icon'>{'🔥'}</span>
                    <span className='trending-title'>{'Trending'}</span>
                </div>
                <div className='trending-error'>{error}</div>
            </div>
        );
    }

    if (threads.length === 0) {
        return (
            <div className='trending-threads-sidebar'>
                <div className='trending-threads-header'>
                    <span className='trending-icon'>{'🔥'}</span>
                    <span className='trending-title'>{'Trending'}</span>
                </div>
                <div className='trending-empty'>{'No trending threads yet'}</div>
            </div>
        );
    }

    return (
        <div className='trending-threads-sidebar'>
            <div className='trending-threads-header'>
                <span className='trending-icon'>{'🔥'}</span>
                <span className='trending-title'>{'Trending'}</span>
            </div>
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
        </div>
    );
};

export default TrendingSidebar;
