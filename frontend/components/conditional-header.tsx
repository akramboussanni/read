'use client';

import { usePathname } from 'next/navigation';
import { Header } from './header';

const AUTH_ROUTES = ['/forgot-password', '/reset-password', '/confirm-email', '/resend-confirmation'];

export function ConditionalHeader() {
  const pathname = usePathname();
  
  // Don't show header on auth pages
  const isAuthPage = AUTH_ROUTES.some(route => pathname.startsWith(route));
  
  if (isAuthPage) {
    return null;
  }
  
  return <Header />;
}
