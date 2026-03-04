'use client';

import { useState, useCallback, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { quizApi } from '@/lib/api/quiz';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ManualQuestionBuilder } from '@/components/quiz/manual-question-builder';
import { VisualDeckSelector } from '@/components/quiz/visual-deck-selector';
import { QuizGenerationOptions } from '@/components/quiz/quiz-generation-options';
import { AdminQuizSettings } from '@/components/quiz/admin-quiz-settings';
import { useAuthStore } from '@/lib/store/auth-store';
import type { ManualQuestionRequest, DeckSelectionRequest, Quiz, CreateQuizRequest } from '@/lib/types/quiz';
import {
    BookOpen, Wand2, ArrowLeft, ArrowRight, Save, Loader2,
    CheckCircle, Settings, Sparkles, PenTool,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';

// =============================================
// TYPES
// =============================================

type QuizMode = 'manual' | 'auto';
type EditorStep = 'mode' | 'questions' | 'settings' | 'review';

interface QuizEditorProps {
    /** If provided, we're editing an existing quiz */
    existingQuiz?: Quiz & { templates?: any[] };
    /** If true, show in a compact inline mode (for course tree embedding) */
    inline?: boolean;
    /** Called when quiz is saved (instead of redirecting) */
    onSave?: (quiz: Quiz) => void;
    /** Called when cancel is pressed in inline mode */
    onCancel?: () => void;
}

// =============================================
// MAIN COMPONENT
// =============================================

export function QuizEditor({ existingQuiz, inline, onSave, onCancel }: QuizEditorProps) {
    const router = useRouter();
    const { user } = useAuthStore();
    const isAdmin = user?.is_admin ?? false;

    // State
    const [step, setStep] = useState<EditorStep>(existingQuiz ? 'questions' : 'mode');
    const [mode, setMode] = useState<QuizMode>(existingQuiz?.question_mode === 'auto' ? 'auto' : 'manual');

    // Basic info
    const [title, setTitle] = useState(existingQuiz?.title || '');
    const [description, setDescription] = useState(existingQuiz?.description || '');
    const [isPublic, setIsPublic] = useState(existingQuiz?.is_public || false);

    // Manual questions
    const [manualQuestions, setManualQuestions] = useState<ManualQuestionRequest[]>([]);

    // Auto-generate config
    const [deckSelections, setDeckSelections] = useState<DeckSelectionRequest[]>([]);
    const [questionCount, setQuestionCount] = useState(10);
    const [questionTypes, setQuestionTypes] = useState<('mcq' | 'write_word' | 'translate')[]>(['mcq', 'translate']);
    const [directions, setDirections] = useState<('source_to_target' | 'target_to_source')[]>(['source_to_target', 'target_to_source']);

    // Admin settings
    const [passPercentage, setPassPercentage] = useState(existingQuiz?.pass_percentage ?? 70);
    const [givesCoins, setGivesCoins] = useState(existingQuiz?.gives_coins ?? false);
    const [coinReward, setCoinReward] = useState(existingQuiz?.coin_reward ?? 10);
    const [levelOrder, setLevelOrder] = useState(existingQuiz?.level_order ?? 0);
    const [isSystem, setIsSystem] = useState(false);
    const [shuffleQuestions, setShuffleQuestions] = useState(existingQuiz?.shuffle_questions ?? true);

    // UI state
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Load existing quiz data into state
    useEffect(() => {
        if (!existingQuiz?.templates?.length) return;

        const templates = existingQuiz.templates;
        const firstTemplate = templates[0];

        if (firstTemplate.generation_mode === 'manual' && firstTemplate.manual_data) {
            try {
                const questions = JSON.parse(firstTemplate.manual_data);
                setManualQuestions(questions);
                setMode('manual');
            } catch { }
        } else if (firstTemplate.generation_mode === 'random_from_deck') {
            setMode('auto');
            // Reconstruct deck selections from templates
            const selections: DeckSelectionRequest[] = [];
            for (const tmpl of templates) {
                if (!tmpl.deck_id) continue;
                const existing = selections.find(s => s.deck_id === String(tmpl.deck_id));
                if (existing) {
                    // Possibly add category
                } else {
                    selections.push({ deck_id: String(tmpl.deck_id), categories: [] });
                }
            }
            setDeckSelections(selections);

            // Parse types/directions from first template
            try {
                setQuestionTypes(JSON.parse(firstTemplate.question_types || '["mcq","translate"]'));
            } catch { }
            try {
                setDirections(JSON.parse(firstTemplate.directions || '["source_to_target","target_to_source"]'));
            } catch { }
            setQuestionCount(firstTemplate.question_count || 10);
        }
    }, [existingQuiz]);

    // Steps config
    const steps: EditorStep[] = existingQuiz ? ['questions', 'settings', 'review'] : ['mode', 'questions', 'settings', 'review'];
    const currentStepIdx = steps.indexOf(step);

    const canAdvance = () => {
        switch (step) {
            case 'mode':
                return true;
            case 'questions':
                if (mode === 'manual') return manualQuestions.length > 0;
                if (mode === 'auto') return deckSelections.length > 0 && questionCount > 0;
                return false;
            case 'settings':
                return title.trim().length > 0;
            case 'review':
                return true;
            default:
                return false;
        }
    };

    const nextStep = () => {
        const idx = steps.indexOf(step);
        if (idx < steps.length - 1) {
            setStep(steps[idx + 1]);
        }
    };

    const prevStep = () => {
        const idx = steps.indexOf(step);
        if (idx > 0) {
            setStep(steps[idx - 1]);
        }
    };

    const handleSave = async () => {
        if (!title.trim()) { setError('Un titre est requis'); return; }
        setSaving(true);
        setError(null);

        try {
            const payload: any = {
                title,
                description,
                is_public: isPublic,
                pass_percentage: passPercentage,
                gives_coins: givesCoins,
                coin_reward: coinReward,
                shuffle_questions: shuffleQuestions,
                is_dynamic: mode === 'auto',
                question_mode: mode,
            };

            if (mode === 'manual') {
                payload.manual_questions = manualQuestions;
            } else {
                payload.auto_generate = {
                    deck_selections: deckSelections,
                    question_types: questionTypes,
                    directions: directions,
                    question_count: questionCount,
                };
            }

            let quiz;
            if (existingQuiz) {
                quiz = await quizApi.updateQuiz(existingQuiz.id, payload);
            } else {
                quiz = await quizApi.createQuiz(payload as CreateQuizRequest);
            }

            if (onSave) {
                onSave(quiz);
            } else {
                router.push(`/quizzes/my`);
            }
        } catch (err: any) {
            setError(err.response?.data?.message || err.message || 'Échec de la sauvegarde');
        } finally {
            setSaving(false);
        }
    };

    // =============================================
    // RENDER
    // =============================================

    return (
        <div className={cn("space-y-6", !inline && "max-w-5xl mx-auto")}>

            {/* Step indicator */}
            <div className="flex items-center justify-center gap-2">
                {steps.map((s, i) => {
                    const isActive = s === step;
                    const isDone = steps.indexOf(step) > i;
                    const labels: Record<EditorStep, string> = {
                        mode: 'Mode',
                        questions: 'Questions',
                        settings: 'Paramètres',
                        review: 'Vérification',
                    };
                    const icons: Record<EditorStep, React.ReactNode> = {
                        mode: <Sparkles className="w-4 h-4" />,
                        questions: <PenTool className="w-4 h-4" />,
                        settings: <Settings className="w-4 h-4" />,
                        review: <CheckCircle className="w-4 h-4" />,
                    };

                    return (
                        <div key={s} className="flex items-center gap-2">
                            <button
                                onClick={() => {
                                    // Allow jumping to any completed step
                                    if (isDone || isActive) setStep(s);
                                }}
                                className={cn(
                                    "flex items-center gap-2 px-4 py-2 rounded-full text-sm font-semibold transition-all",
                                    isActive && "bg-primary text-white shadow-lg shadow-primary/20",
                                    isDone && "bg-primary/10 text-primary cursor-pointer hover:bg-primary/20",
                                    !isActive && !isDone && "bg-muted text-muted-foreground"
                                )}
                            >
                                {isDone ? <CheckCircle className="w-4 h-4" /> : icons[s]}
                                <span className="hidden sm:inline">{labels[s]}</span>
                            </button>
                            {i < steps.length - 1 && (
                                <div className={cn("w-8 h-0.5 rounded", isDone ? "bg-primary" : "bg-border")} />
                            )}
                        </div>
                    );
                })}
            </div>

            {/* Error */}
            {error && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-xl text-sm">
                    {error}
                </div>
            )}

            {/* Step Content */}
            <AnimatePresence mode="wait">
                <motion.div
                    key={step}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    transition={{ duration: 0.2 }}
                >
                    {step === 'mode' && (
                        <StepMode mode={mode} onModeChange={setMode} />
                    )}

                    {step === 'questions' && mode === 'manual' && (
                        <ManualQuestionBuilder
                            questions={manualQuestions}
                            onChange={setManualQuestions}
                        />
                    )}

                    {step === 'questions' && mode === 'auto' && (
                        <StepAutoConfig
                            deckSelections={deckSelections}
                            onDeckSelectionsChange={setDeckSelections}
                            questionCount={questionCount}
                            onQuestionCountChange={setQuestionCount}
                            questionTypes={questionTypes}
                            onQuestionTypesChange={setQuestionTypes}
                            directions={directions}
                            onDirectionsChange={setDirections}
                        />
                    )}

                    {step === 'settings' && (
                        <StepSettings
                            title={title}
                            onTitleChange={setTitle}
                            description={description}
                            onDescriptionChange={setDescription}
                            isPublic={isPublic}
                            onIsPublicChange={setIsPublic}
                            shuffleQuestions={shuffleQuestions}
                            onShuffleQuestionsChange={setShuffleQuestions}
                            isAdmin={isAdmin}
                            passPercentage={passPercentage}
                            onPassPercentageChange={setPassPercentage}
                            givesCoins={givesCoins}
                            onGivesCoinsChange={setGivesCoins}
                            coinReward={coinReward}
                            onCoinRewardChange={setCoinReward}
                            levelOrder={levelOrder}
                            onLevelOrderChange={setLevelOrder}
                            isSystem={isSystem}
                            onIsSystemChange={setIsSystem}
                        />
                    )}

                    {step === 'review' && (
                        <StepReview
                            mode={mode}
                            title={title}
                            description={description}
                            isPublic={isPublic}
                            manualQuestions={manualQuestions}
                            deckSelections={deckSelections}
                            questionCount={questionCount}
                            questionTypes={questionTypes}
                            directions={directions}
                            passPercentage={passPercentage}
                            givesCoins={givesCoins}
                            coinReward={coinReward}
                        />
                    )}
                </motion.div>
            </AnimatePresence>

            {/* Navigation */}
            <div className="flex items-center justify-between pt-4 border-t">
                <div>
                    {currentStepIdx > 0 ? (
                        <Button variant="outline" onClick={prevStep} className="gap-2">
                            <ArrowLeft className="w-4 h-4" />
                            Précédent
                        </Button>
                    ) : (
                        onCancel ? (
                            <Button variant="outline" onClick={onCancel}>Annuler</Button>
                        ) : !inline ? (
                            <Button variant="outline" onClick={() => router.push('/quizzes')} className="gap-2">
                                <ArrowLeft className="w-4 h-4" />
                                Quiz
                            </Button>
                        ) : null
                    )}
                </div>

                <div className="flex items-center gap-3">
                    {step === 'review' ? (
                        <Button
                            onClick={handleSave}
                            disabled={saving || !title.trim()}
                            className="gap-2 bg-emerald-600 hover:bg-emerald-700 text-white px-8"
                        >
                            {saving ? (
                                <><Loader2 className="w-4 h-4 animate-spin" /> Sauvegarde...</>
                            ) : (
                                <><Save className="w-4 h-4" /> {existingQuiz ? 'Mettre à jour' : 'Créer le Quiz'}</>
                            )}
                        </Button>
                    ) : (
                        <Button
                            onClick={nextStep}
                            disabled={!canAdvance()}
                            className="gap-2 px-8"
                        >
                            Suivant
                            <ArrowRight className="w-4 h-4" />
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
}

// =============================================
// STEP: MODE SELECTION
// =============================================

function StepMode({ mode, onModeChange }: { mode: QuizMode; onModeChange: (m: QuizMode) => void }) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-3xl mx-auto">
            <ModeCard
                icon={<PenTool className="w-8 h-8" />}
                title="Mode Manuel"
                description="Créez vos propres questions à la main, ou auto-générez et révisez-les avant de les ajouter."
                features={['Créer des questions MCQ / traduction', 'Auto-générer et éditer', 'Contrôle total sur le contenu']}
                selected={mode === 'manual'}
                onClick={() => onModeChange('manual')}
                gradient="from-blue-500 to-indigo-500"
            />
            <ModeCard
                icon={<Wand2 className="w-8 h-8" />}
                title="Mode Automatique"
                description="Choisissez un deck et des catégories, le quiz génère des questions aléatoires à chaque tentative."
                features={['Questions aléatoires à chaque fois', 'Configuration simple', 'Idéal pour la révision']}
                selected={mode === 'auto'}
                onClick={() => onModeChange('auto')}
                gradient="from-amber-500 to-orange-500"
            />
        </div>
    );
}

function ModeCard({ icon, title, description, features, selected, onClick, gradient }: {
    icon: React.ReactNode;
    title: string;
    description: string;
    features: string[];
    selected: boolean;
    onClick: () => void;
    gradient: string;
}) {
    return (
        <motion.div
            whileHover={{ y: -4 }}
            whileTap={{ scale: 0.98 }}
            onClick={onClick}
            className={cn(
                "relative cursor-pointer rounded-2xl border-2 p-6 transition-all duration-200",
                selected
                    ? "border-primary bg-primary/5 shadow-xl shadow-primary/10 ring-2 ring-primary/20"
                    : "border-border bg-card hover:border-primary/30 hover:shadow-lg"
            )}
        >
            <div className={cn(
                "w-16 h-16 rounded-2xl flex items-center justify-center text-white mb-4 bg-gradient-to-br",
                gradient
            )}>
                {icon}
            </div>
            <h3 className="text-xl font-bold mb-2">{title}</h3>
            <p className="text-muted-foreground text-sm mb-4">{description}</p>
            <ul className="space-y-2">
                {features.map((f, i) => (
                    <li key={i} className="flex items-center gap-2 text-sm">
                        <CheckCircle className={cn("w-4 h-4 flex-shrink-0", selected ? "text-primary" : "text-muted-foreground")} />
                        <span>{f}</span>
                    </li>
                ))}
            </ul>
            {selected && (
                <div className="absolute top-4 right-4">
                    <div className="w-6 h-6 bg-primary rounded-full flex items-center justify-center">
                        <CheckCircle className="w-4 h-4 text-white" />
                    </div>
                </div>
            )}
        </motion.div>
    );
}

// =============================================
// STEP: AUTO CONFIG
// =============================================

function StepAutoConfig({
    deckSelections, onDeckSelectionsChange,
    questionCount, onQuestionCountChange,
    questionTypes, onQuestionTypesChange,
    directions, onDirectionsChange,
}: {
    deckSelections: DeckSelectionRequest[];
    onDeckSelectionsChange: (s: DeckSelectionRequest[]) => void;
    questionCount: number;
    onQuestionCountChange: (n: number) => void;
    questionTypes: ('mcq' | 'write_word' | 'translate')[];
    onQuestionTypesChange: (t: ('mcq' | 'write_word' | 'translate')[]) => void;
    directions: ('source_to_target' | 'target_to_source')[];
    onDirectionsChange: (d: ('source_to_target' | 'target_to_source')[]) => void;
}) {
    return (
        <div className="space-y-6">
            <VisualDeckSelector
                deckSelections={deckSelections}
                onDeckSelectionsChange={onDeckSelectionsChange}
            />
            <QuizGenerationOptions
                questionCount={questionCount}
                questionTypes={questionTypes}
                directions={directions}
                onQuestionCountChange={onQuestionCountChange}
                onQuestionTypesChange={onQuestionTypesChange}
                onDirectionsChange={onDirectionsChange}
            />
        </div>
    );
}

// =============================================
// STEP: SETTINGS
// =============================================

function StepSettings({
    title, onTitleChange,
    description, onDescriptionChange,
    isPublic, onIsPublicChange,
    shuffleQuestions, onShuffleQuestionsChange,
    isAdmin,
    passPercentage, onPassPercentageChange,
    givesCoins, onGivesCoinsChange,
    coinReward, onCoinRewardChange,
    levelOrder, onLevelOrderChange,
    isSystem, onIsSystemChange,
}: {
    title: string; onTitleChange: (v: string) => void;
    description: string; onDescriptionChange: (v: string) => void;
    isPublic: boolean; onIsPublicChange: (v: boolean) => void;
    shuffleQuestions: boolean; onShuffleQuestionsChange: (v: boolean) => void;
    isAdmin: boolean;
    passPercentage: number; onPassPercentageChange: (v: number) => void;
    givesCoins: boolean; onGivesCoinsChange: (v: boolean) => void;
    coinReward: number; onCoinRewardChange: (v: number) => void;
    levelOrder: number; onLevelOrderChange: (v: number) => void;
    isSystem: boolean; onIsSystemChange: (v: boolean) => void;
}) {
    return (
        <div className="space-y-6 max-w-2xl mx-auto">
            <Card>
                <CardHeader>
                    <CardTitle>Informations du Quiz</CardTitle>
                    <CardDescription>Les détails de base de votre quiz</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div>
                        <Label htmlFor="title">Titre *</Label>
                        <Input
                            id="title"
                            value={title}
                            onChange={(e) => onTitleChange(e.target.value)}
                            placeholder="Mon Super Quiz"
                            required
                            className="mt-1"
                        />
                    </div>
                    <div>
                        <Label htmlFor="description">Description</Label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(e) => onDescriptionChange(e.target.value)}
                            placeholder="De quoi parle ce quiz ?"
                            className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background min-h-[80px] resize-y"
                        />
                    </div>
                    <div className="flex flex-col gap-3 pt-2">
                        <label className="flex items-center space-x-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={isPublic}
                                onChange={(e) => onIsPublicChange(e.target.checked)}
                                className="w-4 h-4 rounded border-gray-300 text-primary"
                            />
                            <span className="text-sm">Rendre ce quiz public</span>
                        </label>
                        <label className="flex items-center space-x-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={shuffleQuestions}
                                onChange={(e) => onShuffleQuestionsChange(e.target.checked)}
                                className="w-4 h-4 rounded border-gray-300 text-primary"
                            />
                            <span className="text-sm">Mélanger les questions</span>
                        </label>
                    </div>
                </CardContent>
            </Card>

            {isAdmin && (
                <AdminQuizSettings
                    passPercentage={passPercentage}
                    givesCoins={givesCoins}
                    coinReward={coinReward}
                    levelOrder={levelOrder}
                    isSystem={isSystem}
                    onPassPercentageChange={onPassPercentageChange}
                    onGivesCoinsChange={onGivesCoinsChange}
                    onCoinRewardChange={onCoinRewardChange}
                    onLevelOrderChange={onLevelOrderChange}
                    onIsSystemChange={onIsSystemChange}
                />
            )}
        </div>
    );
}

// =============================================
// STEP: REVIEW
// =============================================

function StepReview({
    mode, title, description, isPublic,
    manualQuestions, deckSelections,
    questionCount, questionTypes, directions,
    passPercentage, givesCoins, coinReward,
}: {
    mode: QuizMode;
    title: string;
    description: string;
    isPublic: boolean;
    manualQuestions: ManualQuestionRequest[];
    deckSelections: DeckSelectionRequest[];
    questionCount: number;
    questionTypes: string[];
    directions: string[];
    passPercentage: number;
    givesCoins: boolean;
    coinReward: number;
}) {
    return (
        <div className="max-w-2xl mx-auto space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <CheckCircle className="w-5 h-5 text-emerald-500" />
                        Résumé du Quiz
                    </CardTitle>
                    <CardDescription>Vérifiez les détails avant de créer</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-muted-foreground">Titre</span>
                            <p className="font-semibold">{title || '—'}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Mode</span>
                            <p className="font-semibold flex items-center gap-1">
                                {mode === 'manual' ? <><PenTool className="w-3 h-3" /> Manuel</> : <><Wand2 className="w-3 h-3" /> Automatique</>}
                            </p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Visibilité</span>
                            <p className="font-semibold">{isPublic ? 'Public' : 'Privé'}</p>
                        </div>
                        <div>
                            <span className="text-muted-foreground">Questions</span>
                            <p className="font-semibold">
                                {mode === 'manual' ? `${manualQuestions.length} questions` : `${questionCount} (auto)`}
                            </p>
                        </div>
                        {description && (
                            <div className="col-span-2">
                                <span className="text-muted-foreground">Description</span>
                                <p className="font-semibold">{description}</p>
                            </div>
                        )}
                    </div>

                    <div className="border-t pt-4 space-y-2 text-sm">
                        <div className="flex justify-between">
                            <span className="text-muted-foreground">Pourcentage de réussite</span>
                            <span className="font-semibold">{passPercentage}%</span>
                        </div>
                        {givesCoins && (
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Récompense pièces</span>
                                <span className="font-semibold">{coinReward} pièces</span>
                            </div>
                        )}
                        {mode === 'auto' && (
                            <>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Types de questions</span>
                                    <span className="font-semibold">{questionTypes.join(', ')}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Directions</span>
                                    <span className="font-semibold">{directions.map(d => d === 'source_to_target' ? 'AR → FR' : 'FR → AR').join(', ')}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Decks sélectionnés</span>
                                    <span className="font-semibold">{deckSelections.length}</span>
                                </div>
                            </>
                        )}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
