'use client';

import { useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Plus, Trash2, CheckCircle2, Eye, Sparkles, Loader2 } from 'lucide-react';
import { QuestionRenderer } from '@/components/quiz/question-renderer';
import type { ManualQuestionRequest, DeckSelectionRequest } from '@/lib/types/quiz';
import { VisualDeckSelector } from '@/components/quiz/visual-deck-selector';
import { quizApi } from '@/lib/api/quiz';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';

interface ManualQuestionBuilderProps {
    questions: ManualQuestionRequest[];
    onChange: (questions: ManualQuestionRequest[]) => void;
}

export function ManualQuestionBuilder({
    questions,
    onChange,
}: ManualQuestionBuilderProps) {
    const [activeTab, setActiveTab] = useState<'list' | 'editor' | 'generate'>('list');
    const [editingIndex, setEditingIndex] = useState<number | null>(null);

    // Editor State
    const [questionText, setQuestionText] = useState('');
    const [correctAnswer, setCorrectAnswer] = useState('');
    const [wrongAnswers, setWrongAnswers] = useState<string[]>(['', '', '']);
    const [questionType, setQuestionType] = useState<'mcq' | 'translate'>('mcq');

    // Generator State
    const [isGenerating, setIsGenerating] = useState(false);
    const [genDeckSelections, setGenDeckSelections] = useState<DeckSelectionRequest[]>([]);
    const [genQuestionType, setGenQuestionType] = useState<'mcq' | 'translate'>('mcq');
    const [genDirection, setGenDirection] = useState<'source_to_target' | 'target_to_source'>('source_to_target');

    const resetEditor = () => {
        setQuestionText('');
        setCorrectAnswer('');
        setWrongAnswers(['', '', '']);
        setQuestionType('mcq');
        setEditingIndex(null);
        setActiveTab('editor');
    };

    const handleEdit = (index: number) => {
        const q = questions[index];
        setQuestionText(q.question_text);
        setCorrectAnswer(q.correct_answer);

        if (q.options) {
            const wrongs = q.options.filter(o => o !== q.correct_answer);
            while (wrongs.length < 3) wrongs.push('');
            setWrongAnswers(wrongs.slice(0, 3));
        } else {
            setWrongAnswers(['', '', '']);
        }

        setQuestionType(q.question_type as any);
        setEditingIndex(index);
        setActiveTab('editor');
    };

    const handleSave = () => {
        if (!questionText.trim() || !correctAnswer.trim()) return;

        const wrongs = wrongAnswers.filter(w => w.trim());
        const options = [...wrongs, correctAnswer].sort(() => Math.random() - 0.5);

        const newQuestion: ManualQuestionRequest = {
            question_text: questionText,
            correct_answer: correctAnswer,
            options: options,
            question_type: questionType,
            direction: 'source_to_target'
        };

        if (editingIndex !== null) {
            const updated = [...questions];
            updated[editingIndex] = newQuestion;
            onChange(updated);
        } else {
            onChange([...questions, newQuestion]);
        }

        resetEditor();
        setEditingIndex(null);
        setActiveTab('list');
    };

    const handleDelete = (index: number) => {
        onChange(questions.filter((_, i) => i !== index));
        if (editingIndex === index) {
            resetEditor();
            setEditingIndex(null);
            // Stay in list tab if deleting from list
        }
    };

    const handleGenerate = async () => {
        if (genDeckSelections.length === 0) return;
        setIsGenerating(true);
        try {
            const result = await quizApi.generatePreview({
                deck_selections: genDeckSelections,
                question_types: [genQuestionType],
                directions: [genDirection],
                question_count: 1,
            });

            if (result && result.length > 0) {
                const q = result[0];
                setQuestionText(q.question_text);
                setCorrectAnswer(q.correct_answer);

                if (q.options) {
                    const wrongs = q.options.filter(o => o !== q.correct_answer);
                    while (wrongs.length < 3) wrongs.push('');
                    setWrongAnswers(wrongs.slice(0, 3));
                } else {
                    setWrongAnswers(['', '', '']);
                }

                setQuestionType(q.question_type as any);
                setEditingIndex(null);
                setActiveTab('editor');
            }
        } catch (error) {
            console.error("Failed to generate question:", error);
        } finally {
            setIsGenerating(false);
        }
    };

    // Live Preview Data
    const previewData = useMemo(() => {
        const wrongs = wrongAnswers.filter(w => w.trim());
        const optionsList = [...wrongs];
        if (correctAnswer.trim()) optionsList.push(correctAnswer);

        const options = optionsList.map((opt, i) => ({
            id: i.toString(),
            option_text: opt
        }));

        return {
            id: "0",
            question_text: questionText || "Your Question Here",
            question_type: questionType,
            points: 10,
            options: options
        };
    }, [questionText, correctAnswer, wrongAnswers, questionType]);

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold">Questions ({questions.length})</h3>
                <div className="flex gap-2">
                    <Button
                        onClick={() => setActiveTab('generate')}
                        variant={activeTab === 'generate' ? "default" : "outline"}
                        className="gap-2"
                    >
                        <Sparkles className="w-4 h-4" />
                        Auto Generate
                    </Button>
                    <Button
                        onClick={resetEditor}
                        variant={activeTab === 'editor' && editingIndex === null ? "default" : "outline"}
                    >
                        <Plus className="w-4 h-4 mr-2" />
                        Add Manually
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 h-[800px]">
                {/* Question List Column */}
                <div className="lg:col-span-4 space-y-3 h-full overflow-y-auto pr-2 custom-scrollbar border-r">
                    <AnimatePresence>
                        {questions.map((q, idx) => (
                            <motion.div
                                key={idx}
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                exit={{ opacity: 0, x: -20 }}
                                onClick={() => handleEdit(idx)}
                                className={cn(
                                    "p-4 rounded-lg border cursor-pointer hover:bg-muted/50 transition-colors relative group",
                                    editingIndex === idx && activeTab === 'editor' ? "border-primary bg-primary/5 shadow-sm" : "bg-card"
                                )}
                            >
                                <div className="flex justify-between items-start">
                                    <h4 className="font-medium text-sm line-clamp-2 mb-2">{q.question_text}</h4>
                                    <Button
                                        size="icon"
                                        variant="ghost"
                                        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10 -mt-1 -mr-1"
                                        onClick={(e) => { e.stopPropagation(); handleDelete(idx); }}
                                    >
                                        <Trash2 className="w-3 h-3" />
                                    </Button>
                                </div>
                                <div className="flex gap-2">
                                    <div className="text-xs bg-muted px-2 py-0.5 rounded text-muted-foreground uppercase">{q.question_type}</div>
                                    <div className="text-xs px-2 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 truncate max-w-[120px]">
                                        {q.correct_answer}
                                    </div>
                                </div>
                            </motion.div>
                        ))}
                        {questions.length === 0 && (
                            <div className="text-center py-12 border-2 border-dashed rounded-xl text-muted-foreground">
                                <p>No questions added yet.</p>
                            </div>
                        )}
                    </AnimatePresence>
                </div>

                {/* Editor / Generator Column */}
                <div className="lg:col-span-8 flex flex-col gap-6 h-full overflow-y-auto pl-2">

                    {activeTab === 'generate' ? (
                        <div className="space-y-6">
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2">
                                        <Sparkles className="w-5 h-5 text-primary" />
                                        Generate Question
                                    </CardTitle>
                                    <CardDescription>
                                        Automatically create a question from your content.
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-6">
                                    <VisualDeckSelector
                                        deckSelections={genDeckSelections}
                                        onDeckSelectionsChange={setGenDeckSelections}
                                    />

                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <Label>Question Type</Label>
                                            <select
                                                className="w-full px-3 py-2 border rounded-md"
                                                value={genQuestionType}
                                                onChange={(e) => setGenQuestionType(e.target.value as any)}
                                            >
                                                <option value="mcq">Multiple Choice</option>
                                                <option value="translate">Translate</option>
                                            </select>
                                        </div>
                                        <div className="space-y-2">
                                            <Label>Direction</Label>
                                            <select
                                                className="w-full px-3 py-2 border rounded-md"
                                                value={genDirection}
                                                onChange={(e) => setGenDirection(e.target.value as any)}
                                            >
                                                <option value="source_to_target">Source to Target</option>
                                                <option value="target_to_source">Target to Source</option>
                                            </select>
                                        </div>
                                    </div>

                                    <Button
                                        className="w-full"
                                        onClick={handleGenerate}
                                        disabled={isGenerating || genDeckSelections.length === 0}
                                    >
                                        {isGenerating ? (
                                            <>
                                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                                Generating...
                                            </>
                                        ) : (
                                            <>
                                                <Sparkles className="w-4 h-4 mr-2" />
                                                Generate Question
                                            </>
                                        )}
                                    </Button>
                                </CardContent>
                            </Card>
                        </div>
                    ) : (
                        <>
                            {/* Live Preview Section */}
                            <div className="space-y-2">
                                <Label className="flex items-center gap-2 text-primary font-semibold">
                                    <Eye className="w-4 h-4" />
                                    Live Preview
                                </Label>
                                <div className="pointer-events-none opacity-90 scale-95 origin-top-left">
                                    <QuestionRenderer
                                        question={previewData}
                                        answer={undefined}
                                        onAnswerChange={() => { }}
                                        onSubmit={() => { }}
                                    />
                                </div>
                            </div>

                            <Card>
                                <CardHeader>
                                    <CardTitle>{editingIndex !== null ? 'Edit Question' : 'New Question'}</CardTitle>
                                    <CardDescription>Configure the question details.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="space-y-2">
                                        <Label>Question Type</Label>
                                        <div className="flex gap-2">
                                            <Button
                                                type="button"
                                                variant={questionType === 'mcq' ? 'default' : 'outline'}
                                                onClick={() => setQuestionType('mcq')}
                                                className="flex-1"
                                            >
                                                Multiple Choice
                                            </Button>

                                            <Button
                                                type="button"
                                                variant={questionType === 'translate' ? 'default' : 'outline'}
                                                onClick={() => setQuestionType('translate')}
                                                className="flex-1"
                                            >
                                                Translate
                                            </Button>
                                        </div>
                                    </div>

                                    <div className="space-y-2">
                                        <Label>Question Text</Label>
                                        <Input
                                            value={questionText}
                                            onChange={(e) => setQuestionText(e.target.value)}
                                            placeholder="e.g., What is the capital of France?"
                                            className="font-medium"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label className="flex items-center gap-2">
                                            Correct Answer
                                            <CheckCircle2 className="w-3 h-3 text-green-500" />
                                        </Label>
                                        <Input
                                            value={correctAnswer}
                                            onChange={(e) => setCorrectAnswer(e.target.value)}
                                            placeholder="e.g., Paris"
                                            className="border-green-500/50 focus-visible:ring-green-500"
                                        />
                                    </div>

                                    {questionType === 'mcq' && (
                                        <div className="space-y-3 pt-2">
                                            <Label>Distractors (Wrong Answers)</Label>
                                            {wrongAnswers.map((ans, i) => (
                                                <Input
                                                    key={i}
                                                    value={ans}
                                                    onChange={(e) => {
                                                        const newWrong = [...wrongAnswers];
                                                        newWrong[i] = e.target.value;
                                                        setWrongAnswers(newWrong);
                                                    }}
                                                    placeholder={`Wrong option ${i + 1}`}
                                                />
                                            ))}
                                        </div>
                                    )}

                                    <div className="pt-4 flex gap-3 justify-end border-t mt-4">
                                        {/* Cancel button logic - return to list if adding new, or just reset editor */}
                                        <Button
                                            variant="ghost"
                                            onClick={() => {
                                                setEditingIndex(null);
                                                resetEditor();
                                                setActiveTab('list');
                                            }}
                                        >
                                            Cancel
                                        </Button>
                                        <Button onClick={handleSave} disabled={!questionText || !correctAnswer}>
                                            {editingIndex !== null ? 'Update Question' : 'Add Question'}
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
