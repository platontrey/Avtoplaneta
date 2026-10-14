/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import operatorMd from '../../docs/operator-instructions.md?raw';
import managerMd from '../../docs/manager-instructions.md?raw';
import archMd from '../../docs/ARCHITECTURE.md?raw';
export { archMd };
import traefikMd from '../../docs/TRAEFIK-README.md?raw';
import monitoringMd from '../../docs/MONITORING-README.md?raw';
import mobileMd from '../../docs/mobile-README.md?raw';
import frontendMd from '../../docs/frontend-README.md?raw';
import migrationMd from '../../docs/migration-README.md?raw';
import rootReadmeMd from '../../docs/root-README.md?raw';

import {
  Package,
  ShoppingCart,
  Server,
  Shield,
  Activity,
  Smartphone,
  Code2,
  Database,
  BookOpen,
} from 'lucide-react';

export interface RepoDocument {
  id: string;
  title: string;
  category: 'Инструкции персонала' | 'Архитектура и бэкенд' | 'Инфраструктура и DevOps' | 'Клиенты и приложения' | 'Общие сведения';
  filePath: string;
  icon: React.ElementType;
  summary: string;
  badge: string;
  content: string;
}

export const REPO_DOCS: RepoDocument[] = [
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
];
