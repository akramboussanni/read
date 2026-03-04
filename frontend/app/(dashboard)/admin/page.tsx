'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { getSystemStats, listCoursesAdmin } from '@/lib/api/admin';
import type { SystemStatsResponse } from '@/lib/types/admin';
import type { Course } from '@/lib/types/course';
import Link from 'next/link';
import {
  Users, BookOpen, Shield, Shapes,
  TrendingUp, Activity, Plus, Settings,
  ArrowRight, Sparkles, LayoutDashboard
} from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

export default function AdminDashboard() {
  const router = useRouter();
  const [systemStats, setSystemStats] = useState<SystemStatsResponse | null>(null);
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [stats, coursesData] = await Promise.all([
        getSystemStats(),
        listCoursesAdmin(),
      ]);
      setSystemStats(stats);
      setCourses(coursesData || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load admin data');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center pt-20">
        <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin mb-4" />
        <p className="text-primary font-black animate-pulse tracking-widest uppercase text-xs">Accès à la zone secrète...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="fun-card border-red-200 p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-red-100 rounded-2xl flex items-center justify-center mx-auto mb-4 border-b-4 border-red-500">
            <Shield className="w-8 h-8 text-red-600" />
          </div>
          <h2 className="text-2xl font-black text-slate-800 mb-2">Accès Refusé</h2>
          <p className="text-muted-foreground font-bold mb-6">{error}</p>
          <Button onClick={() => router.push('/')} className="w-full bg-slate-800 text-white font-black h-12 rounded-xl border-b-4 border-slate-950">
            Retour à l'accueil
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground pt-24 pb-32 relative overflow-hidden">
      {/* Background Decorations */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div className="blob-green -top-20 -left-20 opacity-20" />
        <div className="blob-orange -bottom-20 -right-20 opacity-10" />
        <div className="blob-teal top-1/2 right-10 opacity-10" />
      </div>

      <main className="container max-w-6xl mx-auto px-4 relative z-10 space-y-10">
        <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="flex items-center gap-5">
            <div className="w-16 h-16 bg-indigo-600 text-white rounded-2xl flex items-center justify-center shadow-xl border-b-6 border-indigo-800 animate-float">
              <Shield className="w-8 h-8" strokeWidth={2.5} />
            </div>
            <div>
              <h1 className="text-4xl font-black tracking-tight">Zone Secrète</h1>
              <p className="text-muted-foreground font-black uppercase tracking-widest text-[10px]">Administration & Gestion</p>
            </div>
          </div>
          <Link href="/admin/courses/new">
            <Button className="bg-primary text-white font-black h-14 px-8 rounded-2xl border-b-6 border-primary-hover shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all text-lg gap-2">
              <Plus className="w-6 h-6" /> NOUVEAU COURS
            </Button>
          </Link>
        </motion.div>

        {/* System Statistics */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card className="fun-card border-blue-200 bg-blue-50/10 p-6 flex flex-col justify-between group">
            <div className="flex items-center justify-between mb-4">
              <div className="w-10 h-10 bg-blue-100 rounded-xl flex items-center justify-center text-blue-600 border-b-2 border-blue-300 group-hover:scale-110 transition-transform">
                <Users className="w-5 h-5" />
              </div>
              <span className="text-[10px] font-black uppercase tracking-widest text-blue-500">Utilisateurs</span>
            </div>
            <div>
              <div className="text-4xl font-black text-slate-800 mb-1">{systemStats?.total_users || 0}</div>
              <div className="text-xs font-bold text-slate-500 flex items-center gap-1">
                <TrendingUp className="w-3 h-3 text-emerald-500" /> {systemStats?.active_users_7d || 0} actifs ces 7j
              </div>
            </div>
          </Card>

          <Card className="fun-card border-violet-200 bg-violet-50/10 p-6 flex flex-col justify-between group">
            <div className="flex items-center justify-between mb-4">
              <div className="w-10 h-10 bg-violet-100 rounded-xl flex items-center justify-center text-violet-600 border-b-2 border-violet-300 group-hover:scale-110 transition-transform">
                <BookOpen className="w-5 h-5" />
              </div>
              <span className="text-[10px] font-black uppercase tracking-widest text-violet-500">Parcours</span>
            </div>
            <div>
              <div className="text-4xl font-black text-slate-800 mb-1">{courses.length}</div>
              <div className="text-xs font-bold text-slate-500 flex items-center gap-1">
                <Activity className="w-3 h-3 text-primary" /> Contenu disponible
              </div>
            </div>
          </Card>

          <Card className="fun-card border-emerald-200 bg-emerald-50/10 p-6 flex flex-col justify-between group">
            <div className="flex items-center justify-between mb-4">
              <div className="w-10 h-10 bg-emerald-100 rounded-xl flex items-center justify-center text-emerald-600 border-b-2 border-emerald-300 group-hover:scale-110 transition-transform">
                <Shapes className="w-5 h-5" />
              </div>
              <span className="text-[10px] font-black uppercase tracking-widest text-emerald-500">Classes</span>
            </div>
            <div>
              <div className="text-4xl font-black text-slate-800 mb-1">12</div>
              <div className="text-xs font-bold text-slate-500">Salles virtuelles</div>
            </div>
          </Card>

          <Link href="/admin/quizzes" className="block">
            <Card className="fun-card border-orange-200 bg-orange-50/10 p-6 h-full flex flex-col justify-between group cursor-pointer hover:border-orange-400">
              <div className="flex items-center justify-between mb-4">
                <div className="w-10 h-10 bg-orange-100 rounded-xl flex items-center justify-center text-orange-600 border-b-2 border-orange-300 group-hover:scale-110 transition-transform">
                  <Sparkles className="w-5 h-5" />
                </div>
                <span className="text-[10px] font-black uppercase tracking-widest text-orange-500">Évaluation</span>
              </div>
              <div>
                <div className="text-xl font-black text-slate-800 mb-1">Gérer Quiz</div>
                <div className="text-xs font-bold text-slate-500 flex items-center gap-1">
                  Aller à l'éditeur <ArrowRight className="w-3 h-3" />
                </div>
              </div>
            </Card>
          </Link>
        </div>

        {/* Top Courses */}
        <Card className="fun-card border-slate-200 p-0 overflow-hidden shadow-xl bg-white">
          <div className="p-6 border-b-2 border-border flex items-center justify-between bg-muted/30">
            <div className="flex items-center gap-3">
              <LayoutDashboard className="w-6 h-6 text-slate-400" />
              <h2 className="text-xl font-black">Gestion des Parcours</h2>
            </div>
            <div className="text-xs font-black text-muted-foreground uppercase tracking-widest bg-white px-3 py-1 rounded-full border">
              {courses.length} PARCOURS
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="bg-muted/10 border-b text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                  <th className="py-4 px-6">Contenu</th>
                  <th className="py-4 px-6 text-center">Noeuds</th>
                  <th className="py-4 px-6 text-center">Statut</th>
                  <th className="py-4 px-6 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y border-b-2">
                {courses.map((course) => (
                  <tr key={course.id} className="hover:bg-muted/5 transition-colors group">
                    <td className="py-6 px-6">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-xl flex items-center justify-center text-white font-black text-lg shadow-sm border-b-4 group-hover:-translate-y-1 transition-transform" style={{ background: course.color || '#333', borderColor: `${course.color || '#333'}AA` }}>
                          {course.title.charAt(0)}
                        </div>
                        <div>
                          <div className="font-black text-lg group-hover:text-primary transition-colors">{course.title}</div>
                          <div className="text-xs font-bold text-muted-foreground line-clamp-1">{course.description || "Aucune description"}</div>
                        </div>
                      </div>
                    </td>
                    <td className="py-6 px-6 text-center">
                      <span className="text-sm font-black text-slate-600 bg-muted px-2.5 py-1 rounded-lg border-2 border-border shadow-inner">
                        {course.nodes?.length || 0}
                      </span>
                    </td>
                    <td className="py-6 px-6 text-center">
                      {course.is_published ? (
                        <span className="text-emerald-700 font-black text-[10px] uppercase tracking-widest bg-emerald-100 px-3 py-1 rounded-full border border-emerald-200">En Ligne</span>
                      ) : (
                        <span className="text-amber-700 font-black text-[10px] uppercase tracking-widest bg-amber-100 px-3 py-1 rounded-full border border-amber-200">Brouillon</span>
                      )}
                    </td>
                    <td className="py-6 px-6 text-right">
                      <Link href={`/admin/courses/${course.id}/visual-editor`}>
                        <Button className="h-10 px-6 bg-primary text-white font-black rounded-xl border-b-4 border-primary-hover hover:scale-105 transition-all text-xs">
                          OUVRIR L'ÉDITEUR
                        </Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </main>
    </div>
  );
}
