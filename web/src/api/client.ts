import axios from 'axios';

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

const client = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

client.interceptors.request.use((config) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;

    // Don't retry auth endpoints — prevents infinite refresh loop.
    const isAuthRequest = original.url?.startsWith('/auth/');
    if (error.response?.status === 401 && !original._retry && !isAuthRequest) {
      original._retry = true;
      try {
        const res = await axios.post('/api/auth/refresh', {}, { withCredentials: true });
        const newToken = res.data.access_token;
        setAccessToken(newToken);
        original.headers.Authorization = `Bearer ${newToken}`;
        return client(original);
      } catch {
        setAccessToken(null);
        return Promise.reject(error);
      }
    }
    return Promise.reject(error);
  }
);

export default client;
