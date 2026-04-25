export interface User {
  id: string
  email: string
  name: string
  role: string
  created_at: string
}

export interface Project {
  id: string
  name: string
  display_name: string
  visibility: string
  created_at: string
  updated_at: string
}

export interface Member {
  id: string
  project_id: string
  user_id: string
  role: string
  email: string
  name: string
  created_at: string
}

export interface Repository {
  id: string
  project_id: string
  name: string
  full_name: string
  description: string
  created_at: string
  updated_at: string
}

export interface Tag {
  id: string
  repository_id: string
  name: string
  digest: string
  media_type: string
  size_bytes: number
  pushed_at: string
  updated_at: string
}

export interface RobotToken {
  id: string
  username: string
  name: string
  project_id: string
  created_at: string
  token?: string
}

export interface AuditLog {
  id: string
  actor_type: string
  actor_id: string
  action: string
  target: string
  metadata: Record<string, unknown> | null
  ip: string
  user_agent: string
  created_at: string
}
