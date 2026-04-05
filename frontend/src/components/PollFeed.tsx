import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { pollService } from '../services/api';
import { Poll } from '../types';

const PollFeed: React.FC = () => {
  const navigate = useNavigate();
  const [polls, setPolls] = useState<Poll[]>([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadPolls();
  }, []);

  const loadPolls = async () => {
    try {
      setLoading(true);
      const data = await pollService.getPublicPolls();
      setPolls(data);
    } catch {
      setError('Failed to load polls');
    } finally {
      setLoading(false);
    }
  };

  const filteredPolls = polls.filter(p =>
    p.title.toLowerCase().includes(search.toLowerCase())
  );

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return <div className="container">Loading polls...</div>;
  }

  return (
    <div className="container">
      <div className="feed-header">
        <h1>Poll Feed</h1>
        <button onClick={() => navigate('/create')} className="btn-primary">
          Create Poll
        </button>
      </div>

      <div className="search-bar">
        <input
          type="text"
          placeholder="🔍 Search polls by title..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {error && <div className="error">{error}</div>}

      {filteredPolls.length === 0 && polls.length > 0 ? (
        <div className="empty-state">
          <p>No polls matching "{search}"</p>
        </div>
      ) : polls.length === 0 ? (
        <div className="empty-state">
          <p>No polls yet. Be the first to create one!</p>
          <button onClick={() => navigate('/create')} className="btn-primary">
            Create Poll
          </button>
        </div>
      ) : (
        <div className="poll-list">
          {filteredPolls.map(poll => (
            <div
              key={poll.id}
              className="poll-card"
              onClick={() => navigate(`/poll/${poll.id}`)}
            >
              <h3>{poll.title}</h3>
              <p className="poll-meta">
                Created {formatDate(poll.created_at)}
              </p>
              <p className="poll-visibility">
                {poll.visibility === 'unlisted' && '🔗 Unlisted · '}
                {poll.result_visibility === 'always' && '📊 Public results'}
                {poll.result_visibility === 'after_vote' && '🗳️ Results after voting'}
                {poll.result_visibility === 'author_only' && '🔒 Author-only results'}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default PollFeed;
