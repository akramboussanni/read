'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { createCourse, autoGenerateCourse, listDecks, listTemplates, createFromTemplate } from '@/lib/api/admin';
import type { DeckWithCounts, TemplateInfo } from '@/lib/types/admin';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    BookOpen, Sparkles, ArrowLeft, Wand2,
    CheckCircle2, Layers, Zap, GraduationCap,
    Pencil, ChevronRight, Loader2,
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

type Mode = 'pick' | 'template' | 'auto' | 'manual';

export default function CreateCoursePage() {
    const router = useRouter();
    const [mode, setMode] = useState<Mode>('pick');

    // Template state
    const [templates, setTemplates] = useState<TemplateInfo[]>([]);
    const [loadingTemplates, setLoadingTemplates] = useState(false);
    const [creatingTemplate, setCreatingTemplate] = useState<string | null>(null);

    // Auto state
    const [autoTitle, setAutoTitle] = useState('');
    const [deckId, setDeckId] = useState('');
    const [decks, setDecks] = useState<DeckWithCounts[]>([]);
    const [loadingDecks, setLoadingDecks] = useState(false);

    // Manual state
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [icon, setIcon] = useState('book');
    const [color, setColor] = useState('#6C5CE7');

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (mode === 'template' && templates.length === 0) {
            setLoadingTemplates(true);
            listTemplates()
                .then(setTemplates)
                .catch(console.error)
                .finally(() => setLoadingTemplates(false));
        }
        if (mode === 'auto' && decks.length === 0) {
            setLoadingDecks(true);
            listDecks()
                .then(setDecks)
                .catch(console.error)
                .finally(() => setLoadingDecks(false));
        }
    }, [mode]);

    const handleCreateFromTemplate = async (tmpl: TemplateInfo) => {
        setCreatingTemplate(tmpl.filename);
        setError(null);
        try {
            const course = await createFromTemplate({ template_filename: tmpl.filename });
            router.push(`/admin/courses/${course.id}/visual-editor`);
        } catch (err: any) {
            setError(err?.response?.data || 'Erreur lors de la création');
            setCreatingTemplate(null);
        }
    };

    const handleAutoCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!deckId) return;
        setLoading(true); setError(null);
        try {
            const course = await autoGenerateCourse({ title: autoTitle, deck_id: deckId });
            router.push(`/admin/courses/${course.id}/visual-editor`);
        } catch (err: any) {
            setError(err?.response?.data?.error || 'Échec de la génération');
            setLoading(false);
        }
    };

    const handleManualCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true); setError(null);
        try {
            const course = await createCourse({ title, description, icon, color, is_default: false });
            router.push(`/admin/courses/${course.id}/visual-editor`);
        } catch (err: any) {
            setError(err?.response?.data?.error || 'Échec de la création');
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32">
            {/* Blobs */}
            <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
                <div className="absolute -top-40 -left-20 w-96 h-96 bg-violet-400/10 rounded-full blur-3xl" />
                <div className="absolute bottom-20 right-10 w-80 h-80 bg-blue-400/8 rounded-full blur-3xl" />
            </div>

            <main className="container max-w-3xl mx-auto px-4 relative z-10 space-y-8">
                {/* Header */}
                <div className="space-y-1">
                    <Button
                        variant="ghost"
                        className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent"
                        onClick={() => mode === 'pick' ? router.push('/admin') : setMode('pick')}
                    >
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        {mode === 'pick' ? 'Retour Admin' : 'Choisir autrement'}
                    </Button>
                    <h1 className="text-4xl font-black tracking-tight">
                        {mode === 'pick' ? 'Créer un Cours' :
                            mode === 'template' ? '✨ Depuis un Modèle' :
                                mode === 'auto' ? '⚡ Auto-Génération' :
                                    '✏️ Création Manuelle'}
                    </h1>
                    <p className="text-muted-foreground font-medium">
                        {mode === 'pick' ? 'Comment voulez-vous créer votre cours ?' :
                            mode === 'template' ? 'Un cours complet en un clic — fiches + quiz générés automatiquement.' :
                                mode === 'auto' ? 'Générez un cours à partir d\'un deck de vocabulaire existant.' :
                                    'Créez un cours vide et construisez-le dans l\'éditeur visuel.'}
                    </p>
                </div>

                {/* Error */}
                <AnimatePresence>
                    {error && (
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
                            className="bg-red-50 border-2 border-red-200 rounded-2xl p-4">
                            <p className="text-red-700 font-bold text-sm">{error}</p>
                        </motion.div>
                    )}
                </AnimatePresence>

                <AnimatePresence mode="wait">
                    {/* ─── PICK MODE ─── */}
                    {mode === 'pick' && (
                        <motion.div
                            key="pick"
                            initial={{ opacity: 0, y: 16 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -16 }}
                            className="grid grid-cols-1 gap-4"
                        >
                            {/* Template — recommended */}
                            <button
                                onClick={() => setMode('template')}
                                className="group relative overflow-hidden text-left w-full rounded-3xl border-2 border-violet-300 bg-gradient-to-br from-violet-50 to-purple-50 dark:from-violet-900/20 dark:to-purple-900/20 hover:border-violet-500 hover:shadow-xl hover:shadow-violet-500/10 transition-all duration-300 p-6"
                            >
                                <div className="absolute top-4 right-4 bg-violet-500 text-white text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-full">
                                    Recommandé
                                </div>
                                <div className="flex items-start gap-5">
                                    <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-lg border-b-4 border-violet-700 shrink-0">
                                        <Sparkles className="w-8 h-8 text-white" strokeWidth={2.5} />
                                    </div>
                                    <div className="flex-1">
                                        <h3 className="text-xl font-black mb-1 group-hover:text-violet-700 transition-colors">
                                            Créer depuis un Modèle
                                        </h3>
                                        <p className="text-muted-foreground font-medium text-sm mb-4">
                                            Un cours pré-structuré avec des fiches de révision et des quiz pour chaque chapitre.
                                        </p>
                                        <div className="flex gap-3 text-xs font-bold text-muted-foreground">
                                            <span className="flex items-center gap-1"><BookOpen className="w-3.5 h-3.5 text-violet-500" /> Fiches de révision</span>
                                            <span className="flex items-center gap-1"><Layers className="w-3.5 h-3.5 text-violet-500" /> Chapitres structurés</span>
                                            <span className="flex items-center gap-1"><Zap className="w-3.5 h-3.5 text-violet-500" /> Quiz auto</span>
                                        </div>
                                    </div>
                                    <ChevronRight className="w-6 h-6 text-violet-400 group-hover:translate-x-1 transition-transform mt-4 shrink-0" />
                                </div>
                            </button>

                            {/* Auto-generate */}
                            <button
                                onClick={() => setMode('auto')}
                                className="group text-left w-full rounded-3xl border-2 border-border bg-card hover:border-blue-400/60 hover:shadow-lg transition-all duration-300 p-6"
                            >
                                <div className="flex items-start gap-5">
                                    <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-600 flex items-center justify-center shadow-lg border-b-4 border-blue-700 shrink-0">
                                        <Wand2 className="w-8 h-8 text-white" strokeWidth={2.5} />
                                    </div>
                                    <div className="flex-1">
                                        <h3 className="text-xl font-black mb-1 group-hover:text-blue-700 transition-colors">
                                            Auto-Génération via Deck
                                        </h3>
                                        <p className="text-muted-foreground font-medium text-sm">
                                            Choisissez un deck de vocabulaire existant et le système crée automatiquement un cours progressif.
                                        </p>
                                    </div>
                                    <ChevronRight className="w-6 h-6 text-muted-foreground group-hover:translate-x-1 transition-transform mt-4 shrink-0" />
                                </div>
                            </button>

                            {/* Manual */}
                            <button
                                onClick={() => setMode('manual')}
                                className="group text-left w-full rounded-3xl border-2 border-border bg-card hover:border-slate-400 hover:shadow-lg transition-all duration-300 p-6"
                            >
                                <div className="flex items-start gap-5">
                                    <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-slate-500 to-slate-600 flex items-center justify-center shadow-lg border-b-4 border-slate-700 shrink-0">
                                        <Pencil className="w-8 h-8 text-white" strokeWidth={2.5} />
                                    </div>
                                    <div className="flex-1">
                                        <h3 className="text-xl font-black mb-1 group-hover:text-slate-700 transition-colors">
                                            Création Manuelle
                                        </h3>
                                        <p className="text-muted-foreground font-medium text-sm">
                                            Partez d'une page blanche et construisez votre cours dans l'éditeur visuel.
                                        </p>
                                    </div>
                                    <ChevronRight className="w-6 h-6 text-muted-foreground group-hover:translate-x-1 transition-transform mt-4 shrink-0" />
                                </div>
                            </button>
                        </motion.div>
                    )}

                    {/* ─── TEMPLATE MODE ─── */}
                    {mode === 'template' && (
                        <motion.div key="template" initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="space-y-4">
                            {loadingTemplates ? (
                                <div className="py-24 flex flex-col items-center gap-4">
                                    <Loader2 className="w-10 h-10 text-violet-500 animate-spin" />
                                    <p className="text-muted-foreground font-bold animate-pulse">Chargement des modèles...</p>
                                </div>
                            ) : templates.length === 0 ? (
                                <div className="py-24 text-center border-2 border-dashed border-border rounded-3xl">
                                    <p className="text-muted-foreground font-bold">Aucun modèle disponible.</p>
                                </div>
                            ) : (
                                templates.map((tmpl, i) => {
                                    const isCreating = creatingTemplate === tmpl.filename;
                                    const isDone = creatingTemplate !== null && creatingTemplate !== tmpl.filename;
                                    return (
                                        <motion.div
                                            key={tmpl.filename}
                                            initial={{ opacity: 0, y: 12 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            transition={{ delay: i * 0.07 }}
                                            className="rounded-3xl border-2 border-border bg-card overflow-hidden"
                                        >
                                            <div className="h-2 bg-gradient-to-r from-violet-500 via-purple-500 to-fuchsia-500" />
                                            <div className="p-6 flex items-start gap-5">
                                                <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0 shadow-lg border-b-4 border-violet-700">
                                                    <GraduationCap className="w-7 h-7 text-white" strokeWidth={2.5} />
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <h3 className="text-lg font-black mb-1">{tmpl.title}</h3>
                                                    <p className="text-sm text-muted-foreground line-clamp-2 mb-3 font-medium">{tmpl.description}</p>
                                                    <div className="flex gap-3 text-[11px] font-bold text-muted-foreground">
                                                        <span className="flex items-center gap-1"><Layers className="w-3 h-3 text-violet-500" />{tmpl.group_count} chapitres</span>
                                                        <span className="flex items-center gap-1"><BookOpen className="w-3 h-3 text-emerald-500" />{tmpl.group_count} fiches</span>
                                                        <span className="flex items-center gap-1"><Zap className="w-3 h-3 text-amber-500" />Quiz auto</span>
                                                    </div>
                                                </div>
                                                <Button
                                                    disabled={isCreating || isDone}
                                                    onClick={() => handleCreateFromTemplate(tmpl)}
                                                    className={cn(
                                                        "shrink-0 font-black px-5 gap-2 rounded-2xl border-b-4 active:border-b-0 active:translate-y-1 transition-all shadow-md",
                                                        "bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white border-violet-700"
                                                    )}
                                                >
                                                    {isCreating ? (
                                                        <><Loader2 className="w-4 h-4 animate-spin" /> Création...</>
                                                    ) : (
                                                        <><Sparkles className="w-4 h-4" /> Créer</>
                                                    )}
                                                </Button>
                                            </div>
                                        </motion.div>
                                    );
                                })
                            )}
                        </motion.div>
                    )}

                    {/* ─── AUTO MODE ─── */}
                    {mode === 'auto' && (
                        <motion.div key="auto" initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
                            <form onSubmit={handleAutoCreate} className="bg-card border-2 border-border rounded-3xl p-8 space-y-6">
                                <div className="space-y-2">
                                    <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Titre du Parcours</Label>
                                    <Input
                                        placeholder="Ex: Mon Super Parcours Auto"
                                        value={autoTitle}
                                        onChange={e => setAutoTitle(e.target.value)}
                                        className="h-12 font-bold"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Deck Source</Label>
                                    {loadingDecks ? (
                                        <p className="text-sm text-muted-foreground animate-pulse py-3">Chargement des Decks...</p>
                                    ) : (
                                        <select
                                            required
                                            className="w-full h-12 px-3 border-2 border-border rounded-xl bg-background font-bold text-sm focus:outline-none focus:border-primary"
                                            value={deckId}
                                            onChange={e => setDeckId(e.target.value)}
                                        >
                                            <option value="">-- Sélectionnez un Deck --</option>
                                            {decks.map(d => (
                                                <option key={d.id} value={d.id}>{d.title} ({d.category_count} catégories)</option>
                                            ))}
                                        </select>
                                    )}
                                </div>
                                <Button
                                    type="submit"
                                    disabled={loading || !deckId}
                                    className="w-full h-14 text-base font-black bg-gradient-to-r from-blue-500 to-cyan-600 hover:from-blue-600 hover:to-cyan-700 text-white rounded-2xl border-b-4 border-blue-700 active:border-b-0 active:translate-y-1 transition-all shadow-lg gap-2"
                                >
                                    {loading ? <><Loader2 className="w-5 h-5 animate-spin" /> Génération...</> : <><Wand2 className="w-5 h-5" /> Générer le Cours</>}
                                </Button>
                            </form>
                        </motion.div>
                    )}

                    {/* ─── MANUAL MODE ─── */}
                    {mode === 'manual' && (
                        <motion.div key="manual" initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
                            <form onSubmit={handleManualCreate} className="bg-card border-2 border-border rounded-3xl p-8 space-y-6">
                                <div className="space-y-2">
                                    <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Titre du Parcours</Label>
                                    <Input required placeholder="Ex: Apprendre l'Arabe" value={title} onChange={e => setTitle(e.target.value)} className="h-12 font-bold" />
                                </div>
                                <div className="space-y-2">
                                    <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Description</Label>
                                    <Input placeholder="Une petite description sympa..." value={description} onChange={e => setDescription(e.target.value)} className="h-12 font-bold" />
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Couleur</Label>
                                        <input type="color" className="h-12 w-full rounded-xl cursor-pointer border-2 border-border" value={color} onChange={e => setColor(e.target.value)} />
                                    </div>
                                    <div className="space-y-2">
                                        <Label className="text-xs font-black uppercase tracking-widest text-muted-foreground">Icône</Label>
                                        <Input placeholder="star, book, flame..." value={icon} onChange={e => setIcon(e.target.value)} className="h-12 font-bold" />
                                    </div>
                                </div>
                                <Button
                                    type="submit"
                                    disabled={loading}
                                    className="w-full h-14 text-base font-black bg-gradient-to-r from-slate-600 to-slate-700 hover:from-slate-700 hover:to-slate-800 text-white rounded-2xl border-b-4 border-slate-800 active:border-b-0 active:translate-y-1 transition-all shadow-lg gap-2"
                                >
                                    {loading ? <><Loader2 className="w-5 h-5 animate-spin" /> Création...</> : <><Pencil className="w-5 h-5" /> Créer et Ouvrir l'Éditeur</>}
                                </Button>
                            </form>
                        </motion.div>
                    )}
                </AnimatePresence>
            </main>
        </div>
    );
}
