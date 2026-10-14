/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Crop, Trash2, Plus, Copy, Check, Save } from "lucide-react";
import { PartBasicFields } from "@/features/parts/components/PartBasicFields";
import { PartSpecificationsFields } from "@/features/parts/components/PartSpecificationsFields";
import { partSchema, type PartFormData } from "@/features/parts/schemas/partSchema";
import { useQueryClient } from "@tanstack/react-query";
import { partsKeys } from "@/hooks/useParts";
import { partsApi } from "@/features/parts/api/partsApi";
import { getAuthHeaders } from "@/lib/csrf";
import { copyPhotoToClipboard } from "@/lib/photoUtils";
import { API_BASE_URL } from "@/lib/api";
import type { UsePartEditReturn } from "@/hooks/usePartEdit";
import type { Part } from "@/features/parts/types";

interface EditPartDialogProps {
    partEdit: UsePartEditReturn;
    part: Part;
    onPhotoChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
    onCrop: (src: string | File | undefined, photoPathToReplace?: string) => void;
    onDeletePhoto: (photoPath?: string) => void;
    originalFile: File | null;
}

export default function EditPartDialog({
    partEdit,
    part,
    onPhotoChange,
    onCrop,
    onDeletePhoto,
    originalFile,
}: EditPartDialogProps) {
    const queryClient = useQueryClient();
    const [isSaving, setIsSaving] = useState(false);
    const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

    const form = useForm<PartFormData>({
        resolver: zodResolver(partSchema),
        defaultValues: {
            name: part.name || "",
            brand: part.brand || "",
            model: part.model || "",
            quantity: part.quantity ?? 0,
            price: part.price ?? undefined,
            description: part.description || "",
            category: part.category || "",
            location: part.location || "",
            address: part.address || "",
            seller_id: part.seller_id ? String(part.seller_id) : undefined,
            vin: part.vin || "",
            body_brand: part.body_brand || "",
            engine_brand: part.engine_brand || "",
            car_release_date: part.car_release_date || "",
            car_release_period: part.car_release_period || "",
            front_rear: part.front_rear || "",
            left_right: part.left_right || "",
            top_bottom: part.top_bottom || "",
            color: part.color || "",
            number: part.number || "",
            manufacturer: part.manufacturer || "",
            manufacturer_code: part.manufacturer_code || "",
            oem_code: part.oem_code || "",
            supplier_code: part.supplier_code || "",
            condition: part.condition || "",
            defect: part.defect || "",
            transmission: part.transmission || "",
            transmission_model: part.transmission_model || "",
            drive: part.drive || "",
            wear_percentage: part.wear_percentage || "",
            season: part.season || "",
            diameter: part.diameter || "",
            width: part.width || "",
            profile: part.profile || "",
            tire_quantity: part.tire_quantity || "",
            drilling: part.drilling || "",
            offset: part.offset || "",
            center_hole_diameter: part.center_hole_diameter || "",
            tire_model: part.tire_model || "",
        },
    });

    const existingPhotos = (part.photos && part.photos.length > 0)
        ? part.photos
        : (part.photo ? [part.photo] : []);

    const handleCopyExistingPhoto = async (photoUrl: string, index: number) => {
        try {
            await copyPhotoToClipboard(photoUrl);
            setCopiedIndex(index);
            setTimeout(() => setCopiedIndex(null), 2000);
        } catch (err) {
            console.error("Ошибка копирования фото:", err);
        }
    };

    const handleFormSubmit = async (values: PartFormData) => {
        setIsSaving(true);
        try {
            const updatePayload = {
                ...values,
                seller_id: values.seller_id ? Number(values.seller_id) : undefined,
                price: values.price ? Number(values.price) : 0,
                quantity: Number(values.quantity) || 0,
            };

            const response = await fetch(`${API_BASE_URL}/api/v1/parts/${part.id}`, {
                method: "PUT",
                headers: {
                    ...getAuthHeaders(),
                    "Content-Type": "application/json",
                },
                credentials: "include",
                body: JSON.stringify(updatePayload),
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(`HTTP ${response.status}: ${text}`);
            }

            if (originalFile) {
                await partsApi.uploadPhoto(part.id, originalFile);
            }

            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
            partEdit.setIsEditing(false);
        } catch (err) {
            console.error("Ошибка сохранения запчасти:", err);
            alert("Не удалось сохранить изменения детали.");
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <Dialog open={partEdit.isEditing} onOpenChange={partEdit.setIsEditing}>
            <DialogContent
                className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto"
                onClick={(e) => e.stopPropagation()}
            >
                <DialogHeader>
                    <DialogTitle>Редактировать деталь #{part.id}</DialogTitle>
                    <DialogDescription>
                        Внесите изменения в характеристики запчасти и нажмите &laquo;Сохранить изменения&raquo;.
                    </DialogDescription>
                </DialogHeader>

                <FormProvider {...form}>
                    <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-4">
                        <Tabs defaultValue="basic" className="w-full">
                            <TabsList className="grid w-full grid-cols-2">
                                <TabsTrigger value="basic">Основная информация</TabsTrigger>
                                <TabsTrigger value="specifications">Характеристики</TabsTrigger>
                            </TabsList>

                            {/* Вкладка 1: Основное */}
                            <TabsContent value="basic" className="space-y-4 mt-2">
                                <PartBasicFields />

                                {/* Управление фото */}
                                <div className="border-t border-border pt-4">
                                    <h4 className="text-sm font-semibold mb-3">Фотографии детали</h4>
                                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 mb-3">
                                        {existingPhotos.map((photoPath, index) => {
                                            const photoUrl = photoPath.startsWith("http") || photoPath.startsWith("data:")
                                                ? photoPath
                                                : `${API_BASE_URL}${photoPath.startsWith("/") ? photoPath : `/${photoPath}`}`;
                                            return (
                                                <div key={index} className="relative aspect-square border rounded-lg overflow-hidden group bg-muted/20">
                                                    <img
                                                        src={photoUrl}
                                                        alt={`Фото ${index + 1}`}
                                                        className="w-full h-full object-cover"
                                                    />
                                                    <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-1.5 p-1">
                                                        <Button
                                                            type="button"
                                                            variant="secondary"
                                                            size="icon"
                                                            className="h-7 w-7"
                                                            onClick={() => onCrop(photoUrl, photoPath)}
                                                            title="Отредактировать в фоторедакторе"
                                                        >
                                                            <Crop className="h-3.5 w-3.5" />
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="secondary"
                                                            size="icon"
                                                            className="h-7 w-7"
                                                            onClick={() => handleCopyExistingPhoto(photoUrl, index)}
                                                            title="Скопировать"
                                                        >
                                                            {copiedIndex === index ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
                                                        </Button>
                                                        <Button
                                                            type="button"
                                                            variant="destructive"
                                                            size="icon"
                                                            className="h-7 w-7"
                                                            onClick={() => onDeletePhoto(photoPath)}
                                                            title="Удалить фото"
                                                        >
                                                            <Trash2 className="h-3.5 w-3.5" />
                                                        </Button>
                                                    </div>
                                                </div>
                                            );
                                        })}

                                        {/* Добавить новое фото */}
                                        <label className="border-2 border-dashed border-border hover:border-primary rounded-lg aspect-square flex flex-col items-center justify-center cursor-pointer transition-colors p-2 text-center">
                                            <input
                                                type="file"
                                                accept="image/*"
                                                onChange={onPhotoChange}
                                                className="hidden"
                                            />
                                            <Plus className="h-5 w-5 text-muted-foreground mb-1" />
                                            <span className="text-xs text-muted-foreground">Добавить фото</span>
                                        </label>
                                    </div>

                                    {originalFile && (
                                        <div className="flex items-center justify-between p-2 rounded-lg bg-primary/10 text-xs">
                                            <span className="truncate">Новое фото: {originalFile.name}</span>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                className="h-7 text-xs"
                                                onClick={() => onCrop(originalFile)}
                                            >
                                                <Crop className="h-3 w-3 mr-1" />
                                                Обрезать
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            </TabsContent>

                            {/* Вкладка 2: Характеристики */}
                            <TabsContent value="specifications" className="mt-2">
                                <PartSpecificationsFields />
                            </TabsContent>
                        </Tabs>

                        <DialogFooter className="mt-6 gap-2">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => partEdit.setIsEditing(false)}
                                disabled={isSaving}
                            >
                                Отмена
                            </Button>
                            <Button type="submit" disabled={isSaving} className="gap-2">
                                <Save className="h-4 w-4" />
                                {isSaving ? "Сохранение..." : "Сохранить изменения"}
                            </Button>
                        </DialogFooter>
                    </form>
                </FormProvider>
            </DialogContent>
        </Dialog>
    );
}
