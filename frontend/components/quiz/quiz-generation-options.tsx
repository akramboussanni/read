'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Wand2 } from 'lucide-react';

interface QuizGenerationOptionsProps {
  questionCount: number;
  questionTypes: ('mcq' | 'write_word' | 'translate')[];
  directions: ('source_to_target' | 'target_to_source')[];
  onQuestionCountChange: (value: number) => void;
  onQuestionTypesChange: (types: ('mcq' | 'write_word' | 'translate')[]) => void;
  onDirectionsChange: (directions: ('source_to_target' | 'target_to_source')[]) => void;
}

export function QuizGenerationOptions({
  questionCount,
  questionTypes,
  directions,
  onQuestionCountChange,
  onQuestionTypesChange,
  onDirectionsChange,
}: QuizGenerationOptionsProps) {
  const toggleQuestionType = (type: 'mcq' | 'write_word' | 'translate') => {
    if (questionTypes.includes(type)) {
      onQuestionTypesChange(questionTypes.filter((t) => t !== type));
    } else {
      onQuestionTypesChange([...questionTypes, type]);
    }
  };

  const toggleDirection = (dir: 'source_to_target' | 'target_to_source') => {
    if (directions.includes(dir)) {
      onDirectionsChange(directions.filter((d) => d !== dir));
    } else {
      onDirectionsChange([...directions, dir]);
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Wand2 className="w-5 h-5 text-primary" />
          <CardTitle>Quiz Generation</CardTitle>
        </div>
        <CardDescription>Configure how questions should be generated</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="questionCount">Number of Questions</Label>
          <Input
            id="questionCount"
            type="number"
            min="1"
            max="100"
            value={questionCount}
            onChange={(e) => onQuestionCountChange(Number(e.target.value))}
            className="mt-1"
          />
          <p className="text-xs text-muted-foreground mt-1">
            How many questions to generate (1-100)
          </p>
        </div>

        <div className="space-y-2">
          <Label>Question Types</Label>
          <div className="space-y-2">
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={questionTypes.includes('mcq')}
                onChange={() => toggleQuestionType('mcq')}
                className="w-4 h-4 rounded border-gray-300 text-primary"
              />
              <span className="text-sm">Multiple Choice (MCQ)</span>
            </label>
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={questionTypes.includes('write_word')}
                onChange={() => toggleQuestionType('write_word')}
                className="w-4 h-4 rounded border-gray-300 text-primary"
              />
              <span className="text-sm">Write Word</span>
            </label>
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={questionTypes.includes('translate')}
                onChange={() => toggleQuestionType('translate')}
                className="w-4 h-4 rounded border-gray-300 text-primary"
              />
              <span className="text-sm">Translation</span>
            </label>
          </div>
        </div>

        <div className="space-y-2">
          <Label>Question Direction</Label>
          <div className="space-y-2">
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={directions.includes('source_to_target')}
                onChange={() => toggleDirection('source_to_target')}
                className="w-4 h-4 rounded border-gray-300 text-primary"
              />
              <span className="text-sm">Source → Target (Arabic → French)</span>
            </label>
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={directions.includes('target_to_source')}
                onChange={() => toggleDirection('target_to_source')}
                className="w-4 h-4 rounded border-gray-300 text-primary"
              />
              <span className="text-sm">Target → Source (French → Arabic)</span>
            </label>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
