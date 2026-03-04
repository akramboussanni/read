'use client';

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { addCourseEdge, deleteCourseEdge, deleteCourseNode, updateCourseNode, addCourseNode } from '@/lib/api/admin';
import { quizApi } from '@/lib/api/quiz';
import { QuizEditor } from '@/components/quiz/quiz-editor';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type { Course, CourseNode, CourseEdge, CourseStatus } from '@/lib/types/course';
import type { Quiz, Deck } from '@/lib/types/quiz';
import { X, Trash, Save, BookOpen, Pencil, ExternalLink, Plus, Lock, CheckCircle2, Star, Play, Flag, Coins } from 'lucide-react';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

export type CourseGraphMode = 'edit' | 'view';

interface CourseFlowGraphProps {
    course: Course;
    mode: CourseGraphMode;
    courseStatus?: CourseStatus | null;
    onNodePlay?: (node: CourseNode) => void;
    onGraphChange?: () => void;
}

// ─────────────────────────────────────────────────────────────
// Main Component
// ─────────────────────────────────────────────────────────────

export function CourseFlowGraph({
    course,
    mode,
    courseStatus,
    onNodePlay,
    onGraphChange,
}: CourseFlowGraphProps) {
    const { user } = useAuthStore();
    const courseNodes = course.nodes ?? [];
    const courseEdges = course.edges ?? [];

    // DEBUG: teacher status
    useEffect(() => {
        if (user?.role === 'teacher') {
            console.log('User is teacher:', user.email, 'Confirmed:', user.email_confirmed);
        }
    }, [user]);

    // Edit-mode sidebar state
    const [selectedNode, setSelectedNode] = useState<CourseNode | null>(null);
    const [editTitle, setEditTitle] = useState('');
    const [editDesc, setEditDesc] = useState('');
    const [editType, setEditType] = useState('lesson');

    // Auto Gen States
    const [decks, setDecks] = useState<Deck[]>([]);
    const [autoGenDeckId, setAutoGenDeckId] = useState('');
    const [autoGenCount, setAutoGenCount] = useState<number>(5);
    const [isGenerating, setIsGenerating] = useState(false);

    useEffect(() => {
        if (mode === 'edit') {
            quizApi.listDecks().then(setDecks).catch(console.error);
        }
    }, [mode]);

    // Layout processing: DAG Level Assignment
    const inDegree = new Map<string, number>();
    const adj = new Map<string, string[]>();

    courseNodes.forEach(n => {
        inDegree.set(n.id, 0);
        adj.set(n.id, []);
    });

    courseEdges.forEach(e => {
        if (inDegree.has(e.target)) inDegree.set(e.target, inDegree.get(e.target)! + 1);
        if (adj.has(e.source)) adj.get(e.source)!.push(e.target);
    });

    const level = new Map<string, number>();
    let queue = courseNodes.filter(n => inDegree.get(n.id) === 0);
    queue.forEach(n => level.set(n.id, 0));

    // Loop limits to prevent crazy graphs blocking UI
    let safety = 10000;
    while (queue.length > 0 && safety-- > 0) {
        const u = queue.shift()!;
        const uLevel = level.get(u.id)!;
        for (const v of adj.get(u.id) || []) {
            level.set(v, Math.max(level.get(v) || 0, uLevel + 1));
            inDegree.set(v, inDegree.get(v)! - 1);
            if (inDegree.get(v) === 0) {
                const tgt = courseNodes.find(n => n.id === v);
                if (tgt) queue.push(tgt);
            }
        }
    }

    // Handle cycles / floating nodes
    courseNodes.forEach(n => {
        if (!level.has(n.id)) level.set(n.id, 0); // Put orphaned nodes at level 0
    });

    let maxLevel = Math.max(...Array.from(level.values()), -1);
    if (maxLevel < 0) maxLevel = 0;

    const rows = Array.from({ length: maxLevel + 1 }, () => [] as CourseNode[]);
    courseNodes.forEach(n => rows[level.get(n.id) || 0].push(n));
    rows.forEach(row => row.sort((a, b) => a.sort_order - b.sort_order));

    // ── DOM measuring for SVG lines ──
    const containerRef = useRef<HTMLDivElement>(null);
    const [rects, setRects] = useState<Record<string, { x: number, y: number, w: number, h: number }>>({});

    const measureNodes = useCallback(() => {
        if (!containerRef.current) return;
        const container = containerRef.current.getBoundingClientRect();
        const els = containerRef.current.querySelectorAll('[data-node-id]');
        const newRects: Record<string, { x: number, y: number, w: number, h: number }> = {};
        els.forEach((el) => {
            const id = el.getAttribute('data-node-id')!;
            const rect = el.getBoundingClientRect();
            newRects[id] = {
                x: rect.x - container.x,
                y: rect.y - container.y,
                w: rect.width,
                h: rect.height
            };
        });
        setRects(newRects);
    }, []);

    useEffect(() => {
        // Small delay to ensure DOM is fully rendered
        const t1 = setTimeout(measureNodes, 50);
        const t2 = setTimeout(measureNodes, 300);
        // Observe resize 
        const rz = new ResizeObserver(() => measureNodes());
        if (containerRef.current) rz.observe(containerRef.current);

        return () => {
            clearTimeout(t1);
            clearTimeout(t2);
            rz.disconnect();
        };
    }, [courseNodes, courseEdges, measureNodes]);

    // ── Edit-mode callbacks ──────────────────────────────────────
    const handleNodeClick = (n: CourseNode) => {
        if (mode === 'view') {
            onNodePlay?.(n);
            return;
        }
        setEditTitle(n.title);
        setEditDesc(n.description);
        setEditType(n.node_type);
        setSelectedNode(n);
    };

    const handleCreateChild = async (parentId: string, e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            // 1. Create node
            const newNode = await addCourseNode(course.id, {
                title: 'Nouvelle Étape',
                node_type: 'quiz',
                description: '',
                icon: 'star',
                position_x: 0,
                position_y: 0,
                sort_order: courseNodes.length,
            });
            // 2. Create edge linking them
            await addCourseEdge(course.id, {
                source: parentId,
                target: newNode.id,
                edge_type: 'required',
            });
            onGraphChange?.();
        } catch (err) {
            console.error(err);
        }
    };

    const handleAutoGen = async () => {
        if (!selectedNode || !autoGenDeckId) return;
        setIsGenerating(true);
        try {
            const tempDeck = decks.find(d => d.id === autoGenDeckId);
            const quizTitle = editTitle || (tempDeck ? `Quiz : ${tempDeck.title}` : 'Quiz Auto');

            const quiz = await quizApi.createQuiz({
                title: quizTitle,
                is_public: true,
                auto_generate: {
                    deck_selections: [{ deck_id: autoGenDeckId, categories: [] }],
                    question_types: ['mcq', 'write_word', 'translate'],
                    directions: ['source_to_target', 'target_to_source'],
                    question_count: autoGenCount
                }
            });
            await linkQuizToNode(quiz.id, quizTitle);
        } catch (err) {
            console.error(err);
        } finally {
            setIsGenerating(false);
        }
    };

    const handleCreateEmptyQuiz = async () => {
        if (!selectedNode) return;
        setIsGenerating(true);
        try {
            const quiz = await quizApi.createQuiz({
                title: editTitle || 'Nouveau Quiz',
                is_public: true,
            });
            await linkQuizToNode(quiz.id, editTitle || 'Nouveau Quiz');
        } catch (err) {
            console.error(err);
        } finally {
            setIsGenerating(false);
        }
    };

    const linkQuizToNode = async (quizId: string, nodeTitle: string) => {
        if (!selectedNode) return;
        const config = JSON.parse((selectedNode.quiz_config as string) || '{}');
        config.quiz_id = quizId;
        const updatedNode = await updateCourseNode(selectedNode.id, {
            title: nodeTitle,
            description: editDesc,
            icon: selectedNode.icon,
            position_x: selectedNode.position_x,
            position_y: selectedNode.position_y,
            sort_order: selectedNode.sort_order,
            node_type: editType as string,
            quiz_config: JSON.stringify(config),
            lesson_content: (selectedNode.lesson_content as string) || '',
        });

        // Use the updated node
        setEditTitle(updatedNode.title);
        setSelectedNode(updatedNode);
        onGraphChange?.();
    };

    const saveSelectedNode = async () => {
        if (!selectedNode) return;
        await updateCourseNode(selectedNode.id, {
            title: editTitle,
            description: editDesc,
            icon: selectedNode.icon,
            position_x: selectedNode.position_x,
            position_y: selectedNode.position_y,
            sort_order: selectedNode.sort_order,
            node_type: editType as string,
            quiz_config: selectedNode.quiz_config as string,
            lesson_content: (selectedNode.lesson_content as string) || '',
        });

        // Update local object to reflect the save without closing
        setSelectedNode({ ...selectedNode, title: editTitle, description: editDesc, node_type: editType as any });
        onGraphChange?.();
    };

    const deleteNode = async () => {
        if (!selectedNode) return;
        await deleteCourseNode(selectedNode.id);
        setSelectedNode(null);
        onGraphChange?.();
    };

    // ────────────────────────────────────────────────────────────
    // Calculate completed colors
    // ────────────────────────────────────────────────────────────
    const statusColorMap = {
        // Very flat grey for locked, completely different from vibrant colors
        locked: { bg: 'bg-[#E5E5E5]', text: 'text-[#A3A3A3]', border: 'border-[#D4D4D4]' },
        unlocked: { bg: 'bg-[#1CB0F6]', text: 'text-white', border: 'border-[#1899D6]' },
        completed: { bg: 'bg-[#FFC800]', text: 'text-white', border: 'border-[#E5B400]' },
        mastered: { bg: 'bg-[#FFD900]', text: 'text-white', border: 'border-[#FFC800]' },
        // Admin mode generic colors
        admin: { bg: 'bg-[#CE82FF]', text: 'text-white', border: 'border-[#B461EB]' },
    };

    const sidebarWidth = 'w-80';

    return (
        <div className="flex-1 flex w-full relative h-[calc(100vh-64px)]">
            {/* Scrollable Map Canvas */}
            <div className="flex-1 overflow-y-auto w-full relative bg-white pb-32">
                <div className="absolute inset-0 pointer-events-none opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(circle at center, black 1.5px, transparent 2px)', backgroundSize: '16px 16px' }} />

                <div ref={containerRef} className="relative max-w-lg mx-auto w-full flex flex-col items-center min-h-full py-16 gap-8 md:gap-10">

                    {/* SVG Connection Lines Layer */}
                    <svg className="absolute inset-0 w-full h-full pointer-events-none z-0">
                        {courseEdges.map(e => {
                            const src = rects[e.source];
                            const tgt = rects[e.target];
                            if (!src || !tgt) return null;

                            const x1 = src.x + src.w / 2;
                            const y1 = src.y + src.h - 8;
                            const x2 = tgt.x + tgt.w / 2;
                            const y2 = tgt.y + 8;

                            // Smooth bezier curve connecting layers perfectly
                            const ctrlY = Math.max(y1 + 40, (y1 + y2) / 2);

                            const isOptional = e.edge_type === 'optional';
                            const lineColor = isOptional ? '#10B981' : '#E5E5E5';
                            return (
                                <path
                                    key={e.id}
                                    d={`M ${x1} ${y1} C ${x1} ${ctrlY}, ${x2} ${ctrlY}, ${x2} ${y2}`}
                                    stroke={lineColor}
                                    strokeWidth={isOptional ? 6 : 14}
                                    fill="none"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeDasharray={isOptional ? '10 7' : undefined}
                                    opacity={isOptional ? 0.7 : 1}
                                />
                            );
                        })}
                    </svg>

                    {/* Nodes Layer: Ordered by topological levels */}
                    {rows.map((row, levelIndex) => (
                        <div key={levelIndex} className="relative z-10 w-full flex flex-wrap justify-center gap-4 md:gap-6 px-4">
                            {row.map((node) => {
                                const statusObj = mode === 'view' ? courseStatus?.node_statuses?.[node.id] : undefined;
                                // Reference nodes are always accessible — force unlocked
                                const isReference = node.node_type === 'reference';
                                const rawStatus = mode === 'view' ? (statusObj ? statusObj.state : 'locked') : 'admin';
                                let status = isReference && mode === 'view' ? 'unlocked' : rawStatus;

                                // Teachers can skip progression
                                if (user?.role === 'teacher' && mode === 'view') {
                                    status = 'unlocked';
                                }
                                const isLocked = status === 'locked';
                                const isUnlocked = status === 'unlocked';
                                const sColor = statusColorMap[status as keyof typeof statusColorMap] || statusColorMap.locked;
                                const isSelected = selectedNode?.id === node.id;

                                return (
                                    <div key={node.id} className="relative flex flex-col items-center">
                                        {/* "JOUER" Tooltip */}
                                        {mode === 'view' && isUnlocked && (
                                            <div className="absolute -top-12 z-20 flex flex-col items-center animate-bounce">
                                                <div className={cn(
                                                    "bg-white border-2 border-border shadow-md px-3 py-1.5 rounded-xl uppercase tracking-widest font-black text-xs whitespace-nowrap flex items-center gap-2",
                                                    node.node_type === 'reference' ? "text-emerald-600" : "text-primary"
                                                )}>
                                                    {node.node_type === 'reference' ? 'Lire la fiche' : (
                                                        <>
                                                            Commencer
                                                            {(() => {
                                                                try {
                                                                    const config = JSON.parse((node.quiz_config as string) || '{}');
                                                                    if (config.coin_reward) {
                                                                        return (
                                                                            <span className="flex items-center gap-1 bg-yellow-400 text-white px-1.5 py-0.5 rounded-lg text-[10px] ml-1 shadow-sm">
                                                                                <Coins className="w-3 h-3" />
                                                                                +{config.coin_reward}
                                                                            </span>
                                                                        );
                                                                    }
                                                                } catch { }
                                                                return null;
                                                            })()}
                                                        </>
                                                    )}
                                                </div>
                                                <div className="w-3 h-3 bg-white border-b-2 border-r-2 border-border rotate-45 -mt-[7px] z-10" />
                                            </div>
                                        )}

                                        {/* Circle Node Button */}
                                        <button
                                            data-node-id={node.id}
                                            onClick={() => handleNodeClick(node)}
                                            className={cn(
                                                "w-20 h-20 rounded-full flex items-center justify-center border-b-8 transition-all relative transform active:scale-95 active:border-b-4 active:translate-y-1 shadow-sm",
                                                sColor.bg, sColor.text, sColor.border,
                                                isLocked ? "cursor-not-allowed opacity-100" : "cursor-pointer hover:-translate-y-1 hover:shadow-lg",
                                                isSelected && "ring-8 ring-indigo-500/30"
                                            )}
                                        >
                                            {/* Icon inner stamp */}
                                            <div className="w-14 h-14 rounded-full bg-white/20 flex flex-col items-center justify-center">
                                                {isLocked ? (
                                                    <Lock className="w-7 h-7 fill-black/10 stroke-[2.5]" />
                                                ) : status === 'completed' || status === 'mastered' ? (
                                                    <CheckCircle2 className="w-8 h-8 fill-white/20 stroke-[2.5]" />
                                                ) : node.node_type === 'quiz' ? (
                                                    <Star className="w-7 h-7 fill-white/20 stroke-[2.5]" />
                                                ) : node.node_type === 'reference' ? (
                                                    <BookOpen className="w-7 h-7 fill-white/20 stroke-[2.5]" />
                                                ) : (
                                                    <BookOpen className="w-6 h-6 fill-white/20 stroke-[2.5]" />
                                                )}
                                            </div>
                                        </button>

                                        {/* Label below node */}
                                        <span className={cn(
                                            "mt-3 text-[11px] font-black uppercase tracking-wider text-center max-w-[110px] leading-tight",
                                            "relative z-20 px-2.5 py-1 rounded-xl bg-white/90 backdrop-blur shadow-sm",
                                            isLocked ? "text-[#B0B0B0]" : "text-slate-700"
                                        )}>
                                            {node.title}
                                        </span>

                                        {/* Edit mode: Quick add child button */}
                                        {mode === 'edit' && (
                                            <button
                                                onClick={(e) => handleCreateChild(node.id, e)}
                                                className="absolute -right-3 -top-3 w-7 h-7 bg-emerald-500 hover:bg-emerald-600 text-white rounded-full flex items-center justify-center shadow-md border-2 border-white transition-transform hover:scale-110 z-20"
                                                title="Ajouter une étape suivante"
                                            >
                                                <Plus className="w-4 h-4" strokeWidth={3} />
                                            </button>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    ))}

                    {/* End Flag if path is empty or at very bottom */}
                    {courseNodes.length > 0 && (
                        <div className="relative z-10 w-full flex justify-center mt-4">
                            <div className="w-16 h-16 rounded-3xl bg-amber-400 border-b-8 border-amber-600 flex items-center justify-center shadow-lg">
                                <Flag className="w-7 h-7 text-white fill-white" />
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* ── Admin Edit Sidebar ── */}
            {mode === 'edit' && (
                <AnimatePresence>
                    {selectedNode && (
                        <motion.div
                            initial={{ x: 320, opacity: 0 }}
                            animate={{ x: 0, opacity: 1 }}
                            exit={{ x: 320, opacity: 0 }}
                            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
                            className={cn(
                                'absolute right-0 top-0 h-full border-l border-border bg-card shadow-2xl flex flex-col shrink-0 overflow-hidden transition-all duration-300 z-50',
                                sidebarWidth
                            )}
                        >
                            {/* Sidebar header */}
                            <div className="p-4 border-b border-border flex items-center justify-between shrink-0 bg-slate-50">
                                <div className="flex items-center gap-3">
                                    <div className={cn(
                                        'w-10 h-10 rounded-xl flex items-center justify-center text-white text-xs font-bold shadow-sm',
                                        selectedNode.node_type === 'quiz' ? 'bg-[#1CB0F6]' :
                                            selectedNode.node_type === 'start' ? 'bg-indigo-500' :
                                                selectedNode.node_type === 'milestone' ? 'bg-[#FFC800]' :
                                                    selectedNode.node_type === 'checkpoint' ? 'bg-[#FF4B4B]' :
                                                        selectedNode.node_type === 'reference' ? 'bg-emerald-500' : 'bg-[#CE82FF]'
                                    )}>
                                        {String(selectedNode.node_type).slice(0, 2).toUpperCase()}
                                    </div>
                                    <div>
                                        <h3 className="font-black text-sm text-slate-800">Paramètres de l'Étape</h3>
                                        <p className="text-[10px] font-bold tracking-widest uppercase text-muted-foreground">{selectedNode.node_type}</p>
                                    </div>
                                </div>
                                <Button variant="ghost" size="icon" onClick={() => setSelectedNode(null)} className="h-8 w-8 hover:bg-slate-200">
                                    <X className="w-4 h-4 text-slate-500" />
                                </Button>
                            </div>

                            {/* Sidebar body */}
                            <div className="flex-1 overflow-y-auto p-5 space-y-6">
                                <div className="space-y-1.5">
                                    <Label className="text-[10px] font-black tracking-widest uppercase text-slate-500">Titre de l'Étape</Label>
                                    <Input value={editTitle} onChange={(e) => setEditTitle(e.target.value)} className="h-11 font-bold bg-slate-50" />
                                </div>
                                <div className="space-y-1.5">
                                    <Label className="text-[10px] font-black tracking-widest uppercase text-slate-500">Type de l'Étape</Label>
                                    <select
                                        value={editType}
                                        onChange={(e) => setEditType(e.target.value)}
                                        className="flex h-11 w-full rounded-md border border-input bg-slate-50 px-3 py-2 text-sm font-bold shadow-sm"
                                    >
                                        <option value="lesson">Leçon classique</option>
                                        <option value="quiz">Quiz Interactif</option>
                                        <option value="reference">📖 Fiche de révision</option>
                                        <option value="checkpoint">Checkpoint</option>
                                        <option value="milestone">Jalon</option>
                                        <option value="start">Départ</option>
                                    </select>
                                </div>
                                <div className="space-y-1.5">
                                    <Label className="text-[10px] font-black tracking-widest uppercase text-slate-500">Description</Label>
                                    <Input value={editDesc} onChange={(e) => setEditDesc(e.target.value)} className="bg-slate-50" />
                                </div>

                                {/* Connections section replacing full canvas drag/drop */}
                                <div className="pt-4 border-t border-border space-y-3">
                                    <Label className="text-[10px] font-black tracking-widest uppercase text-slate-500">Structure</Label>
                                    <div className="bg-slate-50 border border-border p-4 rounded-xl">
                                        <p className="text-xs font-semibold text-slate-600 mb-3">Relations avec les autres étapes</p>
                                        {/* Parents */}
                                        <div className="mb-3">
                                            <span className="text-[10px] font-bold text-slate-400 block mb-1">ÉTAPE PRÉCÉDENTE (PARENTS)</span>
                                            {courseEdges.filter(e => e.target === selectedNode.id).length === 0 ? (
                                                <p className="text-xs text-slate-400 italic">Ceci est une étape de départ.</p>
                                            ) : (
                                                <div className="space-y-1.5">
                                                    {courseEdges.filter(e => e.target === selectedNode.id).map(e => {
                                                        const parent = courseNodes.find(n => n.id === e.source);
                                                        return (
                                                            <div key={e.id} className="flex items-center justify-between bg-white border border-border px-3 py-1.5 rounded-lg">
                                                                <span className="text-xs font-bold text-slate-700 truncate max-w-[150px]">{parent?.title || 'Étape'}</span>
                                                                <button onClick={async () => { await deleteCourseEdge(e.id); onGraphChange?.(); }} className="text-red-500 hover:text-red-600"><Trash className="w-3.5 h-3.5" /></button>
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            )}
                                        </div>
                                        {/* Children */}
                                        <div>
                                            <span className="text-[10px] font-bold text-slate-400 block mb-1">ÉTAPE SUIVANTE (ENFANTS)</span>
                                            {courseEdges.filter(e => e.source === selectedNode.id).length === 0 ? (
                                                <p className="text-xs text-slate-400 italic mb-2">Ceci est une étape finale.</p>
                                            ) : (
                                                <div className="space-y-1.5 mb-2">
                                                    {courseEdges.filter(e => e.source === selectedNode.id).map(e => {
                                                        const child = courseNodes.find(n => n.id === e.target);
                                                        return (
                                                            <div key={e.id} className="flex items-center justify-between bg-white border border-border px-3 py-1.5 rounded-lg">
                                                                <span className="text-xs font-bold text-slate-700 truncate max-w-[150px]">{child?.title || 'Étape'}</span>
                                                                <button onClick={async () => { await deleteCourseEdge(e.id); onGraphChange?.(); }} className="text-red-500 hover:text-red-600"><Trash className="w-3.5 h-3.5" /></button>
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            )}
                                            <Button size="sm" variant="outline" className="w-full text-xs font-bold h-8 border-dashed mt-2" onClick={(e) => handleCreateChild(selectedNode.id, e)}>
                                                <Plus className="w-3.5 h-3.5 mr-1" /> Créer une suite
                                            </Button>
                                        </div>
                                    </div>
                                </div>

                                {/* Quiz sub-panel */}
                                {editType === 'quiz' && (
                                    <div className="pt-4 border-t space-y-3 pb-8">
                                        <Label className="text-[10px] font-black tracking-widest uppercase text-slate-500 flex items-center gap-1.5 mb-2">
                                            <BookOpen className="w-3.5 h-3.5" /> Contenu du Quiz
                                        </Label>

                                        {(() => {
                                            try {
                                                const config = JSON.parse((selectedNode.quiz_config as string) || '{}');
                                                if (config.quiz_id) {
                                                    return (
                                                        <div className="p-4 rounded-xl bg-[#1CB0F6]/10 border-2 border-[#1CB0F6]/20">
                                                            <p className="text-xs font-black text-[#1CB0F6] mb-3 uppercase tracking-widest">Quiz lié (ID: {config.quiz_id})</p>
                                                            <Button size="sm" onClick={() => window.open(`/quizzes/${config.quiz_id}`, '_blank')} className="w-full font-bold bg-[#1CB0F6] hover:bg-[#1899D6] text-white shadow-sm h-10 border-b-4 border-[#1899D6] active:border-b-0 active:translate-y-1">
                                                                <ExternalLink className="w-4 h-4 mr-2" /> Ouvrir l'éditeur de quiz (Nouvel onglet)
                                                            </Button>
                                                        </div>
                                                    );
                                                }
                                            } catch { }

                                            return (
                                                <div className="space-y-4">
                                                    {/* Quick auto gen */}
                                                    <div className="bg-slate-50 p-4 border rounded-xl space-y-3">
                                                        <p className="text-xs font-bold text-slate-700">1. Génération Automatique</p>
                                                        <select
                                                            className="w-full h-9 rounded text-xs border bg-white px-2 font-medium"
                                                            value={autoGenDeckId} onChange={e => setAutoGenDeckId(e.target.value)}>
                                                            <option value="">-- Choisir un deck --</option>
                                                            {decks.map(d => <option key={d.id} value={d.id}>{d.title}</option>)}
                                                        </select>
                                                        <div className="flex items-center gap-2">
                                                            <Input type="number" min="1" max="50" className="h-9 w-20 text-xs" value={autoGenCount} onChange={e => setAutoGenCount(parseInt(e.target.value))} />
                                                            <span className="text-xs text-slate-500">questions auto-générées</span>
                                                        </div>
                                                        <Button
                                                            size="sm"
                                                            disabled={!autoGenDeckId || isGenerating}
                                                            onClick={handleAutoGen}
                                                            className="w-full h-9 text-xs font-bold bg-green-500 hover:bg-green-600 text-white shadow-sm"
                                                        >
                                                            {isGenerating ? "Création..." : "Générer & Lier le Quiz"}
                                                        </Button>
                                                    </div>

                                                    <div className="text-center text-[10px] font-black uppercase text-slate-400 tracking-wider">-- OU --</div>

                                                    {/* Create Empty Manual Quiz */}
                                                    <div className="bg-slate-50 p-4 border rounded-xl space-y-3">
                                                        <p className="text-xs font-bold text-slate-700">2. Création Manuelle</p>
                                                        <Button
                                                            size="sm" variant="outline"
                                                            onClick={handleCreateEmptyQuiz}
                                                            disabled={isGenerating}
                                                            className="w-full h-9 text-xs font-bold border-2 border-[#1CB0F6] text-[#1CB0F6] hover:bg-[#1CB0F6]/10"
                                                        >
                                                            Créer un Quiz Manuel (Vide)
                                                        </Button>
                                                    </div>
                                                </div>
                                            );
                                        })()}
                                    </div>
                                )}
                            </div>

                            {/* Sidebar footer ALWAYS SHOWS */}
                            <div className="p-5 border-t border-border flex items-center justify-between shrink-0 bg-white shadow-[0_-4px_10px_rgba(0,0,0,0.05)] relative z-20">
                                <Button variant="ghost" size="sm" onClick={deleteNode} className="text-red-500 hover:text-red-600 hover:bg-red-50 font-bold">
                                    <Trash className="w-4 h-4 mr-1.5" /> Supprimer
                                </Button>
                                <Button size="sm" onClick={saveSelectedNode} className="font-bold bg-[#22C55E] hover:bg-[#16A34A] text-white h-11 px-6 shadow-sm border-b-4 border-[#16A34A] active:border-b-0 active:translate-y-1 rounded-xl">
                                    <Save className="w-4 h-4 mr-2" /> Enregistrer
                                </Button>
                            </div>
                        </motion.div>
                    )}
                </AnimatePresence>
            )}
        </div >
    );
}
