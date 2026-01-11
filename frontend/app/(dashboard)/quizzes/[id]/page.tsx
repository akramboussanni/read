'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { progressionApi } from '@/lib/api/progression';
import { StartQuizResponse, SubmitAnswer } from '@/lib/types/quiz';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ArrowLeft, Clock, AlertCircle } from 'lucide-react';

export default function QuizTakePage() {
  const params = useParams();
  const router = useRouter();
  const quizId = parseInt(params.id as string);

  const [quizData, setQuizData] = useState<StartQuizResponse | null>(null);
  const [answers, setAnswers] = useState<Map<number, string>>(new Map());
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [timeRemaining, setTimeRemaining] = useState<number | null>(null);

  useEffect(() => {
    startQuiz();
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

  const startQuiz = async () => {
    try {
      setLoading(true);
      const data = await progressionApi.startQuiz(quizId);
      setQuizData(data);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to start quiz');
    } finally {
      setLoading(false);
    }
  };

  const handleAnswerChange = (questionId: number, answer: string) => {
    setAnswers((prev) => {
      const newAnswers = new Map(prev);
      newAnswers.set(questionId, answer);
      return newAnswers;
    });
  };

  const handleSubmit = async () => {
    if (!quizData) return;

    // Check if all questions are answered
    const unanswered = quizData.questions.filter(
      (q) => !answers.has(q.id)
    );
    
    if (unanswered.length > 0 && timeRemaining !== 0) {
      if (!confirm(`You have ${unanswered.length} unanswered question(s). Submit anyway?`)) {
        return;
      }
    }

    try {
      setSubmitting(true);
      
      const submitData: SubmitAnswer[] = Array.from(answers.entries()).map(
        ([question_id, answer]) => ({
          question_id,
          answer,
        })
      );

      const result = await progressionApi.submitQuiz(quizId, {
        attempt_id: quizData.attempt_id,
        answers: submitData,
      });

      // Store results in sessionStorage to pass to results page
      sessionStorage.setItem(
        `quiz-results-${quizId}`,
        JSON.stringify({
          ...result,
          quiz_title: quizData.title,
        })
      );

      // Navigate to results page
      router.push(`/quizzes/${quizId}/results`);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to submit quiz');
      setSubmitting(false);
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <p className="text-gray-600 dark:text-gray-400">Loading quiz...</p>
      </div>
    );
  }

  if (error || !quizData) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-red-600 dark:text-red-400 mb-4">{error || 'Quiz not found'}</p>
          <Button onClick={() => router.push('/quizzes')}>
            Back to Quizzes
          </Button>
        </div>
      </div>
    );
  }

  const question = quizData.questions[currentQuestion];
  const progress = ((currentQuestion + 1) / quizData.questions.length) * 100;

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-6 flex items-center justify-between">
          <Button
            variant="outline"
            onClick={() => {
              if (confirm('Are you sure you want to quit? Your progress will be lost.')) {
                router.push('/quizzes');
              }
            }}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Quit
          </Button>
          
          {timeRemaining !== null && (
            <div className="flex items-center gap-2 text-lg font-semibold">
              <Clock className="w-5 h-5" />
              <span className={timeRemaining < 60 ? 'text-red-500' : ''}>
                {formatTime(timeRemaining)}
              </span>
            </div>
          )}
        </div>

        {/* Quiz Title */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>{quizData.title}</CardTitle>
            <CardDescription>
              Question {currentQuestion + 1} of {quizData.questions.length}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all"
                style={{ width: `${progress}%` }}
              />
            </div>
          </CardContent>
        </Card>

        {/* Question Card */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-xl">
              {question.question_text}
            </CardTitle>
            <CardDescription>
              Type: {question.question_type} • Points: {question.points}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {question.question_type === 'mcq' && question.options ? (
              <div className="space-y-3">
                {question.options.map((option) => (
                  <Label
                    key={option.id}
                    className="flex items-center p-4 border rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                  >
                    <input
                      type="radio"
                      name={`question-${question.id}`}
                      value={option.id.toString()}
                      checked={answers.get(question.id) === option.id.toString()}
                      onChange={(e) =>
                        handleAnswerChange(question.id, e.target.value)
                      }
                      className="mr-3"
                    />
                    <span>{option.option_text}</span>
                  </Label>
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                <Label htmlFor="answer">Your Answer</Label>
                <Input
                  id="answer"
                  type="text"
                  value={answers.get(question.id) || ''}
                  onChange={(e) =>
                    handleAnswerChange(question.id, e.target.value)
                  }
                  placeholder="Type your answer here..."
                  className="text-lg"
                />
              </div>
            )}
          </CardContent>
        </Card>

        {/* Navigation */}
        <div className="flex items-center justify-between">
          <Button
            variant="outline"
            onClick={() => setCurrentQuestion((prev) => Math.max(0, prev - 1))}
            disabled={currentQuestion === 0}
          >
            Previous
          </Button>

          <div className="text-sm text-gray-600 dark:text-gray-400">
            {answers.size} of {quizData.questions.length} answered
          </div>

          {currentQuestion < quizData.questions.length - 1 ? (
            <Button
              onClick={() =>
                setCurrentQuestion((prev) =>
                  Math.min(quizData.questions.length - 1, prev + 1)
                )
              }
            >
              Next
            </Button>
          ) : (
            <Button
              onClick={handleSubmit}
              disabled={submitting}
              className="bg-green-600 hover:bg-green-700"
            >
              {submitting ? 'Submitting...' : 'Submit Quiz'}
            </Button>
          )}
        </div>

        {/* Question Navigation */}
        <div className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Question Navigation</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-2">
                {quizData.questions.map((q, idx) => (
                  <Button
                    key={q.id}
                    variant={idx === currentQuestion ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setCurrentQuestion(idx)}
                    className={`w-12 h-12 ${
                      answers.has(q.id)
                        ? 'border-green-500 dark:border-green-400'
                        : ''
                    }`}
                  >
                    {idx + 1}
                  </Button>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  );
}
