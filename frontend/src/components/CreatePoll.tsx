import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { pollService } from '../services/api';
import { CreatePollRequest } from '../types';

const CreatePoll: React.FC = () => {
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [visibility, setVisibility] = useState<'public' | 'unlisted'>('public');
  const [resultVisibility, setResultVisibility] = useState<'always' | 'after_vote' | 'author_only'>('always');
  const [questions, setQuestions] = useState([{ text: '', options: ['', ''] }]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const addQuestion = () => {
    setQuestions([...questions, { text: '', options: ['', ''] }]);
  };

  const addOption = (questionIndex: number) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex].options.push('');
    setQuestions(newQuestions);
  };

  const updateQuestion = (questionIndex: number, text: string) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex].text = text;
    setQuestions(newQuestions);
  };

  const updateOption = (questionIndex: number, optionIndex: number, text: string) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex].options[optionIndex] = text;
    setQuestions(newQuestions);
  };

  const removeQuestion = (questionIndex: number) => {
    if (questions.length <= 1) return;
    const newQuestions = questions.filter((_, idx) => idx !== questionIndex);
    setQuestions(newQuestions);
  };

  const removeOption = (questionIndex: number, optionIndex: number) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex].options = newQuestions[questionIndex].options.filter((_, idx) => idx !== optionIndex);
    if (newQuestions[questionIndex].options.length === 0) {
      newQuestions[questionIndex].options = ['', ''];
    }
    setQuestions(newQuestions);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const filteredQuestions = questions
      .filter(q => q.text.trim() && q.options.some(o => o.trim()))
      .map(q => ({
        text: q.text.trim(),
        options: q.options.filter(o => o.trim()),
      }))
      .filter(q => q.options.length >= 2);

    if (!title.trim()) {
      setError('Title is required');
      return;
    }

    if (filteredQuestions.length === 0) {
      setError('At least one question with 2+ options is required');
      return;
    }

    setLoading(true);

    try {
      const data: CreatePollRequest = {
        title: title.trim(),
        visibility: visibility,
        result_visibility: resultVisibility,
        questions: filteredQuestions,
      };

      const poll = await pollService.createPoll(data);
      navigate(`/poll/${poll.id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create poll');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <h1>Create Poll</h1>
      {error && <div className="error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="title">Poll Title</label>
          <input
            id="title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Enter poll title"
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="visibility">Poll Visibility</label>
          <select
            id="visibility"
            value={visibility}
            onChange={(e) => setVisibility(e.target.value as any)}
          >
            <option value="public">🌐 Public — visible in feed to everyone</option>
            <option value="unlisted">🔗 Unlisted — only visible to me in feed, accessible by link</option>
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="resultVisibility">Results Visibility</label>
          <select
            id="resultVisibility"
            value={resultVisibility}
            onChange={(e) => setResultVisibility(e.target.value as any)}
          >
            <option value="always">📊 Everyone can see results immediately</option>
            <option value="after_vote">🗳️ Everyone can see results after voting</option>
            <option value="author_only">🔒 Only I can see results</option>
          </select>
        </div>

        {questions.map((question, questionIndex) => (
          <div key={questionIndex} className="question-block">
            <div className="question-header">
              <h3>Question {questionIndex + 1}</h3>
              {questions.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeQuestion(questionIndex)}
                  className="btn-danger"
                >
                  ✕
                </button>
              )}
            </div>
            <input
              type="text"
              value={question.text}
              onChange={(e) => updateQuestion(questionIndex, e.target.value)}
              placeholder="Enter question text"
              required
            />
            {question.options.map((option, optionIndex) => (
              <div key={optionIndex} className="option-row">
                <input
                  type="text"
                  value={option}
                  onChange={(e) => updateOption(questionIndex, optionIndex, e.target.value)}
                  placeholder={`Option ${optionIndex + 1}`}
                />
                {question.options.length > 2 && (
                  <button
                    type="button"
                    onClick={() => removeOption(questionIndex, optionIndex)}
                    className="btn-danger-small"
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addOption(questionIndex)}
              className="btn-secondary"
            >
              + Add Option
            </button>
          </div>
        ))}

        <button type="button" onClick={addQuestion} className="btn-secondary">
          + Add Question
        </button>

        <div className="form-actions">
          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Creating...' : 'Create Poll'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreatePoll;
