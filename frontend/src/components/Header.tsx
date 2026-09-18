/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React from 'react';
import styled from 'styled-components';
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
 * Стилизованный контейнер заголовка
 */
const HeaderContainer = styled.header`
  border-bottom: 1px solid #d1d5db;

  .dark & {
    border-bottom: 1px solid oklch(1 0 0 / 12%);
  }
`;

/**
 * Стилизованный контейнер содержимого заголовка
 */
const HeaderContent = styled.div`
  max-width: 80rem;
  margin: 0 auto;
  padding: 1rem 1rem 1rem 1.75rem;
  display: flex;
  align-items: center;
  justify-content: space-between;

  @media (min-width: 640px) {
    padding: 1rem 1.75rem;
  }
`;

/**
 * Стилизованный контейнер левой части заголовка
 */
const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;

  @media (min-width: 640px) {
    gap: 1rem;
  }
`;

/**
 * Стилизованный контейнер правой части заголовка
 */
const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

/**
 * Стилизованный логотип
 */
const LogoContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

/**
 * Стилизованный контейнер логотипа с иконкой
 */
const LogoIcon = styled.div`
  border-radius: 0.375rem;
  border: 1px solid #d1d5db;
  padding: 0.5rem;

  .dark & {
    border: 1px solid oklch(1 0 0 / 12%);
  }
`;

/**
 * Стилизованный заголовок
 */
const HeaderTitle = styled.h1`
  font-size: 1.5rem;
  font-weight: 700;

  @media (min-width: 640px) {
    font-size: 2rem;
  }
`;

/**
 * Стилизованный контейнер навигационных кнопок
 */
const NavigationContainer = styled.div`
  display: none;

  @media (min-width: 640px) {
    display: flex;
    gap: 0.25rem;
  }
`;

/**
 * Стилизованная навигационная кнопка
 */
const NavButton = styled(Link)<{ $isActive: boolean }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  white-space: nowrap;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  outline: none;
  height: 2.5rem;
  padding: 0.5rem 1.5rem;
  width: 8rem;

  &:focus-visible {
    border-color: hsl(var(--ring));
    box-shadow: 0 0 0 3px hsl(var(--ring) / 0.5);
  }

  svg {
    pointer-events: none;
    height: 1rem;
    width: 1rem;
    flex-shrink: 0;
  }

  ${props => props.$isActive ? `
    background-color: #000;
    color: #fff;
    box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);

    .dark & {
      background-color: #222222;
      color: #fff;
    }

    &:hover {
      background-color: #000;
      color: #fff;

      .dark & {
        background-color: #222222;
        color: #fff;
      }
    }
  ` : `
    background-color: transparent;
    box-shadow: none;

    &:hover {
      background-color: #000;
      color: #fff;

      .dark & {
        background-color: #222222;
        color: #fff;
      }
    }
  `}
`;

/**
 * Стилизованный контейнер мобильной навигации
 */
const MobileNavigation = styled.div`
  @media (min-width: 640px) {
    display: none;
  }
`;

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

  return (
    <HeaderContainer>
      <HeaderContent>
        <HeaderLeft>
          <LogoContainer>
            <LogoIcon>
              <Package size={24} />
            </LogoIcon>
            <HeaderTitle>Автопланета</HeaderTitle>
          </LogoContainer>

          <NavigationContainer>
            <NavButton to="/inventory" $isActive={isActivePath('/inventory')}>
              <Package />
              Инвентарь
            </NavButton>
            <NavButton to="/statistics" $isActive={isActivePath('/statistics')}>
              <BarChart3 />
              Статистика
            </NavButton>
            <NavButton to="/messages" $isActive={isActivePath('/messages')}>
              <MessageCircle />
              Сообщения
            </NavButton>
            <NavButton to="/add-car" $isActive={isActivePath('/add-car')} style={{ paddingLeft: '3rem', paddingRight: '3rem' }}>
              <Plus />
              Добавить
            </NavButton>
            <NavButton to="/orders" $isActive={isActivePath('/orders')}>
              <Package />
              Заказы
            </NavButton>
          </NavigationContainer>

          <MobileNavigation>
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
          </MobileNavigation>
        </HeaderLeft>


        <HeaderRight>
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
                  <>
                    <DropdownMenuItem asChild>
                      <Link to="/admin" className="flex items-center">
                        <Settings className="w-4 h-4 mr-2" />
                        Админ панель
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link to="/readme" className="flex items-center">
                        <BookOpen className="w-4 h-4 mr-2" />
                        Документация
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                  </>
                )}
                {(user.role === 'operator' || user.role === 'manager') && <DropdownMenuSeparator />}
                <DropdownMenuItem onClick={onLogout} className="text-red-600 dark:text-red-400">
                  <LogOut className="w-4 h-4 mr-2" />
                  Выйти
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </HeaderRight>

      </HeaderContent>
    </HeaderContainer>
  );
};

export default Header;
