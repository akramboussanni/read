'use client';

import { useRouter } from 'next/navigation';
import { courseApi } from '@/lib/api/course';
import { ClassroomAssignment } from '@/lib/api/classroom';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { BookOpen, Calendar, CheckCircle2, AlertCircle, PlayCircle, Clock } from 'lucide-react';

interface AssignmentCardProps {
    assignment: ClassroomAssignment;
}

export function AssignmentCard({ assignment: asgn }: AssignmentCardProps) {
    const router = useRouter();
    const now = Date.now() / 1000;
    const isOverdue = asgn.due_date > 0 && asgn.due_date < now && !asgn.is_completed;
    const daysLeft = asgn.due_date > 0 ? Math.ceil((asgn.due_date - now) / 86400) : null;

    const handleGo = async () => {
        try {
            const fullC = await courseApi.getCourse(asgn.course_id);
            const n = fullC.nodes?.find((x) => x.id === asgn.node_id);
            if (n && n.node_type === 'quiz') {
                const conf = JSON.parse(n.quiz_config || '{}');
                if (conf.quiz_id) {
                    router.push(`/quizzes/${conf.quiz_id}?courseId=${asgn.course_id}&nodeId=${asgn.node_id}&asgnId=${asgn.id}`);
                    return;
                }
            }
            alert("Impossible d'ouvrir ce devoir directement.");
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <Card
            className={`fun-card overflow-hidden group border-2 transition-colors ${
                asgn.is_completed
                    ? 'border-green-200 bg-green-50/20'
                    : isOverdue
                    ? 'border-red-200 bg-red-50/20'
                    : 'border-border hover:border-primary/40'
            }`}
        >
            <CardContent className="p-0">
                <div className="flex flex-col sm:flex-row">
                    <div
                        className={`w-full sm:w-20 border-r border-border flex flex-col items-center justify-center p-4 ${
                            asgn.is_completed ? 'bg-green-50' : isOverdue ? 'bg-red-50' : 'bg-orange-50/50'
                        }`}
                    >
                        {asgn.is_completed ? (
                            <>
                                <CheckCircle2 className="w-6 h-6 text-green-600 mb-1" />
                                <span className="text-[10px] font-black uppercase text-green-700 whitespace-nowrap">Fait ✓</span>
                            </>
                        ) : isOverdue ? (
                            <>
                                <AlertCircle className="w-6 h-6 text-red-500 mb-1" />
                                <span className="text-[10px] font-black uppercase text-red-600 whitespace-nowrap">Retard</span>
                            </>
                        ) : (
                            <>
                                <Clock className="w-6 h-6 text-orange-500 mb-1" />
                                <span className="text-[10px] font-black uppercase text-orange-600 whitespace-nowrap">À faire</span>
                            </>
                        )}
                    </div>
                    <div className="flex-1 p-5 flex flex-col sm:flex-row items-center justify-between gap-4">
                        <div className="space-y-1.5">
                            <h3 className="text-lg font-black text-slate-800">{asgn.title}</h3>
                            <p className="text-sm text-muted-foreground font-semibold flex items-center gap-2">
                                <BookOpen className="w-3.5 h-3.5" />
                                {asgn.course_name || `Parcours #${asgn.course_id}`}
                                {asgn.node_title && (
                                    <span className="text-slate-400">→ {asgn.node_title}</span>
                                )}
                            </p>
                            {asgn.due_date > 0 && (
                                <div
                                    className={`flex items-center gap-1.5 text-xs font-bold w-fit px-2 py-0.5 rounded-md ${
                                        asgn.is_completed
                                            ? 'text-green-600 bg-green-100'
                                            : isOverdue
                                            ? 'text-red-600 bg-red-100'
                                            : 'text-orange-600 bg-orange-100'
                                    }`}
                                >
                                    <Calendar className="w-3 h-3" />
                                    {new Date(asgn.due_date * 1000).toLocaleDateString('fr-FR', {
                                        day: 'numeric',
                                        month: 'short',
                                    })}
                                    {!asgn.is_completed && daysLeft !== null && (
                                        <span className="opacity-75">
                                            (
                                            {daysLeft < 0
                                                ? `${Math.abs(daysLeft)}j retard`
                                                : daysLeft === 0
                                                ? "Aujourd'hui"
                                                : `${daysLeft}j`}
                                            )
                                        </span>
                                    )}
                                </div>
                            )}
                        </div>
                        <div className="flex items-center gap-3">
                            {asgn.is_completed ? (
                                <div className="flex items-center gap-1.5 text-green-600 bg-green-50 px-3 py-1.5 rounded-xl border-2 border-green-200">
                                    <CheckCircle2 className="w-4 h-4" />
                                    <span className="text-xs font-black uppercase">
                                        {asgn.score_percent ? `${Math.round(asgn.score_percent)}%` : 'Fait'}
                                    </span>
                                </div>
                            ) : (
                                <Button
                                    onClick={handleGo}
                                    className="font-black bg-primary text-white rounded-xl shadow-md border-b-2 border-primary-hover h-10 px-5"
                                >
                                    <PlayCircle className="w-4 h-4 mr-2" /> Go
                                </Button>
                            )}
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
