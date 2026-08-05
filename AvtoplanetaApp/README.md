# Мобильная Автопланета

Flutter-клиент использует тот же API и те же данные, что веб-приложение.

## Запуск

По умолчанию приложение подключается к `https://backend-server.ru`:

```shell
flutter pub get
flutter run
```

Для другого окружения передайте URL при сборке:

```shell
flutter run --dart-define=API_BASE_URL=https://example.org
```

## Вход через Google

Для получения ID Token мобильной сборке нужен OAuth Client ID типа Web application,
совпадающий с `GOOGLE_CLIENT_ID` auth-service:

```shell
flutter run --dart-define=GOOGLE_SERVER_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

Также добавьте штатную конфигурацию Google Sign-In для Android и iOS в соответствии
с настройками OAuth проекта. Секрет OAuth в приложение добавлять нельзя.

## Проверки перед коммитом

```shell
flutter analyze
flutter test
flutter build apk --debug
```

## Сборка в GitHub Actions

Workflow `Build Android app` запускается при push и pull request в ветки
`OptimizedBackend` и `main`, только если изменились файлы в `AvtoplanetaApp/`.
Он выполняет анализ, тесты и release-сборку. Готовый APK доступен в артефактах
запуска GitHub Actions в течение 30 дней.

Сейчас release-вариант использует debug-ключ из Android-конфигурации проекта.
Перед публикацией в Google Play необходимо настроить постоянный release-ключ.
