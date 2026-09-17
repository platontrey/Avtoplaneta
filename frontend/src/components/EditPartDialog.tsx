/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import FormRow from "./FormRow";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import React from "react";
import { motion } from "framer-motion";
import { Upload } from "lucide-react";
import type { UsePartEditReturn } from "@/hooks/usePartEdit";
import type { Part } from "@/features/parts/types";
import { API_BASE_URL } from "@/lib/api";
import SelectUserDropdown from "./SelectUserDropdown";
import type { User } from "@/features/messaging/types";
import { usePartCatalog } from "@/features/catalog/usePartCatalog";
import { useVehicleOptions } from "@/features/vehicles/useVehicleCatalog";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { formatCarReleasePeriod } from "@/lib/utils";

interface EditPartDialogProps {
    partEdit: UsePartEditReturn;
    part: Part;
    onPhotoChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
    onCrop: (src: string | File | undefined) => void;
    onDeletePhoto: () => void;
    originalFile: File | null;
}


export default function EditPartDialog({ partEdit, part, onPhotoChange, onCrop, onDeletePhoto, originalFile }: EditPartDialogProps) {
    console.log('EditPartDialog render, part.photo:', part.photo, 'photoPreview:', partEdit.photoUpload.photoPreview);
    const { data: partCatalog } = usePartCatalog();
    // Марки и модели — из общего серверного справочника, а не из списка в коде.
    const { brandOptions, modelOptions } = useVehicleOptions(partEdit.editForm.brand);

    const visibleFields = partCatalog?.attributes.map((attribute) => attribute.code) ?? [];
    const shows = (field: string) => visibleFields.length === 0 || visibleFields.includes(field);

    const transmissionOptions = partCatalog?.attributes.find(
        (attribute) => attribute.code === "transmission",
    )?.options ?? ["МКПП", "АКПП", "Роботизированная", "Вариатор"];
    const driveOptions = partCatalog?.attributes.find(
        (attribute) => attribute.code === "drive",
    )?.options ?? ["Передний", "Задний", "Полный"];

    return (
        <Dialog open={partEdit.isEditing} onOpenChange={(open) => { console.log('Edit dialog open state:', open); partEdit.setIsEditing(open); }}>
            <DialogContent
                className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto"
                onClick={(e) => e.stopPropagation()}
                onEscapeKeyDown={() => partEdit.setIsEditing(false)}
                onPointerDownOutside={() => partEdit.setIsEditing(false)}
            >
                <motion.div
                    initial={{ opacity: 0, scale: 0.9, y: 20 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.9, y: 20 }}
                    transition={{ duration: 0.15, ease: "easeOut" }}
                >
                    <DialogHeader>
                        <DialogTitle>Редактировать деталь</DialogTitle>
                        <DialogDescription>
                            Внесите изменения в детали детали здесь.
                        </DialogDescription>
                    </DialogHeader>
                    <Tabs defaultValue="basic" className="w-full">
                        <TabsList className="grid w-full grid-cols-2">
                            <TabsTrigger value="basic">Основная информация</TabsTrigger>
                            <TabsTrigger value="specifications">Характеристики</TabsTrigger>
                        </TabsList>

                        <TabsContent value="basic" className="space-y-4 mt-4">
                            <div className="grid gap-3 sm:gap-4 py-4">
                                <FormRow label="Название" htmlFor="name">
                                    <Input
                                        id="name"
                                        autoComplete="name"
                                        value={partEdit.editForm.name}
                                        onChange={(e) => partEdit.updateFormField('name', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Бренд" htmlFor="brand-select">
                                    <div className="relative">
                                        <SearchableSelect
                                            value={partEdit.editForm.brand || ""}
                                            onValueChange={(value) => {
                                                partEdit.updateFormField('brand', value);
                                                // Модель принадлежит марке — старое значение к новой марке не относится.
                                                partEdit.updateFormField('model', '');
                                            }}
                                            options={brandOptions}
                                            placeholder="Выберите или введите бренд"
                                            searchPlaceholder="Поиск бренда..."
                                            emptyMessage="Бренд не найден"
                                            allowCustom={true}
                                        />
                                        {partEdit.editForm.brand && (
                                            <button
                                                type="button"
                                                onClick={() => partEdit.updateFormField('brand', "")}
                                                className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                title="Очистить"
                                            >
                                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                </svg>
                                            </button>
                                        )}
                                    </div>
                                </FormRow>
                                <FormRow label="Модель" htmlFor="model">
                                    <SearchableSelect
                                        value={partEdit.editForm.model || ''}
                                        onValueChange={(value) => partEdit.updateFormField('model', value)}
                                        options={modelOptions}
                                        placeholder={partEdit.editForm.brand ? "Выберите или введите модель" : "Сначала выберите бренд"}
                                        searchPlaceholder="Поиск модели..."
                                        emptyMessage="Модель не найдена — можно ввести свою"
                                        allowCustom={true}
                                    />
                                </FormRow>
                                <FormRow label="VIN / Номер кузова" htmlFor="vin">
                                    <Input
                                        id="vin"
                                        autoComplete="off"
                                        value={partEdit.editForm.vin || ''}
                                        onChange={(e) => partEdit.updateFormField('vin', e.target.value)}
                                    />
                                </FormRow>
                                {shows('car_release_period') && (
                                    <FormRow label="Период выпуска автомобиля" htmlFor="car_release_period">
                                        <Input
                                            id="car_release_period"
                                            autoComplete="off"
                                            placeholder="Например: 2001-2007"
                                            value={partEdit.editForm.car_release_period || ''}
                                            onChange={(e) => partEdit.updateFormField('car_release_period', formatCarReleasePeriod(e.target.value))}
                                        />
                                    </FormRow>
                                )}
                                <FormRow label="Количество" htmlFor="quantity">
                                    <Input
                                        id="quantity"
                                        type="number"
                                        autoComplete="off"
                                        value={partEdit.editForm.quantity}
                                        onChange={(e) => partEdit.updateFormField('quantity', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Описание" htmlFor="description">
                                    <Textarea
                                        id="description"
                                        autoComplete="off"
                                        value={partEdit.editForm.description}
                                        onChange={(e) => partEdit.updateFormField('description', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Категория" htmlFor="category">
                                    <div>
                                        <Select
                                            value={partEdit.editForm.category || ""}
                                            onValueChange={(value) => {
                                                console.log('Category select changed:', value);
                                                partEdit.updateFormField('category', value);
                                            }}
                                            onOpenChange={(open) => console.log('Category select open state:', open)}
                                        >
                                            <SelectTrigger id="category" className="w-full">
                                                <SelectValue placeholder="Выберите категорию" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {partCatalog?.part_form_categories?.map((item) => (
                                                    <SelectItem key={item.code || item.name} value={item.name}>{item.name}</SelectItem>
                                                ))}
                                                {partEdit.editForm.category && !partCatalog?.part_form_categories?.some(i => i.name === partEdit.editForm.category) && (
                                                    <SelectItem value={partEdit.editForm.category}>{partEdit.editForm.category}</SelectItem>
                                                )}
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </FormRow>
                                <FormRow label="Цена (₽)" htmlFor="price">
                                    <Input
                                        id="price"
                                        type="number"
                                        step="100"
                                        autoComplete="off"
                                        value={partEdit.editForm.price}
                                        onChange={(e) => partEdit.updateFormField('price', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Местоположение" htmlFor="location">
                                    <Input
                                        id="location"
                                        autoComplete="address-line1"
                                        value={partEdit.editForm.location}
                                        onChange={(e) => partEdit.updateFormField('location', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Адрес склада" htmlFor="address">
                                    <Input
                                        id="address"
                                        autoComplete="street-address"
                                        value={partEdit.editForm.address}
                                        onChange={(e) => partEdit.updateFormField('address', e.target.value)}
                                    />
                                </FormRow>
                                <FormRow label="Статус" htmlFor="status">
                                    <div>
                                        <Select value={partEdit.editForm.status ? "active" : "inactive"} onValueChange={(value) => { console.log('Status select changed:', value); partEdit.updateFormField('status', value === "active"); }} onOpenChange={(open) => console.log('Status select open state:', open)}>
                                            <SelectTrigger id="status" className="w-full">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="active">Активный</SelectItem>
                                                <SelectItem value="inactive">Неактивный</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </FormRow>
                                <FormRow label="Продавец" htmlFor="salesman-select">
                                    <div>
                                        <SelectUserDropdown
                                            selectedUsers={
                                                partEdit.editForm.seller_id
                                                ? [{ id: partEdit.editForm.seller_id, name: partEdit.editForm.salesman, email: '' } as User]
                                                : (partEdit.editForm.salesman ? [{ id: 0, name: partEdit.editForm.salesman, email: '' } as User] : [])
                                            }
                                            onSelectionChange={(users) => {
                                                if (users.length > 0) {
                                                    partEdit.updateFormField('seller_id', users[0].id);
                                                    partEdit.updateFormField('salesman', users[0].name || users[0].email || 'Без имени');
                                                } else {
                                                    partEdit.updateFormField('seller_id', undefined);
                                                    partEdit.updateFormField('salesman', '');
                                                }
                                            }}
                                            placeholder="Выберите продавца"
                                            multiple={false}
                                            label=""
                                        />
                                    </div>
                                </FormRow>
                                <FormRow label="Фото" htmlFor="photo">
                                    <div className="sm:col-span-3 flex flex-col space-y-2">
                                        <input
                                            type="file"
                                            id="photo"
                                            name="photo"
                                            accept="image/*"
                                            onChange={onPhotoChange}
                                            className="hidden"
                                        />
                                        <label
                                            htmlFor="photo"
                                            className={`flex items-center justify-center w-full h-32 border-2 border-dashed rounded-lg transition-colors ${
                                                partEdit.photoUpload.isUploading
                                                    ? 'border-blue-300 bg-blue-50 cursor-not-allowed'
                                                    : 'border-gray-300 hover:border-gray-400 cursor-pointer'
                                            }`}
                                            style={{ pointerEvents: partEdit.photoUpload.isUploading ? 'none' : 'auto' }}
                                        >
                                            {partEdit.photoUpload.photoPreview ? (
                                                <img
                                                    src={partEdit.photoUpload.photoPreview}
                                                    alt={`Preview of ${partEdit.editForm.name}`}
                                                    className="max-h-16 max-w-full object-contain"
                                                    onError={(e) => {
                                                        console.error('Image failed to load:', partEdit.photoUpload.photoPreview);
                                                        e.currentTarget.style.display = 'none';
                                                    }}
                                                />
                                            ) : (
                                                <div className="text-center">
                                                    {partEdit.photoUpload.isUploading ? (
                                                        <>
                                                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500 mx-auto"></div>
                                                            <p className="mt-1 text-xs text-blue-600">Загрузка...</p>
                                                        </>
                                                    ) : (
                                                        <>
                                                            <Upload className="mx-auto h-6 w-6 text-gray-400" />
                                                            <p className="mt-1 text-xs text-gray-500">Нажмите для выбора фото</p>
                                                        </>
                                                    )}
                                                </div>
                                            )}
                                        </label>
                                        {partEdit.photoUpload.photoPreview && (
                                            <div className="flex space-x-2">
                                                <Button
                                                    type="button"
                                                    onClick={() => onCrop(originalFile || (part.photo ? `${API_BASE_URL}${part.photo}` : undefined))}
                                                    variant="outline"
                                                    size="sm"
                                                    disabled={!originalFile && !part.photo}
                                                >
                                                    Изменить фото
                                                </Button>
                                                <Button
                                                    type="button"
                                                    onClick={onDeletePhoto}
                                                    className="flex-1 bg-black text-white hover:bg-gray-800"
                                                    variant="outline"
                                                    disabled={partEdit.isLoading}
                                                >
                                                    Удалить фото
                                                </Button>
                                            </div>
                                        )}
                                    </div>
                                </FormRow>
                            </div>
                        </TabsContent>

                        <TabsContent value="specifications" className="space-y-4 mt-4">
                            <div className="grid gap-3 sm:gap-4 py-4">

                                <FormRow label="Марка кузова" htmlFor="body_brand">
                                    <Input id="body_brand" value={partEdit.editForm.body_brand || ''} onChange={(e) => partEdit.updateFormField('body_brand', e.target.value)} />
                                </FormRow>

                                {shows('engine_brand') && (
                                    <FormRow label="Марка двигателя" htmlFor="engine_brand">
                                        <Input
                                            id="engine_brand"
                                            value={partEdit.editForm.engine_brand || ''}
                                            onChange={(e) => partEdit.updateFormField('engine_brand', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('car_release_date') && (
                                    <FormRow label="Год выпуска" htmlFor="car_release_date">
                                        <Input
                                            id="car_release_date"
                                            value={partEdit.editForm.car_release_date || ''}
                                            onChange={(e) => partEdit.updateFormField('car_release_date', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('car_release_period') && (
                                    <FormRow label="Период выпуска автомобиля" htmlFor="car_release_period_spec">
                                        <Input
                                            id="car_release_period_spec"
                                            placeholder="Например: 2001-2007"
                                            value={partEdit.editForm.car_release_period || ''}
                                            onChange={(e) => partEdit.updateFormField('car_release_period', formatCarReleasePeriod(e.target.value))}
                                        />
                                    </FormRow>
                                )}

                                {shows('front_rear') && (
                                    <FormRow label="Перед/зад" htmlFor="front_rear">
                                        <div className="relative">
                                            <Select value={partEdit.editForm.front_rear || ""} onValueChange={(value) => partEdit.updateFormField('front_rear', value)}>
                                                <SelectTrigger className="w-full">
                                                    <SelectValue placeholder="Выберите" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="F">F (Перед)</SelectItem>
                                                    <SelectItem value="R">R (Зад)</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            {partEdit.editForm.front_rear && (
                                                <button
                                                    type="button"
                                                    onClick={() => partEdit.updateFormField('front_rear', "")}
                                                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                    title="Очистить"
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                    </svg>
                                                </button>
                                            )}
                                        </div>
                                    </FormRow>
                                )}

                                {shows('left_right') && (
                                    <FormRow label="Право/лево" htmlFor="left_right">
                                        <div className="relative">
                                            <Select value={partEdit.editForm.left_right || ""} onValueChange={(value) => partEdit.updateFormField('left_right', value)}>
                                                <SelectTrigger className="w-full">
                                                    <SelectValue placeholder="Выберите" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="L">L (Лево)</SelectItem>
                                                    <SelectItem value="R">R (Право)</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            {partEdit.editForm.left_right && (
                                                <button
                                                    type="button"
                                                    onClick={() => partEdit.updateFormField('left_right', "")}
                                                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                    title="Очистить"
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                    </svg>
                                                </button>
                                            )}
                                        </div>
                                    </FormRow>
                                )}

                                {shows('top_bottom') && (
                                    <FormRow label="Верх/низ" htmlFor="top_bottom">
                                        <div className="relative">
                                            <Select value={partEdit.editForm.top_bottom || ""} onValueChange={(value) => partEdit.updateFormField('top_bottom', value)}>
                                                <SelectTrigger className="w-full">
                                                    <SelectValue placeholder="Выберите" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="Верх">Верх</SelectItem>
                                                    <SelectItem value="Низ">Низ</SelectItem>
                                                    <SelectItem value="Середина">Середина</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            {partEdit.editForm.top_bottom && (
                                                <button
                                                    type="button"
                                                    onClick={() => partEdit.updateFormField('top_bottom', "")}
                                                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                    title="Очистить"
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                    </svg>
                                                </button>
                                            )}
                                        </div>
                                    </FormRow>
                                )}

                                {shows('number') && (
                                    <FormRow label="Номер детали" htmlFor="number">
                                        <Input
                                            id="number"
                                            value={partEdit.editForm.number || ''}
                                            onChange={(e) => partEdit.updateFormField('number', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('manufacturer') && (
                                    <FormRow label="Производитель" htmlFor="manufacturer">
                                        <Input
                                            id="manufacturer"
                                            value={partEdit.editForm.manufacturer || ''}
                                            onChange={(e) => partEdit.updateFormField('manufacturer', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('manufacturer_code') && (
                                    <FormRow label="Код производителя" htmlFor="manufacturer_code">
                                        <Input
                                            id="manufacturer_code"
                                            value={partEdit.editForm.manufacturer_code || ''}
                                            onChange={(e) => partEdit.updateFormField('manufacturer_code', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('oem_code') && (
                                    <FormRow label="OEM код" htmlFor="oem_code">
                                        <Input
                                            id="oem_code"
                                            value={partEdit.editForm.oem_code || ''}
                                            onChange={(e) => partEdit.updateFormField('oem_code', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('color') && (
                                    <FormRow label="Цвет" htmlFor="color">
                                        <Input
                                            id="color"
                                            value={partEdit.editForm.color || ''}
                                            onChange={(e) => partEdit.updateFormField('color', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('condition') && (
                                    <FormRow label="Состояние" htmlFor="condition">
                                        <Input
                                            id="condition"
                                            value={partEdit.editForm.condition || ''}
                                            onChange={(e) => partEdit.updateFormField('condition', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('supplier_code') && (
                                    <FormRow label="Код поставки" htmlFor="supplier_code">
                                        <Input
                                            id="supplier_code"
                                            value={partEdit.editForm.supplier_code || ''}
                                            onChange={(e) => partEdit.updateFormField('supplier_code', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('defect') && (
                                    <FormRow label="Дефект" htmlFor="defect">
                                        <Input
                                            id="defect"
                                            value={partEdit.editForm.defect || ''}
                                            onChange={(e) => partEdit.updateFormField('defect', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('transmission') && (
                                    <FormRow label="Тип трансмиссии" htmlFor="transmission">
                                        <div className="relative">
                                            <Select value={partEdit.editForm.transmission || ""} onValueChange={(value) => partEdit.updateFormField('transmission', value)}>
                                                <SelectTrigger className="w-full">
                                                    <SelectValue placeholder="Выберите тип трансмиссии" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {transmissionOptions.map((opt) => (
                                                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {partEdit.editForm.transmission && (
                                                <button
                                                    type="button"
                                                    onClick={() => partEdit.updateFormField('transmission', "")}
                                                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                    title="Очистить"
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                    </svg>
                                                </button>
                                            )}
                                        </div>
                                    </FormRow>
                                )}

                                {shows('transmission_model') && (
                                    <FormRow label="Модель трансмиссии" htmlFor="transmission_model">
                                        <Input
                                            id="transmission_model"
                                            value={partEdit.editForm.transmission_model || ''}
                                            onChange={(e) => partEdit.updateFormField('transmission_model', e.target.value)}
                                            placeholder="Введите номер трансмиссии"
                                        />
                                    </FormRow>
                                )}

                                {shows('drive') && (
                                    <FormRow label="Привод" htmlFor="drive">
                                        <div className="relative">
                                            <Select value={partEdit.editForm.drive || ""} onValueChange={(value) => partEdit.updateFormField('drive', value)}>
                                                <SelectTrigger className="w-full">
                                                    <SelectValue placeholder="Выберите привод" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {driveOptions.map((opt) => (
                                                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {partEdit.editForm.drive && (
                                                <button
                                                    type="button"
                                                    onClick={() => partEdit.updateFormField('drive', "")}
                                                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                                                    title="Очистить"
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                                    </svg>
                                                </button>
                                            )}
                                        </div>
                                    </FormRow>
                                )}

                                {shows('wear_percentage') && (
                                    <FormRow label="Процент износа (%)" htmlFor="wear_percentage">
                                        <Input
                                            id="wear_percentage"
                                            type="number"
                                            value={partEdit.editForm.wear_percentage || ''}
                                            onChange={(e) => partEdit.updateFormField('wear_percentage', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('diameter') && (
                                    <FormRow label="Диаметр" htmlFor="diameter">
                                        <Input
                                            id="diameter"
                                            value={partEdit.editForm.diameter || ''}
                                            onChange={(e) => partEdit.updateFormField('diameter', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('width') && (
                                    <FormRow label="Ширина" htmlFor="width">
                                        <Input
                                            id="width"
                                            value={partEdit.editForm.width || ''}
                                            onChange={(e) => partEdit.updateFormField('width', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('profile') && (
                                    <FormRow label="Профиль" htmlFor="profile">
                                        <Input
                                            id="profile"
                                            value={partEdit.editForm.profile || ''}
                                            onChange={(e) => partEdit.updateFormField('profile', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('tire_quantity') && (
                                    <FormRow label="Количество шин" htmlFor="tire_quantity">
                                        <Input
                                            id="tire_quantity"
                                            value={partEdit.editForm.tire_quantity || ''}
                                            onChange={(e) => partEdit.updateFormField('tire_quantity', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('drilling') && (
                                    <FormRow label="Сверловка" htmlFor="drilling">
                                        <Input
                                            id="drilling"
                                            value={partEdit.editForm.drilling || ''}
                                            onChange={(e) => partEdit.updateFormField('drilling', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('offset') && (
                                    <FormRow label="Вылет" htmlFor="offset">
                                        <Input
                                            id="offset"
                                            value={partEdit.editForm.offset || ''}
                                            onChange={(e) => partEdit.updateFormField('offset', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('center_hole_diameter') && (
                                    <FormRow label="Диаметр ЦО" htmlFor="center_hole_diameter">
                                        <Input
                                            id="center_hole_diameter"
                                            value={partEdit.editForm.center_hole_diameter || ''}
                                            onChange={(e) => partEdit.updateFormField('center_hole_diameter', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('tire_model') && (
                                    <FormRow label="Модель шины" htmlFor="tire_model">
                                        <Input
                                            id="tire_model"
                                            value={partEdit.editForm.tire_model || ''}
                                            onChange={(e) => partEdit.updateFormField('tire_model', e.target.value)}
                                        />
                                    </FormRow>
                                )}

                                {shows('season') && (
                                    <FormRow label="Сезон" htmlFor="season">
                                        <Input
                                            id="season"
                                            value={partEdit.editForm.season || ''}
                                            onChange={(e) => partEdit.updateFormField('season', e.target.value)}
                                        />
                                    </FormRow>
                                )}
                            </div>
                        </TabsContent>
                    </Tabs>
                    <DialogFooter>
                        <motion.div
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                        >
                            <Button type="submit" onClick={partEdit.handleEdit} disabled={partEdit.isLoading}>
                                {partEdit.isLoading ? 'Сохранение...' : 'Сохранить изменения'}
                            </Button>
                        </motion.div>
                    </DialogFooter>
                </motion.div>
            </DialogContent>
        </Dialog>
    );
}
