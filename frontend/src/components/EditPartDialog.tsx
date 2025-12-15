/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import React from "react";
import { motion } from "framer-motion";
import { Upload } from "lucide-react";
import type { UsePartEditReturn } from "@/hooks/usePartEdit";
import type { Part } from "@/features/parts/types";
import { API_BASE_URL } from "@/lib/api";

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
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="name" className="sm:text-right text-sm">
                                        Название
                                    </Label>
                                    <Input
                                        id="name"
                                        autoComplete="name"
                                        value={partEdit.editForm.name}
                                        onChange={(e) => partEdit.updateFormField('name', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="brand-select" className="sm:text-right text-sm">
                                        Бренд
                                    </Label>
                                    <div className="relative sm:col-span-3">
                                        <Select value={partEdit.editForm.brand || ""} onValueChange={(value) => partEdit.updateFormField('brand', value)}>
                                            <SelectTrigger id="brand-select" className="h-10 w-full">
                                                <SelectValue placeholder="Выберите бренд" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="BMW">BMW</SelectItem>
                                                <SelectItem value="Audi">Audi</SelectItem>
                                                <SelectItem value="Mercedes">Mercedes</SelectItem>
                                                <SelectItem value="Toyota">Toyota</SelectItem>
                                                <SelectItem value="Volkswagen">Volkswagen</SelectItem>
                                                <SelectItem value="Honda">Honda</SelectItem>
                                                <SelectItem value="Ford">Ford</SelectItem>
                                            </SelectContent>
                                        </Select>
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
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="model" className="sm:text-right text-sm">
                                        Модель
                                    </Label>
                                    <Input
                                        id="model"
                                        autoComplete="off"
                                        value={partEdit.editForm.model || ''}
                                        onChange={(e) => partEdit.updateFormField('model', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="vin" className="sm:text-right text-sm">
                                        VIN
                                    </Label>
                                    <Input
                                        id="vin"
                                        autoComplete="off"
                                        value={partEdit.editForm.vin || ''}
                                        onChange={(e) => partEdit.updateFormField('vin', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="quantity" className="sm:text-right text-sm">
                                        Количество
                                    </Label>
                                    <Input
                                        id="quantity"
                                        type="number"
                                        autoComplete="off"
                                        value={partEdit.editForm.quantity}
                                        onChange={(e) => partEdit.updateFormField('quantity', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="description" className="sm:text-right text-sm">
                                        Описание
                                    </Label>
                                    <Textarea
                                        id="description"
                                        autoComplete="off"
                                        value={partEdit.editForm.description}
                                        onChange={(e) => partEdit.updateFormField('description', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="category" className="sm:text-right text-sm">
                                        Категория
                                    </Label>
                                    <Select value={partEdit.editForm.category || ""} onValueChange={(value) => { console.log('Category select changed:', value); partEdit.updateFormField('category', value); }} onOpenChange={(open) => console.log('Category select open state:', open)}>
                                        <SelectTrigger id="category" className="sm:col-span-3">
                                            <SelectValue placeholder="Выберите категорию" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="Тормоза">Тормоза</SelectItem>
                                            <SelectItem value="Двигатель">Двигатель</SelectItem>
                                            <SelectItem value="Подвеска">Подвеска</SelectItem>
                                            <SelectItem value="Подвеска ДВС/КПП">Подвеска ДВС/КПП</SelectItem>
                                            <SelectItem value="Подвеска передних колес">Подвеска передних колес</SelectItem>
                                            <SelectItem value="Подвеска задних колес">Подвеска задних колес</SelectItem>
                                            <SelectItem value="Электрика">Электрика</SelectItem>
                                            <SelectItem value="Кузов">Кузов</SelectItem>
                                            <SelectItem value="Кузов снаружи">Кузов снаружи</SelectItem>
                                            <SelectItem value="Интерьер">Интерьер</SelectItem>
                                            <SelectItem value="Трансмиссия">Трансмиссия</SelectItem>
                                            <SelectItem value="Система охлаждения и отопления">Система охлаждения и отопления</SelectItem>
                                            <SelectItem value="Система выхлопа (Глушитель)">Система выхлопа (Глушитель)</SelectItem>
                                            <SelectItem value="Система рулевого управления">Система рулевого управления</SelectItem>
                                            <SelectItem value="Рулевое управление">Рулевое управление</SelectItem>
                                            <SelectItem value="Система фильтрации (Фильтры)">Система фильтрации (Фильтры)</SelectItem>
                                            <SelectItem value="Шины и диски">Шины и диски</SelectItem>
                                            <SelectItem value="Автохимия и масла">Автохимия и масла</SelectItem>
                                            <SelectItem value="Аксессуары и тюннинг">Аксессуары и тюннинг</SelectItem>
                                            <SelectItem value="Другое">Другое</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="price" className="sm:text-right text-sm">
                                        Цена (₽)
                                    </Label>
                                    <Input
                                        id="price"
                                        type="number"
                                        step="100"
                                        autoComplete="off"
                                        value={partEdit.editForm.price}
                                        onChange={(e) => partEdit.updateFormField('price', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="location" className="sm:text-right text-sm">
                                        Местоположение
                                    </Label>
                                    <Input
                                        id="location"
                                        autoComplete="address-line1"
                                        value={partEdit.editForm.location}
                                        onChange={(e) => partEdit.updateFormField('location', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="status" className="sm:text-right text-sm">
                                        Статус
                                    </Label>
                                    <Select value={partEdit.editForm.status ? "active" : "inactive"} onValueChange={(value) => { console.log('Status select changed:', value); partEdit.updateFormField('status', value === "active"); }} onOpenChange={(open) => console.log('Status select open state:', open)}>
                                        <SelectTrigger id="status" className="sm:col-span-3">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="active">Активный</SelectItem>
                                            <SelectItem value="inactive">Неактивный</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="salesman-select" className="sm:text-right text-sm">
                                        Продавец
                                    </Label>
                                    <Input
                                        id="salesman"
                                        autoComplete="off"
                                        value={partEdit.editForm.salesman}
                                        onChange={(e) => partEdit.updateFormField('salesman', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="manufacturer" className="sm:text-right text-sm">
                                        Производитель
                                    </Label>
                                    <Input
                                        id="manufacturer"
                                        value={partEdit.editForm.manufacturer || ''}
                                        onChange={(e) => partEdit.updateFormField('manufacturer', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="manufacturer_code" className="sm:text-right text-sm">
                                        Код производителя
                                    </Label>
                                    <Input
                                        id="manufacturer_code"
                                        value={partEdit.editForm.manufacturer_code || ''}
                                        onChange={(e) => partEdit.updateFormField('manufacturer_code', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="oem_code" className="sm:text-right text-sm">
                                        OEM код
                                    </Label>
                                    <Input
                                        id="oem_code"
                                        value={partEdit.editForm.oem_code || ''}
                                        onChange={(e) => partEdit.updateFormField('oem_code', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="supplier_code" className="sm:text-right text-sm">
                                        Код поставки
                                    </Label>
                                    <Input
                                        id="supplier_code"
                                        value={partEdit.editForm.supplier_code || ''}
                                        onChange={(e) => partEdit.updateFormField('supplier_code', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="condition" className="sm:text-right text-sm">
                                        Состояние
                                    </Label>
                                    <Input
                                        id="condition"
                                        value={partEdit.editForm.condition || ''}
                                        onChange={(e) => partEdit.updateFormField('condition', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="wear_percentage" className="sm:text-right text-sm">
                                        Процент износа (%)
                                    </Label>
                                    <Input
                                        id="wear_percentage"
                                        value={partEdit.editForm.wear_percentage || ''}
                                        onChange={(e) => partEdit.updateFormField('wear_percentage', e.target.value)}
                                        className="sm:col-span-3"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                    <Label htmlFor="photo" className="sm:text-right text-sm">
                                        Фото
                                    </Label>
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
                                </div>
                            </div>
                        </TabsContent>

                        <TabsContent value="specifications" className="space-y-4 mt-4">
                            <div className="grid gap-3 sm:gap-4 py-4">
                                {/* Category-specific fields */}
                                {(partEdit.editForm.category === 'Двигатель' || partEdit.editForm.category === 'Трансмиссия') && (
                                    <>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="engine_brand" className="sm:text-right text-sm">
                                                Марка двигателя
                                            </Label>
                                            <Input
                                                id="engine_brand"
                                                value={partEdit.editForm.engine_brand || ''}
                                                onChange={(e) => partEdit.updateFormField('engine_brand', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="transmission" className="sm:text-right text-sm">
                                                Трансмиссия
                                            </Label>
                                            <div className="relative sm:col-span-3">
                                                <Select value={partEdit.editForm.transmission || ""} onValueChange={(value) => partEdit.updateFormField('transmission', value)}>
                                                    <SelectTrigger className="h-10 w-full">
                                                        <SelectValue placeholder="Выберите тип трансмиссии" />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="МКПП">МКПП</SelectItem>
                                                        <SelectItem value="АКПП">АКПП</SelectItem>
                                                        <SelectItem value="Роботизированная">Роботизированная</SelectItem>
                                                        <SelectItem value="Вариатор">Вариатор</SelectItem>
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
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="drive" className="sm:text-right text-sm">
                                                Привод
                                            </Label>
                                            <Input
                                                id="drive"
                                                value={partEdit.editForm.drive || ''}
                                                onChange={(e) => partEdit.updateFormField('drive', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                    </>
                                )}

                                {(partEdit.editForm.category === 'Кузов' || partEdit.editForm.category === 'Кузов снаружи' || partEdit.editForm.category === 'Интерьер') && (
                                    <>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="body_brand" className="sm:text-right text-sm">
                                                Марка кузова
                                            </Label>
                                            <Input
                                                id="body_brand"
                                                value={partEdit.editForm.body_brand || ''}
                                                onChange={(e) => partEdit.updateFormField('body_brand', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="color" className="sm:text-right text-sm">
                                                Цвет
                                            </Label>
                                            <Input
                                                id="color"
                                                value={partEdit.editForm.color || ''}
                                                onChange={(e) => partEdit.updateFormField('color', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                    </>
                                )}

                                {partEdit.editForm.category === 'Шины и диски' && (
                                    <>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="diameter" className="sm:text-right text-sm">
                                                Диаметр
                                            </Label>
                                            <Input
                                                id="diameter"
                                                value={partEdit.editForm.diameter || ''}
                                                onChange={(e) => partEdit.updateFormField('diameter', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="width" className="sm:text-right text-sm">
                                                Ширина
                                            </Label>
                                            <Input
                                                id="width"
                                                value={partEdit.editForm.width || ''}
                                                onChange={(e) => partEdit.updateFormField('width', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="profile" className="sm:text-right text-sm">
                                                Профиль
                                            </Label>
                                            <Input
                                                id="profile"
                                                value={partEdit.editForm.profile || ''}
                                                onChange={(e) => partEdit.updateFormField('profile', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="tire_quantity" className="sm:text-right text-sm">
                                                Количество
                                            </Label>
                                            <Input
                                                id="tire_quantity"
                                                value={partEdit.editForm.tire_quantity || ''}
                                                onChange={(e) => partEdit.updateFormField('tire_quantity', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="drilling" className="sm:text-right text-sm">
                                                Сверловка
                                            </Label>
                                            <Input
                                                id="drilling"
                                                value={partEdit.editForm.drilling || ''}
                                                onChange={(e) => partEdit.updateFormField('drilling', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="offset" className="sm:text-right text-sm">
                                                Вылет
                                            </Label>
                                            <Input
                                                id="offset"
                                                value={partEdit.editForm.offset || ''}
                                                onChange={(e) => partEdit.updateFormField('offset', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="center_hole_diameter" className="sm:text-right text-sm">
                                                Диаметр ЦО
                                            </Label>
                                            <Input
                                                id="center_hole_diameter"
                                                value={partEdit.editForm.center_hole_diameter || ''}
                                                onChange={(e) => partEdit.updateFormField('center_hole_diameter', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="tire_model" className="sm:text-right text-sm">
                                                Модель шины
                                            </Label>
                                            <Input
                                                id="tire_model"
                                                value={partEdit.editForm.tire_model || ''}
                                                onChange={(e) => partEdit.updateFormField('tire_model', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="season" className="sm:text-right text-sm">
                                                Сезон
                                            </Label>
                                            <Input
                                                id="season"
                                                value={partEdit.editForm.season || ''}
                                                onChange={(e) => partEdit.updateFormField('season', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                    </>
                                )}

                                {/* Position fields for most categories */}
                                {(partEdit.editForm.category !== 'Автохимия и масла' && partEdit.editForm.category !== 'Аксессуары и тюннинг' && partEdit.editForm.category !== 'Другое') && (
                                    <>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="front_rear" className="sm:text-right text-sm">
                                                Перед/зад
                                            </Label>
                                            <div className="relative sm:col-span-3">
                                                <Select value={partEdit.editForm.front_rear || ""} onValueChange={(value) => partEdit.updateFormField('front_rear', value)}>
                                                    <SelectTrigger className="h-10 w-full">
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
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="left_right" className="sm:text-right text-sm">
                                                Право/лево
                                            </Label>
                                            <div className="relative sm:col-span-3">
                                                <Select value={partEdit.editForm.left_right || ""} onValueChange={(value) => partEdit.updateFormField('left_right', value)}>
                                                    <SelectTrigger className="h-10 w-full">
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
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="top_bottom" className="sm:text-right text-sm">
                                                Верх/низ
                                            </Label>
                                            <div className="relative sm:col-span-3">
                                                <Select value={partEdit.editForm.top_bottom || ""} onValueChange={(value) => partEdit.updateFormField('top_bottom', value)}>
                                                    <SelectTrigger className="h-10 w-full">
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
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                            <Label htmlFor="number" className="sm:text-right text-sm">
                                                Номер
                                            </Label>
                                            <Input
                                                id="number"
                                                value={partEdit.editForm.number || ''}
                                                onChange={(e) => partEdit.updateFormField('number', e.target.value)}
                                                className="sm:col-span-3"
                                            />
                                        </div>
                                    </>
                                )}

                                {/* Defect field for most categories */}
                                {(partEdit.editForm.category !== 'Автохимия и масла' && partEdit.editForm.category !== 'Аксессуары и тюннинг') && (
                                    <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                        <Label htmlFor="defect" className="sm:text-right text-sm">
                                            Дефект
                                        </Label>
                                        <Input
                                            id="defect"
                                            value={partEdit.editForm.defect || ''}
                                            onChange={(e) => partEdit.updateFormField('defect', e.target.value)}
                                            className="sm:col-span-3"
                                        />
                                    </div>
                                )}

                                {/* Car release date for most categories */}
                                {(partEdit.editForm.category !== 'Автохимия и масла' && partEdit.editForm.category !== 'Аксессуары и тюннинг' && partEdit.editForm.category !== 'Другое') && (
                                    <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                                        <Label htmlFor="car_release_date" className="sm:text-right text-sm">
                                            Дата выпуска автомобиля
                                        </Label>
                                        <Input
                                            id="car_release_date"
                                            value={partEdit.editForm.car_release_date || ''}
                                            onChange={(e) => partEdit.updateFormField('car_release_date', e.target.value)}
                                            className="sm:col-span-3"
                                        />
                                    </div>
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
