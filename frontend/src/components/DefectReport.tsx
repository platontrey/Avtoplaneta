/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useEffect, useMemo, useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ClearableSelect } from "@/components/ClearableSelect";
import { SelectItem } from "@/components/ui/select";
import { Save, ArrowLeft, Search } from "lucide-react";
import { Link } from "react-router-dom";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { useVehicleOptions } from "@/features/vehicles/useVehicleCatalog";
import { formatCarReleasePeriod } from "@/lib/utils";
import { usePartCatalog } from "@/features/catalog/usePartCatalog";
import { createDefectReport, previewDefectReport, type DefectReportPayload, type DefectReportPreviewPart } from "@/features/defectReport/api";

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
  year: z
    .number()
    .min(1900, "Введите корректный год")
    .max(new Date().getFullYear() + 1, "Год не может быть в будущем"),
  vin: z.string().optional(),
  car_release_period: z.string().optional(),
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

const buildDefectReportPayload = (data: DefectReportFormData, catalogVersion?: string): DefectReportPayload => ({
  brand: data.brand || "",
  model: data.model || "",
  year: Number(data.year) || 0,
  car_release_period: data.car_release_period,
  vin: data.vin,
  mileage: Number(data.mileage) || 0,
  engine_brand: data.engine_brand,
  body_brand: data.body_brand,
  interior_color: data.interior_color,
  body_color: data.body_color,
  transmission: data.transmission,
  transmission_model: data.transmission_model,
  drive: data.drive,
  description: data.description,
  catalog_version: catalogVersion,
});

export default function DefectReport() {
  const [loading, setLoading] = useState<boolean>(false);
  const [displayLimit, setDisplayLimit] = useState<number>(10);
  const [previewSearch, setPreviewSearch] = useState<string>("");
  const [previewParts, setPreviewParts] = useState<DefectReportPreviewPart[]>([]);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const { data: partCatalog, isLoading: catalogLoading, error: catalogError } = usePartCatalog();

  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<DefectReportFormData>({
    resolver: zodResolver(defectReportSchema),
    defaultValues: {
      brand: "",
      model: "",
      year: new Date().getFullYear(),
      vin: "",
      car_release_period: "",
      mileage: 0,
      engine_brand: "",
      body_brand: "",
      transmission: "",
      transmission_model: "",
      drive: "",
      interior_color: "",
      body_color: "",
      description: "В связи с изменением цены конечную стоимость товара узнавать по WhatsApp 89138538227",
    },
  });

  const transmissionOptions = partCatalog?.attributes.find((attribute) => attribute.code === "transmission")?.options ?? [];
  const driveOptions = partCatalog?.attributes.find((attribute) => attribute.code === "drive")?.options ?? [];

  const {
    brand,
    model,
    year,
    car_release_period,
    vin,
    mileage,
    engine_brand,
    body_brand,
    interior_color,
    body_color,
    transmission,
    transmission_model,
    drive,
    description,
  } = watch();

  // Тот же серверный справочник, что и в мобильном приложении.
  const selectedBrand = brand;
  const { brandOptions, modelOptions } = useVehicleOptions(selectedBrand);

  const reportPayload = useMemo<DefectReportPayload>(
    () =>
      buildDefectReportPayload(
        {
          brand,
          model,
          year,
          car_release_period,
          vin,
          mileage,
          engine_brand,
          body_brand,
          interior_color,
          body_color,
          transmission,
          transmission_model,
          drive,
          description,
        },
        partCatalog?.version,
      ),
    [
      brand,
      model,
      year,
      car_release_period,
      vin,
      mileage,
      engine_brand,
      body_brand,
      interior_color,
      body_color,
      transmission,
      transmission_model,
      drive,
      description,
      partCatalog?.version,
    ],
  );

  useEffect(() => {
    if (!partCatalog) return;

    const abortController = new AbortController();
    const timeout = window.setTimeout(async () => {
      setPreviewLoading(true);
      setPreviewError(null);
      try {
        const preview = await previewDefectReport(reportPayload, abortController.signal);
        setPreviewParts(preview.parts);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setPreviewParts([]);
        setPreviewError(error instanceof Error ? error.message : "Не удалось построить превью");
      } finally {
        if (!abortController.signal.aborted) setPreviewLoading(false);
      }
    }, 350);

    return () => {
      window.clearTimeout(timeout);
      abortController.abort();
    };
  }, [partCatalog, reportPayload]);

  const filteredParts = useMemo(() => {
    if (!previewSearch.trim()) return previewParts;
    const query = previewSearch.toLowerCase().trim();
    return previewParts.filter((part) => part.name.toLowerCase().includes(query) || part.category.toLowerCase().includes(query));
  }, [previewParts, previewSearch]);

  const onSubmit = async (data: DefectReportFormData) => {
    if (!partCatalog) {
      alert("Каталог запчастей ещё не загружен. Повторите попытку.");
      return;
    }
    setLoading(true);

    try {
      const result = await createDefectReport(buildDefectReportPayload(data, partCatalog.version));
      alert(result.message || "Дефектная ведомость успешно создана!");

      // Перейти к инвентарю
      window.location.href = "/inventory";
    } catch (err) {
      console.error("Ошибка сети:", err);
      alert(err instanceof Error ? err.message : "Ошибка при создании дефектной ведомости");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-5xl mx-auto p-4 sm:p-8">
      <Link to="/add-car" className="flex items-center gap-1 text-gray-600 hover:text-black mb-6 text-sm sm:text-base">
        <ArrowLeft size={16} />
        Назад к выбору действия
      </Link>

      <h2 className="text-2xl sm:text-3xl font-semibold mb-4">Создание дефектной ведомости</h2>
      <p className="text-gray-500 mb-8 text-base sm:text-lg">Будет создана дефектная ведомость со всеми распространёнными запчастями для автомобиля.</p>

      <div className="bg-white rounded-lg border p-4 sm:p-8 shadow-lg w-full max-w-4xl">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Информация об автомобиле */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-8">
            {/* Левая колонка: Основные данные автомобиля */}
            <div className="space-y-4">
              <div>
                <Label htmlFor="brand-select">Бренд *</Label>
                <Controller
                  control={control}
                  name="brand"
                  render={({ field }) => (
                    <SearchableSelect
                      value={field.value || ""}
                      onValueChange={field.onChange}
                      options={brandOptions}
                      placeholder="Выберите или введите бренд"
                      searchPlaceholder="Поиск бренда или ввод нового..."
                      allowCustom={true}
                      className="h-10 w-full"
                    />
                  )}
                />
                {errors.brand && <p className="text-red-500 text-sm mt-1">{errors.brand.message}</p>}
              </div>

              <div>
                <Label htmlFor="model">Модель *</Label>
                <Controller
                  control={control}
                  name="model"
                  render={({ field }) => (
                    <SearchableSelect
                      value={field.value || ""}
                      onValueChange={field.onChange}
                      options={modelOptions}
                      placeholder={selectedBrand ? "Выберите или введите модель" : "Сначала выберите бренд"}
                      searchPlaceholder="Поиск модели или ввод новой..."
                      emptyMessage="Модель не найдена — можно ввести свою"
                      allowCustom={true}
                      className="h-10 w-full"
                    />
                  )}
                />
                {errors.model && <p className="text-red-500 text-sm mt-1">{errors.model.message}</p>}
              </div>

              <div>
                <Label htmlFor="year">Год выпуска *</Label>
                <Input id="year" {...register("year", { valueAsNumber: true })} type="number" className="h-10" autoComplete="off" />
                {errors.year && <p className="text-red-500 text-sm mt-1">{errors.year.message}</p>}
              </div>

              <div>
                <Label htmlFor="vin">VIN / Номер кузова</Label>
                <Input id="vin" {...register("vin")} type="text" placeholder="WVWZZZ1JZ3W386549" className="h-10" autoComplete="off" />
              </div>

              <div>
                <Label htmlFor="car_release_period">Период выпуска автомобиля</Label>
                <Controller
                  control={control}
                  name="car_release_period"
                  render={({ field }) => (
                    <Input
                      id="car_release_period"
                      value={field.value || ""}
                      onChange={(e) => field.onChange(formatCarReleasePeriod(e.target.value))}
                      type="text"
                      placeholder="Например: 2001-2007"
                      className="h-10"
                      autoComplete="off"
                    />
                  )}
                />
              </div>

              <div>
                <Label htmlFor="mileage">Пробег (км) *</Label>
                <Input id="mileage" {...register("mileage", { valueAsNumber: true })} type="number" className="h-10" autoComplete="off" />
                {errors.mileage && <p className="text-red-500 text-sm mt-1">{errors.mileage.message}</p>}
              </div>

              <div>
                <Label htmlFor="body-brand">Марка кузова</Label>
                <Input id="body-brand" {...register("body_brand")} type="text" placeholder="Например: E90" className="h-10" autoComplete="off" />
              </div>

              <div>
                <Label htmlFor="engine-brand">Марка двигателя</Label>
                <Input id="engine-brand" {...register("engine_brand")} type="text" placeholder="Например: Toyota 1NZ-FE" className="h-10" autoComplete="off" />
              </div>
            </div>

            {/* Правая колонка: Характеристики и цвета */}
            <div className="space-y-4">
              <div>
                <Label htmlFor="transmission-select">Тип трансмиссии</Label>
                <Controller
                  control={control}
                  name="transmission"
                  render={({ field }) => (
                    <ClearableSelect value={field.value || ""} onValueChange={field.onChange} placeholder="Выберите тип трансмиссии" id="transmission-select" className="h-10 w-full">
                      {transmissionOptions.map((option) => (
                        <SelectItem key={option} value={option}>
                          {option}
                        </SelectItem>
                      ))}
                    </ClearableSelect>
                  )}
                />
              </div>

              <div>
                <Label htmlFor="transmission-model">Модель трансмиссии</Label>
                <Controller
                  control={control}
                  name="transmission_model"
                  render={({ field }) => (
                    <Input id="transmission-model" value={field.value || ""} onChange={field.onChange} type="text" placeholder="Введите номер трансмиссии" className="h-10" autoComplete="off" />
                  )}
                />
                <p className="mt-1 text-xs text-gray-500">Укажите номер трансмиссии. Значение будет добавлено ко всем запчастям ведомости.</p>
              </div>

              <div>
                <Label htmlFor="drive-select">Привод</Label>
                <Controller
                  control={control}
                  name="drive"
                  render={({ field }) => (
                    <ClearableSelect value={field.value || ""} onValueChange={field.onChange} placeholder="Выберите привод" id="drive-select" className="h-10 w-full">
                      {driveOptions.map((option) => (
                        <SelectItem key={option} value={option}>
                          {option}
                        </SelectItem>
                      ))}
                    </ClearableSelect>
                  )}
                />
                <p className="mt-1 text-xs text-gray-500">Значение будет добавлено ко всем запчастям ведомости.</p>
              </div>

              <div>
                <Label htmlFor="body-color-select">Цвет кузовных деталей</Label>
                <Controller
                  control={control}
                  name="body_color"
                  render={({ field }) => (
                    <SearchableSelect
                      value={field.value || ""}
                      onValueChange={field.onChange}
                      options={availableColors.map((color) => ({
                        value: color,
                        label: color,
                      }))}
                      placeholder="Выберите цвет кузовных деталей"
                      searchPlaceholder="Поиск цвета..."
                      className="h-10 w-full"
                    />
                  )}
                />
              </div>

              <div>
                <Label htmlFor="interior-color-select">Цвет салона</Label>
                <Controller
                  control={control}
                  name="interior_color"
                  render={({ field }) => (
                    <SearchableSelect
                      value={field.value || ""}
                      onValueChange={field.onChange}
                      options={availableColors.map((color) => ({
                        value: color,
                        label: color,
                      }))}
                      placeholder="Выберите цвет салона"
                      searchPlaceholder="Поиск цвета..."
                      className="h-10 w-full"
                    />
                  )}
                />
              </div>
            </div>
          </div>

          {/* Заметки */}
          <div>
            <Label htmlFor="description">Заметки</Label>
            <Textarea id="description" {...register("description")} placeholder="Дополнительная информация..." className="min-h-[100px] sm:min-h-[140px] mt-3" autoComplete="off" />
            {errors.description && <p className="text-red-500 text-sm">{errors.description.message}</p>}
          </div>

          {/* Информация о создаваемых запчастях */}
          <div>
            <Label className="text-lg font-medium mb-4 block">Создаваемые запчасти</Label>
            {catalogLoading && <p className="mb-3 text-sm text-gray-500">Загрузка каталога...</p>}
            {catalogError && <p className="mb-3 text-sm text-red-600">{catalogError.message}</p>}
            {previewLoading && <p className="mb-3 text-sm text-gray-500">Сервер обновляет превью...</p>}
            {previewError && <p className="mb-3 text-sm text-red-600">{previewError}</p>}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
              <p className="text-gray-500 text-sm">
                Будет создано {previewParts.length} распространённых запчастей для автомобиля {brand || "—"} {model || ""}:
                {previewSearch.trim() && <span className="ml-1 text-blue-600 font-medium">(найдено: {filteredParts.length})</span>}
              </p>
              <div className="flex flex-wrap items-center gap-2">
                <div className="relative w-full sm:w-60">
                  <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
                  <Input type="text" placeholder="Поиск запчасти / категории..." value={previewSearch} onChange={(e) => setPreviewSearch(e.target.value)} className="pl-9 h-8 text-sm" />
                </div>
                <div className="flex items-center gap-1">
                  {[10, 20, 50, 100].map((limit) => (
                    <Button
                      key={limit}
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setDisplayLimit(limit)}
                      className={`h-8 px-2.5 text-xs ${displayLimit === limit ? "bg-blue-50 border-blue-300 text-blue-700 font-medium" : ""}`}
                    >
                      {limit}
                    </Button>
                  ))}
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setDisplayLimit(filteredParts.length || 1)}
                    className={`h-8 px-2.5 text-xs ${displayLimit >= filteredParts.length ? "bg-blue-50 border-blue-300 text-blue-700 font-medium" : ""}`}
                  >
                    Все
                  </Button>
                </div>
              </div>
            </div>
            <div className="max-h-96 overflow-y-auto border rounded-lg p-4 bg-gray-50">
              {filteredParts.length === 0 ? (
                <div className="text-center py-8 text-gray-400 text-sm">
                  {previewLoading
                    ? "Формируем превью..."
                    : previewError
                      ? "Превью временно недоступно"
                      : previewSearch.trim()
                        ? `Ничего не найдено по запросу "${previewSearch}"`
                        : "Нет доступных запчастей в каталоге"}
                </div>
              ) : (
                <div className="grid grid-cols-1 gap-2">
                  {filteredParts.slice(0, displayLimit).map((part, index) => (
                    <div key={index} className="p-2 border-b border-gray-100 last:border-b-0">
                      <div className="flex-1">
                        <Label className="font-medium text-sm">{part.name}</Label>
                        <div className="mt-1 text-xs text-gray-600">
                          <div>Категория: {part.category}</div>
                          {part.color && <div>Цвет: {part.color}</div>}
                          {part.car_release_date && <div>Год: {part.car_release_date}</div>}
                          {part.car_release_period && <div className="text-blue-700 font-medium">Период выпуска: {part.car_release_period}</div>}
                          {part.body_brand && <div>Марка кузова: {part.body_brand}</div>}
                          {part.engine_brand && <div>Марка двигателя: {part.engine_brand}</div>}
                          {part.vin && <div>VIN: {part.vin}</div>}
                          {part.transmission && <div className="text-blue-700 font-medium">Трансмиссия: {part.transmission}</div>}
                          {part.transmission_model && <div className="text-blue-700 font-medium">Модель трансмиссии: {part.transmission_model}</div>}
                          {part.drive && <div className="text-blue-700 font-medium">Привод: {part.drive}</div>}
                          {part.front_rear && <div>Перед/зад: {part.front_rear}</div>}
                          {part.left_right && <div>Лево/право: {part.left_right}</div>}
                          {part.top_bottom && <div>Верх/низ: {part.top_bottom}</div>}
                          {part.number && <div>Номер: {part.number}</div>}
                          {part.manufacturer && <div>Производитель: {part.manufacturer}</div>}
                          {part.manufacturer_code && <div>Код производителя: {part.manufacturer_code}</div>}
                          {part.oem_code && <div>OEM код: {part.oem_code}</div>}
                          {part.condition && <div>Состояние: {part.condition}</div>}
                          {part.defect && <div>Дефект: {part.defect}</div>}
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
                          <div className="text-green-600">
                            Кол-во: {part.quantity}, Цена: {part.price}₽
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                  {filteredParts.length > displayLimit && <div className="text-center text-sm text-gray-500 mt-2">... и ещё {filteredParts.length - displayLimit} запчастей</div>}
                </div>
              )}
            </div>
          </div>

          {/* Кнопка отправки */}
          <div className="flex justify-start">
            <Button type="submit" disabled={loading || catalogLoading || Boolean(catalogError) || Boolean(previewError)} className="px-8 sm:px-12 py-3 sm:py-4 text-sm sm:text-base">
              <Save className="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
              {loading ? "Создание..." : "Создать ведомость"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
