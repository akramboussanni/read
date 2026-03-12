'use client';

import { useRouter } from 'next/navigation';
import { Classroom } from '@/lib/api/classroom';
import { ArrowRight, Key, AlertCircle } from 'lucide-react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

interface ClassCardProps {
    class: Classroom;
    type: 'teaching' | 'enrolled';
    delay?: number;
}

export function ClassCard({ class: cls, type, delay = 0 }: ClassCardProps) {
    const router = useRouter();
    return (
        <motion.div
            initial={{ opacity: 0, x: type === 'teaching' ? -20 : 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay }}
            onClick={() => router.push(`/classes/${cls.id}`)}
            className={cn(
                'group relative bg-card rounded-2xl border-2 p-5 cursor-pointer transition-all hover:shadow-lg active:scale-[0.98]',
                type === 'teaching'
                    ? 'border-primary/20 hover:border-primary/50'
                    : 'border-secondary/20 hover:border-secondary/50'
            )}
        >
            <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-4">
                    <div
                        className={cn(
                            'w-14 h-14 rounded-xl flex items-center justify-center font-black text-xl border-b-4 shadow-sm group-hover:-translate-y-1 transition-transform',
                            type === 'teaching'
                                ? 'bg-primary text-white border-primary-hover'
                                : 'bg-secondary text-white border-orange-600'
                        )}
                    >
                        {cls.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                        <h3 className="text-lg font-black group-hover:text-primary transition-colors">{cls.name}</h3>
                        <p className="text-sm text-muted-foreground line-clamp-1">{cls.description || 'Aucune description'}</p>
                    </div>
                </div>
                <ArrowRight className="w-5 h-5 text-muted-foreground group-hover:text-foreground group-hover:translate-x-1 transition-all" />
            </div>

            <div className="mt-4 pt-4 border-t border-muted flex items-center justify-between">
                {type === 'teaching' ? (
                    <div className="flex items-center gap-3">
                        <div className="bg-muted px-2 py-1 rounded-lg flex items-center gap-1.5">
                            <Key className="w-3.5 h-3.5 text-primary" />
                            <span className="text-xs font-black tracking-widest uppercase">{cls.join_code}</span>
                        </div>
                        <span className="text-xs font-bold text-muted-foreground">Appuyer pour gérer</span>
                    </div>
                ) : (
                    <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-muted-foreground italic">Relève les défis !</span>
                    </div>
                )}

                {cls.is_locked && (
                    <div className="text-xs font-bold text-red-500 flex items-center gap-1">
                        <AlertCircle className="w-3 h-3" /> Verrouillé
                    </div>
                )}
            </div>
        </motion.div>
    );
}
