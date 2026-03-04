'use client';

import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from './ui/button';
import {
  LogOut, User, BookOpen, ChevronDown, Shield,
  Plus, Users, LayoutDashboard,
  CheckCircle2, Shapes, Sparkles, Compass, ArrowRight,
  Coins
} from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { classroomApi, Classroom } from '@/lib/api/classroom';
import { courseApi } from '@/lib/api/course';
import { Course, UserEnrollment } from '@/lib/types/course';
import { cn } from '@/lib/utils';
import { AnimatePresence, motion } from 'framer-motion';

export function Header() {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, user, logout, refreshSession } = useAuthStore();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isClassesOpen, setIsClassesOpen] = useState(false);
  const [isCoursesOpen, setIsCoursesOpen] = useState(false);

  const [classes, setClasses] = useState<{ teaching: Classroom[], enrolled: Classroom[] }>({ teaching: [], enrolled: [] });
  const [enrollments, setEnrollments] = useState<UserEnrollment[]>([]);
  const [activeCourse, setActiveCourse] = useState<Course | null>(null);

  const dropdownRef = useRef<HTMLDivElement>(null);
  const classesRef = useRef<HTMLDivElement>(null);
  const coursesRef = useRef<HTMLDivElement>(null);

  const handleLogout = async () => {
    try {
      await logout();
      router.push('/');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  useEffect(() => {
    if (isAuthenticated) {
      classroomApi.listMyClasses().then(setClasses).catch(console.error);
      courseApi.getMyEnrollments().then(data => {
        setEnrollments(data);
        const active = data.find(e => e.course_id === user?.active_course_id)?.course;
        if (active) setActiveCourse(active);
      }).catch(console.error);
    }
  }, [isAuthenticated, user?.active_course_id]);

  const handleSwitchCourse = async (courseId: string) => {
    try {
      await courseApi.setActiveCourse(courseId);
      await refreshSession();
      setIsCoursesOpen(false);
      router.push('/');
    } catch (err) {
      console.error('Failed to switch course:', err);
    }
  };

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) setIsProfileOpen(false);
      if (classesRef.current && !classesRef.current.contains(event.target as Node)) setIsClassesOpen(false);
      if (coursesRef.current && !coursesRef.current.contains(event.target as Node)) setIsCoursesOpen(false);
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const isActive = (path: string) => pathname === path;

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-md border-b-4 border-border shadow-sm transition-all duration-300">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-20">
          <Link href="/" className="flex items-center gap-3 group select-none font-black text-2xl tracking-tight">
            <div className="relative bg-primary text-white p-2.5 rounded-2xl shadow-sm border-b-4 border-primary-hover group-hover:-translate-y-1 transition-transform">
              <Shapes className="h-7 w-7" strokeWidth={3} />
            </div>
            <div className="flex flex-col">
              <span>Iqra</span>
              <span className="text-[10px] text-muted-foreground uppercase tracking-widest -mt-1 hidden sm:block">Apprendre en jouant</span>
            </div>
          </Link>

          <nav className="flex items-center">
            {isAuthenticated ? (
              <div className="flex items-center gap-2 sm:gap-4">
                <Link href="/">
                  <Button variant="ghost" className={cn("hidden lg:flex font-bold text-base h-12 rounded-xl border-2 border-transparent hover:border-primary/20 hover:bg-primary-light hover:text-primary transition-all active:scale-95", isActive('/') && "bg-primary-light text-primary border-primary/20")}>
                    <LayoutDashboard className="h-5 w-5 mr-2" strokeWidth={2.5} /> Dashboard
                  </Button>
                </Link>

                <div className="relative" ref={coursesRef}>
                  <Button onClick={() => setIsCoursesOpen(!isCoursesOpen)} variant="ghost" className={cn("hidden sm:flex font-bold text-base h-12 rounded-xl border-2 border-transparent hover:border-primary/20 hover:bg-primary-light hover:text-primary transition-all active:scale-95", isCoursesOpen && "bg-primary-light text-primary border-primary/20")}>
                    {activeCourse ? (
                      <div className="flex items-center gap-2">
                        <div className="w-5 h-5 rounded-md border-b-2" style={{ background: activeCourse.color, borderColor: `${activeCourse.color}AA` }} />
                        <span className="max-w-[100px] truncate">{activeCourse.title}</span>
                      </div>
                    ) : <> <Compass className="h-5 w-5 mr-2" /> Parcours </>}
                    <ChevronDown className={cn("ml-2 h-4 w-4 transition-transform", isCoursesOpen && "rotate-180")} />
                  </Button>
                  <AnimatePresence>
                    {isCoursesOpen && (
                      <motion.div initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="absolute right-0 top-full mt-3 w-80 bg-white border-2 border-border rounded-2xl shadow-xl overflow-hidden z-50 border-b-4">
                        <div className="p-4 bg-muted border-b-2 flex justify-between items-center">
                          <span className="font-black text-[11px] uppercase tracking-[0.2em] text-muted-foreground">Mes Progrès</span>
                          <Link href="/courses">
                            <Button size="sm" variant="ghost" className="h-7 text-xs font-black text-primary hover:bg-primary-light">
                              Explorer
                            </Button>
                          </Link>
                        </div>
                        <div className="max-h-80 overflow-y-auto p-2 space-y-1.5">
                          {enrollments.length > 0 ? (
                            enrollments.map(en => (
                              <button
                                key={en.course_id}
                                onClick={() => handleSwitchCourse(en.course_id)}
                                className={cn(
                                  "w-full text-left p-3 rounded-xl transition-all border-2 border-transparent hover:border-primary/20 hover:bg-primary-light flex items-center gap-4 group",
                                  en.course_id === user?.active_course_id && "bg-primary-light/40 border-primary/20 ring-1 ring-primary/10"
                                )}
                              >
                                <div
                                  className="w-10 h-10 rounded-xl flex items-center justify-center text-white text-sm font-black shadow-sm border-b-4 transition-transform group-hover:-translate-y-0.5"
                                  style={{ background: en.course?.color, borderColor: `${en.course?.color}AA` }}
                                >
                                  {en.course?.title.charAt(0)}
                                </div>
                                <div className="flex-1 min-w-0">
                                  <div className="font-black text-sm text-slate-700 truncate">{en.course?.title}</div>
                                  <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Voir le parcours</div>
                                </div>
                                {en.course_id === user?.active_course_id ? (
                                  <div className="w-6 h-6 rounded-full bg-primary flex items-center justify-center">
                                    <CheckCircle2 className="w-4 h-4 text-white" />
                                  </div>
                                ) : (
                                  <ArrowRight className="w-4 h-4 text-slate-300 group-hover:translate-x-1 transition-transform" />
                                )}
                              </button>
                            ))
                          ) : (
                            <div className="p-10 text-center">
                              <Compass className="w-10 h-10 text-slate-200 mx-auto mb-3" />
                              <p className="text-sm font-bold text-slate-400">Aucun parcours commencé</p>
                            </div>
                          )}
                        </div>
                        <div className="p-2 border-t-2 bg-slate-50/50">
                          <Link href="/courses">
                            <Button className="w-full font-black text-xs h-10 bg-white border-2 border-border shadow-sm hover:border-primary/30 hover:text-primary rounded-xl">
                              Découvrir d'autres cours
                            </Button>
                          </Link>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>

                {((classes?.teaching?.length || 0) + (classes?.enrolled?.length || 0)) > 0 && (
                  <div className="relative" ref={classesRef}>
                    <Button onClick={() => setIsClassesOpen(!isClassesOpen)} variant="ghost" className={cn("hidden sm:flex font-bold text-base h-12 rounded-xl border-2 border-transparent hover:border-accent/20 hover:bg-accent-light hover:text-teal-600 transition-all active:scale-95", isClassesOpen && "bg-accent-light text-teal-600 border-accent/20")}>
                      <Users className="h-5 w-5 mr-2" /> Classes
                      <ChevronDown className={cn("ml-2 h-4 w-4 transition-transform", isClassesOpen && "rotate-180")} />
                    </Button>
                    <AnimatePresence>
                      {isClassesOpen && (
                        <motion.div initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="absolute right-0 top-full mt-3 w-80 bg-white border-2 border-border rounded-2xl shadow-xl overflow-hidden z-50 border-b-4">
                          <div className="p-3 bg-muted border-b-2 font-black text-[10px] uppercase tracking-wider text-muted-foreground flex justify-between items-center">
                            Tes Classes <Link href="/classes"><Button size="sm" variant="ghost" className="h-6 text-primary p-1">Voir tout</Button></Link>
                          </div>
                          <div className="max-h-64 overflow-y-auto p-2 space-y-2">
                            {classes?.teaching?.map(cls => (
                              <Link key={cls.id} href={`/classes/${cls.id}`} onClick={() => setIsClassesOpen(false)} className="block p-3 rounded-xl hover:bg-primary-light border-2 border-transparent hover:border-primary/20">
                                <div className="flex items-center gap-3">
                                  <div className="w-8 h-8 rounded-lg bg-primary text-white flex items-center justify-center font-black text-xs">{cls.name.charAt(0)}</div>
                                  <div className="font-bold text-sm truncate">{cls.name} <span className="text-[10px] text-primary block">Enseignant</span></div>
                                </div>
                              </Link>
                            ))}
                            {classes?.enrolled?.map(cls => (
                              <Link key={cls.id} href={`/classes/${cls.id}`} onClick={() => setIsClassesOpen(false)} className="block p-3 rounded-xl hover:bg-accent-light border-2 border-transparent hover:border-accent/20">
                                <div className="flex items-center gap-3">
                                  <div className="w-8 h-8 rounded-lg bg-teal-600 text-white flex items-center justify-center font-black text-xs">{cls.name.charAt(0)}</div>
                                  <div className="font-bold text-sm truncate">{cls.name} <span className="text-[10px] text-teal-600 block">Élève</span></div>
                                </div>
                              </Link>
                            ))}
                          </div>
                          <div className="p-3 bg-muted/40 border-t-2 grid grid-cols-2 gap-2">
                            <Link href="/classes/create" onClick={() => setIsClassesOpen(false)}><Button size="sm" variant="outline" className="w-full text-xs font-bold">Créer</Button></Link>
                            <Link href="/classes/join" onClick={() => setIsClassesOpen(false)}><Button size="sm" className="w-full text-xs font-black">Rejoindre</Button></Link>
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                )}

                {isAuthenticated && (
                  <div className="flex items-center gap-1.5 px-3 py-1.5 bg-yellow-50 border-2 border-yellow-200 rounded-xl shadow-sm hover:scale-105 transition-transform group cursor-help" title="Vos Pièces">
                    <div className="bg-yellow-400 p-1 rounded-full shadow-inner animate-pulse">
                      <Coins className="h-4 w-4 text-white" strokeWidth={3} />
                    </div>
                    <span className="font-black text-yellow-700 tabular-nums">
                      {user?.coins || 0}
                    </span>
                  </div>
                )}

                <div className="relative" ref={dropdownRef}>
                  <button onClick={() => setIsProfileOpen(!isProfileOpen)} className="w-11 h-11 rounded-[14px] bg-accent/20 flex items-center justify-center text-teal-700 font-black text-lg border-b-4 border-teal-600 shadow-sm bg-white hover:scale-105 transition-transform overflow-hidden">
                    {user?.username?.charAt(0).toUpperCase()}
                  </button>
                  <AnimatePresence>
                    {isProfileOpen && (
                      <motion.div initial={{ opacity: 0, y: 15, scale: 0.9 }} animate={{ opacity: 1, y: 0, scale: 1 }} exit={{ opacity: 0 }} className="absolute right-0 top-full mt-3 w-64 bg-white border-2 border-border rounded-2xl shadow-lg z-50 border-b-4 p-2 space-y-1">
                        <div className="px-3 py-2 border-b-2 mb-1">
                          <p className="font-black truncate">{user?.username}</p>
                          <p className="text-[10px] font-black text-primary uppercase">{user?.is_admin ? 'Admin' : 'Utilisateur'}</p>
                        </div>
                        <Link href="/profile" onClick={() => setIsProfileOpen(false)}><Button variant="ghost" className="w-full justify-start font-bold h-10 rounded-lg text-sm"><User className="mr-3 h-4 w-4" /> Profil</Button></Link>
                        <Link href="/classes" onClick={() => setIsProfileOpen(false)}><Button variant="ghost" className="w-full justify-start font-bold h-10 rounded-lg text-sm"><Users className="mr-3 h-4 w-4" /> Mes Classes</Button></Link>
                        {user?.is_admin && <Link href="/admin" onClick={() => setIsProfileOpen(false)}><Button variant="ghost" className="w-full justify-start font-bold h-10 rounded-lg text-sm text-info"><Shield className="mr-3 h-4 w-4" /> Admin</Button></Link>}
                        <Button onClick={handleLogout} variant="ghost" className="w-full justify-start font-bold h-10 rounded-lg text-sm text-danger"><LogOut className="mr-3 h-4 w-4" /> Déconnexion</Button>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <Link href="/login">
                  <Button variant="ghost" className="font-bold text-secondary hover:bg-secondary-light hover:text-secondary h-11 px-6 rounded-xl transition-all h-10 border-2 border-secondary shadow-sm">
                    Connexion
                  </Button>
                </Link>
                <Link href="/register">
                  <Button variant="secondary" className="h-11 px-8 font-black rounded-xl shadow-md h-10">
                    S'inscrire
                  </Button>
                </Link>
              </div>
            )}
          </nav>
        </div>
      </div>
    </header>
  );
}
