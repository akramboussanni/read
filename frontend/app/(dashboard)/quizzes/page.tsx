'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { progressionApi } from '@/lib/api/progression';
import { ProgressionStatus, QuizPreview } from '@/lib/types/progression';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Lock, CheckCircle, Trophy, Coins, Flame } from 'lucide-react';

export default function QuizzesPage() {
  const router = useRouter();
  const [status, setStatus] = useState<ProgressionStatus | null>(null);
  const [quizzes, setQuizzes] = useState<QuizPreview[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [statusData, quizzesData] = await Promise.all([
        progressionApi.getStatus(),
        progressionApi.listQuizzes(),
      ]);
      setStatus(statusData);
      setQuizzes(quizzesData);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load quizzes');
    } finally {
      setLoading(false);
    }
  };

  const handleStartQuiz = (quizId: number) => {
    router.push(`/quizzes/${quizId}`);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <p className="text-gray-600 dark:text-gray-400">Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 dark:text-red-400 mb-4">{error}</p>
          <Button onClick={loadData}>Retry</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-8">
        {/* Progression Status Header */}
        {status && (
          <Card className="mb-8">
            <CardHeader>
              <CardTitle className="text-2xl">Your Progress</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <div className="flex items-center gap-2">
                  <Trophy className="w-5 h-5 text-yellow-500" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Level</p>
                    <p className="text-xl font-bold">{status.current_level}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Completed</p>
                    <p className="text-xl font-bold">{status.total_quizzes_completed}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Coins className="w-5 h-5 text-amber-500" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Coins</p>
                    <p className="text-xl font-bold">{status.coin_balance}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Coins className="w-5 h-5 text-amber-600" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Total Earned</p>
                    <p className="text-xl font-bold">{status.total_coins_earned}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Flame className="w-5 h-5 text-orange-500" />
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Streak</p>
                    <p className="text-xl font-bold">{status.streak_days} days</p>
                  </div>
                </div>
              </div>

              {status.next_quiz && (
                <div className="mt-6 pt-6 border-t">
                  <h3 className="text-lg font-semibold mb-2">Next Recommended Quiz</h3>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">{status.next_quiz.title}</p>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {status.next_quiz.description}
                      </p>
                    </div>
                    <Button onClick={() => handleStartQuiz(status.next_quiz!.id)}>
                      Start Quiz
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Quizzes List */}
        <div className="space-y-4">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold">All Quizzes</h2>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => router.push('/quizzes/my')}>
                My Quizzes
              </Button>
              <Button onClick={() => router.push('/quizzes/create')}>
                Create Quiz
              </Button>
            </div>
          </div>
          
          {quizzes.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center">
                <p className="text-gray-600 dark:text-gray-400">No quizzes available yet.</p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {quizzes.map((quiz) => (
                <Card
                  key={quiz.id}
                  className={`relative ${
                    quiz.is_locked ? 'opacity-60' : 'hover:shadow-lg transition-shadow'
                  }`}
                >
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="flex items-center gap-2">
                          {quiz.title}
                          {quiz.is_completed && (
                            <CheckCircle className="w-5 h-5 text-green-500" />
                          )}
                          {quiz.is_locked && <Lock className="w-5 h-5 text-gray-400" />}
                        </CardTitle>
                        <CardDescription className="mt-1">
                          {quiz.description}
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-600 dark:text-gray-400">Level</span>
                        <span className="font-medium">{quiz.level}</span>
                      </div>
                      
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-600 dark:text-gray-400">Questions</span>
                        <span className="font-medium">{quiz.question_count}</span>
                      </div>

                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-600 dark:text-gray-400">Reward</span>
                        <span className="font-medium flex items-center gap-1">
                          <Coins className="w-4 h-4 text-amber-500" />
                          {quiz.coin_reward}
                        </span>
                      </div>

                      {quiz.is_completed && quiz.best_percentage !== undefined && (
                        <div className="flex items-center justify-between text-sm pt-2 border-t">
                          <span className="text-gray-600 dark:text-gray-400">Best Score</span>
                          <span className="font-medium text-green-600 dark:text-green-400">
                            {quiz.best_percentage.toFixed(1)}%
                          </span>
                        </div>
                      )}

                      <Button
                        className="w-full mt-4"
                        onClick={() => handleStartQuiz(quiz.id)}
                        disabled={quiz.is_locked}
                      >
                        {quiz.is_locked ? (
                          <>
                            <Lock className="w-4 h-4 mr-2" />
                            Locked
                          </>
                        ) : quiz.is_completed ? (
                          'Retake Quiz'
                        ) : (
                          'Start Quiz'
                        )}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
