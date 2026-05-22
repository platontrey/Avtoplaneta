import React, { useState, useEffect, useRef } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Conversation, Message, User } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';

interface VoiceMessagePlayerProps {
  voiceUrl: string;
}

const VoiceMessagePlayer: React.FC<VoiceMessagePlayerProps> = ({ voiceUrl }) => {
  console.log('VoiceMessagePlayer: Rendering with voiceUrl:', voiceUrl);
  const [isPlaying, setIsPlaying] = useState(false);
  const [duration, setDuration] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);
  const audioRef = useRef<HTMLAudioElement>(null);

    const togglePlay = () => {
    if (audioRef.current) {
      if (isPlaying) {
        audioRef.current.pause();
      } else {
        audioRef.current.play();
      }
      setIsPlaying(!isPlaying);
    }
  };

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration);
      console.log('Duration loaded:', audioRef.current.duration);
    }
  };

  const handleCanPlay = () => {
    if (audioRef.current && !duration) {
      setDuration(audioRef.current.duration);
      console.log('Can play, duration:', audioRef.current.duration);
    }
  };

  const handleTimeUpdate = () => {
    if (audioRef.current) {
      setCurrentTime(audioRef.current.currentTime);
      console.log('Time update:', audioRef.current.currentTime);
    }
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (audioRef.current) {
      const newTime = parseFloat(e.target.value);
      audioRef.current.currentTime = newTime;
      setCurrentTime(newTime);
    }
  };

  const formatTime = (time: number) => {
    const minutes = Math.floor(time / 60);
    const seconds = Math.floor(time % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  return (
    <div className="flex items-center gap-2 w-full">
      <button
        onClick={togglePlay}
        className="p-2 bg-gray-100 dark:bg-gray-700 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors flex-shrink-0"
      >
        {isPlaying ? (
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
          </svg>
        ) : (
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z"/>
          </svg>
        )}
      </button>
      <div className="flex-1 flex flex-col gap-1">
        <input
          type="range"
          min="0"
          max={duration || 1}
          value={currentTime}
          step="0.1"
          onChange={handleSeek}
          className="w-full h-1 bg-gray-200 dark:bg-gray-600 rounded-lg appearance-none cursor-pointer slider"
          style={{
            background: `linear-gradient(to right, #3b82f6 0%, #3b82f6 ${(currentTime / (duration || 1)) * 100}%, #e5e7eb ${(currentTime / (duration || 1)) * 100}%, #e5e7eb 100%)`
          }}
        />
        <div className="flex justify-between text-xs text-gray-500">
          <span>{formatTime(currentTime)}</span>
          <span>{formatTime(duration)}</span>
        </div>
      </div>
      <audio
        ref={audioRef}
        src={voiceUrl}
        preload="metadata"
        onEnded={() => setIsPlaying(false)}
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onLoadedMetadata={handleLoadedMetadata}
        onCanPlay={handleCanPlay}
        onTimeUpdate={handleTimeUpdate}
        onError={(e) => console.error('VoiceMessagePlayer: Audio error for URL:', voiceUrl, 'Error:', e)}
        onLoadStart={() => console.log('VoiceMessagePlayer: Audio load start for URL:', voiceUrl)}
      />
    </div>
  );
};

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
  const [isRecording, setIsRecording] = useState(false);
  const [mediaRecorder, setMediaRecorder] = useState<MediaRecorder | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const handleDeleteMessage = async (messageId: number) => {
    try {
      await messagingApi.deleteMessage(messageId);
      // Remove the message from the local state
      setMessages(prev => prev.filter(msg => msg.id !== messageId));
    } catch (error) {
      console.error('Failed to delete message:', error);
    }
  };

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

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const recorder = new MediaRecorder(stream);
      const chunks: Blob[] = [];

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunks.push(e.data);
        }
      };

      recorder.onstop = async () => {
        const blob = new Blob(chunks, { type: 'audio/wav' });
        try {
          const message = await messagingApi.sendVoiceMessage(conversation.id, blob);
          setMessages(prev => [...prev, message]);
        } catch (error) {
          console.error('Failed to send voice message:', error);
        }
        // Stop all tracks
        stream.getTracks().forEach(track => track.stop());
      };

      setMediaRecorder(recorder);
      recorder.start();
      setIsRecording(true);
    } catch (error) {
      console.error('Failed to start recording:', error);
    }
  };

  const stopRecording = () => {
    if (mediaRecorder && isRecording) {
      mediaRecorder.stop();
      setIsRecording(false);
      setMediaRecorder(null);
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

    console.log('renderMessage: message.id:', message.id, 'sender_id:', message.sender_id, 'type:', typeof message.sender_id, 'currentUser.id:', currentUser.id, 'type:', typeof currentUser.id, 'isOwn:', isOwn);

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
          <div className={`group relative max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
            isOwn
              ? 'bg-blue-600 text-white'
              : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
          }`}>
            {((conversation.participants.length > 2 && sender) || (!isOwn && sender)) && (
              <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{sender.name}</p>
            )}
            {message.message_type === 'voice' && message.voice_url ? (
              <VoiceMessagePlayer voiceUrl={message.voice_url} />
            ) : (
              <p className="text-sm">{message.content}</p>
            )}
            <div className={`text-xs mt-1 flex items-center justify-between ${
              isOwn ? 'text-blue-200' : 'text-gray-500 dark:text-gray-400'
            }`}>
              <span>{formatTime(message.created_at)}</span>
              {message.sender_id === currentUser.id && (
                <button
                  onClick={() => handleDeleteMessage(message.id)}
                  className="text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 ml-2 p-1"
                  title="Удалить сообщение"
                >
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              )}
            </div>
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
            type="button"
            onClick={isRecording ? stopRecording : startRecording}
            variant={isRecording ? "destructive" : "outline"}
            size="default"
            className={isRecording ? "animate-pulse" : ""}
          >
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 1a4 4 0 0 0-4 4v6a4 4 0 0 0 8 0V5a4 4 0 0 0-4-4z"/>
              <path d="M19 10v1a7 7 0 0 1-14 0v-1"/>
              <path d="M12 19v4"/>
              <path d="M8 23h8"/>
            </svg>
          </Button>
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