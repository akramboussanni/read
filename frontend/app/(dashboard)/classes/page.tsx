'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { classroomApi, Classroom } from '@/lib/api/classroom';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Users, Plus, GraduationCap } from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { ClassCard } from '@/components/dashboard/class-card';
import { EmptySection } from '@/components/dashboard/empty-section';

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


