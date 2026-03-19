import client from './client';
import type { Book } from '../types';

export async function listBooks(): Promise<Book[]> {
  const res = await client.get('/books');
  return res.data || [];
}

export async function getBook(id: string): Promise<Book> {
  const res = await client.get(`/books/${id}`);
  return res.data;
}

export async function getSignedURL(id: string): Promise<string> {
  const res = await client.get(`/books/${id}/url`);
  return res.data.url;
}

export async function uploadBook(file: File, title?: string, author?: string): Promise<Book> {
  const form = new FormData();
  form.append('file', file);
  if (title) form.append('title', title);
  if (author) form.append('author', author);
  const res = await client.post('/books', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return res.data;
}

export async function deleteBook(id: string): Promise<void> {
  await client.delete(`/books/${id}`);
}
