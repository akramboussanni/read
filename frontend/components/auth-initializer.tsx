'use client';

export function AuthInitializer({ children }: { children: React.ReactNode }) {
  // No automatic auth check - let protected routes handle it
  return <>{children}</>;
}
