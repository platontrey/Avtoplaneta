/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useEffect, useCallback, useRef } from 'react';
import { Search, Loader2, X, ChevronDown } from 'lucide-react';
import type { Part } from '../lib/types';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Button } from './ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import { API_BASE_URL } from '@/lib/api';

interface PartSelectorProps {
    value?: string;
    onChange: (partId: string, part?: Part) => void;
    placeholder?: string;
    disabled?: boolean;
}

function PartSelector({ value, onChange, placeholder = "Выберите запчасть", disabled = false }: PartSelectorProps) {
    console.log('PartSelector rendered with value:', value);
    const [searchQuery, setSearchQuery] = useState('');
    const [isSearching, setIsSearching] = useState(false);
    const [isResultsVisible, setIsResultsVisible] = useState(false);
    const [results, setResults] = useState<Part[]>([]);
    const [popularParts, setPopularParts] = useState<Part[]>([]);
    const [selectedPart, setSelectedPart] = useState<Part | null>(null);
    const searchInputRef = useRef<HTMLInputElement>(null);
    const resultsRef = useRef<HTMLDivElement>(null);
    const abortControllerRef = useRef<AbortController | null>(null);
    const prevResultsRef = useRef<Part[]>([]);
    const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

    const MIN_SEARCH_LENGTH = 2;
    const SEARCH_DEBOUNCE_DELAY = 300;

    // Поиск запчастей
    const performSearch = useCallback(async (query: string) => {
        console.log('PartSelector - performSearch called with query:', query);
        if (query.length < MIN_SEARCH_LENGTH) {
            setResults([]);
            setIsResultsVisible(false);
            return;
        }

        setResults([]);
        setIsResultsVisible(false);

        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
        }

        abortControllerRef.current = new AbortController();
        setIsSearching(true);

        try {
            const url = `${API_BASE_URL}/api/v1/parts/inventory?search=${encodeURIComponent(query)}&limit=20`;
            const response = await fetch(url, {
                headers: getAuthHeaders(),
                credentials: 'include',
                signal: abortControllerRef.current.signal
            });

            if (!response.ok) {
                console.error(`HTTP error! status: ${response.status}`);
                setIsSearching(false);
                return;
            }

            const data = await response.json();
            const parts = data.parts || [];

            console.log('PartSelector - Search results:', parts.length, 'parts found');
            console.log('PartSelector - Raw response data:', data);
            console.log('PartSelector - Parts array:', parts);
            console.log('PartSelector - Response data type:', typeof data);
            console.log('PartSelector - Data.parts type:', typeof data.parts);

            const resultsChanged = JSON.stringify(parts) !== JSON.stringify(prevResultsRef.current);

            if (resultsChanged) {
                setResults(parts);
                prevResultsRef.current = parts;
            }

            setIsResultsVisible(parts.length > 0);
        }
        catch (error: unknown) {
            if (error instanceof Error && error.name !== 'AbortError') {
                console.error('Ошибка поиска:', error);
                setResults([]);
                setIsResultsVisible(false);
            }
        } finally {
            setIsSearching(false);
        }
    }, []);

    // Debounce для поиска
    useEffect(() => {
        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }

        if (searchQuery.length >= MIN_SEARCH_LENGTH) {
            console.log('Starting search for:', searchQuery);
            debounceTimerRef.current = setTimeout(() => {
                performSearch(searchQuery).catch(console.error);
            }, SEARCH_DEBOUNCE_DELAY);
        } else {
            // Если запрос слишком короткий, скрываем результаты поиска
            console.log('Query too short, hiding results');
            setResults([]);
            setIsResultsVisible(false);
        }

        return () => {
            if (debounceTimerRef.current) {
                clearTimeout(debounceTimerRef.current);
            }
        };
    }, [searchQuery, performSearch]);

    // Загрузка популярных запчастей при монтировании
    useEffect(() => {
        const loadPopularParts = async () => {
            try {
                const response = await fetch(`${API_BASE_URL}/api/v1/parts/inventory?limit=5`, {
                    headers: getAuthHeaders(),
                    credentials: 'include'
                });
                if (response.ok) {
                    const data = await response.json();
                    setPopularParts(data.parts || []);
                }
            } catch (error) {
                console.error('Ошибка загрузки популярных запчастей:', error);
            }
        };

        if (!value) {
            loadPopularParts().catch(console.error);
        }
    }, [value]);

    // Очистка при размонтировании
    useEffect(() => {
        return () => {
            if (abortControllerRef.current) {
                abortControllerRef.current.abort();
            }
            if (debounceTimerRef.current) {
                clearTimeout(debounceTimerRef.current);
            }
        };
    }, []);

    // Загрузка выбранной запчасти при изменении value
    useEffect(() => {
        if (value && !selectedPart) {
            // Загружаем информацию о выбранной запчасти
            const loadSelectedPart = async () => {
                try {
                    const response = await fetch(`${API_BASE_URL}/api/v1/parts/${value}`, {
                        headers: getAuthHeaders(),
                        credentials: 'include'
                    });
                    if (response.ok) {
                        const part = await response.json();
                        setSelectedPart(part);
                        setSearchQuery(part.name);
                    }
                } catch (error) {
                    console.error('Ошибка загрузки запчасти:', error);
                }
            };
            loadSelectedPart().catch(console.error);
        } else if (!value) {
            setSelectedPart(null);
            setSearchQuery('');
        }
    }, [value, selectedPart]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newQuery = e.target.value;
        setSearchQuery(newQuery);

        // Если очистили поле, сбрасываем выбор
        if (!newQuery) {
            setSelectedPart(null);
            onChange('', undefined);
            // Показываем популярные запчасти при очистке
            if (popularParts.length > 0) {
                setIsResultsVisible(true);
            }
        }
    };

    const handleResultClick = (result: Part) => {
        setSelectedPart(result);
        setSearchQuery(result.name);
        setIsResultsVisible(false);
        setResults([]);
        onChange(result.id.toString(), result);
        // Возвращаем фокус на input после выбора
        setTimeout(() => {
            searchInputRef.current?.blur();
        }, 0);
    };

    const clearSelection = () => {
        setSelectedPart(null);
        setSearchQuery('');
        setResults([]);
        setIsResultsVisible(false);
        onChange('', undefined);
        // Возвращаем фокус на input после очистки
        setTimeout(() => {
            searchInputRef.current?.focus();
        }, 0);
    };

    const handleDropdownToggle = () => {
        if (disabled) return;

        const newVisibility = !isResultsVisible;
        setIsResultsVisible(newVisibility);

        // Если открываем и нет результатов, показываем популярные
        if (newVisibility && results.length === 0 && !searchQuery) {
            // Популярные уже загружены через useEffect
        }
    };

    return (
        <div className="space-y-2" onClick={(e) => e.stopPropagation()}>
            <Label htmlFor="part-selector">Запчасть</Label>
            <div className="relative">
                <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground pointer-events-none z-10">
                    {isSearching ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                        <Search className="w-4 h-4" />
                    )}
                </div>
                <Input
                    ref={searchInputRef}
                    id="part-selector"
                    type="text"
                    value={searchQuery}
                    onChange={handleInputChange}
                    onFocus={() => {
                        if (results.length > 0) {
                            setIsResultsVisible(true);
                        } else if (!searchQuery && popularParts.length > 0) {
                            setIsResultsVisible(true);
                        }
                    }}
                    placeholder={placeholder}
                    className="pl-10 pr-10"
                    disabled={disabled}
                    autoComplete="off"
                />
                {searchQuery && (
                    <Button
                        onClick={clearSelection}
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-1 top-1/2 transform -translate-y-1/2 h-7 w-7 p-0 z-10"
                        aria-label="Очистить выбор"
                        disabled={disabled}
                    >
                        <X className="h-4 w-4" />
                    </Button>
                )}
                {!searchQuery && (
                    <Button
                        onClick={handleDropdownToggle}
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-1 top-1/2 transform -translate-y-1/2 h-7 w-7 p-0 z-10"
                        aria-label="Показать список"
                        disabled={disabled}
                    >
                        <ChevronDown className={`h-4 w-4 transition-transform duration-200 ${isResultsVisible ? 'rotate-180' : ''}`} />
                    </Button>
                )}

                {/* Предупреждение о минимальной длине поиска */}
                {searchQuery.length > 0 && searchQuery.length < MIN_SEARCH_LENGTH && !isResultsVisible && (
                    <div className="absolute left-0 top-full mt-1 z-10">
                        <p className="text-amber-600 text-xs flex items-center gap-1">
                            <span>⚠️</span>
                            Минимум {MIN_SEARCH_LENGTH} символа для поиска
                        </p>
                    </div>
                )}

                {/* Результаты поиска */}
                <AnimatePresence>
                    {isResultsVisible && (
                        <motion.div
                            ref={resultsRef}
                            initial={{ opacity: 0, y: -10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            transition={{ duration: 0.15, ease: 'easeOut' }}
                            className="absolute left-0 right-0 mt-1 bg-popover border rounded-md shadow-lg z-50 max-h-60 overflow-y-auto"
                        >
                            {/* Популярные запчасти */}
                            {!searchQuery && popularParts.length > 0 && (
                                <div className="p-2 border-b border-border">
                                    <div className="text-xs font-medium text-muted-foreground px-1 mb-2">Популярные запчасти:</div>
                                    {popularParts.slice(0, 5).map((part) => (
                                        <button
                                            key={part.id}
                                            type="button"
                                            className="w-full text-left p-2 hover:bg-accent cursor-pointer rounded transition-colors duration-150 mb-1"
                                            onClick={() => handleResultClick(part)}
                                        >
                                            <div className="font-medium text-sm">{part.name}</div>
                                            <div className="text-xs text-muted-foreground">
                                                {part.brand} {part.model} • {part.quantity} шт.
                                            </div>
                                        </button>
                                    ))}
                                </div>
                            )}

                            {/* Результаты поиска */}
                            {results.length > 0 && !isSearching && (
                                <div>
                                    {results.map((result) => (
                                        <button
                                            key={result.id}
                                            type="button"
                                            className="w-full text-left p-3 hover:bg-accent cursor-pointer border-b border-border last:border-b-0 transition-colors duration-150"
                                            onClick={() => handleResultClick(result)}
                                        >
                                            <div className="font-medium text-sm">{result.name}</div>
                                            <div className="text-xs text-muted-foreground mt-0.5">
                                                {result.brand} - {result.model} • Кол-во: {result.quantity} • {result.location || 'Нет локации'}
                                            </div>
                                            {result.description && (
                                                <div className="text-xs text-muted-foreground truncate mt-0.5">{result.description}</div>
                                            )}
                                        </button>
                                    ))}
                                </div>
                            )}

                            {/* Загрузка */}
                            {isSearching && (
                                <div className="p-4 text-center text-muted-foreground">
                                    <Loader2 className="w-4 h-4 animate-spin mx-auto mb-2" />
                                    <div className="text-sm">Поиск...</div>
                                </div>
                            )}

                            {/* Нет результатов */}
                            {!isSearching && searchQuery && results.length === 0 && searchQuery.length >= MIN_SEARCH_LENGTH && (
                                <div className="p-4 text-center text-muted-foreground">
                                    <div className="text-sm">Ничего не найдено</div>
                                    <div className="text-xs mt-1">Попробуйте изменить запрос</div>
                                </div>
                            )}

                            {/* Пустое состояние */}
                            {!isSearching && !searchQuery && popularParts.length === 0 && (
                                <div className="p-4 text-center text-muted-foreground">
                                    <div className="text-sm">Нет доступных запчастей</div>
                                </div>
                            )}
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>

            {/* Информация о выбранной запчасти */}
            {selectedPart && (
                <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    className="text-xs text-muted-foreground bg-muted/50 p-2 rounded border border-border"
                >
                    <span className="font-medium">Выбрано:</span> {selectedPart.name} ({selectedPart.brand} {selectedPart.model}) - <span className="font-medium">Доступно:</span> {selectedPart.quantity} шт.
                </motion.div>
            )}
        </div>
    );
}

export default PartSelector;