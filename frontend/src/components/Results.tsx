import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { pollService } from '../services/api';
import { PollResult } from '../types';

const Results: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [result, setResult] = useState<PollResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (id) {
      loadResults();
    }
  }, [id]);

  const loadResults = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const data = await pollService.getResults(id);
      setResult(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load results');
    } finally {
      setLoading(false);
    }
  };

  const getPercentage = (votes: number, total: number): number => {
    if (total === 0) return 0;
    return Math.round((votes / total) * 100);
  };

  if (loading) {
    return <div className="container">Loading...</div>;
  }

  if (!result) {
    return <div className="container">{error || 'No results available'}</div>;
  }

  return (
    <div className="container">
      <h1>{result.poll.title} - Results</h1>
      {error && <div className="error">{error}</div>}
      
      <div className="results-summary">
        <p><strong>Total Votes:</strong> {result.total_votes}</p>
      </div>

      {result.poll.questions && result.poll.questions.map((question, qIndex) => (
        <div key={question.id} className="question-block">
          <h3>{qIndex + 1}. {question.text}</h3>
          <div className="results">
            {question.options.map(option => {
              const percentage = getPercentage(option.votes || 0, result.total_votes);
              return (
                <div key={option.id} className="result-item">
                  <div className="result-label">
                    <span>{option.text}</span>
                    <span>{option.votes} votes ({percentage}%)</span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress-fill"
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ))}

      <div className="form-actions">
        <button onClick={() => navigate('/')} className="btn-secondary">
          Back to Feed
        </button>
        <button onClick={() => navigate(`/poll/${id}`)} className="btn-secondary">
          Back to Poll
        </button>
      </div>
    </div>
  );
};

export default Results;
