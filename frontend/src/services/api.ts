import axios from 'axios';
import { Poll, CreatePollRequest, VoteRequest, PollResult } from '../types';

const API_URL = process.env.REACT_APP_API_URL || '/api';

const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
});

export const pollService = {
  async getPublicPolls(): Promise<Poll[]> {
    const response = await api.get<Poll[]>('/polls');
    return response.data;
  },

  async createPoll(data: CreatePollRequest): Promise<Poll> {
    const response = await api.post<Poll>('/polls', data);
    return response.data;
  },

  async getPoll(id: string): Promise<{ poll: Poll; vote_status: Record<string, boolean> }> {
    const response = await api.get<{ poll: Poll; vote_status: Record<string, boolean> }>(`/polls/${id}`);
    return response.data;
  },

  async submitVote(pollId: string, data: VoteRequest): Promise<void> {
    await api.post(`/polls/${pollId}/vote`, data);
  },

  async getResults(pollId: string): Promise<PollResult> {
    const response = await api.get<PollResult>(`/polls/${pollId}/results`);
    return response.data;
  },
};
