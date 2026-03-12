'use client';

import React, { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { classroomApi, Classroom, ClassroomAssignment } from '@/lib/api/classroom';
import { courseApi } from '@/lib/api/course';
import { Course, CourseNode } from '@/lib/types/course';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
    Users, BookOpen, Clock, ChevronRight,
    ArrowLeft, Plus, Lock, Unlock, User,
    Trophy, BarChart3, Settings, MoreVertical,
    Calendar, CheckCircle2, Shield, Trash2, Info, AlertCircle, PlayCircle, CalendarClock, Star,
    Eye, RefreshCw, XCircle, Pencil, Check, X, Infinity
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";

import { ClassAssignmentCard, formatDueDate, getDueStatus, getDaysLeft } from '@/components/dashboard/class-assignment-card';

export default function ClassroomDetailPage() {
    const { id } = useParams() as { id: string };
    const router = useRouter();
    const { user } = useAuthStore();

    const [loading, setLoading] = useState(true);
    const [data, setData] = useState<any>(null);
    const [courses, setCourses] = useState<Course[]>([]);
    const [error, setError] = useState<string | null>(null);

    // UI state
    const [isAssigning, setIsAssigning] = useState(false);
    const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);
    const [selectedCourseFull, setSelectedCourseFull] = useState<Course | null>(null);
    const [selectedNodeId, setSelectedNodeId] = useState<string>('');
    const [assignmentTitle, setAssignmentTitle] = useState('');
    const [assignmentDescription, setAssignmentDescription] = useState('');
    const [assignmentDueDate, setAssignmentDueDate] = useState('');
    const [isCreatingAssignment, setIsCreatingAssignment] = useState(false);

    // Stats state
    const [viewStatsAsgnId, setViewStatsAsgnId] = useState<string | null>(null);
    const [viewStatsAsgn, setViewStatsAsgn] = useState<any>(null);
    const [assignmentStats, setAssignmentStats] = useState<any[]>([]);
    const [loadingStats, setLoadingStats] = useState(false);

    // Student detail modal
    const [studentDetail, setStudentDetail] = useState<any | null>(null);
    const [studentDetailLoading, setStudentDetailLoading] = useState(false);
    const [studentDetailName, setStudentDetailName] = useState('');

    // Assignment form settings
    const [assignmentPassingGrade, setAssignmentPassingGrade] = useState(70);
    const [assignmentMaxRetakes, setAssignmentMaxRetakes] = useState(-1); // -1 = unlimited
    const [assignmentType, setAssignmentType] = useState<'quiz' | 'path_progress'>('quiz');
    const [editingAsgn, setEditingAsgn] = useState<any | null>(null); // null = create, object = edit
    const [confirmDelete, setConfirmDelete] = useState<any | null>(null); // assignment to delete

    const resetDrawer = () => {
        setSelectedCourse(null);
        setSelectedNodeId('');
        setAssignmentTitle('');
        setAssignmentDescription('');
        setAssignmentDueDate('');
        setAssignmentPassingGrade(70);
        setAssignmentMaxRetakes(-1);
        setAssignmentType('quiz');
        setEditingAsgn(null);
    };

    const openEditDrawer = (asgn: any) => {
        setEditingAsgn(asgn);
        setAssignmentTitle(asgn.title);
        setAssignmentDescription(asgn.description || '');
        setAssignmentPassingGrade(asgn.passing_grade ?? 70);
        setAssignmentMaxRetakes(asgn.max_retakes ?? -1);
        // Convert unix ts to yyyy-MM-dd
        if (asgn.due_date > 0) {
            const d = new Date(asgn.due_date * 1000);
            setAssignmentDueDate(d.toISOString().split('T')[0]);
        } else {
            setAssignmentDueDate('');
        }
        setIsAssigning(true);
    };

    const handleUpdateClass = async (data: Partial<Classroom>) => {
        try {
            await classroomApi.updateClass(id, data);
            loadData();
        } catch (e) {
            console.error('Update failed:', e);
        }
    }

    useEffect(() => {
        if (selectedCourse) {
            setSelectedCourseFull(null);
            setSelectedNodeId('');
            setAssignmentTitle('');
            setAssignmentDescription('');
            courseApi.getCourse(selectedCourse.id)
                .then(c => setSelectedCourseFull(c))
                .catch(console.error);
        }
    }, [selectedCourse]);

    // Set default due date to 1 week from now when opening the drawer
    useEffect(() => {
        if (isAssigning && !assignmentDueDate) {
            const d = new Date();
            d.setDate(d.getDate() + 7);
            setAssignmentDueDate(d.toISOString().split('T')[0]);
        }
    }, [isAssigning]);

    useEffect(() => {
        loadData();
    }, [id]);

    const loadData = async () => {
        setLoading(true);
        try {
            const [classData, allCourses] = await Promise.all([
                classroomApi.getClassDetails(id),
                courseApi.listCourses()
            ]);
            setData(classData);
            setCourses(allCourses);
        } catch (err: any) {
            console.error('Failed to load class details:', err.response?.data || err.message || err);
            setError(err.response?.data?.message || err.message || 'Échec du chargement des détails de la classe');
        } finally {
            setLoading(false);
            setIsCreatingAssignment(false);
        }
    };

    const fetchStats = async (asgnId: string, asgn: any) => {
        setViewStatsAsgnId(asgnId);
        setViewStatsAsgn(asgn);
        setStudentDetail(null);
        setLoadingStats(true);
        try {
            const stats = await classroomApi.getAssignmentStats(id, asgnId);
            setAssignmentStats(stats);
        } catch (e) {
            console.error(e);
        } finally {
            setLoadingStats(false);
        }
    };

    const fetchStudentDetail = async (studentId: string, studentName: string) => {
        if (!viewStatsAsgnId) return;
        setStudentDetailName(studentName);
        setStudentDetailLoading(true);
        try {
            const detail = await classroomApi.getStudentAssignmentDetail(id, viewStatsAsgnId, studentId);
            setStudentDetail(detail);
        } catch (e) {
            console.error(e);
        } finally {
            setStudentDetailLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center pt-20">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                    <p className="text-primary font-black text-sm tracking-widest animate-pulse">CHARGEMENT DE LA CLASSE...</p>
                </div>
            </div>
        );
    }

    if (error || !data) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center p-4 pt-24">
                <Card className="fun-card border-red-200 p-8 max-w-md w-full text-center">
                    <div className="w-16 h-16 bg-red-100 rounded-2xl flex items-center justify-center mx-auto mb-4 border-b-4 border-red-500">
                        <AlertCircle className="w-8 h-8 text-red-600" />
                    </div>
                    <h2 className="text-2xl font-black text-slate-800 mb-2">Oups !</h2>
                    <p className="text-muted-foreground font-bold mb-6">{error || 'Classe non trouvée'}</p>
                    <Button onClick={() => router.push('/classes')} className="w-full bg-slate-800 text-white font-black h-12 rounded-xl border-b-4 border-slate-950">
                        Retour aux classes
                    </Button>
                </Card>
            </div>
        );
    }

    const isTeacher = data.classroom.teacher_id === user?.id?.toString();
    const classroom = data.classroom as Classroom;

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32">
            <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
                <div className="blob-green -bottom-40 -left-20 opacity-20" />
                <div className="blob-orange top-20 right-10 opacity-10" />
            </div>

            <main className="container max-w-5xl mx-auto px-4 relative z-10 space-y-8">
                {/* Header */}
                <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} className="space-y-4">
                    <Button variant="ghost" onClick={() => router.push('/classes')} className="pl-0 text-muted-foreground hover:text-foreground">
                        <ArrowLeft className="w-4 h-4 mr-2" /> Retour aux classes
                    </Button>

                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                        <div className="flex items-center gap-6">
                            <div className={cn(
                                "w-20 h-20 rounded-3xl flex items-center justify-center text-3xl font-black text-white shadow-xl border-b-6",
                                isTeacher ? "bg-primary border-primary-hover" : "bg-secondary border-orange-600"
                            )}>
                                {classroom.name.charAt(0).toUpperCase()}
                            </div>
                            <div>
                                <h1 className="text-4xl font-black tracking-tight">{classroom.name}</h1>
                                <p className="text-muted-foreground font-semibold flex items-center gap-2">
                                    {isTeacher ? <Shield className="w-4 h-4" /> : <User className="w-4 h-4" />}
                                    {isTeacher ? 'Tu es l\'enseignant' : 'Tu es élève'} • {data.students?.length || 0} élèves
                                </p>
                            </div>
                        </div>

                        {isTeacher && (
                            <div className="flex items-center gap-3">
                                <div className="bg-white border-2 border-primary/20 rounded-2xl p-4 flex items-center gap-4 shadow-sm">
                                    <div>
                                        <p className="text-[10px] font-black uppercase text-muted-foreground tracking-widest">Code d'invitation</p>
                                        <p className="text-2xl font-black text-primary tracking-tighter">{classroom.join_code}</p>
                                    </div>
                                    <Button size="icon" variant="ghost" className="rounded-xl hover:bg-primary/5 text-primary" onClick={() => navigator.clipboard.writeText(classroom.join_code)}>
                                        <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2" /><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" /></svg>
                                    </Button>
                                </div>
                            </div>
                        )}
                    </div>
                </motion.div>

                {/* Content Tabs */}
                <Tabs defaultValue="overview" className="space-y-8">
                    <TabsList className="bg-muted/50 p-1.5 rounded-2xl border border-border flex w-fit mx-auto sm:mx-0">
                        <TabsTrigger value="overview" className="rounded-xl px-6 py-2.5 font-bold data-[state=active]:bg-white data-[state=active]:shadow-sm">
                            Vue d'ensemble
                        </TabsTrigger>
                        <TabsTrigger value="assignments" className="rounded-xl px-6 py-2.5 font-bold data-[state=active]:bg-white data-[state=active]:shadow-sm">
                            Devoirs
                        </TabsTrigger>
                        {isTeacher && (
                            <TabsTrigger value="courses" className="rounded-xl px-6 py-2.5 font-bold data-[state=active]:bg-white data-[state=active]:shadow-sm">
                                Parcours
                            </TabsTrigger>
                        )}
                        {isTeacher && (
                            <TabsTrigger value="students" className="rounded-xl px-6 py-2.5 font-bold data-[state=active]:bg-white data-[state=active]:shadow-sm">
                                Élèves
                            </TabsTrigger>
                        )}
                        <TabsTrigger value="settings" className="rounded-xl px-4 py-2.5 font-bold data-[state=active]:bg-white data-[state=active]:shadow-sm">
                            <Settings className="w-4 h-4" />
                        </TabsTrigger>
                    </TabsList>

                    {/* OVERVIEW CONTENT */}
                    <TabsContent value="overview" className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <Card className="fun-card border-accent/20 md:col-span-3">
                                <CardHeader>
                                    <CardTitle className="text-xl font-black flex items-center gap-2">
                                        <Info className="w-5 h-5 text-teal-600" />
                                        Informations
                                    </CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <p className="text-lg font-bold text-slate-600 leading-relaxed">
                                        {classroom.description || "Aucune description pour cette classe."}
                                    </p>
                                    <div className="flex flex-wrap gap-4 pt-4">
                                        <div className="flex items-center gap-2 text-sm font-bold text-muted-foreground">
                                            <Calendar className="w-4 h-4" /> Créé le {new Date(classroom.created_at * 1000).toLocaleDateString()}
                                        </div>
                                        <div className="flex items-center gap-2 text-sm font-bold text-muted-foreground">
                                            <Users className="w-4 h-4" /> {data.students?.length || 0} Élèves
                                        </div>
                                        <div className="flex items-center gap-2 text-sm font-bold text-muted-foreground">
                                            <BookOpen className="w-4 h-4" /> {data.assignments?.length || 0} Devoirs
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    {/* ASSIGNMENTS CONTENT */}
                    <TabsContent value="assignments" className="space-y-6">
                        <div className="flex items-center justify-between">
                            <h2 className="text-2xl font-black">Devoirs Assignés</h2>
                            {isTeacher && (
                                <Button onClick={() => setIsAssigning(true)} className="bg-primary hover:bg-primary-hover text-white font-black rounded-xl border-b-4 border-primary-hover shadow-lg">
                                    <Plus className="w-4 h-4 mr-2" /> Nouveau Devoir
                                </Button>
                            )}
                        </div>

                        <div className="space-y-4">
                            {data.assignments?.length > 0 ? (
                                data.assignments.map((asgn: any) => (
                                    <ClassAssignmentCard
                                        key={asgn.id}
                                        assignment={asgn}
                                        isTeacher={isTeacher}
                                        onStats={fetchStats}
                                        onEdit={openEditDrawer}
                                        onDelete={setConfirmDelete}
                                    />
                                ))
                            ) : (
                                <div className="py-20 text-center border-2 border-dashed border-border rounded-3xl bg-muted/10">
                                    <Clock className="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-30" />
                                    <h3 className="text-xl font-bold text-slate-400">Aucun devoir en cours</h3>
                                    <p className="text-sm text-muted-foreground font-medium">Tout le monde est à jour !</p>
                                </div>
                            )}
                        </div>
                    </TabsContent>

                    {/* COURSES CONTENT */}
                    {isTeacher && (
                        <TabsContent value="courses" className="space-y-6">
                            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                                <div>
                                    <h2 className="text-2xl font-black">Parcours Associés</h2>
                                    <p className="text-muted-foreground font-semibold">Tu dois lier un parcours à cette classe avant de pouvoir donner des devoirs dessus.</p>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {data.courses?.map((course: any) => (
                                    <div key={course.id} className="p-4 rounded-2xl border-2 border-border bg-white flex items-center justify-between">
                                        <div className="flex items-center gap-4">
                                            <div className="w-10 h-10 rounded-xl flex items-center justify-center font-black text-white" style={{ background: course.color }}>
                                                {course.title.charAt(0)}
                                            </div>
                                            <span className="font-bold">{course.title}</span>
                                        </div>
                                        <Button size="icon" variant="ghost" className="text-red-500 hover:bg-red-50" onClick={async () => {
                                            if (confirm("Retirer ce parcours de la classe ?")) {
                                                await classroomApi.removeCourse(id, course.id);
                                                loadData();
                                            }
                                        }}>
                                            <Trash2 className="w-4 h-4" />
                                        </Button>
                                    </div>
                                ))}

                                <Dialog>
                                    <DialogTrigger asChild>
                                        <div className="p-4 rounded-2xl border-2 border-dashed border-primary/40 bg-primary/5 flex items-center justify-center cursor-pointer hover:bg-primary/10 transition-colors">
                                            <Plus className="w-6 h-6 text-primary mr-2" />
                                            <span className="font-black text-primary">Lier un parcours</span>
                                        </div>
                                    </DialogTrigger>
                                    <DialogContent className="sm:max-w-md">
                                        <DialogHeader>
                                            <DialogTitle className="text-xl font-black">Ajouter un Parcours</DialogTitle>
                                        </DialogHeader>
                                        <div className="space-y-4 pt-4">
                                            {courses.filter(c => !data.courses?.some((dc: any) => dc.id === c.id)).map(c => (
                                                <div key={c.id} className="flex items-center justify-between p-3 border-2 border-border rounded-xl">
                                                    <span className="font-bold">{c.title}</span>
                                                    <Button size="sm" onClick={async () => {
                                                        await classroomApi.addCourse(id, c.id);
                                                        loadData();
                                                    }}>Ajouter</Button>
                                                </div>
                                            ))}
                                            {courses.filter(c => !data.courses?.some((dc: any) => dc.id === c.id)).length === 0 && (
                                                <p className="text-center font-bold text-muted-foreground">Tous les parcours disponibles ont déjà été ajoutés.</p>
                                            )}
                                        </div>
                                    </DialogContent>
                                </Dialog>

                            </div>
                        </TabsContent>
                    )}

                    {/* STUDENTS CONTENT */}
                    {isTeacher && (
                        <TabsContent value="students" className="space-y-6">
                            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                                <h2 className="text-2xl font-black">Élèves ({data.students?.length || 0})</h2>
                                <div className="flex items-center gap-3">

                                    <Button variant="outline" className="rounded-xl border-border font-bold">
                                        Exporter (.csv)
                                    </Button>
                                </div>
                            </div>

                            <div className="bg-card border-2 border-border rounded-3xl overflow-hidden shadow-md">
                                <table className="w-full text-left">
                                    <thead className="bg-muted/50 border-b-2 border-border">
                                        <tr>
                                            <th className="px-6 py-4 text-xs font-black uppercase text-muted-foreground tracking-widest">Élève</th>
                                            <th className="px-6 py-4 text-xs font-black uppercase text-muted-foreground tracking-widest">Niveau (Exp)</th>
                                            <th className="px-6 py-4 text-xs font-black uppercase text-muted-foreground tracking-widest">Dernier passage</th>
                                            <th className="px-6 py-4 text-xs font-black uppercase text-muted-foreground tracking-widest">Status</th>
                                            <th className="px-6 py-4 text-center"></th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y-2 divide-border">
                                        {data.students?.map((s: any) => (
                                            <tr key={s.id} className="hover:bg-muted/20 transition-colors">
                                                <td className="px-6 py-4">
                                                    <div className="flex items-center gap-3">
                                                        <div className="w-10 h-10 rounded-xl bg-accent/20 flex items-center justify-center font-black text-teal-700">
                                                            {s.username.charAt(0).toUpperCase()}
                                                        </div>
                                                        <span className="font-bold">{s.username}</span>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <div className="flex items-center gap-2">
                                                        <BarChart3 className="w-4 h-4 text-primary" />
                                                        <span className="font-bold text-sm">Niveau 1 (120 XP)</span>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 text-sm font-semibold text-muted-foreground">
                                                    Aujourd'hui
                                                </td>
                                                <td className="px-6 py-4">
                                                    <span className="bg-emerald-100 text-emerald-700 text-[10px] font-black px-2.5 py-1 rounded-full border border-emerald-200">
                                                        ACTIF
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 text-center">
                                                    <Button size="sm" variant="ghost" className="rounded-xl text-red-500 hover:text-red-600 hover:bg-red-50" onClick={async () => {
                                                        if (confirm(`Êtes-vous sûr de vouloir retirer ${s.username} de la classe ?`)) {
                                                            await classroomApi.removeStudent(id, s.id);
                                                            loadData();
                                                        }
                                                    }}>
                                                        <Trash2 className="w-4 h-4 mr-2" /> Retirer
                                                    </Button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </TabsContent>
                    )}

                    {/* SETTINGS CONTENT */}
                    <TabsContent value="settings" className="space-y-6">
                        <Card className="fun-card border-red-100 bg-red-50/10">
                            <CardHeader>
                                <CardTitle className="text-xl font-black">Zone de Gestion</CardTitle>
                                <CardDescription className="font-bold">Gère l'accès et le futur de cette classe.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                <div className="flex items-center justify-between p-4 bg-white rounded-2xl border-2 border-border border-b-4">
                                    <div className="space-y-0.5">
                                        <p className="font-black text-lg">Verrouiller la classe</p>
                                        <p className="text-sm text-muted-foreground font-semibold">Empêche les nouveaux élèves de rejoindre avec le code.</p>
                                    </div>
                                    <Button variant={classroom.is_locked ? "default" : "outline"} className={cn("rounded-xl font-bold h-12 px-6", classroom.is_locked ? "bg-red-500 hover:bg-red-600 text-white border-b-4 border-red-700" : "")}>
                                        {classroom.is_locked ? <Lock className="w-4 h-4 mr-2" /> : <Unlock className="w-4 h-4 mr-2" />}
                                        {classroom.is_locked ? "Déverrouiller" : "Verrouiller"}
                                    </Button>
                                </div>

                                {isTeacher && (
                                    <div className="flex items-center justify-between p-4 bg-white rounded-2xl border-2 border-destructive/20 border-b-4">
                                        <div className="space-y-0.5">
                                            <p className="font-black text-lg text-red-600">Supprimer la classe</p>
                                            <p className="text-sm text-muted-foreground font-semibold">Attention, action irréversible ! Toutes les données seront perdues.</p>
                                        </div>
                                        <Button variant="destructive" className="rounded-xl font-bold h-12 px-6 border-b-4 border-red-800">
                                            <Trash2 className="w-4 h-4 mr-2" /> Supprimer
                                        </Button>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </TabsContent>
                </Tabs>
            </main>

            {/* ASSIGNMENT DRAWER */}
            <AnimatePresence>
                {isAssigning && (
                    <>
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} onClick={() => { setIsAssigning(false); resetDrawer(); }} className="fixed inset-0 bg-black/40 backdrop-blur-sm z-[60]" />
                        <motion.div initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }} className="fixed right-0 top-0 h-full w-full max-w-md bg-card border-l-4 border-primary/20 shadow-2xl z-[70] flex flex-col p-8 space-y-6">
                            <div className="flex items-center justify-between">
                                <h2 className="text-3xl font-black tracking-tight">
                                    {editingAsgn ? 'Modifier le devoir' : 'Donner un devoir'}
                                </h2>
                                <button onClick={() => { setIsAssigning(false); resetDrawer(); }} className="w-10 h-10 rounded-full bg-muted flex items-center justify-center hover:bg-primary-light transition-colors">✕</button>
                            </div>

                            <p className="text-muted-foreground font-semibold">
                                {editingAsgn ? 'Modifie les paramètres de ce devoir.' : 'Sélectionne un parcours pour tes élèves.'}
                            </p>

                            <div className="flex-1 overflow-y-auto space-y-4 pr-2">
                                {editingAsgn ? (
                                    // Edit mode: show read-only course/node info + editable fields
                                    <div className="space-y-4">
                                        <div className="p-3 rounded-xl bg-muted/40 border border-border text-sm font-semibold text-muted-foreground">
                                            <span className="text-[10px] font-black uppercase tracking-widest block mb-1">Devoir actuel</span>
                                            {editingAsgn.course_name} → {editingAsgn.node_title || editingAsgn.node_id}
                                        </div>

                                        <div className="space-y-2">
                                            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Titre</label>
                                            <input type="text" value={assignmentTitle} onChange={e => setAssignmentTitle(e.target.value)}
                                                className="w-full h-12 rounded-xl border border-input bg-white px-3 font-bold text-sm shadow-sm" />
                                        </div>

                                        <div className="space-y-2">
                                            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Description (optionnel)</label>
                                            <textarea value={assignmentDescription} onChange={e => setAssignmentDescription(e.target.value)}
                                                className="w-full h-20 rounded-xl border border-input bg-white px-3 py-2 font-semibold text-sm shadow-sm resize-none"
                                                placeholder="Instructions supplémentaires..." />
                                        </div>

                                        <div className="space-y-2">
                                            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                <Calendar className="w-3 h-3" /> Date limite
                                            </label>
                                            <input type="date" value={assignmentDueDate} onChange={e => setAssignmentDueDate(e.target.value)}
                                                className="w-full h-12 rounded-xl border border-input bg-white px-3 font-semibold text-sm shadow-sm" />
                                        </div>

                                        <div className="space-y-2">
                                            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                <Trophy className="w-3 h-3" /> Note de passage ({assignmentPassingGrade}%)
                                            </label>
                                            <input type="range" min={0} max={100} step={5} value={assignmentPassingGrade}
                                                onChange={e => setAssignmentPassingGrade(Number(e.target.value))} className="w-full accent-primary" />
                                            <div className="flex justify-between text-[10px] font-bold text-muted-foreground">
                                                <span>0%</span><span>50%</span><span>100%</span>
                                            </div>
                                        </div>

                                        <div className="space-y-2">
                                            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                <RefreshCw className="w-3 h-3" /> Reprises autorisées
                                            </label>
                                            <div className="flex items-center gap-2">
                                                <button type="button" onClick={() => setAssignmentMaxRetakes(-1)}
                                                    className={cn("flex-1 h-10 rounded-xl border-2 font-bold text-sm transition-all flex items-center justify-center gap-1.5",
                                                        assignmentMaxRetakes === -1 ? "bg-primary text-white border-primary" : "bg-white text-muted-foreground border-border hover:border-primary/40")}>
                                                    <Infinity className="w-4 h-4" /> Illimitées
                                                </button>
                                                <button type="button" onClick={() => setAssignmentMaxRetakes(assignmentMaxRetakes === -1 ? 1 : assignmentMaxRetakes)}
                                                    className={cn("flex-1 h-10 rounded-xl border-2 font-bold text-sm transition-all",
                                                        assignmentMaxRetakes !== -1 ? "bg-primary text-white border-primary" : "bg-white text-muted-foreground border-border hover:border-primary/40")}>
                                                    Limité
                                                </button>
                                            </div>
                                            {assignmentMaxRetakes !== -1 && (
                                                <div className="flex items-center gap-3 mt-2">
                                                    <input type="number" min={0} max={99} value={assignmentMaxRetakes}
                                                        onChange={e => setAssignmentMaxRetakes(Math.max(0, parseInt(e.target.value) || 0))}
                                                        className="w-24 h-10 rounded-xl border border-input bg-white px-3 font-bold text-sm shadow-sm text-center" />
                                                    <span className="text-sm font-semibold text-muted-foreground">
                                                        reprise{assignmentMaxRetakes !== 1 ? 's' : ''} ({assignmentMaxRetakes + 1} essai{assignmentMaxRetakes + 1 !== 1 ? 's' : ''} au total)
                                                    </span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                ) : (
                                    // Create mode: course/node picker
                                    !selectedCourse ? (
                                        <>
                                            {/* Assignment type picker */}
                                            <div className="space-y-2">
                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Type de devoir</label>
                                                <div className="grid grid-cols-2 gap-2">
                                                    <button
                                                        type="button"
                                                        onClick={() => { setAssignmentType('quiz'); setSelectedNodeId(''); }}
                                                        className={cn(
                                                            'flex flex-col items-center justify-center gap-1.5 p-3 rounded-xl border-2 font-bold text-sm transition-all',
                                                            assignmentType === 'quiz'
                                                                ? 'bg-primary text-white border-primary'
                                                                : 'bg-white text-muted-foreground border-border hover:border-primary/40'
                                                        )}
                                                    >
                                                        <PlayCircle className="w-5 h-5" />
                                                        Quiz
                                                        <span className={cn('text-[9px] font-semibold leading-tight text-center', assignmentType === 'quiz' ? 'text-white/80' : 'text-muted-foreground/70')}>Faire un quiz du cours</span>
                                                    </button>
                                                    <button
                                                        type="button"
                                                        onClick={() => { setAssignmentType('path_progress'); setSelectedNodeId(''); }}
                                                        className={cn(
                                                            'flex flex-col items-center justify-center gap-1.5 p-3 rounded-xl border-2 font-bold text-sm transition-all',
                                                            assignmentType === 'path_progress'
                                                                ? 'bg-primary text-white border-primary'
                                                                : 'bg-white text-muted-foreground border-border hover:border-primary/40'
                                                        )}
                                                    >
                                                        <ChevronRight className="w-5 h-5" />
                                                        Parcours
                                                        <span className={cn('text-[9px] font-semibold leading-tight text-center', assignmentType === 'path_progress' ? 'text-white/80' : 'text-muted-foreground/70')}>Atteindre une étape du cours</span>
                                                    </button>
                                                </div>
                                            </div>

                                            {data.courses?.length > 0 ? (
                                                data.courses.map((course: any) => (
                                                    <div
                                                        key={course.id}
                                                        onClick={() => setSelectedCourse(course)}
                                                        className="p-4 rounded-2xl border-2 cursor-pointer transition-all flex items-center gap-4 border-border hover:border-primary/40 bg-muted/20"
                                                    >
                                                        <div className="w-12 h-12 rounded-xl flex items-center justify-center bg-white shadow-sm font-black text-xl border-b-2" style={{ color: course.color, borderColor: `${course.color}44` }}>
                                                            {course.title.charAt(0).toUpperCase()}
                                                        </div>
                                                        <div>
                                                            <p className="font-black text-sm">{course.title}</p>
                                                            <p className="text-xs text-muted-foreground italic">{course.description ? course.description.slice(0, 40) + '...' : 'Pas de description'}</p>
                                                        </div>
                                                    </div>
                                                ))
                                            ) : (
                                                <div className="p-6 border-2 border-dashed border-red-200 bg-red-50 rounded-2xl text-center">
                                                    <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-2" />
                                                    <p className="font-bold text-red-700">Aucun parcours lié !</p>
                                                    <p className="text-sm text-red-600/80 mt-1">Va dans l&apos;onglet &quot;Parcours&quot; pour en ajouter un avant de donner des devoirs.</p>
                                                </div>
                                            )}
                                        </>
                                    ) : (
                                        <div className="space-y-4">
                                            <Button variant="ghost" onClick={() => setSelectedCourse(null)} className="pl-0 text-muted-foreground -ml-2">
                                                <ArrowLeft className="w-4 h-4 mr-2" /> Retour aux parcours
                                            </Button>

                                            <div className="p-4 rounded-2xl border-2 border-primary bg-primary-light/50 flex items-center gap-4">
                                                <div className="w-12 h-12 rounded-xl flex items-center justify-center bg-white shadow-sm font-black text-xl border-b-2" style={{ color: selectedCourse?.color, borderColor: `${selectedCourse?.color}44` }}>
                                                    {selectedCourse?.title.charAt(0).toUpperCase()}
                                                </div>
                                                <div>
                                                    <p className="font-black text-sm">{selectedCourse?.title}</p>
                                                    <p className="text-xs text-muted-foreground italic">Sélectionné</p>
                                                </div>
                                                <CheckCircle2 className="w-5 h-5 ml-auto text-primary" />
                                            </div>

                                            {selectedCourseFull ? (
                                                <div className="space-y-4 pt-4 border-t border-border">
                                                    <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                                                                    {assignmentType === 'path_progress'
                                                                        ? 'Étape cible dans ce parcours'
                                                                        : 'Sélectionner un Quiz dans ce parcours'}
                                                                </label>
                                                                <select
                                                                    value={selectedNodeId}
                                                                    onChange={e => {
                                                                        setSelectedNodeId(e.target.value);
                                                                        const node = selectedCourseFull.nodes?.find(n => n.id === e.target.value);
                                                                        if (node) setAssignmentTitle(
                                                                            assignmentType === 'path_progress'
                                                                                ? `Parcours: atteindre "${node.title}"`
                                                                                : `Devoir: ${node.title}`
                                                                        );
                                                                    }}
                                                                    className="w-full h-12 rounded-xl border border-input bg-white px-3 font-semibold text-sm shadow-sm"
                                                                >
                                                                    <option value="">
                                                                        {assignmentType === 'path_progress' ? '-- Choisir une étape cible --' : '-- Choisir un quiz --'}
                                                                    </option>
                                                                    {(assignmentType === 'path_progress'
                                                                        ? selectedCourseFull.nodes?.filter(n => n.node_type !== 'start')
                                                                        : selectedCourseFull.nodes?.filter(n => n.node_type === 'quiz')
                                                                    )?.map(n => (
                                                                        <option key={n.id} value={n.id}>
                                                                            {assignmentType === 'path_progress' ? `[${n.node_type}] ` : ''}{n.title}
                                                                        </option>
                                                                    ))}
                                                                </select>
                                                                {assignmentType === 'path_progress' && (
                                                                    <p className="text-[10px] text-muted-foreground font-semibold">
                                                                        L&apos;élève devra progresser dans le parcours jusqu&apos;à atteindre cette étape.
                                                                    </p>
                                                                )}
                                                            </div>

                                                            {selectedNodeId && (
                                                                <>
                                                            <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Titre du Devoir</label>
                                                                <input
                                                                    type="text"
                                                                    value={assignmentTitle}
                                                                    onChange={e => setAssignmentTitle(e.target.value)}
                                                                    className="w-full h-12 rounded-xl border border-input bg-white px-3 font-bold text-sm shadow-sm"
                                                                    placeholder="Ex: Refaire le chapitre 1"
                                                                />
                                                            </div>

                                                            <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Description (optionnel)</label>
                                                                <textarea
                                                                    value={assignmentDescription}
                                                                    onChange={e => setAssignmentDescription(e.target.value)}
                                                                    className="w-full h-20 rounded-xl border border-input bg-white px-3 py-2 font-semibold text-sm shadow-sm resize-none"
                                                                    placeholder="Instructions supplémentaires pour les élèves..."
                                                                />
                                                            </div>

                                                            <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                                    <Calendar className="w-3 h-3" /> Date limite
                                                                </label>
                                                                <input
                                                                    type="date"
                                                                    value={assignmentDueDate}
                                                                    onChange={e => setAssignmentDueDate(e.target.value)}
                                                                    className="w-full h-12 rounded-xl border border-input bg-white px-3 font-semibold text-sm shadow-sm"
                                                                />
                                                            </div>

                                                            <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                                    <Trophy className="w-3 h-3" /> Note de passage ({assignmentPassingGrade}%)
                                                                </label>
                                                                <input type="range" min={0} max={100} step={5} value={assignmentPassingGrade}
                                                                    onChange={e => setAssignmentPassingGrade(Number(e.target.value))} className="w-full accent-primary" />
                                                                <div className="flex justify-between text-[10px] font-bold text-muted-foreground">
                                                                    <span>0%</span><span>50%</span><span>100%</span>
                                                                </div>
                                                            </div>

                                                            <div className="space-y-2">
                                                                <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
                                                                    <RefreshCw className="w-3 h-3" /> Reprises autorisées
                                                                </label>
                                                                <div className="flex items-center gap-2">
                                                                    <button type="button" onClick={() => setAssignmentMaxRetakes(-1)}
                                                                        className={cn("flex-1 h-10 rounded-xl border-2 font-bold text-sm transition-all flex items-center justify-center gap-1.5",
                                                                            assignmentMaxRetakes === -1 ? "bg-primary text-white border-primary" : "bg-white text-muted-foreground border-border hover:border-primary/40")}>
                                                                        <Infinity className="w-4 h-4" /> Illimitées
                                                                    </button>
                                                                    <button type="button" onClick={() => setAssignmentMaxRetakes(assignmentMaxRetakes === -1 ? 1 : assignmentMaxRetakes)}
                                                                        className={cn("flex-1 h-10 rounded-xl border-2 font-bold text-sm transition-all",
                                                                            assignmentMaxRetakes !== -1 ? "bg-primary text-white border-primary" : "bg-white text-muted-foreground border-border hover:border-primary/40")}>
                                                                        Limité
                                                                    </button>
                                                                </div>
                                                                {assignmentMaxRetakes !== -1 && (
                                                                    <div className="flex items-center gap-3 mt-2">
                                                                        <input type="number" min={0} max={99} value={assignmentMaxRetakes}
                                                                            onChange={e => setAssignmentMaxRetakes(Math.max(0, parseInt(e.target.value) || 0))}
                                                                            className="w-24 h-10 rounded-xl border border-input bg-white px-3 font-bold text-sm shadow-sm text-center" />
                                                                        <span className="text-sm font-semibold text-muted-foreground">
                                                                            reprise{assignmentMaxRetakes !== 1 ? 's' : ''} ({assignmentMaxRetakes + 1} essai{assignmentMaxRetakes + 1 !== 1 ? 's' : ''} au total)
                                                                        </span>
                                                                    </div>
                                                                )}
                                                            </div>
                                                                </>
                                                            )}
                                                        </div>
                                                    ) : (
                                                <div className="py-8 flex flex-col items-center justify-center gap-2">
                                                    <div className="w-8 h-8 rounded-full border-4 border-primary border-t-transparent animate-spin" />
                                                    <p className="text-xs font-bold text-muted-foreground animate-pulse">Chargement des étapes...</p>
                                                </div>
                                            )}
                                        </div>
                                    )
                                )}
                            </div>

                            <div className="pt-6 border-t border-border">
                                <Button
                                    className="w-full h-14 text-xl font-black bg-primary text-white rounded-2xl border-b-6 border-primary-hover shadow-lg"
                                    disabled={(!editingAsgn && (!selectedCourse || !selectedNodeId || !assignmentTitle)) || isCreatingAssignment}
                                    onClick={async () => {
                                        setIsCreatingAssignment(true);
                                        try {
                                            let dueDateTs: number | undefined;
                                            if (assignmentDueDate) {
                                                const d = new Date(assignmentDueDate + 'T23:59:59');
                                                dueDateTs = Math.floor(d.getTime() / 1000);
                                            }
                                            if (editingAsgn) {
                                                // EDIT mode
                                                await classroomApi.updateAssignment(id, editingAsgn.id, {
                                                    title: assignmentTitle,
                                                    description: assignmentDescription || undefined,
                                                    due_date: dueDateTs ?? 0,
                                                    passing_grade: assignmentPassingGrade,
                                                    max_retakes: assignmentMaxRetakes,
                                                });
                                            } else {
                                                // CREATE mode
                                                if (!selectedCourse || !selectedNodeId) return;
                                                await classroomApi.createAssignment(id, {
                                                    course_id: selectedCourse.id,
                                                    node_id: selectedNodeId,
                                                    title: assignmentTitle,
                                                    description: assignmentDescription || undefined,
                                                    due_date: dueDateTs,
                                                    passing_grade: assignmentPassingGrade,
                                                    max_retakes: assignmentMaxRetakes,
                                                    assignment_type: assignmentType,
                                                });
                                            }
                                            setIsAssigning(false);
                                            resetDrawer();
                                            loadData();
                                        } catch (e) {
                                            console.error('Failed saving assignment', e);
                                            alert("Une erreur s'est produite.");
                                        } finally {
                                            setIsCreatingAssignment(false);
                                        }
                                    }}
                                >
                                    {isCreatingAssignment
                                        ? (editingAsgn ? 'ENREGISTREMENT...' : 'ASSIGNATION EN COURS...')
                                        : (editingAsgn ? <span className="flex items-center gap-2">ENREGISTRER <Check className="w-5 h-5" /></span> : 'ASSIGNER LE DEVOIR !')}
                                </Button>
                            </div>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>

            {/* Stats Modal */}
            <Dialog open={viewStatsAsgnId !== null} onOpenChange={(open) => !open && setViewStatsAsgnId(null)}>
                <DialogContent className="max-w-lg w-full bg-slate-50 border-r-4 border-b-4 border-slate-800 rounded-3xl p-6">
                    <DialogHeader>
                        <DialogTitle className="text-2xl font-black text-slate-800 tracking-tight flex items-center gap-3">
                            <BarChart3 className="w-8 h-8 text-primary" />
                            Statistiques du Devoir
                        </DialogTitle>
                        <DialogDescription className="text-slate-500 font-medium">
                            {viewStatsAsgn?.title && (
                                <span className="block text-sm font-bold text-slate-700 mt-1">{viewStatsAsgn.title}</span>
                            )}
                            {viewStatsAsgn?.due_date > 0 && (
                                <span className="flex items-center gap-1.5 text-xs mt-2 text-slate-400">
                                    <Calendar className="w-3 h-3" /> Date limite: {formatDueDate(viewStatsAsgn.due_date)}
                                </span>
                            )}
                        </DialogDescription>
                    </DialogHeader>

                    {/* Summary bar */}
                    {!loadingStats && assignmentStats.length > 0 && (
                        <div className="flex items-center gap-4 p-4 bg-white rounded-2xl border-2 border-slate-100 mt-2">
                            <div className="flex-1">
                                <p className="text-[10px] font-black uppercase tracking-widest text-slate-400">Complétés</p>
                                <p className="text-2xl font-black text-primary">
                                    {assignmentStats.filter(s => s.is_completed).length}/{assignmentStats.length}
                                </p>
                            </div>
                            <div className="w-px h-10 bg-slate-200" />
                            <div className="flex-1">
                                <p className="text-[10px] font-black uppercase tracking-widest text-slate-400">Réussis</p>
                                <p className="text-2xl font-black text-green-700">
                                    {assignmentStats.filter(s => s.is_completed && (s.percentage ?? 0) >= (viewStatsAsgn?.passing_grade ?? 70)).length}
                                    <span className="text-sm text-slate-400 ml-1">/ {viewStatsAsgn?.passing_grade ?? 70}%</span>
                                </p>
                            </div>
                            <div className="w-px h-10 bg-slate-200" />
                            <div className="flex-1">
                                <p className="text-[10px] font-black uppercase tracking-widest text-slate-400">Moyenne</p>
                                <p className="text-2xl font-black text-teal-700">
                                    {assignmentStats.filter(s => s.is_completed).length > 0
                                        ? Math.round(assignmentStats.filter(s => s.is_completed).reduce((acc, s) => acc + (s.percentage || 0), 0) / assignmentStats.filter(s => s.is_completed).length)
                                        : 0}%
                                </p>
                            </div>
                        </div>
                    )}

                    {loadingStats ? (
                        <div className="py-12 flex flex-col items-center justify-center">
                            <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
                            <p className="mt-4 font-bold text-slate-500">Chargement...</p>
                        </div>
                    ) : studentDetail !== null ? (
                        // Student detail view
                        <div className="mt-4 space-y-4">
                            <button onClick={() => setStudentDetail(null)} className="flex items-center gap-2 text-sm font-bold text-primary hover:underline">
                                <ArrowLeft className="w-4 h-4" /> Retour à la liste
                            </button>
                            <div className="p-4 bg-white rounded-2xl border-2 border-slate-100 space-y-2">
                                <div className="flex items-center justify-between">
                                    <span className="font-black text-slate-800">{studentDetailName}</span>
                                    <span className={cn(
                                        "text-xs font-black px-3 py-1.5 rounded-full border-2",
                                        studentDetail.passed_assignment
                                            ? "bg-green-50 text-green-700 border-green-200"
                                            : "bg-red-50 text-red-700 border-red-200"
                                    )}>
                                        {studentDetail.passed_assignment
                                            ? <span className="flex items-center gap-1"><Check className="w-3 h-3" /> Réussi</span>
                                            : <span className="flex items-center gap-1"><X className="w-3 h-3" /> Échoué</span>}
                                    </span>
                                </div>
                                {studentDetail.attempt && (
                                    <div className="flex items-center gap-4 text-sm">
                                        <span className="font-bold text-slate-600">Score: <span className="text-primary">{Math.round(studentDetail.attempt.percentage ?? 0)}%</span></span>
                                        <span className="text-muted-foreground">({studentDetail.attempt.score}/{studentDetail.attempt.max_score})</span>
                                        <span className="text-xs text-slate-400">{studentDetail.attempt_count} tentative{studentDetail.attempt_count > 1 ? 's' : ''}</span>
                                    </div>
                                )}
                                {!studentDetail.attempt && studentDetail.assignment_type !== 'path_progress' && (
                                    <p className="text-sm text-slate-400 font-semibold">Pas encore soumis.</p>
                                )}
                            </div>
                            {studentDetailLoading ? (
                                <div className="py-4 flex justify-center"><div className="w-6 h-6 border-4 border-primary border-t-transparent rounded-full animate-spin" /></div>
                            ) : studentDetail.assignment_type === 'path_progress' ? (
                                <div className={cn(
                                    'mt-4 p-5 rounded-2xl border-2 text-center font-bold',
                                    studentDetail.passed_assignment
                                        ? 'bg-green-50 border-green-200 text-green-800'
                                        : 'bg-slate-50 border-slate-200 text-slate-500'
                                )}>
                                    {studentDetail.passed_assignment
                                        ? '✓ L\'élève a atteint l\'étape cible dans ce parcours.'
                                        : 'L\'élève n\'a pas encore atteint l\'étape cible.'}
                                </div>
                            ) : (
                                <div className="space-y-3 max-h-[50vh] overflow-y-auto custom-scrollbar pr-2">
                                    {(studentDetail.questions ?? []).map((q: any, i: number) => {
                                        const ans = (studentDetail.answers ?? []).find((a: any) => a.question_id === q.id);
                                        return (
                                            <div key={q.id ?? i} className={cn(
                                                "p-4 rounded-2xl border-2",
                                                ans?.is_correct ? "bg-green-50 border-green-200" : "bg-red-50 border-red-200"
                                            )}>
                                                <p className="text-xs font-black uppercase tracking-widest mb-1 flex items-center gap-1"
                                                    style={{ color: ans?.is_correct ? '#15803d' : '#dc2626' }}>
                                                    {ans?.is_correct
                                                        ? <><Check className="w-3 h-3" /> Correct</>
                                                        : <><X className="w-3 h-3" /> Incorrect</>}
                                                </p>
                                                <p className="font-bold text-slate-700 text-sm">{q.question_text}</p>
                                                <div className="mt-2 space-y-1">
                                                    <p className="text-xs font-semibold text-slate-500">Réponse: <span className="font-black text-slate-700">{ans?.user_answer ?? '—'}</span></p>
                                                    {!ans?.is_correct && (
                                                        <p className="text-xs font-semibold text-green-700">Attendu: <span className="font-black">{q.correct_answer}</span></p>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="mt-4 space-y-3 max-h-[60vh] overflow-y-auto custom-scrollbar pr-2">
                            {assignmentStats.length > 0 ? (
                                assignmentStats.map((stat, idx) => {
                                    const passed = stat.percentage != null && stat.percentage >= (viewStatsAsgn?.passing_grade ?? 70);
                                    return (
                                        <button
                                            key={idx}
                                            onClick={() => stat.is_completed && fetchStudentDetail(stat.student_id, stat.username)}
                                            className={cn(
                                                "w-full flex items-center justify-between p-3 bg-white border-2 rounded-xl shadow-sm transition-all text-left",
                                                stat.is_completed
                                                    ? passed
                                                        ? "border-green-100 hover:border-green-300 hover:shadow-md cursor-pointer"
                                                        : "border-red-100 hover:border-red-300 hover:shadow-md cursor-pointer"
                                                    : "border-slate-100 cursor-default",
                                            )}
                                        >
                                            <div className="flex items-center gap-3">
                                                <div className={cn(
                                                    "w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm",
                                                    !stat.is_completed && "bg-slate-100 text-slate-500",
                                                    stat.is_completed && passed && "bg-green-100 text-green-700",
                                                    stat.is_completed && !passed && "bg-red-100 text-red-700",
                                                )}>
                                                    {!stat.is_completed && <User className="w-4 h-4" />}
                                                    {stat.is_completed && passed && <CheckCircle2 className="w-4 h-4" />}
                                                    {stat.is_completed && !passed && <XCircle className="w-4 h-4" />}
                                                </div>
                                                <div>
                                                    <span className="font-bold text-slate-700">{stat.username}</span>
                                                    {stat.is_completed && stat.completed_at > 0 && (
                                                        <p className="text-[10px] text-slate-400 font-medium">{formatDueDate(stat.completed_at)}</p>
                                                    )}
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                {stat.is_completed ? (
                                                    <>
                                                        <div className="text-right mr-1">
                                                            <p className={cn("text-sm font-black", passed ? "text-green-700" : "text-red-600")}>{Math.round(stat.percentage || 0)}%</p>
                                                            <p className="text-[10px] text-slate-400 font-bold">{stat.score}/{stat.max_score}</p>
                                                        </div>
                                                        <div className={cn(
                                                            "flex items-center gap-1.5 px-3 py-1.5 rounded-lg border",
                                                            passed ? "text-green-600 bg-green-50 border-green-200" : "text-red-600 bg-red-50 border-red-200"
                                                        )}>
                                                            {passed ? <CheckCircle2 className="w-4 h-4" /> : <XCircle className="w-4 h-4" />}
                                                            <span className="text-xs font-bold uppercase">{passed ? 'Réussi' : 'Échoué'}</span>
                                                        </div>
                                                        <Eye className="w-4 h-4 text-slate-400" />
                                                    </>
                                                ) : (
                                                    <div className="flex items-center gap-1.5 text-slate-400 bg-slate-50 px-3 py-1.5 rounded-lg border border-slate-200">
                                                        <AlertCircle className="w-4 h-4" />
                                                        <span className="text-xs font-bold uppercase">À faire</span>
                                                    </div>
                                                )}
                                            </div>
                                        </button>
                                    );
                                })
                            ) : (
                                <p className="text-center font-bold text-slate-400 py-8">Aucun élève trouvé.</p>
                            )}
                        </div>
                    )}
                </DialogContent>
            </Dialog>
            {/* Delete confirmation dialog */}
            <Dialog open={confirmDelete !== null} onOpenChange={(open) => !open && setConfirmDelete(null)}>
                <DialogContent className="max-w-sm w-full bg-white border-r-4 border-b-4 border-slate-800 rounded-3xl p-6">
                    <DialogHeader>
                        <DialogTitle className="text-xl font-black text-slate-800 flex items-center gap-2">
                            <Trash2 className="w-5 h-5 text-red-500" />
                            Supprimer le devoir
                        </DialogTitle>
                        <DialogDescription className="text-sm text-slate-500 font-medium mt-1">
                            Tu vas supprimer{' '}
                            <span className="font-black text-slate-700">&ldquo;{confirmDelete?.title}&rdquo;</span>.
                            Cette action est irréversible.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex gap-3 mt-4">
                        <Button
                            variant="outline"
                            className="flex-1 h-11 rounded-xl font-bold border-2"
                            onClick={() => setConfirmDelete(null)}
                        >
                            Annuler
                        </Button>
                        <Button
                            className="flex-1 h-11 rounded-xl font-black bg-red-500 hover:bg-red-600 text-white border-b-4 border-red-700"
                            onClick={async () => {
                                if (!confirmDelete) return;
                                await classroomApi.deleteAssignment(id, confirmDelete.id);
                                setConfirmDelete(null);
                                loadData();
                            }}
                        >
                            <Trash2 className="w-4 h-4 mr-2" /> Supprimer
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div >
    );
}
