'use client';

import { useAuthStore } from '@/lib/store/auth-store';
import { ProtectedRoute } from '@/components/protected-route';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

export default function ProfilePage() {
  const router = useRouter();
  const { user, logout, logoutAll, isLoading } = useAuthStore();

  const handleLogout = async () => {
    await logout();
    router.push('/');
  };

  const handleLogoutAll = async () => {
    await logoutAll();
    router.push('/');
  };

  return (
    <ProtectedRoute>
      <div className="container mx-auto p-4 max-w-2xl">
        <h1 className="text-3xl font-bold mb-6">Profile</h1>
        
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Account Information</CardTitle>
            <CardDescription>Your account details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-gray-500">Username</p>
              <p className="text-lg">{user?.username}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Email</p>
              <div className="flex items-center gap-2">
                <p className="text-lg">{user?.email || 'Not provided'}</p>
                {user?.email && (
                  <span 
                    className={`text-xs px-2 py-1 rounded-full ${
                      user.email_confirmed 
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' 
                        : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                    }`}
                  >
                    {user.email_confirmed ? 'Verified' : 'Unverified'}
                  </span>
                )}
              </div>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Role</p>
              <p className="text-lg capitalize">{user?.role}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Account Created</p>
              <p className="text-lg">
                {user?.created_at 
                  ? new Date(user.created_at * 1000).toLocaleDateString()
                  : 'Unknown'}
              </p>
            </div>
          </CardContent>
          <CardFooter>
            <Link href="/settings" className="w-full">
              <Button variant="outline" className="w-full">
                Edit Profile
              </Button>
            </Link>
          </CardFooter>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Sessions</CardTitle>
            <CardDescription>Manage your active sessions</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-gray-600">
              You can log out from this device or all devices where you're currently logged in.
            </p>
          </CardContent>
          <CardFooter className="flex gap-4">
            <Button 
              variant="outline" 
              onClick={handleLogout}
              disabled={isLoading}
              className="flex-1"
            >
              Logout This Device
            </Button>
            <Button 
              variant="destructive" 
              onClick={handleLogoutAll}
              disabled={isLoading}
              className="flex-1"
            >
              Logout All Devices
            </Button>
          </CardFooter>
        </Card>
      </div>
    </ProtectedRoute>
  );
}
