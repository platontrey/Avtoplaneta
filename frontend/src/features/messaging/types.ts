export interface Conversation {
  id: number;
  participants: number[];
  title?: string;
  created_at: string;
  last_message_at: string;
  last_message?: string;
  unread_count?: number;
}

export interface Message {
  id: number;
  conversation_id: number;
  sender_id: number;
  content: string;
  created_at: string;
  read_by: number[];
  attachments?: string[];
  mentions?: Mention[];
  reactions?: Reaction[];
}

export interface Mention {
  id: number;
  message_id: number;
  type: 'part' | 'order';
  resource_id: number;
  text: string;
}

export interface Reaction {
  id: number;
  message_id: number;
  user_id: number;
  emoji: string;
}

export interface Notification {
  id: number;
  user_id: number;
  type: 'message' | 'mention' | 'reaction';
  message: string;
  read: boolean;
  created_at: string;
  related_id: number;
}

export interface UserStatus {
  user_id: number;
  is_online: boolean;
  last_seen: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  role: string;
  initials?: string;
}