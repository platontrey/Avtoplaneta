/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { AdminUser } from '../types';

interface AddUserDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (user: {
    email: string;
    name: string;
    initials: string;
    inn: string;
    password: string;
    role: string;
  }) => Promise<void>;
  loading: boolean;
}

export const AddUserDialog: React.FC<AddUserDialogProps> = ({
  isOpen,
  onOpenChange,
  onSubmit,
  loading,
}) => {
  const [newUser, setNewUser] = useState({
    email: '',
    name: '',
    initials: '',
    inn: '',
    password: '',
    role: 'operator',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSubmit(newUser);
    setNewUser({ email: '', name: '', initials: '', inn: '', password: '', role: 'operator' });
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Добавить нового пользователя</DialogTitle>
          <DialogDescription>
            Создать новую учетную запись пользователя с email и паролем.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="add-email">Email</Label>
            <Input
              id="add-email"
              type="email"
              value={newUser.email}
              onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="add-name">Имя</Label>
            <Input
              id="add-name"
              value={newUser.name}
              onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="add-initials">Инициалы</Label>
            <Input
              id="add-initials"
              value={newUser.initials}
              onChange={(e) => setNewUser({ ...newUser, initials: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="add-inn">ИНН</Label>
            <Input
              id="add-inn"
              value={newUser.inn}
              onChange={(e) => setNewUser({ ...newUser, inn: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="add-password">Пароль</Label>
            <Input
              id="add-password"
              type="password"
              value={newUser.password}
              onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="add-role">Роль</Label>
            <Select
              value={newUser.role}
              onValueChange={(value) => setNewUser({ ...newUser, role: value })}
            >
              <SelectTrigger id="add-role">
                <SelectValue placeholder="Выберите роль" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="operator">Operator</SelectItem>
                <SelectItem value="manager">Manager</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={loading}>
              {loading ? 'Создание...' : 'Создать пользователя'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

interface EditUserDialogProps {
  user: AdminUser | null;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (userId: number, form: {
    name: string;
    email: string;
    initials: string;
    inn: string;
    role: string;
  }) => Promise<void>;
  loading: boolean;
}

export const EditUserDialog: React.FC<EditUserDialogProps> = ({
  user,
  isOpen,
  onClose,
  onSubmit,
  loading,
}) => {
  const [editForm, setEditForm] = useState({
    name: '',
    email: '',
    initials: '',
    inn: '',
    role: '',
  });

  useEffect(() => {
    if (user) {
      setEditForm({
        name: user.name,
        email: user.email,
        initials: user.initials || '',
        inn: user.inn || '',
        role: user.role,
      });
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;
    await onSubmit(user.id, editForm);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Редактировать пользователя</DialogTitle>
          <DialogDescription>
            Изменить данные пользователя {user?.name}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="edit-name">Имя</Label>
            <Input
              id="edit-name"
              value={editForm.name}
              onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="edit-email">Email</Label>
            <Input
              id="edit-email"
              type="email"
              value={editForm.email}
              onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="edit-initials">Инициалы</Label>
            <Input
              id="edit-initials"
              value={editForm.initials}
              onChange={(e) => setEditForm({ ...editForm, initials: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="edit-inn">ИНН</Label>
            <Input
              id="edit-inn"
              value={editForm.inn}
              onChange={(e) => setEditForm({ ...editForm, inn: e.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-role">Роль</Label>
            <Select
              value={editForm.role}
              onValueChange={(value) => setEditForm({ ...editForm, role: value })}
            >
              <SelectTrigger id="edit-role">
                <SelectValue placeholder="Выберите роль" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="operator">Operator</SelectItem>
                <SelectItem value="manager">Manager</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Сохранение...' : 'Сохранить'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
