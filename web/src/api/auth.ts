import client, { setAccessToken } from './client';
import type { AuthResponse } from '../types';

export async function register(email: string, password: string): Promise<AuthResponse> {
  const res = await client.post('/auth/register', { email, password });
  setAccessToken(res.data.access_token);
  return res.data;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await client.post('/auth/login', { email, password });
  setAccessToken(res.data.access_token);
  return res.data;
}

export async function refresh(): Promise<string> {
  const res = await client.post('/auth/refresh');
  const token = res.data.access_token;
  setAccessToken(token);
  return token;
}
