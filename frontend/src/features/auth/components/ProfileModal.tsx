/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState, useRef, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Camera, Trash2, KeyRound, User as UserIcon, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react';

import type { User } from '../types';
import { authApi } from '../api/authApi';
import { AUTH_ME_QUERY_KEY } from '../hooks/useCurrentUser';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface ProfileModalProps {
  user: User;
  isOpen: boolean;
  onClose: () => void;
}

export const ProfileModal: React.FC<ProfileModalProps> = ({ user, isOpen, onClose }) => {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Состояние профиля
  const [name, setName] = useState(user.name || '');
  const [username, setUsername] = useState(user.username || '');
  const [inn, setInn] = useState(user.inn || '');
  const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);
  const [profileMessage, setProfileMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);

  // Состояние смены пароля
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [passwordMessage, setPasswordMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);

  // Состояние аватара
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [avatarError, setAvatarError] = useState<string | null>(null);

  useEffect(() => {
    setName(user.name || '');
    setUsername(user.username || '');
    setInn(user.inn || '');
    setProfileMessage(null);
    setPasswordMessage(null);
    setAvatarError(null);
    setOldPassword('');
    setNewPassword('');
    setConfirmPassword('');
  }, [user, isOpen]);

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileMessage(null);
    setIsUpdatingProfile(true);

    try {
      const updated = await authApi.updateProfile({
        name: name.trim(),
        username: username.trim(),
        inn: inn.trim(),
      });

      queryClient.setQueryData(AUTH_ME_QUERY_KEY, (prev: User | null) => {
        if (!prev) return updated;
        return {
          ...prev,
          ...updated,
          initials: updated.name
            ? updated.name
                .split(' ')
                .map((n) => n[0])
                .join('')
                .toUpperCase()
            : '',
        };
      });

      setProfileMessage({ text: 'Данные профиля успешно сохранены', type: 'success' });
    } catch (err: any) {
      setProfileMessage({ text: err.message || 'Ошибка обновления профиля', type: 'error' });
    } finally {
      setIsUpdatingProfile(false);
    }
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordMessage(null);

    if (newPassword.length < 8) {
      setPasswordMessage({ text: 'Новый пароль должен содержать не менее 8 символов', type: 'error' });
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordMessage({ text: 'Пароли не совпадают', type: 'error' });
      return;
    }

    setIsChangingPassword(true);

    try {
      await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });

      setPasswordMessage({ text: 'Пароль успешно изменен', type: 'success' });
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setPasswordMessage({ text: err.message || 'Ошибка смены пароля', type: 'error' });
    } finally {
      setIsChangingPassword(false);
    }
  };

  const handleAvatarSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      setAvatarError('Файл слишком большой. Максимальный размер 5 МБ');
      return;
    }

    setAvatarError(null);
    setIsUploadingAvatar(true);

    try {
      const res = await authApi.uploadAvatar(file);
      queryClient.setQueryData(AUTH_ME_QUERY_KEY, (prev: User | null) => {
        if (!prev) return res.user;
        return {
          ...prev,
          ...res.user,
          avatar_url: res.avatar_url,
        };
      });
    } catch (err: any) {
      setAvatarError(err.message || 'Не удалось загрузить аватар');
    } finally {
      setIsUploadingAvatar(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleDeleteAvatar = async () => {
    if (!user.avatar_url) return;
    setAvatarError(null);
    setIsUploadingAvatar(true);

    try {
      const updated = await authApi.deleteAvatar();
      queryClient.setQueryData(AUTH_ME_QUERY_KEY, (prev: User | null) => {
        if (!prev) return updated;
        return {
          ...prev,
          ...updated,
          avatar_url: '',
        };
      });
    } catch (err: any) {
      setAvatarError(err.message || 'Не удалось удалить аватар');
    } finally {
      setIsUploadingAvatar(false);
    }
  };

  const isGoogleUser = user.provider?.toLowerCase() === 'google';

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="text-xl">Личный кабинет</DialogTitle>
          <DialogDescription>
            Управление данными профиля, аватаром и безопасностью учетной записи
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="profile" className="w-full mt-2">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="profile" className="flex items-center gap-2">
              <UserIcon className="w-4 h-4" />
              Профиль и фото
            </TabsTrigger>
            <TabsTrigger value="security" className="flex items-center gap-2">
              <KeyRound className="w-4 h-4" />
              Безопасность
            </TabsTrigger>
          </TabsList>

          {/* Вкладка 1: Профиль и Аватар */}
          <TabsContent value="profile" className="space-y-4 pt-3">
            {/* Секция Аватара */}
            <div className="flex flex-col items-center justify-center p-4 bg-muted/40 rounded-xl border border-dashed border-border/80">
              <div className="relative group">
                <Avatar className="w-24 h-24 text-2xl font-semibold shadow-md ring-2 ring-background">
                  {user.avatar_url && <AvatarImage src={user.avatar_url} alt={user.name} className="object-cover" />}
                  <AvatarFallback className="bg-primary/10 text-primary">
                    {user.initials || user.name.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>

                {isUploadingAvatar && (
                  <div className="absolute inset-0 bg-background/80 rounded-full flex items-center justify-center">
                    <Loader2 className="w-6 h-6 animate-spin text-primary" />
                  </div>
                )}
              </div>

              <div className="flex items-center gap-2 mt-3">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  className="hidden"
                  onChange={handleAvatarSelect}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={isUploadingAvatar}
                  className="text-xs flex items-center gap-1.5"
                >
                  <Camera className="w-3.5 h-3.5" />
                  {user.avatar_url ? 'Заменить фото' : 'Загрузить фото'}
                </Button>

                {user.avatar_url && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleDeleteAvatar}
                    disabled={isUploadingAvatar}
                    className="text-xs text-destructive hover:bg-destructive/10 flex items-center gap-1.5"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                    Удалить
                  </Button>
                )}
              </div>

              {avatarError && (
                <p className="text-xs text-destructive mt-2 flex items-center gap-1">
                  <AlertCircle className="w-3.5 h-3.5" />
                  {avatarError}
                </p>
              )}
            </div>

            {/* Форма личных данных */}
            <form onSubmit={handleProfileSubmit} className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="profile-name" className="text-xs font-medium">ФИО / Отображаемое имя</Label>
                <Input
                  id="profile-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Иван Иванов"
                  required
                />
              </div>

              <div className="space-y-1">
                <Label htmlFor="profile-username" className="text-xs font-medium">Логин</Label>
                <Input
                  id="profile-username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="ivan_ivanov"
                />
              </div>

              <div className="space-y-1">
                <Label htmlFor="profile-inn" className="text-xs font-medium">ИНН продавца / организации</Label>
                <Input
                  id="profile-inn"
                  value={inn}
                  onChange={(e) => setInn(e.target.value)}
                  placeholder="7701234567"
                />
              </div>

              <div className="space-y-1">
                <Label htmlFor="profile-email" className="text-xs font-medium text-muted-foreground">Электронная почта</Label>
                <Input
                  id="profile-email"
                  value={user.email || 'Не указана'}
                  disabled
                  className="bg-muted text-muted-foreground cursor-not-allowed"
                />
              </div>

              {profileMessage && (
                <div
                  className={`p-2.5 rounded-lg text-xs flex items-center gap-2 ${
                    profileMessage.type === 'success'
                      ? 'bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20'
                      : 'bg-destructive/10 text-destructive border border-destructive/20'
                  }`}
                >
                  {profileMessage.type === 'success' ? (
                    <CheckCircle2 className="w-4 h-4 shrink-0" />
                  ) : (
                    <AlertCircle className="w-4 h-4 shrink-0" />
                  )}
                  <span>{profileMessage.text}</span>
                </div>
              )}

              <Button type="submit" disabled={isUpdatingProfile} className="w-full mt-2">
                {isUpdatingProfile ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Сохранение...
                  </>
                ) : (
                  'Сохранить изменения'
                )}
              </Button>
            </form>
          </TabsContent>

          {/* Вкладка 2: Безопасность и смена пароля */}
          <TabsContent value="security" className="space-y-4 pt-3">
            {isGoogleUser ? (
              <div className="p-4 rounded-xl bg-muted/60 border border-border flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-muted-foreground mt-0.5 shrink-0" />
                <div className="text-sm">
                  <p className="font-medium">Вход выполнен через Google</p>
                  <p className="text-muted-foreground text-xs mt-1">
                    Для учетных записей Google смену пароля необходимо производить в настройках вашего Google-аккаунта.
                  </p>
                </div>
              </div>
            ) : (
              <form onSubmit={handlePasswordSubmit} className="space-y-3">
                <div className="space-y-1">
                  <Label htmlFor="old-password" className="text-xs font-medium">Текущий пароль</Label>
                  <Input
                    id="old-password"
                    type="password"
                    value={oldPassword}
                    onChange={(e) => setOldPassword(e.target.value)}
                    placeholder="••••••••"
                    required
                  />
                </div>

                <div className="space-y-1">
                  <Label htmlFor="new-password" className="text-xs font-medium">Новый пароль (минимум 8 символов)</Label>
                  <Input
                    id="new-password"
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    placeholder="••••••••"
                    required
                    minLength={8}
                  />
                </div>

                <div className="space-y-1">
                  <Label htmlFor="confirm-password" className="text-xs font-medium">Повторите новый пароль</Label>
                  <Input
                    id="confirm-password"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="••••••••"
                    required
                    minLength={8}
                  />
                </div>

                {passwordMessage && (
                  <div
                    className={`p-2.5 rounded-lg text-xs flex items-center gap-2 ${
                      passwordMessage.type === 'success'
                        ? 'bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20'
                        : 'bg-destructive/10 text-destructive border border-destructive/20'
                    }`}
                  >
                    {passwordMessage.type === 'success' ? (
                      <CheckCircle2 className="w-4 h-4 shrink-0" />
                    ) : (
                      <AlertCircle className="w-4 h-4 shrink-0" />
                    )}
                    <span>{passwordMessage.text}</span>
                  </div>
                )}

                <Button type="submit" disabled={isChangingPassword} className="w-full mt-2">
                  {isChangingPassword ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      Смена пароля...
                    </>
                  ) : (
                    'Обновить пароль'
                  )}
                </Button>
              </form>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
};
