/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { MessageCircle } from 'lucide-react';
import type { Conversation } from '../features/messaging/types';
import { useAuth } from '../features/auth/hooks/useAuth';
import ChatList from './ChatList';
import ChatWindow from './ChatWindow';
import CreateChatDialog from './CreateChatDialog';

export default function MessagesPage() {
  const { user: currentUser, isLoading } = useAuth();
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [isCreateChatOpen, setIsCreateChatOpen] = useState(false);

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

      <div className="bg-card rounded-lg border shadow-sm h-[calc(100vh-200px)] overflow-hidden">
        <div className="flex h-full">
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