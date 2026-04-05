import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { pollService } from '../services/api';
import { Poll, VoteRequest } from '../types';

const ViewPoll: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [poll, setPoll] = useState<Poll | null>(null);
  const [voteStatus, setVoteStatus] = useState<Record<string, boolean>>({});
  const [selectedAnswers, setSelectedAnswers] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (id) {
      loadPoll();
    }
  }, [id]);

  const loadPoll = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const data = await pollService.getPoll(id);
      setPoll(data.poll);
      setVoteStatus(data.vote_status);
    } catch {
      setError('Failed to load poll');
    } finally {
      setLoading(false);
    }
  };

  const handleOptionSelect = (questionId: string, optionId: string) => {
    setSelectedAnswers(prev => ({
      ...prev,
      [questionId]: optionId,
    }));
  };

  const handleSubmitVote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !poll || !poll.questions) return;

    const answers = poll.questions
      .filter(q => selectedAnswers[q.id])
      .map(q => ({
        question_id: q.id,
        option_id: selectedAnswers[q.id],
      }));

    if (answers.length === 0) {
      setError('Please select at least one answer');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      const voteRequest: VoteRequest = { answers };
      await pollService.submitVote(id, voteRequest);
      navigate(`/results/${id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to submit vote');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <div className="container">Loading...</div>;
  }

  if (!poll) {
    return <div className="container">Poll not found</div>;
  }

  const allAnswered = poll.questions && poll.questions.every(q => voteStatus[q.id]);

  return (
    <div className="container">
      <h1>{poll.title}</h1>
      {error && <div className="error">{error}</div>}
      
      {allAnswered ? (
        <div className="info">
          <p>You have already voted for this poll.</p>
          <button onClick={() => navigate(`/results/${id}`)} className="btn-primary">
            View Results
          </button>
        </div>
      ) : (
        <form onSubmit={handleSubmitVote}>
          {poll.questions && poll.questions.map((question, index) => (
            <div key={question.id} className="question-block">
              <h3>{index + 1}. {question.text}</h3>
              {voteStatus[question.id] ? (
                <p className="info">Already answered</p>
              ) : (
                <div className="options">
                  {question.options.map(option => (
                    <label key={option.id} className="option-item">
                      <input
                        type="radio"
                        name={`question-${question.id}`}
                        value={option.id}
                        checked={selectedAnswers[question.id] === option.id}
                        onChange={() => handleOptionSelect(question.id, option.id)}
                      />
                      <span className="option-text">{option.text}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>
          ))}
          <button
            type="submit"
            disabled={submitting || Object.keys(selectedAnswers).length === 0}
            className="btn-primary"
          >
            {submitting ? 'Submitting...' : 'Submit Vote'}
          </button>
        </form>
      )}
    </div>
  );
};

export default ViewPoll;
