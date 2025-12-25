/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { Edit, Trash2, X } from 'lucide-react';
import type { Conversation } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from './ui/alert-dialog';

interface IndividualChatSettingsProps {
  conversation: Conversation;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: (conversation: Conversation) => void;
  onDelete: (conversationId: number) => void;
}

export default function IndividualChatSettings({
  conversation,
  isOpen,
  onClose,
  onUpdate,
  onDelete,
}: IndividualChatSettingsProps) {
  const [isRenaming, setIsRenaming] = useState(false);
  const [newTitle, setNewTitle] = useState(conversation.title || '');
  const [isUpdating, setIsUpdating] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleRename = async () => {
    if (!newTitle.trim()) return;

    setIsUpdating(true);
    try {
      const updatedConversation = await messagingApi.updateConversation(conversation.id, newTitle.trim());
      onUpdate(updatedConversation);
      setIsRenaming(false);
    } catch (error) {
      console.error('Failed to rename conversation:', error);
      alert('Не удалось переименовать чат');
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      await messagingApi.deleteConversation(conversation.id);
      onDelete(conversation.id);
      onClose();
    } catch (error) {
      console.error('Failed to delete conversation:', error);
      alert('Не удалось удалить чат');
    } finally {
      setIsDeleting(false);
      setShowDeleteConfirm(false);
    }
  };

  const handleClose = () => {
    setIsRenaming(false);
    setNewTitle(conversation.title || '');
    setShowDeleteConfirm(false);
    onClose();
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Настройки чата</DialogTitle>
            <DialogDescription>
              Управляйте настройками этого чата
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Current title */}
            <div>
              <Label className="text-sm font-medium">Название чата</Label>
              <div className="flex items-center gap-2 mt-1">
                {isRenaming ? (
                  <div className="flex-1 flex gap-2">
                    <Input
                      value={newTitle}
                      onChange={(e) => setNewTitle(e.target.value)}
                      placeholder="Введите название чата"
                      disabled={isUpdating}
                    />
                    <Button
                      size="sm"
                      onClick={handleRename}
                      disabled={isUpdating || !newTitle.trim()}
                    >
                      {isUpdating ? 'Сохранение...' : 'Сохранить'}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        setIsRenaming(false);
                        setNewTitle(conversation.title || '');
                      }}
                      disabled={isUpdating}
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </div>
                ) : (
                  <>
                    <span className="flex-1 text-sm">
                      {conversation.title || 'Без названия'}
                    </span>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setIsRenaming(true)}
                    >
                      <Edit className="w-4 h-4" />
                    </Button>
                  </>
                )}
              </div>
            </div>

            {/* Delete button */}
            <div className="pt-4 border-t">
              <Button
                variant="destructive"
                onClick={() => setShowDeleteConfirm(true)}
                className="w-full"
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Удалить чат
              </Button>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={handleClose}>
              Закрыть
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <AlertDialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удалить чат?</AlertDialogTitle>
            <AlertDialogDescription>
              Это действие нельзя отменить. Чат и все сообщения будут удалены навсегда.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-red-600 hover:bg-red-700"
            >
              {isDeleting ? 'Удаление...' : 'Удалить'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}