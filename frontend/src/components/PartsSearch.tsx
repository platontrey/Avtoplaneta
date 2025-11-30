/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

// Типы для Web Speech API
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
  onstart: ((this: SpeechRecognition, ev: Event) => any) | null;
  onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => any) | null;
  onend: ((this: SpeechRecognition, ev: Event) => any) | null;
  onerror: ((this: SpeechRecognition, ev: SpeechRecognitionErrorEvent) => any) | null;
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

import { useState, useEffect, useCallback, useRef } from 'react';
import { Search, Loader2, X, Filter, ChevronDown, ChevronUp, Mic, MicOff } from 'lucide-react';
import type { Part } from '../lib/types';
import { Card, CardContent } from './ui/card';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { SearchableSelect } from './ui/searchable-select';
import type { SelectOption } from './ui/searchable-select';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { motion, AnimatePresence } from 'framer-motion';

interface PartsSearchProps {
    onSearchChange?: (searchQuery: string) => void;
    onFiltersChange?: (filters: {
        search: string;
        category: string;
        brand: string;
        model: string;
        location: string;
        status: string;
        hasPhoto?: string;
    }) => void;
    onDisplayLimitChange?: (limit: number | undefined) => void;
    currentDisplayLimit?: number | undefined;
}

function PartsSearch({ onFiltersChange, onDisplayLimitChange, currentDisplayLimit }: PartsSearchProps) {
    const [searchQuery, setSearchQuery] = useState('');
    const [category, setCategory] = useState('');
    const [brand, setBrand] = useState('');
    const [model, setModel] = useState('');
    const [location, setLocation] = useState('');
    const [status, setStatus] = useState('');
    const [hasPhoto, setHasPhoto] = useState('with');
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

    const MIN_SEARCH_LENGTH = 2;
    const SEARCH_DEBOUNCE_DELAY = 500;
    const FILTER_DEBOUNCE_DELAY = 300;
    
    // Опции для выпадающих списков
    const categoryOptions: SelectOption[] = [
      { value: "Тормоза", label: "Тормоза" },
      { value: "Двигатель", label: "Двигатель" },
      { value: "Подвеска", label: "Подвеска" },
      { value: "Электрика", label: "Электрика" },
      { value: "Кузов", label: "Кузов" },
      { value: "Интерьер", label: "Интерьер" },
      { value: "Трансмиссия", label: "Трансмиссия" },
      { value: "Система охлаждения и отопления", label: "Система охлаждения и отопления" },
      { value: "Система выхлопа (Глушитель)", label: "Система выхлопа (Глушитель)" },
      { value: "Система рулевого управления", label: "Система рулевого управления" },
      { value: "Система фильтрации (Фильтры)", label: "Система фильтрации (Фильтры)" },
      { value: "Шины и диски", label: "Шины и диски" },
      { value: "Автохимия и масла", label: "Автохимия и масла" },
      { value: "Аксессуары и тюннинг", label: "Аксессуары и тюннинг" },
      { value: "Другое", label: "Другое" },
    ];
    
    const brandOptions: SelectOption[] = [
        { value: "Acura", label: "Acura" },
        { value: "Aston Martin", label: "Aston Martin" },
        { value: "Audi", label: "Audi" },
        { value: "Bentley", label: "Bentley" },
        { value: "BMW", label: "BMW" },
        { value: "Buick", label: "Buick" },
        { value: "Cadillac", label: "Cadillac" },
        { value: "Chevrolet", label: "Chevrolet" },
        { value: "Chrysler", label: "Chrysler" },
        { value: "Citroen", label: "Citroën" },
        { value: "DAF", label: "DAF" },
        { value: "Daihatsu", label: "Daihatsu" },
        { value: "Dodge", label: "Dodge" },
        { value: "Ferrari", label: "Ferrari" },
        { value: "Fiat", label: "Fiat" },
        { value: "Ford", label: "Ford" },
        { value: "GMC", label: "GMC" },
        { value: "Hino", label: "Hino" },
        { value: "Honda", label: "Honda" },
        { value: "Hyundai", label: "Hyundai" },
        { value: "Infiniti", label: "Infiniti" },
        { value: "Isuzu", label: "Isuzu" },
        { value: "Iveco", label: "Iveco" },
        { value: "Jaguar", label: "Jaguar" },
        { value: "Jeep", label: "Jeep" },
        { value: "Kenworth", label: "Kenworth" },
        { value: "Kia", label: "Kia" },
        { value: "Lamborghini", label: "Lamborghini" },
        { value: "Land Rover", label: "Land Rover" },
        { value: "Lexus", label: "Lexus" },
        { value: "Lincoln", label: "Lincoln" },
        { value: "Mack", label: "Mack" },
        { value: "MAN", label: "MAN" },
        { value: "Mazda", label: "Mazda" },
        { value: "Mercedes", label: "Mercedes-Benz" },
        { value: "Mercedes-Benz Trucks", label: "Mercedes-Benz Trucks" },
        { value: "Mitsubishi", label: "Mitsubishi" },
        { value: "Nissan", label: "Nissan" },
        { value: "Opel", label: "Opel" },
        { value: "Peterbilt", label: "Peterbilt" },
        { value: "Peugeot", label: "Peugeot" },
        { value: "Porsche", label: "Porsche" },
        { value: "Ram", label: "Ram" },
        { value: "Renault", label: "Renault" },
        { value: "Rolls-Royce", label: "Rolls-Royce" },
        { value: "Scania", label: "Scania" },
        { value: "Subaru", label: "Subaru" },
        { value: "Suzuki", label: "Suzuki" },
        { value: "Tesla", label: "Tesla" },
        { value: "Toyota", label: "Toyota" },
        { value: "UD Trucks", label: "UD Trucks" },
        { value: "Volkswagen", label: "Volkswagen" },
        { value: "Volvo", label: "Volvo" },
        { value: "Volvo Trucks", label: "Volvo Trucks" },
        { value: "Western Star", label: "Western Star" }
    ];
    
    const statusOptions: SelectOption[] = [
      { value: "true", label: "Доступно" },
      { value: "false", label: "Недоступно" },
    ];
    
    const photoOptions: SelectOption[] = [
      { value: "with", label: "С фото" },
      { value: "without", label: "Без фото" },
      { value: "all", label: "Все" },
    ];

    // Функция применения фильтров
    const applyFilters = useCallback(() => {
        if (onFiltersChange) {
            onFiltersChange({
                search: searchQuery,
                category,
                brand,
                model,
                location,
                status,
                hasPhoto
            });
        }
    }, [searchQuery, category, brand, model, location, status, hasPhoto, onFiltersChange]);

    // Поиск с автодополнением (только для dropdown)
    const performSearch = useCallback(async (query: string) => {
        if (query.length < MIN_SEARCH_LENGTH) {
            clearResults();
            return;
        }

        // Очищаем предыдущие результаты перед новым поиском
        clearResults();

        // Отменяем предыдущий запрос
        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
        }

        abortControllerRef.current = new AbortController();
        setIsSearching(true);

        try {
            const url = `http://localhost:8080/api/inventory?search=${encodeURIComponent(query)}&limit=10`;
            const response = await fetch(url, {
                credentials: 'include',
                signal: abortControllerRef.current.signal
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            const parts = data.parts || [];

            // Проверяем, изменились ли результаты
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
                clearResults();
            }
        } finally {
            setIsSearching(false);
        }
    }, []);

    // Debounce для поиска автодополнения
    useEffect(() => {
        const timeoutId = setTimeout(() => {
            performSearch(searchQuery);
        }, SEARCH_DEBOUNCE_DELAY);

        return () => clearTimeout(timeoutId);
    }, [searchQuery, performSearch]);

    // Debounce для применения фильтров (включая поисковый запрос)
    useEffect(() => {
        const timeoutId = setTimeout(() => {
            applyFilters();
        }, FILTER_DEBOUNCE_DELAY);

        return () => {
            clearTimeout(timeoutId);
        };
    }, [searchQuery, category, brand, model, location, status, hasPhoto, applyFilters]);

    // Очистка при размонтировании
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

    // Проверка поддержки Web Speech API
    const isSpeechRecognitionSupported = () => {
        return 'SpeechRecognition' in window || 'webkitSpeechRecognition' in window;
    };

    // Функция начала голосового поиска
    const startVoiceSearch = useCallback(() => {
        if (!isSpeechRecognitionSupported()) {
            setVoiceSearchError('Ваш браузер не поддерживает голосовой поиск');
            return;
        }

        setVoiceSearchError(null);

        const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
        recognitionRef.current = new SpeechRecognition();

        recognitionRef.current.lang = 'ru-RU';
        recognitionRef.current.continuous = false;
        recognitionRef.current.interimResults = false;

        recognitionRef.current.onstart = () => {
            setIsListening(true);
        };

        recognitionRef.current.onresult = (event: SpeechRecognitionEvent) => {
            const transcript = event.results[0][0].transcript;
            setSearchQuery(transcript);
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
        } catch (error) {
            setVoiceSearchError('Не удалось начать распознавание речи');
            setIsListening(false);
        }
    }, []);

    // Функция остановки голосового поиска
    const stopVoiceSearch = useCallback(() => {
        if (recognitionRef.current) {
            recognitionRef.current.stop();
        }
        setIsListening(false);
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setSearchQuery(value);
    };

    const handleResultClick = (result: Part) => {
        setSearchQuery(result.name);
        setIsResultsVisible(false);
        setResults([]);
    };

    const clearResults = () => {
        setResults([]);
        prevResultsRef.current = [];
        setIsResultsVisible(false);
    };

    const clearSearchQuery = () => {
        setSearchQuery('');
        clearResults();
        setVoiceSearchError(null); // Очистить ошибки голосового поиска
    };

    const clearFilters = () => {
        setSearchQuery('');
        setCategory('');
        setBrand('');
        setModel('');
        setLocation('');
        setStatus('');
        setHasPhoto('with');
        clearResults();
    };

    // Закрытие результатов при клике вне компонента
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

    // Подсчет активных фильтров (не считая поиск и фото по умолчанию)
    const activeFiltersCount = [
        category,
        brand,
        model,
        location,
        status,
        hasPhoto && hasPhoto !== 'with' ? hasPhoto : ''
    ].filter(Boolean).length;

    const handleDisplayLimitChange = (value: string) => {
        const limit = parseInt(value) || 0;
        if (onDisplayLimitChange) {
            onDisplayLimitChange(limit === 0 ? undefined : limit);
        }
    };

    return (
        <Card className="mt-6 bg-white dark:bg-card border border-gray-300 shadow-none">
            <CardContent className="px-4 py-2">
                <div className="space-y-4">
                    {/* Поисковая строка и кнопка фильтров */}
                    <div className="flex flex-col sm:flex-row gap-2">
                        <div className="relative flex-1">
                            <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground pointer-events-none">
                                {isSearching ? (
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                    <Search className="w-4 h-4" />
                                )}
                            </div>
                            <Input
                                ref={searchInputRef}
                                id="search"
                                type="text"
                                value={searchQuery}
                                onChange={handleInputChange}
                                onFocus={() => results.length > 0 && setIsResultsVisible(true)}
                                placeholder="Поиск по названию или описанию..."
                                className="pl-10 pr-20 bg-white border-input-border"
                            />
                            {/* Кнопка голосового поиска */}
                            {isSpeechRecognitionSupported() && (
                                <Button
                                    onClick={isListening ? stopVoiceSearch : startVoiceSearch}
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    className={`absolute right-8 top-1/2 transform -translate-y-1/2 h-7 w-7 p-0 ${
                                        isListening ? 'text-red-500 bg-red-50' : ''
                                    }`}
                                    aria-label={isListening ? 'Остановить голосовой поиск' : 'Голосовой поиск'}
                                    title={isListening ? 'Остановить голосовой поиск' : 'Голосовой поиск'}
                                >
                                    {isListening ? (
                                        <MicOff className="h-4 w-4" />
                                    ) : (
                                        <Mic className="h-4 w-4" />
                                    )}
                                </Button>
                            )}
                            {searchQuery && (
                                <Button
                                    onClick={clearSearchQuery}
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    className="absolute right-1 top-1/2 transform -translate-y-1/2 h-7 w-7 p-0"
                                    aria-label="Очистить поиск"
                                >
                                    <X className="h-4 w-4" />
                                </Button>
                            )}

                            {/* Сообщения под поисковой строкой */}
                            {(searchQuery.length > 0 && searchQuery.length < MIN_SEARCH_LENGTH) && (
                                <p className="text-amber-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full">
                                    <span>⚠️</span>
                                    Минимум {MIN_SEARCH_LENGTH} символа для автодополнения
                                </p>
                            )}

                            {isListening && (
                                <p className="text-red-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full">
                                    <span>🎤</span>
                                    Говорите... (слушаю)
                                </p>
                            )}

                            {voiceSearchError && (
                                <p className="text-red-600 text-xs mt-1.5 flex items-center gap-1 absolute left-0 top-full">
                                    <span>❌</span>
                                    {voiceSearchError}
                                </p>
                            )}

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

                        {/* Кнопка фильтров и селектор количества */}
                        <div className="flex flex-col sm:flex-row gap-2 sm:gap-4">
                            <Button
                                variant={isFiltersOpen ? "default" : "outline"}
                                onClick={() => setIsFiltersOpen(!isFiltersOpen)}
                                className="gap-2 shrink-0 border-input"
                            >
                                <Filter className="h-4 w-4" />
                                Фильтры
                                {activeFiltersCount > 0 && (
                                    <Badge variant="secondary" className="ml-1 bg-transparent border border-gray-300">
                                        {activeFiltersCount}
                                    </Badge>
                                )}
                                {isFiltersOpen ? (
                                    <ChevronUp className="h-4 w-4" />
                                ) : (
                                    <ChevronDown className="h-4 w-4" />
                                )}
                            </Button>

                            {/* Выбор количества отображаемых элементов */}
                            <div className="flex items-center gap-2">
                                <Label htmlFor="display-limit" className="text-sm font-medium">Показать:</Label>
                                <Select
                                    value={currentDisplayLimit ? currentDisplayLimit.toString() : "0"}
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
                                    <CardContent className="pt-6">
                                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                            {/* Категория */}
                                            <div>
                                                <Label htmlFor="category-filter">Категория</Label>
                                                <div className="mt-1.5">
                                                    <SearchableSelect
                                                        value={category}
                                                        onValueChange={(value) => setCategory(value)}
                                                        options={categoryOptions}
                                                        placeholder="Все категории"
                                                        searchPlaceholder="Поиск категории..."
                                                        emptyMessage="Категория не найдена"
                                                        className="bg-transparent border border-gray-300"
                                                    />
                                                </div>
                                            </div>

                                            {/* Бренд */}
                                            <div>
                                                <Label htmlFor="brand-filter">Бренд</Label>
                                                <div className="mt-1.5">
                                                    <SearchableSelect
                                                        value={brand}
                                                        onValueChange={(value) => setBrand(value)}
                                                        options={brandOptions}
                                                        placeholder="Все бренды"
                                                        searchPlaceholder="Поиск бренда..."
                                                        emptyMessage="Бренд не найден"
                                                        className="bg-transparent border border-gray-300"
                                                    />
                                                </div>
                                            </div>

                                            {/* Модель */}
                                            <div>
                                                <Label htmlFor="model-filter">Модель</Label>
                                                <Input
                                                    id="model-filter"
                                                    type="text"
                                                    value={model}
                                                    onChange={(e) => setModel(e.target.value)}
                                                    placeholder="E90, A4..."
                                                    className="mt-1.5 bg-transparent border border-gray-300"
                                                />
                                            </div>

                                            {/* Местоположение */}
                                            <div>
                                                <Label htmlFor="location-filter">Местоположение</Label>
                                                <Input
                                                    id="location-filter"
                                                    type="text"
                                                    value={location}
                                                    onChange={(e) => setLocation(e.target.value)}
                                                    placeholder="Shelf A-12..."
                                                    className="mt-1.5 bg-transparent border border-gray-300"
                                                />
                                            </div>


                                            {/* Статус */}
                                            <div>
                                                <Label htmlFor="status-filter">Статус</Label>
                                                <div className="mt-1.5">
                                                    <SearchableSelect
                                                        value={status}
                                                        onValueChange={(value) => setStatus(value)}
                                                        options={statusOptions}
                                                        placeholder="Все статусы"
                                                        searchPlaceholder="Поиск статуса..."
                                                        emptyMessage="Статус не найден"
                                                        className="bg-transparent border border-gray-300"
                                                    />
                                                </div>
                                            </div>

                                            {/* Фото */}
                                            <div>
                                                <Label htmlFor="photo-filter">Фото</Label>
                                                <div className="mt-1.5">
                                                    <SearchableSelect
                                                        value={hasPhoto}
                                                        onValueChange={(value) => setHasPhoto(value)}
                                                        options={photoOptions}
                                                        placeholder="Все"
                                                        searchPlaceholder="Поиск по фото..."
                                                        emptyMessage="Опция не найдена"
                                                        className="bg-transparent border border-gray-300"
                                                    />
                                                </div>
                                            </div>
                                        </div>

                                        {/* Кнопка очистки фильтров */}
                                        <div className="mt-4 flex justify-end">
                                            <Button
                                                onClick={clearFilters}
                                                type="button"
                                                variant="outline"
                                                disabled={isSearching || (activeFiltersCount === 0 && !searchQuery)}
                                                className="gap-2"
                                            >
                                                <X className="h-4 w-4" />
                                                Очистить все фильтры
                                            </Button>
                                        </div>
                                    </CardContent>
                                </Card>
                            </motion.div>
                        )}
                    </AnimatePresence>

                    {/* Активные фильтры */}
                    {(activeFiltersCount > 0 || searchQuery) && (
                        <div className="flex flex-wrap gap-2">
                            {searchQuery && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Поиск: {searchQuery}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            clearSearchQuery();
                                        }}
                                    />
                                </Badge>
                            )}
                            {category && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Категория: {category}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setCategory('');
                                        }}
                                    />
                                </Badge>
                            )}
                            {brand && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Бренд: {brand}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setBrand('');
                                        }}
                                    />
                                </Badge>
                            )}
                            {model && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Модель: {model}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setModel('');
                                        }}
                                    />
                                </Badge>
                            )}
                            {location && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Местоположение: {location}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setLocation('');
                                        }}
                                    />
                                </Badge>
                            )}
                            {status && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Статус: {status === 'true' ? 'Доступно' : 'Недоступно'}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setStatus('');
                                        }}
                                    />
                                </Badge>
                            )}
                            {hasPhoto && hasPhoto !== 'with' && (
                                <Badge variant="secondary" className="gap-1 bg-transparent border border-gray-300">
                                    <span className="pointer-events-none">Фото: {hasPhoto === 'without' ? 'Без фото' : 'Все'}</span>
                                    <X
                                        className="h-3 w-3 cursor-pointer hover:text-destructive"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            setHasPhoto('with');
                                        }}
                                    />
                                </Badge>
                            )}
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}

export default PartsSearch;
