import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import '../../../app/theme.dart';
import '../models/messaging_models.dart';
import '../providers/messaging_providers.dart';

class DromAuthDialog extends StatefulWidget {
  final VoidCallback onSuccess;

  const DromAuthDialog({super.key, required this.onSuccess});

  static Future<void> show(BuildContext context, {required VoidCallback onSuccess}) {
    return showDialog(
      context: context,
      builder: (_) => DromAuthDialog(onSuccess: onSuccess),
    );
  }

  @override
  State<DromAuthDialog> createState() => _DromAuthDialogState();
}

class _DromAuthDialogState extends State<DromAuthDialog> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final _loginController = TextEditingController();
  final _passwordController = TextEditingController();
  final _codeController = TextEditingController();
  final _cookieController = TextEditingController();

  bool _isCheckingStatus = false;
  bool _isSubmitting = false;
  bool _needCode = false;
  String? _errorMessage;
  String? _successMessage;
  DromAuthStatus? _authStatus;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _checkStatus();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _loginController.dispose();
    _passwordController.dispose();
    _codeController.dispose();
    _cookieController.dispose();
    super.dispose();
  }

  Future<void> _checkStatus() async {
    setState(() => _isCheckingStatus = true);
    try {
      final status = await MessagingRepository.getDromAuthStatus();
      if (mounted) setState(() => _authStatus = status);
    } catch (_) {
      if (mounted) {
        setState(() => _authStatus = const DromAuthStatus(
              authenticated: false,
              error: 'Ошибка проверки связи с сервером',
            ));
      }
    } finally {
      if (mounted) setState(() => _isCheckingStatus = false);
    }
  }

  Future<void> _submitLogin() async {
    final login = _loginController.text.trim();
    final password = _passwordController.text.trim();
    final code = _codeController.text.trim();

    if (login.isEmpty || password.isEmpty) {
      setState(() => _errorMessage = 'Заполните логин и пароль');
      return;
    }

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
      _successMessage = null;
    });

    try {
      final res = await MessagingRepository.loginDrom(
        login,
        password,
        code: _needCode && code.isNotEmpty ? code : null,
      );

      final success = res['success'] == true;
      final needCode = res['need_code'] == true;

      if (success) {
        if (mounted) {
          setState(() {
            _successMessage = 'Успешная авторизация! Сессия Drom сохранена.';
            _needCode = false;
            _passwordController.clear();
            _codeController.clear();
          });
          await _checkStatus();
          widget.onSuccess();
        }
      } else if (needCode) {
        if (mounted) {
          setState(() {
            _needCode = true;
            _successMessage = res['message']?.toString() ??
                'Требуется ввести код подтверждения из SMS или приложения Drom';
          });
        }
      } else {
        if (mounted) {
          setState(() => _errorMessage = res['error']?.toString() ?? 'Не удалось войти на Drom');
        }
      }
    } catch (e) {
      if (mounted) setState(() => _errorMessage = 'Ошибка: $e');
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  Future<void> _submitCookie() async {
    final sessionData = _cookieController.text.trim();
    if (sessionData.isEmpty) {
      setState(() => _errorMessage = 'Вставьте данные сессии или куки');
      return;
    }

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
      _successMessage = null;
    });

    try {
      final res = await MessagingRepository.saveDromSession(sessionData);
      if (res['success'] == true) {
        if (mounted) {
          setState(() {
            _successMessage = 'Сессия успешно сохранена и проверена!';
            _cookieController.clear();
          });
          await _checkStatus();
          widget.onSuccess();
        }
      } else {
        if (mounted) {
          setState(() => _errorMessage = res['error']?.toString() ?? 'Ошибка сохранения сессии');
        }
      }
    } catch (e) {
      if (mounted) setState(() => _errorMessage = 'Ошибка: $e');
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
        side: const BorderSide(color: AppTheme.borderColor),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Подключение к Drom.ru',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  IconButton(
                    icon: _isCheckingStatus
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(LucideIcons.refresh_cw, size: 18),
                    onPressed: _isCheckingStatus ? null : _checkStatus,
                    tooltip: 'Проверить статус сессии',
                  ),
                ],
              ),
              const SizedBox(height: 4),
              const Text(
                'Авторизуйте аккаунт Drom для чтения сообщений и ответов покупателям деталей',
                style: TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 12,
                ),
              ),
              const SizedBox(height: 16),

              // Status Box
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppTheme.cardColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: Row(
                  children: [
                    Icon(
                      _authStatus?.authenticated == true
                          ? LucideIcons.circle_check_big
                          : LucideIcons.circle_alert,
                      size: 20,
                      color: _authStatus?.authenticated == true
                          ? AppTheme.successColor
                          : AppTheme.warningColor,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _authStatus?.authenticated == true
                                ? 'Сессия активна'
                                : 'Не авторизован',
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                              color: Colors.white,
                            ),
                          ),
                          if (_authStatus?.dialogsCount != null)
                            Text(
                              'Загружено диалогов: ${_authStatus!.dialogsCount}',
                              style: const TextStyle(
                                color: AppTheme.mutedColor,
                                fontSize: 11,
                              ),
                            ),
                        ],
                      ),
                    ),
                    if (_authStatus?.checkedAt != null)
                      Text(
                        _authStatus!.checkedAt!,
                        style: const TextStyle(
                          color: AppTheme.mutedColor,
                          fontSize: 10,
                        ),
                      ),
                  ],
                ),
              ),

              // Alerts
              if (_errorMessage != null) ...[
                const SizedBox(height: 10),
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: AppTheme.dangerColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: AppTheme.dangerColor.withValues(alpha: 0.4)),
                  ),
                  child: Row(
                    children: [
                      const Icon(LucideIcons.circle_alert, size: 16, color: AppTheme.dangerColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _errorMessage!,
                          style: const TextStyle(color: AppTheme.dangerColor, fontSize: 12),
                        ),
                      ),
                    ],
                  ),
                ),
              ],

              if (_successMessage != null) ...[
                const SizedBox(height: 10),
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: AppTheme.successColor.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: AppTheme.successColor.withValues(alpha: 0.4)),
                  ),
                  child: Row(
                    children: [
                      const Icon(LucideIcons.circle_check, size: 16, color: AppTheme.successColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _successMessage!,
                          style: const TextStyle(color: AppTheme.successColor, fontSize: 12),
                        ),
                      ),
                    ],
                  ),
                ),
              ],

              const SizedBox(height: 14),

              // Tabs
              Container(
                decoration: BoxDecoration(
                  color: AppTheme.cardColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: TabBar(
                  controller: _tabController,
                  indicatorSize: TabBarIndicatorSize.tab,
                  indicator: BoxDecoration(
                    color: AppTheme.primaryColor,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  labelColor: Colors.white,
                  unselectedLabelColor: AppTheme.mutedColor,
                  labelStyle: const TextStyle(fontSize: 12, fontWeight: FontWeight.w700),
                  tabs: const [
                    Tab(
                      icon: Icon(LucideIcons.key_round, size: 16),
                      text: 'Логин и пароль',
                    ),
                    Tab(
                      icon: Icon(LucideIcons.file_code, size: 16),
                      text: 'Cookie / Сессия',
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 14),

              SizedBox(
                height: 240,
                child: TabBarView(
                  controller: _tabController,
                  children: [
                    // Tab 1: Login & Password
                    SingleChildScrollView(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          TextField(
                            controller: _loginController,
                            decoration: const InputDecoration(
                              labelText: 'Телефон или Логин на Drom',
                              hintText: '+79991234567 или логин',
                              prefixIcon: Icon(LucideIcons.user),
                            ),
                          ),
                          const SizedBox(height: 10),
                          TextField(
                            controller: _passwordController,
                            obscureText: true,
                            decoration: const InputDecoration(
                              labelText: 'Пароль',
                              hintText: 'Пароль от аккаунта Drom',
                              prefixIcon: Icon(LucideIcons.lock),
                            ),
                          ),
                          if (_needCode) ...[
                            const SizedBox(height: 10),
                            TextField(
                              controller: _codeController,
                              keyboardType: TextInputType.number,
                              decoration: const InputDecoration(
                                labelText: 'Код подтверждения из SMS',
                                hintText: '4-6 цифр',
                                prefixIcon: Icon(LucideIcons.smartphone),
                              ),
                            ),
                          ],
                          const SizedBox(height: 14),
                          FilledButton.icon(
                            icon: _isSubmitting
                                ? const SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                  )
                                : const Icon(LucideIcons.log_in, size: 18),
                            label: Text(_needCode ? 'Подтвердить код' : 'Войти на Drom'),
                            onPressed: _isSubmitting ? null : _submitLogin,
                          ),
                        ],
                      ),
                    ),

                    // Tab 2: Raw Cookie
                    SingleChildScrollView(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          TextField(
                            controller: _cookieController,
                            maxLines: 5,
                            decoration: const InputDecoration(
                              labelText: 'Cookie / Session JSON',
                              hintText: 'Вставьте cookie из браузера или сессию drom_session.json',
                            ),
                          ),
                          const SizedBox(height: 12),
                          FilledButton.icon(
                            icon: _isSubmitting
                                ? const SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                                  )
                                : const Icon(LucideIcons.save, size: 18),
                            label: const Text('Сохранить сессию'),
                            onPressed: _isSubmitting ? null : _submitCookie,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('Закрыть'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
