'use client';

import { useEffect, useState } from 'react';
import { quizApi } from '@/lib/api/quiz';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { BookOpen, Check, ChevronsUpDown, Filter, Loader2 } from 'lucide-react';
import type { Deck, Category } from '@/lib/types/quiz';
import type { DeckSelectionRequest } from '@/lib/types/quiz';
import { cn } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';

interface VisualDeckSelectorProps {
    deckSelections: DeckSelectionRequest[];
    onDeckSelectionsChange: (selections: DeckSelectionRequest[]) => void;
}

export function VisualDeckSelector({
    deckSelections,
    onDeckSelectionsChange,
}: VisualDeckSelectorProps) {
    const [decks, setDecks] = useState<Deck[]>([]);
    const [deckCategories, setDeckCategories] = useState<Record<string, Category[]>>({});
    const [loading, setLoading] = useState(true);
    const [expandedDeckId, setExpandedDeckId] = useState<string | null>(null);
    const [loadingCategories, setLoadingCategories] = useState<string | null>(null);

    useEffect(() => {
        loadDecks();
    }, []);

    const loadDecks = async () => {
        try {
            const data = await quizApi.listDecks();
            setDecks(data);
        } catch (err) {
            console.error('Failed to load decks:', err);
        } finally {
            setLoading(false);
        }
    };

    const loadCategories = async (deckId: string) => {
        if (deckCategories[deckId]) return; // Already loaded

        setLoadingCategories(deckId);
        try {
            const data = await quizApi.getCategories(deckId);
            setDeckCategories((prev) => ({ ...prev, [deckId]: data }));
        } catch (err) {
            console.error('Failed to load categories:', err);
        } finally {
            setLoadingCategories(null);
        }
    };

    const handleDeckClick = async (deckId: string) => {
        if (expandedDeckId === deckId) {
            setExpandedDeckId(null);
            return;
        }

        setExpandedDeckId(deckId);
        await loadCategories(deckId);
    };

    const isDeckSelected = (deckId: string) => {
        return deckSelections.some((s) => s.deck_id === deckId);
    };

    const getSelectionForDeck = (deckId: string) => {
        return deckSelections.find((s) => s.deck_id === deckId);
    };

    const toggleDeck = (deckId: string) => {
        if (isDeckSelected(deckId)) {
            onDeckSelectionsChange(deckSelections.filter((s) => s.deck_id !== deckId));
        } else {
            // Select deck with NO specific categories (meaning ALL)
            onDeckSelectionsChange([
                ...deckSelections,
                { deck_id: deckId, categories: [] },
            ]);
        }
    };

    const toggleCategory = (deckId: string, categoryKey: string) => {
        const currentSelection = getSelectionForDeck(deckId);

        if (!currentSelection) {
            // If deck not selected, select it with just this category
            onDeckSelectionsChange([
                ...deckSelections,
                { deck_id: deckId, categories: [categoryKey] },
            ]);
            return;
        }

        let newCategories = [...currentSelection.categories];

        if (newCategories.length === 0) {
            // Currently "All" are selected. 
            // If we toggle one, we need to explicitly list all OTHERS, or logic implies:
            // Empty list = All.
            // So if user clicks one category to "toggle it", does it mean "Select ONLY this one" or "Deselect this one from All"?
            // Usually "Select Only This One" is the intuition if starting from "Whole Deck".
            // BUT, let's look at the requirement: "pick a category from a deck, not everything".

            // If currently "All" (empty list), and user clicks a category, maybe they want to restrict to just that category?
            // Or maybe they want to exclude it. 
            // Let's assume: clicking a category means "I want this category included".
            // If the deck was "All", and we click one, we switch to "Just this one".
            newCategories = [categoryKey];
        } else {
            // Already has a specific list
            if (newCategories.includes(categoryKey)) {
                newCategories = newCategories.filter((c) => c !== categoryKey);
            } else {
                newCategories.push(categoryKey);
            }
        }

        // If no categories left, should we deselect the deck? Or switch back to "All"? 
        // Let's say if no categories selected manually, we remove the deck selection entirely?
        // Or we keep it as "All"? 
        // Let's go with: explicit selection. If list becomes empty, we remove the deck selection (deselect deck).
        // Wait, earlier logic was: empty list = ALL. 
        // So we need to be careful.

        // Improved Logic:
        // If deck is selected as "All" (empty categories):
        //   User clicks 'Cat A' -> Becomes selected with just 'Cat A'.
        // If deck is selected with specific categories ['Cat A']:
        //   User clicks 'Cat A' -> Becomes empty -> Deselect deck? Yes.
        //   User clicks 'Cat B' -> Becomes ['Cat A', 'Cat B'].

        if (newCategories.length === 0) {
            onDeckSelectionsChange(deckSelections.filter((s) => s.deck_id !== deckId));
        } else {
            const updatedSelections = deckSelections.map((s) =>
                s.deck_id === deckId ? { ...s, categories: newCategories } : s
            );
            onDeckSelectionsChange(updatedSelections);
        }
    };

    const selectAllCategories = (deckId: string) => {
        // Set to empty list which means ALL
        const currentSelection = getSelectionForDeck(deckId);
        if (currentSelection) {
            const updatedSelections = deckSelections.map((s) =>
                s.deck_id === deckId ? { ...s, categories: [] } : s
            );
            onDeckSelectionsChange(updatedSelections);
        } else {
            onDeckSelectionsChange([
                ...deckSelections,
                { deck_id: deckId, categories: [] },
            ]);
        }
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center p-12">
                <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {decks.map((deck) => {
                    const isSelected = isDeckSelected(deck.id);
                    const selection = getSelectionForDeck(deck.id);
                    const selectedCount = selection?.categories?.length || 0;
                    const isAllSelected = isSelected && selectedCount === 0;

                    return (
                        <motion.div
                            layout
                            key={deck.id}
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ duration: 0.3 }}
                            className={cn(
                                "group relative rounded-xl border-2 transition-all duration-200 overflow-hidden bg-card hover:shadow-md cursor-pointer",
                                isSelected ? "border-primary ring-2 ring-primary/20" : "border-muted hover:border-primary/50"
                            )}
                        >
                            {/* Card Header / Main Click Area */}
                            <div
                                className="p-4"
                                onClick={() => handleDeckClick(deck.id)}
                            >
                                <div className="flex justify-between items-start mb-2">
                                    <div className="p-2 bg-primary/10 rounded-lg text-primary">
                                        <BookOpen className="w-5 h-5" />
                                    </div>
                                    {isSelected && (
                                        <div className="bg-primary text-primary-foreground text-xs font-bold px-2 py-1 rounded-full flex items-center gap-1">
                                            <Check className="w-3 h-3" />
                                            {isAllSelected ? 'All' : `${selectedCount}`}
                                        </div>
                                    )}
                                </div>

                                <h3 className="font-bold text-lg mb-1 line-clamp-1">{deck.title}</h3>
                                <p className="text-sm text-muted-foreground mb-3">
                                    {deck.question_count || 0} questions • {deck.category_count || 0} categories
                                </p>

                                <div className="flex items-center text-xs text-primary font-medium group-hover:underline">
                                    {expandedDeckId === deck.id ? 'Collapse' : 'Select Categories'}
                                    <ChevronsUpDown className="w-3 h-3 ml-1" />
                                </div>
                            </div>

                            {/* Collapsible Content for Categories */}
                            <AnimatePresence>
                                {expandedDeckId === deck.id && (
                                    <motion.div
                                        initial={{ height: 0, opacity: 0 }}
                                        animate={{ height: "auto", opacity: 1 }}
                                        exit={{ height: 0, opacity: 0 }}
                                        className="border-t bg-muted/30"
                                    >
                                        <div className="p-4 space-y-3">
                                            <div className="flex items-center justify-between">
                                                <Label className="text-xs font-semibold uppercase text-muted-foreground">Categories</Label>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-6 text-xs"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        selectAllCategories(deck.id);
                                                    }}
                                                >
                                                    Select All
                                                </Button>
                                            </div>

                                            {loadingCategories === deck.id ? (
                                                <div className="flex justify-center py-4">
                                                    <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
                                                </div>
                                            ) : (
                                                <div className="flex flex-wrap gap-2">
                                                    {deckCategories[deck.id]?.map((cat) => {
                                                        const isCatSelected = isAllSelected || (selection?.categories?.includes(cat.category_key));

                                                        return (
                                                            <button
                                                                key={cat.id}
                                                                type="button"
                                                                onClick={(e) => {
                                                                    e.stopPropagation();
                                                                    toggleCategory(deck.id, cat.category_key);
                                                                }}
                                                                className={cn(
                                                                    "text-xs px-2.5 py-1 rounded-full border transition-colors",
                                                                    isCatSelected
                                                                        ? "bg-primary text-primary-foreground border-primary"
                                                                        : "bg-background text-foreground border-input hover:border-primary hover:text-primary"
                                                                )}
                                                            >
                                                                {cat.title}
                                                            </button>
                                                        )
                                                    })}
                                                    {deckCategories[deck.id]?.length === 0 && (
                                                        <p className="text-xs text-muted-foreground italic">No categories found in this deck.</p>
                                                    )}
                                                </div>
                                            )}
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </motion.div>
                    );
                })}
            </div>

            {decks.length === 0 && !loading && (
                <div className="text-center py-12 border-2 border-dashed rounded-xl">
                    <Filter className="w-10 h-10 mx-auto text-muted-foreground mb-3" />
                    <h3 className="text-lg font-medium">No Decks Found</h3>
                    <p className="text-muted-foreground">Get started by creating your first deck.</p>
                </div>
            )}
        </div>
    );
}
