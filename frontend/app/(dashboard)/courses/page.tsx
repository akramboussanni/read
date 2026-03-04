'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { courseApi } from '@/lib/api/course';
import { Course, UserEnrollment } from '@/lib/types/course';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/lib/store/auth-store';
import {
    BookOpen, Star, Plus, Check, ChevronRight, Flame, Trophy,
    Compass, GraduationCap, ArrowLeft, Play, Loader2, Sparkles,
    RefreshCw, CheckCircle2, Clock, BarChart3,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';

type Tab = 'browse' | 'mine';

export default function CoursesPage() {
    const router = useRouter();
    const { user } = useAuthStore();

    const [tab, setTab] = useState<Tab>('browse');
    const [allCourses, setAllCourses] = useState<Course[]>([]);
    const [enrollments, setEnrollments] = useState<UserEnrollment[]>([]);
    const [loading, setLoading] = useState(true);
    const [enrollingId, setEnrollingId] = useState<string | null>(null);
    const [switchingId, setSwitchingId] = useState<string | null>(null);
    const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);

    const activeCourseId = user?.active_course_id;

    useEffect(() => {
        loadAll();
    }, []);

    const loadAll = async () => {
        setLoading(true);
        try {
            const [courses, myEnrollments] = await Promise.all([
                courseApi.listCourses().catch(() => []),
                courseApi.getMyEnrollments().catch(() => []),
            ]);
            setAllCourses(courses ?? []);
            setEnrollments(myEnrollments ?? []);
        } finally {
            setLoading(false);
        }
    };

    const enrolledIds = new Set((enrollments ?? []).map(e => e.course_id));

    const handleEnroll = async (courseId: string) => {
        setEnrollingId(courseId);
        try {
            await courseApi.enroll(courseId);
            await courseApi.setActiveCourse(courseId);
            await loadAll();
            // Update auth store user
            window.location.reload(); // Reload to refresh active_course_id in user store
        } catch (err) {
            console.error('Enrollment failed', err);
        } finally {
            setEnrollingId(null);
        }
    };

    const handleSetActive = async (courseId: string) => {
        setSwitchingId(courseId);
        try {
            await courseApi.setActiveCourse(courseId);
            window.location.reload();
        } catch (err) {
            console.error('Failed to set active course', err);
            setSwitchingId(null);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center pt-20">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                    <p className="text-primary font-bold text-sm tracking-widest animate-pulse">CHARGEMENT DES PARCOURS...</p>
                </div>
            </div>
        );
    }

    const myCourses = allCourses.filter(c => enrolledIds.has(c.id));
    const browseCourses = allCourses;

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32">
            {/* Background blobs */}
            <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
                <div className="blob-green -top-40 -left-20 opacity-30" />
                <div className="blob-orange bottom-20 right-10 opacity-20" />
            </div>

            <main className="container max-w-5xl mx-auto px-4 relative z-10 space-y-8">

                {/* Header */}
                <motion.div initial={{ opacity: 0, y: -16 }} animate={{ opacity: 1, y: 0 }} className="space-y-1">
                    <Button variant="ghost" className="text-muted-foreground hover:text-foreground pl-0 hover:bg-transparent" onClick={() => router.push('/')}>
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        Tableau de Bord
                    </Button>
                    <h1 className="text-4xl font-black tracking-tight flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary rounded-2xl flex items-center justify-center shadow-lg shadow-primary/20">
                            <GraduationCap className="w-7 h-7 text-white" strokeWidth={2.5} />
                        </div>
                        Mes Parcours
                    </h1>
                    <p className="text-muted-foreground font-medium">Découvre, inscris-toi et change ton parcours actif à tout moment.</p>
                </motion.div>

                {/* Active Course Banner */}
                <AnimatePresence>
                    {activeCourseId && (() => {
                        const active = allCourses.find(c => c.id === activeCourseId);
                        if (!active) return null;
                        return (
                            <motion.div
                                initial={{ opacity: 0, scale: 0.97 }}
                                animate={{ opacity: 1, scale: 1 }}
                                className="relative overflow-hidden rounded-3xl border-2 border-primary/30 bg-gradient-to-br from-primary/10 via-primary/5 to-transparent p-6 flex items-center gap-6"
                            >
                                <div className="absolute top-0 right-0 w-48 h-48 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-2xl" />
                                <div
                                    className="w-16 h-16 rounded-2xl flex items-center justify-center shrink-0 shadow-lg border-b-4"
                                    style={{ background: active.color || '#6C5CE7', borderColor: `${active.color || '#6C5CE7'}99` }}
                                >
                                    <BookOpen className="w-8 h-8 text-white" strokeWidth={2.5} />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className="text-xs font-bold uppercase tracking-widest text-primary bg-primary/10 px-2 py-0.5 rounded-full flex items-center gap-1">
                                            <Flame className="w-3 h-3" /> En cours
                                        </span>
                                    </div>
                                    <h2 className="text-xl font-black truncate">{active.title}</h2>
                                    <p className="text-sm text-muted-foreground line-clamp-1">{active.description || 'Aucune description'}</p>
                                </div>
                                <Button onClick={() => router.push('/')} className="shrink-0 bg-primary text-white font-bold shadow-lg shadow-primary/20 gap-2">
                                    <Play className="w-4 h-4 fill-white" />
                                    Continuer
                                </Button>
                            </motion.div>
                        );
                    })()}
                </AnimatePresence>

                {/* Tabs */}
                <div className="flex items-center gap-1 bg-muted p-1 rounded-2xl w-fit border border-border">
                    {([
                        { key: 'browse', label: 'Explorer', icon: Compass },
                        { key: 'mine', label: 'Mes Inscriptions', icon: GraduationCap },
                    ] as { key: Tab; label: string; icon: any }[]).map(({ key, label, icon: Icon }) => (
                        <button
                            key={key}
                            onClick={() => { setTab(key); setSelectedCourse(null); }}
                            className={cn(
                                'flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all',
                                tab === key
                                    ? 'bg-white text-foreground shadow-sm border border-border'
                                    : 'text-muted-foreground hover:text-foreground'
                            )}
                        >
                            <Icon className="w-4 h-4" />
                            {label}
                            {key === 'mine' && enrollments.length > 0 && (
                                <span className="ml-1 bg-primary text-white text-[10px] font-black w-5 h-5 rounded-full flex items-center justify-center">
                                    {enrollments.length}
                                </span>
                            )}
                        </button>
                    ))}
                </div>

                {/* Content */}
                <AnimatePresence mode="wait">
                    {tab === 'browse' && (
                        <motion.div key="browse" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
                            {browseCourses.length === 0 ? (
                                <EmptyState
                                    icon={<Compass className="w-12 h-12 text-muted-foreground" />}
                                    title="Aucun parcours disponible"
                                    description="Revenez bientôt, de nouveaux parcours arrivent !"
                                />
                            ) : (
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                                    {browseCourses.map((course, i) => (
                                        <CourseCard
                                            key={course.id}
                                            course={course}
                                            index={i}
                                            isEnrolled={enrolledIds.has(course.id)}
                                            isActive={course.id === activeCourseId}
                                            isEnrolling={enrollingId === course.id}
                                            isSwitching={switchingId === course.id}
                                            onEnroll={() => handleEnroll(course.id)}
                                            onSetActive={() => handleSetActive(course.id)}
                                            onDetails={() => setSelectedCourse(course)}
                                        />
                                    ))}
                                </div>
                            )}
                        </motion.div>
                    )}

                    {tab === 'mine' && (
                        <motion.div key="mine" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
                            {myCourses.length === 0 ? (
                                <EmptyState
                                    icon={<GraduationCap className="w-12 h-12 text-muted-foreground" />}
                                    title="Aucune inscription"
                                    description="Explore les parcours disponibles et inscris-toi pour commencer !"
                                    action={<Button onClick={() => setTab('browse')} className="gap-2"><Compass className="w-4 h-4" />Explorer les Parcours</Button>}
                                />
                            ) : (
                                <div className="space-y-4">
                                    <p className="text-sm font-semibold text-muted-foreground">
                                        Clique sur "Activer" pour changer ton parcours actif.
                                    </p>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        {myCourses.map((course, i) => {
                                            const enrollment = enrollments.find(e => e.course_id === course.id);
                                            return (
                                                <EnrolledCourseCard
                                                    key={course.id}
                                                    course={course}
                                                    enrollment={enrollment}
                                                    index={i}
                                                    isActive={course.id === activeCourseId}
                                                    isSwitching={switchingId === course.id}
                                                    onSetActive={() => handleSetActive(course.id)}
                                                    onPlay={() => router.push('/')}
                                                />
                                            );
                                        })}
                                    </div>
                                </div>
                            )}
                        </motion.div>
                    )}
                </AnimatePresence>
            </main>

            {/* Course Detail Drawer */}
            <AnimatePresence>
                {selectedCourse && (
                    <CourseDetailDrawer
                        course={selectedCourse}
                        isEnrolled={enrolledIds.has(selectedCourse.id)}
                        isActive={selectedCourse.id === activeCourseId}
                        isEnrolling={enrollingId === selectedCourse.id}
                        isSwitching={switchingId === selectedCourse.id}
                        onClose={() => setSelectedCourse(null)}
                        onEnroll={() => handleEnroll(selectedCourse.id)}
                        onSetActive={() => handleSetActive(selectedCourse.id)}
                    />
                )}
            </AnimatePresence>
        </div>
    );
}

// ==========================================
// COURSE CARD (browse)
// ==========================================
function CourseCard({ course, index, isEnrolled, isActive, isEnrolling, isSwitching, onEnroll, onSetActive, onDetails }: {
    course: Course;
    index: number;
    isEnrolled: boolean;
    isActive: boolean;
    isEnrolling: boolean;
    isSwitching: boolean;
    onEnroll: () => void;
    onSetActive: () => void;
    onDetails: () => void;
}) {
    const color = course.color || '#6C5CE7';

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
            className={cn(
                'group relative overflow-hidden rounded-3xl border-2 bg-card transition-all duration-300 flex flex-col cursor-pointer hover:shadow-xl',
                isActive ? 'border-primary shadow-lg shadow-primary/10' : 'border-border hover:border-primary/40'
            )}
            onClick={onDetails}
        >
            {/* Color strip */}
            <div className="h-2 w-full" style={{ background: `linear-gradient(90deg, ${color}, ${color}88)` }} />

            {/* Active badge */}
            {isActive && (
                <div className="absolute top-4 right-4 flex items-center gap-1 bg-primary text-white text-[10px] font-black px-2.5 py-1 rounded-full">
                    <Flame className="w-2.5 h-2.5" /> ACTIF
                </div>
            )}

            <div className="p-5 flex-1 flex flex-col">
                {/* Icon */}
                <div
                    className="w-12 h-12 rounded-2xl flex items-center justify-center mb-4 shadow-md border-b-4"
                    style={{ background: color, borderColor: `${color}88` }}
                >
                    <BookOpen className="w-6 h-6 text-white" strokeWidth={2.5} />
                </div>

                <h3 className="text-lg font-black mb-1 line-clamp-1 group-hover:text-primary transition-colors">{course.title}</h3>
                <p className="text-sm text-muted-foreground line-clamp-2 mb-4 flex-1">{course.description || 'Aucune description'}</p>

                {/* CTA */}
                <div className="flex gap-2" onClick={e => e.stopPropagation()}>
                    {!isEnrolled ? (
                        <Button
                            size="sm"
                            className="flex-1 font-bold bg-primary text-white gap-1.5"
                            onClick={onEnroll}
                            disabled={isEnrolling}
                        >
                            {isEnrolling ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Plus className="w-3.5 h-3.5" />}
                            S'inscrire
                        </Button>
                    ) : isActive ? (
                        <Button size="sm" variant="outline" className="flex-1 font-bold border-primary text-primary gap-1.5" onClick={() => { }}>
                            <CheckCircle2 className="w-3.5 h-3.5" /> Actif
                        </Button>
                    ) : (
                        <Button
                            size="sm"
                            variant="outline"
                            className="flex-1 font-bold gap-1.5 hover:border-primary hover:text-primary"
                            onClick={onSetActive}
                            disabled={isSwitching}
                        >
                            {isSwitching ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
                            Activer
                        </Button>
                    )}
                    <Button size="sm" variant="ghost" className="px-2 text-muted-foreground hover:text-foreground" onClick={onDetails}>
                        <ChevronRight className="w-4 h-4" />
                    </Button>
                </div>
            </div>
        </motion.div>
    );
}

// ==========================================
// ENROLLED COURSE CARD (my courses)
// ==========================================
function EnrolledCourseCard({ course, enrollment, index, isActive, isSwitching, onSetActive, onPlay }: {
    course: Course;
    enrollment?: UserEnrollment;
    index: number;
    isActive: boolean;
    isSwitching: boolean;
    onSetActive: () => void;
    onPlay: () => void;
}) {
    const color = course.color || '#6C5CE7';
    const progress = enrollment?.progress ?? 0;

    return (
        <motion.div
            initial={{ opacity: 0, x: -16 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.06 }}
            className={cn(
                'relative overflow-hidden rounded-3xl border-2 bg-card p-5 flex items-center gap-4 transition-all',
                isActive ? 'border-primary shadow-lg shadow-primary/10' : 'border-border'
            )}
        >
            {/* Icon */}
            <div
                className="w-14 h-14 rounded-2xl flex items-center justify-center shrink-0 shadow-md border-b-4"
                style={{ background: color, borderColor: `${color}88` }}
            >
                <BookOpen className="w-7 h-7 text-white" strokeWidth={2.5} />
            </div>

            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-0.5">
                    <h3 className="font-black text-base truncate">{course.title}</h3>
                    {isActive && (
                        <span className="shrink-0 flex items-center gap-1 bg-primary/10 text-primary text-[10px] font-black px-2 py-0.5 rounded-full">
                            <Flame className="w-2.5 h-2.5" /> EN COURS
                        </span>
                    )}
                </div>

                {/* Progress bar */}
                <div className="flex items-center gap-2 mt-1">
                    <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                        <div
                            className="h-full rounded-full transition-all duration-500"
                            style={{ width: `${progress}%`, background: color }}
                        />
                    </div>
                    <span className="text-xs font-bold text-muted-foreground shrink-0">{Math.round(progress)}%</span>
                </div>
            </div>

            {/* Actions */}
            <div className="flex flex-col gap-2 shrink-0">
                {isActive ? (
                    <Button size="sm" onClick={onPlay} className="font-bold bg-primary text-white gap-1.5 text-xs">
                        <Play className="w-3 h-3 fill-white" /> Jouer
                    </Button>
                ) : (
                    <Button
                        size="sm"
                        variant="outline"
                        onClick={onSetActive}
                        disabled={isSwitching}
                        className="font-bold gap-1.5 text-xs hover:border-primary hover:text-primary"
                    >
                        {isSwitching ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
                        Activer
                    </Button>
                )}
            </div>
        </motion.div>
    );
}

// ==========================================
// COURSE DETAIL DRAWER
// ==========================================
function CourseDetailDrawer({ course, isEnrolled, isActive, isEnrolling, isSwitching, onClose, onEnroll, onSetActive }: {
    course: Course;
    isEnrolled: boolean;
    isActive: boolean;
    isEnrolling: boolean;
    isSwitching: boolean;
    onClose: () => void;
    onEnroll: () => void;
    onSetActive: () => void;
}) {
    const color = course.color || '#6C5CE7';
    const nodeCount = course.nodes?.length ?? 0;

    return (
        <>
            {/* Backdrop */}
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40"
                onClick={onClose}
            />
            {/* Panel */}
            <motion.div
                initial={{ x: '100%' }}
                animate={{ x: 0 }}
                exit={{ x: '100%' }}
                transition={{ type: 'spring', damping: 28, stiffness: 260 }}
                className="fixed right-0 top-0 h-full w-full max-w-md bg-card border-l-2 border-border shadow-2xl z-50 flex flex-col overflow-hidden"
            >
                {/* Hero */}
                <div className="relative h-48 flex items-end p-6 shrink-0" style={{ background: `linear-gradient(135deg, ${color}, ${color}cc)` }}>
                    <div className="absolute inset-0 opacity-10" style={{ backgroundImage: 'radial-gradient(circle at 80% 20%, white 1px, transparent 1px)', backgroundSize: '24px 24px' }} />
                    <button onClick={onClose} className="absolute top-4 right-4 w-8 h-8 rounded-full bg-black/20 text-white flex items-center justify-center hover:bg-black/30 transition-colors">
                        ✕
                    </button>
                    <div>
                        <h2 className="text-2xl font-black text-white leading-tight">{course.title}</h2>
                        {isActive && (
                            <span className="inline-flex items-center gap-1 mt-1 bg-white/20 text-white text-xs font-bold px-2 py-0.5 rounded-full">
                                <Flame className="w-3 h-3" /> Parcours Actif
                            </span>
                        )}
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6">
                    {/* Description */}
                    <div>
                        <h3 className="text-xs font-bold uppercase text-muted-foreground tracking-wider mb-2">Description</h3>
                        <p className="text-foreground font-medium leading-relaxed">{course.description || 'Aucune description disponible.'}</p>
                    </div>

                    {/* Stats */}
                    <div className="grid grid-cols-3 gap-3">
                        {[
                            { label: 'Nœuds', value: nodeCount || '—', icon: BarChart3 },
                            { label: 'Statut', value: isEnrolled ? 'Inscrit' : 'Disponible', icon: CheckCircle2 },
                            { label: 'Type', value: course.is_default ? 'Défaut' : 'Personnalisé', icon: Sparkles },
                        ].map(({ label, value, icon: Icon }) => (
                            <div key={label} className="bg-muted/50 rounded-2xl p-3 text-center border border-border">
                                <Icon className="w-4 h-4 mx-auto mb-1 text-muted-foreground" />
                                <div className="text-sm font-black">{value}</div>
                                <div className="text-[10px] text-muted-foreground font-semibold">{label}</div>
                            </div>
                        ))}
                    </div>

                    {/* Nodes preview */}
                    {course.nodes && course.nodes.length > 0 && (
                        <div>
                            <h3 className="text-xs font-bold uppercase text-muted-foreground tracking-wider mb-3">Aperçu du Parcours</h3>
                            <div className="space-y-2">
                                {[...course.nodes].sort((a, b) => a.sort_order - b.sort_order).slice(0, 6).map((node, i) => {
                                    const typeColor: Record<string, string> = {
                                        quiz: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
                                        lesson: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
                                        milestone: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
                                        start: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
                                        checkpoint: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
                                    };
                                    return (
                                        <div key={node.id} className="flex items-center gap-3 py-2 px-3 rounded-xl bg-muted/30 border border-border/50">
                                            <span className="w-5 h-5 rounded-full bg-muted flex items-center justify-center text-xs font-black text-muted-foreground shrink-0">{i + 1}</span>
                                            <span className="font-semibold text-sm flex-1 truncate">{node.title}</span>
                                            <span className={cn('text-[10px] font-bold px-2 py-0.5 rounded-full capitalize shrink-0', typeColor[node.node_type] || 'bg-muted text-muted-foreground')}>
                                                {node.node_type}
                                            </span>
                                        </div>
                                    );
                                })}
                                {course.nodes.length > 6 && (
                                    <p className="text-xs text-center text-muted-foreground pt-1">+ {course.nodes.length - 6} nœuds supplémentaires</p>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer CTA */}
                <div className="p-6 border-t border-border shrink-0">
                    {!isEnrolled ? (
                        <Button className="w-full h-12 text-base font-black bg-primary text-white gap-2 shadow-lg shadow-primary/20" onClick={onEnroll} disabled={isEnrolling}>
                            {isEnrolling ? <Loader2 className="w-5 h-5 animate-spin" /> : <Plus className="w-5 h-5" />}
                            {isEnrolling ? 'Inscription...' : "S'inscrire et Commencer"}
                        </Button>
                    ) : isActive ? (
                        <Button className="w-full h-12 text-base font-black gap-2" onClick={() => window.location.href = '/'}>
                            <Play className="w-5 h-5 fill-white" /> Continuer le Parcours
                        </Button>
                    ) : (
                        <Button
                            className="w-full h-12 text-base font-black gap-2"
                            variant="outline"
                            onClick={onSetActive}
                            disabled={isSwitching}
                        >
                            {isSwitching ? <Loader2 className="w-5 h-5 animate-spin" /> : <RefreshCw className="w-5 h-5" />}
                            {isSwitching ? 'Changement...' : 'Activer ce Parcours'}
                        </Button>
                    )}
                </div>
            </motion.div>
        </>
    );
}

// ==========================================
// EMPTY STATE
// ==========================================
function EmptyState({ icon, title, description, action }: {
    icon: React.ReactNode;
    title: string;
    description: string;
    action?: React.ReactNode;
}) {
    return (
        <div className="py-24 text-center space-y-4 border-2 border-dashed border-border rounded-3xl bg-muted/20">
            <div className="flex justify-center">{icon}</div>
            <h3 className="text-xl font-black">{title}</h3>
            <p className="text-muted-foreground font-medium max-w-sm mx-auto">{description}</p>
            {action && <div className="pt-2">{action}</div>}
        </div>
    );
}
