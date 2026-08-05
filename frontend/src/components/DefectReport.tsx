/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ClearableSelect } from "@/components/ClearableSelect";
import { SelectItem } from "@/components/ui/select";
import { Save, ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";
import { getAuthHeaders } from "@/lib/csrf";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { API_BASE_URL } from '@/lib/api';

const availableColors = [
  "Черный",
  "Белый",
  "Серебристый",
  "Серый",
  "Темно-серый",
  "Светло-серый",
  "Синий",
  "Темно-синий",
  "Светло-синий",
  "Красный",
  "Темно-красный",
  "Бордовый",
  "Зеленый",
  "Темно-зеленый",
  "Светло-зеленый",
  "Желтый",
  "Оранжевый",
  "Фиолетовый",
  "Коричневый",
  "Бежевый",
  "Золотой",
  "Бронзовый",
  "Перламутровый",
  "Металлик",
  "Матовый",
];

const transmissionModelCategories = [
  "Подвеска ДВС/КПП",
  "Трансмиссия",
  "Подвеска передних колес",
];

const defectReportSchema = z.object({
  brand: z.string().min(1, "Выберите бренд"),
  model: z.string().min(1, "Введите модель"),
  year: z.number().min(1900, "Введите корректный год").max(new Date().getFullYear() + 1, "Год не может быть в будущем"),
  vin: z.string().optional(),
  mileage: z.number().min(0, "Пробег должен быть положительным числом"),
  transmission: z.string().optional(),
  transmission_model: z.string().optional(),
  engine_brand: z.string().optional(),
  body_brand: z.string().optional(),
  interior_color: z.string().optional(),
  body_color: z.string().optional(),
  description: z.string().max(1000, "Описание слишком длинное (макс 1000 символов)").optional(),
});

type DefectReportFormData = z.infer<typeof defectReportSchema>;


interface CommonPart {
  name: string;
  category: string;
  description?: string;
  quantity?: number;
  price?: number;
  // Характеристики запчасти
  body_brand?: string;
  engine_brand?: string;
  car_release_date?: string;
  front_rear?: string;
  left_right?: string;
  top_bottom?: string;
  number?: string;
  manufacturer?: string;
  manufacturer_code?: string;
  oem_code?: string;
  color?: string;
  condition?: string;
  supplier_code?: string;
  defect?: string;
  transmission?: string;
  transmission_model?: string;
  drive?: string;
  wear_percentage?: string;
  season?: string;
  diameter?: string;
  width?: string;
  profile?: string;
  tire_quantity?: string;
  drilling?: string;
  offset?: string;
  center_hole_diameter?: string;
  tire_model?: string;
}

export default function DefectReport() {
  const [loading, setLoading] = useState<boolean>(false);
  const [displayLimit, setDisplayLimit] = useState<number>(10);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<DefectReportFormData>({
    resolver: zodResolver(defectReportSchema),
    defaultValues: {
      description: "В связи с изменением цены конечную стоимость товара узнавать по WhatsApp 89138538227",
    },
  });

  const brand = watch("brand");

  // Список запчастей для дефектной ведомости
  const commonParts: CommonPart[] = [
       // === Категория: Двигатель (Engine) ===
       // Запчасти, относящиеся к двигателю автомобиля, включая компоненты системы впрыска, охлаждения, смазки и т.д.
       { name: "Абсорбер", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Амортизатор натяжителя ремня", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Аутлет", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Аутлет", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Аутлет", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Аутлет", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Башмак натяжителя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Башмак успокоителя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Блок цилиндров", category: "Двигатель", quantity: 0, price: 50000 },
       { name: "Болт головки блока цилиндров", category: "Двигатель", quantity: 0, price: 100 },
       { name: "Болт маховика", category: "Двигатель", quantity: 0, price: 100 },
       { name: "Болт шкива коленвала", category: "Двигатель", quantity: 0, price: 100 },
       { name: "Вакуумный ресивер", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Вакуумный насос", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Головка блока цилиндров", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Головка блока цилиндров", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Головка блока цилиндров", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Головка блока цилиндров", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Головка блока цилиндров", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Датчик давления масла двигателя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Датчик давления топлива", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Даун-пайп", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Двигатель ", category: "Двигатель", quantity: 0, price: 100000 },
       { name: "Декоративная крышка двигателя", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Декоративная крышка двигателя", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Декоративная крышка двигателя", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Декоративная крышка двигателя", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Декоративная крышка двигателя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Демпфер топливный", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Держатель топливной рейки", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Держатель топливной рейки", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Дроссельная заслонка", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Звезда распредвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Звездочка грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Интеркулер", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Карбюратор", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан egr", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан вакуумный", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан вентиляции картерных газов", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан вентиляции топливного бака", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан впускной ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапан выпускной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Клапанная крышка", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Клапанная крышка", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Клапанная крышка", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Клапанная крышка", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Клапанная крышка", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Кожух выпускного коллектора", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Кожух выпускного коллектора", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Кожух выпускного коллектора", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Кожух выпускного коллектора", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Кожух выпускного коллектора", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Коленвал", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Коллектор впускной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Коллектор выпускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Коллектор выпускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Коллектор выпускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Коллектор выпускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Коллектор выпускной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Корпус воздушного фильтра", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Крепление генератора", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Крепление масляного фильтра", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Кронштейн ролика натяжителя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Крышка двигателя задняя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Крышка маслозаливной горловины", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Верх" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Низ" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Верх" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Низ" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "R", top_bottom: "Верх" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "R", top_bottom: "Низ" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "L", top_bottom: "Верх" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "L", top_bottom: "Низ" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крышка ремня грм ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Лобовина грм", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Лобовина грм", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Лобовина грм", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Лобовина грм", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Лобовина грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Маслоприемник", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Масляный насос", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Маховик акпп", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Маховик мкпп", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Механизм изменения длины впускного коллектора", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Муфта vvt-i ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Направляющая масляного щупа", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Насос дополнительной подачи воздуха", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Насос ручной подкачки", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель  ремня генератора ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель ремня гидроусилителя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель ремня грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель ремня навесного оборудования", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель ремня навесного оборудования ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель цепи грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель цепи грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжитель цепи грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Натяжной ролик ремня грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Обводной ролик ремня грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Патрубок воздушного фильтра", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Патрубок турбины", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Планка натяжителя ремня", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Пластина между двигателем и акпп", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поддон масляный двигателя", category: "Двигатель", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Поддон масляный двигателя", category: "Двигатель", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Полукольцо коленвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Поршень с шатуном", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Пружина клапана", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Распредвал впускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Распредвал впускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Распредвал впускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Распредвал впускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Распредвал впускной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Распредвал выпускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Распредвал выпускной", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Распредвал выпускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Распредвал выпускной", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Распредвал выпускной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Регулятор давления топлива", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Регулятор положения распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Регулятор положения распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Регулятор положения распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Регулятор положения распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Регулятор положения распредвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя генератора", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя ремня гидроусилителя", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя ремня навесного оборудования", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя ремня навесного оборудования", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Ролик обводной", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Скоба крепления форсунки", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Сухарь клапана", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Топливная рейка", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка охлаждающей жидкости", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка топливная ", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка картерных газов", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Трубка форсунки тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Турбина", category: "Двигатель", quantity: 0, price: 30000 },
       { name: "Форсунка топливная", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Форсунка холодного пуска", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Цепь грм", category: "Двигатель", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Цепь грм", category: "Двигатель", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Цепь грм", category: "Двигатель", quantity: 0, price: 1000, top_bottom: "Середина" },
       { name: "Шайба клапана регулировочная", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шайба маховика", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шайба пружины клапана", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шестерня грм", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шестерня коленвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "R" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000, left_right: "L" },
       { name: "Шестерня распредвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шестерня тнвд", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шкив вискомуфты", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шкив коленвала", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шланг вентиляции картерных газов", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Шланг топливной рейки", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Щуп масляный", category: "Двигатель", quantity: 0, price: 1000 },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Низ" },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "L", top_bottom: "Низ" },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "R", top_bottom: "Низ" },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Низ" },
       { name: "Подушка двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Верх" },
       { name: "Подушка двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "L", top_bottom: "Верх" },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "R", top_bottom: "Верх" },
       { name: "Подушка двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Верх" },
       { name: "Крепление опоры двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Нижнее" },
       { name: "Крепление опоры двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "L", top_bottom: "Нижнее" },
       { name: "Крепление опоры двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "R", top_bottom: "Нижнее" },
       { name: "Крепление опоры двигателя ", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Нижнее" },
       { name: "Крепление опоры двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", top_bottom: "Верхнее" },
       { name: "Крепление опоры двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "L", top_bottom: "Верхнее" },
       { name: "Крепление опоры двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, left_right: "R", top_bottom: "Верхнее" },
       { name: "Крепление опоры двигателя", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", top_bottom: "Верхнее" },
       { name: "Подушка кпп", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000 },
       { name: "Крепление опоры кпп", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000 },
       { name: "Балка продольная", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000 },
       { name: "Балка под двс", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000 },
       { name: "Балка под кпп", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000 },
       { name: "Кронштейн балки", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кронштейн балки", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Кронштейн балки", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Кронштейн балки", category: "Подвеска ДВС/КПП", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Бачок расширительный", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Вискомуфта", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Диффузор радиатора охлаждения двигателя", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Дополнительная помпа", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Клапан вакуумный", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Клапан отопителя", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Кожух радиатора печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Корпус печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Корпус термостата", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Кран печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Крепление бачка расширительного", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Крепление вискомуфты", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Крепление радиатора охлаждения двигателя", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крепление радиатора охлаждения двигателя", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крыльчатка вентилятора", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Крышка термостата ", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Патрубок печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "R" },
       { name: "Патрубок печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "L" },
       { name: "Патрубок печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Патрубок печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Патрубок радиатора", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Патрубок радиатора", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Помпа", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Радиатор акпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Радиатор мкпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Радиатор охлаждения двигателя", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Радиатор печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Теплообменник", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Термостат", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Тросик печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Тросик печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Тросик печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Трубка охлаждения акпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "R" },
       { name: "Трубка охлаждения акпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "L" },
       { name: "Трубка охлаждения мкпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "R" },
       { name: "Трубка охлаждения мкпп", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "L" },
       { name: "Трубка патрубка печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "R" },
       { name: "Трубка патрубка печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, left_right: "L" },
       { name: "Трубка патрубка печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Трубка патрубка печки", category: "Система охлаждения и отопления", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Шкив помпы", category: "Система охлаждения и отопления", quantity: 0, price: 1000 },
       { name: "Акпп", category: "Трансмиссия", quantity: 0, price: 50000 },
       { name: "Бачок  сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вал вторичный акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вал первичный акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вал промежуточный акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вал селектора", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вилка коробки передач", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Вилка сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "гидроаккумулятор акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "гидроблок", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Гидротрансформатор акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Главный цилиндр сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Датчик положения селектора акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "датчик скорости", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Диск стальной акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Диск сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Диск фрикционный акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Дифференциал", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Игольчатый подшипник", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Карданный вал", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Карданный вал ", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Корзина сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Корпус акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Корпус акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Корпус акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Кронштейн крепления карданного вала", category: "Трансмиссия", quantity: 0, price: 1000, left_right: "R" },
       { name: "Кронштейн крепления карданного вала", category: "Трансмиссия", quantity: 0, price: 1000, left_right: "L" },
       { name: "Кронштейн крепления раздаточной коробки", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Крышка задняя акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Кулиса кпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Лента тормозная акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Масляный насос акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Мкпп", category: "Трансмиссия", quantity: 0, price: 30000 },
       { name: "Мост", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Мост", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Муфта карданного вала", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Опора редуктора", category: "Трансмиссия", quantity: 0, price: 1000, left_right: "R" },
       { name: "Опора редуктора", category: "Трансмиссия", quantity: 0, price: 1000, left_right: "L" },
       { name: "Поддон акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Подшипник выжимной", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Поршень пакета акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Поршень пакета акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Поршень пакета акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Поршень пакета акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Привод", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Привод", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Привод", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Привод", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Проводка акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Рабочий цилиндр сцепления", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Раздаточная коробка", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Редуктор", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Редуктор ", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Редуктор  моста", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Редуктор  моста", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Ретайнер", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Ретайнер", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Ретайнер", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Ручка рычага переключения скоростей", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Рычаг включения раздаточной коробки", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Сервомеханизм ленточного тормоза", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Соленоид акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Ступица", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ступица", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ступица", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ступица", category: "Трансмиссия", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Трос кикдауна ", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Трос переключения акпп ", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Трос переключения мкпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Тяга кулисы кпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Фиксатор положения парковки", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Центробежный регулятор акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Шестерня акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Шестерня акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Щуп акпп", category: "Трансмиссия", quantity: 0, price: 1000 },
       { name: "Амортизатор ", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Амортизатор ", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Болт развальный", category: "Подвеска передних колес", quantity: 0, price: 100, front_rear: "F", left_right: "L" },
       { name: "Болт развальный", category: "Подвеска передних колес", quantity: 0, price: 100, front_rear: "F", left_right: "R" },
       { name: "Бушинг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Бушинг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Вилка амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Вилка амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Гайка развальная", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Гайка развальная", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Демпфер динамический", category: "Подвеска передних колес", quantity: 0, price: 1000 },
       { name: "Крепление рычага подвески", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крепление рычага подвески", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крепление стабилизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крепление стабилизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крепление торсиона", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крепление торсиона", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крепление торсиона", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крепление торсиона", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Кронштейн шаровой опоры", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кронштейн шаровой опоры", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крышка ступицы", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крышка ступицы", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Кулак поворотный", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кулак поворотный", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Опора амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Опора амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Отбойник амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Отбойник амортизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Прокладка под пружину", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Прокладка под пружину", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Пружина", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пружина", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Пыльник стойки", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пыльник стойки", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Верх" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Низ" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Низ" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Верх" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Низ" },
       { name: "Рычаг", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Низ" },
       { name: "Стабилизатор поперечной устойчивости", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Стойка стабилизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Стойка стабилизатора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Торсион", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Торсион", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Чашка пружины", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Чашка пружины", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Шайба развальная", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Шайба развальная", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Шаровая опора", category: "Подвеска передних колес", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Амортизатор ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Амортизатор ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Балка  задняя", category: "Подвеска задних колес", quantity: 0, price: 1000 },
       { name: "Болт развальный", category: "Подвеска задних колес", quantity: 0, price: 100, front_rear: "R", left_right: "R" },
       { name: "Болт развальный", category: "Подвеска задних колес", quantity: 0, price: 100, front_rear: "R", left_right: "L" },
       { name: "Бушинг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Бушинг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Вилка амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Вилка амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крепление стабилизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крепление стабилизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Опора амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Опора амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Отбойник амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Отбойник амортизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Прокладка под пружину", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Прокладка под пружину", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Пружина", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Пружина", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Пыльник стойки", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Пыльник стойки", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Рессора", category: "Подвеска задних колес", quantity: 0, price: 1000, left_right: "R" },
       { name: "Рессора", category: "Подвеска задних колес", quantity: 0, price: 1000, left_right: "L" },
       { name: "Рычаг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R", top_bottom: "Верх" },
       { name: "Рычаг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R", top_bottom: "Низ" },
       { name: "Рычаг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L", top_bottom: "Верх" },
       { name: "Рычаг", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L", top_bottom: "Низ" },
       { name: "Рычаг ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Рычаг ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Стабилизатор поперечной устойчивости", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Стойка стабилизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Стойка стабилизатора", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тяга поперечная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тяга поперечная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тяга поперечная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тяга поперечная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тяга поперечная ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тяга поперечная ", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тяга продольная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тяга продольная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тяга продольная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тяга продольная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Цапфа", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Цапфа", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Чашка пружины", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Чашка пружины", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Шайба развальная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Шайба развальная", category: "Подвеска задних колес", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Бачок гидроусилителя руля", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Карданчик рулевой", category: "Рулевое управление", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Карданчик рулевой", category: "Рулевое управление", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Клапан гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Кожух рулевой колонки", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Крепление бачка гидроусилителя ", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Крепление насоса гидроусилителя руля", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Крепление рулевой рейки", category: "Рулевое управление", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крепление рулевой рейки", category: "Рулевое управление", quantity: 0, price: 1000, left_right: "L" },
       { name: "Насос гидроусилителя руля", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Пыльник рулевой колонки", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Радиатор гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Рулевая колонка", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Рулевая рейка", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Рулевой маятник", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Рулевой редуктор", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Рулевой редуктор угловой", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Сошка рулевая", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Трубка гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Трубка гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Трапеция рулевая", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Тяга рулевая", category: "Рулевое управление", quantity: 0, price: 1000, left_right: "R" },
       { name: "Тяга рулевая", category: "Рулевое управление", quantity: 0, price: 1000, left_right: "L" },
       { name: "Шкив насоса гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Шланг гидроусилителя", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Шланг гидроусилителя ", category: "Рулевое управление", quantity: 0, price: 1000 },
       { name: "Кузов снаружи", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Абсорбер бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Арка заднего колеса ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Арка заднего колеса ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Бампер", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Защита бампера ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Бампер", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Бампер средняя часть ", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Бампер средняя часть ", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Боковина крыши", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Боковина крыши", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Брызговик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Брызговик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Брызговик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Брызговик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ветровик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ветровик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ветровик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ветровик", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Горловина топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Дверь багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Держатель дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Держатель дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Держатель дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Держатель капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Колпачек гайки держателя дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Колпачек гайки держателя дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Колпачек гайки держателя дворника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Дефлектор капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Дефлектор радиатора", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Дефлектор радиатора", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Заглушка бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Заглушка бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Заглушка буксировочного крюка", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Заглушка буксировочного крюка", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Замок двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Замок капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Замок крышки багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Замок лючка бензобака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Замок стекла двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Замок стекла двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Замок стекла собачника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Замок стекла собачника", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Защита горловины топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Защита двигателя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Защита двигателя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Защита двигателя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Защита двигателя", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Защита днища кузова", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Защита днища кузова", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Защита радиатора охолаждения", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Защита радиатора охолаждения", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Защита топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Защита топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Защита топливных трубок", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Зеркало заднего вида боковое", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Зеркало заднего вида боковое", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Зеркало на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Зеркальный элемент", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Зеркальный элемент", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Капот", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Кенгурятник", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Клапан вентиляционный", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Клапан вентиляционный", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Клык бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Клык бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Клык бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Клык бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Колпак запасного колеса", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Корпус зеркала заднего вида", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Корпус зеркала заднего вида", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крепление бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Крепление бокового  стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крепление бокового  стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крепление запасного колеса", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Кронштейн усилителя бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кронштейн усилителя бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Кронштейн усилителя бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Кронштейн усилителя бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крыша", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Крышка багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Крышка бензобака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Крышка зеркала заднего вида", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Крышка зеркала заднего вида", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Крышка форсунки омывателя фар", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крышка форсунки омывателя фар", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Личинка замка", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Личинка замка", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Лонжерон", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Лонжерон", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Чашка кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Чашка кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Люк", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Люк ", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Лючок бензонасоса", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Лючок бензобака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Молдинг на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Молдинг бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг дверной форточки", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг дверной форточки", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг дверной форточки", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг дверной форточки", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Молдинг крыши", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Молдинг крыши", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Молдинг лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Молдинг лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Молдинг лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Молдинг лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Молдинг на дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг на дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг на дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг на дверь", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг на кузов", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Молдинг на кузов", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Молдинг стекла заднего", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Молдинг стекла заднего", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Молдинг стекла заднего", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Молдинг стекла заднего", category: "Кузов снаружи", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Молдинг стекла кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг стекла кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг стекла наружный", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг стекла наружный", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг стекла наружный", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг стекла наружный", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Молдинг форточки кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг форточки кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг форточки кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг форточки кузова", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Накладка двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, top_bottom: "Низ" },
       { name: "Накладка крышки багажника ", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Накладка на дверь наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на дверь наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на дверь наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на дверь наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на крыло", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка на переднюю панель", category: "Кузов снаружи", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Накладка на порог наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Накладка на порог наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Накладка на рамку двери наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на рамку двери наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на рамку двери наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на рамку двери наружняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ограничитель двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ограничитель двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ограничитель двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ограничитель двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ограничитель двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Панель кузова задняяя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Панель кузова передняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Панель стекла заднего", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Верх" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Низ" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Верх" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Низ" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L", top_bottom: "Верх" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L", top_bottom: "Низ" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R", top_bottom: "Верх" },
       { name: "Петля двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R", top_bottom: "Низ" },
       { name: "Петля двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Петля двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Петля крышки багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Петля крышки багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Планка под фару", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Планка под фару", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Планка под фару", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Планка замка капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Подкрылок", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подкрылок", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Подкрылок", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Подкрылок", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Подножка", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Подножка", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Подушка рамы/кузова", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Порог со средней стойкой", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Порог со средней стойкой", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Прокладка заливной горловины", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Рама", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Рейлинг", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Рейлинг", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Решетка бамперная", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Решетка бамперная", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Решетка бамперная", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Решетка под дворники ", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Решетка под дворники ", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Решетка под дворники ", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Решетка радиатора", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Ручка двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Ручка двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ручка двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ручка двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ручка двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ручка сдвижной двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Ручка сдвижной двери внешняя", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Скоба замка капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Спойлер бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Спойлер бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Стойка кузова ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Стойка кузова ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Топливный бак", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Трос  капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Уголок двери внешний", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уголок двери внешний", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уголок двери внешний", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уголок двери внешний", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Уголок крыла ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уголок крыла ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уголок крыла ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уголок крыла ", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Уплотнитель стекла двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уплотнитель стекла двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уплотнитель стекла двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уплотнитель стекла двери", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Упор газовый  двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Упор газовый  двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Упор газовый капота", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Упор газовый капота", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Упор газовый крышки багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Упор газовый крышки багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Упор газовый стекла двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Упор газовый стекла двери багажника", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Усилитель бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Усилитель бампера", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Утеплитель капота", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Форсунка омывателя заднего стекла", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Форсунка омывателя лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Форсунка омывателя лобового стекла", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Форсунка омывателя фары", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Форсунка омывателя фары", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Хомут крепления топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Хомут крепления топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Чехол для запасного колеса", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Шарнир капота", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "R" },
       { name: "Шарнир капота", category: "Кузов снаружи", quantity: 0, price: 1000, left_right: "L" },
       { name: "Шланг горловины топливного бака", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Шланг омывателя фар", category: "Кузов снаружи", quantity: 0, price: 1000 },
       { name: "Эмблема", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Эмблема", category: "Кузов снаружи", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Амортизатор бардачка", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Бардачок", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Бардачок", category: "Кузов внутри", quantity: 0, price: 1000, top_bottom: "Верх" },
       { name: "Бардачок между сидений", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Бачок стеклоомывателя", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Воздуховод", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Воздуховод", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Воздуховод", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Воздуховод", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Воздухозаборник", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Гнездо наушников", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Горловина бачка стеклоомывателя", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Дефлектор воздуховода", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Дефлектор воздуховода", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Дефлектор воздуховода", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Дефлектор воздуховода", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Дефлектор воздуховода ", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Заглушка в руль", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Заглушка датчика света", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Замок бардачка", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Замок двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Замок двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Замок двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Замок двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Замок сдвижной двери", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Замок сдвижной двери", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Замок лючка бензобака", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Замок сиденья", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Замок сиденья", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Защита стоп сигнала", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Защита стоп сигнала", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Зеркало салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Карман ", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Карман обшивки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Карман обшивки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Карман обшивки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Карман обшивки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Клапан вентиляции", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Клапан вентиляции", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Коврик", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Коврик", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Коврик", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Коврик", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Консоль  приборов", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Консоль магнитофона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Консоль между сидений", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Консоль под рулевую колонку", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Корпус блока управления двигателем", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Крепеж сиденья", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крепеж сиденья", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крепление бачка стеклоомывателя", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Крепление салонного зеркала", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Крепление топливного насоса ", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Крепление магнитолы", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крепление магнитолы", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крепление магнитолы", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Кронштейн абсорбера", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Крышка крепления салонного зеркала", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крючок для одежды", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крючок для одежды", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крючок солнцезащитного козырька", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Крючок солнцезащитного козырька", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Молдинг стекла внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Молдинг стекла внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Молдинг стекла внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Молдинг стекла внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка консоли кпп", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Накладка на дверь внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на дверь внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на дверь внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на дверь внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка на порог внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на порог внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на порог внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка на порог внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на рамку двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Накладка на рамку двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Накладка на рамку двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Накладка на рамку двери внутренняя ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Верх" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Верх" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R", top_bottom: "Низ" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L", top_bottom: "Низ" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R", top_bottom: "Верх" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L", top_bottom: "Верх" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R", top_bottom: "Низ" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L", top_bottom: "Низ" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R", top_bottom: "Верх" },
       { name: "Накладка на стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L", top_bottom: "Верх" },
       { name: "Направляющая подголовника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Направляющая подголовника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Направляющая подголовника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Направляющая подголовника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Обрамление ручки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Обрамление ручки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Обрамление ручки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Обрамление ручки двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Обшивка багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R", top_bottom: "Низ" },
       { name: "Обшивка багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L", top_bottom: "Низ" },
       { name: "Обшивка багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R", top_bottom: "Верх" },
       { name: "Обшивка багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L", top_bottom: "Верх" },
       { name: "Обшивка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Обшивка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Обшивка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Обшивка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Обшивка двери багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Обшивка задней панели", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Обшивка крышки багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Обшивка потолка", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Блок педалей", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Отбойник двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Отбойник двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Отбойник двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Отбойник двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Отбойник двери багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Отбойник двери багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Отбойник крышки багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Отбойник крышки багажника", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Подставка под аккумулятор", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Педаль газа", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Педаль сцепления", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Пепельница", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Переключатель скоростей", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик багажного отсека", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик потолка", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Пластик салона", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Подголовник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подголовник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Подголовник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Подголовник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Подочечник", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Подстаканник", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Подушка безопастности в стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подушка безопастности в стойку", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Подушка безопастности в стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Подушка безопастности в стойку", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Покрытие напольное", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Полка стекла заднего", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Рамка магнитолы", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Регулятор высоты ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Регулятор высоты ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Резонатор воздушного фильтра", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ремень безопасности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ремень безопасности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ремень безопасности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ремень безопасности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Ремень безопасности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Руль", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ручка в салоне", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ручка в салоне", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ручка в салоне", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ручка в салоне", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ручка двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ручка двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ручка двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ручка двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ручка сдвижной двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Ручка сдвижной двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Ручка закрывания двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ручка закрывания двери внутренняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ручка закрывания двери внутреняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ручка закрывания двери внутреняя", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ручка открывания багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ручка открывания бака багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ручка открывания бензобака ", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ручка открывания капота", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Ручка стеклоподъемника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Ручка стеклоподъемника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Ручка стеклоподъемника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Ручка стеклоподъемника", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Сиденье", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Сиденье", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Сиденье", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Скоба замка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Скоба замка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Скоба замка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Скоба замка двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Скоба замка крышки багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "скоба замка двери багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Солнцезащитный козырек", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "R" },
       { name: "Солнцезащитный козырек", category: "Кузов внутри", quantity: 0, price: 1000, left_right: "L" },
       { name: "Стеклоподъемник двери багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Стеклоподьемник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Стеклоподьемник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Стеклоподьемник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Стеклоподьемник", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Торпедо", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трапеция дворников", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос газа", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос замка зажигания", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос капота", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос открывания двери  ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Трос открывания двери  ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Трос открывания двери  ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Трос открывания двери  ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Трос открывания крышки багажника", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос открывания лючка бензобака", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Трос сцепления", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Тросик спидометра ", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Тяга блокировки замка двери", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Уголок двери внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уголок двери внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уголок двери внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уголок двери внутренний", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Уплотнитель двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уплотнитель двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уплотнитель двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уплотнитель двери", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Уплотнитель дверного проема", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Уплотнитель дверного проема", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Уплотнитель дверного проема", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Уплотнитель дверного проема", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Усилитель торпедо", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Фальш педаль", category: "Кузов внутри", quantity: 0, price: 1000 },
       { name: "Фиксатор ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Фиксатор ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Фиксатор ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Фиксатор ремня безопастности", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Фиксатор ремня безопастности ", category: "Кузов внутри", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Стекло двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Стекло двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Стекло двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Стекло двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Стекло сдвижной двери", category: "Стекла", quantity: 0, price: 1000, left_right: "L" },
       { name: "Стекло сдвижной двери", category: "Стекла", quantity: 0, price: 1000, left_right: "R" },
       { name: "Стекло двери багажника ", category: "Стекла", quantity: 0, price: 1000 },
       { name: "Стекло заднее", category: "Стекла", quantity: 0, price: 1000 },
       { name: "Стекло кузова  ", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Стекло кузова  ", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Стекло кузова боковое", category: "Стекла", quantity: 0, price: 1000, left_right: "R" },
       { name: "Стекло кузова боковое", category: "Стекла", quantity: 0, price: 1000, left_right: "L" },
       { name: "Стекло лобовое", category: "Стекла", quantity: 0, price: 1000 },
       { name: "Форточка двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Форточка двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Форточка двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Форточка двери", category: "Стекла", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Форточка кузова", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Форточка кузова", category: "Стекла", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Вставка в крышку багажника", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Габарит", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Габарит", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Катафот в бампер", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Катафот в бампер", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Катафот в бампер", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Катафот в бампер", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Корпус фары", category: "Оптика", quantity: 0, price: 1000, left_right: "R" },
       { name: "Корпус фары", category: "Оптика", quantity: 0, price: 1000, left_right: "L" },
       { name: "Корректор фар", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Корректор фар", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Корректор фар", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Корректор фар", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Крышка фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крышка фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Лампочка", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Планка под фонарь", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Планка под фонарь", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Повторитель  бамперный", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Повторитель  бамперный", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Повторитель поворота  в крыло", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Повторитель поворота  в крыло", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подсветка заднего номера", category: "Оптика", quantity: 0, price: 1000 },
       { name: "Пыльник фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пыльник фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Стекло фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Стекло фары", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Стоп-сигнал в дверь багажника", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Фара", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Фара", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Фара противотуманная", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Фара противотуманная", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Фара противотуманная", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Фара противотуманная", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Фонарь в дверь багажника", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Фонарь в дверь багажника ", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Фонарь в крыло", category: "Оптика", quantity: 0, price: 1000, left_right: "R" },
       { name: "Фонарь в крыло", category: "Оптика", quantity: 0, price: 1000, left_right: "L" },
       { name: "Фонарь в крышку багажника", category: "Оптика", quantity: 0, price: 1000, left_right: "R" },
       { name: "Фонарь в крышку багажника", category: "Оптика", quantity: 0, price: 1000, left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Цоколь лампочки", category: "Оптика", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Глушитель", category: "Выхлопная система", quantity: 0, price: 1000 },
       { name: "Задняя часть глушителя", category: "Выхлопная система", quantity: 0, price: 1000 },
       { name: "Катализатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Катализатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Катализатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Катализатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Крепление глушителя", category: "Выхлопная система", quantity: 0, price: 1000, left_right: "R" },
       { name: "Крепление глушителя", category: "Выхлопная система", quantity: 0, price: 1000, left_right: "L" },
       { name: "Насос продувки катализатора", category: "Выхлопная система", quantity: 0, price: 1000 },
       { name: "Приемная труба глушителя", category: "Выхлопная система", quantity: 0, price: 1000, left_right: "L" },
       { name: "Приемная труба глушителя", category: "Выхлопная система", quantity: 0, price: 1000, left_right: "R" },
       { name: "Приемная труба глушителя", category: "Выхлопная система", quantity: 0, price: 1000 },
       { name: "Резонатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Резонатор", category: "Выхлопная система", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Тепловой экран выхлопной системы", category: "Выхлопная система", quantity: 0, price: 1000 },
       { name: "Бачок для тормозной жидкости", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Блок абс ", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Вакуумный усилитель тормозов", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Главный тормозной цилиндр", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Диск тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Диск тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Диск тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Диск тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Колодки стояночного тормоза", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Колодки стояночного тормоза", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Колодки тормозные", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Колодки тормозные", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Колодки тормозные", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Колодки тормозные", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Кронштейн блока абс", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Насос усилителя тормозов", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Педаль ручника", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Педаль тормоза", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Пыльник тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пыльник тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Пыльник тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Пыльник тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Регулятор давления тормозов", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Рычаг ручника", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Суппорт тормозной ", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Суппорт тормозной ", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Суппорт тормозной ", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Суппорт тормозной ", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тормозной барабан", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тормозной барабан", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Тормозной цилиндр", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Тормозной цилиндр", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Трос ручника", category: "Тормозная система", quantity: 0, price: 1000, left_right: "R" },
       { name: "Трос ручника", category: "Тормозная система", quantity: 0, price: 1000, left_right: "L" },
       { name: "Трос ручника", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Трубка тормозная", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Трубка тормозная ", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Трубка тормозная  ", category: "Тормозная система", quantity: 0, price: 1000 },
       { name: "Установочный комплект тормозных колодок", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Установочный комплект тормозных колодок", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Шланг тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Шланг тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Шланг тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Шланг тормозной", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Щиток тормозного диска", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Щиток тормозного диска", category: "Тормозная система", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       // Запчасти, относящиеся к системе кондиционирования автомобиля, включая компрессор, радиатор и трубопроводы
       { name: "Датчик давления кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Испаритель кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Клапан кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Компрессор кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Корпус испарителя", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Крепление компрессора кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Крепление радиатора кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Крепление радиатора кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000, left_right: "L" },
       { name: "Натяжитель  ремня кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Радиатор кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Ресивер кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Ролик натяжителя кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Трубка кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Трубка кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Трубка кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       { name: "Хомут трубки кондиционера", category: "Система кондиционирования", quantity: 0, price: 1000 },
       // Запчасти, относящиеся к электрооснащению автомобиля, включая аккумулятор, генератор, датчики и электронные блоки управления
       { name: "Аккумулятор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Активатор замка  крышки багажника", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Активатор замка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Активатор замка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Активатор замка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Активатор замка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Активатор замка двери  багажника", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Активатор замка лючка бензобака", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Антенна иммобилайзера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок кнопок", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок кнопок управления магнитолой", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок кнопок управления сиденьем", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Блок кнопок управления сиденьем", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Блок комфорта", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок подрулевых переключателей", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок предохранителей ", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок предохранителей ", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок розжига", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Блок розжига", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Блок сам", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления 4WD", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления abs", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления airbag", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления ESP", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления акселератором", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления вентилятором", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления дверьми", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления двигателем", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления замками", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления иммобилайзером", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления климат- контролем", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления кондиционером", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления подвеской", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления раздаткой", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления рулевой рейкой", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления ручным тормозом", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления свечами накала", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блок управления стеклоподъемниками", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Блок управления стеклоподъемниками", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Блок электронный", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Блокировка замка зажигания", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Бронепровод", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Вентилятор радиатора кондиционера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Включение круиз-контроля", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Генератор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Гнездо прикуривателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик airbag", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Датчик airbag", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Датчик airbag", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Датчик airbag", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Датчик аbs", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Датчик аbs", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Датчик аbs", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Датчик аbs", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Датчик абсолютного давления", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик вакуумного усилителя тормозов", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик включения вентилятора ", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик числа оборотов вала акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик гидроусилителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик давления масла акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик давления масла двигателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик давления пневмоподвески", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик детонации", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Датчик детонации", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Датчик детонации", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Датчик детонации", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Датчик детонации", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик дождя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик заднего хода", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик замка зажигания", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик износа тормозных колодок", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик массового расхода воздуха", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик неровной дороги", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Датчик парковки", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Датчик положения дроссельной заслонки", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик положения коленвала", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик положения распредвала", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Датчик положения распредвала", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Датчик положения распредвала", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Датчик положения распредвала", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Датчик положения распредвала", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик положения руля", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик положения селектора акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик регулировки дорожного просвета", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик света", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик скорости", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры впускного коллектора", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры выхлопных газов", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры кондиционера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры охлаждающей жидкости", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик температуры охлаждающей жидкости", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик уровня масла", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик уровня омывающей жидкости", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик уровня охлаждающей жидкости", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик уровня топлива", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Датчик ускорения", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Дисплей информационный", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Диффузор вентилятора", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Замок зажигания", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Зуммер", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Инвертор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Камера заднего вида", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Камера заднего вида", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Катушка зажигания", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кислородный датчик", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кислородный датчик", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Кислородный датчик", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Кислородный датчик", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Кислородный датчик", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Клапан vvt-i", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Клапан холостого хода", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка  антибуксировочной системы", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка аварийной сигнализации", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка блокировки дверей", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка включения 4WD", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка включения TV тюнера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка включения задней печки", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка включения кондиционера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка включения противотуманных фар", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка выключения массы", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка выключения подушки безопасности", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка запуска двигателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка круиз-контроля", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка люка", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка многофункциональная", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка обогрева заднего стекла", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка обогрева лобового стекла", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка омывателя заднего стекла", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка омывателя лобового стекла", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка омывателя фар", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка отключения airbag", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка отключения массы двигателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка открывания багажника", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка открывания бензобака", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка памяти сидений", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка подогрева сидений", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Кнопка подогрева сидений", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Кнопка помощи спуска со склона", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка регулировки рулевой колонкой", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка регулировки яркости подсветки приборной панели", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка режима стеклоочистителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Кнопка стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Кнопка стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Кнопка стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Кнопка управления акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления антенной", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления дополнительным омывателем ", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления дополнительным отопителем", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления зеркалами", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления корректором  фар", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления подвеской", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кнопка управления часами", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Коммутатор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Конденсатор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Контактная группа сдвижной двери", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Концевик двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Концевик двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Концевик двери", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Концевик двери", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Концевик крышки багажника", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Концевик под педаль тормоза", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Кренометр", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Крышка блока предохранителей", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Крышка трамблера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Лента накала свечей соединительная", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Магнитола", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Механизм зеркала ", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Механизм зеркала ", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Мотор вентилятора охлаждения", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор заслонки отопителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор заслонки отопителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор заслонки отопителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор заслонки отопителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор зеркала заднего вида", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Мотор зеркала заднего вида", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Мотор люка", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор печки", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор рулевой колонки ", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Мотор стеклоочистителя", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Мотор стеклоочистителя задний", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Мотор стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Мотор стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Мотор стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Мотор стеклоподъемника", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Моторчик регулировки сиденья ", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Моторчик регулировки сиденья ", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Насос омывателя фар", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Насос стеклоомывателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Насос стеклоомывателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Насос стеклоомывателя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Панель приборов", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Переключатель света фар", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Переключатель стеклоочистителей", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Плата заднего фонаря", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Плата заднего фонаря", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Подрулевой переключатель дворников", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Подрулевой переключатель света", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Подсветка багажника ", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Подсветка багажника ", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Подсветка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подсветка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Подсветка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Подсветка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Подсветка заднего номера", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Подушка безопасности боковая", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Подушка безопасности боковая", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Подушка безопасности в дверь", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Подушка безопасности в дверь", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Подушка безопасности в дверь", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Подушка безопасности в дверь", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Подушка безопасности водителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Подушка безопасности пассажира", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Предохранитель", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Прикуриватель", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Провод антенны", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Провод для датчика ABS", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Провод для датчика ABS", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Провод для датчика ABS", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Провод для датчика ABS", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Проводка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Проводка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Проводка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Проводка двери", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Проводка подкапотная", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Проводка салонная", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Резистор", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Резистор акпп", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Резистор отопителя", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Резистор топливного насоса", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Реле", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Светильник салона", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "F" },
       { name: "Светильник салона", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R" },
       { name: "Светильник салона", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Светильник салона", category: "Электрооснащение", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Свеча накала", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Cирена", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Сигнал", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Сигнализация", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Стартер", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Стоп-сигнал в дверь багажника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "R" },
       { name: "Стоп-сигнал в дверь багажника", category: "Электрооснащение", quantity: 0, price: 1000, left_right: "L" },
       { name: "Стоп-сигнал дополнительный", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Топливный модуль", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Топливный насос", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Трамблер", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Усилитель антенны", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Усилитель магнитолы", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Часы", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Шлейф рулевой", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Электромагнитный клапан", category: "Электрооснащение", quantity: 0, price: 1000 },
       { name: "Электроподогреватель", category: "Электрооснащение", quantity: 0, price: 1000 },
       // Запчасти, относящиеся к дискам и шинам автомобиля, включая колеса, шины и крепежные элементы
       { name: "Болт колесный", category: "Диски и шины", quantity: 0, price: 100 },
       { name: "Болт секретный", category: "Диски и шины", quantity: 0, price: 100 },
       { name: "Болт крепления запасного колеса", category: "Диски и шины", quantity: 0, price: 100 },
       { name: "Гайка на колесо", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Гайка секретная", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск литой", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск литой", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск литой", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск литой", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск штампованный", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск штампованный", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск штампованный", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Диск штампованный", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Ключ секретный", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колпачок на литой диск", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колпак колеса", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колпак колеса", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колпак колеса", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колпак колеса", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колесо запасное", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Колесо запасное", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Шина", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Шина", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Шина", category: "Диски и шины", quantity: 0, price: 1000 },
       { name: "Шина", category: "Диски и шины", quantity: 0, price: 1000 },
       // Запчасти, относящиеся к пневмосистеме автомобиля, включая компрессоры, баллоны и элементы подвески
       { name: "Клапан регулировки подвески", category: "Пневмосистема", quantity: 0, price: 1000 },
       { name: "Компрессор пневмоподвески", category: "Пневмосистема", quantity: 0, price: 1000 },
       { name: "Пневмобаллон", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Пневмобаллон", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Пневмобаллон", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Пневмобаллон", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Ресивер пневмоподвески", category: "Пневмосистема", quantity: 0, price: 1000 },
       { name: "Трубка подкачки пневмотической подвески", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Трубка подкачки пневмотической подвески", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Трубка подкачки пневмотической подвески", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Трубка подкачки пневмотической подвески", category: "Пневмосистема", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       // Сопутствующие товары для автомобиля, включая аксессуары, инструменты и дополнительные элементы
       { name: "Антенна", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Антенна", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Буксировочный крюк", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Динамик", category: "Сопутствующие товары", quantity: 0, price: 1000, front_rear: "F", left_right: "R" },
       { name: "Динамик", category: "Сопутствующие товары", quantity: 0, price: 1000, front_rear: "F", left_right: "L" },
       { name: "Динамик", category: "Сопутствующие товары", quantity: 0, price: 1000, front_rear: "R", left_right: "R" },
       { name: "Динамик", category: "Сопутствующие товары", quantity: 0, price: 1000, front_rear: "R", left_right: "L" },
       { name: "Домкрат", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Ключ балонный ", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Коврик багажника", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Крышка запасного колеса", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Полка  багажника", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Сабвуфер", category: "Сопутствующие товары", quantity: 0, price: 1000 },
       { name: "Шторка багажника", category: "Сопутствующие товары", quantity: 0, price: 1000 },
  ];

  // Добавляем цвет по умолчанию для каждой категории
  const commonPartsWithColor = commonParts.map(part => ({
    ...part,
    color: ['Электрооснащение', 'Система кондиционирования', 'Сопутствующие товары'].includes(part.category) ? 'Черный' : 'Белый'
  }));

  const onSubmit = async (data: DefectReportFormData) => {
    setLoading(true);

    try {
      // Создание дефектной ведомости
      const defectReportData = {
        brand: data.brand,
        model: data.model,
        year: data.year,
        vin: data.vin,
        mileage: data.mileage,
        engine_brand: data.engine_brand,
        body_brand: data.body_brand,
        interior_color: data.interior_color,
        body_color: data.body_color,
        description: data.description,
        selectedParts: commonPartsWithColor.map(part => {
          // Определяем цвет на основе категории
          const isInteriorCategory = ['Электрооснащение', 'Система кондиционирования', 'Сопутствующие товары'].includes(part.category);
          const partColor = isInteriorCategory ? data.interior_color : data.body_color || part.color;

          return {
            name: part.name,
            category: part.category,
            description: "",
            quantity: part.quantity,
            price: part.price,
            brand: data.brand,
            model: data.model,
            // Характеристики запчасти
            body_brand: data.body_brand,
            engine_brand: data.engine_brand,
            car_release_date: data.year.toString(), // Год выпуска автомобиля
            front_rear: part.front_rear,
            left_right: part.left_right,
            top_bottom: part.top_bottom,
            number: part.number,
            manufacturer: part.manufacturer,
            manufacturer_code: part.manufacturer_code,
            oem_code: part.oem_code,
            color: partColor,
            condition: part.condition,
            supplier_code: Date.now().toString(), // Автоматически выставляем код поставки как ID
            defect: part.defect,
            transmission: part.category === "Трансмиссия" ? data.transmission : part.transmission,
            transmission_model: transmissionModelCategories.includes(part.category)
              ? data.transmission_model
              : part.transmission_model,
            drive: part.drive,
            wear_percentage: part.wear_percentage,
            season: part.season,
            diameter: part.diameter,
            width: part.width,
            profile: part.profile,
            tire_quantity: part.tire_quantity,
            drilling: part.drilling,
            offset: part.offset,
            center_hole_diameter: part.center_hole_diameter,
            tire_model: part.tire_model,
            vin: data.vin, // VIN автомобиля
          };
        }),
      };

      const response = await fetch(`${API_BASE_URL}/api/defect-reports`, {
        method: "POST",
        headers: getAuthHeaders(),
        credentials: 'include',
        body: JSON.stringify(defectReportData),
      });

      if (!response.ok) {
        const text = await response.text().catch(() => null);
        alert(text || `Ошибка сервера: ${response.status}`);
        setLoading(false);
        return;
      }

      const result = await response.json().catch(() => ({ message: "OK" }));
      alert(result.message || "Дефектная ведомость успешно создана!");

      // Перейти к инвентарю
      window.location.href = '/inventory';
    } catch (err) {
      console.error("Ошибка сети:", err);
      alert("Ошибка сети при создании дефектной ведомости");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-5xl mx-auto p-4 sm:p-8">
      <Link
        to="/add-car"
        className="flex items-center gap-1 text-gray-600 hover:text-black mb-6 text-sm sm:text-base"
      >
        <ArrowLeft size={16} />
        Назад к выбору действия
      </Link>

      <h2 className="text-2xl sm:text-3xl font-semibold mb-4">Создание дефектной ведомости</h2>
      <p className="text-gray-500 mb-8 text-base sm:text-lg">
        Будет создана дефектная ведомость со всеми распространёнными запчастями для автомобиля.
      </p>

      <div className="bg-white rounded-lg border p-4 sm:p-8 shadow-lg w-full max-w-4xl">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Информация об автомобиле */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-8">
            <div className="space-y-4">
              <div>
                <Label htmlFor="brand-select">Бренд *</Label>
                <ClearableSelect
    value={brand || ""}
    onValueChange={(value) => setValue("brand", value)}
    placeholder="Выберите бренд" id="brand-select" className="h-10 w-full"
>
    <SelectItem value="BMW">BMW</SelectItem>
                    <SelectItem value="Audi">Audi</SelectItem>
                    <SelectItem value="Mercedes">Mercedes</SelectItem>
                    <SelectItem value="Toyota">Toyota</SelectItem>
                    <SelectItem value="Volkswagen">Volkswagen</SelectItem>
                    <SelectItem value="Honda">Honda</SelectItem>
                    <SelectItem value="Ford">Ford</SelectItem>
</ClearableSelect>
                <input
                  type="hidden"
                  {...register("brand")}
                  autoComplete="organization"
                />
                {errors.brand && <p className="text-red-500 text-sm">{errors.brand.message}</p>}
              </div>

              <div>
                <Label htmlFor="model">Модель *</Label>
                <Input
                  id="model"
                  {...register("model")}
                  type="text"
                  placeholder="E90"
                  className="h-10"
                  autoComplete="model"
                />
                {errors.model && <p className="text-red-500 text-sm">{errors.model.message}</p>}
              </div>

              <div>
                <Label htmlFor="body-brand">Марка кузова</Label>
                <Input
                  id="body-brand"
                  {...register("body_brand")}
                  type="text"
                  placeholder="Например: Toyota Corolla"
                  className="h-10"
                  autoComplete="off"
                />
              </div>

              <div>
                <Label htmlFor="engine-brand">Марка двигателя</Label>
                <Input
                  id="engine-brand"
                  {...register("engine_brand")}
                  type="text"
                  placeholder="Например: Toyota 1NZ-FE"
                  className="h-10"
                  autoComplete="off"
                />
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <Label htmlFor="vin">VIN</Label>
                <Input
                  id="vin"
                  {...register("vin")}
                  type="text"
                  placeholder="WVWZZZ1JZ3W386549"
                  className="h-10"
                  autoComplete="off"
                />
              </div>

              <div>
                <Label htmlFor="mileage">Пробег (км) *</Label>
                <Input
                  id="mileage"
                  {...register("mileage", { valueAsNumber: true })}
                  type="number"
                  className="h-10"
                  autoComplete="off"
                />
                {errors.mileage && <p className="text-red-500 text-sm">{errors.mileage.message}</p>}
              </div>
            </div>

            <div className="space-y-4">
              <div>
                <Label htmlFor="year">Год выпуска *</Label>
                <Input
                  id="year"
                  {...register("year", { valueAsNumber: true })}
                  type="number"
                  className="h-10"
                  autoComplete="off"
                />
                {errors.year && <p className="text-red-500 text-sm">{errors.year.message}</p>}
              </div>

              <div>
                <Label htmlFor="transmission-select">Тип трансмиссии</Label>
                <ClearableSelect
    value={watch("transmission") || ""}
    onValueChange={(value) => setValue("transmission", value)}
    placeholder="Выберите тип трансмиссии" id="transmission-select" className="h-10 w-full"
>
    <SelectItem value="МКПП">МКПП</SelectItem>
                    <SelectItem value="АКПП">АКПП</SelectItem>
                    <SelectItem value="Роботизированная">Роботизированная</SelectItem>
                    <SelectItem value="Вариатор">Вариатор</SelectItem>
</ClearableSelect>
                <input
                  type="hidden"
                  {...register("transmission")}
                  autoComplete="off"
                />
              </div>

              <div>
                <Label htmlFor="transmission-model">Модель трансмиссии</Label>
                <Input
                  id="transmission-model"
                  {...register("transmission_model")}
                  type="text"
                  placeholder="Введите номер трансмиссии"
                  className="h-10"
                  autoComplete="off"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Укажите номер трансмиссии. Применяется к подвеске ДВС/КПП,
                  трансмиссии и подвеске передних колес.
                </p>
              </div>

              <div>
                <Label htmlFor="body-color-select">Цвет кузовных деталей</Label>
                <SearchableSelect
                  value={watch("body_color") || ""}
                  onValueChange={(value) => setValue("body_color", value)}
                  options={availableColors.map((color) => ({ value: color, label: color }))}
                  placeholder="Выберите цвет кузовных деталей"
                  searchPlaceholder="Поиск цвета..."
                  className="h-10 w-full"
                />
                <input
                  type="hidden"
                  {...register("body_color")}
                  autoComplete="off"
                />
              </div>

              <div>
                <Label htmlFor="interior-color-select">Цвет салона</Label>
                <SearchableSelect
                  value={watch("interior_color") || ""}
                  onValueChange={(value) => setValue("interior_color", value)}
                  options={availableColors.map((color) => ({ value: color, label: color }))}
                  placeholder="Выберите цвет салона"
                  searchPlaceholder="Поиск цвета..."
                  className="h-10 w-full"
                />
                <input
                  type="hidden"
                  {...register("interior_color")}
                  autoComplete="off"
                />
              </div>
            </div>
          </div>

          {/* Заметки */}
          <div>
            <Label htmlFor="description">Заметки</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder="Дополнительная информация..."
              className="min-h-[100px] sm:min-h-[140px] mt-3"
              autoComplete="off"
            />
            {errors.description && <p className="text-red-500 text-sm">{errors.description.message}</p>}
          </div>

          {/* Информация о создаваемых запчастях */}
          <div>
            <Label className="text-lg font-medium mb-4 block">Создаваемые запчасти</Label>
            <div className="flex items-center gap-4 mb-4">
              <p className="text-gray-500">
                Будет создано {commonPartsWithColor.length} распространённых запчастей для автомобиля {brand} {watch("model")}:
              </p>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(10)}
                  className={displayLimit === 10 ? "bg-blue-50 border-blue-200" : ""}
                >
                  10
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(20)}
                  className={displayLimit === 20 ? "bg-blue-50 border-blue-200" : ""}
                >
                  20
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(30)}
                  className={displayLimit === 30 ? "bg-blue-50 border-blue-200" : ""}
                >
                  30
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(50)}
                  className={displayLimit === 50 ? "bg-blue-50 border-blue-200" : ""}
                >
                  50
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(100)}
                  className={displayLimit === 100 ? "bg-blue-50 border-blue-200" : ""}
                >
                  100
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setDisplayLimit(commonPartsWithColor.length)}
                  className={displayLimit === commonPartsWithColor.length ? "bg-blue-50 border-blue-200" : ""}
                >
                  Все
                </Button>
              </div>
            </div>
            <div className="max-h-96 overflow-y-auto border rounded-lg p-4 bg-gray-50">
              <div className="grid grid-cols-1 gap-2">
                {commonPartsWithColor.slice(0, displayLimit).map((part, index) => (
                  <div key={index} className="p-2 border-b border-gray-100 last:border-b-0">
                    <div className="flex-1">
                      <Label className="font-medium text-sm">
                        {part.name}
                      </Label>
                      <div className="mt-1 text-xs text-gray-600">
                        <div>Категория: {part.category}</div>
                        <div>Цвет: {part.color}</div>
                        {part.front_rear && <div>Перед/зад: {part.front_rear}</div>}
                        {part.left_right && <div>Лево/право: {part.left_right}</div>}
                        {part.top_bottom && <div>Верх/низ: {part.top_bottom}</div>}
                        {part.number && <div>Номер: {part.number}</div>}
                        {part.manufacturer && <div>Производитель: {part.manufacturer}</div>}
                        {part.manufacturer_code && <div>Код производителя: {part.manufacturer_code}</div>}
                        {part.oem_code && <div>OEM код: {part.oem_code}</div>}
                        {part.condition && <div>Состояние: {part.condition}</div>}
                        {part.defect && <div>Дефект: {part.defect}</div>}
                        {part.transmission && <div>Трансмиссия: {part.transmission}</div>}
                        {transmissionModelCategories.includes(part.category) && watch("transmission_model") && (
                          <div>Модель трансмиссии: {watch("transmission_model")}</div>
                        )}
                        {part.drive && <div>Привод: {part.drive}</div>}
                        {part.wear_percentage && <div>Процент износа: {part.wear_percentage}</div>}
                        {part.season && <div>Сезон: {part.season}</div>}
                        {part.diameter && <div>Диаметр: {part.diameter}</div>}
                        {part.width && <div>Ширина: {part.width}</div>}
                        {part.profile && <div>Профиль: {part.profile}</div>}
                        {part.tire_quantity && <div>Количество шин: {part.tire_quantity}</div>}
                        {part.drilling && <div>Сверловка: {part.drilling}</div>}
                        {part.offset && <div>Вылет: {part.offset}</div>}
                        {part.center_hole_diameter && <div>Диаметр ЦО: {part.center_hole_diameter}</div>}
                        {part.tire_model && <div>Модель шины: {part.tire_model}</div>}
                        <div className="text-green-600">Кол-во: {part.quantity}, Цена: {part.price}₽</div>
                      </div>
                    </div>
                  </div>
                ))}
                {commonPartsWithColor.length > displayLimit && (
                  <div className="text-center text-sm text-gray-500 mt-2">
                    ... и ещё {commonPartsWithColor.length - displayLimit} запчастей
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Кнопка отправки */}
          <div className="flex justify-start">
            <Button
              type="submit"
              disabled={loading}
              className="px-8 sm:px-12 py-3 sm:py-4 text-sm sm:text-base"
            >
              <Save className="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
              {loading ? "Создание..." : "Создать ведомость"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
