interface EmptySectionProps {
    title: string;
    desc: string;
}

export function EmptySection({ title, desc }: EmptySectionProps) {
    return (
        <div className="p-10 text-center border-2 border-dashed border-border rounded-2xl bg-muted/20">
            <h3 className="font-bold text-foreground mb-2">{title}</h3>
            <p className="text-sm text-muted-foreground max-w-[240px] mx-auto">{desc}</p>
        </div>
    );
}
