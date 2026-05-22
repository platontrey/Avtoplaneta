/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect } from 'react';
import { Settings, Bell } from 'lucide-react';
import { Button } from './ui/button';
import { Switch } from './ui/switch';
import { Label } from './ui/label';
import { Input } from './ui/input';

interface ChatSettingsProps {
  onSettingsChange?: (settings: ChatSettingsData) => void;
}

export interface ChatSettingsData {
  enableAutoRefresh: boolean;
  refreshInterval: number; // in seconds
  enableNotifications: boolean;
}

const defaultSettings: ChatSettingsData = {
  enableAutoRefresh: true,
  refreshInterval: 30, // 30 seconds for dialogs, 5 for messages
  enableNotifications: true,
};

export default function ChatSettings({ onSettingsChange }: ChatSettingsProps) {
  const [settings, setSettings] = useState<ChatSettingsData>(defaultSettings);

  useEffect(() => {
    // Load settings from localStorage
    const savedSettings = localStorage.getItem('chatSettings');
    if (savedSettings) {
      try {
        const parsed = JSON.parse(savedSettings);
        setSettings({ ...defaultSettings, ...parsed });
      } catch (error) {
        console.error('Failed to parse chat settings:', error);
      }
    }
  }, []);

  const updateSettings = (newSettings: Partial<ChatSettingsData>) => {
    const updated = { ...settings, ...newSettings };
    setSettings(updated);
    localStorage.setItem('chatSettings', JSON.stringify(updated));
    onSettingsChange?.(updated);
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <Settings className="w-8 h-8 text-primary" />
        <h2 className="text-2xl font-bold text-foreground">Настройки чата</h2>
      </div>

      <div className="space-y-6">
        {/* Auto Refresh */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <Label htmlFor="auto-refresh" className="text-base font-medium">
              Автообновление Drom сообщений
            </Label>
            <p className="text-sm text-muted-foreground">
              Автоматически обновлять диалоги и сообщения
            </p>
          </div>
          <Switch
            id="auto-refresh"
            checked={settings.enableAutoRefresh}
            onCheckedChange={(checked) => updateSettings({ enableAutoRefresh: checked })}
          />
        </div>

        {/* Refresh Interval */}
        <div className="space-y-2">
          <Label htmlFor="refresh-interval" className="text-base font-medium">
            Интервал обновления (секунды)
          </Label>
          <p className="text-sm text-muted-foreground">
            Интервал для обновления диалогов (без открытого чата)
          </p>
          <Input
            id="refresh-interval"
            type="number"
            min="10"
            max="300"
            value={settings.refreshInterval}
            onChange={(e) => updateSettings({ refreshInterval: parseInt(e.target.value) || 30 })}
            disabled={!settings.enableAutoRefresh}
            className="w-32"
          />
        </div>

        {/* Notifications */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <Label htmlFor="notifications" className="text-base font-medium flex items-center gap-2">
              <Bell className="w-4 h-4" />
              Уведомления
            </Label>
            <p className="text-sm text-muted-foreground">
              Показывать уведомления о новых сообщениях
            </p>
          </div>
          <Switch
            id="notifications"
            checked={settings.enableNotifications}
            onCheckedChange={(checked) => updateSettings({ enableNotifications: checked })}
          />
        </div>

        {/* Reset to defaults */}
        <div className="pt-4 border-t">
          <Button
            variant="outline"
            onClick={() => {
              setSettings(defaultSettings);
              localStorage.removeItem('chatSettings');
              onSettingsChange?.(defaultSettings);
            }}
          >
            Сбросить настройки
          </Button>
        </div>
      </div>
    </div>
  );
}