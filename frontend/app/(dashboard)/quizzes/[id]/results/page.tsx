'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Trophy, Coins, CheckCircle, XCircle, ArrowRight, Home } from 'lucide-react';

// Store results in sessionStorage to pass between pages
interface QuizResults {
  score: number;
  max_score: number;
  percentage: number;
  passed: boolean;
  coins_earned: number;
  next_unlocked?: number;
  results: {
    question_id: number;
    user_answer: string;
    correct_answer: string;
    is_correct: boolean;
    points_earned: number;
  }[];
  quiz_title?: string;
}

export default function QuizResultsPage() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const quizId = parseInt(params.id as string);
  const [results, setResults] = useState<QuizResults | null>(null);
  const [showDetails, setShowDetails] = useState(false);

  useEffect(() => {
    // Try to get results from sessionStorage
    const storedResults = sessionStorage.getItem(`quiz-results-${quizId}`);
    if (storedResults) {
      setResults(JSON.parse(storedResults));
      // Clean up
      sessionStorage.removeItem(`quiz-results-${quizId}`);
    }
  }, [quizId]);

  // This will be called from the quiz page after submission
  if (!results) {
    return (
      <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            No results found. Please complete a quiz first.
          </p>
          <Button onClick={() => router.push('/quizzes')}>
            Back to Quizzes
          </Button>
        </div>
      </div>
    );
  }

  const correctCount = results.results.filter((r) => r.is_correct).length;
  const totalQuestions = results.results.length;

  return (
    <div className="min-h-screen bg-white dark:bg-black">
      <main className="container mx-auto px-4 py-8">
        {/* Results Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 mb-4">
            <Trophy className="w-10 h-10 text-white" />
          </div>
          <h1 className="text-4xl font-bold mb-2">
            {results.passed ? 'Congratulations!' : 'Quiz Complete!'}
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-400">
            {results.quiz_title || 'Quiz Results'}
          </p>
        </div>

        {/* Score Card */}
        <Card className="mb-6 border-2">
          <CardContent className="pt-6">
            <div className="text-center mb-6">
              <div className="text-6xl font-bold mb-2">
                {results.percentage.toFixed(1)}%
              </div>
              <p className="text-gray-600 dark:text-gray-400">
                {results.score.toFixed(1)} / {results.max_score} points
              </p>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <CheckCircle className="w-6 h-6 text-green-500 mx-auto mb-2" />
                <p className="text-sm text-gray-600 dark:text-gray-400">Correct</p>
                <p className="text-2xl font-bold">{correctCount}</p>
              </div>

              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <XCircle className="w-6 h-6 text-red-500 mx-auto mb-2" />
                <p className="text-sm text-gray-600 dark:text-gray-400">Incorrect</p>
                <p className="text-2xl font-bold">{totalQuestions - correctCount}</p>
              </div>

              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <Coins className="w-6 h-6 text-amber-500 mx-auto mb-2" />
                <p className="text-sm text-gray-600 dark:text-gray-400">Coins Earned</p>
                <p className="text-2xl font-bold">{results.coins_earned}</p>
              </div>

              <div className="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                <Trophy
                  className={`w-6 h-6 mx-auto mb-2 ${
                    results.passed ? 'text-yellow-500' : 'text-gray-400'
                  }`}
                />
                <p className="text-sm text-gray-600 dark:text-gray-400">Status</p>
                <p className="text-xl font-bold">
                  {results.passed ? 'Passed' : 'Not Passed'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Next Quiz Unlocked */}
        {results.next_unlocked && (
          <Card className="mb-6 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-950 dark:to-purple-950 border-2 border-blue-200 dark:border-blue-800">
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-full bg-blue-500 flex items-center justify-center">
                    <CheckCircle className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <p className="font-semibold text-lg">New Quiz Unlocked!</p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      You've unlocked a new challenge
                    </p>
                  </div>
                </div>
                <Button
                  onClick={() => router.push(`/quizzes/${results.next_unlocked}`)}
                  className="bg-blue-600 hover:bg-blue-700"
                >
                  Try Next Quiz
                  <ArrowRight className="w-4 h-4 ml-2" />
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Detailed Results */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Answer Details</CardTitle>
              <Button
                variant="outline"
                onClick={() => setShowDetails(!showDetails)}
              >
                {showDetails ? 'Hide' : 'Show'} Details
              </Button>
            </div>
          </CardHeader>
          {showDetails && (
            <CardContent>
              <div className="space-y-4">
                {results.results.map((result, idx) => (
                  <div
                    key={result.question_id}
                    className={`p-4 rounded-lg border-2 ${
                      result.is_correct
                        ? 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950'
                        : 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h4 className="font-semibold">Question {idx + 1}</h4>
                      <div className="flex items-center gap-2">
                        {result.is_correct ? (
                          <CheckCircle className="w-5 h-5 text-green-600" />
                        ) : (
                          <XCircle className="w-5 h-5 text-red-600" />
                        )}
                        <span className="text-sm font-medium">
                          {result.points_earned.toFixed(1)} pts
                        </span>
                      </div>
                    </div>

                    <div className="space-y-2 text-sm">
                      <div>
                        <span className="text-gray-600 dark:text-gray-400">
                          Your Answer:{' '}
                        </span>
                        <span
                          className={
                            result.is_correct
                              ? 'text-green-700 dark:text-green-300 font-medium'
                              : 'text-red-700 dark:text-red-300 font-medium'
                          }
                        >
                          {result.user_answer || '(Not answered)'}
                        </span>
                      </div>
                      {!result.is_correct && (
                        <div>
                          <span className="text-gray-600 dark:text-gray-400">
                            Correct Answer:{' '}
                          </span>
                          <span className="text-green-700 dark:text-green-300 font-medium">
                            {result.correct_answer}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          )}
        </Card>

        {/* Actions */}
        <div className="flex gap-4 justify-center">
          <Button variant="outline" onClick={() => router.push('/quizzes')}>
            <Home className="w-4 h-4 mr-2" />
            Back to Quizzes
          </Button>
          <Button onClick={() => router.push(`/quizzes/${quizId}`)}>
            Retake Quiz
          </Button>
        </div>
      </main>
    </div>
  );
}
