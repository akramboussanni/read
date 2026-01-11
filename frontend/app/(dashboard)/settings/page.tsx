'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/lib/store/auth-store';
import { ProtectedRoute } from '@/components/protected-route';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';

export default function SettingsPage() {
  const { user, checkAuth } = useAuthStore();
  const [email, setEmail] = useState('');
  const [emailPassword, setEmailPassword] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoadingEmail, setIsLoadingEmail] = useState(false);
  const [isLoadingPassword, setIsLoadingPassword] = useState(false);
  const [emailError, setEmailError] = useState('');
  const [emailSuccess, setEmailSuccess] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // Sync email state with user data
  useEffect(() => {
    if (user?.email) {
      setEmail(user.email);
    }
  }, [user?.email]);

  const handleEmailUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    setEmailSuccess('');
    setIsLoadingEmail(true);

    try {
      const url = `${window.location.origin}/confirm-email`;
      const payload: any = { email, url };
      
      // Include password if current email is verified
      if (user?.email && user?.email_confirmed) {
        if (!emailPassword) {
          setEmailError('Password required to change verified email');
          setIsLoadingEmail(false);
          return;
        }
        payload.password = emailPassword;
      }
      
      await authApi.addEmail(payload);
      setEmailSuccess('Verification email sent! Check your inbox.');
      setEmailPassword(''); // Clear password field
      // Refresh user data
      await checkAuth();
    } catch (err: any) {
      setEmailError(err.response?.data?.message || 'Failed to update email');
    } finally {
      setIsLoadingEmail(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    if (newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match');
      return;
    }

    setIsLoadingPassword(true);

    try {
      await authApi.changePassword({
        old_password: currentPassword,
        new_password: newPassword,
      });
      setPasswordSuccess('Password changed successfully');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setPasswordError(err.response?.data?.message || 'Failed to change password');
    } finally {
      setIsLoadingPassword(false);
    }
  };

  return (
    <ProtectedRoute>
      <div className="container mx-auto p-4 max-w-2xl">
        <div className="mb-6">
          <Link href="/profile" className="text-blue-600 hover:underline">
            ← Back to Profile
          </Link>
        </div>

        <h1 className="text-3xl font-bold mb-6">Settings</h1>

        {/* Email Settings */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Email Address</CardTitle>
            <CardDescription>
              {user?.email 
                ? user.email_confirmed
                  ? 'Update your email address (requires password verification for security)'
                  : 'Update your unverified email address'
                : 'Add an email address (required for quiz creation)'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleEmailUpdate} className="space-y-4">
              {emailError && (
                <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
                  {emailError}
                </div>
              )}
              {emailSuccess && (
                <div className="p-3 text-sm text-green-600 bg-green-50 rounded-md">
                  {emailSuccess}
                </div>
              )}
              
              {user?.email && user.email_confirmed && (
                <div className="p-3 text-sm text-blue-600 bg-blue-50 rounded-md">
                  🔒 Your email is verified. Password confirmation required for security.
                </div>
              )}
              
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="john@example.com"
                  required
                  disabled={isLoadingEmail}
                />
              </div>
              
              {user?.email && user.email_confirmed && (
                <div className="space-y-2">
                  <Label htmlFor="emailPassword">Password</Label>
                  <Input
                    id="emailPassword"
                    type="password"
                    value={emailPassword}
                    onChange={(e) => setEmailPassword(e.target.value)}
                    placeholder="••••••••"
                    required
                    disabled={isLoadingEmail}
                  />
                  <p className="text-xs text-gray-500">
                    Enter your password to confirm this change
                  </p>
                </div>
              )}
              
              <Button 
                type="submit" 
                disabled={isLoadingEmail || email === user?.email}
              >
                {isLoadingEmail ? 'Updating...' : 'Update Email'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Password Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Change Password</CardTitle>
            <CardDescription>
              Update your password. You'll be logged out from all devices after changing.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handlePasswordChange} className="space-y-4">
              {passwordError && (
                <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
                  {passwordError}
                </div>
              )}
              {passwordSuccess && (
                <div className="p-3 text-sm text-green-600 bg-green-50 rounded-md">
                  {passwordSuccess}
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="currentPassword">Current Password</Label>
                <Input
                  id="currentPassword"
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  disabled={isLoadingPassword}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="newPassword">New Password</Label>
                <Input
                  id="newPassword"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  disabled={isLoadingPassword}
                  minLength={8}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm New Password</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  disabled={isLoadingPassword}
                  minLength={8}
                />
              </div>
              <Button 
                type="submit" 
                disabled={isLoadingPassword}
              >
                {isLoadingPassword ? 'Changing...' : 'Change Password'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </ProtectedRoute>
  );
}
