'use client';

import { QuizEditor } from '@/components/quiz/quiz-editor';
import { ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';

export default function CreateQuizPage() {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-background text-foreground pb-20">
      <main className="container mx-auto px-4 py-8 pt-24 space-y-8">

        {/* Header */}
        <div className="space-y-1">
          <Button
            variant="ghost"
            className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent"
            onClick={() => router.push('/quizzes')}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Retour aux Quiz
          </Button>
          <h1 className="text-3xl font-black text-foreground tracking-tight">Créer un Quiz</h1>
          <p className="text-muted-foreground">Créez un quiz manuel ou automatique en quelques étapes</p>
        </div>

        {/* Quiz Editor */}
        <QuizEditor />
      </main>
    </div>
  );
}
