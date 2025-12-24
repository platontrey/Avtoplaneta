import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Package, BarChart3, MessageCircle, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { haptic } from '@/lib/haptic';

const MobileBottomNav: React.FC = () => {
  const location = useLocation();

  const navItems = [
    {
      to: '/inventory',
      icon: Package,
      label: 'Инвентарь',
      active: location.pathname === '/' || location.pathname === '/inventory'
    },
    {
      to: '/statistics',
      icon: BarChart3,
      label: 'Статистика',
      active: location.pathname === '/statistics'
    },
    {
      to: '/messages',
      icon: MessageCircle,
      label: 'Сообщения',
      active: location.pathname === '/messages'
    },
    {
      to: '/add-car',
      icon: Plus,
      label: 'Добавить',
      active: location.pathname === '/add-car'
    },
    {
      to: '/orders',
      icon: Package,
      label: 'Заказы',
      active: location.pathname === '/orders'
    }
  ];

  return (
    <div className="md:hidden fixed bottom-0 left-0 right-0 bg-background border-t border-border z-50">
      <div className="flex items-center justify-around px-2 py-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.to}
              to={item.to}
              onClick={() => haptic.selection()}
              className={cn(
                "flex flex-col items-center justify-center p-3 rounded-lg transition-colors min-w-0 flex-1 min-h-[44px]",
                item.active
                  ? "text-primary bg-primary/10"
                  : "text-muted-foreground hover:text-foreground hover:bg-accent"
              )}
            >
              <Icon className="h-5 w-5 mb-1" />
              <span className="text-xs font-medium truncate">{item.label}</span>
            </Link>
          );
        })}
      </div>
    </div>
  );
};

export default MobileBottomNav;