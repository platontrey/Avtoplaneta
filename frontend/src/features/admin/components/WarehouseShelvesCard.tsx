/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Layers, Plus, Trash2, Search } from 'lucide-react';
import { useShelves } from '@/hooks/useShelves';

export const WarehouseShelvesCard: React.FC = () => {
  const { shelves, isLoading, createShelf, isCreating, deleteShelf, isDeleting } = useShelves();
  const [newShelfName, setNewShelfName] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleAddShelf = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = newShelfName.trim();
    if (!trimmed) return;

    setError(null);
    try {
      await createShelf(trimmed);
      setNewShelfName('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось добавить полку');
    }
  };

  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(`Удалить полку "${name}"?`)) return;
    setError(null);
    try {
      await deleteShelf(id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить полку');
    }
  };

  const filteredShelves = shelves.filter((s) =>
    s.name.toLowerCase().includes(searchQuery.trim().toLowerCase()),
  );

  return (
    <Card className="mt-8">
      <CardHeader>
        <CardTitle className="flex items-center">
          <Layers className="w-5 h-5 mr-2" />
          Полки склада / Места хранения ({shelves.length})
        </CardTitle>
        <CardDescription>
          Управление списком полок для выпадающего списка на сайте и в мобильном приложении
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <div className="mb-4 p-3 rounded-md bg-destructive/15 text-destructive text-sm font-medium">
            {error}
          </div>
        )}

        {/* Форма добавления новой полки */}
        <form onSubmit={handleAddShelf} className="flex flex-col sm:flex-row gap-3 mb-6">
          <Input
            value={newShelfName}
            onChange={(e) => setNewShelfName(e.target.value)}
            placeholder="Введите название полки (например: Стеллаж 1 полка 2, Ряд А, 10-Б)"
            className="flex-1"
            disabled={isCreating}
          />
          <Button
            type="submit"
            disabled={!newShelfName.trim() || isCreating}
            className="whitespace-nowrap"
          >
            <Plus className="w-4 h-4 mr-1.5" />
            {isCreating ? 'Добавление...' : 'Добавить полку'}
          </Button>
        </form>

        {/* Поиск и список полок */}
        <div className="space-y-3">
          {shelves.length > 5 && (
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Поиск среди добавленных полок..."
                className="pl-9"
              />
            </div>
          )}

          {isLoading ? (
            <div className="text-sm text-muted-foreground py-4 text-center">
              Загрузка списка полок...
            </div>
          ) : shelves.length === 0 ? (
            <div className="text-sm text-muted-foreground py-6 text-center border border-dashed rounded-lg">
              Пока не добавлено ни одной полки. Введите название выше, чтобы добавить первую.
            </div>
          ) : filteredShelves.length === 0 ? (
            <div className="text-sm text-muted-foreground py-4 text-center">
              Ничего не найдено по запросу «{searchQuery}»
            </div>
          ) : (
            <div className="max-h-72 overflow-y-auto divide-y border rounded-lg">
              {filteredShelves.map((shelf) => (
                <div
                  key={shelf.id}
                  className="flex items-center justify-between px-4 py-2.5 hover:bg-muted/50 transition-colors"
                >
                  <span className="font-medium text-sm text-foreground">{shelf.name}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDelete(shelf.id, shelf.name)}
                    disabled={isDeleting}
                    className="text-muted-foreground hover:text-destructive h-8 w-8 p-0"
                    title="Удалить полку"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};
