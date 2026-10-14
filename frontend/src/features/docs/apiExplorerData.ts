/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import swaggerSpec from '../../data/swagger.json';
export { swaggerSpec };

export interface ApiEndpoint {
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

export const supplementaryRoutes: ApiEndpoint[] = [
  // Orders
  {
    id: 'orders-list-native',
    method: 'GET',
    path: '/orders',
    service: 'OrdersService',
    title: 'Список активных заказов',
    description: 'Возвращает текущие заказы с цветовыми статусами (красный/коричневый/желтый), данными покупателя и позициями.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-create-native',
    method: 'POST',
    path: '/orders',
    service: 'OrdersService',
    title: 'Создание заказа покупателя',
    description: 'Создает заказ, автоматически привязывает клиента по номеру телефона, резервирует детали и публикует событие в Redis Streams.',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "buyer_number": "+79528094032 Иванов Иван",\n  "transport_company": "СДЭК",\n  "notes": "Отправить до пятницы",\n  "items": [{ "part_id": 1042, "quantity": 1, "price": 28000 }]\n}',
  },
  {
    id: 'orders-complete-native',
    method: 'PUT',
    path: '/orders/{id}/complete',
    service: 'OrdersService',
    title: 'Завершение заказа со списанием',
    description: 'Переводит заказ в green (Завершен), вызывает gRPC PartsService для списания остатков (выставляя quantity = -1 для полностью проданных).',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-customers-list-native',
    method: 'GET',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Клиентская база CRM с LTV',
    description: 'Возвращает список клиентов с количеством заказов, суммарным LTV покупок, накопительной скидкой и любимой ТК.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-customers-create-native',
    method: 'POST',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Создание карточки клиента',
    description: 'Добавляет клиента с уникальным номером телефона (+7XXXXXXXXXX). Защита от дублей вернет 409 Conflict при повторе.',
    authRequired: true,
    requiredRole: 'manager',
  },
  // Messaging
  {
    id: 'messaging-conversations',
    method: 'GET',
    path: '/api/messaging/conversations',
    service: 'MessagingService',
    title: 'Список диалогов и чатов',
    description: 'Возвращает цепочки переписки команды, диалоги по заказам и обращениям клиентов.',
    authRequired: true,
  },
  {
    id: 'messaging-send-message',
    method: 'POST',
    path: '/api/messaging/conversations/{id}/messages',
    service: 'MessagingService',
    title: 'Отправка сообщения в диалог',
    description: 'Публикует сообщение в тред и рассылает real-time push через WebSocket.',
    authRequired: true,
    requestBodyExample: '{\n  "content": "Деталь упакована, трек-номер СДЭК: 1234567890"\n}',
  },
  {
    id: 'messaging-drom-dialogs',
    method: 'GET',
    path: '/api/messaging/drom/dialogs',
    service: 'MessagingService',
    title: 'Диалоги покупателей с Drom.ru',
    description: 'Интеграция с личным кабинетом Дрома: получение входящих сообщений от покупателей.',
    authRequired: true,
    requiredRole: 'manager',
  },
  // Export
  {
    id: 'export-drom-pricelist',
    method: 'GET',
    path: '/uploads/pricelist.xml',
    service: 'ExportService',
    title: 'XML фид прайс-листа для Drom.ru',
    description: 'Публичный XML каталог для автовыгрузки на Дром. Включает детали в наличии (quantity > 0) и дефектные запчасти (quantity = 0). Проданные (quantity = -1) исключаются.',
    authRequired: false,
  },
  {
    id: 'export-drom-live',
    method: 'GET',
    path: '/api/export/drom.xml',
    service: 'ExportService',
    title: 'Потоковая генерация XML фида Дром',
    description: 'Генерирует свежий XML фид через gRPC запрос к PartsService с фильтрацией quantity >= 0.',
    authRequired: false,
  },
  {
    id: 'export-avito-xml',
    method: 'GET',
    path: '/api/export/avito.xml',
    service: 'ExportService',
    title: 'XML выгрузка для Avito Авто',
    description: 'Генерирует фид в стандарте Avito с категориями запчастей, фотографиями и характеристиками.',
    authRequired: false,
  },
  {
    id: 'export-rebuild-force',
    method: 'POST',
    path: '/api/export/rebuild',
    service: 'ExportService',
    title: 'Принудительная пересборка XML фидов',
    description: 'Сбрасывает кэш отпечатка склада InventoryVersion и заново формирует файлы выгрузок.',
    authRequired: true,
    requiredRole: 'admin',
  },
];

export function getParsedEndpoints(): ApiEndpoint[] {
  const endpoints: ApiEndpoint[] = [];

  if (swaggerSpec && swaggerSpec.paths) {
    for (const [pathKey, methods] of Object.entries(swaggerSpec.paths)) {
      for (const [methodKey, opAny] of Object.entries(methods as Record<string, unknown>)) {
        const method = methodKey.toUpperCase() as 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
        const op = opAny as {
          operationId?: string;
          summary?: string;
          description?: string;
          tags?: string[];
          parameters?: Array<{
            name: string;
            in: 'path' | 'query' | 'body' | 'header';
            type?: string;
            required?: boolean;
            description?: string;
            schema?: unknown;
          }>;
          responses?: Record<string, unknown>;
        };

        const rawService = (op.tags && op.tags[0]) || 'PartsService';
        let service: ApiEndpoint['service'] = 'PartsService';
        if (rawService.includes('Auth')) service = 'AuthService';
        else if (rawService.includes('Orders')) service = 'OrdersService';
        else if (rawService.includes('Messaging')) service = 'MessagingService';
        else if (rawService.includes('Export')) service = 'ExportService';

        const params: ApiEndpoint['parameters'] = (op.parameters || []).map((p) => ({
          name: p.name,
          in: p.in,
          type: p.type || (p.schema ? 'object' : 'string'),
          required: !!p.required,
          description: p.description || '',
        }));

        endpoints.push({
          id: op.operationId || `${method}-${pathKey}`,
          method,
          path: pathKey,
          service,
          title: op.summary || `${method} ${pathKey}`,
          description: op.description || (op.summary ? `${op.summary} эндпоинт микросервиса ${service}.` : `HTTP эндпоинт ${pathKey}`),
          authRequired: !pathKey.includes('/auth/login') && !pathKey.includes('/validate'),
          parameters: params,
        });
      }
    }
  }

  for (const r of supplementaryRoutes) {
    if (!endpoints.some((e) => e.path === r.path && e.method === r.method)) {
      endpoints.push(r);
    }
  }

  return endpoints;
}
