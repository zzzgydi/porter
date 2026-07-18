import { Link, useLocation } from "react-router-dom";
import { cn } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import {
  LayoutDashboard,
  FolderKanban,
  KeyRound,
  Users,
  ClipboardList,
  Settings,
  Container,
  PanelLeft,
} from "lucide-react";
import { useState } from "react";

const nav = [
  { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { name: "Projects", href: "/projects", icon: FolderKanban },
  {
    name: "Robot Tokens",
    href: "/robot-tokens",
    icon: KeyRound,
    adminOnly: true,
  },
  { name: "Users", href: "/users", icon: Users, adminOnly: true },
  {
    name: "Audit Logs",
    href: "/audit-logs",
    icon: ClipboardList,
    adminOnly: true,
  },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const location = useLocation();
  const { user } = useAuth();
  const isAdmin = user?.role === "platform_admin";
  const [collapsed, setCollapsed] = useState(false);

  const visibleNav = nav.filter((item) => !item.adminOnly || isAdmin);

  return (
    <aside
      className={cn(
        "flex flex-col border-r bg-card transition-all duration-300",
        collapsed ? "w-16" : "w-64",
      )}
    >
      <div
        className={cn(
          "flex h-16 items-center justify-between border-b transition-all duration-300",
          collapsed ? "px-5" : "px-4",
        )}
      >
        <div className="flex items-center overflow-hidden">
          <Container className="h-6 w-6 shrink-0 text-primary" />
          <span
            className={cn(
              "whitespace-nowrap text-lg font-bold transition-all duration-300",
              collapsed
                ? "ml-0 max-w-0 opacity-0"
                : "ml-2 max-w-24 opacity-100",
            )}
          >
            Porter
          </span>
        </div>
        <button
          onClick={() => setCollapsed(true)}
          className={cn(
            "shrink-0 overflow-hidden rounded-md p-1 text-muted-foreground transition-all duration-300 hover:bg-accent hover:text-foreground",
            collapsed ? "invisible w-0 px-0 opacity-0" : "w-6 opacity-100",
          )}
          title="Collapse sidebar"
        >
          <PanelLeft className="h-4 w-4" />
        </button>
      </div>
      <nav className="flex-1 space-y-1 p-3">
        {visibleNav.map((item) => {
          const active = location.pathname.startsWith(item.href);
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                "flex items-center rounded-lg py-2.5 text-sm font-medium transition-all duration-300",
                active
                  ? "bg-primary/10 text-primary shadow-sm"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                collapsed ? "px-2.5" : "px-3",
              )}
              title={collapsed ? item.name : undefined}
            >
              <item.icon
                className={cn("h-5 w-5 shrink-0", active && "text-primary")}
              />
              <span
                className={cn(
                  "overflow-hidden whitespace-nowrap transition-all duration-300",
                  collapsed
                    ? "ml-0 max-w-0 opacity-0"
                    : "ml-3 max-w-32 opacity-100",
                )}
              >
                {item.name}
              </span>
            </Link>
          );
        })}
      </nav>
      <div
        className={cn(
          "overflow-hidden border-t p-3 transition-all duration-300",
          collapsed
            ? "max-h-[61px] border-border opacity-100"
            : "invisible max-h-0 border-transparent opacity-0",
        )}
      >
        <button
          onClick={() => setCollapsed(false)}
          className="flex w-full items-center justify-center rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          title="Expand sidebar"
        >
          <PanelLeft className="h-5 w-5" />
        </button>
      </div>
    </aside>
  );
}
