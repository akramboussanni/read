'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { getQuizDetail } from '@/lib/api/quiz';
import type { Quiz } from '@/lib/types/quiz';

export default function QuizStatsPage() {
  const params = useParams();
  const router = useRouter();
  const quizId = params.id as string;

  const [quiz, setQuiz] = useState<Quiz | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadQuizStats();
  }, [quizId]);

  const loadQuizStats = async () => {
    try {
      setLoading(true);
      const data = await getQuizDetail(quizId);
      setQuiz(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load quiz statistics');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-lg">Loading statistics...</div>
        </div>
      </div>
    );
  }

  if (error || !quiz) {
    return (
      <div className="container mx-auto py-8">
        <Card className="p-6 bg-red-50 dark:bg-red-900/20">
          <p className="text-red-600 dark:text-red-300">{error || 'Quiz not found'}</p>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{quiz.title}</h1>
          <div className="flex items-center gap-3 mt-2">
            <p className="text-gray-600 dark:text-gray-400">Quiz Statistics</p>
            <span className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded font-mono">
              v{quiz.version || 1}
            </span>
          </div>
        </div>
        <Button variant="outline" onClick={() => router.push('/admin/quizzes')}>
          Back to Quizzes
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Version</div>
          <div className="text-3xl font-bold mt-2 font-mono">v{quiz.version || 1}</div>
          <div className="text-xs text-gray-500 mt-1">Current version</div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Pass Percentage</div>
          <div className="text-3xl font-bold mt-2">{quiz.pass_percentage || 70}%</div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Coin Reward</div>
          <div className="text-3xl font-bold mt-2">{quiz.coin_reward || 0}</div>
        </Card>

        <Card className="p-6">
          <div className="text-sm text-gray-600 dark:text-gray-400">Level Order</div>
          <div className="text-3xl font-bold mt-2">{quiz.level_order || 0}</div>
        </Card>
      </div>

      <Card className="p-6">
        <h2 className="text-xl font-bold mb-4">Quiz Information</h2>
        <div className="space-y-3">
          <div>
            <span className="text-gray-600 dark:text-gray-400">Description:</span>
            <p className="mt-1">{quiz.description || 'No description'}</p>
          </div>
          <div>
            <span className="text-gray-600 dark:text-gray-400">Status:</span>
            <span className="ml-2">{quiz.is_public ? 'Public' : 'Private'}</span>
          </div>
          <div>
            <span className="text-gray-600 dark:text-gray-400">Shuffle Questions:</span>
            <span className="ml-2">{quiz.shuffle_questions ? 'Yes' : 'No'}</span>
          </div>
          {quiz.updated_at && (
            <div>
              <span className="text-gray-600 dark:text-gray-400">Last Updated:</span>
              <span className="ml-2">
                {new Date(quiz.updated_at * 1000).toLocaleDateString()}
              </span>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
