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
  Database,
  Boxes,
  Workflow,
  Sparkles,
  ChevronDown,
  ChevronRight,
  Package,
  ShoppingCart
} from 'lucide-react';

interface ReadmeProps {
  user?: User | null;
  defaultTab?: 'api' | 'architecture' | 'operator' | 'manager' | 'roles';
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

// Реестр всех эндпоинтов микросервисной системы Avtoplaneta
const API_ENDPOINTS: ApiEndpoint[] = [
  // --- PARTS SERVICE ---
  {
    id: 'parts-inventory',
    method: 'GET',
    path: '/api/v1/inventory',
    service: 'PartsService',
    title: 'Каталог запчастей (Elasticsearch + SQL)',
    description: 'Возвращает список запчастей с фильтрацией, пагинацией и полнотекстовым поиском. Включает детали в наличии (quantity > 0) и дефектные ведомости (quantity = 0).',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'query', in: 'query', type: 'string', required: false, description: 'Поисковый запрос (артикул, марка, название)' },
      { name: 'category', in: 'query', type: 'string', required: false, description: 'Фильтр по категории детали' },
      { name: 'page', in: 'query', type: 'integer', required: false, description: 'Номер страницы (дефолт 1)' },
      { name: 'limit', in: 'query', type: 'integer', required: false, description: 'Количество элементов на странице (дефолт 50)' },
    ],
    responseExample: '{\n  "parts": [\n    {\n      "id": 1042,\n      "name": "Дверь передняя правая",\n      "brand": "Toyota",\n      "model": "Camry",\n      "price": 15000,\n      "quantity": 1,\n      "location": "Склад 1, Ряд B"\n    }\n  ],\n  "total": 12480\n}',
  },
  {
    id: 'parts-get-item',
    method: 'GET',
    path: '/api/v1/parts/item/{id}',
    service: 'PartsService',
    title: 'Карточка запчасти по ID',
    description: 'Возвращает полную карточку автозапчасти со списком фотографий, характеристиками и складской локацией.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'Уникальный ID запчасти' }
    ],
  },
  {
    id: 'parts-create',
    method: 'POST',
    path: '/api/v1/parts',
    service: 'PartsService',
    title: 'Создание новой детали',
    description: 'Добавляет новую запчасть в базу данных и синхронизирует её в поисковый индекс Elasticsearch.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'body', in: 'body', type: 'object', required: true, description: 'Данные новой запчасти' }
    ],
    requestBodyExample: '{\n  "name": "Фара левая LED",\n  "brand": "Nissan",\n  "model": "X-Trail",\n  "category": "Оптика",\n  "price": 28000,\n  "quantity": 1,\n  "location": "Секция A-14",\n  "photos": ["/uploads/parts/photo1.webp"]\n}',
  },
  {
    id: 'parts-update',
    method: 'PUT',
    path: '/api/v1/parts/{id}',
    service: 'PartsService',
    title: 'Редактирование детали',
    description: 'Обновляет характеристики запчасти и переиндексирует в Elasticsearch.',
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
    title: 'Удаление запчасти (Soft Delete)',
    description: 'Помечает деталь как удаленную (`deleted_at = NOW()`), удаляя её из каталога и XML фидов.',
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
    description: 'Уменьшает остаток. Если деталь списывается до 0 при продаже, выставляется quantity = -1 для скрытия из каталога.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID детали' },
      { name: 'amount', in: 'body', type: 'integer', required: true, description: 'Количество для списания' }
    ],
  },
  {
    id: 'parts-defect-reports',
    method: 'POST',
    path: '/api/v1/defect-reports',
    service: 'PartsService',
    title: 'Создание дефектной ведомости',
    description: 'Массовое создание разукомплектованных или дефектных деталей с quantity = 0. Они сохраняются в системе и видны в каталоге.',
    authRequired: true,
    requiredRole: 'any',
    requestBodyExample: '{\n  "car_id": 42,\n  "parts": [\n    {\n      "name": "Бампер передний (трещина)",\n      "category": "Кузов наружные элементы",\n      "price": 5000,\n      "defect": "Трещина справа внизу, под пайку"\n    }\n  ]\n}',
  },
  {
    id: 'parts-stats',
    method: 'GET',
    path: '/api/v1/statistics',
    service: 'PartsService',
    title: 'Сводная статистика инвентаря',
    description: 'Возвращает общее количество деталей на складе, суммарную стоимость в рублях и разбивку по категориям.',
    authRequired: true,
    requiredRole: 'manager',
    responseExample: '{\n  "total_parts": 8450,\n  "total_value": 42150000,\n  "categories": [\n    { "name": "Кузовные детали", "count": 2100 },\n    { "name": "Двигатель и навесное", "count": 1420 }\n  ]\n}',
  },

  // --- ORDERS SERVICE ---
  {
    id: 'orders-list',
    method: 'GET',
    path: '/orders',
    service: 'OrdersService',
    title: 'Список активных заказов',
    description: 'Возвращает текущие незавершенные заказы со статусами, информацией о покупателе и позициями.',
    authRequired: true,
    requiredRole: 'manager',
  },
  {
    id: 'orders-create',
    method: 'POST',
    path: '/orders',
    service: 'OrdersService',
    title: 'Создание нового заказа',
    description: 'Создает заказ, автоматически привязывает или создает клиента (с парсингом телефона и ФИО), резервирует позиции и публикует событие в Redis Streams.',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "buyer_number": "+7 (952) 809-40-32 Иванов Иван",\n  "transport_company": "СДЭК",\n  "notes": "Отправить до пятницы",\n  "items": [\n    { "part_id": 1042, "quantity": 1, "price": 15000 }\n  ]\n}',
  },
  {
    id: 'orders-complete',
    method: 'PUT',
    path: '/orders/{id}/complete',
    service: 'OrdersService',
    title: 'Завершение заказа (Списание)',
    description: 'Переводит заказ в статус "зеленый" (Выполнен), архивирует его и списывает остатки деталей через gRPC в PartsService (выставляя quantity = -1 при исчерпании).',
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
    title: 'Обновление статуса заказа',
    description: 'Меняет статус (red = Новый, brown = В сборке, yellow = Готов к выдаче, green = Завершен).',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID заказа' },
      { name: 'status', in: 'body', type: 'string', required: true, description: 'Код статуса: red | brown | yellow | green' }
    ],
  },
  {
    id: 'orders-customers-list',
    method: 'GET',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Клиентская база с аналитикой',
    description: 'Возвращает список клиентов с количеством заказов, суммарным чеком (LTV), городом, любимой ТК и накопительной скидкой.',
    authRequired: true,
    requiredRole: 'manager',
    parameters: [
      { name: 'search', in: 'query', type: 'string', required: false, description: 'Поиск по имени, телефону (+7...) или городу' },
      { name: 'category', in: 'query', type: 'string', required: false, description: 'Фильтр: Обычный | Оптовик | Постоянный' }
    ],
  },
  {
    id: 'orders-customers-create',
    method: 'POST',
    path: '/orders/customers',
    service: 'OrdersService',
    title: 'Создание карточки клиента',
    description: 'Добавляет клиента с уникальным телефонным номером (+7XXXXXXXXXX). Защита от дублей вернет 409 Conflict при попытке повтора.',
    authRequired: true,
    requiredRole: 'manager',
    requestBodyExample: '{\n  "name": "Алексей Смирнов",\n  "phone": "+79528094032",\n  "city": "Томск",\n  "preferred_tk": "Энергия",\n  "category": "Постоянный",\n  "discount_percent": 5\n}',
  },
  {
    id: 'orders-monthly-sales',
    method: 'GET',
    path: '/orders/monthly-sales',
    service: 'OrdersService',
    title: 'Ежемесячная аналитика продаж',
    description: 'Суммарные продажи по месяцам на основе успешно завершенных заказов.',
    authRequired: true,
    requiredRole: 'manager',
  },

  // --- AUTH SERVICE ---
  {
    id: 'auth-login',
    method: 'POST',
    path: '/auth/login',
    service: 'AuthService',
    title: 'Вход по паролю',
    description: 'Аутентифицирует пользователя и выставляет защищенную сессионную cookie с флагами HttpOnly и SameSite=Lax.',
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
    id: 'auth-users-list',
    method: 'GET',
    path: '/api/v1/users',
    service: 'AuthService',
    title: 'Список пользователей системы',
    description: 'Возвращает учетные записи операторов, менеджеров и администраторов.',
    authRequired: true,
    requiredRole: 'admin',
  },
  {
    id: 'auth-activity',
    method: 'GET',
    path: '/api/v1/activity',
    service: 'AuthService',
    title: 'Журнал активности (Аудит)',
    description: 'Лог действий пользователей в системе (кто, когда и какую операцию совершил) для обеспечения безопасности.',
    authRequired: true,
    requiredRole: 'admin',
  },

  // --- MESSAGING SERVICE ---
  {
    id: 'messaging-threads',
    method: 'GET',
    path: '/api/messaging/threads',
    service: 'MessagingService',
    title: 'Список диалогов и чатов',
    description: 'Возвращает треды переписки по клиентам и заказам с количеством непрочитанных сообщений.',
    authRequired: true,
    requiredRole: 'any',
  },
  {
    id: 'messaging-send',
    method: 'POST',
    path: '/api/messaging/threads/{id}/messages',
    service: 'MessagingService',
    title: 'Отправка сообщения в тред',
    description: 'Публикует сообщение в чат с клиентом или внутреннюю заметку команды.',
    authRequired: true,
    requiredRole: 'any',
    parameters: [
      { name: 'id', in: 'path', type: 'integer', required: true, description: 'ID треда' },
      { name: 'content', in: 'body', type: 'string', required: true, description: 'Текст сообщения' }
    ],
  },

  // --- EXPORT SERVICE ---
  {
    id: 'export-drom',
    method: 'GET',
    path: '/api/export/drom.xml',
    service: 'ExportService',
    title: 'XML фид для Drom.ru',
    description: 'Генерирует актуальный XML каталог для автовыгрузки на Дром. Включает детали в наличии (quantity > 0) и дефектные запчасти (quantity = 0). Проданные детали (quantity = -1) исключаются.',
    authRequired: false,
  },
  {
    id: 'export-avito',
    method: 'GET',
    path: '/api/export/avito.xml',
    service: 'ExportService',
    title: 'XML фид для Avito Авто',
    description: 'Генерирует фид в стандарте Avito с категориями запчастей, фотографиями и характеристиками.',
    authRequired: false,
  },
];

export default function Readme({ user, defaultTab = 'api' }: ReadmeProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const urlTab = searchParams.get('tab') as ReadmeProps['defaultTab'] | null;

  const [activeTab, setActiveTab] = useState<'api' | 'architecture' | 'operator' | 'manager' | 'roles'>(
    urlTab || defaultTab || 'api'
  );

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedService, setSelectedService] = useState<string>('ALL');
  const [selectedMethod, setSelectedMethod] = useState<string>('ALL');
  const [expandedEndpoints, setExpandedEndpoints] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Синхронизация таба с URL параметрами
  useEffect(() => {
    if (urlTab && ['api', 'architecture', 'operator', 'manager', 'roles'].includes(urlTab)) {
      setActiveTab(urlTab);
    }
  }, [urlTab]);

  const handleTabChange = (tab: 'api' | 'architecture' | 'operator' | 'manager' | 'roles') => {
    setActiveTab(tab);
    setSearchParams({ tab });
  };

  const toggleEndpoint = (id: string) => {
    setExpandedEndpoints((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  // Фильтрация эндпоинтов API
  const filteredEndpoints = useMemo(() => {
    return API_ENDPOINTS.filter((item) => {
      if (selectedService !== 'ALL' && item.service !== selectedService) return false;
      if (selectedMethod !== 'ALL' && item.method !== selectedMethod) return false;
      if (searchQuery.trim() !== '') {
        const query = searchQuery.toLowerCase();
        const inPath = item.path.toLowerCase().includes(query);
        const inTitle = item.title.toLowerCase().includes(query);
        const inDesc = item.description.toLowerCase().includes(query);
        const inService = item.service.toLowerCase().includes(query);
        if (!inPath && !inTitle && !inDesc && !inService) return false;
      }
      return true;
    });
  }, [searchQuery, selectedService, selectedMethod]);

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

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8">
      {/* Шапка портала документации */}
      <div className="mb-8 border-b border-border pb-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary mb-3">
              <Sparkles className="w-3.5 h-3.5" />
              База знаний и Спецификация v2.4
            </div>
            <h1 className="text-3xl font-extrabold tracking-tight">Центр документации Avtoplaneta</h1>
            <p className="text-muted-foreground mt-1 text-sm md:text-base">
              Интерактивная спецификация REST/gRPC API, архитектура микросервисов и рабочие регламенты команды.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <a
              href="/api.swagger.json"
              download="api.swagger.json"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card hover:bg-accent text-sm font-medium transition-colors shadow-sm"
              title="Скачать полную спецификацию OpenAPI / Swagger в формате JSON"
            >
              <Download className="w-4 h-4 text-primary" />
              <span>Swagger JSON</span>
            </a>

            {user && (
              <div className="hidden sm:flex items-center gap-2 px-3 py-2 rounded-lg bg-muted/60 text-xs text-muted-foreground">
                <span>Ваша роль:</span>
                <span className="font-semibold text-foreground capitalize">
                  {user.role === 'admin' ? 'Администратор' : user.role === 'manager' ? 'Менеджер' : 'Оператор'}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Навигационные вкладки */}
        <div className="flex items-center gap-1 sm:gap-2 mt-6 overflow-x-auto pb-1 scrollbar-none">
          <button
            onClick={() => handleTabChange('api')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'api'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <Code2 className="w-4 h-4" />
            <span>Спецификация API</span>
            <span className="ml-1 text-xs px-1.5 py-0.2 rounded-full bg-background/20 font-mono">
              {API_ENDPOINTS.length}
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
            <span>Архитектура & Сервисы</span>
          </button>

          <button
            onClick={() => handleTabChange('operator')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'operator'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <Package className="w-4 h-4" />
            <span>Инструкция оператора</span>
          </button>

          <button
            onClick={() => handleTabChange('manager')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all shrink-0 ${
              activeTab === 'manager'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            <ShoppingCart className="w-4 h-4" />
            <span>Инструкция менеджера</span>
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
            <span>Роли и безопасность</span>
          </button>
        </div>
      </div>

      {/* ========================================================================= */}
      {/* ВКЛАДКА 1: СПЕЦИФИКАЦИЯ API (SWAGGER / REST EXPLORER)                     */}
      {/* ========================================================================= */}
      {activeTab === 'api' && (
        <div className="space-y-6">
          {/* Поиск и фильтры по сервисам */}
          <div className="flex flex-col md:flex-row gap-4 items-stretch md:items-center justify-between bg-card border border-border rounded-xl p-4 shadow-sm">
            <div className="relative flex-1">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Поиск эндпоинта, пути (/api/v1/...), метода или описания..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-4 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <select
                value={selectedService}
                onChange={(e) => setSelectedService(e.target.value)}
                className="px-3 py-2 bg-background border border-border rounded-lg text-xs font-medium focus:outline-none focus:ring-2 focus:ring-primary/40"
              >
                <option value="ALL">Все микросервисы</option>
                <option value="PartsService">Parts Service</option>
                <option value="OrdersService">Orders Service</option>
                <option value="AuthService">Auth Service</option>
                <option value="MessagingService">Messaging Service</option>
                <option value="ExportService">Export Service</option>
              </select>

              <select
                value={selectedMethod}
                onChange={(e) => setSelectedMethod(e.target.value)}
                className="px-3 py-2 bg-background border border-border rounded-lg text-xs font-medium focus:outline-none focus:ring-2 focus:ring-primary/40"
              >
                <option value="ALL">Все HTTP методы</option>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
              </select>
            </div>
          </div>

          {/* Список эндпоинтов */}
          <div className="space-y-3">
            {filteredEndpoints.length === 0 ? (
              <div className="text-center py-12 border border-dashed border-border rounded-xl">
                <Search className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                <p className="text-muted-foreground text-sm">По вашему запросу не найдено ни одного эндпоинта.</p>
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

                    {/* Развернутые детали эндпоинта */}
                    {isExpanded && (
                      <div className="border-t border-border px-4 py-5 bg-muted/10 space-y-4 text-sm">
                        <div>
                          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                            Назначение и логика
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
                                    <th className="py-2 px-3 text-left">Где передается</th>
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
                              Пример успешного ответа (200 OK)
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
                            <span>{copiedId === ep.id + '-url' ? 'Скопировано!' : 'Копировать путь'}</span>
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
                            <span>{copiedId === ep.id + '-curl' ? 'Скопировано cURL!' : 'Копировать cURL'}</span>
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
      {/* ВКЛАДКА 2: АРХИТЕКТУРА & МИКРОСЕРВИСЫ                                      */}
      {/* ========================================================================= */}
      {activeTab === 'architecture' && (
        <div className="space-y-8">
          {/* Интерактивная карта архитектуры */}
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-3 flex items-center gap-2">
              <Workflow className="w-5 h-5 text-primary" />
              Схема топологии и потоков данных
            </h2>
            <p className="text-sm text-muted-foreground mb-6">
              Внешний трафик поступает через единую точку входа (Traefik Ingress). Сессии проверяются на лету через
              Traefik ForwardAuth. Микросервисы взаимодействуют по высокоскоростному gRPC (Protobuf), а асинхронные
              события передаются через Redis Streams.
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

          {/* Карточки ключевых правил системы */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
              <h3 className="text-lg font-bold mb-3 flex items-center gap-2">
                <Boxes className="w-5 h-5 text-amber-500" />
                Логика остатков (Quantity Lifecycle)
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                Фундаментальное бизнес-правило учета количества запчастей в базе данных, каталоге и XML выгрузках:
              </p>
              <ul className="space-y-3 text-xs md:text-sm">
                <li className="flex items-start gap-2">
                  <span className="font-mono font-bold text-emerald-500 bg-emerald-500/10 px-2 py-0.5 rounded">
                    quantity &gt; 0
                  </span>
                  <span className="text-muted-foreground">
                    Деталь физически доступна на складе для продажи. Отображается в общем каталоге и фидах.
                  </span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="font-mono font-bold text-blue-500 bg-blue-500/10 px-2 py-0.5 rounded">
                    quantity == 0
                  </span>
                  <span className="text-muted-foreground">
                    Деталь из дефектной ведомости (разукомплектовка). <strong>ОБЯЗАТЕЛЬНО</strong> остается в каталоге и
                    XML выгрузках (<code className="text-xs">quantity &gt;= 0</code>).
                  </span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="font-mono font-bold text-rose-500 bg-rose-500/10 px-2 py-0.5 rounded">
                    quantity == -1
                  </span>
                  <span className="text-muted-foreground">
                    Маркер проданной детали! При завершении заказа OrdersService выставляет -1. Это скрывает деталь из
                    каталога без физического удаления из базы.
                  </span>
                </li>
              </ul>
            </div>

            <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
              <h3 className="text-lg font-bold mb-3 flex items-center gap-2">
                <Database className="w-5 h-5 text-primary" />
                Стек хранения данных (Storage Tier)
              </h3>
              <ul className="space-y-4 text-xs md:text-sm">
                <li className="border-b border-border pb-3">
                  <div className="font-semibold text-foreground flex items-center gap-2">
                    <span>PostgreSQL (Database-per-Service)</span>
                    <span className="text-[11px] font-mono text-primary bg-primary/10 px-1.5 py-0.5 rounded">sqlc + pgx/v5</span>
                  </div>
                  <p className="text-muted-foreground mt-1 text-xs">
                    Каждый микросервис владеет изолированной схемой БД. Запросы строго типизированы через sqlc с нулевым
                    оверхедом и ACID транзакциями.
                  </p>
                </li>
                <li className="border-b border-border pb-3">
                  <div className="font-semibold text-foreground flex items-center gap-2">
                    <span>Elasticsearch</span>
                    <span className="text-[11px] font-mono text-primary bg-primary/10 px-1.5 py-0.5 rounded">Full-Text Search</span>
                  </div>
                  <p className="text-muted-foreground mt-1 text-xs">
                    Молниеносный поиск по десяткам тысяч запчастей, опечаткам, OEM-кодам и фасетным фильтрам (марка,
                    модель, кузов).
                  </p>
                </li>
                <li>
                  <div className="font-semibold text-foreground flex items-center gap-2">
                    <span>Redis & Redis Streams</span>
                    <span className="text-[11px] font-mono text-primary bg-primary/10 px-1.5 py-0.5 rounded">Async Events</span>
                  </div>
                  <p className="text-muted-foreground mt-1 text-xs">
                    Кэширование горячих сессий, защита от частых запросов и шина асинхронных событий между сервисами.
                  </p>
                </li>
              </ul>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* ВКЛАДКА 3: ИНСТРУКЦИЯ ОПЕРАТОРА                                           */}
      {/* ========================================================================= */}
      {activeTab === 'operator' && (
        <div className="space-y-6">
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-2">Инструкция оператора склада</h2>
            <p className="text-sm text-muted-foreground mb-6">
              Оператор отвечает за ежедневное управление каталогом запчастей, создание карточек деталей, загрузку фотографий
              и оформление дефектных ведомостей.
            </p>

            <div className="space-y-6 text-sm">
              <div className="border-l-2 border-primary pl-4">
                <h3 className="font-semibold text-base mb-1">1. Добавление новой запчасти в каталог</h3>
                <p className="text-muted-foreground mb-3 text-xs leading-relaxed">
                  Перейдите в раздел <strong>"Инвентарь"</strong> и нажмите <strong>"Добавить запчасть"</strong>.
                </p>
                <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                  <li><strong>Название:</strong> Указывайте точное наименование (например: "Крыло переднее правое").</li>
                  <li><strong>Марка и модель:</strong> Выберите из автодополнения (справочник Drom).</li>
                  <li><strong>Артикул / OEM:</strong> Обязательно вносите заводской номер детали для точного поиска.</li>
                  <li><strong>Количество:</strong> Для целой запчасти всегда ставьте фактический остаток (1, 2...).</li>
                  <li><strong>Фотографии:</strong> Загружайте минимум 3 фото (общий вид, дефекты/маркировка, обратная сторона).</li>
                </ul>
              </div>

              <div className="border-l-2 border-primary pl-4">
                <h3 className="font-semibold text-base mb-1">2. Работа с дефектными ведомостями</h3>
                <p className="text-muted-foreground mb-2 text-xs leading-relaxed">
                  При разборе автомобиля или выявлении брака создавайте дефектный отчёт через раздел <strong>"Дефектовка"</strong>.
                </p>
                <div className="p-3 bg-muted rounded-lg text-xs space-y-1 text-muted-foreground">
                  <p className="font-medium text-foreground">Важное правило системы:</p>
                  <p>
                    Дефектным запчастям система автоматически присваивает количество <code>0</code>.
                    Они остаются полностью видимыми в каталоге и выгружаются на Drom/Avito, так как клиенты часто ищут
                    запчасти под восстановление.
                  </p>
                </div>
              </div>

              <div className="border-l-2 border-primary pl-4">
                <h3 className="font-semibold text-base mb-1">3. Мобильное приложение (Flutter)</h3>
                <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                  <li>Скачайте актуальный APK из релизов и выполните вход под логином оператора.</li>
                  <li>Используйте камеру смартфона для быстрой фотосъемки и привязки изображений к детали.</li>
                  <li>Проверяйте локацию хранения (стеллаж, полка) прямо на складе.</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* ВКЛАДКА 4: ИНСТРУКЦИЯ МЕНЕДЖЕРА                                            */}
      {/* ========================================================================= */}
      {activeTab === 'manager' && (
        <div className="space-y-6">
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-2">Инструкция менеджера по продажам</h2>
            <p className="text-sm text-muted-foreground mb-6">
              Менеджер обрабатывает входящие заказы, формирует клиентскую базу, контролирует доставку через транспортные
              компании и списывает проданные позиции.
            </p>

            <div className="space-y-6 text-sm">
              <div className="border-l-2 border-primary pl-4">
                <h3 className="font-semibold text-base mb-2">1. Цветовая кодировка жизненного цикла заказа</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs">
                  <div className="p-3 rounded-lg border border-rose-500/30 bg-rose-500/10">
                    <div className="font-bold text-rose-600 dark:text-rose-400 mb-1">Красный (Новый)</div>
                    <p className="text-muted-foreground">Заказ поступил с сайта или мобильного приложения. Требует звонка клиенту.</p>
                  </div>
                  <div className="p-3 rounded-lg border border-amber-600/30 bg-amber-600/10">
                    <div className="font-bold text-amber-700 dark:text-amber-400 mb-1">Коричневый (В сборке)</div>
                    <p className="text-muted-foreground">Оплата подтверждена. Склад комплектует детали к отправке.</p>
                  </div>
                  <div className="p-3 rounded-lg border border-yellow-500/30 bg-yellow-500/10">
                    <div className="font-bold text-yellow-600 dark:text-yellow-400 mb-1">Желтый (Готов)</div>
                    <p className="text-muted-foreground">Передан курьеру ТК или ожидает самовывоза в пункте выдачи.</p>
                  </div>
                  <div className="p-3 rounded-lg border border-emerald-500/30 bg-emerald-500/10">
                    <div className="font-bold text-emerald-600 dark:text-emerald-400 mb-1">Зеленый (Выполнен)</div>
                    <p className="text-muted-foreground">Клиент получил заказ. Происходит автоматическое списание детали (quantity = -1).</p>
                  </div>
                </div>
              </div>

              <div className="border-l-2 border-primary pl-4">
                <h3 className="font-semibold text-base mb-1">2. Работа с клиентской базой (CRM)</h3>
                <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                  <li>
                    <strong>Автоматическая нормализация:</strong> Система автоматически приводит телефоны клиентов к каноническому виду{' '}
                    <code className="text-xs">+7XXXXXXXXXX</code> и отделяет имя покупателя от служебных меток ("Получатель").
                  </li>
                  <li>
                    <strong>Аналитика LTV:</strong> В таблице клиентов сразу видны суммарный объем покупок и история обращений.
                  </li>
                  <li>
                    <strong>Скидки:</strong> Можно назначать персональный процент скидки клиенту, который учитывается при расчете заказа.
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* ВКЛАДКА 5: РОЛИ И БЕЗОПАСНОСТЬ (RBAC)                                     */}
      {/* ========================================================================= */}
      {activeTab === 'roles' && (
        <div className="space-y-6">
          <div className="border border-border rounded-xl bg-card p-6 shadow-sm">
            <h2 className="text-xl font-bold mb-2">Ролевая модель доступа (RBAC)</h2>
            <p className="text-sm text-muted-foreground mb-6">
              Доступ к операциям разграничен на уровне Traefik ForwardAuth и микросервисов.
            </p>

            <div className="overflow-x-auto border border-border rounded-xl">
              <table className="w-full text-xs md:text-sm">
                <thead>
                  <tr className="bg-muted/60 text-muted-foreground border-b border-border">
                    <th className="py-3 px-4 text-left font-semibold">Операция / Ресурс</th>
                    <th className="py-3 px-4 text-center font-semibold">Оператор</th>
                    <th className="py-3 px-4 text-center font-semibold">Менеджер</th>
                    <th className="py-3 px-4 text-center font-semibold">Администратор</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="py-3 px-4 font-medium">Просмотр и поиск по каталогу</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Добавление деталей и дефектовки</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Удаление запчастей из базы</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Создание и управление заказами</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Клиентская база и скидки</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Управление сотрудниками (создание, роли)</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-rose-500 font-bold">❌</td>
                    <td className="py-3 px-4 text-center text-emerald-500 font-bold">✅</td>
                  </tr>
                  <tr>
                    <td className="py-3 px-4 font-medium">Просмотр аудита действий и логов</td>
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
    </div>
  );
}
