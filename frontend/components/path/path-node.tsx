'use client';

import { memo } from 'react';
import { Handle, Position, NodeProps, Node } from '@xyflow/react';
import { cn } from '@/lib/utils';
import { Lock, CheckCircle, Star, Play, BookOpen } from 'lucide-react';

// Using NodeProps generic is tricky in JS/TS without full setup, assuming standard props

interface PathNodeData extends Record<string, unknown> {
    label: string;
    status: string;
    type: string;
}

type CustomNode = Node<PathNodeData>;

const PathNode = ({ data, isConnectable }: NodeProps<CustomNode>) => {
    const { label, status, type } = data; // status: 'locked', 'unlocked', 'completed', 'current'

    const isLocked = status === 'locked';
    const isCompleted = status === 'completed';
    const isCurrent = status === 'current';

    return (
        <div className={cn(
            "relative group w-48 h-24 rounded-2xl border-2 transition-all duration-300 shadow-sm flex items-center justify-center p-4 cursor-pointer hover:scale-105 active:scale-95 select-none",
            isLocked ? "bg-muted border-border opacity-60 grayscale" :
                isCompleted ? "bg-emerald-50 border-emerald-200 shadow-emerald-100" :
                    isCurrent ? "bg-white border-primary shadow-lg shadow-primary/20 ring-4 ring-primary/10 animate-pulse-slow" :
                        "bg-white border-border hover:border-primary/50"
        )}>
            <Handle type="target" position={Position.Top} isConnectable={isConnectable} className="!bg-muted-foreground w-3 h-3 border-2 border-white" />

            <div className="flex flex-col items-center text-center gap-2">
                {isLocked && <Lock className="w-6 h-6 text-muted-foreground" />}
                {isCompleted && <CheckCircle className="w-6 h-6 text-emerald-500 fill-emerald-100" />}
                {isCurrent && <Play className="w-8 h-8 text-primary fill-primary/20 ml-1" />}
                {!isLocked && !isCompleted && !isCurrent && <BookOpen className="w-6 h-6 text-primary" />}

                <span className={cn(
                    "font-bold text-xs leading-tight line-clamp-2",
                    isLocked ? "text-muted-foreground" : "text-foreground"
                )}>
                    {label}
                </span>
            </div>

            <Handle type="source" position={Position.Bottom} isConnectable={isConnectable} className="!bg-muted-foreground w-3 h-3 border-2 border-white" />
        </div>
    );
};

export default memo(PathNode);
