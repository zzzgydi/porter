/// <reference types="vite/client" />
import type { Project, Member, Repository, Tag, RobotToken, User, AuditLog } from './types'

const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

export class APIError extends Error {
  constructor(public status: number, public data: unknown) {
    super(`API error ${status}`)
  }
}

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...opts?.headers,
    },
    ...opts,
  })
  if (res.status === 401) {
    throw new APIError(401, { message: 'Unauthorized' })
  }
  if (!res.ok) {
    let data: unknown
    const contentType = res.headers.get('content-type')
    if (contentType && contentType.includes('application/json')) {
      data = await res.json().catch(() => ({}))
    } else {
      const text = await res.text().catch(() => '')
      data = { message: text || `Request failed with status ${res.status}` }
    }
    throw new APIError(res.status, data)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  login: (email: string, password: string) =>
    request<{ id: string; email: string; name: string; role: string }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<{ ok: boolean }>('/api/auth/logout', { method: 'POST' }),
  me: () => request<{ id: string; email: string; name: string; role: string }>('/api/me'),

  projects: {
    list: () => request<Project[]>('/api/projects'),
    create: (data: { name: string; display_name?: string; visibility?: string }) =>
      request<Project>('/api/projects', { method: 'POST', body: JSON.stringify(data) }),
    get: (name: string) => request<Project>(`/api/projects/${name}`),
    update: (name: string, data: { display_name?: string; visibility?: string }) =>
      request<{ ok: boolean }>(`/api/projects/${name}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (name: string) => request<{ ok: boolean }>(`/api/projects/${name}`, { method: 'DELETE' }),
    members: {
      list: (project: string) => request<Member[]>(`/api/projects/${project}/members`),
      add: (project: string, data: { email: string; role: string }) =>
        request<Member>(`/api/projects/${project}/members`, { method: 'POST', body: JSON.stringify(data) }),
      remove: (project: string, userId: string) =>
        request<{ ok: boolean }>(`/api/projects/${project}/members/${userId}`, { method: 'DELETE' }),
    },
  },

  repositories: {
    list: (project: string) => request<Repository[]>(`/api/projects/${project}/repositories`),
    get: (project: string, repo: string) => request<Repository>(`/api/projects/${project}/repositories/${repo}`),
  },

  tags: {
    list: (project: string, repo: string) =>
      request<Tag[]>(`/api/projects/${project}/repositories/${repo}/tags`),
    delete: (project: string, repo: string, tag: string) =>
      request<{ ok: boolean }>(`/api/projects/${project}/repositories/${repo}/tags/${tag}`, { method: 'DELETE' }),
  },

  robots: {
    list: (project?: string) => {
      const qs = new URLSearchParams()
      if (project) qs.set('project', project)
      const query = qs.toString()
      return request<RobotToken[]>(`/api/robot-tokens${query ? '?' + query : ''}`)
    },
    create: (data: { project_id: string; name: string; permissions: Record<string, string[]> }) =>
      request<RobotToken>('/api/robot-tokens', { method: 'POST', body: JSON.stringify(data) }),
    revoke: (id: string) => request<{ ok: boolean }>(`/api/robot-tokens/${id}`, { method: 'DELETE' }),
  },

  users: {
    list: () => request<User[]>('/api/users'),
    create: (data: { email: string; name: string; password: string; role: string }) =>
      request<User>('/api/users', { method: 'POST', body: JSON.stringify(data) }),
    delete: (id: string) => request<{ ok: boolean }>(`/api/users/${id}`, { method: 'DELETE' }),
  },

  audit: {
    list: (limit?: number, offset?: number) =>
      request<AuditLog[]>(`/api/audit-logs?limit=${limit || 50}&offset=${offset || 0}`),
  },
}
