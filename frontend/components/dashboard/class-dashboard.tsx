'use client';

import { Course } from '@/lib/types/course';
import { ClassroomAssignment } from '@/lib/api/classroom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ClassAssignmentCard } from './class-assignment-card';
import { Calendar, ChevronRight, CheckCircle2, Flame, Users } from 'lucide-react';
import { motion } from 'framer-motion';

interface ClassDashboardData {
    classroom: { name: string };
    assignments?: ClassroomAssignment[];
    students?: { id: string; username: string }[];
}

interface ClassDashboardProps {
    classData: ClassDashboardData;
    activeCourse: Course | null;
    progressPct: number;
    onViewCourse: () => void;
}

export function ClassDashboard({ classData, activeCourse, progressPct, onViewCourse }: ClassDashboardProps) {
    return (
        <div className="flex-1 overflow-y-auto w-full max-w-7xl mx-auto px-4 lg:px-8 py-8 space-y-8">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <div className="space-y-1">
                    <p className="text-sm font-black text-primary uppercase tracking-widest flex items-center gap-2">
                        <Users className="w-4 h-4" /> Ma Classe
                    </p>
                    <h1 className="text-4xl font-black tracking-tight">{classData.classroom.name}</h1>
                    <p className="text-muted-foreground font-semibold">
                        Voici ton tableau de bord. Regarde tes devoirs et continue ton parcours !
                    </p>
                </div>

                {activeCourse && (
                    <Button
                        onClick={onViewCourse}
                        className="shrink-0 h-14 px-6 text-lg font-black bg-primary text-white rounded-2xl border-b-4 border-primary-hover hover:-translate-y-1 active:translate-y-0 transition-all flex items-center gap-2"
                    >
                        Parcours: {activeCourse.title} <ChevronRight className="w-5 h-5" />
                    </Button>
                )}
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Main Content: Assignments */}
                <div className="lg:col-span-2 space-y-6">
                    <h2 className="text-2xl font-black flex items-center gap-3">
                        <Calendar className="w-6 h-6 text-primary" /> Devoirs à faire
                    </h2>

                    <div className="space-y-4">
                        {classData.assignments && classData.assignments.length > 0 ? (
                            classData.assignments.map((asgn) => (
                                <ClassAssignmentCard key={asgn.id} assignment={asgn} isTeacher={false} />
                            ))
                        ) : (
                            <div className="py-16 text-center border-2 border-dashed border-border rounded-3xl bg-muted/10">
                                <CheckCircle2 className="w-12 h-12 text-green-500 mx-auto mb-4 opacity-50" />
                                <h3 className="text-xl font-bold text-slate-500">Aucun devoir en cours</h3>
                                <p className="text-sm text-muted-foreground font-medium">Bon travail, tu es à jour !</p>
                            </div>
                        )}
                    </div>
                </div>

                {/* Sidebar: Ranking & Active Course */}
                <div className="space-y-6">
                    {activeCourse && (
                        <Card className="fun-card border-primary/20 bg-primary/5">
                            <CardHeader className="pb-3 border-b border-primary/10 bg-white">
                                <CardTitle className="text-sm font-black uppercase tracking-widest text-primary flex items-center gap-2">
                                    <Flame className="w-4 h-4" /> Mon Parcours
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="p-5 space-y-4 bg-white/50">
                                <div className="flex justify-between items-center text-sm font-bold">
                                    <span className="text-slate-600">Progression</span>
                                    <span className="text-primary tabular-nums">{progressPct}%</span>
                                </div>
                                <div className="w-full h-3 bg-muted rounded-full overflow-hidden">
                                    <motion.div
                                        initial={{ width: 0 }}
                                        animate={{ width: `${progressPct}%` }}
                                        transition={{ duration: 0.8 }}
                                        className="h-full rounded-full bg-primary"
                                    />
                                </div>
                                <Button
                                    onClick={onViewCourse}
                                    className="w-full font-black text-sm bg-white border-2 border-primary text-primary hover:bg-primary hover:text-white transition-colors h-12 rounded-xl"
                                >
                                    Voir la carte du parcours
                                </Button>
                            </CardContent>
                        </Card>
                    )}

                </div>
            </div>
        </div>
    );
}
