'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Shield, Coins, CheckCircle } from 'lucide-react';

interface AdminQuizSettingsProps {
  passPercentage: number;
  givesCoins: boolean;
  coinReward: number;
  levelOrder: number;
  isSystem: boolean;
  onPassPercentageChange: (value: number) => void;
  onGivesCoinsChange: (value: boolean) => void;
  onCoinRewardChange: (value: number) => void;
  onLevelOrderChange: (value: number) => void;
  onIsSystemChange: (value: boolean) => void;
}

export function AdminQuizSettings({
  passPercentage,
  givesCoins,
  coinReward,
  levelOrder,
  isSystem,
  onPassPercentageChange,
  onGivesCoinsChange,
  onCoinRewardChange,
  onLevelOrderChange,
  onIsSystemChange,
}: AdminQuizSettingsProps) {
  return (
    <Card className="border-2 border-accent/50 bg-accent/5">
      <CardHeader>
        <div className="flex items-center gap-2">
          <Shield className="w-5 h-5 text-accent" />
          <CardTitle>Admin Settings</CardTitle>
        </div>
        <CardDescription>Configure advanced quiz properties (admin only)</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Label htmlFor="passPercentage" className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4" />
              Pass Percentage
            </Label>
            <Input
              id="passPercentage"
              type="number"
              min="0"
              max="100"
              value={passPercentage}
              onChange={(e) => onPassPercentageChange(Number(e.target.value))}
              className="mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Minimum score to pass (0-100%)
            </p>
          </div>

          <div>
            <Label htmlFor="levelOrder" className="flex items-center gap-2">
              Level Order
            </Label>
            <Input
              id="levelOrder"
              type="number"
              min="0"
              value={levelOrder}
              onChange={(e) => onLevelOrderChange(Number(e.target.value))}
              className="mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Quiz progression order (0 = first)
            </p>
          </div>
        </div>

        <div className="border-t pt-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="givesCoins"
                checked={givesCoins}
                onChange={(e) => onGivesCoinsChange(e.target.checked)}
                className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <Label htmlFor="givesCoins" className="font-normal cursor-pointer flex items-center gap-2">
                <Coins className="w-4 h-4 text-yellow-500" />
                Reward coins upon completion
              </Label>
            </div>
          </div>

          {givesCoins && (
            <div className="ml-6 mb-4">
              <Label htmlFor="coinReward" className="text-sm">Coins per Question</Label>
              <Input
                id="coinReward"
                type="number"
                min="0"
                value={coinReward}
                onChange={(e) => onCoinRewardChange(Number(e.target.value))}
                className="mt-1 max-w-xs"
                placeholder="e.g., 10"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Total reward = coins × number of questions
              </p>
            </div>
          )}
        </div>

        <div className="flex items-center space-x-2 border-t pt-4">
          <input
            type="checkbox"
            id="isSystem"
            checked={isSystem}
            onChange={(e) => onIsSystemChange(e.target.checked)}
            className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
          />
          <Label htmlFor="isSystem" className="font-normal cursor-pointer">
            Mark as system quiz (official content)
          </Label>
        </div>
      </CardContent>
    </Card>
  );
}
