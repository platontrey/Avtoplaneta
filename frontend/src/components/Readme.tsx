/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { useState, useMemo, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { User } from '../features/auth/types';
import MarkdownViewer from './MarkdownViewer';
import swaggerSpec from '../data/swagger.json';

// Прямой импорт оригинальных .md файлов репозитория через Vite (Single Source of Truth)
import operatorMd from '../../../operator-instructions.md?raw';
import managerMd from '../../../manager-instructions.md?raw';
import archMd from '../../../backend/ARCHITECTURE.md?raw';
import traefikMd from '../../../TRAEFIK-README.md?raw';
import monitoringMd from '../../../MONITORING-README.md?raw';
import mobileMd from '../../../AvtoplanetaApp/README.md?raw';
import frontendMd from '../../README.md?raw';
import migrationMd from '../../../scripts/migration/README.md?raw';
import rootReadmeMd from '../../../README.md?raw';

import {
  Code2,
  Server,
  Shield,
  Search,
  Copy,
  Check,
  Download,
  Sparkles,
  ChevronDown,
  ChevronRight,
  Package,
  ShoppingCart,
  BookOpen,
  Smartphone,
  FileCode,
  ExternalLink,
  Database,
  Activity,
  FileText
} from 'lucide-react';

interface ReadmeProps {
  user?: User | null;
  defaultTab?: 'guides' | 'api' | 'architecture' | 'roles' | 'faq' | 'operator' | 'manager';
}

// Документ из репозитория
interface RepoDocument {
  id: string;
  title: string;
  category: 'Инструкции персонала' | 'Архитектура и бэкенд' | 'Инфраструктура и DevOps' | 'Клиенты и приложения' | 'Общие сведения';
  filePath: string;
  icon: React.ElementType;
  summary: string;
  badge: string;
  content: string;
}

// Эндпоинт API
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

export default function Readme({ defaultTab = 'guides' }: ReadmeProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get('tab');
  const urlDoc = searchParams.get('doc');

  // Каталог оригинальных .md файлов репозитория
  const REPO_DOCS: RepoDocument[] = useMemo(() => [
    {
      id: 'operator',
      title: 'Инструкция оператора склада',
      category: 'Инструкции персонала',
      filePath: 'operator-instructions.md',
      icon: Package,
      summary: 'Приемка деталей, фотографирование, складская адресация, оформление дефектных ведомостей.',
      badge: 'Оператор',
      content: operatorMd,
    },
    {
      id: 'manager',
      title: 'Инструкция менеджера по продажам',
      category: 'Инструкции персонала',
      filePath: 'manager-instructions.md',
      icon: ShoppingCart,
      summary: 'Работа с заказами, цветовые статусы, списание деталей со склада, CRM клиенты и ТК.',
      badge: 'Менеджер',
      content: managerMd,
    },
    {
      id: 'architecture',
      title: 'Архитектура микросервисов бэкенда',
      category: 'Архитектура и бэкенд',
      filePath: 'backend/ARCHITECTURE.md',
      icon: Server,
      summary: 'Архитектура сервисов (Auth, Parts, Orders, Messaging, Export), gRPC, Traefik, PostgreSQL, Redis Streams, Elasticsearch.',
      badge: 'Бэкенд',
      content: archMd,
    },
    {
      id: 'traefik',
      title: 'Traefik Ingress, TLS и ForwardAuth',
      category: 'Инфраструктура и DevOps',
      filePath: 'TRAEFIK-README.md',
      icon: Shield,
      summary: 'Маршрутизация HTTP трафика, SSL сертификаты Let\'s Encrypt, валидация сессий через ForwardAuth.',
      badge: 'DevOps',
      content: traefikMd,
    },
    {
      id: 'monitoring',
      title: 'Мониторинг (Prometheus & Grafana)',
      category: 'Инфраструктура и DevOps',
      filePath: 'MONITORING-README.md',
      icon: Activity,
      summary: 'Сбор метрик со всех сервисов, дашборды Grafana, cAdvisor, Node Exporter и мониторинг latency.',
      badge: 'DevOps',
      content: monitoringMd,
    },
    {
      id: 'mobile',
      title: 'Мобильное приложение (Flutter)',
      category: 'Клиенты и приложения',
      filePath: 'AvtoplanetaApp/README.md',
      icon: Smartphone,
      summary: 'Клиент для Android и iOS: кроссплатформенная сборка, переменные окружения, работа со складом и Google Auth.',
      badge: 'Mobile',
      content: mobileMd,
    },
    {
      id: 'frontend',
      title: 'Веб-клиент (React + TypeScript)',
      category: 'Клиенты и приложения',
      filePath: 'frontend/README.md',
      icon: Code2,
      summary: 'SPA клиент, Vite, PWA, Tailwind CSS v4, TanStack Query, Radix UI и ролевые экраны.',
      badge: 'Frontend',
      content: frontendMd,
    },
    {
      id: 'migration',
      title: 'Миграция каталога JoomShopping',
      category: 'Архитектура и бэкенд',
      filePath: 'scripts/migration/README.md',
      icon: Database,
      summary: 'Скрипты импорта товаров из legacy MySQL Joomla/JoomShopping в PostgreSQL склада.',
      badge: 'База данных',
      content: migrationMd,
    },
    {
      id: 'root-readme',
      title: 'Главный обзор проекта Avtoplaneta',
      category: 'Общие сведения',
      filePath: 'README.md',
      icon: BookOpen,
      summary: 'Обзор всей экосистемы проекта, Docker Compose запуск одной командой и порты сервисов.',
      badge: 'Обзор',
      content: rootReadmeMd,
    },
  ], []);

  // Первоначальный выбор вкладки
  const initialTab = useMemo<'guides' | 'api' | 'architecture' | 'roles' | 'faq'>(() => {
    if (rawTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(rawTab)) {
      return rawTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq';
    }
    if (defaultTab === 'operator' || defaultTab === 'manager') {
      return 'guides';
    }
    if (defaultTab === 'architecture') {
      return 'architecture';
    }
    if (defaultTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(defaultTab)) {
      return defaultTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq';
    }
    return 'guides';
  }, [rawTab, defaultTab]);

  const [activeTab, setActiveTab] = useState<'guides' | 'api' | 'architecture' | 'roles' | 'faq'>(initialTab);

  // Выбранный документ репозитория
  const initialDocId = useMemo(() => {
    if (urlDoc && REPO_DOCS.some((d) => d.id === urlDoc)) {
      return urlDoc;
    }
    if (defaultTab === 'operator') return 'operator';
    if (defaultTab === 'manager') return 'manager';
    if (defaultTab === 'architecture') return 'architecture';
    return 'root-readme';
  }, [urlDoc, defaultTab, REPO_DOCS]);

  const [selectedDocId, setSelectedDocId] = useState<string>(initialDocId);
  const [docSearchQuery, setDocSearchQuery] = useState('');
  const [copiedDoc, setCopiedDoc] = useState(false);

  // Состояние API Explorer
  const [apiSearchQuery, setApiSearchQuery] = useState('');
  const [selectedService, setSelectedService] = useState<string>('ALL');
  const [selectedMethod, setSelectedMethod] = useState<string>('ALL');
  const [expandedEndpoints, setExpandedEndpoints] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Синхронизация с URL при смене роута
  useEffect(() => {
    if (rawTab && ['guides', 'api', 'architecture', 'roles', 'faq'].includes(rawTab)) {
      setActiveTab(rawTab as 'guides' | 'api' | 'architecture' | 'roles' | 'faq');
    }
    if (urlDoc && REPO_DOCS.some((d) => d.id === urlDoc)) {
      setSelectedDocId(urlDoc);
    }
  }, [rawTab, urlDoc, REPO_DOCS]);

  const handleTabChange = (tab: 'guides' | 'api' | 'architecture' | 'roles' | 'faq') => {
    setActiveTab(tab);
    setSearchParams({ tab, doc: selectedDocId });
  };

  const handleSelectDoc = (docId: string) => {
    setSelectedDocId(docId);
    setSearchParams({ tab: activeTab, doc: docId });
  };

  const activeDoc = useMemo(() => {
    return REPO_DOCS.find((d) => d.id === selectedDocId) || REPO_DOCS[0];
  }, [selectedDocId, REPO_DOCS]);

  // Фильтрация списка документов
  const filteredDocs = useMemo(() => {
    if (!docSearchQuery.trim()) return REPO_DOCS;
    const query = docSearchQuery.toLowerCase();
    return REPO_DOCS.filter(
      (d) =>
        d.title.toLowerCase().includes(query) ||
        d.filePath.toLowerCase().includes(query) ||
        d.summary.toLowerCase().includes(query) ||
        d.content.toLowerCase().includes(query)
    );
  }, [docSearchQuery, REPO_DOCS]);

  // Скачивание .md файла
  const handleDownloadDoc = (doc: RepoDocument) => {
    const blob = new Blob([doc.content], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = doc.filePath.split('/').pop() || `${doc.id}.md`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // Копирование Markdown текста
  const handleCopyDocContent = (content: string) => {
    navigator.clipboard.writeText(content);
    setCopiedDoc(true);
    setTimeout(() => setCopiedDoc(false), 2000);
  };

  // ---------------------------------------------------------------------------
  // ДИНАМИЧЕСКИЙ РЕЕСТР API (Парсинг OpenAPI swagger.json + дополнительные HTTP роуты)
  // ---------------------------------------------------------------------------
  const parsedEndpoints: ApiEndpoint[] = useMemo(() => {
    const endpoints: ApiEndpoint[] = [];

    // 1. Динамический парсинг путей из swagger.json
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

    // 2. Дополнительные нативные HTTP роуты OrdersService, MessagingService, ExportService
    const supplementaryRoutes: ApiEndpoint[] = [
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

    // Добавляем дополнительные маршруты, если они не были в OpenAPI
    for (const r of supplementaryRoutes) {
      if (!endpoints.some((e) => e.path === r.path && e.method === r.method)) {
        endpoints.push(r);
      }
    }

    return endpoints;
  }, []);

  // Фильтрация API
  const filteredEndpoints = useMemo(() => {
    return parsedEndpoints.filter((item) => {
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
  }, [parsedEndpoints, apiSearchQuery, selectedService, selectedMethod]);

  const toggleEndpoint = (id: string) => {
    setExpandedEndpoints((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleDownloadSwagger = () => {
    const blob = new Blob([JSON.stringify(swaggerSpec, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'api.swagger.json';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

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
    <div className="min-h-screen bg-background text-foreground pb-20">
      {/* Шапка Центра документации */}
      <header className="border-b border-border bg-card/60 backdrop-blur-md sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-xl bg-primary/10 text-primary">
                  <BookOpen className="w-6 h-6" />
                </div>
                <div>
                  <h1 className="text-2xl font-bold tracking-tight">Документация и API Avtoplaneta</h1>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Оригинальные Markdown документы репозитория (Docs-as-Code) и OpenAPI 2.0 спецификация
                  </p>
                </div>
              </div>
            </div>

            {/* Вкладки навигации */}
            <div className="flex items-center gap-1.5 p-1 bg-muted/70 rounded-xl border border-border text-xs font-medium overflow-x-auto">
              <button
                onClick={() => handleTabChange('guides')}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-lg transition-all ${
                  activeTab === 'guides'
                    ? 'bg-card text-foreground shadow-sm font-semibold'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <FileText className="w-4 h-4 text-primary" />
                <span>Документы репозитория ({REPO_DOCS.length})</span>
              </button>

              <button
                onClick={() => handleTabChange('api')}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-lg transition-all ${
                  activeTab === 'api'
                    ? 'bg-card text-foreground shadow-sm font-semibold'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Code2 className="w-4 h-4 text-emerald-500" />
                <span>API Explorer ({parsedEndpoints.length})</span>
              </button>

              <button
                onClick={() => handleTabChange('architecture')}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-lg transition-all ${
                  activeTab === 'architecture'
                    ? 'bg-card text-foreground shadow-sm font-semibold'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Server className="w-4 h-4 text-blue-500" />
                <span>Архитектура системы</span>
              </button>

              <button
                onClick={() => handleTabChange('roles')}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-lg transition-all ${
                  activeTab === 'roles'
                    ? 'bg-card text-foreground shadow-sm font-semibold'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Shield className="w-4 h-4 text-amber-500" />
                <span>Матрица ролей</span>
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* =================================================================== */}
        {/* ВКЛАДКА 1: ДОКУМЕНТЫ РЕПОЗИТОРИЯ (.MD ФАЙЛЫ)                      */}
        {/* =================================================================== */}
        {activeTab === 'guides' && (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            {/* Боковая панель списка документов */}
            <aside className="lg:col-span-4 space-y-4">
              <div className="p-4 rounded-2xl border border-border bg-card/80 shadow-sm space-y-3">
                <div className="relative">
                  <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <input
                    type="text"
                    placeholder="Поиск по документам..."
                    value={docSearchQuery}
                    onChange={(e) => setDocSearchQuery(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 text-xs rounded-xl border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div className="space-y-1.5 max-h-[calc(100vh-280px)] overflow-y-auto pr-1">
                  {filteredDocs.map((doc) => {
                    const isSelected = doc.id === activeDoc.id;
                    const IconComponent = doc.icon;
                    return (
                      <button
                        key={doc.id}
                        onClick={() => handleSelectDoc(doc.id)}
                        className={`w-full text-left p-3 rounded-xl border transition-all flex items-start gap-3 ${
                          isSelected
                            ? 'bg-primary/10 border-primary/40 shadow-sm'
                            : 'bg-background hover:bg-muted/60 border-border'
                        }`}
                      >
                        <div
                          className={`p-2 rounded-lg shrink-0 mt-0.5 ${
                            isSelected ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'
                          }`}
                        >
                          <IconComponent className="w-4 h-4" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between gap-1 mb-1">
                            <span className="text-xs font-semibold truncate text-foreground">{doc.title}</span>
                            <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-muted text-muted-foreground font-mono shrink-0">
                              {doc.badge}
                            </span>
                          </div>
                          <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed">
                            {doc.summary}
                          </p>
                          <div className="mt-1.5 flex items-center gap-1.5 text-[10px] font-mono text-muted-foreground">
                            <FileCode className="w-3 h-3 text-primary/70" />
                            <span className="truncate">{doc.filePath}</span>
                          </div>
                        </div>
                      </button>
                    );
                  })}
                  {filteredDocs.length === 0 && (
                    <div className="p-6 text-center text-xs text-muted-foreground">
                      Документы по запросу не найдены
                    </div>
                  )}
                </div>
              </div>

              {/* Информационный блок Single Source of Truth */}
              <div className="p-4 rounded-2xl border border-primary/20 bg-primary/5 text-xs space-y-2">
                <div className="flex items-center gap-2 font-semibold text-primary">
                  <Sparkles className="w-4 h-4" />
                  <span>Docs-as-Code (Единый источник)</span>
                </div>
                <p className="text-muted-foreground text-[11px] leading-relaxed">
                  Все инструкции на этой странице импортируются напрямую из оригинальных <code className="font-mono text-primary font-semibold">.md</code> файлов репозитория. Любые правки в файлах на сервере моментально отображаются на сайте без изменения UI кода.
                </p>
              </div>
            </aside>

            {/* Область просмотра документа */}
            <article className="lg:col-span-8 space-y-6">
              <div className="p-6 sm:p-8 rounded-2xl border border-border bg-card shadow-sm">
                {/* Метаданные документа */}
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-border mb-6">
                  <div>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="px-2 py-0.5 text-[10px] font-semibold rounded-full bg-primary/10 text-primary">
                        {activeDoc.category}
                      </span>
                      <span className="text-xs font-mono text-muted-foreground flex items-center gap-1">
                        <FileCode className="w-3.5 h-3.5" />
                        {activeDoc.filePath}
                      </span>
                    </div>
                    <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">
                      {activeDoc.title}
                    </h2>
                  </div>

                  {/* Кнопки действий над файлом */}
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      onClick={() => handleCopyDocContent(activeDoc.content)}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-xl border border-border bg-background hover:bg-muted text-foreground transition-all shadow-sm"
                      title="Скопировать оригинальный Markdown"
                    >
                      {copiedDoc ? (
                        <>
                          <Check className="w-3.5 h-3.5 text-emerald-500" />
                          <span className="text-emerald-500">Скопировано</span>
                        </>
                      ) : (
                        <>
                          <Copy className="w-3.5 h-3.5 text-muted-foreground" />
                          <span>Копировать</span>
                        </>
                      )}
                    </button>

                    <button
                      onClick={() => handleDownloadDoc(activeDoc)}
                      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-xl border border-primary/30 bg-primary/10 hover:bg-primary/20 text-primary transition-all shadow-sm"
                      title="Скачать исходный .md файл"
                    >
                      <Download className="w-3.5 h-3.5" />
                      <span>Скачать .md</span>
                    </button>
                  </div>
                </div>

                {/* Рендеринг содержимого через MarkdownViewer */}
                <div className="prose prose-sm dark:prose-invert max-w-none">
                  <MarkdownViewer content={activeDoc.content} />
                </div>
              </div>
            </article>
          </div>
        )}

        {/* =================================================================== */}
        {/* ВКЛАДКА 2: ИНТЕРАКТИВНЫЙ API EXPLORER (OPENAPI 2.0 / SWAGGER)       */}
        {/* =================================================================== */}
        {activeTab === 'api' && (
          <div className="space-y-6">
            {/* Панель фильтров и экспорта OpenAPI */}
            <div className="p-5 rounded-2xl border border-border bg-card shadow-sm space-y-4">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                  <h2 className="text-lg font-bold flex items-center gap-2">
                    <Code2 className="w-5 h-5 text-emerald-500" />
                    OpenAPI 2.0 / gRPC Gateway Эндпоинты
                  </h2>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Всего зарегистрировано {parsedEndpoints.length} маршрутов во всех 5 микросервисах
                  </p>
                </div>

                <button
                  onClick={handleDownloadSwagger}
                  className="flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/20 transition-all shadow-sm self-start md:self-auto"
                >
                  <Download className="w-4 h-4" />
                  <span>Скачать api.swagger.json</span>
                </button>
              </div>

              {/* Фильтры */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3 pt-2">
                <div className="relative">
                  <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <input
                    type="text"
                    placeholder="Поиск по пути (/api/v1/...), названию..."
                    value={apiSearchQuery}
                    onChange={(e) => setApiSearchQuery(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 text-xs rounded-xl border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div>
                  <select
                    value={selectedService}
                    onChange={(e) => setSelectedService(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-xl border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="ALL">Все микросервисы (All Services)</option>
                    <option value="PartsService">PartsService (Склад, дефекты, фото)</option>
                    <option value="OrdersService">OrdersService (Заказы, CRM, LTV)</option>
                    <option value="AuthService">AuthService (Авторизация, сессии)</option>
                    <option value="MessagingService">MessagingService (Чаты, Дром)</option>
                    <option value="ExportService">ExportService (XML фиды Дром/Авито)</option>
                  </select>
                </div>

                <div>
                  <select
                    value={selectedMethod}
                    onChange={(e) => setSelectedMethod(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-xl border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="ALL">Все HTTP методы</option>
                    <option value="GET">GET</option>
                    <option value="POST">POST</option>
                    <option value="PUT">PUT</option>
                    <option value="DELETE">DELETE</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Список найденных эндпоинтов */}
            <div className="space-y-3">
              {filteredEndpoints.map((item) => {
                const isExpanded = !!expandedEndpoints[item.id];
                return (
                  <div
                    key={item.id}
                    className="rounded-2xl border border-border bg-card overflow-hidden shadow-sm hover:border-border/80 transition-all"
                  >
                    <div
                      onClick={() => toggleEndpoint(item.id)}
                      className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 cursor-pointer hover:bg-muted/30 transition-colors"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <span
                          className={`px-2.5 py-1 text-xs font-mono font-bold rounded-lg border shrink-0 ${methodColorClass(
                            item.method
                          )}`}
                        >
                          {item.method}
                        </span>
                        <code className="text-xs sm:text-sm font-mono font-semibold text-foreground truncate">
                          {item.path}
                        </code>
                        <span className="hidden sm:inline-block text-[11px] px-2 py-0.5 rounded-full bg-muted text-muted-foreground font-mono">
                          {item.service}
                        </span>
                      </div>

                      <div className="flex items-center gap-3 shrink-0">
                        <span className="text-xs text-muted-foreground truncate max-w-[240px] text-right">
                          {item.title}
                        </span>
                        {isExpanded ? (
                          <ChevronDown className="w-4 h-4 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="w-4 h-4 text-muted-foreground" />
                        )}
                      </div>
                    </div>

                    {/* Развернутое описание эндпоинта */}
                    {isExpanded && (
                      <div className="px-5 py-4 border-t border-border bg-muted/20 space-y-4 text-xs">
                        <div>
                          <h4 className="font-semibold text-foreground mb-1">Назначение эндпоинта</h4>
                          <p className="text-muted-foreground leading-relaxed">{item.description}</p>
                        </div>

                        {/* Параметры */}
                        {item.parameters && item.parameters.length > 0 && (
                          <div>
                            <h4 className="font-semibold text-foreground mb-2">Параметры запроса</h4>
                            <div className="overflow-x-auto border border-border rounded-xl">
                              <table className="w-full text-left text-xs">
                                <thead>
                                  <tr className="bg-muted/70 border-b border-border text-muted-foreground">
                                    <th className="py-2 px-3">Параметр</th>
                                    <th className="py-2 px-3">Передача в</th>
                                    <th className="py-2 px-3">Тип</th>
                                    <th className="py-2 px-3">Обязательность</th>
                                    <th className="py-2 px-3">Описание</th>
                                  </tr>
                                </thead>
                                <tbody className="divide-y divide-border">
                                  {item.parameters.map((p, pIdx) => (
                                    <tr key={pIdx}>
                                      <td className="py-2 px-3 font-mono font-semibold text-primary">{p.name}</td>
                                      <td className="py-2 px-3 font-mono text-muted-foreground">{p.in}</td>
                                      <td className="py-2 px-3 font-mono text-muted-foreground">{p.type}</td>
                                      <td className="py-2 px-3">
                                        {p.required ? (
                                          <span className="text-rose-500 font-semibold">Обязательно</span>
                                        ) : (
                                          <span className="text-muted-foreground">Опционально</span>
                                        )}
                                      </td>
                                      <td className="py-2 px-3 text-muted-foreground">{p.description || '—'}</td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            </div>
                          </div>
                        )}

                        {/* Пример тела запроса */}
                        {item.requestBodyExample && (
                          <div>
                            <div className="flex items-center justify-between mb-1.5">
                              <h4 className="font-semibold text-foreground">Пример тела запроса (JSON)</h4>
                              <button
                                onClick={() => copyToClipboard(item.requestBodyExample!, `body-${item.id}`)}
                                className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
                              >
                                {copiedId === `body-${item.id}` ? (
                                  <>
                                    <Check className="w-3 h-3 text-emerald-500" />
                                    <span className="text-emerald-500">Скопировано</span>
                                  </>
                                ) : (
                                  <>
                                    <Copy className="w-3 h-3" />
                                    <span>Копировать</span>
                                  </>
                                )}
                              </button>
                            </div>
                            <pre className="p-3 rounded-xl bg-muted font-mono text-[11px] overflow-x-auto text-foreground">
                              {item.requestBodyExample}
                            </pre>
                          </div>
                        )}

                        {/* cURL команда */}
                        <div>
                          <div className="flex items-center justify-between mb-1.5">
                            <h4 className="font-semibold text-foreground">cURL пример</h4>
                            <button
                              onClick={() => {
                                const curl = `curl -X ${item.method} "https://avtoplaneta.ru${item.path}" ${
                                  item.authRequired ? '-H "Cookie: session_id=..."' : ''
                                }`;
                                copyToClipboard(curl, `curl-${item.id}`);
                              }}
                              className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
                            >
                              {copiedId === `curl-${item.id}` ? (
                                <>
                                  <Check className="w-3 h-3 text-emerald-500" />
                                  <span className="text-emerald-500">Скопировано</span>
                                </>
                              ) : (
                                <>
                                  <Copy className="w-3 h-3" />
                                  <span>Копировать cURL</span>
                                </>
                              )}
                            </button>
                          </div>
                          <pre className="p-3 rounded-xl bg-muted font-mono text-[11px] overflow-x-auto text-primary">
                            {`curl -X ${item.method} "https://avtoplaneta.ru${item.path}" ${
                              item.authRequired ? '-H "Cookie: session_id=..."' : ''
                            }`}
                          </pre>
                        </div>
                      </div>
                    )}
                  </div>
                );
              })}

              {filteredEndpoints.length === 0 && (
                <div className="p-12 text-center text-sm text-muted-foreground border border-dashed border-border rounded-2xl bg-card">
                  Эндпоинты по заданным критериям фильтрации не найдены
                </div>
              )}
            </div>
          </div>
        )}

        {/* =================================================================== */}
        {/* ВКЛАДКА 3: АРХИТЕКТУРА СИСТЕМЫ И ДИАГРАММА                          */}
        {/* =================================================================== */}
        {activeTab === 'architecture' && (
          <div className="space-y-8">
            {/* Архитектурная диаграмма */}
            <div className="p-6 sm:p-8 rounded-2xl border border-border bg-card shadow-sm space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <h2 className="text-xl sm:text-2xl font-bold flex items-center gap-2">
                    <Server className="w-6 h-6 text-blue-500" />
                    Интерактивная архитектурная диаграмма
                  </h2>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Маршрутизация Traefik, микросервисы на Go, хранилища PostgreSQL, Elasticsearch и шина событий Redis Streams
                  </p>
                </div>
                <a
                  href="/architecture-diagram.svg"
                  target="_blank"
                  rel="noreferrer"
                  className="flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold rounded-xl bg-primary/10 text-primary border border-primary/20 hover:bg-primary/20 transition-all self-start sm:self-auto"
                >
                  <ExternalLink className="w-3.5 h-3.5" />
                  <span>Открыть схему в полном размере</span>
                </a>
              </div>

              <div className="rounded-xl border border-border bg-muted/40 p-4 overflow-hidden flex items-center justify-center">
                <img
                  src="/architecture-diagram.svg"
                  alt="Архитектура микросервисов Avtoplaneta"
                  className="max-h-[500px] w-auto object-contain rounded-lg"
                />
              </div>
            </div>

            {/* Полный текст backend/ARCHITECTURE.md */}
            <div className="p-6 sm:p-8 rounded-2xl border border-border bg-card shadow-sm">
              <div className="flex items-center justify-between pb-6 border-b border-border mb-6">
                <div>
                  <span className="text-xs font-mono text-muted-foreground">backend/ARCHITECTURE.md</span>
                  <h3 className="text-xl font-bold mt-1">Спецификация микросервисной архитектуры</h3>
                </div>
                <button
                  onClick={() => handleCopyDocContent(archMd)}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-xl border border-border bg-background hover:bg-muted text-foreground transition-all"
                >
                  {copiedDoc ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                      <span className="text-emerald-500">Скопировано</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5 text-muted-foreground" />
                      <span>Копировать</span>
                    </>
                  )}
                </button>
              </div>

              <div className="prose prose-sm dark:prose-invert max-w-none">
                <MarkdownViewer content={archMd} />
              </div>
            </div>
          </div>
        )}

        {/* =================================================================== */}
        {/* ВКЛАДКА 4: МАТРИЦА РОЛЕЙ И ПРАВ ДОСТУПА                             */}
        {/* =================================================================== */}
        {activeTab === 'roles' && (
          <div className="space-y-6">
            <div className="p-6 sm:p-8 rounded-2xl border border-border bg-card shadow-sm space-y-6">
              <div>
                <h2 className="text-2xl font-bold flex items-center gap-2">
                  <Shield className="w-6 h-6 text-amber-500" />
                  Ролевая модель и разграничение прав доступа
                </h2>
                <p className="text-xs text-muted-foreground mt-1">
                  Каждый входящий запрос валидируется Traefik ForwardAuth в AuthService за &lt;1 мс
                </p>
              </div>

              {/* Сравнительная таблица прав */}
              <div className="overflow-x-auto border border-border rounded-xl">
                <table className="w-full text-xs text-left">
                  <thead>
                    <tr className="bg-muted/70 border-b border-border text-muted-foreground font-semibold">
                      <th className="py-3 px-4">Функциональный модуль</th>
                      <th className="py-3 px-4">Оператор склада (operator)</th>
                      <th className="py-3 px-4">Менеджер продаж (manager)</th>
                      <th className="py-3 px-4">Администратор (admin)</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Просмотр каталога запчастей</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Приемка, фото и дефектовка</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Создание и правка</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Создание и правка</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Создание и ведение заказов</td>
                      <td className="py-3 px-4 text-muted-foreground">❌ Только просмотр</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полное управление</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Списание деталей (quantity = -1)</td>
                      <td className="py-3 px-4 text-muted-foreground">❌ Нет прав</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ При закрытии заказа</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">CRM Клиенты и персональные скидки</td>
                      <td className="py-3 px-4 text-muted-foreground">❌ Нет доступа</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Создание и редактирование</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный доступ</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Управление сотрудниками системы</td>
                      <td className="py-3 px-4 text-rose-500 font-semibold">❌ Запрещено</td>
                      <td className="py-3 px-4 text-rose-500 font-semibold">❌ Запрещено</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Полный CRUD</td>
                    </tr>
                    <tr>
                      <td className="py-3 px-4 font-semibold text-foreground">Журнал аудита действий (Audit Logs)</td>
                      <td className="py-3 px-4 text-rose-500 font-semibold">❌ Запрещено</td>
                      <td className="py-3 px-4 text-rose-500 font-semibold">❌ Запрещено</td>
                      <td className="py-3 px-4 text-emerald-500 font-semibold">✅ Просмотр журнала</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              {/* Быстрые ссылки на инструкции ролей */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4">
                <button
                  onClick={() => {
                    handleTabChange('guides');
                    handleSelectDoc('operator');
                  }}
                  className="p-4 rounded-xl border border-border bg-muted/40 hover:bg-muted text-left transition-all flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-primary/10 text-primary">
                      <Package className="w-5 h-5" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-foreground text-sm">Инструкция для оператора</h4>
                      <p className="text-xs text-muted-foreground">Складские процедуры, фото, дефектовка</p>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-muted-foreground" />
                </button>

                <button
                  onClick={() => {
                    handleTabChange('guides');
                    handleSelectDoc('manager');
                  }}
                  className="p-4 rounded-xl border border-border bg-muted/40 hover:bg-muted text-left transition-all flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-500">
                      <ShoppingCart className="w-5 h-5" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-foreground text-sm">Инструкция для менеджера</h4>
                      <p className="text-xs text-muted-foreground">Заказы, клиенты, статусы, списание деталей</p>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-muted-foreground" />
                </button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
