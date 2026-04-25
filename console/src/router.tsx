import { createBrowserRouter, Navigate } from 'react-router-dom'
import { App } from './app'
import { LoginPage } from './routes/login'
import { DashboardPage } from './routes/dashboard'
import { ProjectsPage } from './routes/projects/list'
import { ProjectDetailPage } from './routes/projects/detail'
import { ProjectMembersPage } from './routes/projects/members'
import { RepositoryDetailPage } from './routes/repositories/detail'
import { RobotTokensPage } from './routes/robot-tokens'
import { UsersPage } from './routes/users'
import { AuditLogsPage } from './routes/audit-logs'
import { SettingsPage } from './routes/settings'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <App />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <DashboardPage /> },
      { path: 'projects', element: <ProjectsPage /> },
      { path: 'projects/:project', element: <ProjectDetailPage /> },
      { path: 'projects/:project/members', element: <ProjectMembersPage /> },
      { path: 'projects/:project/repositories/:repo', element: <RepositoryDetailPage /> },
      { path: 'robot-tokens', element: <RobotTokensPage /> },
      { path: 'users', element: <UsersPage /> },
      { path: 'audit-logs', element: <AuditLogsPage /> },
      { path: 'settings', element: <SettingsPage /> },
    ],
  },
])
