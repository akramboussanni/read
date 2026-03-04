'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { classroomApi } from '@/lib/api/classroom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Users, Plus, ArrowLeft, Shield, Sparkles, Shapes } from 'lucide-react';
import { motion } from 'framer-motion';

export default function CreateClassroomPage() {
    const router = useRouter();
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!name) return;
        setLoading(true);
        setError('');
        try {
            const classData = await classroomApi.createClassroom(name, description);
            router.push(`/classes/${classData.id}`);
        } catch (err) {
            console.error('Failed to create classroom:', err);
            setError("Erreur lors de la création de la classe. Réessaie !");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-background text-foreground pt-24 pb-32 relative">
            <div className="blob-green top-10 right-20" />
            <div className="blob-orange bottom-20 left-10" />

            <main className="container max-w-2xl mx-auto px-4 relative z-10">
                <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} className="mb-8">
                    <Button variant="ghost" onClick={() => router.back()} className="mb-4 pl-0 text-muted-foreground hover:text-foreground">
                        <ArrowLeft className="w-4 h-4 mr-2" /> Retour
                    </Button>
                    <h1 className="text-4xl font-black tracking-tight flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary rounded-2xl flex items-center justify-center text-white shadow-lg shadow-primary/20">
                            <Plus className="w-7 h-7" strokeWidth={2.5} />
                        </div>
                        Nouvelle classe
                    </h1>
                </motion.div>

                <Card className="fun-card border-primary/20 p-4">
                    <CardHeader className="text-center">
                        <div className="mx-auto w-16 h-16 bg-accent rounded-2xl flex items-center justify-center text-white shadow-lg shadow-accent/20 border-b-4 border-teal-600 mb-4 animate-float">
                            <Shapes className="w-8 h-8" />
                        </div>
                        <CardTitle className="text-2xl font-black">Prêt à enseigner ?</CardTitle>
                        <CardDescription className="text-base font-bold text-muted-foreground">Donne un nom sympa à ta classe pour tes élèves.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-6">
                            {error && (
                                <div className="p-4 text-sm font-bold text-red-600 bg-red-50 border-2 border-red-200 rounded-xl">
                                    {error}
                                </div>
                            )}

                            <div className="space-y-2">
                                <Label htmlFor="name" className="text-base font-bold flex items-center gap-2">
                                    <Users className="w-5 h-5 text-primary" />
                                    Nom de la classe <span className="text-red-500">*</span>
                                </Label>
                                <Input
                                    id="name"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    placeholder="Ex: Apprentis du Coran - Niveau 1"
                                    required
                                    className="h-12 text-base font-bold"
                                    maxLength={50}
                                />
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="description" className="text-base font-bold flex items-center gap-2">
                                    <Shield className="w-5 h-5 text-primary" />
                                    Description <span className="text-xs font-normal text-muted-foreground">(facultatif)</span>
                                </Label>
                                <Textarea
                                    id="description"
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    placeholder="De quoi parle la classe ? Quels sont les objectifs ?"
                                    className="min-h-[120px] text-base font-semibold"
                                    maxLength={250}
                                />
                            </div>

                            <Button type="submit" className="w-full h-14 text-xl font-black bg-primary text-white rounded-2xl border-b-6 border-primary-hover shadow-lg hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all gap-2" disabled={loading}>
                                {loading ? (
                                    <span className="flex items-center gap-2">
                                        <div className="w-6 h-6 border-4 border-white border-t-transparent rounded-full animate-spin" />
                                        CREATION...
                                    </span>
                                ) : (
                                    <span className="flex items-center gap-2">
                                        <Sparkles className="w-6 h-6 animate-pulse" />
                                        CREER LA CLASSE !
                                    </span>
                                )}
                            </Button>
                        </form>
                    </CardContent>
                </Card>
            </main>
        </div>
    );
}
