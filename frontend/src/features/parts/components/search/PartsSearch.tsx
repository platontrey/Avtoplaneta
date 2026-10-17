/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { Search, Loader2, X, Filter, ChevronDown, ChevronUp, Mic, MicOff, Car, Wrench, GitFork, Package, Disc } from 'lucide-react';
import type { Part } from '@/lib/types';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { SelectOption } from '@/components/ui/searchable-select';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { motion, AnimatePresence } from 'framer-motion';
import { API_BASE_URL } from '@/lib/api';
import { useVehicleOptions } from '@/features/vehicles/useVehicleCatalog';
import { usePartCatalog } from '@/features/catalog/usePartCatalog';
import type { PartFilters } from '@/hooks/useParts';

import { ActiveFilterBadges, type ActiveFilterItem } from './ActiveFilterBadges';
import { GeneralFiltersTab } from './GeneralFiltersTab';
import { BodyEngineFiltersTab } from './BodyEngineFiltersTab';
import { TransmissionFiltersTab } from './TransmissionFiltersTab';
import { WarehouseFiltersTab } from './WarehouseFiltersTab';
import { WheelsFiltersTab } from './WheelsFiltersTab';
import {
  type SearchFiltersState,
  initialSearchFiltersState,
  conditionOptions,
  statusOptions,
  photoOptions,
  transmissionOptions,
  driveOptions,
  frontRearOptions,
  leftRightOptions,
  topBottomOptions,
  seasonOptions,
  FILTER_DESCRIPTORS,
  toPartFilters,
} from './searchTypes';

// Web Speech API Types
declare global {
  interface Window {
    SpeechRecognition: typeof SpeechRecognition;
    webkitSpeechRecognition: typeof SpeechRecognition;
  }
}

interface SpeechRecognition extends EventTarget {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  start(): void;
  stop(): void;
  abort(): void;
  onstart: ((this: SpeechRecognition, ev: Event) => void) | null;
  onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => void) | null;
  onend: ((this: SpeechRecognition, ev: Event) => void) | null;
  onerror: ((this: SpeechRecognition, ev: SpeechRecognitionErrorEvent) => void) | null;
}

interface SpeechRecognitionEvent extends Event {
  results: SpeechRecognitionResultList;
  resultIndex: number;
}

interface SpeechRecognitionErrorEvent extends Event {
  error: string;
  message: string;
}

interface SpeechRecognitionResultList {
  readonly length: number;
  item(index: number): SpeechRecognitionResult;
  [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
  readonly length: number;
  item(index: number): SpeechRecognitionAlternative;
  [index: number]: SpeechRecognitionAlternative;
  isFinal: boolean;
}

interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

// eslint-disable-next-line no-var
declare var SpeechRecognition: {
  prototype: SpeechRecognition;
  new(): SpeechRecognition;
};

export interface PartsSearchProps {
  onSearchChange?: (searchQuery: string) => void;
  onFiltersChange?: (filters: PartFilters) => void;
  onDisplayLimitChange?: (limit: number | undefined) => void;
  currentDisplayLimit?: number | undefined;
}

const MIN_SEARCH_LENGTH = 2;
const SEARCH_DEBOUNCE_DELAY = 500;
const FILTER_DEBOUNCE_DELAY = 300;

export const PartsSearch: React.FC<PartsSearchProps> = ({
  onFiltersChange,
  onDisplayLimitChange,
  currentDisplayLimit,
}) => {
  const [filters, setFilters] = useState<SearchFiltersState>(initialSearchFiltersState);

  const { brandOptions, modelOptions, bodyOptions, engineOptions } = useVehicleOptions(
    filters.brand,
    filters.model,
  );

  const [isSearching, setIsSearching] = useState(false);
  const [isResultsVisible, setIsResultsVisible] = useState(false);
  const [isFiltersOpen, setIsFiltersOpen] = useState(false);
  const [results, setResults] = useState<Part[]>([]);
  const [isListening, setIsListening] = useState(false);
  const [voiceSearchError, setVoiceSearchError] = useState<string | null>(null);

  const searchInputRef = useRef<HTMLInputElement>(null);
  const resultsRef = useRef<HTMLDivElement>(null);
  const abortControllerRef = useRef<AbortController | null>(null);
  const recognitionRef = useRef<SpeechRecognition | null>(null);
  const prevResultsRef = useRef<Part[]>([]);

  const { data: partCatalog } = usePartCatalog();

  const categoryOptions: SelectOption[] = useMemo(() => {
    const catalogCategories = partCatalog?.part_form_categories?.map((c) => c.name) ?? [];
    const allCategories = Array.from(
      new Set([
        ...catalogCategories,
        'Выхлопная система',
        'Двигатель',
        'Диски и шины',
        'Кузов',
        'Кузов внутри',
        'Кузов снаружи',
        'Оптика',
        'Пневмосистема',
        'Подвеска',
        'Подвеска ДВС/КПП',
        'Подвеска задних колес',
        'Подвеска передних колес',
        'Рулевое управление',
        'Система кондиционирования',
        'Система охлаждения и отопления',
        'Система выхлопа (Глушитель)',
        'Система рулевого управления',
        'Система фильтрации (Фильтры)',
        'Сопутствующие товары',
        'Стекла',
        'Тормоза',
        'Тормозная система',
        'Трансмиссия',
        'Шины и диски',
        'Электрика',
        'Электрооснащение',
        'Автохимия и масла',
        'Аксессуары и тюннинг',
        'Интерьер',
        'Другое',
      ])
    )
      .filter(Boolean)
      .sort((a, b) => a.localeCompare(b, 'ru'));

    return allCategories.map((cat) => ({ value: cat, label: cat }));
  }, [partCatalog]);

  const updateFilter = useCallback(
    <K extends keyof SearchFiltersState>(key: K, val: SearchFiltersState[K]) => {
      setFilters((prev) => (prev[key] === val ? prev : { ...prev, [key]: val }));
    },
    []
  );

  const clearResults = useCallback(() => {
    setResults([]);
    prevResultsRef.current = [];
    setIsResultsVisible(false);
  }, []);

  const clearSearchQuery = useCallback(() => {
    updateFilter('searchQuery', '');
    clearResults();
    setVoiceSearchError(null);
  }, [updateFilter, clearResults]);

  const clearFilters = useCallback(() => {
    setFilters(initialSearchFiltersState);
    clearResults();
  }, [clearResults]);

  const performSearch = useCallback(async (query: string) => {
    if (query.length < MIN_SEARCH_LENGTH) {
      clearResults();
      return;
    }

    clearResults();

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    abortControllerRef.current = new AbortController();
    setIsSearching(true);

    try {
      const url = `${API_BASE_URL}/api/v1/inventory?search=${encodeURIComponent(query)}&limit=10`;
      const response = await fetch(url, {
        credentials: 'include',
        signal: abortControllerRef.current.signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      const parts = data.parts || [];

      const resultsChanged =
        parts.length !== prevResultsRef.current.length ||
        parts.some((part: Part, index: number) => part.id !== prevResultsRef.current[index]?.id);

      if (resultsChanged) {
        setResults(parts);
        prevResultsRef.current = parts;
      }

      setIsResultsVisible(parts.length > 0);
    } catch (error: unknown) {
      if (error instanceof Error && error.name !== 'AbortError') {
        console.error('Ошибка поиска:', error);
        clearResults();
      }
    } finally {
      setIsSearching(false);
    }
  }, [clearResults]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      performSearch(filters.searchQuery);
    }, SEARCH_DEBOUNCE_DELAY);

    return () => clearTimeout(timeoutId);
  }, [filters.searchQuery, performSearch]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (onFiltersChange) {
        onFiltersChange(toPartFilters(filters));
      }
    }, FILTER_DEBOUNCE_DELAY);

    return () => clearTimeout(timeoutId);
  }, [filters, onFiltersChange]);

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      if (recognitionRef.current) {
        recognitionRef.current.abort();
      }
    };
  }, []);

  const isSpeechRecognitionSupported = () => {
    return 'SpeechRecognition' in window || 'webkitSpeechRecognition' in window;
  };

  const startVoiceSearch = useCallback(() => {
    if (!isSpeechRecognitionSupported()) {
      setVoiceSearchError('Ваш браузер не поддерживает голосовой поиск');
      return;
    }

    setVoiceSearchError(null);

    const SpeechRecognitionClass = window.SpeechRecognition || window.webkitSpeechRecognition;
    recognitionRef.current = new SpeechRecognitionClass();
    recognitionRef.current.lang = 'ru-RU';
    recognitionRef.current.continuous = false;
    recognitionRef.current.interimResults = false;

    recognitionRef.current.onstart = () => {
      setIsListening(true);
    };

    recognitionRef.current.onresult = (event: SpeechRecognitionEvent) => {
      const transcript = event.results[0][0].transcript;
      updateFilter('searchQuery', transcript);
      setIsListening(false);
    };

    recognitionRef.current.onend = () => {
      setIsListening(false);
    };

    recognitionRef.current.onerror = (event: SpeechRecognitionErrorEvent) => {
      setIsListening(false);
      setVoiceSearchError(`Ошибка распознавания речи: ${event.error}`);
    };

    try {
      recognitionRef.current.start();
    } catch {
      setVoiceSearchError('Не удалось начать распознавание речи');
      setIsListening(false);
    }
  }, [updateFilter]);

  const stopVoiceSearch = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.stop();
    }
    setIsListening(false);
  }, []);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    updateFilter('searchQuery', e.target.value);
  }, [updateFilter]);

  const handleResultClick = (result: Part) => {
    updateFilter('searchQuery', result.name);
    setIsResultsVisible(false);
    setResults([]);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        searchInputRef.current &&
        !searchInputRef.current.contains(event.target as Node) &&
        resultsRef.current &&
        !resultsRef.current.contains(event.target as Node)
      ) {
        setIsResultsVisible(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Подсчет активных фильтров
  const activeFiltersCount = useMemo(() => {
    return [
      filters.category,
      filters.brand,
      filters.model,
      filters.location,
      filters.address,
      filters.status,
      filters.hasPhoto && filters.hasPhoto !== 'with' ? filters.hasPhoto : '',
      filters.number,
      filters.oemCode,
      filters.vin,
      filters.carReleasePeriod,
      filters.bodyBrand,
      filters.engineBrand,
      filters.carReleaseDate,
      filters.transmission,
      filters.drive,
      filters.condition,
      filters.manufacturer,
      filters.defect,
      filters.color,
      filters.salesman,
      filters.minPrice,
      filters.maxPrice,
      filters.minQuantity,
      filters.maxQuantity,
      filters.frontRear,
      filters.leftRight,
      filters.topBottom,
      filters.manufacturerCode,
      filters.supplierCode,
      filters.transmissionModel,
      filters.wearPercentage,
      filters.season,
      filters.diameter,
      filters.width,
      filters.profile,
      filters.tireQuantity,
      filters.drilling,
      filters.offset,
      filters.centerHoleDiameter,
      filters.tireModel,
    ].filter(Boolean).length;
  }, [filters]);

  const mainFiltersCount = useMemo(() => {
    return [
      filters.category,
      filters.brand,
      filters.model,
      filters.carReleaseDate,
      filters.condition,
      filters.minPrice,
      filters.maxPrice,
      filters.minQuantity,
      filters.maxQuantity,
      filters.status,
      filters.hasPhoto && filters.hasPhoto !== 'with' ? filters.hasPhoto : '',
    ].filter(Boolean).length;
  }, [filters]);

  const bodyEngineFiltersCount = useMemo(() => {
    return [
      filters.bodyBrand,
      filters.engineBrand,
      filters.vin,
      filters.carReleasePeriod,
      filters.number,
      filters.oemCode,
      filters.defect,
      filters.color,
    ].filter(Boolean).length;
  }, [filters]);

  const transmissionFiltersCount = useMemo(() => {
    return [
      filters.transmission,
      filters.transmissionModel,
      filters.drive,
      filters.frontRear,
      filters.leftRight,
      filters.topBottom,
    ].filter(Boolean).length;
  }, [filters]);

  const warehouseFiltersCount = useMemo(() => {
    return [
      filters.location,
      filters.address,
      filters.salesman,
      filters.manufacturer,
      filters.manufacturerCode,
      filters.supplierCode,
      filters.wearPercentage,
    ].filter(Boolean).length;
  }, [filters]);

  const wheelsFiltersCount = useMemo(() => {
    return [
      filters.season,
      filters.diameter,
      filters.width,
      filters.profile,
      filters.tireQuantity,
      filters.drilling,
      filters.offset,
      filters.centerHoleDiameter,
      filters.tireModel,
    ].filter(Boolean).length;
  }, [filters]);

  // Формирование активных фильтров для бейджей
  const activeFilterItems = useMemo<ActiveFilterItem[]>(() => {
    const items: ActiveFilterItem[] = [];
    const query = filters.searchQuery.trim();

    if (query) {
      const multiTerms = query
        .split(/(?:\s+(?:или|or)\s+|[,\n\r;|]+)/i)
        .map((t) => t.trim())
        .filter(Boolean);

      if (multiTerms.length > 1) {
        multiTerms.forEach((term, index) => {
          items.push({
            id: `search-${index}`,
            label: `Поиск: ${term}`,
            onClear: () => {
              const remaining = multiTerms.filter((_, i) => i !== index);
              updateFilter('searchQuery', remaining.join(', '));
            },
          });
        });
      } else {
        items.push({ id: 'search', label: `Поиск: ${query}`, onClear: clearSearchQuery });
      }
    }

    for (const desc of FILTER_DESCRIPTORS) {
      const val = filters[desc.key];
      if (!val || (desc.key === 'hasPhoto' && val === 'all')) continue;
      items.push({
        id: desc.key,
        label: desc.getLabel(val),
        onClear: () => updateFilter(desc.key, desc.clearVal ?? ''),
      });
    }

    return items;
  }, [filters, updateFilter, clearSearchQuery]);

  const handleDisplayLimitChange = (value: string) => {
    const limit = parseInt(value) || 0;
    if (onDisplayLimitChange) {
      onDisplayLimitChange(limit === 0 ? undefined : limit);
    }
  };

  return (
    <motion.div layout>
      <Card className="mt-6 bg-white dark:bg-card border border-gray-300 shadow-none">
        <CardContent className="px-4 py-2">
          <div className="space-y-4">
            {/* Поисковая строка и кнопка фильтров */}
            <div className="flex flex-col sm:flex-row gap-2">
              <div className="relative flex-1">
                <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground pointer-events-none">
                  <AnimatePresence mode="wait">
                    {isSearching ? (
                      <motion.div
                        key="loader"
                        initial={{ opacity: 0, rotate: -180 }}
                        animate={{ opacity: 1, rotate: 0 }}
                        exit={{ opacity: 0, rotate: 180 }}
                        transition={{ duration: 0.3 }}
                      >
                        <Loader2 className="w-4 h-4 animate-spin" />
                      </motion.div>
                    ) : (
                      <motion.div
                        key="search"
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        exit={{ opacity: 0, scale: 0.8 }}
                        transition={{ duration: 0.3 }}
                      >
                        <Search className="w-4 h-4" />
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
                <Input
                  ref={searchInputRef}
                  id="search"
                  type="text"
                  value={filters.searchQuery}
                  onChange={handleInputChange}
                  onFocus={() => results.length > 0 && setIsResultsVisible(true)}
                  placeholder="Поиск по названию, номеру (можно через запятую: фара, крыло)..."
                  className="pl-10 pr-20 bg-white dark:bg-card border-input-border"
                />
                {/* Кнопка голосового поиска */}
                {isSpeechRecognitionSupported() && (
                  <AnimatePresence>
                    <motion.div
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.8 }}
                      transition={{ duration: 0.2 }}
                      className="absolute right-8 top-1/2 transform -translate-y-1/2"
                    >
                      <Button
                        onClick={isListening ? stopVoiceSearch : startVoiceSearch}
                        type="button"
                        variant="ghost"
                        size="sm"
                        className={`h-7 w-7 p-0 ${isListening ? 'text-red-500 bg-red-50' : ''}`}
                        aria-label={isListening ? 'Остановить голосовой поиск' : 'Голосовой поиск'}
                        title={isListening ? 'Остановить голосовой поиск' : 'Голосовой поиск'}
                      >
                        <AnimatePresence mode="wait">
                          {isListening ? (
                            <motion.div
                              key="micoff"
                              initial={{ opacity: 0, rotate: -90 }}
                              animate={{ opacity: 1, rotate: 0 }}
                              exit={{ opacity: 0, rotate: 90 }}
                              transition={{ duration: 0.2 }}
                            >
                              <MicOff className="h-4 w-4" />
                            </motion.div>
                          ) : (
                            <motion.div
                              key="mic"
                              initial={{ opacity: 0, scale: 0.8 }}
                              animate={{ opacity: 1, scale: 1 }}
                              exit={{ opacity: 0, scale: 0.8 }}
                              transition={{ duration: 0.2 }}
                            >
                              <Mic className="h-4 w-4" />
                            </motion.div>
                          )}
                        </AnimatePresence>
                      </Button>
                    </motion.div>
                  </AnimatePresence>
                )}
                <AnimatePresence>
                  {filters.searchQuery && (
                    <motion.div
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.8 }}
                      transition={{ duration: 0.2 }}
                      className="absolute right-1 top-1/2 transform -translate-y-1/2"
                    >
                      <Button
                        onClick={clearSearchQuery}
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-7 w-7 p-0"
                        aria-label="Очистить поиск"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </motion.div>
                  )}
                </AnimatePresence>

                {/* Сообщения под поисковой строкой */}
                <AnimatePresence mode="wait">
                  {filters.searchQuery.length > 0 && filters.searchQuery.length < MIN_SEARCH_LENGTH && (
                    <motion.p
                      key="min-length"
                      initial={{ opacity: 0, y: -5 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -5 }}
                      transition={{ duration: 0.2 }}
                      className="text-amber-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full"
                    >
                      <span>⚠️</span>
                      Минимум {MIN_SEARCH_LENGTH} символа для автодополнения
                    </motion.p>
                  )}

                  {isListening && (
                    <motion.p
                      key="listening"
                      initial={{ opacity: 0, y: -5 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -5 }}
                      transition={{ duration: 0.2 }}
                      className="text-red-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full"
                    >
                      <span>🎤</span>
                      Говорите... (слушаю)
                    </motion.p>
                  )}

                  {voiceSearchError && (
                    <motion.p
                      key="error"
                      initial={{ opacity: 0, y: -5 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -5 }}
                      transition={{ duration: 0.2 }}
                      className="text-red-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full"
                    >
                      <span>❌</span>
                      {voiceSearchError}
                    </motion.p>
                  )}
                </AnimatePresence>

                {/* Результаты автодополнения */}
                <AnimatePresence>
                  {isResultsVisible && results.length > 0 && !isSearching && (
                    <motion.div
                      ref={resultsRef}
                      initial={{ opacity: 0, scale: 0.95 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.95 }}
                      transition={{ duration: 0.3, ease: 'easeOut' }}
                      className="absolute left-0 right-0 mt-1 bg-popover border rounded-md shadow-lg z-50 max-h-60 overflow-y-auto top-full"
                    >
                      {results.map((result) => (
                        <div
                          key={result.id}
                          className="p-3 hover:bg-accent cursor-pointer border-b last:border-b-0 transition-all duration-200"
                          onClick={() => handleResultClick(result)}
                        >
                          <div className="font-medium text-sm">{result.name}</div>
                          {result.description && (
                            <div className="text-xs text-muted-foreground truncate mt-0.5">{result.description}</div>
                          )}
                        </div>
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              {/* Кнопка фильтров, сброса и селектор количества */}
              <div className="flex flex-col sm:flex-row gap-2 sm:gap-4">
                <div className="flex gap-2">
                  <motion.div layout>
                    <Button
                      variant={isFiltersOpen ? 'default' : 'outline'}
                      onClick={() => setIsFiltersOpen(!isFiltersOpen)}
                      className="gap-2 shrink-0 border-input"
                    >
                      <Filter className="h-4 w-4" />
                      Фильтры
                      <AnimatePresence>
                        {activeFiltersCount > 0 && (
                          <motion.div
                            initial={{ opacity: 0, scale: 0.8 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.8 }}
                            transition={{ duration: 0.2 }}
                          >
                            <Badge variant="secondary" className="ml-1 bg-transparent border border-gray-300">
                              {activeFiltersCount}
                            </Badge>
                          </motion.div>
                        )}
                      </AnimatePresence>
                      {isFiltersOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                    </Button>
                  </motion.div>

                  {/* Кнопка сброса всех фильтров */}
                  <AnimatePresence>
                    {activeFiltersCount > 0 && (
                      <motion.div
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        exit={{ opacity: 0, scale: 0.8 }}
                        transition={{ duration: 0.2 }}
                      >
                        <Button
                          onClick={clearFilters}
                          type="button"
                          variant="outline"
                          size="sm"
                          className="gap-2 shrink-0 border-input"
                        >
                          <X className="h-4 w-4" />
                          Сбросить фильтры
                        </Button>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>

                {/* Выбор количества отображаемых элементов */}
                <div className="flex items-center gap-2">
                  <Label htmlFor="display-limit" className="text-sm font-medium">
                    Показать:
                  </Label>
                  <Select
                    value={currentDisplayLimit ? currentDisplayLimit.toString() : '0'}
                    onValueChange={handleDisplayLimitChange}
                  >
                    <SelectTrigger id="display-limit" className="w-24 h-8">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="20">20</SelectItem>
                      <SelectItem value="40">40</SelectItem>
                      <SelectItem value="60">60</SelectItem>
                      <SelectItem value="80">80</SelectItem>
                      <SelectItem value="100">100</SelectItem>
                      <SelectItem value="0">все</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            {/* Раскрывающиеся фильтры */}
            <AnimatePresence initial={false}>
              {isFiltersOpen && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.3, ease: 'easeInOut' }}
                  style={{ overflow: 'hidden' }}
                >
                  <Card className="bg-transparent border border-gray-300">
                    <CardContent className="pt-4 pb-4">
                      <Tabs defaultValue="main" className="w-full">
                        <TabsList className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 w-full h-auto p-1 gap-1 bg-muted/60">
                          <TabsTrigger
                            value="main"
                            className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm"
                          >
                            <Car className="h-4 w-4 shrink-0" />
                            <span>Основные</span>
                            {mainFiltersCount > 0 && (
                              <Badge
                                variant="secondary"
                                className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold"
                              >
                                {mainFiltersCount}
                              </Badge>
                            )}
                          </TabsTrigger>
                          <TabsTrigger
                            value="bodyEngine"
                            className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm"
                          >
                            <Wrench className="h-4 w-4 shrink-0" />
                            <span>Кузов и ДВС</span>
                            {bodyEngineFiltersCount > 0 && (
                              <Badge
                                variant="secondary"
                                className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold"
                              >
                                {bodyEngineFiltersCount}
                              </Badge>
                            )}
                          </TabsTrigger>
                          <TabsTrigger
                            value="transmission"
                            className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm"
                          >
                            <GitFork className="h-4 w-4 shrink-0" />
                            <span>КПП и привод</span>
                            {transmissionFiltersCount > 0 && (
                              <Badge
                                variant="secondary"
                                className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold"
                              >
                                {transmissionFiltersCount}
                              </Badge>
                            )}
                          </TabsTrigger>
                          <TabsTrigger
                            value="warehouse"
                            className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm"
                          >
                            <Package className="h-4 w-4 shrink-0" />
                            <span>Склад</span>
                            {warehouseFiltersCount > 0 && (
                              <Badge
                                variant="secondary"
                                className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold"
                              >
                                {warehouseFiltersCount}
                              </Badge>
                            )}
                          </TabsTrigger>
                          <TabsTrigger
                            value="wheels"
                            className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm"
                          >
                            <Disc className="h-4 w-4 shrink-0" />
                            <span>Шины и диски</span>
                            {wheelsFiltersCount > 0 && (
                              <Badge
                                variant="secondary"
                                className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold"
                              >
                                {wheelsFiltersCount}
                              </Badge>
                            )}
                          </TabsTrigger>
                        </TabsList>

                        <TabsContent value="main" className="mt-4">
                          <GeneralFiltersTab
                            filters={filters}
                            onChange={updateFilter}
                            brandOptions={brandOptions}
                            modelOptions={modelOptions}
                            categoryOptions={categoryOptions}
                            statusOptions={statusOptions}
                            photoOptions={photoOptions}
                            conditionOptions={conditionOptions}
                          />
                        </TabsContent>

                        <TabsContent value="bodyEngine" className="mt-4">
                          <BodyEngineFiltersTab
                            filters={filters}
                            onChange={updateFilter}
                            bodyOptions={bodyOptions}
                            engineOptions={engineOptions}
                          />
                        </TabsContent>

                        <TabsContent value="transmission" className="mt-4">
                          <TransmissionFiltersTab
                            filters={filters}
                            onChange={updateFilter}
                            transmissionOptions={transmissionOptions}
                            driveOptions={driveOptions}
                            frontRearOptions={frontRearOptions}
                            leftRightOptions={leftRightOptions}
                            topBottomOptions={topBottomOptions}
                          />
                        </TabsContent>

                        <TabsContent value="warehouse" className="mt-4">
                          <WarehouseFiltersTab
                            filters={filters}
                            onChange={updateFilter}
                          />
                        </TabsContent>

                        <TabsContent value="wheels" className="mt-4">
                          <WheelsFiltersTab
                            filters={filters}
                            onChange={updateFilter}
                            seasonOptions={seasonOptions}
                          />
                        </TabsContent>
                      </Tabs>
                    </CardContent>
                  </Card>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Активные фильтры */}
            <ActiveFilterBadges filters={activeFilterItems} />
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
};

export default PartsSearch;
