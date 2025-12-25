/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect } from 'react';
import { Edit, Trash2, X, Users, UserMinus } from 'lucide-react';
import type { Conversation, User } from '../features/messaging/types';
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
  const [users, setUsers] = useState<User[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  const [removingParticipant, setRemovingParticipant] = useState<number | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadUsers();
      setNewTitle(conversation.title || '');
    }
  }, [isOpen, conversation.title]);

  const loadUsers = async () => {
    setIsLoadingUsers(true);
    try {
      const fetchedUsers = await messagingApi.getUsers();
      setUsers(fetchedUsers);
      const userId = parseInt(localStorage.getItem('userId') || '0');
      const current = fetchedUsers.find(u => u.id === userId) || null;
      setCurrentUser(current);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setIsLoadingUsers(false);
    }
  };

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

  const handleRemoveParticipant = async (participantId: number) => {
    setRemovingParticipant(participantId);
    try {
      await messagingApi.removeParticipant(conversation.id, participantId);
      // Update conversation participants
      const updatedConversation = {
        ...conversation,
        participants: conversation.participants.filter(id => id !== participantId)
      };
      onUpdate(updatedConversation);
    } catch (error) {
      console.error('Failed to remove participant:', error);
      alert('Не удалось удалить участника');
    } finally {
      setRemovingParticipant(null);
    }
  };

  const handleClose = () => {
    setIsRenaming(false);
    setNewTitle(conversation.title || '');
    setShowDeleteConfirm(false);
    setUsers([]);
    setCurrentUser(null);
    setRemovingParticipant(null);
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

            {/* Participants */}
            <div>
              <Label className="text-sm font-medium flex items-center gap-2 mb-2">
                <Users className="w-4 h-4" />
                Участники ({conversation.participants.length})
              </Label>
              <div className="space-y-2">
                {isLoadingUsers ? (
                  <div className="text-sm text-muted-foreground">Загрузка участников...</div>
                ) : (
                  conversation.participants.map(participantId => {
                    const user = users.find(u => u.id === participantId);
                    const isCurrentUser = currentUser?.id === participantId;
                    const canRemove = currentUser?.role === 'admin' && !isCurrentUser;
                    return (
                      <div key={participantId} className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center text-sm font-medium">
                            {user?.name?.[0] || '?'}
                          </div>
                          <div>
                            <div className="text-sm font-medium">
                              {user?.name || `Пользователь ${participantId}`}
                              {isCurrentUser && <span className="text-muted-foreground ml-1">(вы)</span>}
                            </div>
                            <div className="text-xs text-muted-foreground">{user?.role}</div>
                          </div>
                        </div>
                        {canRemove && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRemoveParticipant(participantId)}
                            disabled={removingParticipant === participantId}
                            className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          >
                            <UserMinus className="w-4 h-4" />
                          </Button>
                        )}
                      </div>
                    );
                  })
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