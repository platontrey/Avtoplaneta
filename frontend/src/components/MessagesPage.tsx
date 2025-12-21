/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect } from 'react';
import { MessageCircle, RefreshCw } from 'lucide-react';
import type { Conversation, DromDialog, DromMessage } from '../features/messaging/types';
import { useAuth } from '../features/auth/hooks/useAuth';
import { messagingApi } from '../features/messaging/api/messagingApi';
import ChatList from './ChatList';
import ChatWindow from './ChatWindow';
import CreateChatDialog from './CreateChatDialog';
import { Button } from './ui/button';
import { Input } from './ui/input';

export default function MessagesPage() {
   const { user: currentUser, isLoading } = useAuth();
   const [activeTab, setActiveTab] = useState<'messages' | 'drom'>('messages');
   const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
   const [isCreateChatOpen, setIsCreateChatOpen] = useState(false);

   // Drom state
   const [selectedDromDialog, setSelectedDromDialog] = useState<DromDialog | null>(null);
   const [dromMessageInput, setDromMessageInput] = useState('');
   const [dromDialogs, setDromDialogs] = useState<DromDialog[]>([]);
   const [dromMessages, setDromMessages] = useState<DromMessage[]>([]);
   const [dromDialogsLoading, setDromDialogsLoading] = useState(true);
   const [dromMessagesLoading, setDromMessagesLoading] = useState(false);
   const [dromSending, setDromSending] = useState(false);
   const [dromFetching, setDromFetching] = useState(false);
   const [hasNewDromMessages, setHasNewDromMessages] = useState(false);
   const [previousMessageCount, setPreviousMessageCount] = useState(0);

  const handleSelectConversation = (conversation: Conversation) => {
    setSelectedConversation(conversation);
  };

  const handleCloseChat = () => {
    setSelectedConversation(null);
  };

  const handleCreateNewChat = () => {
    setIsCreateChatOpen(true);
  };

  const handleCreateChatSuccess = (conversation: Conversation) => {
    setSelectedConversation(conversation);
  };

  // Drom functions
  useEffect(() => {
    if (activeTab === 'drom') {
      loadDromDialogs();
    }
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'drom' && selectedDromDialog) {
      loadDromMessages(selectedDromDialog.dialog_id);
    } else {
      setDromMessages([]);
    }
  }, [selectedDromDialog, activeTab]);

  const loadDromDialogs = async () => {
    setDromDialogsLoading(true);
    try {
      const data = await messagingApi.getDromDialogs();
      setDromDialogs(data);
    } catch (error) {
      console.error('Failed to load Drom dialogs:', error);
    } finally {
      setDromDialogsLoading(false);
    }
  };

  const loadDromMessages = async (dialogId: string) => {
    setDromMessagesLoading(true);
    try {
      const data = await messagingApi.getDromMessages(dialogId);
      if (data.length > previousMessageCount) {
        setHasNewDromMessages(true);
      }
      setPreviousMessageCount(data.length);
      setDromMessages(data);
    } catch (error) {
      console.error('Failed to load Drom messages:', error);
    } finally {
      setDromMessagesLoading(false);
    }
  };

  const handleSelectDromDialog = (dialog: DromDialog) => {
    setSelectedDromDialog(dialog);
    setHasNewDromMessages(false);
  };

  const handleSendDromMessage = async () => {
    if (!selectedDromDialog || !dromMessageInput.trim()) return;
    setDromSending(true);
    try {
      await messagingApi.sendDromMessage(selectedDromDialog.dialog_id, dromMessageInput.trim());
      setDromMessageInput('');
      setHasNewDromMessages(false);
      await loadDromMessages(selectedDromDialog.dialog_id);
    } catch (error) {
      console.error('Failed to send Drom message:', error);
    } finally {
      setDromSending(false);
    }
  };

  const handleFetchDromMessages = async () => {
    setDromFetching(true);
    try {
      await loadDromDialogs();
    } catch (error) {
      console.error('Failed to fetch Drom dialogs:', error);
    } finally {
      setDromFetching(false);
    }
  };

  // Polling for Drom updates
  useEffect(() => {
    if (activeTab !== 'drom') return;

    let interval: NodeJS.Timeout;

    if (selectedDromDialog) {
      // Poll messages every 5 seconds when dialog is open
      interval = setInterval(async () => {
        try {
          await loadDromMessages(selectedDromDialog.dialog_id);
        } catch (error) {
          console.error('Failed to poll Drom messages:', error);
        }
      }, 5000);
    } else {
      // Poll dialogs every 30 seconds when no dialog is selected
      interval = setInterval(async () => {
        try {
          await loadDromDialogs();
        } catch (error) {
          console.error('Failed to poll Drom dialogs:', error);
        }
      }, 30000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [activeTab, selectedDromDialog]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
      </div>
    );
  }

  if (!currentUser) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-500 mb-4">Необходимо авторизоваться</p>
          <a href="/login" className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
            Войти
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-4 sm:p-6 lg:p-8">
      <div className="mb-6">
        <div className="flex items-center gap-3 mb-2">
          <MessageCircle className="w-8 h-8 text-primary" />
          <h1 className="text-3xl font-bold text-foreground">Сообщения</h1>
        </div>
        <p className="text-muted-foreground">
          Общайтесь с коллегами и управляйте чатами
        </p>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border mb-4">
        <button
          onClick={() => setActiveTab('messages')}
          className={`px-4 py-2 font-medium ${activeTab === 'messages' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground'}`}
        >
          Сообщения
        </button>
        <button
          onClick={() => {
            setActiveTab('drom');
            setHasNewDromMessages(false);
          }}
          className={`px-4 py-2 font-medium relative ${activeTab === 'drom' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground'}`}
        >
          Drom сообщения
          {hasNewDromMessages && (
            <div className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full"></div>
          )}
        </button>
      </div>

      <div className="bg-card rounded-lg border shadow-sm h-[calc(100vh-200px)] overflow-hidden">
        <div className="flex h-full">
          {activeTab === 'messages' ? (
            <>
              {/* Sidebar with chat list */}
              <div className="w-1/3 border-r border-border">
                <ChatList
                  onSelectConversation={handleSelectConversation}
                  selectedConversationId={selectedConversation?.id}
                  onCreateNewChat={handleCreateNewChat}
                  currentUser={currentUser}
                />
              </div>

              {/* Chat window */}
              <div className="flex-1">
                {selectedConversation ? (
                  <ChatWindow
                    conversation={selectedConversation}
                    currentUser={currentUser}
                    onClose={handleCloseChat}
                  />
                ) : (
                  <div className="flex items-center justify-center h-full text-muted-foreground">
                    <div className="text-center">
                      <MessageCircle className="w-16 h-16 mx-auto mb-4 opacity-50" />
                      <h3 className="text-lg font-medium mb-2">Выберите чат</h3>
                      <p>Выберите существующий чат или создайте новый</p>
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <>
              {/* Drom Sidebar */}
              <div className="w-1/3 border-r border-border flex flex-col">
                <div className="p-4 border-b border-border">
                  <div className="flex items-center justify-between">
                    <h3 className="font-medium">Drom диалоги</h3>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleFetchDromMessages}
                      disabled={dromFetching}
                    >
                      <RefreshCw className={`w-4 h-4 ${dromFetching ? 'animate-spin' : ''}`} />
                    </Button>
                  </div>
                </div>
                <div className="flex-1 overflow-y-auto">
                  {dromDialogsLoading ? (
                    <div className="p-4 space-y-2">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <div key={i} className="h-16 bg-muted rounded animate-pulse" />
                      ))}
                    </div>
                  ) : dromDialogs && dromDialogs.length > 0 ? (
                    dromDialogs.map((dialog: DromDialog) => (
                      <div
                        key={dialog.id}
                        className={`p-4 border-b border-border cursor-pointer hover:bg-accent/50 ${
                          selectedDromDialog?.id === dialog.id ? 'bg-accent' : ''
                        }`}
                        onClick={() => handleSelectDromDialog(dialog)}
                      >
                        <div className="font-medium">{dialog.interlocutor}</div>
                        <div className="text-sm text-muted-foreground truncate">
                          {dialog.last_message || 'Нет сообщений'}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {new Date(dialog.last_message_at).toLocaleString('ru-RU')}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="p-4 text-center text-muted-foreground">
                      Нет диалогов. Нажмите обновить для загрузки.
                    </div>
                  )}
                </div>
              </div>

              {/* Drom Chat window */}
              <div className="flex-1 flex flex-col">
                {selectedDromDialog ? (
                  <>
                    <div className="p-4 border-b border-border">
                      <h3 className="font-medium">{selectedDromDialog.interlocutor}</h3>
                    </div>
                    <div className="flex-1 overflow-y-auto p-4 space-y-4">
                      {dromMessagesLoading ? (
                        <div className="space-y-2">
                          {Array.from({ length: 3 }).map((_, i) => (
                            <div key={i} className="h-12 bg-muted rounded animate-pulse" />
                          ))}
                        </div>
                      ) : dromMessages && dromMessages.length > 0 ? (
                        dromMessages.map((message: DromMessage) => (
                          <div
                            key={message.id}
                            className={`flex ${message.direction === 'outgoing' ? 'justify-end' : 'justify-start'}`}
                          >
                            <div
                              className={`max-w-xs px-3 py-2 rounded-lg text-sm ${
                                message.direction === 'outgoing'
                                  ? 'bg-primary text-primary-foreground'
                                  : 'bg-muted'
                              }`}
                            >
                              <div>{message.text}</div>
                              <div className="text-xs opacity-70 mt-1">
                                {message.time}
                              </div>
                            </div>
                          </div>
                        ))
                      ) : (
                        <div className="text-center text-muted-foreground">
                          Нет сообщений
                        </div>
                      )}
                    </div>
                    <div className="p-4 border-t border-border">
                      <div className="flex gap-2">
                        <Input
                          value={dromMessageInput}
                          onChange={(e) => setDromMessageInput(e.target.value)}
                          placeholder="Введите сообщение..."
                          onKeyPress={(e) => e.key === 'Enter' && handleSendDromMessage()}
                          disabled={dromSending}
                        />
                        <Button onClick={handleSendDromMessage} disabled={dromSending || !dromMessageInput.trim()}>
                          {dromSending ? 'Отправка...' : 'Отправить'}
                        </Button>
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="flex items-center justify-center h-full text-muted-foreground">
                    <div className="text-center">
                      <MessageCircle className="w-16 h-16 mx-auto mb-4 opacity-50" />
                      <h3 className="text-lg font-medium mb-2">Выберите диалог</h3>
                      <p>Выберите диалог для просмотра сообщений</p>
                    </div>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>

      <CreateChatDialog
        isOpen={isCreateChatOpen}
        onClose={() => setIsCreateChatOpen(false)}
        onSuccess={handleCreateChatSuccess}
      />
    </div>
  );
}