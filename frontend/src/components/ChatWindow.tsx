import React, { useState, useEffect, useRef } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Conversation, Message, User } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';

interface ChatWindowProps {
  conversation: Conversation;
  currentUser: User;
  onClose: () => void;
}

const ChatWindow: React.FC<ChatWindowProps> = ({ conversation, currentUser, onClose }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadMessages();
    loadUsers();
  }, [conversation.id]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadMessages = async () => {
    try {
      const data = await messagingApi.getMessages(conversation.id);
      setMessages(data);
    } catch (error) {
      console.error('Failed to load messages:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadUsers = async () => {
    try {
      const data = await messagingApi.getUsers();
      console.log('Loaded users:', data);
      setUsers(data);
    } catch (error) {
      console.error('Failed to load users:', error);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMessage.trim()) return;

    try {
      const message = await messagingApi.sendMessage(conversation.id, newMessage.trim());
      setMessages(prev => [...prev, message]);
      setNewMessage('');
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  };

  const handleAddReaction = async (messageId: number, emoji: string) => {
    try {
      await messagingApi.addReaction(messageId, emoji);
      // Refresh messages to show new reaction
      const updatedMessages = await messagingApi.getMessages(conversation.id);
      setMessages(updatedMessages);
    } catch (error) {
      console.error('Failed to add reaction:', error);
    }
  };

  const handleRemoveReaction = async (messageId: number, reactionId: number) => {
    try {
      await messagingApi.removeReaction(messageId, reactionId);
      // Refresh messages to show removed reaction
      const updatedMessages = await messagingApi.getMessages(conversation.id);
      setMessages(updatedMessages);
    } catch (error) {
      console.error('Failed to remove reaction:', error);
    }
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (date.toDateString() === today.toDateString()) {
      return 'Сегодня';
    } else if (date.toDateString() === yesterday.toDateString()) {
      return 'Вчера';
    } else {
      return date.toLocaleDateString();
    }
  };

  const renderMessage = (message: Message, index: number) => {
    const isOwn = message.sender_id === currentUser.id;
    const sender = users.find(u => u.id === message.sender_id);
    const prevMessage = index > 0 ? messages[index - 1] : null;
    const showDate = !prevMessage || formatDate(message.created_at) !== formatDate(prevMessage.created_at);

    return (
      <div key={message.id}>
        {showDate && (
          <div className="flex justify-center my-4">
            <span className="px-3 py-1 text-xs text-gray-500 bg-gray-100 dark:bg-gray-800 rounded-full">
              {formatDate(message.created_at)}
            </span>
          </div>
        )}

        <div className={`flex mb-4 ${isOwn ? 'justify-end' : 'justify-start'}`}>
          <div className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
            isOwn
              ? 'bg-blue-600 text-white'
              : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
          }`}>
            {((conversation.participants.length > 2 && sender) || (!isOwn && sender)) && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{sender.name}</p>
            )}
            <p className="text-sm">{message.content}</p>
            <span className={`text-xs mt-1 block ${
              isOwn ? 'text-blue-200' : 'text-gray-500 dark:text-gray-400'
            }`}>
              {formatTime(message.created_at)}
            </span>
            {/* Reactions */}
            {message.reactions && message.reactions.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {message.reactions.map((reaction) => (
                  <button
                    key={reaction.id}
                    onClick={() => reaction.user_id === currentUser.id && handleRemoveReaction(message.id, reaction.id)}
                    className={`text-xs px-2 py-1 rounded-full border ${
                      reaction.user_id === currentUser.id
                        ? 'bg-blue-100 dark:bg-blue-900 border-blue-300 dark:border-blue-700 text-blue-800 dark:text-blue-200'
                        : 'bg-gray-100 dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-800 dark:text-gray-200'
                    } hover:bg-opacity-80 transition-colors`}
                  >
                    {reaction.emoji} {reaction.user_id === currentUser.id ? ' (вы)' : ''}
                  </button>
                ))}
              </div>
            )}
            {/* Add reaction buttons - only for other users' messages */}
            {message.sender_id !== currentUser.id && (
              <div className="flex gap-1 mt-2">
                {['👍', '❤️', '😂', '😮', '😢', '😡'].map((emoji) => (
                  <button
                    key={emoji}
                    onClick={() => handleAddReaction(message.id, emoji)}
                    className="text-xs hover:bg-gray-300 dark:hover:bg-gray-600 px-2 py-1 rounded transition-colors"
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  const getChatTitle = () => {
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

  const chatTitle = getChatTitle();
  console.log('Chat title:', chatTitle, 'users:', users, 'participants:', conversation.participants, 'currentUser.id:', currentUser.id);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            {chatTitle}
          </h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {conversation.participants.length} участников
          </p>
        </div>
        <button
          onClick={onClose}
          className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
            <div className="text-center">
              <svg className="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <p>Начните разговор</p>
            </div>
          </div>
        ) : (
          messages.map((message, index) => renderMessage(message, index))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Message Input */}
      <div className="p-4 border-t border-gray-200 dark:border-gray-700">
        <form onSubmit={handleSendMessage} className="flex gap-2">
          <Input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder="Введите сообщение..."
            className="flex-1"
          />
          <Button
            type="submit"
            disabled={!newMessage.trim()}
            size="default"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </Button>
        </form>
      </div>
    </div>
  );
};

export default ChatWindow;