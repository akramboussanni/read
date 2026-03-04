'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface QuizBasicInfoProps {
  title: string;
  description: string;
  isPublic: boolean;
  onTitleChange: (value: string) => void;
  onDescriptionChange: (value: string) => void;
  onIsPublicChange: (value: boolean) => void;
}

export function QuizBasicInfo({
  title,
  description,
  isPublic,
  onTitleChange,
  onDescriptionChange,
  onIsPublicChange,
}: QuizBasicInfoProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Basic Information</CardTitle>
        <CardDescription>Enter the basic details for your quiz</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="title">Quiz Title *</Label>
          <Input
            id="title"
            value={title}
            onChange={(e) => onTitleChange(e.target.value)}
            placeholder="My Awesome Quiz"
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
            placeholder="What is this quiz about?"
            className="w-full mt-1 px-3 py-2 border border-input rounded-md bg-background min-h-[80px] resize-y"
          />
        </div>

        <div className="flex items-center space-x-2">
          <input
            type="checkbox"
            id="isPublic"
            checked={isPublic}
            onChange={(e) => onIsPublicChange(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
          />
          <Label htmlFor="isPublic" className="font-normal cursor-pointer">
            Make this quiz public (others can see and take it)
          </Label>
        </div>
      </CardContent>
    </Card>
  );
}
