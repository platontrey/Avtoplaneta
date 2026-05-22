import React, { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { Plus, MoreVertical, Edit } from "lucide-react";
import type { Conversation, UserStatus, User } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import IndividualChatSettings from './IndividualChatSettings';

interface ChatListProps {
  onSelectConversation: (conversation: Conversation) => void;
  selectedConversationId?: number;
  onCreateNewChat?: () => void;
  currentUser: User;
  onUpdateConversation?: (conversation: Conversation) => void;
  onDeleteConversation?: (conversationId: number) => void;
}

const ChatList: React.FC<ChatListProps> = ({
  onSelectConversation,
  selectedConversationId,
  onCreateNewChat,
  currentUser,
  onUpdateConversation,
  onDeleteConversation
}) => {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [userStatuses, setUserStatuses] = useState<UserStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [settingsConversation, setSettingsConversation] = useState<Conversation | null>(null);

  useEffect(() => {
    loadData().catch(console.error);
  }, []);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [conversationsData, usersData, statusesData] = await Promise.allSettled([
        messagingApi.getConversations(),
        messagingApi.getUsers(),
        messagingApi.getUserStatuses()
      ]);

      if (conversationsData.status === 'fulfilled') {
        setConversations(conversationsData.value);
      } else {
        console.error('Failed to load conversations:', conversationsData.reason);
        setError('Не удалось загрузить чаты. Проверьте, запущен ли messaging-service.');
      }

      if (usersData.status === 'fulfilled') {
        console.log('Loaded users in ChatList:', usersData.value);
        setUsers(usersData.value);
      } else {
        console.error('Failed to load users:', usersData.reason);
      }

      if (statusesData.status === 'fulfilled') {
        setUserStatuses(statusesData.value);
      } else {
        console.error('Failed to load user statuses:', statusesData.reason);
      }
    } catch (error) {
      console.error('Failed to load data:', error);
      setError('Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'только что';
    if (minutes < 60) return `${minutes} мин`;
    if (hours < 24) return `${hours} ч`;
    if (days < 7) return `${days} д`;
    return date.toLocaleDateString();
  };

  const getOnlineCount = (participantIds: number[]) => {
    return participantIds.filter(id =>
      userStatuses.find(status => status.user_id === id)?.is_online
    ).length;
  };

  const getChatTitle = (conversation: Conversation) => {
    if (conversation.title) return conversation.title;

    const participantNames = conversation.participants
      .filter(id => id !== currentUser.id)
      .map(id => users.find(u => u.id === id)?.name)
      .filter(name => name)
      .join(', ');

    if (conversation.participants.length === 2) {
      return participantNames || `Чат #${conversation.id}`;
    } else if (conversation.participants.length > 2) {
      const firstParticipant = participantNames.split(', ')[0];
      return firstParticipant ? `${firstParticipant}: Групповой чат` : 'Групповой чат';
    }
    return `Чат #${conversation.id}`;
  };

  const handleUpdateConversation = (updatedConversation: Conversation) => {
    setConversations(prev => prev.map(conv =>
      conv.id === updatedConversation.id ? updatedConversation : conv
    ));
    // Update settings conversation if it's the same chat
    if (settingsConversation && settingsConversation.id === updatedConversation.id) {
      setSettingsConversation(updatedConversation);
    }
    onUpdateConversation?.(updatedConversation);
  };

  const handleDeleteConversation = (conversationId: number) => {
    setConversations(prev => prev.filter(conv => conv.id !== conversationId));
    onDeleteConversation?.(conversationId);
  };

  const handleOpenSettings = (conversation: Conversation) => {
    setSettingsConversation(conversation);
  };

  const handleCloseSettings = () => {
    setSettingsConversation(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-red-500 mb-4">{error}</p>
          <button
            onClick={loadData}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Повторить
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Чаты</h2>
          {onCreateNewChat && (
            <Button
              onClick={onCreateNewChat}
              size="sm"
              variant="outline"
            >
              <Plus className="w-4 h-4 mr-2" />
              Новый чат
            </Button>
          )}
        </div>
      </div>

      <div className="divide-y divide-gray-200 dark:divide-gray-700">
        {conversations.length === 0 ? (
          <div className="p-4 text-center text-gray-500 dark:text-gray-400">
            Нет активных чатов
          </div>
        ) : (
          conversations.map((conversation) => (
            <div
              key={conversation.id}
              onClick={() => onSelectConversation(conversation)}
              className={`group p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors ${
                selectedConversationId === conversation.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {getChatTitle(conversation)}
                    </h3>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {formatTime(conversation.last_message_at)}
                      </span>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 p-0"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleOpenSettings(conversation)}>
                            <Edit className="w-4 h-4 mr-2" />
                            Настройки чата
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>

                  {conversation.last_message && (
                    <p className="text-sm text-gray-600 dark:text-gray-400 truncate mt-1">
                      {conversation.last_message}
                    </p>
                  )}

                  <div className="flex items-center justify-between mt-2">
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {conversation.participants.length} участников
                      </span>
                      {(() => {
                        const onlineCount = getOnlineCount(conversation.participants);
                        return onlineCount > 0 ? (
                          <div className="flex items-center gap-1">
                            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                            <span className="text-xs text-green-600 dark:text-green-400">
                              {onlineCount}
                            </span>
                          </div>
                        ) : null;
                      })()}
                    </div>
                    {conversation.unread_count && conversation.unread_count > 0 && (
                      <span className="inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white bg-blue-600 rounded-full">
                        {conversation.unread_count}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Individual Chat Settings */}
      {settingsConversation && (
        <IndividualChatSettings
          conversation={settingsConversation}
          isOpen={true}
          onClose={handleCloseSettings}
          onUpdate={handleUpdateConversation}
          onDelete={handleDeleteConversation}
        />
      )}
    </div>
  );
};

export default ChatList;