/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useMemo, useState } from "react";
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
import {
  flattenCatalogParts,
  expandDefectReportParts,
  reportBindingCategories,
  usePartCatalog,
} from '@/features/catalog/usePartCatalog';

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

const defectReportSchema = z.object({
  brand: z.string().min(1, "Выберите бренд"),
  model: z.string().min(1, "Введите модель"),
  year: z.number().min(1900, "Введите корректный год").max(new Date().getFullYear() + 1, "Год не может быть в будущем"),
  vin: z.string().optional(),
  mileage: z.number().min(0, "Пробег должен быть положительным числом"),
  transmission: z.string().optional(),
  transmission_model: z.string().optional(),
  drive: z.string().optional(),
  engine_brand: z.string().optional(),
  body_brand: z.string().optional(),
  interior_color: z.string().optional(),
  body_color: z.string().optional(),
  description: z.string().max(1000, "Описание слишком длинное (макс 1000 символов)").optional(),
});

type DefectReportFormData = z.infer<typeof defectReportSchema>;

export default function DefectReport() {
  const [loading, setLoading] = useState<boolean>(false);
  const [displayLimit, setDisplayLimit] = useState<number>(10);
  const {
    data: partCatalog,
    isLoading: catalogLoading,
    error: catalogError,
  } = usePartCatalog();

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

  // Каталог и зависимости загружаются из единого источника parts-service.
  const commonParts = useMemo(() => flattenCatalogParts(partCatalog), [partCatalog]);
  const transmissionModelCategories = useMemo(
    () => reportBindingCategories(partCatalog, 'transmission_model'),
    [partCatalog],
  );
  const driveCategories = useMemo(
    () => reportBindingCategories(partCatalog, 'drive'),
    [partCatalog],
  );
  const transmissionOptions =
    partCatalog?.attributes.find((attribute) => attribute.code === 'transmission')?.options ?? [];
  const driveOptions =
    partCatalog?.attributes.find((attribute) => attribute.code === 'drive')?.options ?? [];

  const commonPartsWithColor = commonParts;

  const onSubmit = async (data: DefectReportFormData) => {
    if (!partCatalog) {
      alert("Каталог запчастей ещё не загружен. Повторите попытку.");
      return;
    }
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
        transmission: data.transmission,
        transmission_model: data.transmission_model,
        drive: data.drive,
        description: data.description,
        catalog_version: partCatalog.version,
        selectedParts: expandDefectReportParts(partCatalog, {
          body_brand: data.body_brand,
          engine_brand: data.engine_brand,
          year: data.year,
          vin: data.vin,
          transmission: data.transmission,
          transmission_model: data.transmission_model,
          drive: data.drive,
          interior_color: data.interior_color,
          body_color: data.body_color,
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
                <Label htmlFor="vin">VIN / Марка кузова</Label>
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
    {transmissionOptions.map((option) => (
                      <SelectItem key={option} value={option}>{option}</SelectItem>
                    ))}
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
                <Label htmlFor="drive-select">Привод</Label>
                <ClearableSelect
                  value={watch("drive") || ""}
                  onValueChange={(value) => setValue("drive", value)}
                  placeholder="Выберите привод"
                  id="drive-select"
                  className="h-10 w-full"
                >
                  {driveOptions.map((option) => (
                    <SelectItem key={option} value={option}>{option}</SelectItem>
                  ))}
                </ClearableSelect>
                <input type="hidden" {...register("drive")} autoComplete="off" />
                <p className="mt-1 text-xs text-gray-500">
                  Значение будет добавлено только к подходящим запчастям подвески,
                  трансмиссии, рулевого управления, выхлопной и тормозной систем,
                  электрооснащения и двигателя.
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
            {catalogLoading && <p className="mb-3 text-sm text-gray-500">Загрузка каталога...</p>}
            {catalogError && (
              <p className="mb-3 text-sm text-red-600">{catalogError.message}</p>
            )}
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
                        {transmissionModelCategories.has(part.category) && watch("transmission_model") && (
                          <div>Модель трансмиссии: {watch("transmission_model")}</div>
                        )}
                        {driveCategories.has(part.category) && watch("drive") && (
                          <div>Привод: {watch("drive")}</div>
                        )}
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
              disabled={loading || catalogLoading || Boolean(catalogError)}
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
