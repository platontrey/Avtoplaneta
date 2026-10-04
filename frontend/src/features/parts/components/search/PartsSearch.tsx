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
import { formatCarReleasePeriod } from '@/lib/utils';

import { ActiveFilterBadges, type ActiveFilterItem } from './ActiveFilterBadges';
import { GeneralFiltersTab } from './GeneralFiltersTab';
import { BodyEngineFiltersTab } from './BodyEngineFiltersTab';
import { TransmissionFiltersTab } from './TransmissionFiltersTab';
import { WarehouseFiltersTab } from './WarehouseFiltersTab';
import { WheelsFiltersTab } from './WheelsFiltersTab';

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

export const PartsSearch: React.FC<PartsSearchProps> = ({
  onFiltersChange,
  onDisplayLimitChange,
  currentDisplayLimit,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState('');
  const [brand, setBrand] = useState('');
  const [model, setModel] = useState('');

  const { brandOptions, modelOptions, bodyOptions, engineOptions } = useVehicleOptions(brand, model);

  const [location, setLocation] = useState('');
  const [address, setAddress] = useState('');
  const [status, setStatus] = useState('');
  const [hasPhoto, setHasPhoto] = useState('all');
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

  const statusOptions: SelectOption[] = [
    { value: 'true', label: 'Доступно' },
    { value: 'false', label: 'Недоступно' },
  ];

  const photoOptions: SelectOption[] = [
    { value: 'all', label: 'Все' },
    { value: 'with', label: 'С фото' },
    { value: 'without', label: 'Без фото' },
  ];

  const transmissionOptions: SelectOption[] = [
    { value: 'АКПП', label: 'АКПП' },
    { value: 'МКПП', label: 'МКПП' },
    { value: 'Вариатор', label: 'Вариатор' },
    { value: 'Роботизированная', label: 'Роботизированная' },
  ];

  const driveOptions: SelectOption[] = [
    { value: 'Передний', label: 'Передний' },
    { value: 'Задний', label: 'Задний' },
    { value: 'Полный', label: 'Полный' },
  ];

  const frontRearOptions: SelectOption[] = [
    { value: 'Передний', label: 'Передний' },
    { value: 'Задний', label: 'Задний' },
  ];

  const leftRightOptions: SelectOption[] = [
    { value: 'Левый', label: 'Левый' },
    { value: 'Правый', label: 'Правый' },
  ];

  const topBottomOptions: SelectOption[] = [
    { value: 'Верхний', label: 'Верхний' },
    { value: 'Нижний', label: 'Нижний' },
  ];

  const seasonOptions: SelectOption[] = [
    { value: 'Лето', label: 'Лето' },
    { value: 'Зима', label: 'Зима' },
    { value: 'Всесезонная', label: 'Всесезонная' },
  ];

  const clearResults = useCallback(() => {
    setResults([]);
    prevResultsRef.current = [];
    setIsResultsVisible(false);
  }, []);

  const clearSearchQuery = useCallback(() => {
    setSearchQuery('');
    clearResults();
    setVoiceSearchError(null);
  }, [clearResults]);

  const clearFilters = () => {
    setSearchQuery('');
    setCategory('');
    setBrand('');
    setModel('');
    setLocation('');
    setAddress('');
    setStatus('');
    setHasPhoto('all');
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
        hasPhoto: hasPhoto === 'all' ? '' : hasPhoto,
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
    onFiltersChange,
  ]);

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
      performSearch(searchQuery);
    }, SEARCH_DEBOUNCE_DELAY);

    return () => clearTimeout(timeoutId);
  }, [searchQuery, performSearch]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      applyFilters();
    }, FILTER_DEBOUNCE_DELAY);

    return () => clearTimeout(timeoutId);
  }, [searchQuery, category, brand, model, location, address, status, hasPhoto, applyFilters]);

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

  const stopVoiceSearch = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.stop();
    }
    setIsListening(false);
  }, []);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  }, []);

  const handleResultClick = (result: Part) => {
    setSearchQuery(result.name);
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
  const activeFiltersCount = useMemo(
    () =>
      [
        category, brand, model, location, address, status,
        hasPhoto && hasPhoto !== 'all' ? hasPhoto : '',
        number, oemCode, vin, carReleasePeriod, bodyBrand, engineBrand, carReleaseDate,
        transmission, drive, condition, manufacturer, defect, color,
        salesman, minPrice, maxPrice, minQuantity, maxQuantity,
        frontRear, leftRight, topBottom, manufacturerCode, supplierCode,
        transmissionModel, wearPercentage, season, diameter, width,
        profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel,
      ].filter(Boolean).length,
    [
      category, brand, model, location, address, status, hasPhoto,
      number, oemCode, vin, carReleasePeriod, bodyBrand, engineBrand, carReleaseDate,
      transmission, drive, condition, manufacturer, defect, color,
      salesman, minPrice, maxPrice, minQuantity, maxQuantity,
      frontRear, leftRight, topBottom, manufacturerCode, supplierCode,
      transmissionModel, wearPercentage, season, diameter, width,
      profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel,
    ]
  );

  const mainFiltersCount = useMemo(
    () =>
      [
        category, brand, model, carReleaseDate, condition,
        minPrice, maxPrice, minQuantity, maxQuantity, status,
        hasPhoto && hasPhoto !== 'all' ? hasPhoto : '',
      ].filter(Boolean).length,
    [category, brand, model, carReleaseDate, condition, minPrice, maxPrice, minQuantity, maxQuantity, status, hasPhoto]
  );

  const bodyEngineFiltersCount = useMemo(
    () =>
      [bodyBrand, engineBrand, vin, carReleasePeriod, number, oemCode, defect, color].filter(Boolean).length,
    [bodyBrand, engineBrand, vin, carReleasePeriod, number, oemCode, defect, color]
  );

  const transmissionFiltersCount = useMemo(
    () =>
      [transmission, transmissionModel, drive, frontRear, leftRight, topBottom].filter(Boolean).length,
    [transmission, transmissionModel, drive, frontRear, leftRight, topBottom]
  );

  const warehouseFiltersCount = useMemo(
    () =>
      [location, address, salesman, manufacturer, manufacturerCode, supplierCode, wearPercentage].filter(Boolean).length,
    [location, address, salesman, manufacturer, manufacturerCode, supplierCode, wearPercentage]
  );

  const wheelsFiltersCount = useMemo(
    () =>
      [season, diameter, width, profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel].filter(Boolean).length,
    [season, diameter, width, profile, tireQuantity, drilling, offset, centerHoleDiameter, tireModel]
  );

  // Формирование активных фильтров для бейджей
  const activeFilterItems = useMemo<ActiveFilterItem[]>(() => {
    const items: ActiveFilterItem[] = [];
    if (searchQuery.trim()) {
      const multiTerms = searchQuery
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
              setSearchQuery(remaining.join(', '));
            },
          });
        });
      } else {
        items.push({ id: 'search', label: `Поиск: ${searchQuery}`, onClear: clearSearchQuery });
      }
    }
    if (address) items.push({ id: 'address', label: `Адрес склада: ${address}`, onClear: () => setAddress('') });
    if (bodyBrand) items.push({ id: 'bodyBrand', label: `Марка кузова: ${bodyBrand}`, onClear: () => setBodyBrand('') });
    if (brand) items.push({ id: 'brand', label: `Марка авто: ${brand}`, onClear: () => setBrand('') });
    if (carReleaseDate) items.push({ id: 'carReleaseDate', label: `Год выпуска: ${carReleaseDate}`, onClear: () => setCarReleaseDate('') });
    if (carReleasePeriod) items.push({ id: 'carReleasePeriod', label: `Период выпуска: ${formatCarReleasePeriod(carReleasePeriod)}`, onClear: () => setCarReleasePeriod('') });
    if (category) items.push({ id: 'category', label: `Категория: ${category}`, onClear: () => setCategory('') });
    if (centerHoleDiameter) items.push({ id: 'centerHoleDiameter', label: `Диаметр ЦО: ${centerHoleDiameter} мм`, onClear: () => setCenterHoleDiameter('') });
    if (color) items.push({ id: 'color', label: `Цвет: ${color}`, onClear: () => setColor('') });
    if (condition) items.push({ id: 'condition', label: `Состояние: ${condition}`, onClear: () => setCondition('') });
    if (defect) items.push({ id: 'defect', label: `Дефект: ${defect}`, onClear: () => setDefect('') });
    if (diameter) items.push({ id: 'diameter', label: `Диаметр: R${diameter}`, onClear: () => setDiameter('') });
    if (drilling) items.push({ id: 'drilling', label: `Сверловка: ${drilling}`, onClear: () => setDrilling('') });
    if (drive) items.push({ id: 'drive', label: `Привод: ${drive}`, onClear: () => setDrive('') });
    if (engineBrand) items.push({ id: 'engineBrand', label: `Марка двигателя: ${engineBrand}`, onClear: () => setEngineBrand('') });
    if (frontRear) items.push({ id: 'frontRear', label: `Расположение: ${frontRear}`, onClear: () => setFrontRear('') });
    if (hasPhoto && hasPhoto !== 'all') items.push({ id: 'hasPhoto', label: `Фото: ${hasPhoto === 'with' ? 'С фото' : 'Без фото'}`, onClear: () => setHasPhoto('all') });
    if (leftRight) items.push({ id: 'leftRight', label: `Сторона: ${leftRight}`, onClear: () => setLeftRight('') });
    if (location) items.push({ id: 'location', label: `Местоположение: ${location}`, onClear: () => setLocation('') });
    if (manufacturer) items.push({ id: 'manufacturer', label: `Производитель: ${manufacturer}`, onClear: () => setManufacturer('') });
    if (manufacturerCode) items.push({ id: 'manufacturerCode', label: `Код производителя: ${manufacturerCode}`, onClear: () => setManufacturerCode('') });
    if (maxPrice) items.push({ id: 'maxPrice', label: `Цена до: ${maxPrice} ₽`, onClear: () => setMaxPrice('') });
    if (maxQuantity) items.push({ id: 'maxQuantity', label: `Количество до: ${maxQuantity} шт`, onClear: () => setMaxQuantity('') });
    if (minPrice) items.push({ id: 'minPrice', label: `Цена от: ${minPrice} ₽`, onClear: () => setMinPrice('') });
    if (minQuantity) items.push({ id: 'minQuantity', label: `Количество от: ${minQuantity} шт`, onClear: () => setMinQuantity('') });
    if (model) items.push({ id: 'model', label: `Модель: ${model}`, onClear: () => setModel('') });
    if (number) items.push({ id: 'number', label: `Номер детали: ${number}`, onClear: () => setNumber('') });
    if (oemCode) items.push({ id: 'oemCode', label: `OEM номер: ${oemCode}`, onClear: () => setOemCode('') });
    if (offset) items.push({ id: 'offset', label: `Вылет (ET): ${offset} мм`, onClear: () => setOffset('') });
    if (profile) items.push({ id: 'profile', label: `Профиль шины: ${profile}%`, onClear: () => setProfile('') });
    if (salesman) items.push({ id: 'salesman', label: `Продавец: ${salesman}`, onClear: () => setSalesman('') });
    if (season) items.push({ id: 'season', label: `Сезонность: ${season}`, onClear: () => setSeason('') });
    if (status) items.push({ id: 'status', label: `Статус: ${status === 'true' ? 'Доступно' : 'Недоступно'}`, onClear: () => setStatus('') });
    if (supplierCode) items.push({ id: 'supplierCode', label: `Код поставщика: ${supplierCode}`, onClear: () => setSupplierCode('') });
    if (tireQuantity) items.push({ id: 'tireQuantity', label: `Количество шин: ${tireQuantity} шт`, onClear: () => setTireQuantity('') });
    if (topBottom) items.push({ id: 'topBottom', label: `Положение: ${topBottom}`, onClear: () => setTopBottom('') });
    if (transmission) items.push({ id: 'transmission', label: `КПП: ${transmission}`, onClear: () => setTransmission('') });
    if (transmissionModel) items.push({ id: 'transmissionModel', label: `Модель КПП: ${transmissionModel}`, onClear: () => setTransmissionModel('') });
    if (vin) items.push({ id: 'vin', label: `VIN: ${vin}`, onClear: () => setVin('') });
    if (wearPercentage) items.push({ id: 'wearPercentage', label: `Износ: ${wearPercentage}%`, onClear: () => setWearPercentage('') });
    if (width) items.push({ id: 'width', label: `Ширина: ${width} мм`, onClear: () => setWidth('') });
    if (tireModel) items.push({ id: 'tireModel', label: `Модель шины: ${tireModel}`, onClear: () => setTireModel('') });
    return items;
  }, [
    searchQuery, clearSearchQuery, address, bodyBrand, brand, carReleaseDate, carReleasePeriod, category,
    centerHoleDiameter, color, condition, defect, diameter, drilling, drive, engineBrand,
    frontRear, hasPhoto, leftRight, location, manufacturer, manufacturerCode, maxPrice,
    maxQuantity, minPrice, minQuantity, model, number, oemCode, offset, profile, salesman,
    season, status, supplierCode, tireQuantity, topBottom, transmission, transmissionModel,
    vin, wearPercentage, width, tireModel,
  ]);

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
                  {searchQuery.length > 0 && searchQuery.length < MIN_SEARCH_LENGTH && (
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
                            brand={brand}
                            setBrand={setBrand}
                            model={model}
                            setModel={setModel}
                            category={category}
                            setCategory={setCategory}
                            carReleaseDate={carReleaseDate}
                            setCarReleaseDate={setCarReleaseDate}
                            condition={condition}
                            setCondition={setCondition}
                            minPrice={minPrice}
                            setMinPrice={setMinPrice}
                            maxPrice={maxPrice}
                            setMaxPrice={setMaxPrice}
                            minQuantity={minQuantity}
                            setMinQuantity={setMinQuantity}
                            maxQuantity={maxQuantity}
                            setMaxQuantity={setMaxQuantity}
                            status={status}
                            setStatus={setStatus}
                            hasPhoto={hasPhoto}
                            setHasPhoto={setHasPhoto}
                            brandOptions={brandOptions}
                            modelOptions={modelOptions}
                            categoryOptions={categoryOptions}
                            statusOptions={statusOptions}
                            photoOptions={photoOptions}
                          />
                        </TabsContent>

                        <TabsContent value="bodyEngine" className="mt-4">
                          <BodyEngineFiltersTab
                            brand={brand}
                            bodyBrand={bodyBrand}
                            setBodyBrand={setBodyBrand}
                            engineBrand={engineBrand}
                            setEngineBrand={setEngineBrand}
                            vin={vin}
                            setVin={setVin}
                            carReleasePeriod={carReleasePeriod}
                            setCarReleasePeriod={setCarReleasePeriod}
                            number={number}
                            setNumber={setNumber}
                            oemCode={oemCode}
                            setOemCode={setOemCode}
                            defect={defect}
                            setDefect={setDefect}
                            color={color}
                            setColor={setColor}
                            bodyOptions={bodyOptions}
                            engineOptions={engineOptions}
                          />
                        </TabsContent>

                        <TabsContent value="transmission" className="mt-4">
                          <TransmissionFiltersTab
                            transmission={transmission}
                            setTransmission={setTransmission}
                            transmissionModel={transmissionModel}
                            setTransmissionModel={setTransmissionModel}
                            drive={drive}
                            setDrive={setDrive}
                            frontRear={frontRear}
                            setFrontRear={setFrontRear}
                            leftRight={leftRight}
                            setLeftRight={setLeftRight}
                            topBottom={topBottom}
                            setTopBottom={setTopBottom}
                            transmissionOptions={transmissionOptions}
                            driveOptions={driveOptions}
                            frontRearOptions={frontRearOptions}
                            leftRightOptions={leftRightOptions}
                            topBottomOptions={topBottomOptions}
                          />
                        </TabsContent>

                        <TabsContent value="warehouse" className="mt-4">
                          <WarehouseFiltersTab
                            location={location}
                            setLocation={setLocation}
                            address={address}
                            setAddress={setAddress}
                            salesman={salesman}
                            setSalesman={setSalesman}
                            manufacturer={manufacturer}
                            setManufacturer={setManufacturer}
                            manufacturerCode={manufacturerCode}
                            setManufacturerCode={setManufacturerCode}
                            supplierCode={supplierCode}
                            setSupplierCode={setSupplierCode}
                            wearPercentage={wearPercentage}
                            setWearPercentage={setWearPercentage}
                          />
                        </TabsContent>

                        <TabsContent value="wheels" className="mt-4">
                          <WheelsFiltersTab
                            season={season}
                            setSeason={setSeason}
                            tireModel={tireModel}
                            setTireModel={setTireModel}
                            diameter={diameter}
                            setDiameter={setDiameter}
                            width={width}
                            setWidth={setWidth}
                            profile={profile}
                            setProfile={setProfile}
                            tireQuantity={tireQuantity}
                            setTireQuantity={setTireQuantity}
                            drilling={drilling}
                            setDrilling={setDrilling}
                            offset={offset}
                            setOffset={setOffset}
                            centerHoleDiameter={centerHoleDiameter}
                            setCenterHoleDiameter={setCenterHoleDiameter}
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
