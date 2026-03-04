'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { quizApi } from '@/lib/api/quiz';
import { Quiz } from '@/lib/types/quiz';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { CheckCircle, Trophy, Star, Gamepad2, Palette, ArrowLeft, ArrowRight, BookOpen, Plus, Sparkles } from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

export default function QuizzesPage() {
  const router = useRouter();
  const [publicQuizzes, setPublicQuizzes] = useState<Quiz[]>([]);
  const [createdQuizzes, setCreatedQuizzes] = useState<Quiz[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [listData, createdData] = await Promise.all([
        quizApi.listQuizzes().catch(() => ({ quizzes: [], total: 0, page: 1, page_size: 20 })),
        quizApi.getMyQuizzes().catch(() => []),
      ]);
      setPublicQuizzes(listData.quizzes || []);
      setCreatedQuizzes(createdData);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Échec du chargement des quiz');
    } finally {
      setLoading(false);
    }
  };

  const handleStartQuiz = (quizId: string) => {
    router.push(`/quizzes/${quizId}`);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center pt-20">
        <div className="flex flex-col items-center gap-4">
          <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-primary font-mono text-sm animate-pulse tracking-widest">CHARGEMENT DU CONTENU...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center pt-20">
        <Card className="max-w-md mx-4 bg-white border-red-200 shadow-xl">
          <CardContent className="pt-6 text-center space-y-4">
            <p className="text-red-600 font-bold text-lg">{error}</p>
            <Button onClick={loadData} size="lg" variant="outline" className="border-red-200 text-red-600 hover:bg-red-50">
              Réessayer
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground font-sans pb-20">
      <main className="container mx-auto px-4 py-8 pt-24 space-y-8">

        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <Button
              variant="ghost"
              className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent"
              onClick={() => router.push('/')}
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Retour au Tableau de Bord
            </Button>
            <h1 className="text-4xl font-black text-foreground tracking-tight">Sélection de Mission</h1>
            <p className="text-muted-foreground">Choisissez votre prochain défi ou gérez vos créations</p>
          </div>
        </div>

        {/* Content Tabs */}
        <Tabs defaultValue="available" className="w-full space-y-8">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <TabsList className="bg-muted p-1 border border-border h-auto rounded-xl">
              <TabsTrigger value="available" className="px-6 py-3 rounded-lg text-sm font-bold data-[state=active]:bg-primary data-[state=active]:text-white transition-all text-muted-foreground">
                <Gamepad2 className="w-4 h-4 mr-2" />
                Quiz Disponibles
              </TabsTrigger>
              <TabsTrigger value="created" className="px-6 py-3 rounded-lg text-sm font-bold data-[state=active]:bg-secondary data-[state=active]:text-white transition-all text-muted-foreground">
                <Palette className="w-4 h-4 mr-2" />
                Mes Créations
              </TabsTrigger>
            </TabsList>

            <Button onClick={() => router.push('/quizzes/create')} className="bg-emerald-600 hover:bg-emerald-700 text-white font-bold shadow-lg shadow-emerald-500/20">
              <Plus className="w-4 h-4 mr-2" />
              Créer Nouveau
            </Button>
          </div>

          {/* Available Quizzes Tab */}
          <TabsContent value="available" className="space-y-6 mt-0">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {publicQuizzes.map((quiz, i) => (
                <QuizCard
                  key={quiz.id}
                  quiz={quiz}
                  index={i}
                  onClick={() => handleStartQuiz(quiz.id)}
                />
              ))}
              {publicQuizzes.length === 0 && (
                <div className="col-span-full py-20 text-center border-2 border-dashed border-border rounded-3xl bg-white/50">
                  <Trophy className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-xl font-bold text-foreground">Aucun Quiz Disponible</h3>
                  <p className="text-muted-foreground mt-2">Revenez plus tard pour de nouveaux défis !</p>
                </div>
              )}
            </div>
          </TabsContent>

          {/* Created Quizzes Tab */}
          <TabsContent value="created" className="space-y-6 mt-0">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {createdQuizzes.map((quiz, i) => (
                <QuizCard
                  key={quiz.id}
                  quiz={quiz}
                  index={i}
                  onClick={() => handleStartQuiz(quiz.id)}
                  isCreator
                />
              ))}
              {createdQuizzes.length === 0 && (
                <div className="col-span-full py-20 text-center border-2 border-dashed border-border rounded-3xl bg-white/50">
                  <Palette className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-xl font-bold text-foreground">Aucun Quiz Créé</h3>
                  <p className="text-muted-foreground mt-2 mb-6">Vous n'avez pas encore créé de quiz.</p>
                  <Button onClick={() => router.push('/quizzes/create')}>Créer Votre Premier Quiz</Button>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>

      </main>
    </div>
  );
}

function QuizCard({ quiz, index, onClick, isCreator }: { quiz: Quiz, index: number, onClick: () => void, isCreator?: boolean }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ delay: index * 0.05 }}
      onClick={onClick}
      className="group relative bg-white border border-border hover:border-primary/50 rounded-3xl overflow-hidden shadow-sm hover:shadow-xl hover:shadow-primary/5 transition-all duration-300 cursor-pointer flex flex-col h-full"
    >
      <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary to-secondary transform origin-left scale-x-0 group-hover:scale-x-100 transition-transform duration-500" />

      <div className="p-6 flex-1 relative z-10">
        <div className="flex justify-between items-start mb-4">
          <div className="flex gap-2">
            <span className={cn(
              "text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded border",
              isCreator
                ? "bg-secondary/10 border-secondary/20 text-secondary"
                : "bg-muted border-border text-muted-foreground"
            )}>
              {isCreator ? 'Créateur' : (quiz.is_public ? 'Public' : 'Privé')}
            </span>
            {quiz.is_dynamic && (
              <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-1 rounded border bg-amber-500/10 border-amber-500/20 text-amber-600 flex items-center gap-1">
                <Sparkles className="w-2.5 h-2.5" /> Dynamique
              </span>
            )}
          </div>
        </div>

        <h3 className="text-xl font-bold text-foreground mb-2 group-hover:text-primary transition-colors line-clamp-2 leading-tight">
          {quiz.title}
        </h3>
        <p className="text-muted-foreground text-sm line-clamp-2 mb-4 leading-relaxed">
          {quiz.description || "Aucune description fournie."}
        </p>
      </div>

      <div className="px-6 py-4 bg-muted/20 border-t border-border flex items-center justify-between text-xs font-semibold text-muted-foreground">
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1.5 hover:text-foreground transition-colors">
            <BookOpen className="w-3.5 h-3.5" />
            {quiz.is_dynamic ? 'Variable' : `${quiz.question_mode || 'mixed'}`}
          </span>
          {quiz.coin_reward && quiz.coin_reward > 0 && (
            <span className="flex items-center gap-1.5 text-amber-500">
              <Star className="w-3.5 h-3.5" />
              {quiz.coin_reward}
            </span>
          )}
        </div>

        <div className="flex items-center group-hover:text-primary transition-colors">
          {isCreator ? 'Gérer / Jouer' : 'Démarrer'}
          <ArrowRight className="w-3.5 h-3.5 ml-1.5 transform group-hover:translate-x-1 transition-transform" />
        </div>
      </div>
    </motion.div>
  );
}
