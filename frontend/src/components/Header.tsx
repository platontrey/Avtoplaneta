/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Package, BarChart3, Plus, LogOut, User as UserIcon, Settings, ChevronDown, BookOpen, MessageCircle } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import ThemeToggle from './ThemeToggle';
import { cn } from "@/lib/utils";

import type { User } from '../features/auth/types';

/**
 * Интерфейс для пропсов компонента Header
 */
interface HeaderProps {
  /** Текущий пользователь */
  user?: User | null;
  /** Функция выхода из системы */
  onLogout?: () => void;
}

/**
 * Основной компонент Header
 * @param user - Текущий пользователь
 * @param onLogout - Функция выхода из системы
 * @returns JSX элемент заголовка
 */
const Header: React.FC<HeaderProps> = ({ user, onLogout }) => {
  const location = useLocation();

  /**
   * Определяет, активна ли текущая страница для данной кнопки
   * @param path - Путь страницы
   * @returns true если страница активна
   */
  const isActivePath = (path: string): boolean => {
    // Специальная обработка для инвентаря - оба пути "/" и "/inventory" активируют кнопку инвентаря
    if (path === '/inventory') {
      return location.pathname === '/' || location.pathname === '/inventory';
    }
    return location.pathname === path;
  };

  const navButtonClass = (isActive: boolean) =>
    cn(
      "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all outline-none h-10 px-6 w-32 focus-visible:ring-2 focus-visible:ring-ring [&>svg]:size-4 [&>svg]:shrink-0",
      isActive
        ? "bg-black text-white dark:bg-[#222222] shadow-sm hover:bg-black dark:hover:bg-[#222222]"
        : "bg-transparent text-foreground hover:bg-black hover:text-white dark:hover:bg-[#222222] dark:hover:text-white"
    );

  return (
    <header className="border-b border-gray-300 dark:border-white/12">
      <div className="max-w-7xl mx-auto px-4 sm:px-7 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2 sm:gap-4">
          <div className="flex items-center gap-2">
            <div className="rounded-md border border-gray-300 dark:border-white/12 p-2">
              <Package size={24} />
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold">Автопланета</h1>
          </div>

          <div className="hidden sm:flex gap-1">
            <Link to="/inventory" className={navButtonClass(isActivePath('/inventory'))}>
              <Package />
              Инвентарь
            </Link>
            <Link to="/statistics" className={navButtonClass(isActivePath('/statistics'))}>
              <BarChart3 />
              Статистика
            </Link>
            <Link to="/messages" className={navButtonClass(isActivePath('/messages'))}>
              <MessageCircle />
              Сообщения
            </Link>
            <Link to="/add-car" className={cn(navButtonClass(isActivePath('/add-car')), "!px-12")}>
              <Plus />
              Добавить
            </Link>
            <Link to="/orders" className={navButtonClass(isActivePath('/orders'))}>
              <Package />
              Заказы
            </Link>
          </div>

          <div className="sm:hidden">
            <DropdownMenu modal={false} onOpenChange={(open) => console.log('Mobile menu open state:', open)}>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" onClick={() => console.log('Mobile menu trigger clicked')}>
                  <ChevronDown className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" side="bottom" sideOffset={5} style={{ zIndex: 9999 }} onPointerDownOutside={(e) => e.preventDefault()}>
                <DropdownMenuItem asChild>
                  <Link to="/inventory" className="flex items-center">
                    <Package className="w-4 h-4 mr-2" />
                    Инвентарь
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link to="/statistics" className="flex items-center">
                    <BarChart3 className="w-4 h-4 mr-2" />
                    Статистика
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link to="/messages" className="flex items-center">
                    <MessageCircle className="w-4 h-4 mr-2" />
                    Сообщения
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link to="/add-car" className="flex items-center">
                    <Plus className="w-4 h-4 mr-2" />
                    Добавить
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link to="/orders" className="flex items-center">
                    <Package className="w-4 h-4 mr-2" />
                    Заказы
                  </Link>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <ThemeToggle />

          {/* User profile dropdown */}
          {user && (
            <DropdownMenu modal={false} onOpenChange={(open) => console.log('Profile menu open state:', open)}>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="flex items-center space-x-1 sm:space-x-2" onClick={() => console.log('Profile menu trigger clicked')}>
                  <UserIcon size={16} />
                  <span className="hidden sm:inline">{user.name}</span>
                  <ChevronDown className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" side="bottom" style={{ zIndex: 9999 }} onPointerDownOutside={(e) => e.preventDefault()}>
                <DropdownMenuLabel>Мой профиль</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <div className="px-2 py-1.5 text-sm">
                  <p className="font-medium">{user.name}</p>
                  <p className="text-muted-foreground">{user.email}</p>
                  {user.role === 'admin' && (
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300 mt-1">
                      Администратор
                    </span>
                  )}
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link to="/messages" className="flex items-center">
                    <MessageCircle className="w-4 h-4 mr-2" />
                    Сообщения
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link to="/readme" className="flex items-center">
                    <BookOpen className="w-4 h-4 mr-2" />
                    Документация и API
                  </Link>
                </DropdownMenuItem>
                {user.role === 'operator' && (
                  <DropdownMenuItem asChild>
                    <Link to="/operator-instructions" className="flex items-center">
                      <BookOpen className="w-4 h-4 mr-2" />
                      Инструкция оператора
                    </Link>
                  </DropdownMenuItem>
                )}
                {user.role === 'manager' && (
                  <DropdownMenuItem asChild>
                    <Link to="/manager-instructions" className="flex items-center">
                      <BookOpen className="w-4 h-4 mr-2" />
                      Инструкция менеджера
                    </Link>
                  </DropdownMenuItem>
                )}
                {user.role === 'admin' && (
                  <DropdownMenuItem asChild>
                    <Link to="/admin" className="flex items-center">
                      <Settings className="w-4 h-4 mr-2" />
                      Админ панель
                    </Link>
                  </DropdownMenuItem>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={onLogout} className="text-red-600 dark:text-red-400">
                  <LogOut className="w-4 h-4 mr-2" />
                  Выйти
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
