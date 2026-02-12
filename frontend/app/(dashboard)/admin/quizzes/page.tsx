'use client';

import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { getQuizStats, deleteQuiz } from '@/lib/api/admin';
import type { QuizStatsResponse } from '@/lib/types/admin';
import Link from 'next/link';

export default function QuizManagement() {
  const [quizzes, setQuizzes] = useState<QuizStatsResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);

  useEffect(() => {
    loadQuizzes();
  }, []);

  const loadQuizzes = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getQuizStats(100, 0);
      setQuizzes(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load quizzes');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (quizId: string, title: string) => {
    if (!confirm(`Are you sure you want to delete "${title}"?`)) {
      return;
    }

    try {
      setDeleting(quizId);
      await deleteQuiz(quizId);
      setQuizzes(quizzes.filter((q) => q.quiz_id !== quizId));
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to delete quiz');
    } finally {
      setDeleting(null);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-lg">Loading quizzes...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Quiz Management</h1>
        <div className="flex gap-3">
          <Link href="/admin">
            <Button variant="outline">Back to Dashboard</Button>
          </Link>
          <Link href="/admin/quizzes/create">
            <Button>Create Quiz</Button>
          </Link>
        </div>
      </div>

      {error && (
        <Card className="p-4 bg-red-50 dark:bg-red-900/20">
          <p className="text-red-600 dark:text-red-300">{error}</p>
        </Card>
      )}

      <Card className="p-6">
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
                <th className="pb-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {quizzes.map((quiz) => (
                <tr key={quiz.quiz_id} className="text-sm">
                  <td className="py-4">
                    <div className="font-medium">{quiz.title}</div>
                    <div className="flex gap-2 mt-1">
                      {quiz.is_system && (
                        <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300 rounded">
                          System
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="py-4 text-gray-600 dark:text-gray-400">
                    {quiz.creator_username || 'System'}
                  </td>
                  <td className="py-4 text-right">
                    <span className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded font-mono">
                      v{quiz.version || 1}
                    </span>
                  </td>
                  <td className="py-4 text-right">{quiz.total_attempts}</td>
                  <td className="py-4 text-right font-mono">{quiz.average_score.toFixed(1)}%</td>
                  <td className="py-4 text-right">
                    <span
                      className={
                        quiz.pass_rate >= 70
                          ? 'text-green-600 font-medium'
                          : quiz.pass_rate >= 50
                          ? 'text-yellow-600'
                          : 'text-red-600'
                      }
                    >
                      {quiz.pass_rate.toFixed(1)}%
                    </span>
                  </td>
                  <td className="py-4 text-right">
                    <div className="flex gap-2 justify-end">
                      <Link href={`/admin/quizzes/${quiz.quiz_id}/edit`}>
                        <Button size="sm" variant="outline">
                          Edit
                        </Button>
                      </Link>
                      <Link href={`/admin/quizzes/${quiz.quiz_id}/stats`}>
                        <Button size="sm" variant="outline">
                          Stats
                        </Button>
                      </Link>
                      {!quiz.is_system && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDelete(quiz.quiz_id, quiz.title)}
                          disabled={deleting === quiz.quiz_id}
                          className="text-red-600 hover:text-red-700"
                        >
                          {deleting === quiz.quiz_id ? 'Deleting...' : 'Delete'}
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {quizzes.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              No quizzes found. Create your first quiz!
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
