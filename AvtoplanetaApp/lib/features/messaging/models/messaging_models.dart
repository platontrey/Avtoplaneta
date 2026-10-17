import 'package:intl/intl.dart';

class Conversation {
  final int id;
  final List<int> participants;
  final String title;
  final DateTime createdAt;
  final DateTime lastMessageAt;
  final String? lastMessage;
  final int unreadCount;

  const Conversation({
    required this.id,
    required this.participants,
    required this.title,
    required this.createdAt,
    required this.lastMessageAt,
    this.lastMessage,
    this.unreadCount = 0,
  });

  factory Conversation.fromJson(Map<String, dynamic> json) {
    final rawParticipants = json['participants'];
    final participants = <int>[];
    if (rawParticipants is List) {
      for (final p in rawParticipants) {
        if (p is num) {
          participants.add(p.toInt());
        }
      }
    }

    final createdAtStr = json['created_at']?.toString() ?? '';
    final lastMessageAtStr = json['last_message_at']?.toString() ?? '';

    return Conversation(
      id: (json['id'] as num?)?.toInt() ?? 0,
      participants: participants,
      title: json['title']?.toString() ?? '',
      createdAt: DateTime.tryParse(createdAtStr) ?? DateTime.now(),
      lastMessageAt: DateTime.tryParse(lastMessageAtStr) ?? DateTime.now(),
      lastMessage: json['last_message']?.toString(),
      unreadCount: (json['unread_count'] as num?)?.toInt() ?? 0,
    );
  }

  Conversation copyWith({
    int? id,
    List<int>? participants,
    String? title,
    DateTime? createdAt,
    DateTime? lastMessageAt,
    String? lastMessage,
    int? unreadCount,
  }) {
    return Conversation(
      id: id ?? this.id,
      participants: participants ?? this.participants,
      title: title ?? this.title,
      createdAt: createdAt ?? this.createdAt,
      lastMessageAt: lastMessageAt ?? this.lastMessageAt,
      lastMessage: lastMessage ?? this.lastMessage,
      unreadCount: unreadCount ?? this.unreadCount,
    );
  }
}

class Message {
  final int id;
  final int conversationId;
  final int senderId;
  final String content;
  final String messageType;
  final String? voiceUrl;
  final DateTime createdAt;
  final List<int> readBy;
  final List<String> attachments;
  final List<Reaction> reactions;

  const Message({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.content,
    this.messageType = 'text',
    this.voiceUrl,
    required this.createdAt,
    this.readBy = const [],
    this.attachments = const [],
    this.reactions = const [],
  });

  factory Message.fromJson(Map<String, dynamic> json) {
    final rawReadBy = json['read_by'];
    final readBy = <int>[];
    if (rawReadBy is List) {
      for (final r in rawReadBy) {
        if (r is num) {
          readBy.add(r.toInt());
        }
      }
    }

    final rawAttachments = json['attachments'];
    final attachments = <String>[];
    if (rawAttachments is List) {
      for (final a in rawAttachments) {
        if (a is String) {
          attachments.add(a);
        }
      }
    }

    final rawReactions = json['reactions'];
    final reactions = <Reaction>[];
    if (rawReactions is List) {
      for (final r in rawReactions) {
        if (r is Map) {
          reactions.add(Reaction.fromJson(Map<String, dynamic>.from(r)));
        }
      }
    }

    final createdAtStr = json['created_at']?.toString() ?? '';

    return Message(
      id: (json['id'] as num?)?.toInt() ?? 0,
      conversationId: (json['conversation_id'] as num?)?.toInt() ?? 0,
      senderId: (json['sender_id'] as num?)?.toInt() ?? 0,
      content: json['content']?.toString() ?? '',
      messageType: json['message_type']?.toString() ?? 'text',
      voiceUrl: json['voice_url']?.toString(),
      createdAt: DateTime.tryParse(createdAtStr) ?? DateTime.now(),
      readBy: readBy,
      attachments: attachments,
      reactions: reactions,
    );
  }
}

class Reaction {
  final int id;
  final int messageId;
  final int userId;
  final String emoji;

  const Reaction({
    required this.id,
    required this.messageId,
    required this.userId,
    required this.emoji,
  });

  factory Reaction.fromJson(Map<String, dynamic> json) {
    return Reaction(
      id: (json['id'] as num?)?.toInt() ?? 0,
      messageId: (json['message_id'] as num?)?.toInt() ?? 0,
      userId: (json['user_id'] as num?)?.toInt() ?? 0,
      emoji: json['emoji']?.toString() ?? '',
    );
  }
}

class UserStatus {
  final int userId;
  final bool isOnline;
  final DateTime? lastSeen;

  const UserStatus({
    required this.userId,
    required this.isOnline,
    this.lastSeen,
  });

  factory UserStatus.fromJson(Map<String, dynamic> json) {
    final lastSeenStr = json['last_seen']?.toString();
    return UserStatus(
      userId: (json['user_id'] as num?)?.toInt() ?? 0,
      isOnline: json['is_online'] == true,
      lastSeen: lastSeenStr != null ? DateTime.tryParse(lastSeenStr) : null,
    );
  }
}

class ChatUser {
  final int id;
  final String name;
  final String username;
  final String email;
  final String role;
  final String initials;

  const ChatUser({
    required this.id,
    required this.name,
    this.username = '',
    required this.email,
    required this.role,
    this.initials = '',
  });

  factory ChatUser.fromJson(Map<String, dynamic> json) {
    final name = (json['name'] ?? json['username'] ?? json['email'] ?? 'Пользователь').toString();
    return ChatUser(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: name,
      username: json['username']?.toString() ?? '',
      email: json['email']?.toString() ?? '',
      role: json['role']?.toString() ?? 'operator',
      initials: (json['initials']?.toString().isNotEmpty == true)
          ? json['initials'].toString()
          : (name.isNotEmpty ? name[0].toUpperCase() : '?'),
    );
  }
}

// === DROM MODELS ===

class DromDialog {
  final int id;
  final String dialogId;
  final String interlocutor;
  final String? createdAt;
  final String? lastMessageAt;
  final String? lastMessage;
  final bool isUnread;
  final String? partTitle;
  final String? partPrice;
  final String? partUrl;
  final String? partPhotoUrl;
  final String? bulletinId;
  final String? userRating;
  final String? userStatus;

  const DromDialog({
    required this.id,
    required this.dialogId,
    required this.interlocutor,
    this.createdAt,
    this.lastMessageAt,
    this.lastMessage,
    this.isUnread = false,
    this.partTitle,
    this.partPrice,
    this.partUrl,
    this.partPhotoUrl,
    this.bulletinId,
    this.userRating,
    this.userStatus,
  });

  factory DromDialog.fromJson(Map<String, dynamic> json) {
    return DromDialog(
      id: (json['id'] as num?)?.toInt() ?? 0,
      dialogId: (json['dialog_id'] ?? json['id'] ?? '').toString(),
      interlocutor: json['interlocutor']?.toString() ?? 'Покупатель',
      createdAt: json['created_at']?.toString(),
      lastMessageAt: json['last_message_at']?.toString(),
      lastMessage: json['last_message']?.toString(),
      isUnread: json['is_unread'] == true,
      partTitle: json['part_title']?.toString(),
      partPrice: json['part_price']?.toString(),
      partUrl: json['part_url']?.toString(),
      partPhotoUrl: json['part_photo_url']?.toString(),
      bulletinId: json['bulletin_id']?.toString(),
      userRating: json['user_rating']?.toString(),
      userStatus: json['user_status']?.toString(),
    );
  }

  String get formattedTime {
    if (lastMessageAt == null || lastMessageAt!.isEmpty) return '';
    final dt = DateTime.tryParse(lastMessageAt!);
    if (dt != null) {
      final local = dt.toLocal();
      final now = DateTime.now();
      if (local.year == now.year && local.month == now.month && local.day == now.day) {
        return DateFormat('HH:mm').format(local);
      }
      return DateFormat('dd.MM.yyyy HH:mm').format(local);
    }
    return lastMessageAt!;
  }
}

class DromMessage {
  final int id;
  final String dialogId;
  final String messageId;
  final String author;
  final String direction; // "incoming" or "outgoing"
  final String time;
  final String text;
  final bool isRead;
  final String createdAt;

  const DromMessage({
    required this.id,
    required this.dialogId,
    required this.messageId,
    required this.author,
    required this.direction,
    required this.time,
    required this.text,
    this.isRead = false,
    required this.createdAt,
  });

  factory DromMessage.fromJson(Map<String, dynamic> json) {
    return DromMessage(
      id: (json['id'] as num?)?.toInt() ?? 0,
      dialogId: json['dialog_id']?.toString() ?? '',
      messageId: json['message_id']?.toString() ?? '',
      author: json['author']?.toString() ?? '',
      direction: json['direction']?.toString() ?? 'incoming',
      time: json['time']?.toString() ?? '',
      text: json['text']?.toString() ?? '',
      isRead: json['is_read'] == true,
      createdAt: json['created_at']?.toString() ?? '',
    );
  }

  bool get isOutgoing => direction == 'outgoing';
}

class DromAuthStatus {
  final bool authenticated;
  final int? dialogsCount;
  final String? checkedAt;
  final String? error;

  const DromAuthStatus({
    required this.authenticated,
    this.dialogsCount,
    this.checkedAt,
    this.error,
  });

  factory DromAuthStatus.fromJson(Map<String, dynamic> json) {
    return DromAuthStatus(
      authenticated: json['authenticated'] == true,
      dialogsCount: (json['dialogs_count'] as num?)?.toInt(),
      checkedAt: json['checked_at']?.toString(),
      error: json['error']?.toString(),
    );
  }
}

// === CHAT SETTINGS ===

class ChatSettingsData {
  final bool enableAutoRefresh;
  final int refreshInterval; // in seconds
  final bool enableNotifications;

  const ChatSettingsData({
    this.enableAutoRefresh = true,
    this.refreshInterval = 30,
    this.enableNotifications = true,
  });

  factory ChatSettingsData.fromMap(Map<dynamic, dynamic> map) {
    return ChatSettingsData(
      enableAutoRefresh: map['enableAutoRefresh'] as bool? ?? true,
      refreshInterval: (map['refreshInterval'] as num?)?.toInt() ?? 30,
      enableNotifications: map['enableNotifications'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toMap() => {
    'enableAutoRefresh': enableAutoRefresh,
    'refreshInterval': refreshInterval,
    'enableNotifications': enableNotifications,
  };

  ChatSettingsData copyWith({
    bool? enableAutoRefresh,
    int? refreshInterval,
    bool? enableNotifications,
  }) {
    return ChatSettingsData(
      enableAutoRefresh: enableAutoRefresh ?? this.enableAutoRefresh,
      refreshInterval: refreshInterval ?? this.refreshInterval,
      enableNotifications: enableNotifications ?? this.enableNotifications,
    );
  }
}
