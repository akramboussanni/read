'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { getCourseAdmin, addCourseNode } from '@/lib/api/admin';
import type { Course } from '@/lib/types/course';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Plus, RefreshCw } from 'lucide-react';
import { CourseFlowGraph } from '@/components/course/course-flow-graph';

export default function VisualEditorPage() {
    const { id } = useParams() as { id: string };
    const router = useRouter();

    const [course, setCourse] = useState<Course | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => { loadCourse(); }, [id]);

    const loadCourse = async () => {
        try {
            setLoading(true);
            setCourse(await getCourseAdmin(id));
        } catch (err) {
            console.error('Failed to load course', err);
        } finally {
            setLoading(false);
        }
    };

    const addNode = async () => {
        await addCourseNode(id, {
            title: 'Nouveau Nœud',
            node_type: 'quiz',
            description: '',
            icon: 'star',
            position_x: Math.random() * 400 + 100,
            position_y: Math.random() * 400 + 100,
            sort_order: course?.nodes?.length ?? 0,
        }).catch(console.error);
        loadCourse();
    };

    if (loading) return (
        <div className="flex items-center justify-center h-screen bg-background">
            <div className="flex flex-col items-center gap-4">
                <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                <p className="text-muted-foreground font-semibold text-sm">Chargement de l'éditeur...</p>
            </div>
        </div>
    );

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-background">
            {/* Header */}
            <div className="h-16 bg-card border-b border-border shadow-sm flex items-center justify-between px-6 shrink-0 z-10">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="sm" onClick={() => router.push('/admin')}>
                        <ArrowLeft className="w-4 h-4 mr-2" /> Retour
                    </Button>
                    <div>
                        <h1 className="text-lg font-bold">{course?.title} — Éditeur</h1>
                        <p className="text-xs text-muted-foreground">
                            {course?.nodes?.length ?? 0} nœuds · {course?.edges?.length ?? 0} connexions
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={addNode} className="font-semibold">
                        <Plus className="w-4 h-4 mr-2" /> Nœud
                    </Button>
                    <Button size="sm" onClick={loadCourse} className="font-semibold bg-primary">
                        <RefreshCw className="w-4 h-4 mr-2" /> Sync
                    </Button>
                </div>
            </div>

            {/* Graph */}
            {course && (
                <CourseFlowGraph
                    course={course}
                    mode="edit"
                    onGraphChange={loadCourse}
                />
            )}
        </div>
    );
}
