import React, { useState, useRef, useEffect } from 'react';
import { useCustomers } from '@/hooks/useCustomers';
import type { CustomerWithStats } from '@/lib/types';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { CATEGORY_LABELS } from './CustomersTab';
import { AlertTriangle, UserCheck, X } from 'lucide-react';

interface CustomerSelectInputProps {
  value: string;
  onChange: (value: string) => void;
  onSelectCustomer: (customer: CustomerWithStats | null) => void;
  selectedCustomer: CustomerWithStats | null;
  placeholder?: string;
  required?: boolean;
}

export const CustomerSelectInput: React.FC<CustomerSelectInputProps> = ({
  value,
  onChange,
  onSelectCustomer,
  selectedCustomer,
  placeholder = '+7 999 ... или имя клиента',
  required = false,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const query = value.trim();
  const { data } = useCustomers({
    q: query.length >= 2 ? query : '',
    limit: 6,
  });

  const matchingCustomers = data?.customers || [];

  // Закрытие дропдауна при клике вне компонента
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (c: CustomerWithStats) => {
    onSelectCustomer(c);
    onChange(c.phone || c.name);
    setIsOpen(false);
  };

  const handleClear = () => {
    onSelectCustomer(null);
    onChange('');
  };

  return (
    <div ref={containerRef} className="space-y-1.5 relative">
      <div className="relative">
        <Input
          value={value}
          onChange={(e) => {
            onChange(e.target.value);
            if (selectedCustomer && e.target.value !== (selectedCustomer.phone || selectedCustomer.name)) {
              onSelectCustomer(null);
            }
            setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          placeholder={placeholder}
          required={required}
          className={selectedCustomer ? 'pr-8 font-medium' : ''}
        />

        {selectedCustomer ? (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            title="Очистить выбор клиента"
          >
            <X className="h-4 w-4" />
          </button>
        ) : null}
      </div>

      {/* Индикатор выбранного клиента */}
      {selectedCustomer && (
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground px-1">
          <UserCheck className="h-3.5 w-3.5 text-emerald-600" />
          <span>Клиент #{selectedCustomer.id}: </span>
          <span className="font-semibold text-foreground">{selectedCustomer.name}</span>
          {selectedCustomer.city && <span>({selectedCustomer.city})</span>}
          {selectedCustomer.discount_percent > 0 && (
            <Badge variant="secondary" className="text-[10px] py-0 px-1 text-emerald-600">
              Скидка {selectedCustomer.discount_percent}%
            </Badge>
          )}
        </div>
      )}

      {/* Предупреждение о ЧС */}
      {selectedCustomer?.category === 'blacklist' && (
        <div className="p-2.5 bg-rose-500/10 border border-rose-500/30 rounded-lg text-xs text-rose-700 dark:text-rose-300 flex items-start gap-2">
          <AlertTriangle className="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
          <div>
            <strong>⚠️ Внимание: клиент в чёрном списке!</strong>
            <div className="mt-0.5">{selectedCustomer.notes || 'Проблемный клиент. Требуется осторожность.'}</div>
          </div>
        </div>
      )}

      {/* Выпадающий список подсказок */}
      {isOpen && query.length >= 2 && matchingCustomers.length > 0 && (
        <div className="absolute left-0 right-0 top-full mt-1 bg-popover text-popover-foreground border rounded-md shadow-lg z-50 overflow-hidden max-h-56 overflow-y-auto">
          <div className="text-[10px] text-muted-foreground uppercase px-2.5 py-1.5 bg-muted/60 font-semibold">
            Найдено в базе клиентов:
          </div>
          {matchingCustomers.map((c) => {
            const cat = CATEGORY_LABELS[c.category] || CATEGORY_LABELS.regular;
            return (
              <div
                key={c.id}
                onClick={() => handleSelect(c)}
                className="p-2 hover:bg-accent hover:text-accent-foreground cursor-pointer transition-colors border-b last:border-b-0 text-xs flex items-center justify-between gap-2"
              >
                <div>
                  <div className="font-semibold flex items-center gap-1.5">
                    <span>{c.name || c.phone}</span>
                    <Badge variant="outline" className={`text-[10px] py-0 px-1 ${cat.badgeClass}`}>
                      {cat.icon} {cat.label}
                    </Badge>
                  </div>
                  <div className="text-[11px] text-muted-foreground flex items-center gap-2 mt-0.5">
                    <span>{c.phone}</span>
                    {c.city && <span>· {c.city}</span>}
                    {c.preferred_tk && <span>· ТК: {c.preferred_tk}</span>}
                  </div>
                </div>

                <div className="text-right shrink-0">
                  {c.discount_percent > 0 && (
                    <div className="text-emerald-600 font-semibold text-[11px]">
                      -{c.discount_percent}%
                    </div>
                  )}
                  <div className="text-[10px] text-muted-foreground">
                    {c.total_orders} зак.
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
