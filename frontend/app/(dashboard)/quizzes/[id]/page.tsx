'use client';

import { useEffect, useState, useCallback, Suspense } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { quizApi } from '@/lib/api/quiz';
import { StartQuizResponse } from '@/lib/types/quiz';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { ArrowLeft, Clock, AlertCircle, ChevronLeft, ChevronRight, AlertTriangle } from 'lucide-react';
import { QuestionRenderer } from '@/components/quiz/question-renderer';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Shapes, Trophy, Star, Sparkles, HelpCircle, Flag, LayoutDashboard } from 'lucide-react';

function QuizTakeContent() {
  const params = useParams();
  const router = useRouter();
  const searchParamsHook = useSearchParams();
  const quizId = params.id as string;

  const [quizData, setQuizData] = useState<StartQuizResponse | null>(null);
  const [answers, setAnswers] = useState<Map<string, string>>(new Map());
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [timeRemaining, setTimeRemaining] = useState<number | null>(null);
  const [feedbackMap, setFeedbackMap] = useState<Map<string, { isCorrect: boolean, correctAnswer?: string, points_earned: number, ai_explanation?: string }>>(new Map());
  const [warningMap, setWarningMap] = useState<Map<string, string>>(new Map()); // questionId -> hint message
  const [checkingInfo, setCheckingInfo] = useState<string | null>(null);

  useEffect(() => {
    if (quizId) {
      startQuiz();
    }
  }, [quizId]);

  useEffect(() => {
    if (quizData?.time_limit && timeRemaining === null) {
      setTimeRemaining(quizData.time_limit * 60); // Convert minutes to seconds
    }

    if (timeRemaining !== null && timeRemaining > 0) {
      const timer = setInterval(() => {
        setTimeRemaining((prev) => {
          if (prev === null || prev <= 1) {
            handleSubmit(); // Auto-submit when time runs out
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => clearInterval(timer);
    }
  }, [timeRemaining, quizData]);

  const startQuiz = async (restart: boolean = false) => {
    try {
      setLoading(true);
      const searchParams = new URLSearchParams(window.location.search);
      const courseId = searchParams.get('courseId') || undefined;
      const nodeId = searchParams.get('nodeId') || undefined;
      const assignmentId = searchParams.get('asgnId') || undefined;

      const data = await quizApi.startQuiz(quizId, courseId, nodeId, assignmentId);
      setQuizData(data);
      if (restart) setFeedbackMap(new Map());

      // Load previous answers if any
      if (data.previous_answers && data.previous_answers.length > 0) {
        const prevById = new Map<string, string>();
        const prevFeedback = new Map<string, any>();
        data.previous_answers.forEach((ans: any) => {
          prevById.set(ans.question_id, ans.answer);
          if (ans.is_correct !== undefined && ans.is_correct !== null) {
            prevFeedback.set(ans.question_id, {
              isCorrect: ans.is_correct,
              correctAnswer: ans.correct_answer,
              ai_explanation: ans.ai_explanation,
              points_earned: ans.points_earned
            });
          }
        });
        setAnswers(prevById);
        setFeedbackMap(prevFeedback);

        // Find first unanswered question
        const firstUnansweredIdx = data.questions.findIndex(q => !prevById.has(q.id));
        if (firstUnansweredIdx >= 0) {
          setCurrentQuestionIdx(firstUnansweredIdx);
        } else {
          // All answered? Go to last
          setCurrentQuestionIdx(data.questions.length - 1);
        }
      } else {
        setAnswers(new Map());
        setCurrentQuestionIdx(0);
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Échec du démarrage du quiz');
    } finally {
      setLoading(false);
    }
  };

  const handleAnswerChange = (questionId: string, answer: string) => {
    if (feedbackMap.has(questionId)) return;
    // Clear any 'needs more detail' warning when the user updates their answer
    if (warningMap.has(questionId)) {
      setWarningMap(prev => { const m = new Map(prev); m.delete(questionId); return m; });
    }
    setAnswers((prev) => {
      const newAnswers = new Map(prev);
      newAnswers.set(questionId, answer);
      return newAnswers;
    });
  };

  const handleNext = () => {
    if (!quizData) return;
    if (currentQuestionIdx < quizData.questions.length - 1) {
      setCurrentQuestionIdx(prev => prev + 1);
    }
  };

  const handlePrev = () => {
    if (currentQuestionIdx > 0) {
      setCurrentQuestionIdx(prev => prev - 1);
    }
  };

  const handleCheck = async () => {
    if (!quizData) return;
    const currentQ = quizData.questions[currentQuestionIdx];
    const answer = answers.get(currentQ.id);

    if (!answer) return;

    try {
      setCheckingInfo(currentQ.id);
      const result = await quizApi.submitAnswer({
        attempt_id: quizData.attempt_id,
        question_id: currentQ.id,
        answer: answer
      });

      if (result.needs_more_detail) {
        // AI flagged as "close but vague" — don't lock the answer, just warn
        setWarningMap(prev => {
          const m = new Map(prev);
          m.set(currentQ.id, result.ai_explanation || 'Soyez plus précis !');
          return m;
        });
      } else {
        // Lock in the answer with full feedback
        setWarningMap(prev => { const m = new Map(prev); m.delete(currentQ.id); return m; });
        setFeedbackMap(prev => {
          const newMap = new Map(prev);
          newMap.set(currentQ.id, {
            isCorrect: result.is_correct,
            correctAnswer: result.correct_answer,
            points_earned: result.points_earned,
            ai_explanation: result.ai_explanation
          });
          return newMap;
        });
      }

    } catch (err) {
      console.error("Failed to check answer:", err);
    } finally {
      setCheckingInfo(null);
    }
  };

  const handleSubmit = useCallback(async () => {
    if (!quizData || submitting) return;

    // Check if all questions are answered
    const unanswered = quizData.questions.filter(
      (q) => !answers.has(q.id)
    );

    if (unanswered.length > 0 && timeRemaining !== 0) {
      if (!confirm(`Il vous reste ${unanswered.length} question(s) sans réponse. Soumettre quand même ?`)) {
        return;
      }
    }

    try {
      setSubmitting(true);

      // Submit any remaining unanswered questions first
      for (const q of quizData.questions) {
        if (!feedbackMap.has(q.id) && answers.has(q.id)) {
          try {
            await quizApi.submitAnswer({
              attempt_id: quizData.attempt_id,
              question_id: q.id,
              answer: answers.get(q.id)!,
            });
          } catch {
            // Ignore - might already be submitted
          }
        }
      }

      const result = await quizApi.completeQuiz(quizData.attempt_id);

      // Store results in sessionStorage — use string quizId (not parseInt!) to avoid
      // precision loss on 64-bit snowflake IDs
      sessionStorage.setItem(
        `quiz-results-${quizId}`,
        JSON.stringify(result)
      );

      // Navigate to results page, forwarding all context params + attemptId as fallback
      const courseId = searchParamsHook.get('courseId');
      const nodeId = searchParamsHook.get('nodeId');
      const asgnId = searchParamsHook.get('asgnId');
      const url = new URL(`/quizzes/${quizId}/results`, window.location.origin);
      url.searchParams.set('attemptId', quizData.attempt_id);
      if (courseId) url.searchParams.set('courseId', courseId);
      if (nodeId) url.searchParams.set('nodeId', nodeId);
      if (asgnId) url.searchParams.set('asgnId', asgnId);
      router.push(url.pathname + url.search);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Échec de la soumission du quiz');
      setSubmitting(false);
    }
  }, [quizData, answers, timeRemaining, submitting, quizId, router, searchParamsHook]);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center space-y-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground animate-pulse">Préparation de votre quiz...</p>
      </div>
    );
  }

  if (error || !quizData || quizData.questions.length === 0) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="max-w-md w-full text-center space-y-4 p-8 border rounded-xl bg-card shadow-lg">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
            <AlertCircle className="w-8 h-8 text-red-600" />
          </div>
          <h2 className="text-xl font-bold">Un problème est survenu</h2>
          <p className="text-muted-foreground">{error || (quizData && quizData.questions.length === 0 ? 'Ce quiz n\'a pas encore de questions.' : 'Quiz non trouvé')}</p>
          <Button onClick={() => router.push('/quizzes')} variant="outline" className="w-full">
            Retour aux Quiz
          </Button>
        </div>
      </div>
    );
  }

  const currentQuestionStr = quizData.questions[currentQuestionIdx];
  const progress = ((answers.size) / quizData.questions.length) * 100;
  const isLastQuestion = currentQuestionIdx === quizData.questions.length - 1;
  const currentFeedback = feedbackMap.get(currentQuestionStr.id);
  const currentWarning = warningMap.get(currentQuestionStr.id); // 'needs_more_detail' hint
  const isChecking = checkingInfo === currentQuestionStr.id;
  const hasAnswer = answers.has(currentQuestionStr.id);

  return (
    <div className="min-h-screen bg-background flex flex-col md:flex-row relative overflow-hidden">
      {/* Decorative Blobs */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div className="blob-green -top-20 -left-20 opacity-20" />
        <div className="blob-orange -bottom-20 -right-20 opacity-20" />
        <div className="blob-teal top-1/2 left-1/2 -translate-x-1/2 opacity-10" />
      </div>

      {/* Sidebar Navigation */}
      <aside className="hidden md:flex flex-col w-72 border-r-4 border-border bg-white h-screen sticky top-0 px-6 py-10 z-10 shadow-xl overflow-hidden">
        <div className="mb-10 text-center relative">
          <div className="w-16 h-16 bg-primary text-white rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg border-b-4 border-primary-hover animate-float">
            <Shapes className="h-8 w-8" />
          </div>
          <h1 className="font-black text-xl line-clamp-2 mb-2 tracking-tight">{quizData.title}</h1>
          <div className="inline-flex items-center gap-2 text-xs font-black uppercase tracking-widest text-muted-foreground bg-muted/50 px-3 py-1.5 rounded-full border border-border">
            <Clock className="w-3.5 h-3.5" />
            {timeRemaining !== null ? (
              <span className={cn(timeRemaining < 60 && "text-red-500 animate-pulse")}>
                {formatTime(timeRemaining)}
              </span>
            ) : "Pas de limite"}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar space-y-8">
          <div className="space-y-3">
            <div className="flex justify-between items-end px-1">
              <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Progression</span>
              <span className="text-sm font-black text-primary">{Math.round(progress)}%</span>
            </div>
            <div className="h-3 w-full bg-muted rounded-full p-0.5 border-2 border-border shadow-inner">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${progress}%` }}
                className="h-full bg-primary rounded-full shadow-sm"
              />
            </div>
          </div>

          <div>
            <h3 className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-4 px-1 flex items-center gap-2">
              <LayoutDashboard className="w-3 h-3" /> Questions
            </h3>
            <div className="grid grid-cols-5 gap-2 pb-6">
              {quizData.questions.map((q, idx) => {
                const isAnswered = answers.has(q.id);
                const isCurrent = idx === currentQuestionIdx;
                const qFeedback = feedbackMap.get(q.id);

                return (
                  <button
                    key={q.id}
                    onClick={() => setCurrentQuestionIdx(idx)}
                    className={cn(
                      "aspect-square rounded-xl flex items-center justify-center text-xs font-black transition-all border-2 border-b-4 active:border-b-0 active:translate-y-1 shadow-sm",
                      isCurrent
                        ? "bg-primary text-white border-primary-hover scale-110 shadow-primary/20 z-10"
                        : qFeedback
                          ? qFeedback.isCorrect
                            ? "bg-green-100 text-green-700 border-green-300"
                            : "bg-red-100 text-red-700 border-red-300"
                          : isAnswered
                            ? "bg-blue-100 text-blue-700 border-blue-300"
                            : "bg-white text-slate-400 border-border"
                    )}
                  >
                    {idx + 1}
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        <div className="mt-auto pt-6 border-t-2 border-border space-y-3">
          <Button
            variant="ghost"
            className="w-full text-xs font-black text-muted-foreground hover:text-primary transition-all rounded-xl h-10"
            onClick={() => {
              if (confirm('Recommencer le quiz ? Votre progression actuelle sera perdue.')) {
                startQuiz(true);
              }
            }}
          >
            Recommencer
          </Button>
          <Button
            variant="ghost"
            className="w-full text-xs font-black text-muted-foreground hover:text-red-500 transition-all rounded-xl h-10"
            onClick={() => router.push('/quizzes')}
          >
            Arrêter & Sauver
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col h-screen overflow-hidden">
        {/* Mobile Header (visible only on small screens) */}
        <header className="md:hidden flex items-center justify-between p-4 border-b bg-background z-10">
          <Button size="icon" variant="ghost" onClick={() => router.push('/quizzes')}>
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div className="font-semibold text-base truncate max-w-[150px]">{quizData.title}</div>
          {timeRemaining !== null && (
            <div className={cn("font-mono font-bold", timeRemaining < 60 ? "text-red-500" : "")}>
              {formatTime(timeRemaining)}
            </div>
          )}
        </header>

        {/* Question Area */}
        <div className="flex-1 overflow-y-auto p-4 md:p-8 lg:p-12 flex flex-col max-w-5xl mx-auto w-full relative z-10">
          <div className="flex-1 flex flex-col justify-center min-h-[450px]">
            <AnimatePresence mode="wait">
              <motion.div
                key={currentQuestionIdx}
                initial={{ opacity: 0, y: 30, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -30, scale: 0.95 }}
                transition={{ type: "spring", bounce: 0.4 }}
                className="w-full"
              >
                <Card className="fun-card border-primary/20 bg-white shadow-2xl overflow-visible">
                  <CardHeader className="text-center relative pt-8">
                    <div className="absolute -top-10 left-1/2 -translate-x-1/2 w-20 h-20 bg-accent rounded-3xl flex items-center justify-center text-white shadow-xl rotate-3 border-b-6 border-teal-600 animate-float">
                      <Sparkles className="w-10 h-10" />
                    </div>
                    <CardTitle className="text-3xl font-black mt-6">Question {currentQuestionIdx + 1}</CardTitle>
                    <div className="flex items-center justify-center gap-2 text-[10px] font-black uppercase text-muted-foreground tracking-widest bg-muted/50 w-fit mx-auto px-4 py-1.5 rounded-full border border-border">
                      <HelpCircle className="w-3.5 h-3.5" /> {quizData.questions.length - currentQuestionIdx} restantes
                    </div>
                  </CardHeader>
                  <CardContent className="p-8 md:p-12 space-y-4">
                    {/* Warning banner: AI says 'close but vague' */}
                    {currentWarning && !currentFeedback && (
                      <motion.div
                        initial={{ opacity: 0, y: -8 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="flex items-start gap-3 bg-amber-50 border-2 border-amber-300 rounded-2xl p-4"
                      >
                        <div className="w-8 h-8 rounded-xl bg-amber-200 flex items-center justify-center shrink-0">
                          <AlertTriangle className="w-4 h-4 text-amber-700" />
                        </div>
                        <div>
                          <p className="font-black text-amber-800 text-sm">Presque ! Soyez plus précis</p>
                          <p className="text-amber-700/80 text-xs font-semibold mt-0.5">{currentWarning}</p>
                        </div>
                      </motion.div>
                    )}
                    <QuestionRenderer
                      question={currentQuestionStr}
                      answer={answers.get(currentQuestionStr.id)}
                      onAnswerChange={(val) => handleAnswerChange(currentQuestionStr.id, val)}
                      feedback={currentFeedback}
                      onSubmit={() => {
                        if (currentFeedback) {
                          if (isLastQuestion) handleSubmit();
                          else handleNext();
                        } else if (hasAnswer) {
                          handleCheck();
                        }
                      }}
                    />
                  </CardContent>
                </Card>
              </motion.div>
            </AnimatePresence>
          </div>

          {/* Bottom Navigation */}
          <div className="mt-12 flex items-center justify-between gap-6 py-6 sticky bottom-0 md:bg-transparent">
            <Button
              variant="ghost"
              onClick={handlePrev}
              disabled={currentQuestionIdx === 0}
              className="w-32 h-14 font-black rounded-2xl border-2 border-border hover:bg-muted text-slate-500 transition-all active:scale-95"
            >
              <ChevronLeft className="w-5 h-5 mr-1" /> PRÉCÉDENT
            </Button>

            <div className="flex items-center gap-4">
              {isLastQuestion && currentFeedback && (
                <Button
                  onClick={() => router.push('/quizzes')}
                  variant="ghost"
                  className="font-black text-muted-foreground hover:text-foreground"
                >
                  <ChevronLeft className="w-4 h-4 mr-1" /> Quitter
                </Button>
              )}

              {!currentFeedback ? (
                <Button
                  onClick={handleCheck}
                  disabled={!hasAnswer || isChecking}
                  className="w-48 h-14 text-xl font-black bg-primary text-white rounded-2xl border-b-6 border-primary-hover shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all"
                >
                  {isChecking ? (
                    <span className="flex items-center gap-2">
                      <div className="w-5 h-5 border-3 border-white border-t-transparent rounded-full animate-spin" />
                      VÉRIFIE...
                    </span>
                  ) : "VÉRIFIER !"}
                </Button>
              ) : (
                isLastQuestion ? (
                  <Button
                    onClick={handleSubmit}
                    disabled={submitting}
                    className="w-56 h-14 text-xl font-black bg-secondary text-white rounded-2xl border-b-6 border-orange-600 shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all flex items-center justify-center gap-3"
                  >
                    {submitting ? 'FINITION...' : <>TERMINER ! <Trophy className="w-6 h-6" /></>}
                  </Button>
                ) : (
                  <Button
                    onClick={handleNext}
                    className="w-56 h-14 text-xl font-black bg-primary text-white rounded-2xl border-b-6 border-primary-hover shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all flex items-center justify-center gap-2"
                  >
                    SUIVANT <ChevronRight className="w-6 h-6" />
                  </Button>
                )
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default function QuizTakePage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-background flex flex-col items-center justify-center space-y-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground animate-pulse">Préparation de votre quiz...</p>
      </div>
    }>
      <QuizTakeContent />
    </Suspense>
  );
}

