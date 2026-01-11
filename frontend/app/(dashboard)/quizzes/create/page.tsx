'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { quizApi } from '@/lib/api/quiz';
import { Deck, Category } from '@/lib/types/quiz';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ArrowLeft, Plus, Trash2 } from 'lucide-react';

interface CategorySelection {
  category_id: string;
  question_count: number;
  category?: Category;
}

export default function CreateQuizPage() {
  const router = useRouter();
  const [decks, setDecks] = useState<Deck[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [selectedDeckId, setSelectedDeckId] = useState<string | null>(null);
  const [categorySelections, setCategorySelections] = useState<CategorySelection[]>([]);
  const [questionMode, setQuestionMode] = useState<'ar_to_fr' | 'fr_to_ar'>('ar_to_fr');
  const [timeLimit, setTimeLimit] = useState<number>(300); // 5 minutes default
  const [passPercentage, setPassPercentage] = useState<number>(70);
  const [shuffleQuestions, setShuffleQuestions] = useState(true);
  const [isPublic, setIsPublic] = useState(false);

  useEffect(() => {
    loadDecks();
  }, []);

  useEffect(() => {
    if (selectedDeckId) {
      loadCategories(selectedDeckId);
    } else {
      setCategories([]);
      setCategorySelections([]);
    }
  }, [selectedDeckId]);

  const loadDecks = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await quizApi.listDecks();
      setDecks(data);
    } catch (err: any) {
      const errorMsg = err.response?.data?.message || err.message || 'Failed to load decks';
      setError(errorMsg);
      console.error('Error loading decks:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadCategories = async (deckId: string) => {
    try {
      const data = await quizApi.getCategories(deckId);
      setCategories(data);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load categories');
    }
  };

  const addCategorySelection = (categoryId: string) => {
    const category = categories.find((c) => c.id === categoryId);
    if (!category) return;

    // Check if already selected
    if (categorySelections.some((cs) => cs.category_id === categoryId)) {
      return;
    }

    setCategorySelections([
      ...categorySelections,
      {
        category_id: categoryId,
        question_count: 5, // Default to 5 questions
        category,
      },
    ]);
  };

  const removeCategorySelection = (categoryId: string) => {
    setCategorySelections(categorySelections.filter((cs) => cs.category_id !== categoryId));
  };

  const updateQuestionCount = (categoryId: string, count: number) => {
    setCategorySelections(
      categorySelections.map((cs) =>
        cs.category_id === categoryId ? { ...cs, question_count: count } : cs
      )
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!title.trim()) {
      setError('Title is required');
      return;
    }

    if (!selectedDeckId) {
      setError('Please select a deck');
      return;
    }

    if (categorySelections.length === 0) {
      setError('Please select at least one category');
      return;
    }

    try {
      setSubmitting(true);
      await quizApi.createQuiz({
        title,
        description,
        deck_id: selectedDeckId,
        category_selections: categorySelections.map((cs) => ({
          category_id: cs.category_id,
          question_count: cs.question_count,
        })),
        question_mode: questionMode,
        time_limit: timeLimit > 0 ? timeLimit : undefined,
        pass_percentage: passPercentage,
        shuffle_questions: shuffleQuestions,
        is_public: isPublic,
      });

      router.push('/quizzes');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create quiz');
    } finally {
      setSubmitting(false);
    }
  };

  const availableCategories = categories.filter(
    (cat) => !categorySelections.some((cs) => cs.category_id === cat.id)
  );

  if (loading) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <p className="text-gray-600 dark:text-gray-400">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-8 max-w-4xl">
        <div className="mb-6">
          <Button variant="outline" onClick={() => router.back()} className="mb-4">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <h1 className="text-3xl font-bold">Create New Quiz</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Create a custom quiz by selecting questions from different categories
          </p>
        </div>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle>Basic Information</CardTitle>
              <CardDescription>Enter the basic details for your quiz</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="title">Quiz Title *</Label>
                <Input
                  id="title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="My Custom Quiz"
                  required
                />
              </div>

              <div>
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="A brief description of your quiz"
                />
              </div>

              <div>
                <Label htmlFor="deck">Select Deck *</Label>
                <select
                  id="deck"
                  value={selectedDeckId || ''}
                  onChange={(e) => setSelectedDeckId(e.target.value || null)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-black"
                  required
                >
                  <option value="">-- Select a deck --</option>
                  {decks.map((deck) => (
                    <option key={deck.id} value={deck.id}>
                      {deck.title} ({deck.question_count || 0} questions)
                    </option>
                  ))}
                </select>
              </div>
            </CardContent>
          </Card>

          {/* Category Selection */}
          {selectedDeckId && (
            <Card>
              <CardHeader>
                <CardTitle>Category Selection</CardTitle>
                <CardDescription>
                  Choose categories and how many questions from each
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Selected Categories */}
                {categorySelections.length > 0 && (
                  <div className="space-y-3">
                    {categorySelections.map((selection) => (
                      <div
                        key={selection.category_id}
                        className="flex items-center gap-4 p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
                      >
                        <div className="flex-1">
                          <p className="font-medium">{selection.category?.title}</p>
                          <p className="text-sm text-gray-600 dark:text-gray-400">
                            {selection.category?.question_count || 0} questions available
                          </p>
                        </div>
                        <div className="flex items-center gap-2">
                          <Label htmlFor={`count-${selection.category_id}`} className="text-sm">
                            Questions:
                          </Label>
                          <Input
                            id={`count-${selection.category_id}`}
                            type="number"
                            min="1"
                            max={selection.category?.question_count || 100}
                            value={selection.question_count}
                            onChange={(e) =>
                              updateQuestionCount(selection.category_id, Number(e.target.value))
                            }
                            className="w-20"
                          />
                        </div>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => removeCategorySelection(selection.category_id)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Add Category */}
                {availableCategories.length > 0 && (
                  <div>
                    <Label>Add Category</Label>
                    <div className="flex gap-2 mt-2">
                      <select
                        id="add-category"
                        className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-black"
                        onChange={(e) => {
                          const categoryId = e.target.value;
                          if (categoryId) {
                            addCategorySelection(categoryId);
                            e.target.value = '';
                          }
                        }}
                      >
                        <option value="">-- Select a category to add --</option>
                        {availableCategories.map((cat) => (
                          <option key={cat.id} value={cat.id}>
                            {cat.title} ({cat.question_count || 0} questions)
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {availableCategories.length === 0 && categorySelections.length === 0 && (
                  <p className="text-gray-600 dark:text-gray-400 text-center py-4">
                    No categories available in this deck
                  </p>
                )}
              </CardContent>
            </Card>
          )}

          {/* Quiz Settings */}
          <Card>
            <CardHeader>
              <CardTitle>Quiz Settings</CardTitle>
              <CardDescription>Configure quiz behavior and options</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="question-mode">Question Mode</Label>
                <select
                  id="question-mode"
                  value={questionMode}
                  onChange={(e) => setQuestionMode(e.target.value as 'ar_to_fr' | 'fr_to_ar')}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-black"
                >
                  <option value="ar_to_fr">Arabic to French</option>
                  <option value="fr_to_ar">French to Arabic</option>
                </select>
              </div>

              <div>
                <Label htmlFor="time-limit">Time Limit (seconds, 0 for no limit)</Label>
                <Input
                  id="time-limit"
                  type="number"
                  min="0"
                  value={timeLimit}
                  onChange={(e) => setTimeLimit(Number(e.target.value))}
                />
              </div>

              <div>
                <Label htmlFor="pass-percentage">Pass Percentage (%)</Label>
                <Input
                  id="pass-percentage"
                  type="number"
                  min="0"
                  max="100"
                  value={passPercentage}
                  onChange={(e) => setPassPercentage(Number(e.target.value))}
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="shuffle"
                  checked={shuffleQuestions}
                  onChange={(e) => setShuffleQuestions(e.target.checked)}
                  className="w-4 h-4"
                />
                <Label htmlFor="shuffle">Shuffle Questions</Label>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="public"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-4 h-4"
                />
                <Label htmlFor="public">Make Public (others can see and take this quiz)</Label>
              </div>
            </CardContent>
          </Card>

          {/* Submit */}
          <div className="flex gap-4 justify-end">
            <Button type="button" variant="outline" onClick={() => router.back()}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Creating...' : 'Create Quiz'}
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}
