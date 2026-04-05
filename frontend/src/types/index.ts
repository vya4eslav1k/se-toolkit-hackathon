export interface Poll {
  id: string;
  title: string;
  created_at: string;
  created_by: string;
  visibility: 'public' | 'unlisted';
  result_visibility: 'always' | 'after_vote' | 'author_only';
  questions?: Question[];
}

export interface Question {
  id: string;
  poll_id: string;
  text: string;
  options: Option[];
}

export interface Option {
  id: string;
  question_id: string;
  text: string;
  votes?: number;
}

export interface CreatePollRequest {
  title: string;
  visibility: 'public' | 'unlisted';
  result_visibility: 'always' | 'after_vote' | 'author_only';
  questions: QuestionRequest[];
}

export interface QuestionRequest {
  text: string;
  options: string[];
}

export interface VoteRequest {
  answers: AnswerRequest[];
}

export interface AnswerRequest {
  question_id: string;
  option_id: string;
}

export interface PollResult {
  poll: Poll;
  total_votes: number;
  can_view: boolean;
}
