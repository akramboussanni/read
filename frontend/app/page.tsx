'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/lib/store/auth-store';
import { courseApi } from '@/lib/api/course';
import { Course, CourseNode, CourseStatus } from '@/lib/types/course';
import { Button } from '@/components/ui/button';
import { BookOpen, ChevronRight, GraduationCap, Compass, Trophy, Flame, Plus, CheckCircle2, Users, LayoutDashboard } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { motion } from 'framer-motion';
import { CourseFlowGraph } from '@/components/course/course-flow-graph';
import { ClassDashboard } from '@/components/dashboard/class-dashboard';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

export default function Home() {
  const router = useRouter();
  const { isAuthenticated, user } = useAuthStore();

  const [activeCourse, setActiveCourse] = useState<Course | null>(null);
  const [courseStatus, setCourseStatus] = useState<CourseStatus | null>(null);
  const [classData, setClassData] = useState<any>(null);
  const [viewMode, setViewMode] = useState<'dashboard' | 'course'>('dashboard');
  const [suggestedCourses, setSuggestedCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogConfig, setDialogConfig] = useState<{
    open: boolean,
    title: string,
    description: string,
    actionLabel?: string,
    onAction?: () => void,
    loadingAction?: boolean
  }>({
    open: false,
    title: "",
    description: "",
  });

  useEffect(() => {
    if (isAuthenticated) loadUserData();
    else setLoading(false);
  }, [isAuthenticated, user?.active_course_id]);

  const loadUserData = async () => {
    try {
      const { classroomApi } = await import('@/lib/api/classroom');

      const [allCourses, classes] = await Promise.all([
        courseApi.listCourses(),
        classroomApi.listMyClasses().catch(() => ({ teaching: [], enrolled: [] }))
      ]);

      setSuggestedCourses(allCourses.slice(0, 3));

      if (classes.enrolled?.length > 0) {
        const details = await classroomApi.getClassDetails(classes.enrolled[0].id);
        setClassData(details);
      }

      if (user?.active_course_id) {
        const [course, status] = await Promise.all([
          courseApi.getCourse(user.active_course_id),
          courseApi.getCourseStatus(user.active_course_id),
        ]);
        setActiveCourse(course);
        setCourseStatus(status);
      } else {
        const enrollments = await courseApi.getMyEnrollments().catch(() => []);
        if (enrollments?.length > 0 && enrollments[0].course) {
          const c = enrollments[0].course!;
          await courseApi.setActiveCourse(c.id);
          const [course, status] = await Promise.all([
            courseApi.getCourse(c.id),
            courseApi.getCourseStatus(c.id),
          ]);
          setActiveCourse(course);
          setCourseStatus(status);
        }
      }
    } catch (err) {
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleNodePlay = (node: CourseNode) => {
    const statusObj = courseStatus?.node_statuses?.[node.id];
    const status = statusObj ? statusObj.state : 'locked';

    // Reference nodes are always accessible (bypass lock)
    if (node.node_type === 'reference') {
      router.push(`/reference/${node.id}?courseId=${activeCourse?.id}`);
      return;
    }

    if (status === 'locked' && user?.role !== 'teacher') {
      setDialogConfig({
        open: true,
        title: "Étape verrouillée 🔒",
        description: "Terminez les étapes précédentes pour débloquer celle-ci !"
      });
      return;
    }

    if (node.node_type === 'quiz') {
      try {
        const config = JSON.parse(node.quiz_config || '{}');
        if (config.quiz_id) {
          const asgn = classData?.assignments?.find((a: any) => a.node_id === node.id && (a.course_id === activeCourse?.id || a.course_id === activeCourse?.id.toString()));
          const asgnParam = asgn ? `&asgnId=${asgn.id}` : '';
          router.push(`/quizzes/${config.quiz_id}?courseId=${activeCourse?.id}&nodeId=${node.id}${asgnParam}`);
          return;
        }
      } catch (e) {
        console.error('Failed to parse quiz config', e);
      }
      setDialogConfig({
        open: true,
        title: "Quiz en construction 🛠️",
        description: "L'éditeur n'a pas encore ajouté de questions à ce quiz. Revenez plus tard !"
      });
      return;
    }

    // milestone/checkpoint: auto-completed by useEffect above, clicking just refreshes
    if (node.node_type === 'milestone' || node.node_type === 'checkpoint') {
      courseApi.completeNode(activeCourse!.id, node.id)
        .then(() => courseApi.getCourseStatus(activeCourse!.id))
        .then(setCourseStatus)
        .catch(console.error);
      return;
    }
  };

  // Auto-complete nonce (milestone/checkpoint) nodes silently when they unlock
  useEffect(() => {
    if (!courseStatus || !activeCourse) return;
    const nodes = activeCourse.nodes ?? [];
    const statuses = courseStatus.node_statuses ?? {};
    const toComplete = nodes.filter(
      (n: CourseNode) => ['milestone', 'checkpoint'].includes(n.node_type)
        && statuses[n.id]?.state === 'unlocked'
    );
    if (toComplete.length === 0) return;
    Promise.all(toComplete.map((n: CourseNode) => courseApi.completeNode(activeCourse.id, n.id)))
      .then(() => courseApi.getCourseStatus(activeCourse.id))
      .then(setCourseStatus)
      .catch(console.error);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [courseStatus?.node_statuses, activeCourse?.id]);


  // ─── Landing page ───────────────────────────────────────────────
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-background text-foreground overflow-hidden relative">
        <div className="blob-green top-10 left-10" />
        <div className="blob-orange bottom-20 right-10" />
        <div className="blob-teal top-1/2 left-1/2 -translate-x-1/2" />
        <main className="container mx-auto px-4 pt-20 pb-32 relative z-10 flex flex-col items-center text-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.5 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ type: 'spring', bounce: 0.5 }}
            className="w-40 h-40 bg-primary text-white rounded-[3rem] flex items-center justify-center mb-10 shadow-2xl border-b-8 border-primary-hover animate-float relative"
          >
            <div className="absolute inset-0 bg-white/10 rounded-[3rem] -translate-y-2" />
            <BookOpen className="w-20 h-20 relative z-10" strokeWidth={3} />
          </motion.div>
          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1, duration: 0.5 }}
            className="text-5xl md:text-7xl font-black tracking-tight text-foreground mb-6"
          >
            Apprends l'Islam <br />en{' '}
            <span className="text-secondary inline-block animate-wobble" style={{ animationDelay: '1s' }}>jouant !</span>
          </motion.h1>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="text-xl md:text-2xl text-muted-foreground max-w-2xl font-semibold mb-12"
          >
            Gagne des points, débloque des niveaux et suis un parcours pensé pour toi.
          </motion.p>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="flex flex-col sm:flex-row gap-4"
          >
            <Link href="/register" className="w-full sm:w-auto">
              <Button size="lg" className="h-20 px-12 text-2xl font-black bg-primary hover:bg-primary-hover text-white rounded-[2rem] border-b-[10px] border-primary-hover hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all w-full">
                C'EST PARTI !
              </Button>
            </Link>
            <Link href="/login" className="w-full sm:w-auto">
              <Button size="lg" variant="outline" className="h-20 px-12 text-xl font-black bg-white hover:bg-muted text-primary rounded-[2rem] border-4 border-border hover:border-primary/50 hover:-translate-y-1 active:translate-y-1 transition-all w-full">
                J'AI DÉJÀ UN COMPTE
              </Button>
            </Link>
          </motion.div>
        </main>
      </div>
    );
  }

  // ─── Loading ────────────────────────────────────────────────────
  if (loading) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center">
        <BookOpen className="w-16 h-16 text-primary animate-wobble mb-6" strokeWidth={3} />
        <p className="text-foreground font-black text-xl animate-pulse">Chargement du parcours...</p>
      </div>
    );
  }

  const nodes = activeCourse?.nodes ?? [];
  const completedCount = nodes.filter(n => {
    const s = courseStatus?.node_statuses?.[n.id]?.state;
    return s === 'completed' || s === 'mastered';
  }).length;
  const progressPct = nodes.length > 0 ? Math.round((completedCount / nodes.length) * 100) : 0;
  const currentStep = nodes.findIndex(n => courseStatus?.node_statuses?.[n.id]?.state === 'unlocked');

  // ─── Dashboard ──────────────────────────────────────────────────
  return (
    <div className="h-screen flex flex-col bg-background text-foreground overflow-hidden pt-16">

      {classData && viewMode === 'dashboard' ? (
        <ClassDashboard
          classData={classData}
          activeCourse={activeCourse}
          progressPct={progressPct}
          onViewCourse={() => setViewMode('course')}
        />
      ) : activeCourse ? (
        <>
          {/* Course Header */}
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            className="shrink-0 mx-4 mt-4 mb-2 rounded-xl border border-border bg-card shadow-sm overflow-hidden"
          >
            <div className="h-1 w-full" style={{ background: activeCourse.color || '#6C5CE7' }} />
            <div className="px-3 py-2 flex items-center gap-2 md:gap-4">
              {/* Icon + name */}
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <div
                  className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 border-b-2"
                  style={{ background: activeCourse.color || '#6C5CE7', borderColor: `${activeCourse.color || '#6C5CE7'}88` }}
                >
                  <GraduationCap className="w-5 h-5 text-white" strokeWidth={2.5} />
                </div>
                <div className="min-w-0">
                  <p className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Parcours Actif</p>
                  <h1 className="text-sm font-black leading-tight truncate">{activeCourse.title}</h1>
                </div>
              </div>

              {/* Progress */}
              <div className="hidden sm:flex items-center gap-3 shrink-0">
                <div className="flex items-center gap-1.5">
                  <Trophy className="w-3.5 h-3.5 text-amber-500 shrink-0" />
                  <span className="text-xs font-bold text-muted-foreground whitespace-nowrap">{completedCount}/{nodes.length}</span>
                </div>
                {currentStep >= 0 && (
                  <div className="flex items-center gap-1.5">
                    <Flame className="w-3.5 h-3.5 text-primary shrink-0" />
                    <span className="text-xs font-bold text-muted-foreground whitespace-nowrap">Étape {currentStep + 1}</span>
                  </div>
                )}
                <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${progressPct}%` }}
                    transition={{ duration: 0.8, ease: 'easeOut' }}
                    className="h-full rounded-full"
                    style={{ background: activeCourse.color || '#6C5CE7' }}
                  />
                </div>
                <span className="text-xs font-black tabular-nums" style={{ color: activeCourse.color || '#6C5CE7' }}>
                  {progressPct}%
                </span>
              </div>

              <div className="flex items-center gap-2">
                {classData && (
                  <Button
                    variant="ghost"
                    onClick={() => setViewMode('dashboard')}
                    className="h-9 px-4 font-black text-xs text-primary bg-primary/10 hover:bg-primary/20 hover:text-primary rounded-xl transition-all border-2 border-primary/20 flex items-center gap-2"
                  >
                    <LayoutDashboard className="w-4 h-4" /> TABLEAU DE BORD
                  </Button>
                )}
                <button
                  onClick={() => router.push('/courses')}
                  className="shrink-0 flex items-center gap-1 text-xs font-bold text-muted-foreground hover:text-primary transition-colors px-2 py-1 rounded-lg hover:bg-primary/5"
                >
                  Changer <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </motion.div>

          {/* The graph — fills remaining space */}
          <div className="flex-1 flex flex-col lg:flex-row gap-4 overflow-hidden mx-4 mb-4">
            <div className="flex-1 rounded-2xl border border-border bg-card shadow-sm overflow-hidden min-h-0">
              <CourseFlowGraph
                course={activeCourse}
                mode="view"
                courseStatus={courseStatus}
                onNodePlay={handleNodePlay}
              />
            </div>

            {/* Sidebar for Suggestions et Devoirs */}
            {(classData || suggestedCourses.length > 0) && (
              <div className="lg:w-80 shrink-0 space-y-4 h-[auto] lg:h-full lg:overflow-y-auto w-full lg:block">
                {classData && (
                  <Button
                    onClick={() => setViewMode('dashboard')}
                    className="w-full h-14 text-base font-black bg-white border-2 border-primary text-primary hover:bg-primary hover:text-white rounded-2xl shadow-sm transition-all flex items-center justify-center gap-2 border-b-4 active:translate-y-0.5 active:border-b-2"
                  >
                    <Users className="w-5 h-5" /> Retour à Ma Classe
                  </Button>
                )}
                {classData?.assignments?.length > 0 && (
                  <Card className="fun-card border-secondary/20 bg-secondary/5">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-xs font-black uppercase tracking-widest text-secondary flex items-center gap-2">
                        <Plus className="w-3 h-3" /> Devoirs Enseignant
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {classData.assignments.slice(0, 2).map((a: any) => (
                        <div key={a.id} onClick={async () => {
                          if (a.is_completed) return; // Don't redirect if already done
                          try {
                            const fullC = await courseApi.getCourse(a.course_id);
                            const n = fullC.nodes?.find((x: CourseNode) => x.id === a.node_id);
                            if (n && n.node_type === 'quiz') {
                              const conf = JSON.parse(n.quiz_config || '{}');
                              if (conf.quiz_id) {
                                router.push(`/quizzes/${conf.quiz_id}?courseId=${a.course_id}&nodeId=${a.node_id}&asgnId=${a.id}`);
                                return;
                              }
                            }
                            alert("Impossible d'ouvrir ce quizz.");
                          } catch (e) {
                            console.error(e);
                          }
                        }}
                          className={`p-3 bg-white rounded-xl border-2 shadow-sm transition-all ${a.is_completed ? 'border-green-200 opacity-75' : 'border-secondary/10 hover:border-secondary/40 cursor-pointer'}`}
                        >
                          <div className="flex items-center justify-between">
                            <p className="text-sm font-black line-clamp-1">{a.title}</p>
                            {a.is_completed && <CheckCircle2 className="w-4 h-4 text-green-600 shrink-0" />}
                          </div>
                          <p className="text-[10px] text-muted-foreground font-bold">{a.course_name || a.course_id}</p>
                          {a.due_date > 0 && (
                            <p className="text-[10px] text-orange-600 font-bold mt-0.5">
                              Limite: {new Date(a.due_date * 1000).toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })}
                            </p>
                          )}
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                )}

                <Card className="fun-card border-accent/20">
                  <CardHeader className="pb-2">
                    <CardTitle className="text-xs font-black uppercase tracking-widest text-teal-600 flex items-center gap-2">
                      <Compass className="w-3 h-3" /> Pour Toi
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {suggestedCourses.filter(c => c.id !== activeCourse?.id).map(c => (
                      <div key={c.id} onClick={() => router.push('/courses')} className="group flex items-center gap-3 p-2.5 rounded-xl hover:bg-accent-light transition-all cursor-pointer">
                        <div className="w-8 h-8 rounded-lg flex items-center justify-center text-white text-xs font-black border-b-2" style={{ background: c.color, borderColor: `${c.color}77` }}>
                          {c.title.charAt(0)}
                        </div>
                        <span className="text-sm font-bold truncate group-hover:text-teal-700">{c.title}</span>
                      </div>
                    ))}
                  </CardContent>
                </Card>

                <div className="p-4 bg-primary-light/30 rounded-2xl border-2 border-dashed border-primary/20 text-center relative overflow-hidden group">
                  <div className="absolute inset-0 bg-white/40 translate-y-full group-hover:translate-y-0 transition-transform duration-700" />
                  <Trophy className="w-8 h-8 text-amber-500 mx-auto mb-2 relative z-10 animate-bounce-gentle" />
                  <p className="text-xs font-black text-primary relative z-10 leading-tight">Continue ton parcours pour débloquer des trophées !</p>
                </div>
              </div>
            )}
          </div>
        </>
      ) : (
        /* No course */
        <div className="flex-1 flex flex-col items-center justify-center text-center px-4">
          <div className="w-20 h-20 bg-muted rounded-3xl flex items-center justify-center mx-auto mb-6 border-2 border-border">
            <GraduationCap className="w-10 h-10 text-muted-foreground" strokeWidth={2} />
          </div>
          <h2 className="text-3xl font-black text-foreground mb-3">Choisis ton Parcours</h2>
          <p className="text-muted-foreground font-semibold mb-8 max-w-xs mx-auto">
            Inscris-toi à un cours pour commencer ton aventure d'apprentissage !
          </p>
          <Button
            size="lg"
            onClick={() => router.push('/courses')}
            className="h-14 px-8 font-black text-base gap-2 bg-primary text-white shadow-lg shadow-primary/20 rounded-2xl border-b-4 border-primary-hover hover:-translate-y-1 active:translate-y-0 active:border-b-0 transition-all"
          >
            <Compass className="w-5 h-5" />
            Explorer les Parcours
          </Button>
        </div>
      )}

      <Dialog open={dialogConfig.open} onOpenChange={(open) => setDialogConfig(prev => ({ ...prev, open }))}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-xl font-black">{dialogConfig.title}</DialogTitle>
            <DialogDescription className="font-semibold text-base py-4 text-slate-600">
              {dialogConfig.description}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" onClick={() => setDialogConfig(prev => ({ ...prev, open: false }))} className="font-bold">
              {dialogConfig.onAction ? "Annuler" : "Compris"}
            </Button>
            {dialogConfig.actionLabel && dialogConfig.onAction && (
              <Button
                onClick={dialogConfig.onAction}
                disabled={dialogConfig.loadingAction}
                className="font-bold bg-green-500 hover:bg-green-600 text-white"
              >
                {dialogConfig.loadingAction ? "Chargement..." : dialogConfig.actionLabel}
              </Button>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
