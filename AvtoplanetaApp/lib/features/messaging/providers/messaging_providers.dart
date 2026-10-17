import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/storage/cache_storage.dart';
import '../models/messaging_models.dart';

// === REPOSITORY ===

class MessagingRepository {
  // --- Conversations ---
  static Future<List<Conversation>> getConversations() async {
    final response = await apiClient.dio.get('/api/messaging/conversations');
    final data = response.data;
    List<dynamic> list = [];
    if (data is List) {
      list = data;
    } else if (data is Map && data['conversations'] != null) {
      list = data['conversations'] as List<dynamic>;
    }
    return list
        .whereType<Map>()
        .map((item) => Conversation.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  static Future<Conversation> createConversation(List<int> participants, String? title) async {
    final payload = <String, dynamic>{
      'participants': participants,
    };
    if (title != null && title.trim().isNotEmpty) {
      payload['title'] = title.trim();
    }
    final response = await apiClient.dio.post(
      '/api/messaging/conversations',
      data: payload,
    );
    return Conversation.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<Conversation> updateConversation(int convId, String title, [List<int>? participants]) async {
    final payload = <String, dynamic>{
      'title': title,
    };
    if (participants != null) {
      payload['participants'] = participants;
    }
    final response = await apiClient.dio.put(
      '/api/messaging/conversations/$convId',
      data: payload,
    );
    return Conversation.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<void> deleteConversation(int convId) async {
    await apiClient.dio.delete('/api/messaging/conversations/$convId');
  }

  static Future<void> removeParticipant(int convId, int participantId) async {
    await apiClient.dio.delete('/api/messaging/conversations/$convId/participants/$participantId');
  }

  // --- Messages ---
  static Future<List<Message>> getMessages(int convId) async {
    final response = await apiClient.dio.get('/api/messaging/conversations/$convId/messages');
    final data = response.data;
    List<dynamic> list = [];
    if (data is Map && data['messages'] is List) {
      list = data['messages'] as List<dynamic>;
    } else if (data is List) {
      list = data;
    }
    return list
        .whereType<Map>()
        .map((item) => Message.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  static Future<Message> sendMessage(int convId, String content, [List<String>? attachments]) async {
    final response = await apiClient.dio.post(
      '/api/messaging/conversations/$convId/messages',
      data: {
        'content': content,
        if (attachments != null && attachments.isNotEmpty) 'attachments': attachments,
      },
    );
    return Message.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<void> deleteMessage(int messageId) async {
    await apiClient.dio.delete('/api/messaging/messages/$messageId');
  }

  static Future<void> markMessageRead(int messageId) async {
    await apiClient.dio.put('/api/messaging/messages/$messageId/read');
  }

  static Future<Reaction> addReaction(int messageId, String emoji) async {
    final response = await apiClient.dio.post(
      '/api/messaging/messages/$messageId/reactions',
      data: {'emoji': emoji},
    );
    return Reaction.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<void> removeReaction(int messageId, int reactionId) async {
    await apiClient.dio.delete('/api/messaging/messages/$messageId/reactions/$reactionId');
  }

  // --- Users & Statuses ---
  static Future<Map<int, ChatUser>> getUsers() async {
    final response = await apiClient.dio.get('/api/messaging/users');
    final data = response.data;
    final list = data is Map && data['users'] is List ? data['users'] as List : const [];
    final map = <int, ChatUser>{};
    for (final item in list.whereType<Map>()) {
      final user = ChatUser.fromJson(Map<String, dynamic>.from(item));
      if (user.id > 0) {
        map[user.id] = user;
      }
    }
    return map;
  }

  static Future<Map<int, UserStatus>> getUserStatuses() async {
    try {
      final response = await apiClient.dio.get('/api/messaging/users/status');
      final data = response.data;
      final list = data is Map && data['statuses'] is List ? data['statuses'] as List : const [];
      final map = <int, UserStatus>{};
      for (final item in list.whereType<Map>()) {
        final status = UserStatus.fromJson(Map<String, dynamic>.from(item));
        if (status.userId > 0) {
          map[status.userId] = status;
        }
      }
      return map;
    } catch (_) {
      return {};
    }
  }

  // --- Drom API ---
  static Future<List<DromDialog>> getDromDialogs() async {
    final response = await apiClient.dio.get('/api/messaging/drom/dialogs');
    final data = response.data;
    List<dynamic> list = [];
    if (data is List) {
      list = data;
    } else if (data is Map && data['dialogs'] != null) {
      list = data['dialogs'] as List<dynamic>;
    }
    return list
        .whereType<Map>()
        .map((item) => DromDialog.fromJson(Map<String, dynamic>.from(item)))
        .toList();
  }

  static Future<Map<String, dynamic>> getDromMessages(String dialogId) async {
    final response = await apiClient.dio.get(
      '/api/messaging/drom/messages',
      queryParameters: {'dialog_id': dialogId},
    );
    final data = response.data;
    List<DromMessage> messages = [];
    DromDialog? dialog;

    if (data is Map) {
      if (data['messages'] is List) {
        messages = (data['messages'] as List)
            .whereType<Map>()
            .map((item) => DromMessage.fromJson(Map<String, dynamic>.from(item)))
            .toList();
      }
      if (data['dialog'] is Map) {
        dialog = DromDialog.fromJson(Map<String, dynamic>.from(data['dialog'] as Map));
      }
    } else if (data is List) {
      messages = data
          .whereType<Map>()
          .map((item) => DromMessage.fromJson(Map<String, dynamic>.from(item)))
          .toList();
    }

    return {
      'messages': messages,
      'dialog': dialog,
    };
  }

  static Future<DromMessage> sendDromMessage(String dialogId, String content) async {
    final response = await apiClient.dio.post(
      '/api/messaging/drom/messages',
      data: {
        'dialog_id': dialogId,
        'content': content,
      },
    );
    return DromMessage.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<DromAuthStatus> getDromAuthStatus() async {
    final response = await apiClient.dio.get('/api/messaging/drom/auth/status');
    return DromAuthStatus.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<Map<String, dynamic>> loginDrom(String login, String password, {String? code}) async {
    final payload = <String, dynamic>{
      'login': login,
      'password': password,
    };
    if (code != null && code.trim().isNotEmpty) {
      payload['code'] = code.trim();
    }
    final response = await apiClient.dio.post(
      '/api/messaging/drom/auth/login',
      data: payload,
    );
    return Map<String, dynamic>.from(response.data as Map);
  }

  static Future<Map<String, dynamic>> saveDromSession(String sessionData) async {
    final response = await apiClient.dio.post(
      '/api/messaging/drom/auth/session',
      data: sessionData,
    );
    return Map<String, dynamic>.from(response.data as Map);
  }
}

// === RIVERPOD PROVIDERS ===

final conversationsProvider = FutureProvider<List<Conversation>>((ref) async {
  return MessagingRepository.getConversations();
});

final messagingUsersProvider = FutureProvider<Map<int, ChatUser>>((ref) async {
  return MessagingRepository.getUsers();
});

final userStatusesProvider = FutureProvider<Map<int, UserStatus>>((ref) async {
  return MessagingRepository.getUserStatuses();
});

final dromDialogsProvider = FutureProvider<List<DromDialog>>((ref) async {
  return MessagingRepository.getDromDialogs();
});

final dromAuthStatusProvider = FutureProvider<DromAuthStatus>((ref) async {
  return MessagingRepository.getDromAuthStatus();
});

// --- Settings StateNotifier ---
class ChatSettingsNotifier extends StateNotifier<ChatSettingsData> {
  ChatSettingsNotifier() : super(const ChatSettingsData()) {
    _loadSettings();
  }

  void _loadSettings() {
    try {
      final saved = CacheStorage.getChatSettings();
      if (saved != null) {
        state = ChatSettingsData.fromMap(saved);
      }
    } catch (_) {}
  }

  Future<void> _persist(ChatSettingsData settings) async {
    state = settings;
    try {
      await CacheStorage.saveChatSettings(settings.toMap());
    } catch (_) {}
  }

  void updateSettings(ChatSettingsData newSettings) {
    _persist(newSettings);
  }

  void setAutoRefresh(bool value) {
    _persist(state.copyWith(enableAutoRefresh: value));
  }

  void setRefreshInterval(int seconds) {
    _persist(state.copyWith(refreshInterval: seconds));
  }

  void setNotifications(bool value) {
    _persist(state.copyWith(enableNotifications: value));
  }

  void resetToDefaults() {
    _persist(const ChatSettingsData());
  }
}

final chatSettingsProvider = StateNotifierProvider<ChatSettingsNotifier, ChatSettingsData>((ref) {
  return ChatSettingsNotifier();
});
