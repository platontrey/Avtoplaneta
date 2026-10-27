/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { useFormContext, Controller } from "react-hook-form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SelectItem } from "@/components/ui/select";
import { ClearableSelect } from "@/components/ClearableSelect";
import { SearchableSelect } from "@/components/ui/searchable-select";
import FormRow from "@/components/FormRow";
import SelectUserDropdown from "@/components/SelectUserDropdown";
import { useVehicleOptions } from "@/features/vehicles/useVehicleCatalog";
import { usePartCatalog } from "@/features/catalog/usePartCatalog";
import { useShelves } from "@/hooks/useShelves";
import { formatCarReleasePeriod } from "@/lib/utils";
import type { PartFormData } from "../schemas/partSchema";
import type { User } from "@/features/messaging/types";

export function PartBasicFields() {
    const { register, control, watch, setValue, formState: { errors } } = useFormContext<PartFormData>();

    const brand = watch("brand") || "";
    const model = watch("model") || "";

    const { brandOptions, modelOptions } = useVehicleOptions(brand, model);
    const { data: partCatalog } = usePartCatalog();
    const { shelves } = useShelves();

    const visibleFields = partCatalog?.attributes?.map((attr) => attr.code) ?? [];
    const shows = (code: string) => visibleFields.length === 0 || visibleFields.includes(code);

    return (
        <div className="grid gap-3 sm:gap-4 py-4">
            {/* Название */}
            <FormRow label="Название *" htmlFor="name">
                <Input
                    id="name"
                    {...register("name")}
                    placeholder="Например: Диск тормозной передний"
                    autoComplete="off"
                />
                {errors.name && <p className="text-destructive text-xs mt-1">{errors.name.message}</p>}
            </FormRow>

            {/* Бренд */}
            <FormRow label="Бренд *" htmlFor="brand">
                <Controller
                    name="brand"
                    control={control}
                    render={({ field }) => (
                        <div className="relative">
                            <SearchableSelect
                                value={field.value || ""}
                                onValueChange={(val) => {
                                    field.onChange(val);
                                    setValue("model", "");
                                }}
                                options={brandOptions}
                                placeholder="Выберите или введите бренд"
                                searchPlaceholder="Поиск бренда..."
                                emptyMessage="Бренд не найден — можно ввести свой"
                                allowCustom
                                className="h-10"
                            />
                        </div>
                    )}
                />
                {errors.brand && <p className="text-destructive text-xs mt-1">{errors.brand.message}</p>}
            </FormRow>

            {/* Модель */}
            <FormRow label="Модель *" htmlFor="model">
                <Controller
                    name="model"
                    control={control}
                    render={({ field }) => (
                        <SearchableSelect
                            value={field.value || ""}
                            onValueChange={field.onChange}
                            options={modelOptions}
                            placeholder={brand ? "Выберите или введите модель" : "Сначала выберите бренд"}
                            searchPlaceholder="Поиск модели..."
                            emptyMessage="Модель не найдена — можно ввести свою"
                            allowCustom
                            className="h-10"
                        />
                    )}
                />
                {errors.model && <p className="text-destructive text-xs mt-1">{errors.model.message}</p>}
            </FormRow>

            {/* VIN */}
            <FormRow label="VIN / Номер кузова" htmlFor="vin">
                <Input
                    id="vin"
                    {...register("vin")}
                    placeholder="WVWZZZ..."
                    autoComplete="off"
                />
            </FormRow>

            {/* Период выпуска автомобиля */}
            {shows("car_release_period") && (
                <FormRow label="Период выпуска" htmlFor="car_release_period">
                    <Input
                        id="car_release_period"
                        placeholder="Например: 2001-2007"
                        autoComplete="off"
                        {...register("car_release_period", {
                            onChange: (e) => {
                                const formatted = formatCarReleasePeriod(e.target.value);
                                e.target.value = formatted;
                                setValue("car_release_period", formatted);
                            },
                        })}
                    />
                </FormRow>
            )}

            {/* Количество */}
            <FormRow label="Количество *" htmlFor="quantity">
                <Input
                    id="quantity"
                    type="number"
                    min={0}
                    {...register("quantity", { valueAsNumber: true })}
                />
                {errors.quantity && <p className="text-destructive text-xs mt-1">{errors.quantity.message}</p>}
            </FormRow>

            {/* Категория */}
            <FormRow label="Категория" htmlFor="category">
                <Controller
                    name="category"
                    control={control}
                    render={({ field }) => (
                        <ClearableSelect
                            id="category"
                            value={field.value || ""}
                            onValueChange={field.onChange}
                            placeholder="Выберите категорию"
                            className="h-10 w-full"
                        >
                            {partCatalog?.part_form_categories?.map((item) => (
                                <SelectItem key={item.code || item.name} value={item.name}>
                                    {item.name}
                                </SelectItem>
                            ))}
                        </ClearableSelect>
                    )}
                />
            </FormRow>

            {/* Цена */}
            <FormRow label="Цена (₽)" htmlFor="price">
                <Input
                    id="price"
                    type="number"
                    step="100"
                    min={0}
                    {...register("price", {
                        setValueAs: (v) => (v === "" || v === undefined ? undefined : Number(v)),
                    })}
                    placeholder="0"
                />
            </FormRow>

            {/* Описание */}
            <FormRow label="Описание" htmlFor="description">
                <Textarea
                    id="description"
                    {...register("description")}
                    placeholder="Дополнительные примечания к запчасти"
                    rows={3}
                />
            </FormRow>

            {/* Местоположение (полка / стеллаж) */}
            <FormRow label="Местоположение" htmlFor="location">
                <Controller
                    name="location"
                    control={control}
                    render={({ field }) => {
                        const currentVal = (field.value || "").trim();
                        const shelfOptions = shelves.map((s) => ({ value: s.name, label: s.name }));
                        if (currentVal && !shelfOptions.some((o) => o.value.toLowerCase() === currentVal.toLowerCase())) {
                            shelfOptions.unshift({ value: currentVal, label: currentVal });
                        }
                        return (
                            <SearchableSelect
                                value={field.value || ""}
                                onValueChange={field.onChange}
                                options={shelfOptions}
                                placeholder="Выберите или введите полку..."
                                searchPlaceholder="Поиск или ввод полки..."
                                emptyMessage="Полка не найдена"
                                allowCustom={true}
                                className="h-10 w-full"
                            />
                        );
                    }}
                />
            </FormRow>

            {/* Адрес склада */}
            <FormRow label="Адрес склада" htmlFor="address">
                <Controller
                    name="address"
                    control={control}
                    render={({ field }) => {
                        const warehouseAddresses = [
                            "Томск, Профсоюзная 2/11",
                            "г. Томск, пос. Кузовлево, ул. Заречная, 36",
                        ];
                        const currentVal = (field.value || "").trim();
                        const options =
                            currentVal && !warehouseAddresses.includes(currentVal)
                                ? [currentVal, ...warehouseAddresses]
                                : warehouseAddresses;
                        return (
                            <ClearableSelect
                                id="address"
                                value={field.value || ""}
                                onValueChange={field.onChange}
                                placeholder="Выберите адрес склада"
                                className="h-10 w-full"
                            >
                                {options.map((addr) => (
                                    <SelectItem key={addr} value={addr}>
                                        {addr}
                                    </SelectItem>
                                ))}
                            </ClearableSelect>
                        );
                    }}
                />
            </FormRow>

            {/* Продавец */}
            <FormRow label="Продавец" htmlFor="seller_id">
                <Controller
                    name="seller_id"
                    control={control}
                    render={({ field }) => (
                        <SelectUserDropdown
                            selectedUsers={
                                field.value
                                    ? [{ id: Number(field.value), name: '', email: '' } as User]
                                    : []
                            }
                            onSelectionChange={(selected) => {
                                if (selected.length > 0) {
                                    field.onChange(String(selected[0].id));
                                } else {
                                    field.onChange(undefined);
                                }
                            }}
                            multiple={false}
                            placeholder="Выберите продавца"
                        />
                    )}
                />
            </FormRow>
        </div>
    );
}
