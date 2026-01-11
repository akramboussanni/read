'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';

export default function ConfirmEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const confirmEmail = async () => {
      const token = searchParams.get('token');
      
      if (!token) {
        setError('Invalid or missing confirmation token');
        setIsLoading(false);
        return;
      }

      try {
        await authApi.confirmEmail({ token });
        setSuccess(true);
      } catch (err: any) {
        const errorMsg = err.response?.data?.message || 'Failed to confirm email';
        setError(errorMsg);
      } finally {
        setIsLoading(false);
      }
    };

    confirmEmail();
  }, [searchParams, router]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Confirming Email</CardTitle>
            <CardDescription>
              Please wait while we verify your email...
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Email Confirmed!</CardTitle>
            <CardDescription>
              Your email has been successfully verified
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-600">
              Your email is now verified. You can now log in and use all features.
            </p>
            <p className="text-sm text-gray-600 mt-2">
              Redirecting to login page...
            </p>
          </CardContent>
          <CardFooter>
            <Link href="/login" className="w-full">
              <Button className="w-full">
                Go to Login
              </Button>
            </Link>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Confirmation Failed</CardTitle>
          <CardDescription>
            Unable to verify your email
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
            {error}
          </div>
          <p className="text-sm text-gray-600">
            {error.includes('expired') 
              ? 'Your confirmation link has expired. Please request a new one.' 
              : 'The confirmation link may be invalid or already used.'}
          </p>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <Link href="/resend-confirmation" className="w-full">
            <Button className="w-full">
              Resend Confirmation Email
            </Button>
          </Link>
          <Link href="/login" className="w-full">
            <Button variant="outline" className="w-full">
              Back to Login
            </Button>
          </Link>
        </CardFooter>
      </Card>
    </div>
  );
}
