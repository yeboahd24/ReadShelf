export interface User {
  id: string;
  email: string;
  plan: string;
  created_at: string;
}

export interface Book {
  id: string;
  user_id: string;
  title: string;
  author?: string;
  page_count?: number;
  color?: string;
  created_at: string;
}

export interface Annotation {
  id: string;
  book_id: string;
  user_id: string;
  type: 'highlight' | 'strikethrough';
  content: string;
  page: number;
  chapter?: string;
  user_note?: string;
  char_start?: number;
  char_end?: number;
  rects?: Array<{ x: number; y: number; width: number; height: number }>;
  created_at: string;
  book_title?: string;
}

export interface RecallResult {
  answer: string;
  sources: Array<{
    annotation_id: string;
    book_id: string;
    book_title: string;
    page: number;
    content: string;
  }>;
}

export interface AuthResponse {
  user: User;
  access_token: string;
}
