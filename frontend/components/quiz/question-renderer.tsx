'use client';

import { motion, AnimatePresence } from 'framer-motion';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import { Check, X, Type, Globe, List, Brain, Tag, FlaskConical, Info, Zap } from 'lucide-react';
import type { QuestionWithOptions } from '@/lib/types/quiz';

interface QuestionRendererProps {
    question: QuestionWithOptions;
    answer: string | undefined;
    onAnswerChange: (answer: string) => void;
    onSubmit: () => void;
    feedback?: {
        isCorrect: boolean;
        correctAnswer?: string;
        ai_explanation?: string;
    };
}

export function QuestionRenderer({
    question,
    answer,
    onAnswerChange,
    onSubmit,
    feedback,
}: QuestionRendererProps) {

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            onSubmit();
        }
    };

    const isSelected = (val: string) => answer === val;

    return (
        <Card className={cn(
            "relative border-none overflow-hidden transition-all duration-500",
            "bg-white/70 backdrop-blur-xl shadow-[0_8px_32px_rgba(0,0,0,0.08)]",
            feedback?.isCorrect === true ? "ring-2 ring-green-500/50 bg-green-50/20" :
                feedback?.isCorrect === false ? "ring-2 ring-red-500/50 bg-red-50/20" : ""
        )}>
            {/* Background Grain/Texture Effect */}
            <div className="absolute inset-0 opacity-[0.03] pointer-events-none bg-[url('https://grainy-gradients.vercel.app/noise.svg')]" />

            <CardHeader className="relative p-8 pb-4">
                <div className="flex flex-col gap-6">
                    <div className="flex items-center justify-between">
                        <div className="flex flex-wrap gap-2">
                            <motion.span
                                initial={{ opacity: 0, x: -10 }}
                                animate={{ opacity: 1, x: 0 }}
                                className={cn(
                                    "text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-widest flex items-center gap-1.5 shadow-sm border",
                                    question.question_type === 'mcq' ? "bg-blue-500 text-white border-blue-400" :
                                        question.question_type === 'translate' ? "bg-indigo-500 text-white border-indigo-400" :
                                            "bg-amber-500 text-white border-amber-400"
                                )}
                            >
                                {question.question_type === 'mcq' && <List className="w-3 h-3" />}
                                {question.question_type === 'translate' && <Globe className="w-3 h-3" />}
                                {question.question_type === 'write_word' && <Type className="w-3 h-3" />}
                                {question.question_type === 'mcq' ? 'QCM' :
                                    question.question_type === 'translate' ? 'Traduction' :
                                        'Écriture'}
                            </motion.span>

                            {question.direction && !['source_to_target', 'target_to_source'].includes(question.direction) && (
                                <motion.span
                                    initial={{ opacity: 0, scale: 0.9 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    className="text-[10px] font-black px-3 py-1 rounded-full uppercase tracking-widest flex items-center gap-1.5 bg-white border border-border shadow-sm text-foreground/80"
                                >
                                    {question.direction === 'identify_grammar' && <FlaskConical className="w-3 h-3 text-emerald-500" />}
                                    {question.direction === 'attach_suffix' && <Brain className="w-3 h-3 text-purple-500" />}
                                    {question.direction === 'conjugate' && <Zap className="w-3 h-3 text-amber-500" />}
                                    {question.direction === 'identify_grammar' ? 'Structure' :
                                        question.direction === 'attach_suffix' ? 'Linguistique' :
                                            question.direction === 'conjugate' ? 'Conjugaison' :
                                                'Analyse'}
                                </motion.span>
                            )}

                            <span className="text-[10px] font-bold text-muted-foreground/60 px-3 py-1 rounded-full bg-black/5 flex items-center gap-1.5">
                                <Zap className="w-3 h-3 text-yellow-500 fill-yellow-500" />
                                {question.points} Points
                            </span>
                        </div>

                        <AnimatePresence>
                            {feedback && (
                                <motion.div
                                    initial={{ opacity: 0, scale: 0.8 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    className={cn(
                                        "px-4 py-1.5 rounded-full text-xs font-black uppercase tracking-widest shadow-lg",
                                        feedback.isCorrect ? "bg-green-500 text-white" : "bg-red-500 text-white"
                                    )}
                                >
                                    {feedback.isCorrect ? "Parfait !" : "Raté !"}
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>

                    <motion.h2
                        layout
                        className="text-2xl md:text-3xl font-black tracking-tight text-foreground leading-[1.2]"
                    >
                        {question.question_text}
                    </motion.h2>
                </div>
            </CardHeader>

            <CardContent className="relative p-8 pt-4">
                <AnimatePresence mode="wait">
                    {question.question_type === 'mcq' && question.options ? (
                        <div className="grid grid-cols-1 gap-3">
                            {question.options.map((option, idx) => {
                                const isSelectedOption = isSelected(option.id);
                                const isCorrectTarget = feedback?.correctAnswer === option.option_text;

                                let stateStyles = "bg-white/50 border-white/40 hover:bg-white/80 hover:border-primary/30 active:scale-[0.98]";
                                if (feedback) {
                                    if (isCorrectTarget) {
                                        stateStyles = "bg-green-500 text-white border-green-400 shadow-[0_0_20px_rgba(34,197,94,0.3)]";
                                    } else if (isSelectedOption && !feedback.isCorrect) {
                                        stateStyles = "bg-red-500 text-white border-red-400";
                                    } else {
                                        stateStyles = "opacity-40 grayscale-[0.5]";
                                    }
                                } else if (isSelectedOption) {
                                    stateStyles = "bg-primary text-white border-primary-foreground/20 shadow-lg shadow-primary/20 scale-[1.02]";
                                }

                                return (
                                    <motion.div
                                        key={option.id}
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        transition={{ delay: idx * 0.05 }}
                                        onClick={() => !feedback && onAnswerChange(option.id)}
                                        className={cn(
                                            "relative group flex items-center p-5 rounded-2xl border-2 transition-all duration-300",
                                            !feedback && "cursor-pointer hover:shadow-[0_4px_12px_rgba(0,0,0,0.05)]",
                                            stateStyles
                                        )}
                                    >
                                        <div className={cn(
                                            "flex items-center justify-center w-6 h-6 rounded-full border-2 mr-4 transition-all duration-300",
                                            isSelectedOption || (feedback && isCorrectTarget) ? "bg-white scale-110" : "bg-black/5"
                                        )}>
                                            {(isSelectedOption || (feedback && isCorrectTarget)) && (
                                                ((isSelectedOption && !feedback) || isCorrectTarget) ? (
                                                    <Check className="w-3.5 h-3.5 text-green-600" />
                                                ) : (
                                                    <X className="w-3.5 h-3.5 text-red-600" />
                                                )
                                            )}
                                        </div>
                                        <span className="text-lg font-bold tracking-tight">{option.option_text}</span>
                                    </motion.div>
                                );
                            })}
                        </div>
                    ) : (
                        <div className="space-y-6">
                            <div className="relative group">
                                <Input
                                    autoFocus
                                    type="text"
                                    value={answer || ''}
                                    onChange={(e) => onAnswerChange(e.target.value)}
                                    onKeyDown={handleKeyDown}
                                    disabled={!!feedback}
                                    placeholder="Répondez ici..."
                                    className={cn(
                                        "text-xl md:text-2xl p-8 h-auto font-bold rounded-2xl border-2 transition-all duration-500 bg-white/50 focus-visible:ring-offset-0 focus-visible:ring-0",
                                        feedback?.isCorrect === true ? "border-green-500 text-green-700 bg-green-50" :
                                            feedback?.isCorrect === false ? "border-red-500 text-red-700 bg-red-50" :
                                                "border-black/5 hover:border-black/10 focus-visible:border-primary focus-visible:bg-white shadow-inner"
                                    )}
                                />
                                {!feedback && !answer && (
                                    <div className="absolute right-6 top-1/2 -translate-y-1/2 flex items-center gap-2 opacity-30 group-hover:opacity-50 transition-opacity">
                                        <kbd className="px-2 py-1 rounded-lg bg-black/5 text-[10px] uppercase font-black">Entrée</kbd>
                                    </div>
                                )}
                            </div>

                            {feedback && !feedback.isCorrect && (
                                <motion.div
                                    initial={{ opacity: 0, y: 10 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    className="p-6 rounded-3xl bg-red-500/10 border border-red-500/20"
                                >
                                    <div className="flex items-center gap-2 text-red-600 mb-2">
                                        <Info className="w-4 h-4" />
                                        <span className="text-[10px] font-black uppercase tracking-widest">Réponse attendue</span>
                                    </div>
                                    <p className="text-2xl font-black text-foreground mb-3">{feedback.correctAnswer}</p>

                                    {feedback.ai_explanation && (
                                        <div className="pt-3 border-t border-red-500/10">
                                            <p className="text-xs font-medium text-red-600/80 leading-relaxed italic">
                                                {feedback.ai_explanation}
                                            </p>
                                        </div>
                                    )}
                                </motion.div>
                            )}

                            {feedback && feedback.isCorrect && feedback.ai_explanation && (
                                <motion.div
                                    initial={{ opacity: 0, y: 5 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    className="px-6 py-2"
                                >
                                    <p className="text-[11px] font-medium text-green-600/60 italic">
                                        {feedback.ai_explanation}
                                    </p>
                                </motion.div>
                            )}

                            <div className="flex items-center justify-center gap-3 px-6 py-4 rounded-2xl bg-black/[0.02] border border-black/[0.04]">
                                <Brain className="w-4 h-4 text-primary/40" />
                                <p className="text-[11px] font-bold text-muted-foreground/60 uppercase tracking-widest">
                                    {question.question_type === 'translate'
                                        ? "Traduisez exactement la phrase"
                                        : "Analysez et répondez avec précision"}
                                </p>
                            </div>
                        </div>
                    )}
                </AnimatePresence>
            </CardContent>
        </Card>
    );
}
