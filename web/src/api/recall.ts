import client from './client';
import type { RecallResult } from '../types';

export async function query(q: string): Promise<RecallResult> {
  const res = await client.post('/recall', { query: q });
  return res.data;
}
