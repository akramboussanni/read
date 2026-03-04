'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { classroomApi, Classroom } from '@/lib/api/classroom';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import {
    Users, Plus, ArrowRight, User,
    Settings, Key, Trash2, GraduationCap,
    Calendar, Info, AlertCircle
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

export default function ClassesPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [classes, setClasses] = useState<{ teaching: Classroom[], enrolled: Classroom[] }>({ teaching: [], enrolled: [] });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadClasses();
    }, []);

    const loadClasses = async () => {
        setLoading(true);
        try {
            const data = await classroomApi.listMyClasses();
            setClasses({
                teaching: data?.teaching || [],
                enrolled: data?.enrolled || []
            });
        } catch (err) {
            console.error('Failed to load classes:', err);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center pt-20">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                    <p className="text-primary font-black text-sm tracking-widest animate-pulse">CHARGEMENT DE TES CLASSES...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32">
            <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
                <div className="blob-green -top-40 -left-20 opacity-30" />
                <div className="blob-orange bottom-20 right-10 opacity-20" />
            </div>

            <main className="container max-w-5xl mx-auto px-4 relative z-10 space-y-12">
                <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} className="space-y-2">
                    <h1 className="text-4xl font-black tracking-tight flex items-center gap-4">
                        <div className="w-12 h-12 bg-accent/20 rounded-2xl flex items-center justify-center text-teal-600 shadow-sm">
                            <Users className="w-7 h-7" strokeWidth={2.5} />
                        </div>
                        Mes Classes
                    </h1>
                    <p className="text-muted-foreground font-semibold text-lg">Apprends en groupe, partage ton progrès et relève les défis de tes enseignants.</p>
                </motion.div>

                <div className={cn(
                    "grid gap-8",
                    user?.role === 'teacher' ? "grid-cols-1 md:grid-cols-2" : "grid-cols-1 max-w-2xl mx-auto"
                )}>
                    {/* TEACHING SECTION */}
                    {user?.role === 'teacher' && (
                        <div className="space-y-6">
                            <div className="flex items-center justify-between">
                                <h2 className="text-xl font-black flex items-center gap-2">
                                    <Plus className="w-5 h-5 text-primary" />
                                    Je suis Enseignant
                                </h2>
                                <Button onClick={() => router.push('/classes/create')} variant="outline" className="rounded-xl border-primary/20 text-primary font-bold hover:bg-primary/5">
                                    Créer une classe
                                </Button>
                            </div>

                            <div className="space-y-4">
                                {classes?.teaching?.length > 0 ? (
                                    classes.teaching.map((cls, i) => (
                                        <ClassCard key={cls.id} class={cls} type="teaching" delay={i * 0.1} />
                                    ))
                                ) : (
                                    <EmptySection
                                        title="Tu n'enseignes aucune classe"
                                        desc="Crée ta première classe pour inviter tes élèves et leur donner des devoirs."
                                    />
                                )}
                            </div>
                        </div>
                    )}

                    {/* ENROLLED SECTION */}
                    <div className="space-y-6">
                        <div className="flex items-center justify-between">
                            <h2 className="text-xl font-black flex items-center gap-2">
                                <GraduationCap className="w-5 h-5 text-secondary" />
                                {user?.role === 'teacher' ? "Je suis Élève" : "Mes Classes"}
                            </h2>
                            <Button onClick={() => router.push('/classes/join')} className="bg-secondary text-white rounded-xl border-b-4 border-orange-600 font-bold hover:bg-orange-500">
                                Rejoindre une classe
                            </Button>
                        </div>

                        <div className="space-y-4">
                            {classes?.enrolled?.length > 0 ? (
                                classes.enrolled.map((cls, i) => (
                                    <ClassCard key={cls.id} class={cls} type="enrolled" delay={i * 0.1} />
                                ))
                            ) : (
                                <EmptySection
                                    title="Tu n'es dans aucune classe"
                                    desc="Demande un code à ton enseignant pour rejoindre sa classe et suivre ses parcours."
                                />
                            )}
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}

function ClassCard({ class: cls, type, delay }: { class: Classroom, type: 'teaching' | 'enrolled', delay: number }) {
    const router = useRouter();
    return (
        <motion.div
            initial={{ opacity: 0, x: type === 'teaching' ? -20 : 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay }}
            onClick={() => router.push(`/classes/${cls.id}`)}
            className={cn(
                "group relative bg-card rounded-2xl border-2 p-5 cursor-pointer transition-all hover:shadow-lg active:scale-[0.98]",
                type === 'teaching' ? "border-primary/20 hover:border-primary/50" : "border-secondary/20 hover:border-secondary/50"
            )}
        >
            <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-4">
                    <div className={cn(
                        "w-14 h-14 rounded-xl flex items-center justify-center font-black text-xl border-b-4 shadow-sm group-hover:-translate-y-1 transition-transform",
                        type === 'teaching' ? "bg-primary text-white border-primary-hover" : "bg-secondary text-white border-orange-600"
                    )}>
                        {cls.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                        <h3 className="text-lg font-black group-hover:text-primary transition-colors">{cls.name}</h3>
                        <p className="text-sm text-muted-foreground line-clamp-1">{cls.description || 'Aucune description'}</p>
                    </div>
                </div>
                <ArrowRight className="w-5 h-5 text-muted-foreground group-hover:text-foreground group-hover:translate-x-1 transition-all" />
            </div>

            <div className="mt-4 pt-4 border-t border-muted flex items-center justify-between">
                {type === 'teaching' ? (
                    <div className="flex items-center gap-3">
                        <div className="bg-muted px-2 py-1 rounded-lg flex items-center gap-1.5">
                            <Key className="w-3.5 h-3.5 text-primary" />
                            <span className="text-xs font-black tracking-widest uppercase">{cls.join_code}</span>
                        </div>
                        <span className="text-xs font-bold text-muted-foreground">Appuyer pour gérer</span>
                    </div>
                ) : (
                    <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-muted-foreground italic">Relève les défis !</span>
                    </div>
                )}

                {cls.is_locked && (
                    <div className="text-xs font-bold text-red-500 flex items-center gap-1">
                        <AlertCircle className="w-3 h-3" /> Verrouillé
                    </div>
                )}
            </div>
        </motion.div>
    );
}

function EmptySection({ title, desc }: { title: string, desc: string }) {
    return (
        <div className="p-10 text-center border-2 border-dashed border-border rounded-2xl bg-muted/20">
            <h3 className="font-bold text-foreground mb-2">{title}</h3>
            <p className="text-sm text-muted-foreground max-w-[240px] mx-auto">{desc}</p>
        </div>
    );
}
