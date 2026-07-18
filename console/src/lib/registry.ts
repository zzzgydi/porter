/// <reference types="vite/client" />

const REGISTRY_PUBLIC_URL =
  import.meta.env.VITE_REGISTRY_PUBLIC_URL || `http://${window.location.host}`;

// Registry host (without protocol) used to build docker login/pull commands.
export const REGISTRY_HOST = REGISTRY_PUBLIC_URL.replace(/^https?:\/\//, "");
