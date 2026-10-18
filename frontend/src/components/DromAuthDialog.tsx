/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CheckCircle2, AlertCircle, RefreshCw, KeyRound, FileCode2 } from 'lucide-react';
import { messagingApi } from '../features/messaging/api/messagingApi';

interface DromAuthDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function DromAuthDialog({ isOpen, onClose, onSuccess }: DromAuthDialogProps) {
  const [activeTab, setActiveTab] = useState<'password' | 'session'>('session');
  const [isCheckingStatus, setIsCheckingStatus] = useState(false);
  const [authStatus, setAuthStatus] = useState<{
    authenticated: boolean;
    dialogs_count?: number;
    checked_at?: string;
    error?: string;
  } | null>(null);

  // Form states
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [needCode, setNeedCode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Raw session states
  const [rawSession, setRawSession] = useState('');
  const [isSavingSession, setIsSavingSession] = useState(false);

  useEffect(() => {
    if (isOpen) {
      checkStatus();
      setErrorMessage(null);
      setSuccessMessage(null);
      setNeedCode(false);
      setCode('');
    }
  }, [isOpen]);

  const checkStatus = async () => {
    setIsCheckingStatus(true);
    try {
      const res = await messagingApi.getDromAuthStatus();
      setAuthStatus(res);
    } catch {
      setAuthStatus({ authenticated: false, error: 'Ошибка проверки связи с сервером' });
    } finally {
      setIsCheckingStatus(false);
    }
  };

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!login.trim() || !password.trim()) {
      setErrorMessage('Заполните логин и пароль');
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const res = await messagingApi.loginDrom(login.trim(), password.trim(), needCode ? code.trim() : undefined);
      if (res.success) {
        setSuccessMessage('Успешная авторизация! Сессия Drom сохранена.');
        setNeedCode(false);
        setPassword('');
        setCode('');
        await checkStatus();
        onSuccess();
      } else if (res.need_code) {
        setNeedCode(true);
        setSuccessMessage(res.message || 'Требуется ввести код подтверждения из SMS или приложения Drom');
      } else if (res.captcha_required) {
        setActiveTab('session');
        setErrorMessage(res.error || 'Drom заблокировал форму входа проверкой reCAPTCHA. Пожалуйста, используйте авторизацию через Cookie.');
      } else {
        setErrorMessage(res.error || 'Не удалось войти на Drom');
      }
    } catch (err: unknown) {
      const error = err as Error;
      setErrorMessage(error.message || 'Ошибка отправки запроса авторизации');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSessionSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = rawSession.trim();
    if (!trimmed) {
      setErrorMessage('Вставьте данные сессии или куки');
      return;
    }

    setIsSavingSession(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const res = await messagingApi.saveDromSession(trimmed);
      if (res.success) {
        setSuccessMessage('Сессия успешно сохранена и проверена!');
        setRawSession('');
        await checkStatus();
        onSuccess();
      } else {
        setErrorMessage(res.error || 'Ошибка сохранения сессии');
      }
    } catch (err: unknown) {
      const error = err as Error;
      setErrorMessage(error.message || 'Ошибка сохранения данных сессии');
    } finally {
      setIsSavingSession(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="text-lg font-semibold">Подключение к Drom.ru</DialogTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={checkStatus}
              disabled={isCheckingStatus}
              title="Проверить статус сессии"
              className="h-8 w-8 p-0"
            >
              <RefreshCw className={`w-4 h-4 ${isCheckingStatus ? 'animate-spin' : ''}`} />
            </Button>
          </div>
          <DialogDescription className="text-xs text-muted-foreground">
            Авторизуйте ваш аккаунт Drom для чтения сообщений и ответов покупателям деталей
          </DialogDescription>
        </DialogHeader>

        {/* Current Connection Status */}
        <div className="bg-muted/50 p-3 rounded-lg border border-border flex items-center justify-between text-xs">
          <div className="flex items-center gap-2">
            {authStatus?.authenticated ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-500 flex-shrink-0" />
            ) : (
              <AlertCircle className="w-4 h-4 text-amber-500 flex-shrink-0" />
            )}
            <div>
              <span className="font-medium">
                {authStatus?.authenticated ? 'Сессия активна' : 'Не авторизован'}
              </span>
              {authStatus?.dialogs_count !== undefined && (
                <span className="text-muted-foreground ml-1.5">
                  ({authStatus.dialogs_count} диалогов)
                </span>
              )}
            </div>
          </div>
          {authStatus?.checked_at && (
            <span className="text-[11px] text-muted-foreground">{authStatus.checked_at}</span>
          )}
        </div>

        {/* Alerts */}
        {errorMessage && (
          <div className="bg-destructive/10 text-destructive border border-destructive/20 p-2.5 rounded-md text-xs flex items-start justify-between gap-2">
            <div className="flex items-start gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>{errorMessage}</span>
            </div>
            {activeTab === 'password' && (
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setActiveTab('session')}
                className="h-6 text-[10px] px-2 whitespace-nowrap ml-2 border-destructive/30 hover:bg-destructive/10 shrink-0"
              >
                Ввести Cookie
              </Button>
            )}
          </div>
        )}

        {successMessage && (
          <div className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 p-2.5 rounded-md text-xs flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 flex-shrink-0" />
            <span>{successMessage}</span>
          </div>
        )}

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'password' | 'session')} className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="session" className="text-xs gap-1.5 font-medium">
              <FileCode2 className="w-3.5 h-3.5" />
              Вставка Cookie (Надёжно)
            </TabsTrigger>
            <TabsTrigger value="password" className="text-xs gap-1.5">
              <KeyRound className="w-3.5 h-3.5" />
              Логин и пароль
            </TabsTrigger>
          </TabsList>

          {/* Tab 1: Manual Cookie input */}
          <TabsContent value="session" className="space-y-3 pt-2">
            <form onSubmit={handleSessionSubmit} className="space-y-3">
              <div className="bg-muted/40 p-2.5 rounded-md border border-border text-[11px] space-y-1.5 text-muted-foreground">
                <p className="font-semibold text-foreground">Инструкция по копированию Cookie (1 минута):</p>
                <ol className="list-decimal list-inside space-y-1">
                  <li>Откройте <a href="https://my.drom.ru" target="_blank" rel="noreferrer" className="text-primary hover:underline">my.drom.ru</a> в браузере под вашим аккаунтом.</li>
                  <li>Нажмите <kbd className="px-1 py-0.5 bg-muted rounded border text-[10px] font-mono">F12</kbd> (или Ctrl+Shift+I) &rarr; вкладка <strong>Сеть (Network)</strong> &rarr; обновите страницу.</li>
                  <li>Кликните на верхний запрос &rarr; скопируйте заголовок <strong>Cookie</strong> (содержит <code>ring=...</code>) и вставьте в поле ниже.</li>
                </ol>
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="drom-raw-session" className="text-xs">
                  Заголовок Cookie или содержимое drom_session.json
                </Label>
                <Textarea
                  id="drom-raw-session"
                  value={rawSession}
                  onChange={(e) => setRawSession(e.target.value)}
                  placeholder='Вставьте куки (например: ring=...; ring_session=...) или JSON'
                  disabled={isSavingSession}
                  className="text-xs font-mono min-h-[90px] resize-none"
                />
              </div>

              <Button
                type="submit"
                disabled={isSavingSession || !rawSession.trim()}
                className="w-full text-xs h-9"
              >
                {isSavingSession ? (
                  <>
                    <RefreshCw className="w-3.5 h-3.5 mr-2 animate-spin" />
                    Сохранение и проверка сессии...
                  </>
                ) : (
                  'Сохранить сессию'
                )}
              </Button>
            </form>
          </TabsContent>

          {/* Tab 2: Login & Password */}
          <TabsContent value="password" className="space-y-3 pt-2">
            <form onSubmit={handleLoginSubmit} className="space-y-3">
              <div className="bg-amber-500/10 text-amber-700 dark:text-amber-400 p-2 rounded-md border border-amber-500/20 text-[11px] leading-relaxed">
                Сервер Drom часто защищает прямой вход проверкой reCAPTCHA. Если автоматический вход не проходит, используйте вкладку «Вставка Cookie».
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="drom-login" className="text-xs">Телефон или Логин на Drom</Label>
                <Input
                  id="drom-login"
                  value={login}
                  onChange={(e) => setLogin(e.target.value)}
                  placeholder="+79991234567 или логин"
                  disabled={isSubmitting}
                  className="text-xs h-9"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="drom-password" className="text-xs">Пароль</Label>
                <Input
                  id="drom-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Пароль от аккаунта"
                  disabled={isSubmitting}
                  className="text-xs h-9"
                  required
                />
              </div>

              {needCode && (
                <div className="space-y-1.5 bg-accent/40 p-3 rounded-md border border-border">
                  <Label htmlFor="drom-code" className="text-xs font-semibold text-primary">
                    Код подтверждения из SMS
                  </Label>
                  <Input
                    id="drom-code"
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    placeholder="Введите 4-6 цифр из SMS"
                    disabled={isSubmitting}
                    className="text-xs h-9 font-mono"
                    autoFocus
                    required
                  />
                  <p className="text-[11px] text-muted-foreground">
                    Drom запросил двухфакторное подтверждение. Введите код и нажмите «Подтвердить».
                  </p>
                </div>
              )}

              <Button
                type="submit"
                disabled={isSubmitting || !login.trim() || !password.trim()}
                className="w-full text-xs h-9"
              >
                {isSubmitting ? (
                  <>
                    <RefreshCw className="w-3.5 h-3.5 mr-2 animate-spin" />
                    {needCode ? 'Проверка кода...' : 'Авторизация в Drom...'}
                  </>
                ) : needCode ? (
                  'Подтвердить код'
                ) : (
                  'Войти в Drom'
                )}
              </Button>
            </form>
          </TabsContent>
        </Tabs>

        <DialogFooter className="pt-2">
          <Button variant="outline" size="sm" onClick={onClose} className="text-xs w-full sm:w-auto">
            Закрыть
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
