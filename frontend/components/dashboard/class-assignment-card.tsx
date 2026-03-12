'use client';

import { useRouter } from 'next/navigation';
import { courseApi } from '@/lib/api/course';
import { ClassroomAssignment } from '@/lib/api/classroom';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
    BookOpen, Calendar, CheckCircle2, AlertCircle, PlayCircle,
    CalendarClock, BarChart3, Pencil, Trash2, Users, Trophy, RefreshCw, XCircle, Navigation
} from 'lucide-react';
import { cn } from '@/lib/utils';

export function formatDueDate(ts: number): string {
    if (!ts) return '';
    const d = new Date(ts * 1000);
    return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' });
}

export function getDueStatus(ts: number, isCompleted: boolean): 'completed' | 'overdue' | 'upcoming' | 'none' {
    if (isCompleted) return 'completed';
    if (!ts) return 'none';
    const now = Date.now() / 1000;
    if (ts < now) return 'overdue';
    return 'upcoming';
}

export function getDaysLeft(ts: number): string {
    if (!ts) return '';
    const now = Date.now() / 1000;
    const diff = ts - now;
    const days = Math.ceil(diff / 86400);
    if (days < 0) return `${Math.abs(days)}j en retard`;
    if (days === 0) return "Aujourd'hui !";
    return `${days}j restants`;
}

interface ClassAssignmentCardProps {
    assignment: ClassroomAssignment & {
        course_name?: string;
        node_title?: string;
        completed_count?: number;
        total_students?: number;
        completed_at?: number;
    };
    isTeacher: boolean;
    onStats?: (asgnId: string, asgn: any) => void;
    onEdit?: (asgn: any) => void;
    onDelete?: (asgn: any) => void;
}

export function ClassAssignmentCard({ assignment: asgn, isTeacher, onStats, onEdit, onDelete }: ClassAssignmentCardProps) {
    const router = useRouter();
    const dueStatus = getDueStatus(asgn.due_date, asgn.is_completed ?? false);
    const isPassing = asgn.is_completed && (asgn.score_percent ?? 0) >= (asgn.passing_grade ?? 70);
    const isFailed = asgn.is_completed && (asgn.score_percent ?? 0) < (asgn.passing_grade ?? 70);
    const isPathProgress = asgn.assignment_type === 'path_progress';

    const handlePlay = async () => {
        if (isPathProgress) {
            // Navigate to the home page where the course graph is shown
            router.push('/');
            return;
        }
        try {
            const fullC = await courseApi.getCourse(asgn.course_id);
            const n = fullC.nodes?.find(x => x.id === asgn.node_id);
            if (n && n.node_type === 'quiz') {
                const conf = JSON.parse(n.quiz_config || '{}');
                if (conf.quiz_id) {
                    router.push(`/quizzes/${conf.quiz_id}?courseId=${asgn.course_id}&nodeId=${asgn.node_id}&asgnId=${asgn.id}`);
                    return;
                }
            }
            alert("Désolé, impossible d'ouvrir ce quizz directement.");
        } catch (e) {
            console.error('Failed to fetch course to redirect', e);
        }
    };

    return (
        <Card
            className={cn(
                'fun-card overflow-hidden group transition-all',
                isPassing && 'border-green-200 bg-green-50/20',
                isFailed && 'border-amber-300 bg-amber-50/30',
                dueStatus === 'overdue' && !isTeacher && !asgn.is_completed && 'border-red-200 bg-red-50/20',
            )}
        >
            <CardContent className="p-0">
                <div className="flex flex-col sm:flex-row">
                    {/* Status sidebar */}
                    <div
                        className={cn(
                            'w-full sm:w-24 border-r border-border flex flex-col items-center justify-center p-4 gap-1',
                            isPassing && 'bg-green-50 border-green-200',
                            isFailed && 'bg-amber-50 border-amber-200',
                            dueStatus === 'overdue' && !asgn.is_completed && 'bg-red-50 border-red-200',
                            dueStatus === 'upcoming' && 'bg-muted/30',
                            dueStatus === 'none' && 'bg-muted/30',
                        )}
                    >
                        {isPassing ? (
                            <>
                                <CheckCircle2 className="w-6 h-6 text-green-600" />
                                <span className="text-[10px] font-black uppercase text-green-700 whitespace-nowrap">Fait ✓</span>
                            </>
                        ) : isFailed ? (
                            <>
                                <XCircle className="w-6 h-6 text-amber-500" />
                                <span className="text-[10px] font-black uppercase text-amber-600 whitespace-nowrap">Échoué</span>
                            </>
                        ) : dueStatus === 'overdue' ? (
                            <>
                                <AlertCircle className="w-6 h-6 text-red-500" />
                                <span className="text-[10px] font-black uppercase text-red-600 whitespace-nowrap">En retard</span>
                            </>
                        ) : (
                            <>
                                <CalendarClock className="w-6 h-6 text-muted-foreground" />
                                <span className="text-[10px] font-black uppercase text-muted-foreground whitespace-nowrap">À faire</span>
                            </>
                        )}
                    </div>

                    {/* Content */}
                    <div className="flex-1 p-5 flex flex-col gap-3">
                        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                            <div className="space-y-1.5">
                                <h3 className="text-lg font-black group-hover:text-primary transition-colors">{asgn.title}</h3>
                                <div className="flex flex-wrap items-center gap-3 text-sm">
                                    {isPathProgress && (
                                        <span className="inline-flex items-center gap-1 text-[10px] font-black px-2 py-0.5 rounded-full bg-teal-100 text-teal-700 border border-teal-200 uppercase tracking-wide">
                                            <Navigation className="w-3 h-3" /> Parcours
                                        </span>
                                    )}
                                    <span className="flex items-center gap-1.5 text-muted-foreground font-semibold">
                                        <BookOpen className="w-3.5 h-3.5" />
                                        {asgn.course_name || asgn.course_id}
                                    </span>
                                    {asgn.node_title && (
                                        <span className="text-muted-foreground font-medium">
                                            {isPathProgress ? '→ Atteindre « ' : '→ '}{asgn.node_title}{isPathProgress ? ' »' : ''}
                                        </span>
                                    )}
                                </div>
                                {asgn.description && (
                                    <p className="text-sm text-slate-500 font-medium">{asgn.description}</p>
                                )}
                            </div>

                            {/* Actions */}
                            <div className="flex items-center gap-2 shrink-0">
                                {asgn.is_completed && !isTeacher && asgn.score_percent !== undefined && (
                                    <div
                                        className={cn(
                                            'flex items-center gap-1.5 px-3 py-1.5 rounded-xl border-2 text-xs font-black',
                                            isPassing
                                                ? 'text-green-700 bg-green-50 border-green-200'
                                                : 'text-amber-700 bg-amber-50 border-amber-300',
                                        )}
                                    >
                                        {isPassing
                                            ? <CheckCircle2 className="w-4 h-4" />
                                            : <XCircle className="w-4 h-4" />}
                                        {Math.round(asgn.score_percent ?? 0)}%
                                        {isFailed && (
                                            <span className="opacity-70">(min {asgn.passing_grade ?? 70}%)</span>
                                        )}
                                    </div>
                                )}
                                <Button
                                    variant="outline"
                                    onClick={handlePlay}
                                    className="rounded-xl border-primary/20 text-primary font-bold hover:bg-primary/5 transition-all"
                                >
                                    {isPathProgress
                                        ? <><Navigation className="mr-2 w-4 h-4" />{isTeacher ? 'Voir parcours' : asgn.is_completed ? 'Parcours' : 'Continuer'}</>
                                        : <><PlayCircle className="mr-2 w-4 h-4" />{isTeacher ? 'Aperçu' : asgn.is_completed ? 'Refaire' : 'Go'}</>
                                    }
                                </Button>

                                {isTeacher && (
                                    <>
                                        <Button
                                            variant="outline"
                                            onClick={() => onStats?.(asgn.id, asgn)}
                                            className="rounded-xl font-bold transition-all"
                                        >
                                            <BarChart3 className="mr-2 w-4 h-4" /> Stats
                                        </Button>
                                        <Button
                                            variant="outline"
                                            size="icon"
                                            onClick={() => onEdit?.(asgn)}
                                            className="rounded-xl transition-all"
                                            title="Modifier"
                                        >
                                            <Pencil className="w-4 h-4" />
                                        </Button>
                                        <Button
                                            variant="outline"
                                            size="icon"
                                            onClick={() => onDelete?.(asgn)}
                                            className="rounded-xl border-red-200 text-red-500 hover:bg-red-50 transition-all"
                                            title="Supprimer"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </Button>
                                    </>
                                )}
                            </div>
                        </div>

                        {/* Footer badges — unified muted style */}
                        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border/50">
                            {asgn.due_date > 0 && (
                                <span className={cn(
                                    'inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-muted/60 text-muted-foreground',
                                    dueStatus === 'overdue' && !asgn.is_completed && 'bg-red-100 text-red-600',
                                )}>
                                    <Calendar className="w-3 h-3" />
                                    {formatDueDate(asgn.due_date)}
                                    {dueStatus !== 'completed' && (
                                        <span className="opacity-70">({getDaysLeft(asgn.due_date)})</span>
                                    )}
                                </span>
                            )}
                            {isTeacher && (asgn as any).completed_count !== undefined && (
                                <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-muted/60 text-muted-foreground">
                                    <Users className="w-3 h-3" />
                                    {(asgn as any).completed_count}/{(asgn as any).total_students} complétés
                                </span>
                            )}
                            <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-muted/60 text-muted-foreground">
                                <Trophy className="w-3 h-3" />
                                Seuil: {asgn.passing_grade ?? 70}%
                            </span>
                            {!isTeacher && (
                                <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-muted/60 text-muted-foreground">
                                    <RefreshCw className="w-3 h-3" />
                                    {asgn.max_retakes === -1
                                        ? 'Reprises illimitées'
                                        : asgn.max_retakes === 0
                                        ? 'Pas de reprise'
                                        : `${asgn.max_retakes} reprise${asgn.max_retakes > 1 ? 's' : ''}`}
                                </span>
                            )}
                            {asgn.is_completed && (asgn as any).completed_at > 0 && !isTeacher && (
                                <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg bg-muted/60 text-muted-foreground">
                                    <CheckCircle2 className="w-3 h-3" />
                                    Terminé le {formatDueDate((asgn as any).completed_at)}
                                </span>
                            )}
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
