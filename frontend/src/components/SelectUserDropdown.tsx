/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect, useRef } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { User } from '../features/messaging/types';
import { messagingApi } from '../features/messaging/api/messagingApi';

interface SelectUserDropdownProps {
  selectedUsers: User[];
  onSelectionChange: (users: User[]) => void;
  placeholder?: string;
  multiple?: boolean;
  currentUser?: User;
}

export default function SelectUserDropdown({
  selectedUsers,
  onSelectionChange,
  placeholder = "Выберите пользователей",
  multiple = true,
  currentUser
}: SelectUserDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      try {
        const fetchedUsers = await messagingApi.getUsers();
        setUsers(fetchedUsers);
      } catch (error) {
        console.error('Failed to fetch users:', error);
      } finally {
        setLoading(false);
      }
    };

    if (isOpen) {
      fetchUsers();
    }
  }, [isOpen]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const filteredUsers = users.filter(user =>
    (user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.email.toLowerCase().includes(searchQuery.toLowerCase())) &&
    (!currentUser || user.id !== currentUser.id)
  );

  const handleUserToggle = (user: User) => {
    if (multiple) {
      const isSelected = selectedUsers.some(u => u.id === user.id);
      if (isSelected) {
        onSelectionChange(selectedUsers.filter(u => u.id !== user.id));
      } else {
        onSelectionChange([...selectedUsers, user]);
      }
    } else {
      onSelectionChange([user]);
      setIsOpen(false);
    }
  };

  const handleRemoveUser = (userId: number) => {
    onSelectionChange(selectedUsers.filter(u => u.id !== userId));
  };

  const displayText = selectedUsers.length > 0
    ? selectedUsers.map(u => u.name).join(', ')
    : placeholder;

  return (
    <div className="relative" ref={dropdownRef}>
      <Label htmlFor="user-select">Участники</Label>
      <div className="relative">
        <Button
          type="button"
          variant="outline"
          className="w-full justify-start text-left font-normal"
          onClick={() => setIsOpen(!isOpen)}
        >
          <span className="truncate">{displayText}</span>
        </Button>

        {isOpen && (
          <div className="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
            <div className="p-2">
              <Input
                placeholder="Поиск пользователей..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="mb-2"
              />
            </div>

            {loading ? (
              <div className="p-2 text-center text-gray-500">Загрузка...</div>
            ) : (
              <div className="max-h-40 overflow-auto">
                {filteredUsers.length === 0 ? (
                  <div className="p-2 text-center text-gray-500">Пользователи не найдены</div>
                ) : (
                  filteredUsers.map(user => {
                    const isSelected = selectedUsers.some(u => u.id === user.id);
                    return (
                      <div
                        key={user.id}
                        className={`p-2 cursor-pointer hover:bg-gray-100 ${
                          isSelected ? 'bg-blue-50' : ''
                        }`}
                        onClick={() => handleUserToggle(user)}
                      >
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-medium">{user.name}</div>
                            <div className="text-sm text-gray-500">{user.email}</div>
                          </div>
                          {isSelected && (
                            <span className="text-blue-600">✓</span>
                          )}
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {selectedUsers.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {selectedUsers.map(user => (
            <span
              key={user.id}
              className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-blue-100 text-blue-800"
            >
              {user.name}
              <button
                type="button"
                onClick={() => handleRemoveUser(user.id)}
                className="ml-1 text-blue-600 hover:text-blue-800"
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}