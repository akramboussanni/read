'use client';

import React, { useEffect } from 'react';
import { useAuthStore } from '@/lib/store/auth-store';

export function AuthInitializer({ children }: { children: React.ReactNode }) {
  const { checkAuth } = useAuthStore();

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  return <>{children}</>;
}
