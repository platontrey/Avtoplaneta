/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { useState, useMemo, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { User } from '../features/auth/types';
import {
  Code2,
  Server,
  Shield,
  Search,
  Copy,
  Check,
  Download,
  Terminal,
  Boxes,
  Workflow,
  Sparkles,
  ChevronDown,
  ChevronRight,
  Package,
  ShoppingCart,
  BookOpen,
  HelpCircle,
  Smartphone,
  Truck,
  Camera,
  FileSpreadsheet,
  Users,
  Layers,
  FileCode,
  Radio,
  CheckCircle2,
  AlertTriangle,
  KeyRound
} from 'lucide-react';

interface ReadmeProps {
  user?: User | null;
  defaultTab?: 'guides' | 'api' | 'architecture' | 'roles' | 'faq' | 'operator' | 'manager';
}

interface ApiEndpoint {
  id: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  path: string;
  service: 'AuthService' | 'PartsService' | 'OrdersService' | 'MessagingService' | 'ExportService';
  title: string;
  description: string;
  authRequired: boolean;
  requiredRole?: 'admin' | 'manager' | 'operator' | 'any';
  parameters?: Array<{
    name: string;
    in: 'path' | 'query' | 'body' | 'header';
    type: string;
    required: boolean;
    description: string;
  }>;
  requestBodyExample?: string;
  responseExample?: string;
}

// -----------------------------------------------------------------------------
// РЕЕСТР ВСЕХ 45+ ЭНДПОИНТОВ МИКРОСЕРВИСНОЙ СИСТЕМЫ AVTOPLANETA
// -----------------------------------------------------------------------------
const FULL_API_ENDPOINTS: ApiEndpoint[] = [
  // === PARTS SERVICE ===
  {
    id: 'parts-inventory',
    method: 'GET',
    path: '/api/v1/inventory',
    service: 'PartsService',
    title: 'Каталог запчастей (Elasticsearch + PostgreSQL)',
    description: 'Возвращает список запчастей с фильтрацией, пагинацией и полнотекстовым поиском. Включает детали в наличии (quantity > 0) и дефектные ведомости (quantity = 0). Проданные (quantity = -1) исключаются.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'query', in: 'query', type: 'string', required: false, description: 'Поисковый запрос (артикул, марка, название, OEM)' },
      { name: 'category', in: 'query', type: 'string', required: false, description: 'Фильтр по категории детали' },
      { name: 'brand', in: 'query', type: 'string', required: false, description: 'Марка автомобиля' },
      { name: 'model', in: 'query', type: 'string', required: false, description: 'Модель автомобиля' },
      { name: 'page', in: 'query', type: 'integer', required: false, description: 'Номер страницы (дефолт 1)' },
      { name: 'limit', in: 'query', type: 'integer', required: false, description: 'Размер страницы (дефолт 50)' },
    ],
    responseExample: '{\n  "parts": [\n    {\n      "id": 1042,\n      "name": "Дверь передняя правая",\n      "brand": "Toyota",\n      "model": "Camry",\n      "price": 15000,\n      "quantity": 1,\n      "location": "Склад 1, Ряд B, Полка 4",\n      "photos": ["/uploads/parts/1042_front.webp"]\n    }\n  ],\n  "total": 12480\n}',
  },
  {
    id: 'parts-get-item',
    method: 'GET',
    path: '/api/v1/parts/item/{id}',
    service: 'PartsService',
    title: 'Карточка запчасти по ID',
    description: 'Возвращает полную карточку автозапчасти со всеми характеристиками, фотографиями, складской локацией и историей изменений.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'Уникальный ID запчасти' }
    ],
    responseExample: '{\n  "id": 1042,\n  "name": "Фара передняя левая LED",\n  "brand": "Nissan",\n  "model": "X-Trail",\n  "oem_code": "26060-4BA0A",\n  "price": 28000,\n  "quantity": 1,\n  "condition": "Оригинал, б/у",\n  "location": "Секция A-14"\n}',
  },
  {
    id: 'parts-create',
    method: 'POST',
    path: '/api/v1/parts',
    service: 'PartsService',
    title: 'Создание новой детали',
    description: 'Добавляет запчасть в базу данных PostgreSQL и моментально индексирует в Elasticsearch.',
    authRequired: true,
    requiredRole: 'any',
    requestBodyExample: '{\n  "name": "Фара левая LED",\n  "brand": "Nissan",\n  "model": "X-Trail",\n  "category": "Оптика",\n  "price": 28000,\n  "quantity": 1,\n  "location": "Секция A-14",\n  "photos": ["/uploads/parts/photo1.webp"]\n}',
  },
  {
    id: 'parts-update',
    method: 'PUT',
    path: '/api/v1/parts/{id}',
    service: 'PartsService',
    title: 'Обновление характеристик детали',
    description: 'Обновляет параметры запчасти (цену, фото, локацию, описание) и обновляет поисковый индекс.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID детали' }
    ],
  },
  {
    id: 'parts-delete',
    method: 'DELETE',
    path: '/api/v1/parts/{id}',
    service: 'PartsService',
    title: 'Мягкое удаление детали (Soft Delete)',
    description: 'Устанавливает deleted_at = NOW(), убирая запчасть из каталога, поиска и фидов выгрузки.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID запчасти' }
    ],
  },
  {
    id: 'parts-decrease',
    method: 'POST',
    path: '/api/v1/parts/{id}/decrease',
    service: 'PartsService',
    title: 'Списание количества детали',
    description: 'Уменьшает остаток. Если деталь списывается до 0 при продаже, выставляется quantity = -1 (маркер продажи).',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID детали' }
    ],
    requestBodyExample: '{\n  "amount": 1\n}',
  },
  {
    id: 'parts-increase',
    method: 'POST',
    path: '/api/v1/parts/{id}/increase',
    service: 'PartsService',
    title: 'Увеличение остатка детали (Возврат на склад)',
    description: 'Увеличивает остаток детали. Если деталь ранее имела quantity = -1, восстанавливает доступность на складе.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID детали' }
    ],
  },
  {
    id: 'parts-mark-deletion',
    method: 'POST',
    path: '/api/v1/parts/{id}/mark-deletion',
    service: 'PartsService',
    title: 'Отметка детали к утилизации / удалению через 14 дней',
    description: 'Выставляет to_delete_at таймер для отложенного списания залежавшихся неликвидных позиций.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'parts-photo-delete',
    method: 'DELETE',
    path: '/api/v1/parts/{partId}/photo',
    service: 'PartsService',
    title: 'Удаление конкретной фотографии детали',
    description: 'Удаляет фото из массива карточки и очищает бинарный файл с диска.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'parts-defect-reports',
    method: 'POST',
    path: '/api/v1/defect-reports',
    service: 'PartsService',
    title: 'Создание дефектной ведомости (Разбор авто)',
    description: 'Массовое создание разобранных или поврежденных деталей с quantity = 0. Они остаются полностью видимыми в каталоге.',
    authRequired: true,
    requiredRole: 'any',
    requestBodyExample: '{\n  "car_id": 42,\n  "parts": [\n    {\n      "name": "Бампер передний (трещина)",\n      "category": "Кузов наружные элементы",\n      "price": 5000,\n      "defect": "Трещина справа внизу, под пайку"\n    }\n  ]\n}',
  },
  {
    id: 'parts-defect-preview',
    method: 'POST',
    path: '/api/v1/defect-reports/preview',
    service: 'PartsService',
    title: 'Предпросмотр дефектной ведомости',
    description: 'Парсит структуру дефектовки и возвращает калькуляцию позиций перед реальной записью в базу.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'parts-stats',
    method: 'GET',
    path: '/api/v1/statistics',
    service: 'PartsService',
    title: 'Сводная статистика инвентаря',
    description: 'Агрегированные показатели: общее число деталей, складская сумма в рублях и распределение по категориям.',
    authRequired: true,
    requiredRole: 'manager',
    responseExample: '{\n  "total_parts": 8450,\n  "total_value": 42150000,\n  "categories": [\n    { "name": "Кузовные детали", "count": 2100 },\n    { "name": "Двигатель и навесное", "count": 1420 }\n  ]\n}',
  },
  {
    id: 'parts-stats-earnings',
    method: 'POST',
    path: '/api/v1/statistics/earnings',
    service: 'PartsService',
    title: 'Фиксация доходности по периодам',
    description: 'Обновляет аналитические показатели доходности по складам.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'parts-supplier-codes',
    method: 'GET',
    path: '/api/v1/admin/supplier-codes',
    service: 'PartsService',
    title: 'Список поставщиков и кодов',
    description: 'Возвращает справочник кодов поставщиков для массовых операций.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'parts-bulk-delete',
    method: 'POST',
    path: '/api/v1/admin/parts/bulk-delete',
    service: 'PartsService',
    title: 'Пакетное удаление деталей',
    description: 'Массовое удаление деталей по списку ID (до 10 000 позиций за раз).',
    authRequired: true,
    requiredRole: 'admin',
    requestBodyExample: '{\n  "ids": [101, 102, 103, 104]\n}',
  },
  {
    id: 'parts-bulk-update',
    method: 'POST',
    path: '/api/v1/admin/parts/bulk-update',
    service: 'PartsService',
    title: 'Пакетное обновление параметров',
    description: 'Массовое изменение поставщика, цен или локации группы деталей.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'parts-zero-quantity',
    method: 'POST',
    path: '/api/v1/admin/parts/zero-quantity',
    service: 'PartsService',
    title: 'Очистка нулевых остатков по поставщику',
    description: 'Удаляет неактуальные нулевые позиции конкретного поставщика.',
    authRequired: true,
    requiredRole: 'admin',
  },

  // === ORDERS SERVICE ===
  {
    id: 'orders-list',
    method: 'GET',
    path: '/orders',
    service: 'OrdersService',
    title: 'Список активных заказов',
    description: 'Возвращает текущие незавершенные заказы с цветовыми статусами (красный/коричневый/желтый), данными покупателя и составом позиций.',
    authRequired: true,
    requiredRole: 'manager',
    responseExample: '[\n  {\n    "id": 512,\n    "order_number": "ORD-2026-0512",\n    "status": "brown",\n    "status_text": "В сборке",\n    "buyer_number": "+7 (952) 809-40-32 Иванов Иван",\n    "transport_company": "СДЭК",\n    "created_at": "2026-03-01T10:15:00Z",\n    "items": [\n      { "id": 1, "part_id": 1042, "part_name": "Фара левая LED", "quantity": 1, "price": 28000 }\n    ]\n  }\n]',
  },
  {
    id: 'orders-completed',
    method: 'GET',
    path: '/orders/completed',
    service: 'OrdersService',
    title: 'История выполненных заказов',
    description: 'Архив заказов со статусом green (Выполнен), по которым произведено списание деталей.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-create',
    method: 'POST',
    path: '/orders',
    service: 'OrdersService',
    title: 'Создание нового заказа',
    description: 'Создает заказ, автоматически привязывает клиента по номеру телефона (+7...), резервирует детали и публикует событие в Redis Streams.',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "buyer_number": "+79528094032 Иванов Иван",\n  "transport_company": "СДЭК",\n  "notes": "Отправить до пятницы",\n  "discount": 500,\n  "items": [\n    { "part_id": 1042, "quantity": 1, "price": 28000 }\n  ]\n}',
  },
  {
    id: 'orders-update-details',
    method: 'PUT',
    path: '/orders/{id}',
    service: 'OrdersService',
    title: 'Редактирование параметров заказа',
    description: 'Обновляет ТК, трек-номер, скидку, примечания и контактные данные.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID заказа' }
    ],
  },
  {
    id: 'orders-update-status',
    method: 'PUT',
    path: '/orders/{id}/status',
    service: 'OrdersService',
    title: 'Смена статуса заказа',
    description: 'Переводит заказ по этапам: red (Новый) -> brown (В сборке) -> yellow (Готов к выдаче).',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "status": "yellow"\n}',
  },
  {
    id: 'orders-complete',
    method: 'PUT',
    path: '/orders/{id}/complete',
    service: 'OrdersService',
    title: 'Завершение заказа (Финальное списание)',
    description: 'Переводит заказ в green (Завершен), вызывает gRPC PartsService для списания остатков (выставляя quantity = -1 для проданных) и фиксирует сумму в истории продаж.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID заказа' }
    ],
  },
  {
    id: 'orders-delete',
    method: 'DELETE',
    path: '/orders/{id}',
    service: 'OrdersService',
    title: 'Отмена и удаление заказа',
    description: 'Удаляет незавершенный заказ и освобождает зарезервированные складские позиции.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID заказа' }
    ],
  },
  {
    id: 'orders-add-item',
    method: 'POST',
    path: '/orders/{id}/items',
    service: 'OrdersService',
    title: 'Добавление запчасти в существующий заказ',
    description: 'Добавляет позицию в заказ с проверкой наличия на складе через gRPC PartsService.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-update-item',
    method: 'PUT',
    path: '/orders/{id}/items/{itemId}',
    service: 'OrdersService',
    title: 'Изменение количества/цены позиции заказа',
    description: 'Обновляет количество или согласованную цену конкретной детали в заказе.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-delete-item',
    method: 'DELETE',
    path: '/orders/{id}/items/{itemId}',
    service: 'OrdersService',
    title: 'Удаление позиции из состава заказа',
    description: 'Убирает деталь из заказа и пересчитывает общую сумму.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-customers-list',
    method: 'GET',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Клиентская база CRM с LTV',
    description: 'Возвращает список клиентов с количеством заказов, суммарным объемом покупок (LTV), накопительной скидкой и любимой ТК.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'search', in: 'query', type: 'string', required: false, description: 'Поиск по имени, номеру (+7...) или городу' },
      { name: 'category', in: 'query', type: 'string', required: false, description: 'Фильтр: Обычный | Оптовик | Постоянный' }
    ],
    responseExample: '[\n  {\n    "id": 84,\n    "name": "Алексей Смирнов",\n    "phone": "+79528094032",\n    "city": "Томск",\n    "preferred_tk": "Энергия",\n    "category": "Постоянный",\n    "discount_percent": 5,\n    "total_orders": 7,\n    "total_spent": 142500,\n    "last_order_at": "2026-02-28T14:30:00Z"\n  }\n]',
  },
  {
    id: 'orders-customers-create',
    method: 'POST',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Создание карточки клиента',
    description: 'Добавляет клиента с уникальным номером (+7XXXXXXXXXX). Защита от дублей вернет 409 Conflict при попытке создать повторный номер.',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "name": "ИП Кузнецов",\n  "phone": "+79138887766",\n  "city": "Новосибирск",\n  "preferred_tk": "СДЭК",\n  "category": "Оптовик",\n  "discount_percent": 10\n}',
  },
  {
    id: 'orders-customers-update',
    method: 'PUT',
    path: '/orders/customers/{id}',
    service: 'OrdersService',
    title: 'Редактирование карточки клиента',
    description: 'Обновляет категорию, процент персональной скидки, паспортные данные или примечания.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-customers-delete',
    method: 'DELETE',
    path: '/orders/customers/{id}',
    service: 'OrdersService',
    title: 'Удаление карточки клиента',
    description: 'Удаляет клиента из CRM справочника.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-monthly-sales',
    method: 'GET',
    path: '/orders/monthly-sales',
    service: 'OrdersService',
    title: 'Ежемесячная аналитика продаж',
    description: 'Возвращает агрегированные данные о выручке по месяцам на основе успешно завершенных заказов.',
    authRequired: true,
    requiredRole: 'manager',
  },

  // === AUTH SERVICE ===
  {
    id: 'auth-login',
    method: 'POST',
    path: '/auth/login',
    service: 'AuthService',
    title: 'Вход сотрудника по логину и паролю',
    description: 'Аутентифицирует пользователя и устанавливает безопасную сессионную cookie с флагами HttpOnly и SameSite=Lax.',
    authRequired: false,
    requestBodyExample: '{\n  "username": "operator1",\n  "password": "••••••••"\n}',
  },
  {
    id: 'auth-verify',
    method: 'GET',
    path: '/auth/verify',
    service: 'AuthService',
    title: 'Traefik ForwardAuth верификация',
    description: 'Горячий путь: вызывается Traefik Ingress на каждый входящий HTTP запрос к микросервисам. Валидирует сессию из cookie за <1 мс и проставляет заголовки X-User-Id, X-User-Role.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'auth-me',
    method: 'GET',
    path: '/auth/me',
    service: 'AuthService',
    title: 'Профиль авторизованного пользователя',
    description: 'Возвращает текущие данные пользователя, роль, имя и доступные права.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'auth-logout',
    method: 'POST',
    path: '/auth/logout',
    service: 'AuthService',
    title: 'Завершение сессии (Выход)',
    description: 'Очищает сессию в Redis и удаляет сессионные cookies.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'auth-users-list',
    method: 'GET',
    path: '/api/v1/users',
    service: 'AuthService',
    title: 'Список всех пользователей системы',
    description: 'Возвращает сотрудников с ролями (Admin, Manager, Operator).',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'auth-users-create',
    method: 'POST',
    path: '/api/v1/users',
    service: 'AuthService',
    title: 'Создание нового сотрудника',
    description: 'Создает учетную запись с заданной ролью и хэшированным bcrypt паролем.',
    authRequired: true,
    requiredRole: 'admin',
    requestBodyExample: '{\n  "email": "manager@avtoplaneta.ru",\n  "name": "Сергей Васильев",\n  "role": "manager",\n  "password": "TemporaryPassword123"\n}',
  },
  {
    id: 'auth-users-update',
    method: 'PUT',
    path: '/api/v1/users/{id}',
    service: 'AuthService',
    title: 'Редактирование профиля сотрудника',
    description: 'Изменение роли, имени или смена пароля.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'auth-users-delete',
    method: 'DELETE',
    path: '/api/v1/users/{id}',
    service: 'AuthService',
    title: 'Блокировка / удаление сотрудника',
    description: 'Блокирует доступ пользователя в систему и завершает активные сессии.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'auth-activity',
    method: 'GET',
    path: '/api/v1/activity',
    service: 'AuthService',
    title: 'Журнал аудита активности',
    description: 'Аудит-лог действий в системе (кто, когда и какую операцию совершил) для обеспечения информационной безопасности.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'auth-admin-status',
    method: 'GET',
    path: '/api/v1/admin/status',
    service: 'AuthService',
    title: 'Системный статус микросервисов',
    description: 'Healthcheck проверка статуса микросервисов, баз данных и подключений.',
    authRequired: true,
    requiredRole: 'admin',
  },

  // === MESSAGING SERVICE ===
  {
    id: 'messaging-threads',
    method: 'GET',
    path: '/api/messaging/threads',
    service: 'MessagingService',
    title: 'Список диалогов и чатов',
    description: 'Возвращает треды переписки по клиентам, заказам и внутренним диалогам команды.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'messaging-create-thread',
    method: 'POST',
    path: '/api/messaging/threads',
    service: 'MessagingService',
    title: 'Создание нового треда диалога',
    description: 'Инициализирует новый чат по конкретному заказу или вопросу покупателя.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'messaging-get-messages',
    method: 'GET',
    path: '/api/messaging/threads/{id}/messages',
    service: 'MessagingService',
    title: 'История сообщений в треде',
    description: 'Возвращает цепочку сообщений диалога с метаданными и вложениями.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'messaging-send',
    method: 'POST',
    path: '/api/messaging/threads/{id}/messages',
    service: 'MessagingService',
    title: 'Отправка сообщения в диалог',
    description: 'Публикует сообщение в чат и отправляет push/уведомление собеседникам.',
    authRequired: true,
    requiredRole: 'any',
    requestBodyExample: '{\n  "content": "Деталь упакована, трек СДЭК: 1234567890"\n}',
  },
  {
    id: 'messaging-unread',
    method: 'GET',
    path: '/api/messaging/unread',
    service: 'MessagingService',
    title: 'Счетчик непрочитанных сообщений',
    description: 'Быстрый легковесный запрос для бейджей уведомлений в шапке сайта.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'messaging-drom-dialogs',
    method: 'GET',
    path: '/api/messaging/drom/dialogs',
    service: 'MessagingService',
    title: 'Диалоги с площадки Drom.ru',
    description: 'Интеграция с личным кабинетом Дрома: получение входящих сообщений от покупателей запчастей.',
    authRequired: true,
    requiredRole: 'manager',
  },

  // === EXPORT SERVICE ===
  {
    id: 'export-pricelist-xml',
    method: 'GET',
    path: '/uploads/pricelist.xml',
    service: 'ExportService',
    title: 'Готовый XML фид для Drom.ru',
    description: 'Публичный XML каталог для автовыгрузки на Дром. Включает детали в наличии (quantity > 0) и дефектные запчасти (quantity = 0). Проданные детали (quantity = -1) исключаются.',
    authRequired: false,
  },
  {
    id: 'export-drom-stream',
    method: 'GET',
    path: '/api/export/drom.xml',
    service: 'ExportService',
    title: 'Генерация актуального фида Drom.ru',
    description: 'Потоковая генерация каталога через gRPC к PartsService с фильтрацией quantity >= 0.',
    authRequired: false,
  },
  {
    id: 'export-avito-xml',
    method: 'GET',
    path: '/api/export/avito.xml',
    service: 'ExportService',
    title: 'XML фид для Avito Авто',
    description: 'Генерирует фид в стандарте Avito с категориями запчастей, фотографиями и характеристиками.',
    authRequired: false,
  },
  {
    id: 'export-rebuild',
    method: 'POST',
    path: '/api/export/rebuild',
    service: 'ExportService',
    title: 'Принудительная пересборка XML фидов',
    description: 'Сбрасывает кэш отпечатка склада InventoryVersion и заново пересобирает файлы выгрузок.',
    authRequired: true,
    requiredRole: 'admin',
  },
];

// -----------------------------------------------------------------------------
// РУБРИКАТОР СТАТЕЙ БАЗЫ ЗНАНИЙ
// -----------------------------------------------------------------------------
interface GuideArticle {
  id: string;
  category: 'onboarding' | 'warehouse' | 'orders' | 'integrations' | 'mobile';
  title: string;
  icon: any;
  summary: string;
  badge: string;
  content: React.ReactNode;
}

export default function Readme({ user, defaultTab = 'guides' }: ReadmeProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get('tab');
  const urlArticle = searchParams.get('article');

  // Определяем активную вкладку
  const initialTab = useMemo<'guides' | 'api' | 'architecture' | 'roles' | 'faq'>(() => {
    if (rawTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(rawTab)) {
      return rawTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq';
    }
    if (defaultTab === 'operator' || defaultTab === 'manager') {
      return 'guides';
    }
    if (defaultTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(defaultTab)) {
      return defaultTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq';
    }
    return 'guides';
  }, [rawTab, defaultTab]);

  const [activeTab, setActiveTab] = useState<'guides' | 'api' | 'architecture' | 'roles' | 'faq'>(initialTab);

  // Выбранная статья в базе знаний
  const [selectedArticleId, setSelectedArticleId] = useState<string>(
    urlArticle || (defaultTab === 'manager' ? 'orders-lifecycle' : defaultTab === 'operator' ? 'warehouse-cycle' : 'onboarding-intro')
  );

  // Состояние API Explorer
  const [apiSearchQuery, setApiSearchQuery] = useState('');
  const [selectedService, setSelectedService] = useState<string>('ALL');
  const [selectedMethod, setSelectedMethod] = useState<string>('ALL');
  const [expandedEndpoints, setExpandedEndpoints] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Синхронизация с URL
  useEffect(() => {
    if (rawTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(rawTab)) {
      setActiveTab(rawTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq');
    }
  }, [rawTab]);

  const handleTabChange = (tab: 'guides' | 'api' | 'architecture' | 'roles' | 'faq') => {
    setActiveTab(tab);
    setSearchParams({ tab, article: selectedArticleId });
  };

  const handleSelectArticle = (articleId: string) => {
    setSelectedArticleId(articleId);
    setSearchParams({ tab: 'guides', article: articleId });
  };

  const toggleEndpoint = (id: string) => {
    setExpandedEndpoints((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  // Фильтрация API
  const filteredEndpoints = useMemo(() => {
    return FULL_API_ENDPOINTS.filter((item) => {
      if (selectedService !== 'ALL' && item.service !== selectedService) return false;
      if (selectedMethod !== 'ALL' && item.method !== selectedMethod) return false;
      if (apiSearchQuery.trim() !== '') {
        const query = apiSearchQuery.toLowerCase();
        const inPath = item.path.toLowerCase().includes(query);
        const inTitle = item.title.toLowerCase().includes(query);
        const inDesc = item.description.toLowerCase().includes(query);
        const inService = item.service.toLowerCase().includes(query);
        if (!inPath && !inTitle && !inDesc && !inService) return false;
      }
      return true;
    });
  }, [apiSearchQuery, selectedService, selectedMethod]);

  const methodColorClass = (method: string) => {
    switch (method) {
      case 'GET':
        return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
      case 'POST':
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30';
      case 'PUT':
        return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
      case 'DELETE':
        return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
      default:
        return 'bg-slate-500/10 text-slate-600 border-slate-500/30';
    }
  };

  // ---------------------------------------------------------------------------
  // СТАТЬИ БАЗЫ ЗНАНИЙ
  // ---------------------------------------------------------------------------
  const GUIDE_ARTICLES: GuideArticle[] = [
    {
      id: 'onboarding-intro',
      category: 'onboarding',
      title: 'Введение в систему и онбординг сотрудника',
      icon: BookOpen,
      badge: 'Для всех ролей',
      summary: 'Архитектурный обзор рабочих процессов, назначение модулей и базовая терминология.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Добро пожаловать в систему Avtoplaneta!</h3>
            <p className="text-muted-foreground mb-4">
              Avtoplaneta — это комплексная экосистема для автоматизации автомобильного разбора, управления инвентарем
              запчастей, сквозной продажи через веб-каталог и классифайды (Drom.ru, Avito), а также координации логистики.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <Package className="w-4 h-4 text-primary" />
                Складской модуль
              </h4>
              <p className="text-xs text-muted-foreground">
                Учет десятков тысяч запчастей, адресация хранения по стеллажам и полкам, параллельная загрузка фотографий
                и мгновенный поиск с подсказками опечаток через Elasticsearch.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <ShoppingCart className="w-4 h-4 text-emerald-500" />
                Управление продажами
              </h4>
              <p className="text-xs text-muted-foreground">
                Обработка входящих заказов, цветовая шкала готовности, интеграция с транспортными компаниями (СДЭК,
                Энергия, ПЭК) и автоматическое списание деталей со склада.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <Users className="w-4 h-4 text-blue-500" />
                CRM Клиентская база
              </h4>
              <p className="text-xs text-muted-foreground">
                Автоматическое распознавание телефонов в формате <code>+7...</code>, расчет LTV (суммы покупок), история
                обращений и индивидуальные накопительные скидки.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <Radio className="w-4 h-4 text-amber-500" />
                Классифайды Drom и Avito
              </h4>
              <p className="text-xs text-muted-foreground">
                Автоматическая генерация XML-прайсов в реальном времени, потоковая выгрузка без замедления основной базы
                и прием сообщений от покупателей с Дрома.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl border border-primary/20 bg-primary/5">
            <h4 className="font-semibold text-primary mb-2 flex items-center gap-2">
              <KeyRound className="w-4 h-4" />
              Правила учетной записи и безопасность
            </h4>
            <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
              <li>Никогда не передавайте свой пароль третьим лицам.</li>
              <li>Все действия операторов и менеджеров фиксируются в журнале аудита безопасности.</li>
              <li>Сессия защищена современным шлюзом Traefik ForwardAuth с валидацией за время менее 1 миллисекунды.</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: 'warehouse-cycle',
      category: 'warehouse',
      title: 'Складской учет: от приемки до ячейки хранения',
      icon: Package,
      badge: 'Операторам склада',
      summary: 'Регламент создания карточек деталей, обязательные поля, артикулы OEM и адресация ячеек.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Полный цикл работы с деталью на складе</h3>
            <p className="text-muted-foreground mb-4">
              Правильное и аккуратное заполнение карточки — залог того, что деталь мгновенно найдется в поиске и будет
              успешно продана через сайт или классифайды.
            </p>
          </div>

          <div className="space-y-4">
            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Шаг 1: Идентификация и создание карточки</h4>
              <p className="text-xs text-muted-foreground mb-2">
                В разделе <strong>«Инвентарь»</strong> нажмите кнопку <strong>«Добавить запчасть»</strong>.
              </p>
              <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                <li><strong>Наименование:</strong> используйте общепринятые названия («Фара передняя правая», «Бампер задний»).</li>
                <li><strong>Марка / Модель:</strong> выбирайте автодополнением из базы Drom для точной привязки в каталоге.</li>
                <li><strong>Артикул / OEM-код:</strong> обязательно переписывайте оригинальный номер с детали — покупатели ищут по нему в 70% случаев!</li>
                <li><strong>Сторона и положение:</strong> укажите «Левый/Правый», «Передний/Задний», «Верхний/Нижний».</li>
              </ul>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Шаг 2: Адресация хранения (Локация)</h4>
              <p className="text-xs text-muted-foreground">
                Поле <strong>«Локация»</strong> обязательно для заполнения по стандарту:{' '}
                <code className="text-xs bg-muted px-2 py-0.5 rounded font-mono">Склад 1 - Ряд [Буква] - Полка [Номер]</code>.
                Это исключает потери времени при сборке заказа менеджером.
              </p>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Шаг 3: Состояние и дефекты</h4>
              <p className="text-xs text-muted-foreground">
                Если деталь имеет потертости, сколы или дефекты — честно опишите их в поле «Описание дефекта».
                Это предотвращает конфликтные возвраты и отказы клиентов при получении в транспортной компании.
              </p>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'photo-standards',
      category: 'warehouse',
      title: 'Стандарты фотосъемки и оптимизация изображений',
      icon: Camera,
      badge: 'Операторам склада',
      summary: 'Требования к ракурсам, освещению, кадрированию и сжатию в современный формат WebP.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Стандарты съемки автозапчастей</h3>
            <p className="text-muted-foreground mb-4">
              Качественные фотографии увеличивают конверсию в продажу на 300%. Клиенты из других регионов покупают
              только тогда, когда видят реальное состояние детали.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
            <div className="p-4 rounded-xl border border-emerald-500/30 bg-emerald-500/5">
              <h4 className="font-bold text-emerald-600 dark:text-emerald-400 mb-2 flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4" />
                Как правильно делать фото:
              </h4>
              <ul className="space-y-1.5 text-muted-foreground">
                <li>• <strong>Ракурс 1:</strong> Общий вид детали целиком на светлом контрастном фоне.</li>
                <li>• <strong>Ракурс 2:</strong> Обратная сторона и все крепления (кронштейны, уши, разъемы).</li>
                <li>• <strong>Ракурс 3:</strong> Маркировка, заводская наклейка, выбитый номер OEM крупным планом.</li>
                <li>• <strong>Ракурс 4:</strong> Все имеющиеся дефекты (царапины, трещины) с близкого расстояния.</li>
              </ul>
            </div>

            <div className="p-4 rounded-xl border border-rose-500/30 bg-rose-500/5">
              <h4 className="font-bold text-rose-600 dark:text-rose-400 mb-2 flex items-center gap-1.5">
                <AlertTriangle className="w-4 h-4" />
                Чего делать нельзя:
              </h4>
              <ul className="space-y-1.5 text-muted-foreground">
                <li>• Не фотографируйте в темных неосвещенных проходах склада.</li>
                <li>• Не обрезайте края детали при съемке (деталь должна быть в кадре целиком).</li>
                <li>• Не закрывайте маркировку и заводские штампы пальцами.</li>
                <li>• Не используйте чужие фото из интернета — только реальные снимки!</li>
              </ul>
            </div>
          </div>

          <div className="p-4 rounded-xl border border-border bg-card">
            <h4 className="font-semibold text-foreground mb-2">Автоматическая оптимизация в системе</h4>
            <p className="text-xs text-muted-foreground">
              Платформа автоматически сжимает загруженные снимки в легковесный формат <strong>WebP</strong> с сохранением
              высокой четкости. Это позволяет клиентам на мобильных устройствах моментально просматривать фото даже при
              медленном мобильном интернете.
            </p>
          </div>
        </div>
      ),
    },
    {
      id: 'defect-reports',
      category: 'warehouse',
      title: 'Дефектовка автомобилей и правило нулевого остатка (quantity = 0)',
      icon: FileSpreadsheet,
      badge: 'Фундаментальное правило',
      summary: 'Почему дефектные детали получают quantity = 0, как работает импорт из Excel и почему они видны в XML.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Дефектовочные ведомости и разбор автомобилей</h3>
            <p className="text-muted-foreground mb-4">
              При разборе донорского автомобиля или оприходовании поврежденных деталей действует ключевое системное правило,
              которое отличает Avtoplaneta от стандартных складских программ.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-amber-500/30 bg-amber-500/10 space-y-3">
            <div className="font-bold text-amber-700 dark:text-amber-400 flex items-center gap-2">
              <Boxes className="w-5 h-5" />
              Золотое правило остатка (quantity = 0)
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed">
              В стандартных магазинах товар с нулевым количеством считается отсутствующим и скрывается. В Avtoplaneta
              детали с дефектовки получают количество ровно <strong>0</strong>. Они <strong>ОБЯЗАНЫ ОСТАВАТЬСЯ ВИДИМЫМИ</strong>{' '}
              в каталоге сайта и в выгрузках на Drom.ru и Avito!
            </p>
            <div className="text-xs font-mono bg-background/50 p-2.5 rounded border border-amber-500/20">
              Поиск деталей: WHERE quantity &gt;= 0 AND deleted_at IS NULL
            </div>
            <p className="text-xs text-muted-foreground">
              Клиенты авторазборов очень часто ищут поврежденные детали (например, фару с треснувшим стеклом под переклейку
              или бампер под локальную пайку). Скрывать такие позиции нельзя!
            </p>
          </div>

          <div className="space-y-3">
            <h4 className="font-semibold text-base">Импорт ведомостей из Excel:</h4>
            <ol className="list-decimal list-inside space-y-1.5 text-xs text-muted-foreground">
              <li>Перейдите в раздел <strong>«Дефектовка»</strong> и выберите приходный файл ведомости.</li>
              <li>Система запустит предпросмотр позиций, валидируя названия, марки и примененные скидки за дефект.</li>
              <li>При подтверждении система генерирует до 1 400+ запчастей за один синхронный SQL-пакет (менее 50 мс).</li>
            </ol>
          </div>
        </div>
      ),
    },
    {
      id: 'orders-lifecycle',
      category: 'orders',
      title: 'Жизненный цикл заказа и автоматическое списание (quantity = -1)',
      icon: ShoppingCart,
      badge: 'Менеджерам по продажам',
      summary: 'Цветовая шкала статусов заказов, списание проданных деталей и работа с транспортными компаниями.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Управление заказами и отгрузкой</h3>
            <p className="text-muted-foreground mb-4">
              Заказы проходят строгий четырехэтапный маршрут. Система визуально подсвечивает состояние каждого заказа
              цветовым маркером.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
            <div className="p-3.5 rounded-xl border border-rose-500/30 bg-rose-500/10">
              <div className="font-bold text-rose-600 dark:text-rose-400 mb-1 flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-rose-500 animate-pulse"></span>
                Красный (Новый заказ)
              </div>
              <p className="text-muted-foreground">
                Поступил с сайта или мобильного приложения. Менеджер проверяет совместимость, звонит клиенту, уточняет
                город доставки и выставляет реквизиты на оплату.
              </p>
            </div>

            <div className="p-3.5 rounded-xl border border-amber-600/30 bg-amber-600/10">
              <div className="font-bold text-amber-700 dark:text-amber-400 mb-1 flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-amber-600"></span>
                Коричневый (В сборке)
              </div>
              <p className="text-muted-foreground">
                Оплата подтверждена. Заказ передается на склад. Операторы снимают деталь с полки и надежно упаковывают в
                пупырчатую пленку и картон.
              </p>
            </div>

            <div className="p-3.5 rounded-xl border border-yellow-500/30 bg-yellow-500/10">
              <div className="font-bold text-yellow-600 dark:text-yellow-400 mb-1 flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-yellow-500"></span>
                Желтый (Готов к отправке)
              </div>
              <p className="text-muted-foreground">
                Посылка полностью упакована, оформлена транспортная накладная (СДЭК, Энергия, ПЭК), ожидается приезд курьера
                ТК или клиент идет на склад за самовывозом.
              </p>
            </div>

            <div className="p-3.5 rounded-xl border border-emerald-500/30 bg-emerald-500/10">
              <div className="font-bold text-emerald-600 dark:text-emerald-400 mb-1 flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                Зеленый (Выполнен & Списан)
              </div>
              <p className="text-muted-foreground">
                Заказ выдан! Нажатие кнопки «Завершить заказ» списывает остаток в базе через gRPC. Детали присваивается маркер{' '}
                <strong>quantity = -1</strong>, скрывающий её из витрины.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl border border-rose-500/20 bg-rose-500/5">
            <h4 className="font-bold text-rose-600 dark:text-rose-400 mb-1">
              Почему при продаже ставится quantity = -1, а не 0?
            </h4>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Так как <code>quantity = 0</code> зарезервирован для видимых дефектных позиций, проданные позиции помечаются{' '}
              <strong>-1</strong>. Это позволяет сохранить историю всех продаж в заказах без физического удаления строк из
              базы данных, при этом деталь надежно скрывается от покупателей на сайте и в фидах.
            </p>
          </div>
        </div>
      ),
    },
    {
      id: 'crm-customers',
      category: 'orders',
      title: 'CRM и Клиентская база: дедупликация и накопительные скидки',
      icon: Users,
      badge: 'Менеджерам по продажам',
      summary: 'Нормализация номеров +7..., автоматическое распознавание имен, защита от дублей и аналитика LTV.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Встроенная CRM-система клиентов</h3>
            <p className="text-muted-foreground mb-4">
              Вся история взаимодействий с клиентами централизована в модуле <strong>«Клиенты»</strong>. Система автоматически
              ведет учет объема покупок (LTV) и исключает дублирование записей.
            </p>
          </div>

          <div className="space-y-4">
            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Умный парсер контактов (parseBuyerInfo)</h4>
              <p className="text-xs text-muted-foreground mb-2">
                При создании заказа с мобильного приложения или сайта покупатель часто передается в виде сырой строки:{' '}
                <em>«Получатель Кадиров Мадамин Анвархонович 9635102022»</em>.
              </p>
              <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                <li>Система автоматически отделяет телефон от имени.</li>
                <li>Номер приводится к международному формату: <code className="text-xs">+7 (963) 510-20-22</code>.</li>
                <li>Служебный префикс «Получатель» автоматически очищается.</li>
                <li>Если клиент с таким номером уже существует — заказ автоматически привязывается к нему!</li>
              </ul>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Категории клиентов и персональные скидки</h4>
              <p className="text-xs text-muted-foreground mb-2">
                Вы можете классифицировать клиентов по категориям:
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-xs">
                <div className="p-2.5 rounded-lg border border-border bg-card">
                  <div className="font-semibold mb-0.5">👤 Обычный</div>
                  <div className="text-muted-foreground">Розничные клиенты, стандартные розничные цены.</div>
                </div>
                <div className="p-2.5 rounded-lg border border-border bg-card">
                  <div className="font-semibold mb-0.5">⭐ Постоянный</div>
                  <div className="text-muted-foreground">Скидка 3–7%, приоритетная отгрузка.</div>
                </div>
                <div className="p-2.5 rounded-lg border border-border bg-card">
                  <div className="font-semibold mb-0.5">💼 Оптовик / СТО</div>
                  <div className="text-muted-foreground">Скидка от 10%, работа по безналичному расчету.</div>
                </div>
              </div>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Аналитика покупок (LTV)</h4>
              <p className="text-xs text-muted-foreground">
                В карточке клиента в реальном времени рассчитывается суммарная выручка по всем его завершенным заказам и
                дата последнего обращения. Это позволяет сразу видеть ключевых клиентов при входящем звонке.
              </p>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'drom-avito-export',
      category: 'integrations',
      title: 'Интеграция с Drom.ru и Avito: автогенерация XML фидов',
      icon: Radio,
      badge: 'Интеграции и выгрузки',
      summary: 'Формирование XML прайс-листов, кэширование отпечатком InventoryVersion и защита персональных данных.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Автоматическая синхронизация с классифайдами</h3>
            <p className="text-muted-foreground mb-4">
              Каталог запчастей Avtoplaneta автоматически транслируется на крупнейшие автомобильные порталы России:
              <strong>Drom.ru</strong> и <strong>Avito Авто</strong>.
            </p>
          </div>

          <div className="space-y-4">
            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <FileCode className="w-4 h-4 text-primary" />
                Публичные адреса XML фидов:
              </h4>
              <div className="space-y-2 text-xs font-mono">
                <div className="p-2 bg-muted rounded flex items-center justify-between">
                  <span>https://avtoplaneta.ru/uploads/pricelist.xml</span>
                  <span className="text-[11px] font-sans text-muted-foreground">Прайс для Drom.ru</span>
                </div>
                <div className="p-2 bg-muted rounded flex items-center justify-between">
                  <span>https://avtoplaneta.ru/api/export/avito.xml</span>
                  <span className="text-[11px] font-sans text-muted-foreground">Фид для Avito Авто</span>
                </div>
              </div>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Умное кэширование (InventoryVersion)</h4>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Генерация XML по каталогу из 50 000 позиций — тяжелая операция. Чтобы не нагружать базу данных при каждом
                обращении робота Drom, сервис сверяет отпечаток склада <code>InventoryVersion</code> (хэш времени последнего
                изменения склада и числа живых деталей). Если изменений не было, готовый XML отдается моментально из кэша!
              </p>
            </div>

            <div className="border-l-2 border-primary pl-4">
              <h4 className="font-semibold text-base mb-1">Изоляция персональных данных сотрудников</h4>
              <p className="text-xs text-muted-foreground leading-relaxed">
                Сервис склада не знает ни про ИНН сотрудников, ни про банковские реквизиты. Экспорт-сервис запрашивает
                только необходимые реквизиты продавцов через строго типизированный внутренний gRPC <code>auth.GetUsers</code>.
              </p>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'mobile-app-flutter',
      category: 'mobile',
      title: 'Мобильное приложение (Flutter) для сотрудников склада',
      icon: Smartphone,
      badge: 'Мобильное приложение',
      summary: 'Установка APK, быстрая ревизия на складе, сканирование и привязка фото прямо с телефона.',
      content: (
        <div className="space-y-6 text-sm leading-relaxed">
          <div>
            <h3 className="text-xl font-bold mb-3">Работа через мобильное приложение Avtoplaneta</h3>
            <p className="text-muted-foreground mb-4">
              Специально для операторов склада разработано нативное мобильное приложение на Flutter (Android / iOS),
              позволяющее работать непосредственно у стеллажей с запчастями без необходимости носить ноутбук.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold mb-1 flex items-center gap-1.5">
                <Camera className="w-4 h-4 text-primary" />
                Съемка камерой смартфона
              </h4>
              <p className="text-muted-foreground">
                Фотографируйте снятую деталь прямо у капота автомобиля и прикрепляйте фото к карточке в одно касание.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold mb-1 flex items-center gap-1.5">
                <Search className="w-4 h-4 text-emerald-500" />
                Быстрая ревизия полок
              </h4>
              <p className="text-muted-foreground">
                Находите нужную запчасть в ряду по артикулу, сверяйте остаток и мгновенно корректируйте локацию.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-card">
              <h4 className="font-semibold mb-1 flex items-center gap-1.5">
                <Truck className="w-4 h-4 text-blue-500" />
                Комплектация заказов
              </h4>
              <p className="text-muted-foreground">
                Просматривайте список деталей к сборке в заказе и отмечайте укомплектованные позиции.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl border border-border bg-muted/30">
            <h4 className="font-semibold text-foreground mb-1">Как установить приложение:</h4>
            <ol className="list-decimal list-inside space-y-1 text-xs text-muted-foreground">
              <li>Откройте раздел релизов в репозитории проекта и скачайте актуальный файл <code>app-release.apk</code>.</li>
              <li>Разрешите установку из доверенного источника на Android устройстве.</li>
              <li>Введите ваш рабочий логин и пароль оператора или менеджера.</li>
            </ol>
          </div>
        </div>
      ),
    },
  ];

  const currentArticle = useMemo(() => {
    return GUIDE_ARTICLES.find((a) => a.id === selectedArticleId) || GUIDE_ARTICLES[0];
  }, [selectedArticleId]);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      {/* Шапка портала документации */}
      <div className="mb-8 border-b border-border pb-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary mb-3">
              <Sparkles className="w-3.5 h-3.5" />
              Корпоративная база знаний & API v2.5
            </div>
            <h1 className="text-3xl font-extrabold tracking-tight">Центр документации Avtoplaneta</h1>
            <p className="text-muted-foreground mt-1 text-sm md:text-base">
              Единый портал регламентов, интерактивная спецификация 45+ API эндпоинтов и архитектура микросервисов.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <a
              href="/api.swagger.json"
              download="api.swagger.json"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card hover:bg-accent text-sm font-medium transition-colors shadow-sm"
              title="Скачать спецификацию OpenAPI / Swagger JSON"
            >
              <Download className="w-4 h-4 text-primary" />
              <span>Swagger JSON</span>
            </a>

            {user && (
              <div className="hidden sm:flex items-center gap-2 px-3 py-2 rounded-lg bg-muted/60 text-xs text-muted-foreground">
                <span>Роль:</span>
                <span className="font-semibold text-foreground capitalize">
                  {user.role === 'admin' ? 'Администратор' : user.role === 'manager' ? 'Менеджер' : 'Оператор'}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Главные навигационные разделы */}
        <div className="flex items-center gap-1 sm:gap-2 mt-6 overflow-x-auto pb-1 scrollbar-none">
          <button
            onClick={() => handleTabChange('guides')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'guides'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <BookOpen className="w-4 h-4" />
            <span>База знаний & Регламенты</span>
            <span className="ml-1 text-xs px-1.5 py-0.2 rounded-full bg-background/20 font-mono">
              {GUIDE_ARTICLES.length}
            </span>
          </button>

          <button
            onClick={() => handleTabChange('api')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'api'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <Code2 className="w-4 h-4" />
            <span>Спецификация API (Swagger)</span>
            <span className="ml-1 text-xs px-1.5 py-0.2 rounded-full bg-background/20 font-mono">
              {FULL_API_ENDPOINTS.length}
            </span>
          </button>

          <button
            onClick={() => handleTabChange('architecture')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'architecture'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <Server className="w-4 h-4" />
            <span>Архитектура & Микросервисы</span>
          </button>

          <button
            onClick={() => handleTabChange('roles')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'roles'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <Shield className="w-4 h-4" />
            <span>Роли и безопасность (RBAC)</span>
          </button>

          <button
            onClick={() => handleTabChange('faq')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'faq'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <HelpCircle className="w-4 h-4" />
            <span>FAQ и Решение проблем</span>
          </button>
        </div>
      </div>

      {/* ========================================================================= */}
      {/* РАЗДЕЛ 1: БАЗА ЗНАНИЙ И РЕГЛАМЕНТЫ (ДВУХКОЛОНОЧНЫЙ WIKI)                  */}
      {/* ========================================================================= */}
      {activeTab === 'guides' && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Левая колонка: список статей */}
          <div className="lg:col-span-4 space-y-2 bg-card border border-border rounded-xl p-3 shadow-sm">
            <div className="px-3 py-2 text-xs font-bold text-muted-foreground uppercase tracking-wider">
              Содержание базы знаний
            </div>
            <div className="space-y-1">
              {GUIDE_ARTICLES.map((article) => {
                const IconComponent = article.icon;
                const isSelected = article.id === selectedArticleId;
                return (
                  <button
                    key={article.id}
                    onClick={() => handleSelectArticle(article.id)}
                    className={`w-full text-left p-3 rounded-lg transition-all flex items-start gap-3 select-none ${
                      isSelected
                        ? 'bg-primary/10 border-l-4 border-primary text-foreground'
                        : 'hover:bg-muted/60 text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    <IconComponent
                      className={`w-4 h-4 mt-0.5 shrink-0 ${isSelected ? 'text-primary' : 'text-muted-foreground'}`}
                    />
                    <div className="min-w-0">
                      <div className="text-sm font-semibold line-clamp-1">{article.title}</div>
                      <div className="text-xs text-muted-foreground line-clamp-1 mt-0.5">{article.summary}</div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Правая колонка: просмотр выбранной статьи */}
          <div className="lg:col-span-8 border border-border rounded-xl bg-card p-6 md:p-8 shadow-sm">
            <div className="flex items-center gap-2 mb-3">
              <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary">
                {currentArticle.badge}
              </span>
            </div>
            <h2 className="text-2xl md:text-3xl font-extrabold tracking-tight mb-4">{currentArticle.title}</h2>
            <div className="border-t border-border pt-6">{currentArticle.content}</div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* РАЗДЕЛ 2: СПЕЦИФИКАЦИЯ API (SWAGGER / 45+ ENDPOINTS)                      */}
      {/* ========================================================================= */}
      {activeTab === 'api' && (
        <div className="space-y-6">
          {/* Панель фильтров */}
          <div className="flex flex-col md:flex-row gap-4 items-stretch md:items-center justify-between bg-card border border-border rounded-xl p-4 shadow-sm">
            <div className="relative flex-1">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Поиск по пути (/api/v1/..., /orders...), названию, методу или тегам..."
                value={apiSearchQuery}
                onChange={(e) => setApiSearchQuery(e.target.value)}
                className="w-full pl-9 pr-4 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <select
                value={selectedService}
                onChange={(e) => setSelectedService(e.target.value)}
                className="px-3 py-2 bg-background border border-border rounded-lg text-xs font-medium focus:outline-none focus:ring-2 focus:ring-primary/40"
              >
                <option value="ALL">Все сервисы (Все {FULL_API_ENDPOINTS.length})</option>
                <option value="PartsService">Parts Service (Каталог и склад)</option>
                <option value="OrdersService">Orders Service (Заказы и CRM)</option>
                <option value="AuthService">Auth Service (Аутентификация)</option>
                <option value="MessagingService">Messaging Service (Чаты)</option>
                <option value="ExportService">Export Service (Выгрузки XML)</option>
              </select>

              <select
                value={selectedMethod}
                onChange={(e) => setSelectedMethod(e.target.value)}
                className="px-3 py-2 bg-background border border-border rounded-lg text-xs font-medium focus:outline-none focus:ring-2 focus:ring-primary/40"
              >
                <option value="ALL">Все методы</option>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
              </select>
            </div>
          </div>

          {/* Список карточек эндпоинтов */}
          <div className="space-y-3">
            {filteredEndpoints.length === 0 ? (
              <div className="text-center py-12 border border-dashed border-border rounded-xl">
                <Search className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                <p className="text-muted-foreground text-sm">По заданному запросу эндпоинты не найдены.</p>
              </div>
            ) : (
              filteredEndpoints.map((ep) => {
                const isExpanded = !!expandedEndpoints[ep.id];
                return (
                  <div
                    key={ep.id}
                    className="border border-border rounded-xl bg-card overflow-hidden transition-all hover:border-primary/40 shadow-sm"
                  >
                    <div
                      onClick={() => toggleEndpoint(ep.id)}
                      className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 cursor-pointer select-none hover:bg-muted/30"
                    >
                      <div className="flex items-center gap-3 flex-wrap">
                        <span
                          className={`px-2.5 py-1 rounded-md text-xs font-mono font-bold border ${methodColorClass(
                            ep.method
                          )}`}
                        >
                          {ep.method}
                        </span>
                        <code className="text-sm font-semibold text-foreground break-all">{ep.path}</code>
                        <span className="text-xs text-muted-foreground px-2 py-0.5 rounded bg-muted">
                          {ep.service}
                        </span>
                      </div>

                      <div className="flex items-center gap-3 justify-between sm:justify-end">
                        <span className="text-xs text-muted-foreground line-clamp-1">{ep.title}</span>
                        {isExpanded ? (
                          <ChevronDown className="w-4 h-4 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="w-4 h-4 text-muted-foreground" />
                        )}
                      </div>
                    </div>

                    {isExpanded && (
                      <div className="border-t border-border px-4 py-5 bg-muted/10 space-y-4 text-sm">
                        <div>
                          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                            Назначение и сценарии
                          </h4>
                          <p className="text-foreground leading-relaxed">{ep.description}</p>
                        </div>

                        {ep.parameters && ep.parameters.length > 0 && (
                          <div>
                            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                              Параметры запроса
                            </h4>
                            <div className="overflow-x-auto border border-border rounded-lg bg-card">
                              <table className="w-full text-xs">
                                <thead>
                                  <tr className="bg-muted/60 text-muted-foreground border-b border-border">
                                    <th className="py-2 px-3 text-left">Имя</th>
                                    <th className="py-2 px-3 text-left">Тип</th>
                                    <th className="py-2 px-3 text-left">Расположение</th>
                                    <th className="py-2 px-3 text-left">Обязателен</th>
                                    <th className="py-2 px-3 text-left">Описание</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {ep.parameters.map((p, idx) => (
                                    <tr key={idx} className="border-b border-border/50 last:border-0">
                                      <td className="py-2 px-3 font-mono font-semibold">{p.name}</td>
                                      <td className="py-2 px-3 text-muted-foreground">{p.type}</td>
                                      <td className="py-2 px-3 font-mono text-xs">{p.in}</td>
                                      <td className="py-2 px-3">
                                        {p.required ? (
                                          <span className="text-rose-500 font-medium">Да</span>
                                        ) : (
                                          <span className="text-muted-foreground">Нет</span>
                                        )}
                                      </td>
                                      <td className="py-2 px-3 text-muted-foreground">{p.description}</td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            </div>
                          </div>
                        )}

                        {ep.requestBodyExample && (
                          <div>
                            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                              Пример тела запроса (JSON Body)
                            </h4>
                            <pre className="p-3 bg-muted rounded-lg font-mono text-xs overflow-x-auto text-foreground">
                              {ep.requestBodyExample}
                            </pre>
                          </div>
                        )}

                        {ep.responseExample && (
                          <div>
                            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                              Пример ответа сервера (200 OK)
                            </h4>
                            <pre className="p-3 bg-muted rounded-lg font-mono text-xs overflow-x-auto text-foreground">
                              {ep.responseExample}
                            </pre>
                          </div>
                        )}

                        <div className="flex flex-wrap items-center gap-2 pt-2">
                          <button
                            onClick={() => copyToClipboard(ep.path, ep.id + '-url')}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-card hover:bg-accent text-xs font-medium transition-colors"
                          >
                            {copiedId === ep.id + '-url' ? (
                              <Check className="w-3.5 h-3.5 text-emerald-500" />
                            ) : (
                              <Copy className="w-3.5 h-3.5" />
                            )}
                            <span>{copiedId === ep.id + '-url' ? 'Путь скопирован!' : 'Копировать путь'}</span>
                          </button>

                          <button
                            onClick={() => {
                              const curlCmd = `curl -X ${ep.method} "https://avtoplaneta.ru${ep.path.replace(
                                /\{id\}/g,
                                '1'
                              )}" -H "Accept: application/json"`;
                              copyToClipboard(curlCmd, ep.id + '-curl');
                            }}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-card hover:bg-accent text-xs font-medium transition-colors"
                          >
                            {copiedId === ep.id + '-curl' ? (
                              <Check className="w-3.5 h-3.5 text-emerald-500" />
                            ) : (
                              <Terminal className="w-3.5 h-3.5" />
                            )}
                            <span>{copiedId === ep.id + '-curl' ? 'cURL скопирован!' : 'Копировать cURL'}</span>
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* РАЗДЕЛ 3: АРХИТЕКТУРА И МИКРОСЕРВИСЫ                                      */}
      {/* ========================================================================= */}
      {activeTab === 'architecture' && (
        <div className="space-y-8">
          {/* Топология */}
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-3 flex items-center gap-2">
              <Workflow className="w-5 h-5 text-primary" />
              Топология и межсервисные каналы связи
            </h2>
            <p className="text-sm text-muted-foreground mb-6 leading-relaxed">
              Архитектура построена на прямом проксировании внешнего трафика через Traefik Ingress. Проверка сессий
              происходит на лету через Traefik ForwardAuth (<code className="text-xs font-mono">/auth/verify</code> за &lt;1 мс).
              100% межсервисных вызовов осуществляются строго по <strong>gRPC (Protobuf v3)</strong>, а события передаются
              асинхронно через шину <strong>Redis Streams</strong>.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-5 gap-3 text-center text-xs">
              <div className="p-4 rounded-xl border border-primary/30 bg-primary/5 flex flex-col justify-between">
                <div>
                  <div className="font-bold text-primary text-sm mb-1">Traefik Ingress</div>
                  <div className="text-muted-foreground">SSL / Маршрутизация / ForwardAuth</div>
                </div>
                <div className="mt-3 py-1 px-2 rounded bg-primary/10 font-mono text-[11px] text-primary">
                  Порты :80 / :443
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-card flex flex-col justify-between">
                <div>
                  <div className="font-bold text-foreground text-sm mb-1">Auth Service</div>
                  <div className="text-muted-foreground">Сессии, пользователи, аудит</div>
                </div>
                <div className="mt-3 py-1 px-2 rounded bg-muted font-mono text-[11px]">
                  HTTP :8083 | gRPC :9083
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-card flex flex-col justify-between">
                <div>
                  <div className="font-bold text-foreground text-sm mb-1">Parts Service</div>
                  <div className="text-muted-foreground">Инвентарь, дефектовки, склад</div>
                </div>
                <div className="mt-3 py-1 px-2 rounded bg-muted font-mono text-[11px]">
                  HTTP :8081 | gRPC :9081
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-card flex flex-col justify-between">
                <div>
                  <div className="font-bold text-foreground text-sm mb-1">Orders Service</div>
                  <div className="text-muted-foreground">Заказы, клиенты, списание</div>
                </div>
                <div className="mt-3 py-1 px-2 rounded bg-muted font-mono text-[11px]">
                  HTTP :8082 | gRPC :9082
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-card flex flex-col justify-between">
                <div>
                  <div className="font-bold text-foreground text-sm mb-1">Export Service</div>
                  <div className="text-muted-foreground">XML Drom.ru & Avito фиды</div>
                </div>
                <div className="mt-3 py-1 px-2 rounded bg-muted font-mono text-[11px]">
                  HTTP :8085 (Cron фоновые задачи)
                </div>
              </div>
            </div>
          </div>

          {/* Таблица gRPC RPC контрактов */}
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h3 className="text-lg font-bold mb-3 flex items-center gap-2">
              <Layers className="w-5 h-5 text-primary" />
              Межсервисные вызовы по gRPC (100% коммуникация)
            </h3>
            <div className="overflow-x-auto border border-border rounded-lg">
              <table className="w-full text-xs">
                <thead>
                  <tr className="bg-muted/60 text-muted-foreground border-b border-border">
                    <th className="py-2.5 px-3 text-left">Инициатор</th>
                    <th className="py-2.5 px-3 text-left">Целевой сервис</th>
                    <th className="py-2.5 px-3 text-left font-mono">RPC метод</th>
                    <th className="py-2.5 px-3 text-left">Назначение</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="py-2.5 px-3 font-semibold">Orders Service</td>
                    <td className="py-2.5 px-3">Parts Service</td>
                    <td className="py-2.5 px-3 font-mono text-primary">ChangePartQuantity</td>
                    <td className="py-2.5 px-3 text-muted-foreground">
                      Идемпотентное списание остатков при завершении заказа (с фиксацией в part_stock_operations)
                    </td>
                  </tr>
                  <tr>
                    <td className="py-2.5 px-3 font-semibold">Orders Service</td>
                    <td className="py-2.5 px-3">Parts Service</td>
                    <td className="py-2.5 px-3 font-mono text-primary">GetPart</td>
                    <td className="py-2.5 px-3 text-muted-foreground">
                      Проверка актуальной цены и наличия перед созданием заказа
                    </td>
                  </tr>
                  <tr>
                    <td className="py-2.5 px-3 font-semibold">Parts Service</td>
                    <td className="py-2.5 px-3">Auth Service</td>
                    <td className="py-2.5 px-3 font-mono text-primary">LogActivity</td>
                    <td className="py-2.5 px-3 text-muted-foreground">
                      Фиксация аудита действий операторов в централизованный лог безопасности
                    </td>
                  </tr>
                  <tr>
                    <td className="py-2.5 px-3 font-semibold">Export Service</td>
                    <td className="py-2.5 px-3">Parts Service</td>
                    <td className="py-2.5 px-3 font-mono text-primary">ListPartsForExport</td>
                    <td className="py-2.5 px-3 text-muted-foreground">
                      Потоковая выгрузка каталога (quantity &gt;= 0) для сборки прайс-листов Drom и Avito
                    </td>
                  </tr>
                  <tr>
                    <td className="py-2.5 px-3 font-semibold">Export Service</td>
                    <td className="py-2.5 px-3">Parts Service</td>
                    <td className="py-2.5 px-3 font-mono text-primary">InventoryVersion</td>
                    <td className="py-2.5 px-3 text-muted-foreground">
                      Получение отпечатка склада для проверки необходимости пересборки XML
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* РАЗДЕЛ 4: РОЛИ И БЕЗОПАСНОСТЬ (RBAC МАТРИЦА)                              */}
      {/* ========================================================================= */}
      {activeTab === 'roles' && (
        <div className="space-y-6">
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-2">Матрица прав доступа сотрудников (RBAC)</h2>
            <p className="text-sm text-muted-foreground mb-6">
              Контроль доступа реализован на двух уровнях: Traefik ForwardAuth блокирует несанкционированные URL на входе,
              а внутри микросервисов права дополнительно проверяются по ролевому контексту.
            </p>

            <div className="overflow-x-auto border border-border rounded-xl">
              <table className="w-full text-xs md:text-sm">
                <thead>
                  <tr className="bg-muted/60 text-muted-foreground border-b border-border">
                    <th className="py-3 px-4 text-left font-semibold">Функциональный блок / Операция</th>
                    <th className="py-3 px-4 text-center font-semibold">Оператор</th>
                    <th className="py-3 px-4 text-center font-semibold">Менеджер</th>
                    <th className="py-3 px-4 text-center font-semibold">Администратор</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="py-3 px-4 font-medium">Просмотр и поиск по каталогу деталей</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Создание запчастей, загрузка фотографий</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Создание дефектных ведомостей (quantity = 0)</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Удаление запчастей из каталога (Soft Delete)</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Создание, редактирование и списание заказов</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Клиентская база CRM и назначение скидок</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Управление сотрудниками (создание, смена ролей)</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Просмотр аудита действий и логов безопасности</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* РАЗДЕЛ 5: FAQ И РЕШЕНИЕ ПРОБЛЕМ                                           */}
      {/* ========================================================================= */}
      {activeTab === 'faq' && (
        <div className="space-y-4">
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-4 flex items-center gap-2">
              <HelpCircle className="w-5 h-5 text-primary" />
              Часто задаваемые вопросы и сценарии решения сбоев
            </h2>

            <div className="space-y-4 text-sm">
              <div className="border-b border-border pb-4">
                <h4 className="font-semibold text-foreground mb-1">
                  1. Клиент вернул деталь. Как вернуть её обратно на склад в каталог?
                </h4>
                <p className="text-muted-foreground text-xs leading-relaxed">
                  Если заказ уже был переведен в статус «Выполнен» (зеленый), а деталь получила статус <code>quantity = -1</code>,
                  достаточно открыть карточку детали и нажать «Увеличить количество» (либо внести приход). Деталь снова станет
                  доступна для поиска и попадет в XML-фиды.
                </p>
              </div>

              <div className="border-b border-border pb-4">
                <h4 className="font-semibold text-foreground mb-1">
                  2. Почему деталь из дефектной ведомости видна на Дроме, хотя её количество 0?
                </h4>
                <p className="text-muted-foreground text-xs leading-relaxed">
                  Это нормальное поведение системы. В соответствии с архитектурным правилом Avtoplaneta, детали с{' '}
                  <code>quantity = 0</code> считаются доступными дефектными позициями. Они выгружаются с подробным описанием
                  дефекта. Скрываются только проданные детали со статусом <code>quantity = -1</code>.
                </p>
              </div>

              <div className="border-b border-border pb-4">
                <h4 className="font-semibold text-foreground mb-1">
                  3. Почему при создании клиента возвращается ошибка «409 Conflict»?
                </h4>
                <p className="text-muted-foreground text-xs leading-relaxed">
                  В базе действует ограничение уникальности по номеру телефона. Если такой номер уже зарегистрирован в CRM,
                  повторное создание отклоняется для защиты от дублирования карточек. Используйте поиск по номеру телефона.
                </p>
              </div>

              <div>
                <h4 className="font-semibold text-foreground mb-1">
                  4. Что делать, если на Дроме отображаются старые цены на запчасти?
                </h4>
                <p className="text-muted-foreground text-xs leading-relaxed">
                  Файл выгрузки обновляется автоматически при изменении отпечатка склада. Администратор может зайти в Админ-панель
                  или вызвать эндпоинт <code>POST /api/export/rebuild</code> для принудительной мгновенной пересборки прайс-листа.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
