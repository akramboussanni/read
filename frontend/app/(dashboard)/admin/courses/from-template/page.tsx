'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { listTemplates, createFromTemplate } from '@/lib/api/admin';
import type { TemplateInfo } from '@/lib/types/admin';
import { Button } from '@/components/ui/button';
import {
    ArrowLeft, BookOpen, Sparkles, Loader2, CheckCircle2,
    FileText, Layers, Zap, GraduationCap,
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

export default function FromTemplatePage() {
    const router = useRouter();
    const [templates, setTemplates] = useState<TemplateInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [creating, setCreating] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        loadTemplates();
    }, []);

    const loadTemplates = async () => {
        try {
            const data = await listTemplates();
            setTemplates(data || []);
        } catch (err) {
            console.error('Failed to load templates', err);
            setError('Impossible de charger les modèles');
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async (template: TemplateInfo) => {
        setCreating(template.filename);
        setError(null);
        try {
            const course = await createFromTemplate({
                template_filename: template.filename,
            });
            setSuccess(course.id);
            setTimeout(() => {
                router.push(`/admin/courses/${course.id}/visual-editor`);
            }, 1500);
        } catch (err: any) {
            console.error('Failed to create course from template', err);
            setError(err?.response?.data || 'Erreur lors de la création du cours');
            setCreating(null);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                    <p className="text-primary font-bold text-sm tracking-widest animate-pulse">CHARGEMENT DES MODÈLES...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32">
            {/* Background */}
            <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
                <div className="absolute -top-40 -left-20 w-96 h-96 bg-violet-400/10 rounded-full blur-3xl" />
                <div className="absolute bottom-20 right-10 w-80 h-80 bg-amber-400/10 rounded-full blur-3xl" />
            </div>

            <main className="container max-w-4xl mx-auto px-4 relative z-10 space-y-8">
                {/* Header */}
                <motion.div initial={{ opacity: 0, y: -16 }} animate={{ opacity: 1, y: 0 }} className="space-y-1">
                    <Button
                        variant="ghost"
                        className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent"
                        onClick={() => router.push('/admin')}
                    >
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        Retour Admin
                    </Button>
                    <h1 className="text-4xl font-black tracking-tight flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-violet-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg shadow-violet-500/20">
                            <Sparkles className="w-7 h-7 text-white" strokeWidth={2.5} />
                        </div>
                        Créer depuis un Modèle
                    </h1>
                    <p className="text-muted-foreground font-medium max-w-xl">
                        Choisissez un modèle de cours pré-configuré. Le cours sera créé avec des fiches de révision et des quiz pour chaque chapitre.
                    </p>
                </motion.div>

                {/* Error */}
                <AnimatePresence>
                    {error && (
                        <motion.div
                            initial={{ opacity: 0, y: -8 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0 }}
                            className="bg-red-50 dark:bg-red-900/20 border-2 border-red-200 dark:border-red-800 rounded-2xl p-4"
                        >
                            <p className="text-red-700 dark:text-red-300 font-bold text-sm">{error}</p>
                        </motion.div>
                    )}
                </AnimatePresence>

                {/* Templates Grid */}
                {templates.length === 0 ? (
                    <div className="py-24 text-center space-y-4 border-2 border-dashed border-border rounded-3xl bg-muted/20">
                        <FileText className="w-12 h-12 text-muted-foreground mx-auto" />
                        <h3 className="text-xl font-black">Aucun modèle disponible</h3>
                        <p className="text-muted-foreground font-medium max-w-sm mx-auto">
                            Les modèles de cours sont intégrés au serveur. Assurez-vous que les fichiers JSON sont en place.
                        </p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 gap-5">
                        {templates.map((template, i) => {
                            const isCreating = creating === template.filename;
                            const isSuccess = success !== null && creating === template.filename;

                            return (
                                <motion.div
                                    key={template.filename}
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ delay: i * 0.08 }}
                                    className={cn(
                                        "group relative overflow-hidden rounded-3xl border-2 bg-card transition-all duration-300",
                                        isSuccess
                                            ? "border-green-400 shadow-lg shadow-green-500/10"
                                            : "border-border hover:border-violet-400/50 hover:shadow-xl"
                                    )}
                                >
                                    {/* Gradient strip */}
                                    <div className="h-2 w-full bg-gradient-to-r from-violet-500 via-purple-500 to-fuchsia-500" />

                                    <div className="p-6 flex items-start gap-5">
                                        {/* Icon */}
                                        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0 shadow-lg border-b-4 border-violet-700">
                                            <GraduationCap className="w-8 h-8 text-white" strokeWidth={2.5} />
                                        </div>

                                        {/* Content */}
                                        <div className="flex-1 min-w-0">
                                            <h3 className="text-xl font-black mb-1 group-hover:text-violet-600 transition-colors">
                                                {template.title}
                                            </h3>
                                            <p className="text-sm text-muted-foreground line-clamp-2 mb-4 font-medium">
                                                {template.description}
                                            </p>

                                            {/* Stats */}
                                            <div className="flex items-center gap-4 mb-4">
                                                <div className="flex items-center gap-1.5 text-xs font-bold text-muted-foreground">
                                                    <Layers className="w-3.5 h-3.5 text-violet-500" />
                                                    {template.group_count} chapitres
                                                </div>
                                                <div className="flex items-center gap-1.5 text-xs font-bold text-muted-foreground">
                                                    <BookOpen className="w-3.5 h-3.5 text-blue-500" />
                                                    {template.group_count} fiches de révision
                                                </div>
                                                <div className="flex items-center gap-1.5 text-xs font-bold text-muted-foreground">
                                                    <Zap className="w-3.5 h-3.5 text-amber-500" />
                                                    Quiz auto-générés
                                                </div>
                                            </div>

                                            {/* Tag */}
                                            <div className="flex items-center gap-2">
                                                <span className="text-[10px] font-bold uppercase tracking-widest text-violet-600 bg-violet-50 dark:bg-violet-900/30 px-2.5 py-1 rounded-full">
                                                    Deck : {template.deck_key}
                                                </span>
                                            </div>
                                        </div>

                                        {/* Action */}
                                        <div className="shrink-0 flex items-center">
                                            {isSuccess ? (
                                                <div className="flex items-center gap-2 text-green-600 font-bold animate-pulse">
                                                    <CheckCircle2 className="w-6 h-6" />
                                                    <span className="text-sm">Créé !</span>
                                                </div>
                                            ) : (
                                                <Button
                                                    size="lg"
                                                    disabled={isCreating || creating !== null}
                                                    onClick={() => handleCreate(template)}
                                                    className={cn(
                                                        "font-black px-6 gap-2 rounded-2xl border-b-4 active:border-b-0 active:translate-y-1 transition-all shadow-lg",
                                                        "bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white border-violet-700"
                                                    )}
                                                >
                                                    {isCreating ? (
                                                        <>
                                                            <Loader2 className="w-5 h-5 animate-spin" />
                                                            Création...
                                                        </>
                                                    ) : (
                                                        <>
                                                            <Sparkles className="w-5 h-5" />
                                                            Créer le Cours
                                                        </>
                                                    )}
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                </motion.div>
                            );
                        })}
                    </div>
                )}
            </main>
        </div>
    );
}
