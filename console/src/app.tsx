import { Outlet, Navigate, useLocation } from "react-router-dom";
import { useAuth } from "./lib/auth";
import { Shell } from "./components/layout/shell";

export function App() {
  const { user, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return (
    <Shell>
      <Outlet />
    </Shell>
  );
}
