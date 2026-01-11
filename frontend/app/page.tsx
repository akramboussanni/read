'use client';

import Link from 'next/link';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { BookOpen, Trophy, Coins, TrendingUp } from 'lucide-react';

export default function Home() {
  const { isAuthenticated } = useAuthStore();

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-16">
        {/* Hero Section */}
        <div className="flex flex-col items-center justify-center text-center space-y-6 py-20">
          <h1 className="text-6xl font-bold text-gray-900 dark:text-white">
            Master Your Knowledge
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-400 max-w-2xl">
            Test your knowledge, track your progress, and earn rewards with our interactive quiz system
          </p>
          
          <div className="flex gap-4 mt-8">
            {isAuthenticated ? (
              <Link href="/quizzes">
                <Button size="lg" className="text-lg px-8 py-6">
                  <BookOpen className="mr-2 h-5 w-5" />
                  Start Learning
                </Button>
              </Link>
            ) : (
              <>
                <Link href="/register">
                  <Button size="lg" className="text-lg px-8 py-6">
                    Get Started
                  </Button>
                </Link>
                <Link href="/login">
                  <Button size="lg" variant="outline" className="text-lg px-8 py-6">
                    Login
                  </Button>
                </Link>
              </>
            )}
          </div>
        </div>

        {/* Features Section */}
        <div className="grid md:grid-cols-3 gap-8 mt-20">
          <Card>
            <CardHeader>
              <div className="w-12 h-12 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center mb-4">
                <BookOpen className="h-6 w-6 text-blue-600 dark:text-blue-400" />
              </div>
              <CardTitle>Progressive Learning</CardTitle>
              <CardDescription>
                Unlock quizzes as you progress through levels, ensuring a structured learning path
              </CardDescription>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader>
              <div className="w-12 h-12 rounded-full bg-yellow-100 dark:bg-yellow-900 flex items-center justify-center mb-4">
                <Trophy className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
              </div>
              <CardTitle>Track Your Progress</CardTitle>
              <CardDescription>
                Monitor your performance, maintain streaks, and see your improvement over time
              </CardDescription>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader>
              <div className="w-12 h-12 rounded-full bg-amber-100 dark:bg-amber-900 flex items-center justify-center mb-4">
                <Coins className="h-6 w-6 text-amber-600 dark:text-amber-400" />
              </div>
              <CardTitle>Earn Rewards</CardTitle>
              <CardDescription>
                Complete quizzes to earn coins and unlock achievements as you master new topics
              </CardDescription>
            </CardHeader>
          </Card>
        </div>

        {/* Stats Section */}
        {isAuthenticated && (
          <div className="mt-20 text-center">
            <h2 className="text-3xl font-bold mb-4">Ready to Continue?</h2>
            <p className="text-gray-600 dark:text-gray-400 mb-8">
              Jump back into your learning journey
            </p>
            <Link href="/quizzes">
              <Button size="lg" className="text-lg px-8 py-6">
                <TrendingUp className="mr-2 h-5 w-5" />
                View My Progress
              </Button>
            </Link>
          </div>
        )}
      </main>
    </div>
  );
}
