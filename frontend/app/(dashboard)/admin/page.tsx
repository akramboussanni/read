'use client';

import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import { getSystemStats, getQuizStats } from '@/lib/api/admin';
import type { SystemStatsResponse, QuizStatsResponse } from '@/lib/types/admin';
import Link from 'next/link';

export default function AdminDashboard() {
  const [systemStats, setSystemStats] = useState<SystemStatsResponse | null>(null);
  const [quizStats, setQuizStats] = useState<QuizStatsResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [stats, quizzes] = await Promise.all([
        getSystemStats(),
        getQuizStats(10, 0),
      ]);
      setSystemStats(stats);
      setQuizStats(quizzes);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load admin data');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-lg">Loading admin dashboard...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-8">
        <Card className="p-6 bg-red-50 dark:bg-red-900/20">
          <h2 className="text-xl font-bold text-red-800 dark:text-red-200 mb-2">
            Error Loading Dashboard
          </h2>
          <p className="text-red-600 dark:text-red-300">{error}</p>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <div className="flex gap-4">
          <Link
            href="/admin/quizzes"
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Manage Quizzes
          </Link>
          <Link
            href="/admin/users"
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
          >
            Manage Users
          </Link>
        </div>
      </div>

      {/* System Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Total Users</div>
          <div className="text-3xl font-bold mt-2">{systemStats?.total_users || 0}</div>
          <div className="text-xs text-gray-500 mt-1">
            {systemStats?.active_users_7d || 0} active (7d)
          </div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Total Quizzes</div>
          <div className="text-3xl font-bold mt-2">{systemStats?.total_quizzes || 0}</div>
          <div className="text-xs text-gray-500 mt-1">Active quizzes</div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Total Attempts</div>
          <div className="text-3xl font-bold mt-2">{systemStats?.total_attempts || 0}</div>
          <div className="text-xs text-gray-500 mt-1">Completed attempts</div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Avg Score</div>
          <div className="text-3xl font-bold mt-2">
            {systemStats?.average_score ? `${systemStats.average_score.toFixed(1)}%` : 'N/A'}
          </div>
          <div className="text-xs text-gray-500 mt-1">Overall average</div>
        </Card>
      </div>

      {/* Top Quizzes */}
      <Card className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold">Quiz Statistics</h2>
          <Link
            href="/admin/quizzes/stats"
            className="text-blue-600 hover:underline text-sm"
          >
            View All
          </Link>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b">
              <tr className="text-left text-sm text-gray-600 dark:text-gray-400">
                <th className="pb-3">Quiz</th>
                <th className="pb-3">Creator</th>
                <th className="pb-3 text-right">Version</th>
                <th className="pb-3 text-right">Attempts</th>
                <th className="pb-3 text-right">Avg Score</th>
                <th className="pb-3 text-right">Pass Rate</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {quizStats.map((quiz) => (
                <tr key={quiz.quiz_id} className="text-sm">
                  <td className="py-3">
                    <div className="font-medium">{quiz.title}</div>
                    {quiz.is_system && (
                      <span className="text-xs text-blue-600 dark:text-blue-400">System</span>
                    )}
                  </td>
                  <td className="py-3 text-gray-600 dark:text-gray-400">
                    {quiz.creator_username || 'System'}
                  </td>
                  <td className="py-3 text-right text-gray-500">
                    v{quiz.version || 1}
                  </td>
                  <td className="py-3 text-right">{quiz.total_attempts}</td>
                  <td className="py-3 text-right">{quiz.average_score.toFixed(1)}%</td>
                  <td className="py-3 text-right">
                    <span
                      className={
                        quiz.pass_rate >= 70
                          ? 'text-green-600'
                          : quiz.pass_rate >= 50
                          ? 'text-yellow-600'
                          : 'text-red-600'
                      }
                    >
                      {quiz.pass_rate.toFixed(1)}%
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
