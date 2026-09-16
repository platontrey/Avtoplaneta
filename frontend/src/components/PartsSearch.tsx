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

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { Search, Loader2, X, Filter, ChevronDown, ChevronUp, Mic, MicOff, Car, Wrench, GitFork, Package, Disc } from 'lucide-react';
import type { Part } from '../lib/types';
import { Card, CardContent } from './ui/card';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Checkbox } from './ui/checkbox';
import { SearchableSelect } from './ui/searchable-select';
import type { SelectOption } from './ui/searchable-select';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs';
import { motion, AnimatePresence } from 'framer-motion';
import { API_BASE_URL } from '@/lib/api';
import { brandOptions } from '@/lib/constants';
import { usePartCatalog } from '@/features/catalog/usePartCatalog';
import type { PartFilters } from '@/hooks/useParts';
import { formatCarReleasePeriod } from '@/lib/utils';

interface PartsSearchProps {
    onSearchChange?: (searchQuery: string) => void;
    onFiltersChange?: (filters: PartFilters) => void;
    onDisplayLimitChange?: (limit: number | undefined) => void;
    currentDisplayLimit?: number | undefined;
}

function PartsSearch({ onFiltersChange, onDisplayLimitChange, currentDisplayLimit }: PartsSearchProps) {
    const [searchQuery, setSearchQuery] = useState('');
    const [category, setCategory] = useState('');
    const [brand, setBrand] = useState('');
    const [model, setModel] = useState('');
    const [location, setLocation] = useState('');
    const [address, setAddress] = useState('');
    const [status, setStatus] = useState('');
    const [hasPhoto, setHasPhoto] = useState('with');
    const [number, setNumber] = useState('');
    const [oemCode, setOemCode] = useState('');
    const [vin, setVin] = useState('');
    const [carReleasePeriod, setCarReleasePeriod] = useState('');
    const [bodyBrand, setBodyBrand] = useState('');
    const [engineBrand, setEngineBrand] = useState('');
    const [carReleaseDate, setCarReleaseDate] = useState('');
    const [transmission, setTransmission] = useState('');
    const [drive, setDrive] = useState('');
    const [condition, setCondition] = useState('');
    const [manufacturer, setManufacturer] = useState('');
    const [defect, setDefect] = useState('');
    const [color, setColor] = useState('');
    const [salesman, setSalesman] = useState('');
    const [minPrice, setMinPrice] = useState('');
    const [maxPrice, setMaxPrice] = useState('');
    const [minQuantity, setMinQuantity] = useState('');
    const [maxQuantity, setMaxQuantity] = useState('');
    const [frontRear, setFrontRear] = useState('');
    const [leftRight, setLeftRight] = useState('');
    const [topBottom, setTopBottom] = useState('');
    const [manufacturerCode, setManufacturerCode] = useState('');
    const [supplierCode, setSupplierCode] = useState('');
    const [transmissionModel, setTransmissionModel] = useState('');
    const [wearPercentage, setWearPercentage] = useState('');
    const [season, setSeason] = useState('');
    const [diameter, setDiameter] = useState('');
    const [width, setWidth] = useState('');
    const [profile, setProfile] = useState('');
    const [tireQuantity, setTireQuantity] = useState('');
    const [drilling, setDrilling] = useState('');
    const [offset, setOffset] = useState('');
    const [centerHoleDiameter, setCenterHoleDiameter] = useState('');
    const [tireModel, setTireModel] = useState('');
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

    const { data: partCatalog } = usePartCatalog();

    // Опции для выпадающих списков категорий (все категории каталога + исторические)
    const categoryOptions: SelectOption[] = useMemo(() => {
        const catalogCategories = partCatalog?.part_form_categories?.map(c => c.name) ?? [];
        const allCategories = Array.from(new Set([
            ...catalogCategories,
            "Выхлопная система",
            "Двигатель",
            "Диски и шины",
            "Кузов",
            "Кузов внутри",
            "Кузов снаружи",
            "Оптика",
            "Пневмосистема",
            "Подвеска",
            "Подвеска ДВС/КПП",
            "Подвеска задних колес",
            "Подвеска передних колес",
            "Рулевое управление",
            "Система кондиционирования",
            "Система охлаждения и отопления",
            "Система выхлопа (Глушитель)",
            "Система рулевого управления",
            "Система фильтрации (Фильтры)",
            "Сопутствующие товары",
            "Стекла",
            "Тормоза",
            "Тормозная система",
            "Трансмиссия",
            "Шины и диски",
            "Электрика",
            "Электрооснащение",
            "Автохимия и масла",
            "Аксессуары и тюннинг",
            "Интерьер",
            "Другое",
        ])).filter(Boolean).sort((a, b) => a.localeCompare(b, 'ru'));

        return allCategories.map(cat => ({ value: cat, label: cat }));
    }, [partCatalog]);

    const statusOptions: SelectOption[] = [
      { value: "true", label: "Доступно" },
      { value: "false", label: "Недоступно" },
    ];

    const photoOptions: SelectOption[] = [
      { value: "with", label: "С фото" },
      { value: "without", label: "Без фото" },
      { value: "all", label: "Все" },
    ];

    const transmissionOptions: SelectOption[] = [
      { value: "АКПП", label: "АКПП" },
      { value: "МКПП", label: "МКПП" },
      { value: "Вариатор", label: "Вариатор" },
      { value: "Роботизированная", label: "Роботизированная" },
    ];

    const driveOptions: SelectOption[] = [
      { value: "Передний", label: "Передний" },
      { value: "Задний", label: "Задний" },
      { value: "Полный", label: "Полный" },
    ];

    const frontRearOptions: SelectOption[] = [
      { value: "Передний", label: "Передний" },
      { value: "Задний", label: "Задний" },
    ];

    const leftRightOptions: SelectOption[] = [
      { value: "Левый", label: "Левый" },
      { value: "Правый", label: "Правый" },
    ];

    const topBottomOptions: SelectOption[] = [
      { value: "Верхний", label: "Верхний" },
      { value: "Нижний", label: "Нижний" },
    ];

    const seasonOptions: SelectOption[] = [
      { value: "Лето", label: "Лето" },
      { value: "Зима", label: "Зима" },
      { value: "Всесезонная", label: "Всесезонная" },
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
                address,
                salesman,
                status,
                hasPhoto,
                number,
                oem_code: oemCode,
                vin,
                car_release_period: carReleasePeriod,
                body_brand: bodyBrand,
                engine_brand: engineBrand,
                car_release_date: carReleaseDate,
                transmission,
                drive,
                condition,
                manufacturer,
                defect,
                color,
                min_price: minPrice,
                max_price: maxPrice,
                min_quantity: minQuantity,
                max_quantity: maxQuantity,
                front_rear: frontRear,
                left_right: leftRight,
                top_bottom: topBottom,
                manufacturer_code: manufacturerCode,
                supplier_code: supplierCode,
                transmission_model: transmissionModel,
                wear_percentage: wearPercentage,
                season,
                diameter,
                width,
                profile,
                tire_quantity: tireQuantity,
                drilling,
                offset,
                center_hole_diameter: centerHoleDiameter,
                tire_model: tireModel,
            });
        }
    }, [
        searchQuery, category, brand, model, location, address, salesman, status, hasPhoto,
        number, oemCode, vin, carReleasePeriod, bodyBrand, engineBrand, carReleaseDate,
        transmission, drive, condition, manufacturer, defect, color,
        minPrice, maxPrice, minQuantity, maxQuantity,
        frontRear, leftRight, topBottom, manufacturerCode, supplierCode, transmissionModel, wearPercentage,
        season, diameter, width, profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel,
        onFiltersChange
    ]);

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
            const url = `${API_BASE_URL}/api/v1/inventory?search=${encodeURIComponent(query)}&limit=10`;
            const response = await fetch(url, {
                credentials: 'include',
                signal: abortControllerRef.current.signal
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            const parts = data.parts || [];

            // Проверяем, изменились ли результаты (оптимизированная проверка)
            const resultsChanged = parts.length !== prevResultsRef.current.length ||
                parts.some((part: Part, index: number) => part.id !== prevResultsRef.current[index]?.id);

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
    }, [searchQuery, category, brand, model, location, address, status, hasPhoto, applyFilters]);

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
        } catch {
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

    const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setSearchQuery(value);
    }, []);

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
        setAddress('');
        setStatus('');
        setHasPhoto('with');
        setNumber('');
        setOemCode('');
        setVin('');
        setCarReleasePeriod('');
        setBodyBrand('');
        setEngineBrand('');
        setCarReleaseDate('');
        setTransmission('');
        setDrive('');
        setCondition('');
        setManufacturer('');
        setDefect('');
        setColor('');
        setSalesman('');
        setMinPrice('');
        setMaxPrice('');
        setMinQuantity('');
        setMaxQuantity('');
        setFrontRear('');
        setLeftRight('');
        setTopBottom('');
        setManufacturerCode('');
        setSupplierCode('');
        setTransmissionModel('');
        setWearPercentage('');
        setSeason('');
        setDiameter('');
        setWidth('');
        setProfile('');
        setTireQuantity('');
        setDrilling('');
        setOffset('');
        setCenterHoleDiameter('');
        setTireModel('');
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

    // Подсчет активных фильтров (не считая поиск и фото по умолчанию) - оптимизировано с useMemo
    const activeFiltersCount = useMemo(() => [
        category,
        brand,
        model,
        location,
        address,
        status,
        hasPhoto && hasPhoto !== 'with' ? hasPhoto : '',
        number,
        oemCode,
        vin,
        carReleasePeriod,
        bodyBrand,
        engineBrand,
        carReleaseDate,
        transmission,
        drive,
        condition,
        manufacturer,
        defect,
        color,
        salesman,
        minPrice,
        maxPrice,
        minQuantity,
        maxQuantity,
        frontRear,
        leftRight,
        topBottom,
        manufacturerCode,
        supplierCode,
        transmissionModel,
        wearPercentage,
        season,
        diameter,
        width,
        profile,
        tireQuantity,
        drilling,
        offset,
        centerHoleDiameter,
        tireModel,
    ].filter(Boolean).length, [
        category, brand, model, location, address, status, hasPhoto,
        number, oemCode, vin, carReleasePeriod, bodyBrand, engineBrand, carReleaseDate,
        transmission, drive, condition, manufacturer, defect, color,
        salesman, minPrice, maxPrice, minQuantity, maxQuantity,
        frontRear, leftRight, topBottom, manufacturerCode, supplierCode,
        transmissionModel, wearPercentage, season, diameter, width,
        profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel,
    ]);

    const mainFiltersCount = useMemo(() => [
        category,
        brand,
        model,
        carReleaseDate,
        condition,
        minPrice,
        maxPrice,
        minQuantity,
        maxQuantity,
        status,
        hasPhoto && hasPhoto !== 'with' ? hasPhoto : '',
    ].filter(Boolean).length, [category, brand, model, carReleaseDate, condition, minPrice, maxPrice, minQuantity, maxQuantity, status, hasPhoto]);

    const bodyEngineFiltersCount = useMemo(() => [
        bodyBrand,
        engineBrand,
        vin,
        carReleasePeriod,
        number,
        oemCode,
        defect,
        color,
    ].filter(Boolean).length, [bodyBrand, engineBrand, vin, carReleasePeriod, number, oemCode, defect, color]);

    const transmissionFiltersCount = useMemo(() => [
        transmission,
        transmissionModel,
        drive,
        frontRear,
        leftRight,
        topBottom,
    ].filter(Boolean).length, [transmission, transmissionModel, drive, frontRear, leftRight, topBottom]);

    const warehouseFiltersCount = useMemo(() => [
        location,
        address,
        salesman,
        manufacturer,
        manufacturerCode,
        supplierCode,
        wearPercentage,
    ].filter(Boolean).length, [location, address, salesman, manufacturer, manufacturerCode, supplierCode, wearPercentage]);

    const wheelsFiltersCount = useMemo(() => [
        season,
        diameter,
        width,
        profile,
        tireQuantity,
        drilling,
        offset,
        centerHoleDiameter,
        tireModel,
    ].filter(Boolean).length, [season, diameter, width, profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel]);

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
                                value={searchQuery}
                                onChange={handleInputChange}
                                onFocus={() => results.length > 0 && setIsResultsVisible(true)}
                                placeholder="Поиск по названию или описанию..."
                                className="pl-10 pr-20 bg-white border-input-border"
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
                                            className={`h-7 w-7 p-0 ${
                                                isListening ? 'text-red-500 bg-red-50' : ''
                                            }`}
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
                                {searchQuery && (
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
                                {(searchQuery.length > 0 && searchQuery.length < MIN_SEARCH_LENGTH) && (
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
                                        variant={isFiltersOpen ? "default" : "outline"}
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
                                    {isFiltersOpen ? (
                                        <ChevronUp className="h-4 w-4" />
                                    ) : (
                                        <ChevronDown className="h-4 w-4" />
                                    )}
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
                                    <CardContent className="pt-4 pb-4">
                                        <Tabs defaultValue="main" className="w-full">
                                            <TabsList className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 w-full h-auto p-1 gap-1 bg-muted/60">
                                                <TabsTrigger value="main" className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm">
                                                    <Car className="h-4 w-4 shrink-0" />
                                                    <span>Основные</span>
                                                    {mainFiltersCount > 0 && (
                                                        <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold">
                                                            {mainFiltersCount}
                                                        </Badge>
                                                    )}
                                                </TabsTrigger>
                                                <TabsTrigger value="bodyEngine" className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm">
                                                    <Wrench className="h-4 w-4 shrink-0" />
                                                    <span>Кузов и ДВС</span>
                                                    {bodyEngineFiltersCount > 0 && (
                                                        <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold">
                                                            {bodyEngineFiltersCount}
                                                        </Badge>
                                                    )}
                                                </TabsTrigger>
                                                <TabsTrigger value="transmission" className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm">
                                                    <GitFork className="h-4 w-4 shrink-0" />
                                                    <span>КПП и привод</span>
                                                    {transmissionFiltersCount > 0 && (
                                                        <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold">
                                                            {transmissionFiltersCount}
                                                        </Badge>
                                                    )}
                                                </TabsTrigger>
                                                <TabsTrigger value="warehouse" className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm">
                                                    <Package className="h-4 w-4 shrink-0" />
                                                    <span>Склад</span>
                                                    {warehouseFiltersCount > 0 && (
                                                        <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold">
                                                            {warehouseFiltersCount}
                                                        </Badge>
                                                    )}
                                                </TabsTrigger>
                                                <TabsTrigger value="wheels" className="flex items-center justify-center gap-1.5 py-2 px-2 text-xs sm:text-sm">
                                                    <Disc className="h-4 w-4 shrink-0" />
                                                    <span>Шины и диски</span>
                                                    {wheelsFiltersCount > 0 && (
                                                        <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs bg-primary/15 text-primary border-none font-semibold">
                                                            {wheelsFiltersCount}
                                                        </Badge>
                                                    )}
                                                </TabsTrigger>
                                            </TabsList>

                                            {/* Вкладка 1: Основные */}
                                            <TabsContent value="main" className="mt-4">
                                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
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
                                                                allowCustom={true}
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

                                                    {/* Год выпуска */}
                                                    <div>
                                                        <Label htmlFor="release-date-filter">Год выпуска</Label>
                                                        <Input
                                                            id="release-date-filter"
                                                            type="text"
                                                            value={carReleaseDate}
                                                            onChange={(e) => setCarReleaseDate(e.target.value)}
                                                            placeholder="2010..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Состояние */}
                                                    <div>
                                                        <Label htmlFor="condition-filter">Состояние</Label>
                                                        <Input
                                                            id="condition-filter"
                                                            type="text"
                                                            value={condition}
                                                            onChange={(e) => setCondition(e.target.value)}
                                                            placeholder="Контрактная, б/у, новая..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Сдвоенный диапазон: Цена */}
                                                    <div>
                                                        <Label>Цена (₽)</Label>
                                                        <div className="grid grid-cols-2 gap-2 mt-1.5">
                                                            <Input
                                                                id="min-price-filter"
                                                                type="number"
                                                                value={minPrice}
                                                                onChange={(e) => setMinPrice(e.target.value)}
                                                                placeholder="От 0"
                                                                min="0"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                            <Input
                                                                id="max-price-filter"
                                                                type="number"
                                                                value={maxPrice}
                                                                onChange={(e) => setMaxPrice(e.target.value)}
                                                                placeholder="До"
                                                                min="0"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Сдвоенный диапазон: Количество */}
                                                    <div>
                                                        <Label>Количество</Label>
                                                        <div className="grid grid-cols-2 gap-2 mt-1.5">
                                                            <Input
                                                                id="min-quantity-filter"
                                                                type="number"
                                                                value={minQuantity}
                                                                onChange={(e) => setMinQuantity(e.target.value)}
                                                                placeholder="От"
                                                                min="0"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                            <Input
                                                                id="max-quantity-filter"
                                                                type="number"
                                                                value={maxQuantity}
                                                                onChange={(e) => setMaxQuantity(e.target.value)}
                                                                placeholder="До"
                                                                min="0"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
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
                                            </TabsContent>

                                            {/* Вкладка 2: Кузов и ДВС */}
                                            <TabsContent value="bodyEngine" className="mt-4">
                                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                                    {/* Марка кузова */}
                                                    <div>
                                                        <Label htmlFor="body-brand-filter">Марка кузова</Label>
                                                        <Input
                                                            id="body-brand-filter"
                                                            type="text"
                                                            value={bodyBrand}
                                                            onChange={(e) => setBodyBrand(e.target.value)}
                                                            placeholder="ACV40, E90, W212..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Марка двигателя */}
                                                    <div>
                                                        <Label htmlFor="engine-brand-filter">Марка двигателя</Label>
                                                        <Input
                                                            id="engine-brand-filter"
                                                            type="text"
                                                            value={engineBrand}
                                                            onChange={(e) => setEngineBrand(e.target.value)}
                                                            placeholder="2AZ-FE, N46, 1JZ..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* VIN / Номер кузова */}
                                                    <div>
                                                        <Label htmlFor="vin-filter">VIN / Номер кузова</Label>
                                                        <Input
                                                            id="vin-filter"
                                                            type="text"
                                                            value={vin}
                                                            onChange={(e) => setVin(e.target.value)}
                                                            placeholder="WVWZZZ1JZ3W386549..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Период выпуска автомобиля */}
                                                    <div>
                                                        <Label htmlFor="car-release-period-filter">Период выпуска автомобиля</Label>
                                                        <Input
                                                            id="car-release-period-filter"
                                                            type="text"
                                                            value={carReleasePeriod}
                                                            onChange={(e) => setCarReleasePeriod(formatCarReleasePeriod(e.target.value))}
                                                            placeholder="2001-2007..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Номер детали */}
                                                    <div>
                                                        <Label htmlFor="number-filter">Номер детали</Label>
                                                        <Input
                                                            id="number-filter"
                                                            type="text"
                                                            value={number}
                                                            onChange={(e) => setNumber(e.target.value)}
                                                            placeholder="52119-33939..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* OEM код */}
                                                    <div>
                                                        <Label htmlFor="oem-filter">OEM код</Label>
                                                        <Input
                                                            id="oem-filter"
                                                            type="text"
                                                            value={oemCode}
                                                            onChange={(e) => setOemCode(e.target.value)}
                                                            placeholder="89661-06G80..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Дефект */}
                                                    <div>
                                                        <Label htmlFor="defect-filter">Дефект</Label>
                                                        <Input
                                                            id="defect-filter"
                                                            type="text"
                                                            value={defect}
                                                            onChange={(e) => setDefect(e.target.value)}
                                                            placeholder="Царапина, скол, трещина..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Цвет */}
                                                    <div>
                                                        <Label htmlFor="color-filter">Цвет</Label>
                                                        <Input
                                                            id="color-filter"
                                                            type="text"
                                                            value={color}
                                                            onChange={(e) => setColor(e.target.value)}
                                                            placeholder="Черный, белый, серебристый..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>
                                                </div>
                                            </TabsContent>

                                            {/* Вкладка 3: КПП и расположение */}
                                            <TabsContent value="transmission" className="mt-4">
                                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                                    {/* Трансмиссия */}
                                                    <div>
                                                        <Label htmlFor="transmission-filter">Трансмиссия</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={transmission}
                                                                onValueChange={(value) => setTransmission(value)}
                                                                options={transmissionOptions}
                                                                placeholder="Все типы КПП"
                                                                searchPlaceholder="Поиск КПП..."
                                                                emptyMessage="Тип КПП не найден"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Модель КПП */}
                                                    <div>
                                                        <Label htmlFor="transmission-model-filter">Модель КПП</Label>
                                                        <Input
                                                            id="transmission-model-filter"
                                                            type="text"
                                                            value={transmissionModel}
                                                            onChange={(e) => setTransmissionModel(e.target.value)}
                                                            placeholder="U140F, RE4F04B..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Привод */}
                                                    <div>
                                                        <Label htmlFor="drive-filter">Привод</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={drive}
                                                                onValueChange={(value) => setDrive(value)}
                                                                options={driveOptions}
                                                                placeholder="Все приводы"
                                                                searchPlaceholder="Поиск привода..."
                                                                emptyMessage="Привод не найден"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Перед / зад */}
                                                    <div>
                                                        <Label htmlFor="front-rear-filter">Перед / зад</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={frontRear}
                                                                onValueChange={(value) => setFrontRear(value)}
                                                                options={frontRearOptions}
                                                                placeholder="Все расположения"
                                                                searchPlaceholder="Поиск..."
                                                                emptyMessage="Не найдено"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Право / лево */}
                                                    <div>
                                                        <Label htmlFor="left-right-filter">Право / лево</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={leftRight}
                                                                onValueChange={(value) => setLeftRight(value)}
                                                                options={leftRightOptions}
                                                                placeholder="Все стороны"
                                                                searchPlaceholder="Поиск..."
                                                                emptyMessage="Не найдено"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Верх / низ */}
                                                    <div>
                                                        <Label htmlFor="top-bottom-filter">Верх / низ</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={topBottom}
                                                                onValueChange={(value) => setTopBottom(value)}
                                                                options={topBottomOptions}
                                                                placeholder="Все положения"
                                                                searchPlaceholder="Поиск..."
                                                                emptyMessage="Не найдено"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>
                                                </div>
                                            </TabsContent>

                                            {/* Вкладка 4: Склад и продавец */}
                                            <TabsContent value="warehouse" className="mt-4">
                                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                                    {/* Местоположение */}
                                                    <div>
                                                        <Label htmlFor="location-filter">Местоположение (полка/стеллаж)</Label>
                                                        <Input
                                                            id="location-filter"
                                                            type="text"
                                                            value={location}
                                                            onChange={(e) => setLocation(e.target.value)}
                                                            placeholder="Shelf A-12..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Адрес склада */}
                                                    <div>
                                                        <Label htmlFor="address-filter">Адрес склада</Label>
                                                        <Input
                                                            id="address-filter"
                                                            type="text"
                                                            value={address}
                                                            onChange={(e) => setAddress(e.target.value)}
                                                            placeholder="Профсоюзная 2/11..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Продавец */}
                                                    <div>
                                                        <Label htmlFor="salesman-filter">Продавец</Label>
                                                        <Input
                                                            id="salesman-filter"
                                                            type="text"
                                                            value={salesman}
                                                            onChange={(e) => setSalesman(e.target.value)}
                                                            placeholder="Имя продавца..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Производитель */}
                                                    <div>
                                                        <Label htmlFor="manufacturer-filter">Производитель</Label>
                                                        <Input
                                                            id="manufacturer-filter"
                                                            type="text"
                                                            value={manufacturer}
                                                            onChange={(e) => setManufacturer(e.target.value)}
                                                            placeholder="Denso, Bosch, Brembo..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Код производителя */}
                                                    <div>
                                                        <Label htmlFor="manufacturer-code-filter">Код производителя</Label>
                                                        <Input
                                                            id="manufacturer-code-filter"
                                                            type="text"
                                                            value={manufacturerCode}
                                                            onChange={(e) => setManufacturerCode(e.target.value)}
                                                            placeholder="Код производителя..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Код поставщика */}
                                                    <div>
                                                        <Label htmlFor="supplier-code-filter">Код поставщика</Label>
                                                        <Input
                                                            id="supplier-code-filter"
                                                            type="text"
                                                            value={supplierCode}
                                                            onChange={(e) => setSupplierCode(e.target.value)}
                                                            placeholder="Код поставщика..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Процент износа */}
                                                    <div>
                                                        <Label htmlFor="wear-percentage-filter">Процент износа</Label>
                                                        <Input
                                                            id="wear-percentage-filter"
                                                            type="text"
                                                            value={wearPercentage}
                                                            onChange={(e) => setWearPercentage(e.target.value)}
                                                            placeholder="5%, 10%..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>
                                                </div>
                                            </TabsContent>

                                            {/* Вкладка 5: Шины и диски */}
                                            <TabsContent value="wheels" className="mt-4">
                                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                                    {/* Сезонность */}
                                                    <div>
                                                        <Label htmlFor="season-filter">Сезонность шин</Label>
                                                        <div className="mt-1.5">
                                                            <SearchableSelect
                                                                value={season}
                                                                onValueChange={(value) => setSeason(value)}
                                                                options={seasonOptions}
                                                                placeholder="Все сезоны"
                                                                searchPlaceholder="Поиск сезона..."
                                                                emptyMessage="Сезон не найден"
                                                                className="bg-transparent border border-gray-300"
                                                            />
                                                        </div>
                                                    </div>

                                                    {/* Модель шины */}
                                                    <div>
                                                        <Label htmlFor="tire-model-filter">Модель шины</Label>
                                                        <Input
                                                            id="tire-model-filter"
                                                            type="text"
                                                            value={tireModel}
                                                            onChange={(e) => setTireModel(e.target.value)}
                                                            placeholder="Hakkapeliitta 8, Ice Cruiser..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Диаметр */}
                                                    <div>
                                                        <Label htmlFor="diameter-filter">Диаметр</Label>
                                                        <Input
                                                            id="diameter-filter"
                                                            type="text"
                                                            value={diameter}
                                                            onChange={(e) => setDiameter(e.target.value)}
                                                            placeholder="R15, R16, R17..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Ширина шины */}
                                                    <div>
                                                        <Label htmlFor="width-filter">Ширина шины</Label>
                                                        <Input
                                                            id="width-filter"
                                                            type="text"
                                                            value={width}
                                                            onChange={(e) => setWidth(e.target.value)}
                                                            placeholder="205, 215, 225..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Профиль */}
                                                    <div>
                                                        <Label htmlFor="profile-filter">Профиль шины</Label>
                                                        <Input
                                                            id="profile-filter"
                                                            type="text"
                                                            value={profile}
                                                            onChange={(e) => setProfile(e.target.value)}
                                                            placeholder="55, 60, 65..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Количество шин */}
                                                    <div>
                                                        <Label htmlFor="tire-quantity-filter">Количество шин</Label>
                                                        <Input
                                                            id="tire-quantity-filter"
                                                            type="text"
                                                            value={tireQuantity}
                                                            onChange={(e) => setTireQuantity(e.target.value)}
                                                            placeholder="4 шт, пара..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Сверловка (PCD) */}
                                                    <div>
                                                        <Label htmlFor="drilling-filter">Сверловка (PCD)</Label>
                                                        <Input
                                                            id="drilling-filter"
                                                            type="text"
                                                            value={drilling}
                                                            onChange={(e) => setDrilling(e.target.value)}
                                                            placeholder="5x114.3, 4x100..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Вылет (ET) */}
                                                    <div>
                                                        <Label htmlFor="offset-filter">Вылет (ET)</Label>
                                                        <Input
                                                            id="offset-filter"
                                                            type="text"
                                                            value={offset}
                                                            onChange={(e) => setOffset(e.target.value)}
                                                            placeholder="ET45, ET38..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>

                                                    {/* Диаметр ЦО (DIA) */}
                                                    <div>
                                                        <Label htmlFor="center-hole-diameter-filter">Диаметр ЦО (DIA)</Label>
                                                        <Input
                                                            id="center-hole-diameter-filter"
                                                            type="text"
                                                            value={centerHoleDiameter}
                                                            onChange={(e) => setCenterHoleDiameter(e.target.value)}
                                                            placeholder="60.1, 67.1..."
                                                            className="mt-1.5 bg-transparent border border-gray-300"
                                                        />
                                                    </div>
                                                </div>
                                            </TabsContent>
                                        </Tabs>

                                        {/* Панель сброса и сводки под вкладками */}
                                        <div className="mt-4 pt-3 border-t flex flex-col sm:flex-row items-center justify-between gap-2">
                                            <div className="text-xs text-muted-foreground">
                                                {activeFiltersCount > 0 ? (
                                                    <span>Активно фильтров: <strong className="text-foreground">{activeFiltersCount}</strong></span>
                                                ) : (
                                                    <span>Фильтры не выбраны</span>
                                                )}
                                            </div>
                                            <Button
                                                onClick={clearFilters}
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                disabled={isSearching || (activeFiltersCount === 0 && !searchQuery)}
                                                className="gap-2 shrink-0 border-input"
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
                    <AnimatePresence mode="wait">
                        {(activeFiltersCount > 0 || searchQuery) && (
                            <motion.div
                                key="active-filters"
                                layout
                                initial={{ opacity: 0, y: -10 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -10 }}
                                transition={{ duration: 0.3, ease: 'easeOut' }}
                                className="flex flex-wrap gap-2"
                            >
                                <AnimatePresence>
                                    {searchQuery && (
                                        <motion.div
                                            key="search"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) clearSearchQuery();
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Поиск: {searchQuery}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {address && (
                                        <motion.div
                                            key="address"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setAddress('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Адрес: {address}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {category && (
                                        <motion.div
                                            key="category"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setCategory('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Категория: {category}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {brand && (
                                        <motion.div
                                            key="brand"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setBrand('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Бренд: {brand}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {model && (
                                        <motion.div
                                            key="model"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setModel('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Модель: {model}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {location && (
                                        <motion.div
                                            key="location"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setLocation('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Местоположение: {location}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {number && (
                                        <motion.div
                                            key="number"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setNumber('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Номер: {number}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {oemCode && (
                                        <motion.div
                                            key="oemCode"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setOemCode('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>OEM: {oemCode}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {vin && (
                                        <motion.div
                                            key="vin"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setVin('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>VIN: {vin}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {carReleasePeriod && (
                                        <motion.div
                                            key="carReleasePeriod"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setCarReleasePeriod('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Период: {carReleasePeriod}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {bodyBrand && (
                                        <motion.div
                                            key="bodyBrand"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setBodyBrand('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Кузов: {bodyBrand}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {engineBrand && (
                                        <motion.div
                                            key="engineBrand"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setEngineBrand('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>ДВС: {engineBrand}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {carReleaseDate && (
                                        <motion.div
                                            key="carReleaseDate"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setCarReleaseDate('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Год: {carReleaseDate}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {transmission && (
                                        <motion.div
                                            key="transmission"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setTransmission('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>КПП: {transmission}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {drive && (
                                        <motion.div
                                            key="drive"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setDrive('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Привод: {drive}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {condition && (
                                        <motion.div
                                            key="condition"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setCondition('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Состояние: {condition}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {manufacturer && (
                                        <motion.div
                                            key="manufacturer"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setManufacturer('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Производитель: {manufacturer}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {defect && (
                                        <motion.div
                                            key="defect"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setDefect('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Дефект: {defect}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {color && (
                                        <motion.div
                                            key="color"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setColor('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Цвет: {color}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {status && (
                                        <motion.div
                                            key="status"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setStatus('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Статус: {status === 'true' ? 'Доступно' : 'Недоступно'}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {hasPhoto && hasPhoto !== 'with' && (
                                        <motion.div
                                            key="hasPhoto"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setHasPhoto('with');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Фото: {hasPhoto === 'without' ? 'Без фото' : 'Все'}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {minPrice && (
                                        <motion.div
                                            key="minPrice"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setMinPrice('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Цена от: {minPrice} ₽</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {maxPrice && (
                                        <motion.div
                                            key="maxPrice"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setMaxPrice('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Цена до: {maxPrice} ₽</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {minQuantity && (
                                        <motion.div
                                            key="minQuantity"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setMinQuantity('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Кол-во от: {minQuantity}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {maxQuantity && (
                                        <motion.div
                                            key="maxQuantity"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setMaxQuantity('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Кол-во до: {maxQuantity}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {salesman && (
                                        <motion.div
                                            key="salesman"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setSalesman('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Продавец: {salesman}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {frontRear && (
                                        <motion.div
                                            key="frontRear"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setFrontRear('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Перед/зад: {frontRear}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {leftRight && (
                                        <motion.div
                                            key="leftRight"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setLeftRight('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Право/лево: {leftRight}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {topBottom && (
                                        <motion.div
                                            key="topBottom"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setTopBottom('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Верх/низ: {topBottom}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {manufacturerCode && (
                                        <motion.div
                                            key="manufacturerCode"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setManufacturerCode('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Код произв.: {manufacturerCode}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {supplierCode && (
                                        <motion.div
                                            key="supplierCode"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setSupplierCode('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Код поставщ.: {supplierCode}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {transmissionModel && (
                                        <motion.div
                                            key="transmissionModel"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setTransmissionModel('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Модель КПП: {transmissionModel}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {wearPercentage && (
                                        <motion.div
                                            key="wearPercentage"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setWearPercentage('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Износ: {wearPercentage}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {season && (
                                        <motion.div
                                            key="season"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setSeason('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Сезон: {season}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {diameter && (
                                        <motion.div
                                            key="diameter"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setDiameter('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Диаметр: {diameter}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {width && (
                                        <motion.div
                                            key="width"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setWidth('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Ширина: {width}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {profile && (
                                        <motion.div
                                            key="profile"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setProfile('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Профиль: {profile}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {tireQuantity && (
                                        <motion.div
                                            key="tireQuantity"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setTireQuantity('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Кол-во шин: {tireQuantity}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {drilling && (
                                        <motion.div
                                            key="drilling"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setDrilling('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Сверловка: {drilling}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {offset && (
                                        <motion.div
                                            key="offset"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setOffset('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Вылет: {offset}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {centerHoleDiameter && (
                                        <motion.div
                                            key="centerHoleDiameter"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setCenterHoleDiameter('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Диаметр ЦО: {centerHoleDiameter}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                <AnimatePresence>
                                    {tireModel && (
                                        <motion.div
                                            key="tireModel"
                                            initial={{ opacity: 0, scale: 0.8 }}
                                            animate={{ opacity: 1, scale: 1 }}
                                            exit={{ opacity: 0, scale: 0.8 }}
                                            transition={{ duration: 0.2 }}
                                        >
                                            <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                                                <Checkbox
                                                    checked={true}
                                                    onCheckedChange={(checked) => {
                                                        if (!checked) setTireModel('');
                                                    }}
                                                    className="h-3 w-3"
                                                />
                                                <span>Модель шины: {tireModel}</span>
                                            </Badge>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
            </CardContent>
        </Card>
        </motion.div>
    );
}

export default PartsSearch;
