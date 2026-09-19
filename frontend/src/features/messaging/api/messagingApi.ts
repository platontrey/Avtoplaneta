import type { Conversation, Message, User, UserStatus, Notification, Reaction, DromDialog, DromMessage } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
const MESSAGING_BASE_URL = `${API_BASE_URL}/api/v1/messaging`;

class MessagingApi {
  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-User-ID': localStorage.getItem('userId') || '',
    };
    const token = localStorage.getItem('token');
    if (token) headers.Authorization = `Bearer ${token}`;
    return headers;
  }

  // === ИСПОЛЬЗУЕМЫЕ МЕТОДЫ ===

  // Conversations
  async getConversations(): Promise<Conversation[]> { // ✅ используется в ChatList
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch conversations');
    const data = await response.json();
    if (!data) return [];
    if (Array.isArray(data)) return data;
    return data.conversations ?? data ?? [];
  }


  // Messages
  async getMessages(conversationId: number): Promise<Message[]> { // ✅ используется в ChatWindow
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}/messages`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch messages');
    const data = await response.json();
    return data.messages;
  }

  async sendMessage(conversationId: number, content: string, attachments?: string[]): Promise<Message> { // ✅ используется в ChatWindow
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}/messages`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ content, attachments }),
    });
    if (!response.ok) throw new Error('Failed to send message');
    return response.json();
  }

  async sendVoiceMessage(conversationId: number, voiceBlob: Blob): Promise<Message> {
    const formData = new FormData();
    formData.append('voice', voiceBlob, 'voice.wav');

    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}/messages/voice`, {
      method: 'POST',
      headers: (() => {
        const headers: Record<string, string> = {
          'X-User-ID': localStorage.getItem('userId') || '',
        };
        const token = localStorage.getItem('token');
        if (token) headers.Authorization = `Bearer ${token}`;
        return headers;
      })(),
      body: formData,
    });
    if (!response.ok) throw new Error('Failed to send voice message');
    return response.json();
  }

  async deleteMessage(messageId: number): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/messages/${messageId}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to delete message');
  }

  // User status
  async getUserStatuses(): Promise<UserStatus[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/users/status`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch user statuses');
    const data = await response.json();
    return data.statuses;
  }

  async updateUserStatus(isOnline: boolean): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/users/status`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify({ is_online: isOnline }),
    });
    if (!response.ok) throw new Error('Failed to update user status');
  }

  // === ДОПОЛНИТЕЛЬНЫЕ МЕТОДЫ ===

  // Conversations
  async createConversation(participants: number[], title?: string): Promise<Conversation> {
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ participants, title }),
    });
    if (!response.ok) throw new Error('Failed to create conversation');
    return response.json();
  }

  async updateConversation(conversationId: number, title: string): Promise<Conversation> {
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify({ title }),
    });
    if (!response.ok) throw new Error('Failed to update conversation');
    return response.json();
  }

  async deleteConversation(conversationId: number): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to delete conversation');
  }

  async removeParticipant(conversationId: number, participantId: number): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/conversations/${conversationId}/participants/${participantId}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to remove participant');
  }

  // Reactions
  async addReaction(messageId: number, emoji: string): Promise<Reaction> {
    const response = await fetch(`${MESSAGING_BASE_URL}/messages/${messageId}/reactions`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ emoji }),
    });
    if (!response.ok) throw new Error('Failed to add reaction');
    return response.json();
  }

  async removeReaction(messageId: number, reactionId: number): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/messages/${messageId}/reactions/${reactionId}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to remove reaction');
  }

  // Notifications
  async getNotifications(): Promise<Notification[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/notifications`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch notifications');
    const data = await response.json();
    return data.notifications;
  }

  async markNotificationRead(notificationId: number): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/notifications/${notificationId}/read`, {
      method: 'PUT',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to mark notification as read');
  }

  async markAllNotificationsRead(): Promise<void> {
    const response = await fetch(`${MESSAGING_BASE_URL}/notifications/read-all`, {
      method: 'PUT',
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to mark all notifications as read');
  }

  // Search
  async searchMessages(query: string): Promise<Message[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/search?q=${encodeURIComponent(query)}`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to search messages');
    const data = await response.json();
    return data.messages;
  }

  // Users
  async getUsers(): Promise<User[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/users`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch users');
    const data = await response.json();
    return data.users;
  }

  // === DROM METHODS ===

  async getDromDialogs(): Promise<DromDialog[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/drom/dialogs`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch Drom dialogs');
    const data = await response.json();
    return data.dialogs;
  }

  async getDromMessages(dialogId: string): Promise<DromMessage[]> {
    const response = await fetch(`${MESSAGING_BASE_URL}/drom/messages?dialog_id=${encodeURIComponent(dialogId)}`, {
      headers: this.getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch Drom messages');
    const data = await response.json();
    return data.messages;
  }

  async sendDromMessage(dialogId: string, content: string): Promise<DromMessage> {
    const response = await fetch(`${MESSAGING_BASE_URL}/drom/messages`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ dialog_id: dialogId, content }),
    });
    if (!response.ok) throw new Error('Failed to send Drom message');
    return response.json();
  }

  async fetchDromMessages(): Promise<DromDialog[]> {
    // Assuming this refreshes dialogs
    return this.getDromDialogs();
  }

}

export const messagingApi = new MessagingApi();
