'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Joystick } from 'lucide-react';
import { quizApi } from '@/lib/api/quiz';
import { QuizEditor } from '@/components/quiz/quiz-editor';
import type { Quiz } from '@/lib/types/quiz';

export default function EditQuizPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [quiz, setQuiz] = useState<(Quiz & { templates?: any[] }) | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadQuiz();
  }, [id]);

  const loadQuiz = async () => {
    try {
      setLoading(true);
      const q = await quizApi.getQuiz(id);
      setQuiz(q);
    } catch (e: any) {
      setError(e.message || 'Failed to load quiz');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
    </div>
  );

  if (error || !quiz) return (
    <div className="container mx-auto p-8 text-center">
      <p className="text-red-500">{error || 'Quiz not found'}</p>
      <Button variant="outline" onClick={() => router.back()} className="mt-4">Retour</Button>
    </div>
  );

  return (
    <div className="min-h-screen bg-background text-foreground pb-20">
      <main className="container mx-auto px-4 py-8 pt-24 space-y-8">
        <div className="space-y-1">
          <Button
            variant="ghost"
            className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent"
            onClick={() => router.back()}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Retour
          </Button>
          <h1 className="text-3xl font-black flex items-center gap-3">
            <Joystick className="w-8 h-8 text-[#6C5CE7]" />
            Éditeur de Quiz
          </h1>
          <p className="text-muted-foreground">Modifiez la configuration du quiz</p>
        </div>

        <QuizEditor
          existingQuiz={quiz}
          onSave={() => router.back()}
        />
      </main>
    </div>
  );
}
