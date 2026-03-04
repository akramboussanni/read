'use client';

import { useEffect, useState, Suspense } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Trophy, CheckCircle2, XCircle, Home,
  BookOpen, RotateCcw, Sparkles, ChevronDown,
  CheckCircle, AlertCircle, ClipboardList
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { quizApi } from '@/lib/api/quiz';

interface QuizResult {
  score: number;
  max_score: number;
  percentage: number;
  passed: boolean;
  coins_earned: number;
  next_unlocked?: number;
  attempt_id?: string;
  results: {
    question_id: string | number;
    question_text?: string;
    user_answer: string;
    correct_answer: string;
    is_correct: boolean;
    points_earned: number;
    ai_explanation?: string;
  }[];
  quiz_title?: string;
}

function QuizResultsContent() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  // Use the raw string ID — do NOT parseInt (loses precision on snowflake IDs)
  const quizId = params.id as string;

  const [results, setResults] = useState<QuizResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [showDetails, setShowDetails] = useState(false);

  // Source context: where did the quiz come from?
  const courseId = searchParams.get('courseId');
  const nodeId = searchParams.get('nodeId');
  const asgnId = searchParams.get('asgnId');
  const attemptId = searchParams.get('attemptId');
  const source = asgnId ? 'homework' : courseId ? 'course' : 'standalone';

  useEffect(() => {
    loadResults();
  }, [quizId]);

  const loadResults = async () => {
    setLoading(true);
    try {
      // 1) Try sessionStorage first (preferred flow: quiz page stored it)
      const storedKey = `quiz-results-${quizId}`;
      const stored = sessionStorage.getItem(storedKey);
      if (stored) {
        const parsed = JSON.parse(stored);
        setResults(parsed);
        // Don't remove immediately — keep it in case of refresh, clean up after delay
        setTimeout(() => sessionStorage.removeItem(storedKey), 5000);
        setLoading(false);
        return;
      }

      // 2) Fallback: if we have an attemptId in the URL, fetch from API
      const aid = attemptId || searchParams.get('attempt_id');
      if (aid) {
        const data = await quizApi.getAttempt(aid);
        if (data?.attempt) {
          const attempt = data.attempt;
          const answers = data.answers || [];
          const questions = data.questions || [];

          // Build results array from attempt data
          const resultsArr = answers.map((a: any) => {
            const q = questions.find((q: any) => String(q.id) === String(a.question_id));
            return {
              question_id: a.question_id,
              question_text: q?.question_text,
              user_answer: a.user_answer,
              correct_answer: q?.correct_answer || '',
              is_correct: a.is_correct,
              points_earned: a.points_earned,
              ai_explanation: a.ai_explanation,
            };
          });

          setResults({
            score: attempt.score ?? 0,
            max_score: attempt.max_score ?? questions.length,
            percentage: attempt.percentage ?? 0,
            passed: attempt.passed ?? false,
            coins_earned: attempt.coins_earned ?? 0,
            attempt_id: String(attempt.id),
            results: resultsArr,
          });
        }
      }
    } catch (err) {
      console.error('Failed to load results:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (results?.passed) {
      // Celebration handled via animation (no external confetti library needed)
    }
  }, [results]);

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center pt-20">
        <div className="flex flex-col items-center gap-4">
          <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-primary font-black text-sm tracking-widest animate-pulse">CHARGEMENT DES RÉSULTATS...</p>
        </div>
      </div>
    );
  }

  if (!results) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4 pt-24">
        <Card className="fun-card border-orange-200 bg-orange-50/10 p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-orange-100 rounded-2xl flex items-center justify-center mx-auto mb-4 border-b-4 border-orange-400">
            <AlertCircle className="w-8 h-8 text-orange-600" />
          </div>
          <h2 className="text-2xl font-black text-slate-800 mb-2">Aucun résultat trouvé</h2>
          <p className="text-muted-foreground font-bold mb-6">
            Termine d'abord un quiz pour voir tes résultats ici.
          </p>
          <div className="flex flex-col gap-3">
            <Button onClick={() => router.push(`/quizzes/${quizId}`)} className="w-full bg-primary text-white font-black h-12 rounded-xl border-b-4 border-primary-hover">
              <RotateCcw className="w-4 h-4 mr-2" /> Recommencer ce quiz
            </Button>
            {asgnId && (
              <Button onClick={() => router.back()} variant="outline" className="w-full font-bold h-12 rounded-xl">
                <ClipboardList className="w-4 h-4 mr-2" /> Retour aux devoirs
              </Button>
            )}
            <Button onClick={() => router.push('/')} variant="ghost" className="w-full font-bold h-10 rounded-xl">
              <Home className="w-4 h-4 mr-2" /> Tableau de bord
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  const correctCount = results.results.filter(r => r.is_correct).length;
  const totalQuestions = results.results.length;
  const pct = Math.round(results.percentage);

  const scoreColor = pct >= 80 ? 'text-green-600' : pct >= 60 ? 'text-orange-500' : 'text-red-500';
  const scoreBg = pct >= 80 ? 'from-green-400 to-emerald-600' : pct >= 60 ? 'from-orange-400 to-amber-600' : 'from-red-400 to-rose-600';

  return (
    <div className="min-h-screen bg-background pb-32 relative overflow-hidden">
      {/* Background blobs */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div className={cn("absolute -top-40 -right-20 w-96 h-96 rounded-full opacity-10 blur-3xl", results.passed ? "bg-green-500" : "bg-red-400")} />
        <div className="blob-green -bottom-40 -left-20 opacity-15" />
      </div>

      <main className="container max-w-2xl mx-auto px-4 pt-24 relative z-10 space-y-6">
        {/* Hero score card */}
        <motion.div initial={{ opacity: 0, y: -30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
          <Card className={cn("fun-card overflow-hidden border-0 shadow-2xl", results.passed ? "border-green-200" : "border-red-200")}>
            <div className={`bg-gradient-to-br ${scoreBg} p-10 text-center text-white relative`}>
              <div className="absolute inset-0 opacity-10" style={{ backgroundImage: 'radial-gradient(circle at 20% 50%, white 1px, transparent 1px), radial-gradient(circle at 80% 20%, white 1px, transparent 1px)', backgroundSize: '30px 30px' }} />

              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ delay: 0.2, type: 'spring', bounce: 0.5 }}
                className="w-24 h-24 bg-white/20 backdrop-blur rounded-3xl flex items-center justify-center mx-auto mb-6 border-4 border-white/30 shadow-xl"
              >
                {results.passed ? <Trophy className="w-12 h-12 text-white" /> : <AlertCircle className="w-12 h-12 text-white" />}
              </motion.div>

              <h1 className="text-3xl font-black mb-2 drop-shadow">
                {results.passed ? '🎉 Félicitations !' : 'Quiz Terminé !'}
              </h1>
              <p className="text-white/80 font-semibold mb-8">
                {results.quiz_title || (source === 'homework' ? 'Devoir complété' : 'Quiz terminé')}
              </p>

              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.4 }}
                className="text-8xl font-black drop-shadow-xl mb-2"
              >
                {pct}%
              </motion.div>
              <p className="text-white/80 font-bold text-lg">
                {results.score.toFixed(1)} / {results.max_score} points
              </p>

              {/* Status badge */}
              <div className={cn(
                "inline-flex items-center gap-2 mt-4 px-6 py-2 rounded-full font-black text-sm border-2",
                results.passed ? "bg-white/20 border-white/40" : "bg-white/10 border-white/20",
              )}>
                {results.passed ? <CheckCircle className="w-4 h-4" /> : <XCircle className="w-4 h-4" />}
                {results.passed ? 'RÉUSSI' : 'À REFAIRE'}
              </div>
            </div>

            {/* Stats row */}
            <CardContent className="p-0">
              <div className="grid grid-cols-3 divide-x-2 divide-border">
                <div className="p-5 text-center">
                  <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-1">Corrects</p>
                  <p className="text-3xl font-black text-green-600">{correctCount}</p>
                  <p className="text-xs text-muted-foreground font-bold">/{totalQuestions}</p>
                </div>
                <div className="p-5 text-center">
                  <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-1">Incorrects</p>
                  <p className="text-3xl font-black text-red-500">{totalQuestions - correctCount}</p>
                  <p className="text-xs text-muted-foreground font-bold">/{totalQuestions}</p>
                </div>
                <div className="p-5 text-center">
                  <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-1">Pièces</p>
                  <p className="text-3xl font-black text-amber-500">+{results.coins_earned}</p>
                  <p className="text-xs text-muted-foreground font-bold">gagnées</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Source-specific message */}
        {source === 'homework' && (
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }}>
            <Card className="fun-card border-teal-200 bg-teal-50/30">
              <CardContent className="flex items-center gap-4 p-5">
                <div className="w-12 h-12 rounded-2xl bg-teal-100 flex items-center justify-center shrink-0 border-b-4 border-teal-400">
                  <ClipboardList className="w-6 h-6 text-teal-700" />
                </div>
                <div>
                  <p className="font-black text-teal-800">Devoir {results.passed ? 'soumis !' : 'terminé'}</p>
                  <p className="text-sm text-teal-600/80 font-semibold">
                    {results.passed
                      ? 'Ton enseignant peut maintenant voir ton résultat. Continue comme ça !'
                      : 'Tu peux réessayer ce devoir pour améliorer ton score.'}
                  </p>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {source === 'course' && results.passed && (
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }}>
            <Card className="fun-card border-primary/20 bg-primary-light/30">
              <CardContent className="flex items-center gap-4 p-5">
                <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center shrink-0 border-b-4 border-primary/30">
                  <Sparkles className="w-6 h-6 text-primary" />
                </div>
                <div>
                  <p className="font-black text-primary">Progression débloquée !</p>
                  <p className="text-sm text-primary/70 font-semibold">La prochaine étape du parcours est maintenant accessible.</p>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Question details accordion */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.5 }}>
          <Card className="fun-card">
            <CardHeader
              className="cursor-pointer select-none"
              onClick={() => setShowDetails(!showDetails)}
            >
              <div className="flex items-center justify-between">
                <CardTitle className="text-xl font-black flex items-center gap-2">
                  <BookOpen className="w-5 h-5 text-primary" />
                  Détails des Réponses
                </CardTitle>
                <ChevronDown className={cn("w-5 h-5 text-muted-foreground transition-transform", showDetails && "rotate-180")} />
              </div>
            </CardHeader>
            <AnimatePresence>
              {showDetails && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                  className="overflow-hidden"
                >
                  <CardContent className="pt-0 space-y-3">
                    {results.results.map((result, idx) => (
                      <div
                        key={`${result.question_id}-${idx}`}
                        className={cn(
                          "p-4 rounded-2xl border-2",
                          result.is_correct
                            ? "border-green-200 bg-green-50/50"
                            : "border-red-200 bg-red-50/50"
                        )}
                      >
                        <div className="flex items-start justify-between gap-4 mb-2">
                          <div className="flex items-center gap-2">
                            <div className={cn(
                              "w-7 h-7 rounded-lg flex items-center justify-center font-black text-xs",
                              result.is_correct ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
                            )}>
                              {idx + 1}
                            </div>
                            {result.question_text && (
                              <p className="text-sm font-bold text-slate-700 line-clamp-2">{result.question_text}</p>
                            )}
                          </div>
                          <div className="flex items-center gap-1.5 shrink-0">
                            {result.is_correct
                              ? <CheckCircle2 className="w-5 h-5 text-green-600" />
                              : <XCircle className="w-5 h-5 text-red-500" />
                            }
                            <span className="text-xs font-black text-slate-500">{result.points_earned.toFixed(1)}pts</span>
                          </div>
                        </div>
                        <div className="space-y-1 text-sm">
                          <div className="flex items-start gap-2">
                            <span className="text-slate-500 font-semibold shrink-0">Ta réponse:</span>
                            <span className={cn("font-bold", result.is_correct ? "text-green-700" : "text-red-700")}>
                              {result.user_answer || '(sans réponse)'}
                            </span>
                          </div>
                          {!result.is_correct && result.correct_answer && (
                            <div className="flex items-start gap-2">
                              <span className="text-slate-500 font-semibold shrink-0">Bonne réponse:</span>
                              <span className="font-bold text-green-700">{result.correct_answer}</span>
                            </div>
                          )}
                          {result.ai_explanation && (
                            <div className="mt-2 p-2 bg-white/70 rounded-lg border border-border">
                              <p className="text-[11px] font-black text-muted-foreground uppercase tracking-wide mb-0.5">Explication IA</p>
                              <p className="text-xs text-slate-600 font-medium">{result.ai_explanation}</p>
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </CardContent>
                </motion.div>
              )}
            </AnimatePresence>
          </Card>
        </motion.div>

        {/* Action buttons — context-aware */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.6 }} className="space-y-3">
          {/* Primary: retry quiz (always useful) */}
          <Button
            onClick={() => {
              const url = new URL(`/quizzes/${quizId}`, window.location.origin);
              if (courseId) url.searchParams.set('courseId', courseId);
              if (nodeId) url.searchParams.set('nodeId', nodeId);
              if (asgnId) url.searchParams.set('asgnId', asgnId);
              router.push(url.pathname + url.search);
            }}
            className="w-full h-14 text-lg font-black bg-primary text-white rounded-2xl border-b-6 border-primary-hover shadow-lg"
          >
            <RotateCcw className="w-5 h-5 mr-2" /> Refaire le Quiz
          </Button>

          {/* Secondary: context-aware back button */}
          {source === 'homework' && (
            <Button
              onClick={() => router.push('/')}
              variant="outline"
              className="w-full h-12 font-bold rounded-xl border-2"
            >
              <ClipboardList className="w-4 h-4 mr-2" /> Voir mes devoirs
            </Button>
          )}

          {source === 'course' && courseId && (
            <Button
              onClick={() => router.push(`/courses/${courseId}`)}
              variant="outline"
              className="w-full h-12 font-bold rounded-xl border-2"
            >
              <BookOpen className="w-4 h-4 mr-2" /> Continuer le parcours
            </Button>
          )}

          <Button
            onClick={() => router.push('/')}
            variant="ghost"
            className="w-full h-10 font-bold rounded-xl text-muted-foreground"
          >
            <Home className="w-4 h-4 mr-2" /> Tableau de bord
          </Button>
        </motion.div>
      </main>
    </div>
  );
}

export default function QuizResultsPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
      </div>
    }>
      <QuizResultsContent />
    </Suspense>
  );
}
