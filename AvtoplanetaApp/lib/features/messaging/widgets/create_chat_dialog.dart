import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../app/theme.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';

class CreateChatDialog extends StatefulWidget {
  final Map<int, ChatUser> users;
  final int? currentUserId;
  final Function(Conversation conversation) onSuccess;

  const CreateChatDialog({
    super.key,
    required this.users,
    required this.currentUserId,
    required this.onSuccess,
  });

  static Future<void> show(
    BuildContext context, {
    required Map<int, ChatUser> users,
    required int? currentUserId,
    required Function(Conversation conversation) onSuccess,
  }) {
    return showDialog(
      context: context,
      builder: (_) => CreateChatDialog(
        users: users,
        currentUserId: currentUserId,
        onSuccess: onSuccess,
      ),
    );
  }

  @override
  State<CreateChatDialog> createState() => _CreateChatDialogState();
}

class _CreateChatDialogState extends State<CreateChatDialog> {
  final _titleController = TextEditingController();
  final _searchController = TextEditingController();
  final _selectedUserIds = <int>{};
  bool _isCreating = false;
  String _searchQuery = '';

  @override
  void dispose() {
    _titleController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _create() async {
    if (_selectedUserIds.isEmpty) return;

    setState(() => _isCreating = true);
    try {
      final title = _titleController.text.trim();
      final conversation = await MessagingRepository.createConversation(
        _selectedUserIds.toList(),
        title.isNotEmpty ? title : null,
      );
      if (mounted) {
        Navigator.pop(context);
        widget.onSuccess(conversation);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка создания чата: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isCreating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final availableUsers = widget.users.values
        .where((u) => u.id != widget.currentUserId)
        .where((u) =>
            _searchQuery.isEmpty ||
            u.name.toLowerCase().contains(_searchQuery.toLowerCase()) ||
            u.email.toLowerCase().contains(_searchQuery.toLowerCase()))
        .toList()
      ..sort((a, b) => a.name.compareTo(b.name));

    return Dialog(
      backgroundColor: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: const BorderSide(color: AppTheme.borderColor),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 420),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Новый диалог',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(LucideIcons.x, size: 20),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const SizedBox(height: 14),

              // Title Field
              TextField(
                controller: _titleController,
                decoration: const InputDecoration(
                  labelText: 'Название чата (опционально)',
                  hintText: 'например, Склад или Вопрос по заказу',
                  prefixIcon: Icon(LucideIcons.message_square),
                ),
              ),
              const SizedBox(height: 14),

              // Search User Field
              TextField(
                controller: _searchController,
                onChanged: (val) => setState(() => _searchQuery = val.trim()),
                decoration: InputDecoration(
                  hintText: 'Поиск коллег…',
                  prefixIcon: const Icon(LucideIcons.search, size: 18),
                  suffixIcon: _searchQuery.isNotEmpty
                      ? IconButton(
                          icon: const Icon(LucideIcons.x, size: 16),
                          onPressed: () {
                            _searchController.clear();
                            setState(() => _searchQuery = '');
                          },
                        )
                      : null,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                ),
              ),
              const SizedBox(height: 10),

              // User Selection List
              Text(
                'Выберите участников (${_selectedUserIds.length})',
                style: const TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 6),
              ConstrainedBox(
                constraints: const BoxConstraints(maxHeight: 240),
                child: availableUsers.isEmpty
                    ? const Padding(
                        padding: EdgeInsets.symmetric(vertical: 24),
                        child: Center(
                          child: Text(
                            'Пользователи не найдены',
                            style: TextStyle(color: AppTheme.mutedColor),
                          ),
                        ),
                      )
                    : ListView.separated(
                        shrinkWrap: true,
                        itemCount: availableUsers.length,
                        separatorBuilder: (_, _) => const SizedBox(height: 6),
                        itemBuilder: (_, index) {
                          final user = availableUsers[index];
                          final isSelected = _selectedUserIds.contains(user.id);

                          return InkWell(
                            onTap: () {
                              setState(() {
                                if (isSelected) {
                                  _selectedUserIds.remove(user.id);
                                } else {
                                  _selectedUserIds.add(user.id);
                                }
                              });
                            },
                            borderRadius: BorderRadius.circular(12),
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                              decoration: BoxDecoration(
                                color: isSelected
                                    ? AppTheme.primaryColor.withValues(alpha: 0.12)
                                    : AppTheme.cardColor,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: isSelected
                                      ? AppTheme.primaryColor.withValues(alpha: 0.5)
                                      : AppTheme.borderColor,
                                ),
                              ),
                              child: Row(
                                children: [
                                  CircleAvatar(
                                    radius: 16,
                                    backgroundColor: isSelected
                                        ? AppTheme.primaryColor
                                        : AppTheme.primaryColor.withValues(alpha: 0.18),
                                    child: Text(
                                      user.initials,
                                      style: TextStyle(
                                        fontSize: 12,
                                        fontWeight: FontWeight.bold,
                                        color: isSelected ? Colors.white : AppTheme.primaryColor,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 10),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          user.name,
                                          style: const TextStyle(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w600,
                                            color: Colors.white,
                                          ),
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                        Text(
                                          '${user.role} • ${user.email}',
                                          style: const TextStyle(
                                            fontSize: 11,
                                            color: AppTheme.mutedColor,
                                          ),
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ],
                                    ),
                                  ),
                                  Checkbox(
                                    value: isSelected,
                                    activeColor: AppTheme.primaryColor,
                                    onChanged: (val) {
                                      setState(() {
                                        if (val == true) {
                                          _selectedUserIds.add(user.id);
                                        } else {
                                          _selectedUserIds.remove(user.id);
                                        }
                                      });
                                    },
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
              ),

              const SizedBox(height: 20),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: const Text('Отмена'),
                  ),
                  const SizedBox(width: 8),
                  FilledButton.icon(
                    icon: _isCreating
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          )
                        : const Icon(LucideIcons.plus, size: 18),
                    label: const Text('Создать'),
                    onPressed: _selectedUserIds.isEmpty || _isCreating ? null : _create,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
