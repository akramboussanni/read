'use client';

import { useAuthStore } from '@/lib/store/auth-store';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Trophy } from 'lucide-react';

interface Student {
    id: string;
    username: string;
}

interface ClassLeaderboardProps {
    students: Student[];
}

export function ClassLeaderboard({ students }: ClassLeaderboardProps) {
    const { user } = useAuthStore();

    return (
        <Card className="fun-card border-orange-200 bg-orange-50/30">
            <CardHeader>
                <CardTitle className="text-lg font-black text-orange-700 flex items-center gap-2">
                    <Trophy className="w-5 h-5" /> Top Classe
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
                {students?.slice(0, 5).map((s, i) => (
                    <div
                        key={s.id}
                        className="flex items-center justify-between p-3 bg-white rounded-xl border border-orange-100 shadow-sm relative overflow-hidden group"
                    >
                        <div className="absolute inset-0 bg-orange-50 translate-x-full group-hover:translate-x-0 transition-transform" />
                        <div className="relative flex items-center gap-3">
                            <div className="w-8 h-8 rounded-lg bg-orange-100 flex items-center justify-center font-black text-orange-700 text-xs shadow-inner">
                                #{i + 1}
                            </div>
                            <span className="font-bold text-sm text-slate-800">{s.username}</span>
                        </div>
                        {s.username === user?.username && (
                            <span className="relative text-[10px] font-black uppercase text-primary tracking-wider bg-primary/10 px-2 py-0.5 rounded-md">
                                Toi
                            </span>
                        )}
                    </div>
                ))}
                {(!students || students.length === 0) && (
                    <p className="text-center py-4 text-xs font-bold text-orange-400">Aucun élève trouvé.</p>
                )}
            </CardContent>
        </Card>
    );
}
