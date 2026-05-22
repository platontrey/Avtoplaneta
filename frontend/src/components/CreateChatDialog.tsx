/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { Conversation, User } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';
import SelectUserDropdown from './SelectUserDropdown';

interface CreateChatDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (conversation: Conversation) => void;
  currentUser: User;
}

export default function CreateChatDialog({ isOpen, onClose, onSuccess, currentUser }: CreateChatDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [form, setForm] = useState({
    title: '',
    participants: [] as User[],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const participantIds = form.participants.map(user => user.id);

      if (participantIds.length === 0) {
        alert('Выберите хотя бы одного участника');
        setIsLoading(false);
        return;
      }

      const conversation = await messagingApi.createConversation(
        participantIds,
        form.title.trim() || undefined
      );

      onSuccess(conversation);
      onClose();
      setForm({ title: '', participants: [] });
    } catch (error) {
      console.error('Failed to create conversation:', error);
      alert('Ошибка при создании чата');
    } finally {
      setIsLoading(false);
    }
  };

  const updateForm = (field: string, value: string | User[]) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Создать новый чат</DialogTitle>
          <DialogDescription>
            Введите название чата и ID участников через запятую
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="title">Название чата (опционально)</Label>
            <Input
              id="title"
              placeholder="Введите название чата"
              value={form.title}
              onChange={(e) => updateForm('title', e.target.value)}
            />
          </div>

          <SelectUserDropdown
            selectedUsers={form.participants}
            onSelectionChange={(users) => updateForm('participants', users)}
            placeholder="Выберите участников чата"
            multiple={true}
            currentUser={currentUser}
          />

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Создание...' : 'Создать чат'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}