import client from './client';
import type { Annotation } from '../types';

export async function listAnnotations(bookId: string): Promise<Annotation[]> {
  const res = await client.get(`/books/${bookId}/annotations`);
  return res.data || [];
}

export interface CreateAnnotationInput {
  type: 'highlight' | 'strikethrough';
  content: string;
  page: number;
  chapter?: string;
  user_note?: string;
  char_start?: number;
  char_end?: number;
  rects?: Array<{ x: number; y: number; width: number; height: number }>;
}

export async function createAnnotation(bookId: string, data: CreateAnnotationInput): Promise<Annotation> {
  const res = await client.post(`/books/${bookId}/annotations`, data);
  return res.data;
}

export async function deleteAnnotation(id: string): Promise<void> {
  await client.delete(`/annotations/${id}`);
}

export async function updateNote(id: string, note: string): Promise<void> {
  await client.patch(`/annotations/${id}/note`, { note });
}

export async function searchAnnotations(query: string): Promise<Annotation[]> {
  const res = await client.get('/annotations/search', { params: { q: query } });
  return res.data || [];
}
