'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { updateQuiz } from '@/lib/api/admin';
import { getQuizDetail } from '@/lib/api/quiz';
import type { UpdateQuizRequest } from '@/lib/types/admin';

export default function EditQuiz() {
  const router = useRouter();
  const params = useParams();
  const quizId = params.id as string;

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    title: '',
    description: '',
    pass_percentage: 70,
    shuffle_questions: true,
    gives_coins: false,
    coin_reward: 0,
    level_order: 0,
    is_public: true,
  });

  useEffect(() => {
    loadQuiz();
  }, [quizId]);

  const loadQuiz = async () => {
    try {
      setLoading(true);
      const quiz = await getQuizDetail(quizId);
      setFormData({
        title: quiz.title || '',
        description: quiz.description || '',
        pass_percentage: quiz.pass_percentage || 70,
        shuffle_questions: quiz.shuffle_questions ?? true,
        gives_coins: quiz.gives_coins ?? false,
        coin_reward: quiz.coin_reward || 0,
        level_order: quiz.level_order || 0,
        is_public: quiz.is_public ?? true,
      });
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load quiz');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);

    try {
      const updateData: UpdateQuizRequest = {
        title: formData.title,
        description: formData.description,
        pass_percentage: formData.pass_percentage,
        shuffle_questions: formData.shuffle_questions,
        gives_coins: formData.gives_coins,
        coin_reward: formData.coin_reward,
        level_order: formData.level_order,
        is_public: formData.is_public,
      };

      await updateQuiz(quizId, updateData);
      router.push('/admin/quizzes');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update quiz');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-lg">Loading quiz...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 max-w-2xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Edit Quiz</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">Update quiz settings and properties</p>
      </div>

      {error && (
        <Card className="p-4 mb-6 bg-red-50 dark:bg-red-900/20">
          <p className="text-red-600 dark:text-red-300">{error}</p>
        </Card>
      )}

      <form onSubmit={handleSubmit}>
        <Card className="p-6 space-y-6">
          <div className="space-y-2">
            <Label htmlFor="title">Title *</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full min-h-[100px] px-3 py-2 border rounded-md"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="pass_percentage">Pass Percentage (%)</Label>
              <Input
                id="pass_percentage"
                type="number"
                min="0"
                max="100"
                value={formData.pass_percentage}
                onChange={(e) => setFormData({ ...formData, pass_percentage: parseInt(e.target.value) })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="level_order">Level Order</Label>
              <Input
                id="level_order"
                type="number"
                min="0"
                value={formData.level_order}
                onChange={(e) => setFormData({ ...formData, level_order: parseInt(e.target.value) })}
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="coin_reward">Coin Reward</Label>
              <Input
                id="coin_reward"
                type="number"
                min="0"
                value={formData.coin_reward}
                onChange={(e) => setFormData({ ...formData, coin_reward: parseInt(e.target.value) })}
              />
            </div>

            <div className="flex items-center space-x-2 pt-8">
              <input
                type="checkbox"
                id="gives_coins"
                checked={formData.gives_coins}
                onChange={(e) => setFormData({ ...formData, gives_coins: e.target.checked })}
                className="rounded"
              />
              <Label htmlFor="gives_coins">Gives Coins</Label>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="shuffle_questions"
                checked={formData.shuffle_questions}
                onChange={(e) => setFormData({ ...formData, shuffle_questions: e.target.checked })}
                className="rounded"
              />
              <Label htmlFor="shuffle_questions">Shuffle Questions</Label>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="is_public"
                checked={formData.is_public}
                onChange={(e) => setFormData({ ...formData, is_public: e.target.checked })}
                className="rounded"
              />
              <Label htmlFor="is_public">Public</Label>
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <Button type="submit" disabled={saving} className="flex-1">
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push('/admin/quizzes')}
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </Card>
      </form>
    </div>
  );
}
