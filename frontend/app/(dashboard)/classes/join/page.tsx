'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { classroomApi } from '@/lib/api/classroom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Users, GraduationCap, ArrowLeft, Key, Sparkles, CheckCircle2, Plus } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

import { useAuthStore } from '@/lib/store/auth-store';

export default function JoinClassroomPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [code, setCode] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [joined, setJoined] = useState<{ id: string, name: string } | null>(null);

    // Teachers shouldn't join classes
    React.useEffect(() => {
        if (user?.role === 'teacher') {
            router.replace('/classes');
        }
    }, [user, router]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!code) return;
        setLoading(true);
        setError('');
        try {
            const result = await classroomApi.joinClassroom(code);
            setJoined({ id: result.id, name: result.name });
            // Redirect after a short delay
            setTimeout(() => {
                router.push(`/classes/${result.id}`);
            }, 2000);
        } catch (err: any) {
            console.error('Failed to join classroom:', err);
            setError(err.response?.data?.error || "Code invalide ou classe verrouillée. Vérifie avec ton enseignant !");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32 relative">
            <div className="blob-green top-10 right-20" />
            <div className="blob-orange bottom-20 left-10" />

            <main className="container max-w-xl mx-auto px-4 relative z-10">
                <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} className="mb-8">
                    <Button variant="ghost" onClick={() => router.back()} className="mb-4 pl-0 text-muted-foreground hover:text-foreground">
                        <ArrowLeft className="w-4 h-4 mr-2" /> Retour
                    </Button>
                    <h1 className="text-4xl font-black tracking-tight flex items-center gap-3">
                        <div className="w-12 h-12 bg-secondary rounded-2xl flex items-center justify-center text-white shadow-lg shadow-orange-500/20">
                            <Key className="w-7 h-7" strokeWidth={2.5} />
                        </div>
                        Rejoindre une classe
                    </h1>
                </motion.div>

                <AnimatePresence mode="wait">
                    {!joined ? (
                        <motion.div key="form" initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.9 }}>
                            <Card className="fun-card border-secondary/20 p-4">
                                <CardHeader className="text-center">
                                    <div className="mx-auto w-16 h-16 bg-secondary rounded-2xl flex items-center justify-center text-white shadow-lg shadow-orange-500/20 border-b-4 border-orange-600 mb-4 animate-wobble">
                                        <GraduationCap className="w-8 h-8" strokeWidth={2.5} />
                                    </div>
                                    <CardTitle className="text-2xl font-black text-foreground mb-1">Entrer le code secret</CardTitle>
                                    <CardDescription className="text-base font-bold text-muted-foreground">Demande le code à ton enseignant pour rejoindre sa classe.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <form id="join-form" onSubmit={handleSubmit} className="space-y-6">
                                        {error && (
                                            <div className="p-4 text-sm font-bold text-red-600 bg-red-50 border-2 border-red-200 rounded-xl animate-bounce-gentle">
                                                {error}
                                            </div>
                                        )}

                                        <div className="space-y-4">
                                            <Label htmlFor="code" className="sr-only">Code de classe</Label>
                                            <Input
                                                id="code"
                                                value={code}
                                                onChange={(e) => setCode(e.target.value.toUpperCase())}
                                                placeholder="ABC-123"
                                                required
                                                className="h-16 text-2xl font-black text-center uppercase tracking-widest border-2 border-muted hover:border-secondary focus:border-secondary transition-all"
                                                maxLength={10}
                                            />
                                            <p className="text-xs text-center text-muted-foreground font-black uppercase tracking-widest">
                                                Généralement 6-8 caractères
                                            </p>
                                        </div>
                                    </form>
                                </CardContent>
                                <CardFooter>
                                    <Button form="join-form" type="submit" className="w-full h-14 text-xl font-black bg-secondary text-white rounded-2xl border-b-6 border-orange-600 shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all gap-2" disabled={loading}>
                                        {loading ? (
                                            <span className="flex items-center gap-2">
                                                <div className="w-6 h-6 border-4 border-white border-t-transparent rounded-full animate-spin" />
                                                CONNEXION...
                                            </span>
                                        ) : (
                                            <span className="flex items-center gap-2">
                                                <Plus className="w-6 h-6" strokeWidth={3} />
                                                REJOINDRE !
                                            </span>
                                        )}
                                    </Button>
                                </CardFooter>
                            </Card>
                        </motion.div>
                    ) : (
                        <motion.div key="success" initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="text-center p-12 bg-white rounded-3xl border-2 border-teal-500 shadow-xl space-y-6">
                            <div className="w-20 h-20 bg-teal-500 text-white rounded-full flex items-center justify-center mx-auto shadow-lg shadow-teal-500/40 border-b-6 border-teal-700">
                                <CheckCircle2 className="w-12 h-12" strokeWidth={3} />
                            </div>
                            <div className="space-y-2">
                                <h1 className="text-3xl font-black text-foreground tracking-tight">C'est gagné ! 🎉</h1>
                                <p className="text-lg font-bold text-teal-600">Bienvenue dans la classe <span className="underline underline-offset-4 decoration-4">{joined.name}</span> !</p>
                            </div>
                            <div className="pt-4 flex flex-col items-center gap-3">
                                <div className="w-12 h-1 pt-0.5 bg-zinc-100 rounded-full overflow-hidden">
                                    <motion.div initial={{ x: '-100%' }} animate={{ x: 0 }} transition={{ duration: 1.5, repeat: Infinity, ease: 'linear' }} className="h-full w-1/3 bg-teal-500" />
                                </div>
                                <p className="text-xs text-muted-foreground font-black uppercase tracking-widest">On te prépare ta place...</p>
                            </div>
                        </motion.div>
                    )}
                </AnimatePresence>
            </main>
        </div>
    );
}
