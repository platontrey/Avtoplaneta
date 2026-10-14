/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Save, ArrowLeft, Upload, Crop, Trash2 } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { getAuthHeaders } from "@/lib/csrf";
import { PartBasicFields } from "@/features/parts/components/PartBasicFields";
import { PartSpecificationsFields } from "@/features/parts/components/PartSpecificationsFields";
import { partSchema, type PartFormData } from "@/features/parts/schemas/partSchema";
import ImageEditor from "./ImageEditor";
import { API_BASE_URL } from "@/lib/api";

export default function AddPart() {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [photoFiles, setPhotoFiles] = useState<File[]>([]);
    const [photoPreviews, setPhotoPreviews] = useState<string[]>([]);
    const [showCropper, setShowCropper] = useState(false);
    const [cropImageIndex, setCropImageIndex] = useState<number | null>(null);

    const form = useForm<PartFormData>({
        resolver: zodResolver(partSchema),
        defaultValues: {
            brand: "",
            name: "",
            model: "",
            quantity: 1,
            price: undefined,
            description: "",
            category: "",
            location: "",
            address: "",
            vin: "",
        },
    });

    const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []);
        const validFiles: File[] = [];
        const validPreviews: string[] = [];

        for (const file of files) {
            if (!file.type.startsWith("image/")) {
                alert(`Файл ${file.name} не является изображением.`);
                continue;
            }
            if (file.size > 30 * 1024 * 1024) {
                alert(`Файл ${file.name} превышает 30МБ.`);
                continue;
            }
            validFiles.push(file);
            validPreviews.push(URL.createObjectURL(file));
        }

        if (validFiles.length > 0) {
            // Если поле имени запчасти пустое, автозаполняем именем первого файла
            const currentName = form.getValues("name");
            if (!currentName) {
                const fileNameWithoutExt = validFiles[0].name.replace(/\.[^/.]+$/, "");
                form.setValue("name", fileNameWithoutExt);
            }

            setPhotoFiles((prev) => [...prev, ...validFiles]);
            setPhotoPreviews((prev) => [...prev, ...validPreviews]);
        }
    };

    const handleRemovePhoto = (index: number) => {
        setPhotoFiles((prev) => prev.filter((_, i) => i !== index));
        setPhotoPreviews((prev) => prev.filter((_, i) => i !== index));
    };

    const handleCropComplete = (croppedImageBlob: Blob) => {
        if (cropImageIndex === null || croppedImageBlob.size === 0) return;
        const croppedFile = new File([croppedImageBlob], "cropped-image.jpg", { type: "image/jpeg" });
        const newPreview = URL.createObjectURL(croppedFile);

        setPhotoFiles((prev) => {
            const next = [...prev];
            next[cropImageIndex] = croppedFile;
            return next;
        });

        setPhotoPreviews((prev) => {
            const next = [...prev];
            next[cropImageIndex] = newPreview;
            return next;
        });

        setShowCropper(false);
        setCropImageIndex(null);
    };

    const onSubmit = async (data: PartFormData) => {
        setLoading(true);
        const newPart = {
            ...data,
            quantity: Number(data.quantity) || 0,
            price: data.price ? Number(data.price) : 0,
            seller_id: data.seller_id ? parseInt(data.seller_id) : undefined,
        };

        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/parts`, {
                method: "POST",
                headers: {
                    ...getAuthHeaders(),
                    "Content-Type": "application/json",
                },
                credentials: "include",
                body: JSON.stringify(newPart),
            });

            if (!response.ok) {
                const text = await response.text();
                alert(text || `Ошибка сервера: ${response.status}`);
                setLoading(false);
                return;
            }

            const result = await response.json().catch(() => ({ message: "OK" }));

            // Параллельная загрузка фотографий
            if (photoFiles.length > 0 && result.id) {
                const uploadPromises = photoFiles.map(async (file) => {
                    const formData = new FormData();
                    formData.append("photo", file);
                    return fetch(`${API_BASE_URL}/api/v1/uploadpartphoto/${result.id}`, {
                        method: "POST",
                        headers: {
                            "X-CSRF-Token": getAuthHeaders()["X-CSRF-Token"] || "",
                        },
                        credentials: "include",
                        body: formData,
                    });
                });

                await Promise.allSettled(uploadPromises);
            }

            alert(photoFiles.length > 0 ? `Запчасть и ${photoFiles.length} фото успешно добавлены!` : "Запчасть успешно добавлена!");
            navigate("/inventory");
        } catch (err) {
            console.error("Ошибка при сохранении:", err);
            alert("Ошибка сети при добавлении запчасти");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-5xl mx-auto p-4 sm:p-8">
            <Link
                to="/inventory"
                className="flex items-center gap-1 text-muted-foreground hover:text-foreground mb-6 text-sm sm:text-base"
            >
                <ArrowLeft size={16} />
                Назад к инвентарю
            </Link>

            <h2 className="text-2xl sm:text-3xl font-semibold mb-2">Добавить запчасть</h2>
            <p className="text-muted-foreground mb-6 text-sm sm:text-base">
                Заполните форму для добавления новой запчасти в складской каталог.
            </p>

            <div className="bg-card text-card-foreground rounded-xl border border-border p-4 sm:p-8 shadow-sm">
                <FormProvider {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                        <Tabs defaultValue="basic" className="w-full">
                            <TabsList className="grid w-full grid-cols-2">
                                <TabsTrigger value="basic">Основная информация</TabsTrigger>
                                <TabsTrigger value="specifications">Характеристики</TabsTrigger>
                            </TabsList>

                            {/* Вкладка 1: Основные поля и загрузка фото */}
                            <TabsContent value="basic" className="space-y-6 mt-4">
                                <PartBasicFields />

                                {/* Блок выбора фотографий */}
                                <div className="border-t border-border pt-4">
                                    <h4 className="text-sm font-semibold mb-2">Фотографии запчасти</h4>
                                    <p className="text-xs text-muted-foreground mb-4">
                                        Первое фото будет обложкой детали. Имя первой фотографии автоматически подставляется в название детали.
                                    </p>

                                    <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-5 gap-3">
                                        {photoPreviews.map((previewUrl, index) => (
                                            <div key={index} className="relative aspect-square border rounded-lg overflow-hidden group bg-muted/20">
                                                <img
                                                    src={previewUrl}
                                                    alt={`Превью ${index + 1}`}
                                                    className="w-full h-full object-cover"
                                                />
                                                <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                                                    <Button
                                                        type="button"
                                                        variant="secondary"
                                                        size="icon"
                                                        className="h-8 w-8"
                                                        onClick={() => {
                                                            setCropImageIndex(index);
                                                            setShowCropper(true);
                                                        }}
                                                        title="Отредактировать фото"
                                                    >
                                                        <Crop className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        type="button"
                                                        variant="destructive"
                                                        size="icon"
                                                        className="h-8 w-8"
                                                        onClick={() => handleRemovePhoto(index)}
                                                        title="Удалить"
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </div>
                                        ))}

                                        {/* Загрузка новых фото */}
                                        <label className="border-2 border-dashed border-border hover:border-primary rounded-lg aspect-square flex flex-col items-center justify-center cursor-pointer transition-colors p-3 text-center">
                                            <input
                                                type="file"
                                                multiple
                                                accept="image/*"
                                                onChange={handlePhotoChange}
                                                className="hidden"
                                            />
                                            <Upload className="h-6 w-6 text-muted-foreground mb-2" />
                                            <span className="text-xs font-medium text-foreground">Загрузить фото</span>
                                            <span className="text-[10px] text-muted-foreground mt-0.5">JPG, PNG до 30МБ</span>
                                        </label>
                                    </div>
                                </div>
                            </TabsContent>

                            {/* Вкладка 2: Характеристики */}
                            <TabsContent value="specifications" className="mt-4">
                                <PartSpecificationsFields />
                            </TabsContent>
                        </Tabs>

                        <div className="flex justify-end gap-3 pt-4 border-t border-border">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => navigate("/inventory")}
                                disabled={loading}
                            >
                                Отмена
                            </Button>
                            <Button type="submit" disabled={loading} className="gap-2">
                                <Save className="h-4 w-4" />
                                {loading ? "Сохранение..." : "Добавить запчасть"}
                            </Button>
                        </div>
                    </form>
                </FormProvider>
            </div>

            {/* Фоторедактор (монтируется только по клику на обрезку) */}
            {showCropper && cropImageIndex !== null && photoPreviews[cropImageIndex] && (
                <ImageEditor
                    src={photoPreviews[cropImageIndex]}
                    aspect={null}
                    onEditComplete={handleCropComplete}
                    onCancel={() => {
                        setShowCropper(false);
                        setCropImageIndex(null);
                    }}
                />
            )}
        </div>
    );
}
