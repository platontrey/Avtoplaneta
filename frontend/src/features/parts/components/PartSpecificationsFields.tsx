/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { useFormContext, Controller } from "react-hook-form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SearchableSelect } from "@/components/ui/searchable-select";
import FormRow from "@/components/FormRow";
import { useVehicleOptions } from "@/features/vehicles/useVehicleCatalog";
import { usePartCatalog } from "@/features/catalog/usePartCatalog";
import type { PartFormData } from "../schemas/partSchema";

interface PartSpecificationsFieldsProps {
    brand?: string;
    model?: string;
}

export function PartSpecificationsFields({ brand: propBrand, model: propModel }: PartSpecificationsFieldsProps) {
    const { register, control, watch } = useFormContext<PartFormData>();

    const brand = propBrand ?? watch("brand") ?? "";
    const model = propModel ?? watch("model") ?? "";

    const { bodyOptions, engineOptions } = useVehicleOptions(brand, model);
    const { data: partCatalog } = usePartCatalog();

    const visibleFields = partCatalog?.attributes?.map((attr) => attr.code) ?? [];
    const shows = (code: string) => visibleFields.length === 0 || visibleFields.includes(code);

    const transmissionOptions = partCatalog?.attributes?.find(
        (attr) => attr.code === "transmission"
    )?.options ?? ["АКПП", "МКПП", "Вариатор", "Роботизированная"];

    const driveOptions = partCatalog?.attributes?.find(
        (attr) => attr.code === "drive"
    )?.options ?? ["Передний", "Задний", "Полный"];

    const conditionOptions = partCatalog?.attributes?.find(
        (attr) => attr.code === "condition"
    )?.options ?? ["Б/У", "Новое"];

    const seasonOptions = partCatalog?.attributes?.find(
        (attr) => attr.code === "season"
    )?.options ?? ["Зимние", "Летние", "Всесезонные"];

    return (
        <div className="grid gap-3 sm:gap-4 py-4">
            {/* Марка кузова */}
            <FormRow label="Марка кузова" htmlFor="body_brand">
                <Controller
                    name="body_brand"
                    control={control}
                    render={({ field }) => (
                        <SearchableSelect
                            value={field.value || ""}
                            onValueChange={field.onChange}
                            options={bodyOptions}
                            placeholder={brand ? "Выберите или введите марку кузова" : "Сначала выберите бренд"}
                            searchPlaceholder="Поиск кузова..."
                            emptyMessage="Кузов не найден — можно ввести свой"
                            allowCustom
                            className="h-10"
                        />
                    )}
                />
            </FormRow>

            {/* Марка двигателя */}
            {shows("engine_brand") && (
                <FormRow label="Марка двигателя" htmlFor="engine_brand">
                    <Controller
                        name="engine_brand"
                        control={control}
                        render={({ field }) => (
                            <SearchableSelect
                                value={field.value || ""}
                                onValueChange={field.onChange}
                                options={engineOptions}
                                placeholder={brand ? "Выберите или введите марку двигателя" : "Сначала выберите бренд"}
                                searchPlaceholder="Поиск двигателя..."
                                emptyMessage="Двигатель не найден — можно ввести свой"
                                allowCustom
                                className="h-10"
                            />
                        )}
                    />
                </FormRow>
            )}

            {/* Год выпуска */}
            {shows("car_release_date") && (
                <FormRow label="Год выпуска" htmlFor="car_release_date">
                    <Input id="car_release_date" {...register("car_release_date")} placeholder="Например: 2005" />
                </FormRow>
            )}

            {/* Перед / Зад */}
            {shows("front_rear") && (
                <FormRow label="Перед/зад" htmlFor="front_rear">
                    <Controller
                        name="front_rear"
                        control={control}
                        render={({ field }) => (
                            <Select value={field.value || ""} onValueChange={field.onChange}>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Выберите" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="F">F (Перед)</SelectItem>
                                    <SelectItem value="R">R (Зад)</SelectItem>
                                </SelectContent>
                            </Select>
                        )}
                    />
                </FormRow>
            )}

            {/* Лево / Право */}
            {shows("left_right") && (
                <FormRow label="Право/лево" htmlFor="left_right">
                    <Controller
                        name="left_right"
                        control={control}
                        render={({ field }) => (
                            <Select value={field.value || ""} onValueChange={field.onChange}>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Выберите" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="L">L (Лево)</SelectItem>
                                    <SelectItem value="R">R (Право)</SelectItem>
                                </SelectContent>
                            </Select>
                        )}
                    />
                </FormRow>
            )}

            {/* Верх / Низ */}
            {shows("top_bottom") && (
                <FormRow label="Верх/низ" htmlFor="top_bottom">
                    <Input id="top_bottom" {...register("top_bottom")} placeholder="Например: Верх" />
                </FormRow>
            )}

            {/* Цвет */}
            {shows("color") && (
                <FormRow label="Цвет деталей" htmlFor="color">
                    <Input id="color" {...register("color")} placeholder="Например: Черный, Серебристый" />
                </FormRow>
            )}

            {/* Номер детали */}
            {shows("number") && (
                <FormRow label="Номер детали" htmlFor="number">
                    <Input id="number" {...register("number")} placeholder="Номер детали" />
                </FormRow>
            )}

            {/* Производитель */}
            {shows("manufacturer") && (
                <FormRow label="Производитель" htmlFor="manufacturer">
                    <Input id="manufacturer" {...register("manufacturer")} placeholder="Производитель" />
                </FormRow>
            )}

            {/* Код производителя */}
            {shows("manufacturer_code") && (
                <FormRow label="Код производителя" htmlFor="manufacturer_code">
                    <Input id="manufacturer_code" {...register("manufacturer_code")} placeholder="Код производителя" />
                </FormRow>
            )}

            {/* OEM код */}
            {shows("oem_code") && (
                <FormRow label="OEM код" htmlFor="oem_code">
                    <Input id="oem_code" {...register("oem_code")} placeholder="OEM код" />
                </FormRow>
            )}

            {/* Код поставщика */}
            {shows("supplier_code") && (
                <FormRow label="Код поставщика" htmlFor="supplier_code">
                    <Input id="supplier_code" {...register("supplier_code")} placeholder="Код поставщика" />
                </FormRow>
            )}

            {/* Состояние */}
            {shows("condition") && (
                <FormRow label="Состояние" htmlFor="condition">
                    <Controller
                        name="condition"
                        control={control}
                        render={({ field }) => {
                            const val = (field.value || "").trim();
                            let normalized = val;
                            if (val.toLowerCase() === "б/у" || val.toLowerCase() === "бу") {
                                normalized = "Б/У";
                            } else if (val.toLowerCase() === "новое" || val.toLowerCase() === "новый") {
                                normalized = "Новое";
                            }
                            return (
                                <Select value={normalized} onValueChange={field.onChange}>
                                    <SelectTrigger className="w-full">
                                        <SelectValue placeholder="Выберите состояние" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {conditionOptions.map((opt) => (
                                            <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                        ))}
                                        {normalized && !conditionOptions.includes(normalized) && (
                                            <SelectItem value={normalized}>{normalized}</SelectItem>
                                        )}
                                    </SelectContent>
                                </Select>
                            );
                        }}
                    />
                </FormRow>
            )}

            {/* Дефект */}
            {shows("defect") && (
                <FormRow label="Дефект" htmlFor="defect">
                    <Input id="defect" {...register("defect")} placeholder="Описание дефекта (при наличии)" />
                </FormRow>
            )}

            {/* Трансмиссия */}
            {shows("transmission") && (
                <FormRow label="Трансмиссия" htmlFor="transmission">
                    <Controller
                        name="transmission"
                        control={control}
                        render={({ field }) => (
                            <Select value={field.value || ""} onValueChange={field.onChange}>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Выберите тип трансмиссии" />
                                </SelectTrigger>
                                <SelectContent>
                                    {transmissionOptions.map((opt) => (
                                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        )}
                    />
                </FormRow>
            )}

            {/* Модель трансмиссии */}
            {shows("transmission_model") && (
                <FormRow label="Модель трансмиссии" htmlFor="transmission_model">
                    <Input id="transmission_model" {...register("transmission_model")} placeholder="Модель трансмиссии" />
                </FormRow>
            )}

            {/* Привод */}
            {shows("drive") && (
                <FormRow label="Привод" htmlFor="drive">
                    <Controller
                        name="drive"
                        control={control}
                        render={({ field }) => (
                            <Select value={field.value || ""} onValueChange={field.onChange}>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="Выберите привод" />
                                </SelectTrigger>
                                <SelectContent>
                                    {driveOptions.map((opt) => (
                                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        )}
                    />
                </FormRow>
            )}

            {/* Процент износа */}
            {shows("wear_percentage") && (
                <FormRow label="Процент износа (%)" htmlFor="wear_percentage">
                    <Input id="wear_percentage" type="number" {...register("wear_percentage")} placeholder="0-100" />
                </FormRow>
            )}

            {/* Колеса и шины */}
            {shows("diameter") && (
                <FormRow label="Диаметр (R)" htmlFor="diameter">
                    <Input id="diameter" {...register("diameter")} placeholder="R16" />
                </FormRow>
            )}

            {shows("width") && (
                <FormRow label="Ширина" htmlFor="width">
                    <Input id="width" {...register("width")} placeholder="205" />
                </FormRow>
            )}

            {shows("profile") && (
                <FormRow label="Профиль" htmlFor="profile">
                    <Input id="profile" {...register("profile")} placeholder="55" />
                </FormRow>
            )}

            {shows("tire_quantity") && (
                <FormRow label="Количество шин" htmlFor="tire_quantity">
                    <Input id="tire_quantity" {...register("tire_quantity")} placeholder="4" />
                </FormRow>
            )}

            {shows("drilling") && (
                <FormRow label="Сверловка (PCD)" htmlFor="drilling">
                    <Input id="drilling" {...register("drilling")} placeholder="5x114.3" />
                </FormRow>
            )}

            {shows("offset") && (
                <FormRow label="Вылет (ET)" htmlFor="offset">
                    <Input id="offset" {...register("offset")} placeholder="ET45" />
                </FormRow>
            )}

            {shows("center_hole_diameter") && (
                <FormRow label="Диаметр ЦО (DIA)" htmlFor="center_hole_diameter">
                    <Input id="center_hole_diameter" {...register("center_hole_diameter")} placeholder="67.1" />
                </FormRow>
            )}

            {shows("tire_model") && (
                <FormRow label="Модель шины" htmlFor="tire_model">
                    <Input id="tire_model" {...register("tire_model")} placeholder="Модель шины" />
                </FormRow>
            )}

            {shows("season") && (
                <FormRow label="Сезон" htmlFor="season">
                    <Controller
                        name="season"
                        control={control}
                        render={({ field }) => {
                            const val = (field.value || "").trim();
                            const lower = val.toLowerCase();
                            let normalized = val;
                            if (lower.includes("всесез")) {
                                normalized = "Всесезонные";
                            } else if (lower.includes("зим")) {
                                normalized = "Зимние";
                            } else if (lower.includes("лет")) {
                                normalized = "Летние";
                            }
                            return (
                                <Select value={normalized} onValueChange={field.onChange}>
                                    <SelectTrigger className="w-full">
                                        <SelectValue placeholder="Выберите сезон" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {seasonOptions.map((opt) => (
                                            <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                        ))}
                                        {normalized && !seasonOptions.includes(normalized) && (
                                            <SelectItem value={normalized}>{normalized}</SelectItem>
                                        )}
                                    </SelectContent>
                                </Select>
                            );
                        }}
                    />
                </FormRow>
            )}
        </div>
    );
}
