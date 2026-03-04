'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { quizApi } from '@/lib/api/quiz';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, Edit, Trash2, Eye } from 'lucide-react';

interface UserQuiz {
  id: string;
  title: string;
  description: string;
  is_public: boolean;
  created_at: number;
  question_count?: number;
}

export default function MyQuizzesPage() {
  const router = useRouter();
  const [quizzes, setQuizzes] = useState<UserQuiz[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadQuizzes();
  }, []);

  const loadQuizzes = async () => {
    try {
      setLoading(true);
      const data = await quizApi.getMyQuizzes();
      setQuizzes(data as unknown as UserQuiz[]);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load quizzes');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (quizId: string) => {
    if (!confirm('Are you sure you want to delete this quiz?')) {
      return;
    }

    try {
      await quizApi.deleteQuiz(quizId);
      setQuizzes(quizzes.filter((q) => q.id !== quizId));
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to delete quiz');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <p className="text-gray-600 dark:text-gray-400">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold">My Quizzes</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-2">
              Manage your custom quizzes
            </p>
          </div>
          <Button onClick={() => router.push('/quizzes/create')}>
            <Plus className="w-4 h-4 mr-2" />
            Create Quiz
          </Button>
        </div>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        {quizzes.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                You haven't created any quizzes yet
              </p>
              <Button onClick={() => router.push('/quizzes/create')}>
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Quiz
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {quizzes.map((quiz) => (
              <Card key={quiz.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="flex items-center gap-2">
                        {quiz.title}
                        {quiz.is_public && (
                          <span className="text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 px-2 py-1 rounded">
                            Public
                          </span>
                        )}
                      </CardTitle>
                      <CardDescription className="mt-1">
                        {quiz.description || 'No description'}
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-gray-600 dark:text-gray-400">Created</span>
                      <span className="font-medium">
                        {new Date(quiz.created_at * 1000).toLocaleDateString()}
                      </span>
                    </div>

                    {quiz.question_count !== undefined && (
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-600 dark:text-gray-400">Questions</span>
                        <span className="font-medium">{quiz.question_count}</span>
                      </div>
                    )}

                    <div className="flex gap-2 pt-4 border-t">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1"
                        onClick={() => router.push(`/quizzes/${quiz.id}`)}
                      >
                        <Eye className="w-4 h-4 mr-1" />
                        View
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => router.push(`/quizzes/${quiz.id}/edit`)}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDelete(quiz.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
