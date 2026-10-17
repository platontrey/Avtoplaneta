import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/features/messaging/models/messaging_models.dart';

void main() {
  group('Messaging Models Tests', () {
    test('Conversation.fromJson parses correctly', () {
      final json = {
        'id': 42,
        'participants': [1, 2, 3],
        'title': 'Рабочий чат',
        'created_at': '2026-10-15T10:00:00Z',
        'last_message_at': '2026-10-15T12:30:00Z',
        'last_message': 'Привет, запчасть приехала',
        'unread_count': 3,
      };

      final conv = Conversation.fromJson(json);

      expect(conv.id, 42);
      expect(conv.participants, [1, 2, 3]);
      expect(conv.title, 'Рабочий чат');
      expect(conv.lastMessage, 'Привет, запчасть приехала');
      expect(conv.unreadCount, 3);
    });

    test('Message.fromJson parses reactions and read_by', () {
      final json = {
        'id': 101,
        'conversation_id': 42,
        'sender_id': 1,
        'content': 'Отлично!',
        'message_type': 'text',
        'created_at': '2026-10-15T12:30:00Z',
        'read_by': [1, 2],
        'attachments': ['https://example.com/photo.jpg'],
        'reactions': [
          {'id': 1, 'message_id': 101, 'user_id': 2, 'emoji': '👍'},
          {'id': 2, 'message_id': 101, 'user_id': 3, 'emoji': '❤️'},
        ],
      };

      final msg = Message.fromJson(json);

      expect(msg.id, 101);
      expect(msg.content, 'Отлично!');
      expect(msg.readBy, [1, 2]);
      expect(msg.attachments.length, 1);
      expect(msg.reactions.length, 2);
      expect(msg.reactions[0].emoji, '👍');
      expect(msg.reactions[0].userId, 2);
    });

    test('DromDialog.fromJson and formattedTime', () {
      final json = {
        'id': 10,
        'dialog_id': 'drom_dlg_123',
        'interlocutor': 'Алексей',
        'created_at': '2026-10-14T08:00:00Z',
        'last_message_at': '2026-10-14T09:15:00Z',
        'last_message': 'Есть ли в наличии?',
        'is_unread': true,
        'part_title': 'Фара левая Toyota Camry',
        'part_price': '15 000 ₽',
        'part_url': 'https://drom.ru/part/12345',
        'user_rating': '★ 5.0',
        'user_status': 'онлайн',
      };

      final dlg = DromDialog.fromJson(json);

      expect(dlg.dialogId, 'drom_dlg_123');
      expect(dlg.interlocutor, 'Алексей');
      expect(dlg.isUnread, true);
      expect(dlg.partTitle, 'Фара левая Toyota Camry');
      expect(dlg.partPrice, '15 000 ₽');
      expect(dlg.userRating, '★ 5.0');
      expect(dlg.formattedTime, isNotEmpty);
    });

    test('DromMessage.fromJson identifies outgoing direction', () {
      final json = {
        'id': 505,
        'dialog_id': 'drom_dlg_123',
        'message_id': 'msg_88',
        'author': 'Менеджер',
        'direction': 'outgoing',
        'time': '14:20',
        'text': 'Да, фара в наличии на складе',
        'is_read': true,
        'created_at': '2026-10-14T09:20:00Z',
      };

      final msg = DromMessage.fromJson(json);

      expect(msg.isOutgoing, true);
      expect(msg.text, 'Да, фара в наличии на складе');
      expect(msg.time, '14:20');
    });

    test('ChatSettingsData serialization and defaults', () {
      const defaultSettings = ChatSettingsData();
      expect(defaultSettings.enableAutoRefresh, true);
      expect(defaultSettings.refreshInterval, 30);
      expect(defaultSettings.enableNotifications, true);

      final map = defaultSettings.toMap();
      final restored = ChatSettingsData.fromMap(map);

      expect(restored.enableAutoRefresh, true);
      expect(restored.refreshInterval, 30);
      expect(restored.enableNotifications, true);

      final custom = defaultSettings.copyWith(
        refreshInterval: 15,
        enableAutoRefresh: false,
      );
      expect(custom.refreshInterval, 15);
      expect(custom.enableAutoRefresh, false);
      expect(custom.enableNotifications, true);
    });
  });
}
