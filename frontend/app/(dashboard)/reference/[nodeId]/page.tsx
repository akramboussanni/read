'use client';

import React, { useEffect, useState, use, useRef, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { courseApi } from '@/lib/api/course';
import type { CourseNode } from '@/lib/types/course';
import { Button } from '@/components/ui/button';
import { ArrowLeft, BookOpen, CheckCircle2, ChevronRight, ChevronLeft, Loader2 } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

// ─── Block Types ──────────────────────────────────────────────
interface TextBlock { type: 'text'; content: string }
interface FlashcardsBlock {
    type: 'flashcards';
    cards: { arabic: string; french: string; tip?: string }[];
}
interface QuranBlock {
    type: 'quran';
    arabic: string;
    translation: string;
    highlight: string;
    source: string;
    context?: string; // friendly explanation for kids
}
interface QuizBlock {
    type: 'quiz';
    question: string;
    options: string[];
    answer: number;
    explanation: string;
}
interface TipBlock { type: 'tip'; title: string; content: string }
type LessonBlock = TextBlock | FlashcardsBlock | QuranBlock | QuizBlock | TipBlock;
interface LessonContent { type: 'lesson'; title: string; blocks: LessonBlock[] }

// ─── Mini markdown: bold → emerald, italic ────────────────────
function renderMd(text: string) {
    return text.split(/(\*\*[^*]+\*\*|\*[^*]+\*)/g).map((p, i) => {
        if (p.startsWith('**')) return <strong key={i} className="font-bold text-emerald-700 dark:text-emerald-400">{p.slice(2, -2)}</strong>;
        if (p.startsWith('*')) return <em key={i}>{p.slice(1, -1)}</em>;
        return p;
    });
}

// ─── Text Block — plain article prose ────────────────────────
function TextBlockView({ block }: { block: TextBlock }) {
    return (
        <p className="text-[15px] md:text-base leading-[1.85] text-slate-700 dark:text-slate-300">
            {renderMd(block.content)}
        </p>
    );
}

// ─── Flashcards — full width single card swiper ───────────────
function FlashcardsBlockView({ block }: { block: FlashcardsBlock }) {
    const [idx, setIdx] = useState(0);
    const [flipped, setFlipped] = useState(false);
    const touchStart = useRef<number | null>(null);
    const card = block.cards[idx];
    const total = block.cards.length;

    const next = useCallback(() => {
        setFlipped(false);
        setTimeout(() => setIdx(i => (i + 1) % total), 150);
    }, [total]);

    const prev = useCallback(() => {
        setFlipped(false);
        setTimeout(() => setIdx(i => (i - 1 + total) % total), 150);
    }, [total]);

    return (
        <div className="my-2">
            {/* Label */}
            <div className="flex items-center gap-3 mb-4">
                <div className="h-px flex-1 bg-border" />
                <span className="text-[10px] font-black uppercase tracking-[0.15em] text-muted-foreground">
                    🃏 Vocabulaire — {idx + 1} / {total}
                </span>
                <div className="h-px flex-1 bg-border" />
            </div>

            {/* Card */}
            <div
                className="relative mx-auto cursor-pointer select-none"
                style={{ perspective: '1200px', maxWidth: 480 }}
                onClick={() => setFlipped(f => !f)}
                onTouchStart={e => { touchStart.current = e.touches[0].clientX; }}
                onTouchEnd={e => {
                    if (touchStart.current === null) return;
                    const diff = touchStart.current - e.changedTouches[0].clientX;
                    if (Math.abs(diff) > 50) diff > 0 ? next() : prev();
                    touchStart.current = null;
                }}
            >
                <AnimatePresence mode="wait">
                    <motion.div
                        key={idx}
                        initial={{ opacity: 0, x: 30 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -30 }}
                        transition={{ duration: 0.18 }}
                    >
                        <motion.div
                            animate={{ rotateY: flipped ? 180 : 0 }}
                            transition={{ duration: 0.5, type: 'spring', stiffness: 200, damping: 24 }}
                            style={{ transformStyle: 'preserve-3d', position: 'relative', height: 220 }}
                        >
                            {/* Front */}
                            <div className={cn(
                                "absolute inset-0 rounded-3xl flex flex-col items-center justify-center p-8",
                                "bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-900/30 dark:to-orange-900/15",
                                "border-2 border-amber-200/80 dark:border-amber-700/40 border-b-[5px] shadow-lg"
                            )} style={{ backfaceVisibility: 'hidden' }}>
                                <p className="text-6xl font-bold text-amber-800 dark:text-amber-200 mb-4" dir="rtl">{card.arabic}</p>
                                <p className="text-[11px] font-black uppercase tracking-widest text-amber-400/80">
                                    Toucher pour révéler
                                </p>
                            </div>
                            {/* Back */}
                            <div className={cn(
                                "absolute inset-0 rounded-3xl flex flex-col items-center justify-center p-8",
                                "bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/30 dark:to-teal-900/15",
                                "border-2 border-emerald-200/80 dark:border-emerald-700/40 border-b-[5px] shadow-lg"
                            )} style={{ backfaceVisibility: 'hidden', transform: 'rotateY(180deg)' }}>
                                <p className="text-3xl font-black text-emerald-800 dark:text-emerald-200 text-center mb-3">{card.french}</p>
                                {card.tip && (
                                    <p className="text-xs text-emerald-600/80 dark:text-emerald-400/70 text-center font-medium leading-relaxed max-w-xs">{card.tip}</p>
                                )}
                                <div className="mt-4 px-3 py-1.5 rounded-full bg-emerald-100 dark:bg-emerald-900/50">
                                    <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300" dir="rtl">{card.arabic}</p>
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                </AnimatePresence>
            </div>

            {/* Navigation */}
            <div className="flex items-center justify-between mt-5 max-w-[480px] mx-auto">
                <button onClick={prev}
                    className="w-10 h-10 rounded-full border-2 border-border flex items-center justify-center hover:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/20 transition-all">
                    <ChevronLeft className="w-4 h-4" />
                </button>
                {/* Progress dots */}
                <div className="flex gap-1.5 items-center">
                    {block.cards.map((_, i) => (
                        <button key={i} onClick={() => { setFlipped(false); setTimeout(() => setIdx(i), 150); }}
                            className={cn("rounded-full transition-all", i === idx
                                ? "w-5 h-2 bg-emerald-500"
                                : "w-2 h-2 bg-muted-foreground/25 hover:bg-muted-foreground/50"
                            )} />
                    ))}
                </div>
                <button onClick={next}
                    className="w-10 h-10 rounded-full border-2 border-border flex items-center justify-center hover:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/20 transition-all">
                    <ChevronRight className="w-4 h-4" />
                </button>
            </div>

            <p className="text-center text-[10px] text-muted-foreground/60 mt-3 font-medium">
                ← Glisser ou utiliser les flèches →
            </p>
        </div>
    );
}

// ─── Quran — editorial pullquote style ───────────────────────
function QuranBlockView({ block }: { block: QuranBlock }) {
    const parts = block.arabic.split(block.highlight);

    return (
        <div className="my-2 pl-5 border-l-4 border-amber-400 dark:border-amber-500">
            <p className="text-[10px] font-black uppercase tracking-widest text-amber-500 dark:text-amber-400 mb-3">
                Exemple coranique · {block.source}
            </p>
            {/* Kid-friendly context */}
            {block.context && (
                <p className="text-[13px] text-slate-500 dark:text-slate-400 font-medium italic mb-4 leading-relaxed">
                    {block.context}
                </p>
            )}
            <p className="text-3xl md:text-4xl font-bold leading-loose text-slate-800 dark:text-slate-100 mb-4" dir="rtl">
                {parts.map((part, i) => (
                    <React.Fragment key={i}>
                        {part}
                        {i < parts.length - 1 && (
                            <span className="text-amber-600 dark:text-amber-400 bg-amber-100 dark:bg-amber-900/40 rounded px-1 mx-0.5">
                                {block.highlight}
                            </span>
                        )}
                    </React.Fragment>
                ))}
            </p>
            <p className="text-sm text-slate-500 dark:text-slate-400 italic font-medium">
                « {block.translation} »
            </p>
        </div>
    );
}

// ─── Mini Quiz — inline challenge ────────────────────────────
function QuizBlockView({ block }: { block: QuizBlock }) {
    const [selected, setSelected] = useState<number | null>(null);
    const answered = selected !== null;

    return (
        <div className="my-2">
            <div className="flex items-center gap-3 mb-4">
                <div className="h-px flex-1 bg-border" />
                <span className="text-[10px] font-black uppercase tracking-[0.15em] text-muted-foreground">
                    ✏️ Question rapide
                </span>
                <div className="h-px flex-1 bg-border" />
            </div>

            <p className="text-[15px] font-bold text-slate-800 dark:text-slate-100 mb-4">{block.question}</p>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {block.options.map((opt, i) => {
                    const isCorrect = i === block.answer;
                    const isSelected = i === selected;
                    return (
                        <button key={i} disabled={answered} onClick={() => setSelected(i)}
                            className={cn(
                                "w-full text-left px-4 py-3 rounded-2xl border-2 font-semibold text-sm transition-all",
                                !answered && "border-border bg-white dark:bg-slate-800/50 hover:border-emerald-400 hover:bg-emerald-50/50 dark:hover:bg-emerald-900/10",
                                answered && isCorrect && "border-emerald-400 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-800 dark:text-emerald-300",
                                answered && isSelected && !isCorrect && "border-red-400 bg-red-50 dark:bg-red-900/30 text-red-800 dark:text-red-300",
                                answered && !isSelected && !isCorrect && "border-border opacity-40"
                            )}>
                            <span className="flex items-center gap-2">
                                <span className={cn(
                                    "w-5 h-5 rounded-full text-[11px] flex items-center justify-center font-black shrink-0",
                                    !answered && "bg-muted",
                                    answered && isCorrect && "bg-emerald-500 text-white",
                                    answered && isSelected && !isCorrect && "bg-red-500 text-white",
                                    answered && !isSelected && !isCorrect && "bg-muted"
                                )}>
                                    {answered ? (isCorrect ? '✓' : isSelected ? '✗' : String.fromCharCode(65 + i)) : String.fromCharCode(65 + i)}
                                </span>
                                {opt}
                            </span>
                        </button>
                    );
                })}
            </div>

            <AnimatePresence>
                {answered && (
                    <motion.p initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }}
                        className={cn(
                            "mt-4 text-sm font-medium leading-relaxed",
                            selected === block.answer ? "text-emerald-700 dark:text-emerald-400" : "text-slate-600 dark:text-slate-400"
                        )}>
                        {selected === block.answer ? '✓ ' : ''}
                        {block.explanation}
                    </motion.p>
                )}
            </AnimatePresence>
        </div>
    );
}

// ─── Tip — inline annotation, not a card ─────────────────────
function TipBlockView({ block }: { block: TipBlock }) {
    return (
        <div className="flex gap-3 my-1 p-4 rounded-2xl bg-amber-50/80 dark:bg-amber-900/10 border border-amber-200/80 dark:border-amber-700/30">
            <span className="text-lg mt-0.5 shrink-0">💡</span>
            <div>
                <p className="text-xs font-black uppercase tracking-wider text-amber-600 dark:text-amber-400 mb-1">{block.title}</p>
                <p className="text-[13px] text-amber-900/80 dark:text-amber-200/80 font-medium leading-relaxed">{block.content}</p>
            </div>
        </div>
    );
}

// ─── Main Page ────────────────────────────────────────────────
export default function ReferencePage({ params }: { params: Promise<{ nodeId: string }> }) {
    const { nodeId } = use(params);
    const router = useRouter();
    const searchParams = useSearchParams();
    const courseId = searchParams.get('courseId') || '';

    const [node, setNode] = useState<CourseNode | null>(null);
    const [lesson, setLesson] = useState<LessonContent | null>(null);
    const [loading, setLoading] = useState(true);
    const [completing, setCompleting] = useState(false);
    const [completed, setCompleted] = useState(false);

    useEffect(() => { load(); }, [nodeId, courseId]);

    const load = async () => {
        try {
            if (!courseId) return;
            const course = await courseApi.getCourse(courseId);
            const found = course.nodes?.find((n: CourseNode) => n.id === nodeId);
            if (found) {
                setNode(found);
                if (found.lesson_content) {
                    try {
                        const parsed = JSON.parse(found.lesson_content);
                        if (parsed.type === 'lesson') {
                            const blocks = Array.isArray(parsed.blocks) ? parsed.blocks : JSON.parse(parsed.blocks);
                            setLesson({ type: 'lesson', title: parsed.title, blocks });
                        }
                    } catch (e) {
                        console.error('Failed to parse lesson content', e);
                    }
                }
            }
            const status = await courseApi.getCourseStatus(courseId);
            if (status?.node_statuses?.[nodeId]?.state === 'completed') setCompleted(true);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const handleComplete = async () => {
        if (!courseId) return;
        setCompleting(true);
        try {
            await courseApi.completeNode(courseId, nodeId);
            setCompleted(true);
            setTimeout(() => router.push('/'), 900);
        } catch (e) { console.error(e); }
        finally { setCompleting(false); }
    };

    if (loading) return (
        <div className="min-h-screen flex items-center justify-center">
            <div className="w-14 h-14 rounded-2xl bg-emerald-500 flex items-center justify-center shadow-xl animate-pulse">
                <BookOpen className="w-7 h-7 text-white" />
            </div>
        </div>
    );

    const blockCount = lesson?.blocks?.length ?? 0;

    return (
        <div className="min-h-screen bg-white dark:bg-slate-950 text-foreground pb-40">

            {/* Sticky header — minimal */}
            <div className="sticky top-0 z-20 bg-white/80 dark:bg-slate-950/80 backdrop-blur border-b border-border/50">
                <div className="container max-w-2xl mx-auto px-4 h-13 flex items-center justify-between py-3">
                    <button className="flex items-center gap-1.5 text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors"
                        onClick={() => router.push('/')}>
                        <ArrowLeft className="w-4 h-4" /> Retour
                    </button>
                    <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                        {blockCount} sections
                    </span>
                </div>
            </div>

            {/* Article header */}
            <div className="container max-w-2xl mx-auto px-5 pt-10 pb-8">
                <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
                    <div className="flex items-center gap-2 mb-3">
                        <div className="w-7 h-7 rounded-xl bg-emerald-500 flex items-center justify-center">
                            <BookOpen className="w-3.5 h-3.5 text-white" strokeWidth={2.5} />
                        </div>
                        <span className="text-[10px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">
                            Leçon interactive
                        </span>
                    </div>
                    <h1 className="text-3xl md:text-4xl font-black text-slate-900 dark:text-slate-50 leading-tight mb-2">
                        {lesson?.title || node?.title}
                    </h1>
                    {node?.description && (
                        <p className="text-base text-muted-foreground font-medium">{node.description}</p>
                    )}
                    {/* Thin accent line */}
                    <div className="mt-6 h-0.5 w-16 bg-emerald-500 rounded-full" />
                </motion.div>
            </div>

            {/* Article body — flowing content */}
            <main className="container max-w-2xl mx-auto px-5">
                {lesson?.blocks?.map((block, i) => (
                    <motion.div key={i}
                        initial={{ opacity: 0, y: 12 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: i * 0.05 }}
                        className="mb-8"
                    >
                        {block.type === 'text' && <TextBlockView block={block} />}
                        {block.type === 'flashcards' && <FlashcardsBlockView block={block} />}
                        {block.type === 'quran' && <QuranBlockView block={block} />}
                        {block.type === 'quiz' && <QuizBlockView block={block} />}
                        {block.type === 'tip' && <TipBlockView block={block} />}
                    </motion.div>
                ))}

                {!lesson && (
                    <div className="text-center py-20 text-muted-foreground">
                        <BookOpen className="w-12 h-12 mx-auto mb-4 opacity-30" />
                        <p className="font-bold">Contenu non disponible</p>
                    </div>
                )}

                {/* Completion */}
                <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.4 }}
                    className="pt-4 pb-16">

                    <div className="h-px bg-border mb-8" />

                    {completed ? (
                        <div className="flex items-center justify-center gap-3 text-emerald-600 dark:text-emerald-400 py-4">
                            <CheckCircle2 className="w-6 h-6" />
                            <span className="font-black text-base">Leçon terminée !</span>
                        </div>
                    ) : (
                        <Button size="lg" onClick={handleComplete} disabled={completing}
                            className={cn(
                                "w-full h-14 text-base font-black rounded-2xl",
                                "bg-emerald-500 hover:bg-emerald-600 text-white",
                                "border-b-[5px] border-emerald-700",
                                "hover:-translate-y-0.5 active:translate-y-0.5 active:border-b-0",
                                "transition-all shadow-lg shadow-emerald-500/20 gap-2"
                            )}>
                            {completing
                                ? <><Loader2 className="w-4 h-4 animate-spin" /> Enregistrement...</>
                                : <>J'ai compris — Continuer <ChevronRight className="w-4 h-4" /></>
                            }
                        </Button>
                    )}
                </motion.div>
            </main>
        </div>
    );
}
