/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ClearableSelect } from "@/components/ClearableSelect";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { Save, ArrowLeft, Upload } from "lucide-react";
import { Link } from "react-router-dom";
import { getAuthHeaders } from "@/lib/csrf";
import { sanitizeHtml } from "@/lib/security";
import ImageEditor from "./ImageEditor";
import { API_BASE_URL } from '@/lib/api';
import { usePartCatalog } from '@/features/catalog/usePartCatalog';

const brandOptions = [
        { value: "Acura", label: "Acura" },
        { value: "Aston Martin", label: "Aston Martin" },
        { value: "Audi", label: "Audi" },
        { value: "Bentley", label: "Bentley" },
        { value: "BMW", label: "BMW" },
        { value: "Buick", label: "Buick" },
        { value: "Cadillac", label: "Cadillac" },
        { value: "Chevrolet", label: "Chevrolet" },
        { value: "Chrysler", label: "Chrysler" },
        { value: "Citroen", label: "Citroën" },
        { value: "DAF", label: "DAF" },
        { value: "Daihatsu", label: "Daihatsu" },
        { value: "Dodge", label: "Dodge" },
        { value: "Ferrari", label: "Ferrari" },
        { value: "Fiat", label: "Fiat" },
        { value: "Ford", label: "Ford" },
        { value: "GMC", label: "GMC" },
        { value: "Hino", label: "Hino" },
        { value: "Honda", label: "Honda" },
        { value: "Hyundai", label: "Hyundai" },
        { value: "Infiniti", label: "Infiniti" },
        { value: "Isuzu", label: "Isuzu" },
        { value: "Iveco", label: "Iveco" },
        { value: "Jaguar", label: "Jaguar" },
        { value: "Jeep", label: "Jeep" },
        { value: "Kenworth", label: "Kenworth" },
        { value: "Kia", label: "Kia" },
        { value: "Lamborghini", label: "Lamborghini" },
        { value: "Land Rover", label: "Land Rover" },
        { value: "Lexus", label: "Lexus" },
        { value: "Lincoln", label: "Lincoln" },
        { value: "Mack", label: "Mack" },
        { value: "MAN", label: "MAN" },
        { value: "Mazda", label: "Mazda" },
        { value: "Mercedes", label: "Mercedes-Benz" },
        { value: "Mercedes-Benz Trucks", label: "Mercedes-Benz Trucks" },
        { value: "Mitsubishi", label: "Mitsubishi" },
        { value: "Nissan", label: "Nissan" },
        { value: "Opel", label: "Opel" },
        { value: "Peterbilt", label: "Peterbilt" },
        { value: "Peugeot", label: "Peugeot" },
        { value: "Porsche", label: "Porsche" },
        { value: "Ram", label: "Ram" },
        { value: "Renault", label: "Renault" },
        { value: "Rolls-Royce", label: "Rolls-Royce" },
        { value: "Scania", label: "Scania" },
        { value: "Subaru", label: "Subaru" },
        { value: "Suzuki", label: "Suzuki" },
        { value: "Tesla", label: "Tesla" },
        { value: "Toyota", label: "Toyota" },
        { value: "UD Trucks", label: "UD Trucks" },
        { value: "Volkswagen", label: "Volkswagen" },
        { value: "Volvo", label: "Volvo" },
        { value: "Volvo Trucks", label: "Volvo Trucks" },
        { value: "Western Star", label: "Western Star" }
];

const partSchema = z.object({
    brand: z.string().min(1, "Выберите бренд"),
    name: z.string().min(1, "Введите название запчасти").max(255, "Название слишком длинное (макс 255 символов)"),
    model: z.string().min(1, "Введите модель"),
    quantity: z.number().min(0, "Количество должно быть положительным числом"),
    description: z.string().max(1000, "Описание слишком длинное (макс 1000 символов)").optional(),
    category: z.string().optional(),
    location: z.string().optional(),
    address: z.string().optional(),
    price: z.number().optional(),
    seller_id: z.string().optional(),
    // Характеристики запчасти
    body_brand: z.string().optional(),
    engine_brand: z.string().optional(),
    car_release_date: z.string().optional(),
    front_rear: z.string().optional(),
    left_right: z.string().optional(),
    top_bottom: z.string().optional(),
    number: z.string().optional(),
    manufacturer: z.string().optional(),
    manufacturer_code: z.string().optional(),
    oem_code: z.string().optional(),
    color: z.string().optional(),
    condition: z.string().optional(),
    supplier_code: z.string().optional(),
    defect: z.string().optional(),
    transmission: z.string().optional(),
    transmission_model: z.string().optional(),
    drive: z.string().optional(),
    wear_percentage: z.string().optional(),
    season: z.string().optional(),
    diameter: z.string().optional(),
    width: z.string().optional(),
    profile: z.string().optional(),
    tire_quantity: z.string().optional(),
    drilling: z.string().optional(),
    offset: z.string().optional(),
    center_hole_diameter: z.string().optional(),
    tire_model: z.string().optional(),
    vin: z.string().optional(),
});

type PartFormData = z.infer<typeof partSchema>;

interface Part {
    brand: string;
    name: string;
    quantity: number;
    description?: string;
    category?: string;
    model: string;
    location?: string;
    address?: string;
    price?: number;
    seller_id?: number;
    photo?: string;
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
    vin?: string;
}

interface User {
    id: number;
    name: string;
    email: string;
}


export default function AddPart() {
    const [loading, setLoading] = useState<boolean>(false);
    const [photoFiles, setPhotoFiles] = useState<File[]>([]);
    const [photoPreviews, setPhotoPreviews] = useState<string[]>([]);
    const [users, setUsers] = useState<User[]>([]);
    const [loadingUsers, setLoadingUsers] = useState<boolean>(true);
    const [showCropper, setShowCropper] = useState<boolean>(false);
    const [tempImageSrc, setTempImageSrc] = useState<string | File>("");
    const { data: partCatalog, error: catalogError } = usePartCatalog();

    const {
        register,
        handleSubmit,
        setValue,
        watch,
        formState: { errors },
    } = useForm<PartFormData>({
        resolver: zodResolver(partSchema),
        defaultValues: {
            quantity: 1,
            price: 100,
            seller_id: undefined,
        },
    });

    const brand = watch("brand");
    const category = watch("category");
    const sellerId = watch("seller_id");


    const visibleFields = category
        ? partCatalog?.part_form_categories.find((item) => item.name === category)?.attributes ?? []
        : partCatalog?.attributes.map((attribute) => attribute.code) ?? [];
    const transmissionOptions = partCatalog?.attributes.find(
        (attribute) => attribute.code === "transmission",
    )?.options ?? [];
    const driveOptions = partCatalog?.attributes.find(
        (attribute) => attribute.code === "drive",
    )?.options ?? [];

    // Загрузка списка пользователей и шаблонов характеристик при монтировании компонента
    useEffect(() => {
        const fetchUsers = async () => {
            try {
                const response = await fetch(`${API_BASE_URL}/admin/users`, {
                    method: "GET",
                    headers: getAuthHeaders(),
                    credentials: 'include',
                });

                if (response.ok) {
                    const data = await response.json();
                    setUsers(data.users || []);
                } else {
                    console.error("Failed to fetch users:", response.status);
                }
            } catch (error) {
                console.error("Error fetching users:", error);
            } finally {
                setLoadingUsers(false);
            }
        };

        void fetchUsers();
    }, []);


    const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []);
        console.log('AddPart handlePhotoChange: files selected:', files.map(f => f.name));

        const validFiles: File[] = [];
        const validPreviews: string[] = [];

        for (const file of files) {
            // Validate file type - only allow images
            if (!file.type.startsWith('image/')) {
                alert(`Файл ${file.name} не является изображением.`);
                continue;
            }

            // Validate file size (max 5MB to prevent memory issues)
            const maxSize = 5 * 1024 * 1024; // 5MB
            if (file.size > maxSize) {
                alert(`Размер файла ${file.name} должен быть менее 5МБ.`);
                continue;
            }

            validFiles.push(file);
            validPreviews.push(URL.createObjectURL(file));
        }

        if (validFiles.length > 0) {
            // Copy the first file name (without extension) to the name field
            const firstFile = validFiles[0];
            const fileNameWithoutExt = firstFile.name.replace(/\.[^/.]+$/, "");
            setValue('name', fileNameWithoutExt);

            setPhotoFiles(prev => [...prev, ...validFiles]);
            setPhotoPreviews(prev => [...prev, ...validPreviews]);
        }
    };

    const handleCropComplete = (croppedImageBlob: Blob) => {
        console.log('AddPart handleCropComplete: called with blob size:', croppedImageBlob.size, 'type:', croppedImageBlob.type);
        if (croppedImageBlob.size === 0) {
            console.log('AddPart handleCropComplete: Error - cropped image is empty');
            alert('Ошибка: обрезанное изображение пустое');
            return;
        }
        const croppedFile = new File([croppedImageBlob], 'cropped-image.jpg', { type: 'image/jpeg' });
        console.log('AddPart handleCropComplete: Created file:', croppedFile.name, 'size:', croppedFile.size, 'type:', croppedFile.type);

        // Replace the first photo with cropped version
        setPhotoFiles(prev => {
            const newFiles = [...prev];
            if (newFiles.length > 0) {
                newFiles[0] = croppedFile;
            } else {
                newFiles.push(croppedFile);
            }
            return newFiles;
        });

        setPhotoPreviews(prev => {
            const newPreviews = [...prev];
            if (newPreviews.length > 0) {
                // Revoke old URL to prevent memory leaks
                URL.revokeObjectURL(newPreviews[0]);
                newPreviews[0] = URL.createObjectURL(croppedFile);
            } else {
                newPreviews.push(URL.createObjectURL(croppedFile));
            }
            return newPreviews;
        });

        setShowCropper(false);
        setTempImageSrc("");
    };

    const handleCropCancel = () => {
        setShowCropper(false);
        setTempImageSrc("");
        // Reset the input
        const input = document.getElementById('photo') as HTMLInputElement;
        if (input) input.value = '';
    };



    const onSubmit = async (data: PartFormData) => {
        setLoading(true);

        // Санитизировать входные данные с помощью DOMPurify
        const sanitizedName = sanitizeHtml(data.name, 'NAME');
        const sanitizedDescription = data.description ? sanitizeHtml(data.description, 'DESCRIPTION') : undefined;
        const sanitizedLocation = data.location ? sanitizeHtml(data.location, 'NAME') : undefined;
        const sanitizedAddress = data.address ? sanitizeHtml(data.address, 'NAME') : undefined;

        const newPart: Part = {
            brand: data.brand,
            name: sanitizedName,
            model: data.model,
            quantity: data.quantity,
            description: sanitizedDescription,
            category: data.category,
            location: sanitizedLocation,
            address: sanitizedAddress,
            price: data.price,
            seller_id: data.seller_id ? parseInt(data.seller_id) : undefined,
            // Характеристики запчасти (теперь хранятся в основной таблице Part)
            body_brand: data.body_brand,
            engine_brand: data.engine_brand,
            car_release_date: data.car_release_date,
            front_rear: data.front_rear,
            left_right: data.left_right,
            top_bottom: data.top_bottom,
            number: data.number,
            manufacturer: data.manufacturer,
            manufacturer_code: data.manufacturer_code,
            oem_code: data.oem_code,
            color: data.color,
            condition: data.condition,
            supplier_code: data.supplier_code,
            defect: data.defect,
            transmission: data.transmission,
            transmission_model: data.transmission_model,
            drive: data.drive,
            wear_percentage: data.wear_percentage,
            season: data.season,
            diameter: data.diameter,
            width: data.width,
            profile: data.profile,
            tire_quantity: data.tire_quantity,
            drilling: data.drilling,
            offset: data.offset,
            center_hole_diameter: data.center_hole_diameter,
            tire_model: data.tire_model,
            vin: data.vin
        };

        try {
            // Сначала добавить запчасть
            const response = await fetch(`${API_BASE_URL}/api/v1/parts`, {
                method: "POST",
                headers: getAuthHeaders(),
                credentials: 'include', // Include cookies in the request
                body: JSON.stringify(newPart),
            });

            if (!response.ok) {
                const text = await response.text().catch(() => null);
                alert(text || `Ошибка сервера: ${response.status}`);
                setLoading(false);
                return;
            }

            const result = await response.json().catch(() => ({ message: "OK" }));


            // Если фото выбраны, загрузить их параллельно для улучшения производительности
             if (photoFiles.length > 0) {
                  console.log('AddPart onSubmit: Uploading photos for new part, photoFiles:', photoFiles.length, 'part ID:', result.id);

                  // Краткая задержка для обеспечения полного создания запчасти
                  await new Promise(resolve => setTimeout(resolve, 100));

                  try {
                      // Загружаем все фото параллельно для оптимизации INP
                      const uploadPromises = photoFiles.map(async (photoFile, i) => {
                          console.log(`AddPart onSubmit: Uploading photo ${i + 1}/${photoFiles.length}:`, photoFile.name, 'size:', photoFile.size, 'type:', photoFile.type);

                          const photoFormData = new FormData();
                          photoFormData.append("photo", photoFile);

                          const uploadUrl = `${API_BASE_URL}/api/uploadpartphoto/${result.id}`;
                          console.log('AddPart onSubmit: Making photo upload request to:', uploadUrl);

                          const photoResponse = await fetch(uploadUrl, {
                              method: "POST",
                              headers: {
                                  'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
                              },
                              credentials: 'include',
                              body: photoFormData,
                          });

                          console.log(`AddPart onSubmit: Photo ${i + 1} upload response status:`, photoResponse.status);

                          if (!photoResponse.ok) {
                              const errorText = await photoResponse.text();
                              console.error(`AddPart onSubmit: Failed to upload photo ${i + 1}:`, errorText);
                              throw new Error(`Failed to upload photo ${i + 1}`);
                          } else {
                              const photoResult = await photoResponse.json();
                              console.log(`AddPart onSubmit: Photo ${i + 1} upload successful:`, photoResult);
                              return photoResult;
                          }
                      });

                      // Ждем завершения всех загрузок параллельно
                      await Promise.allSettled(uploadPromises);
                  } catch (error) {
                      console.error('AddPart onSubmit: Error during photo uploads:', error);
                      alert('Запчасть создана, но загрузка некоторых фото не удалась.');
                  }
              } else {
                  console.log('AddPart onSubmit: No photo files selected for upload');
              }

            if (photoFiles.length > 0) {
                 alert(`Запчасть и ${photoFiles.length} фото успешно добавлены!`);
             } else {
                 alert(result.message || "Запчасть успешно добавлена!");
             }

            // Перейти к инвентарю и обновить для отображения новой запчасти
            window.location.href = '/inventory';
        } catch (err) {
            console.error("Ошибка сети:", err);
            alert("Ошибка сети при добавлении запчасти");
        } finally {
            setLoading(false);
            console.timeEnd('AddPart.onSubmit');
        }
    };

    return (
        <div className="max-w-5xl mx-auto p-4 sm:p-8">
            <Link
                to="/inventory"
                className="flex items-center gap-1 text-gray-600 hover:text-black mb-6 text-sm sm:text-base"
            >
                <ArrowLeft size={16} />
                Назад к инвентарю
            </Link>

            <h2 className="text-2xl sm:text-3xl font-semibold mb-4">Добавить запчасть</h2>
            <p className="text-gray-500 mb-8 text-base sm:text-lg">
                Заполните форму для добавления новой запчасти в инвентарь.
            </p>

            <div className="bg-white rounded-lg border p-4 sm:p-8 shadow-lg w-full max-w-4xl h-auto min-h-[600px] sm:min-h-[800px]">
                <Tabs defaultValue="basic" className="w-full">
                    <TabsList className="grid w-full grid-cols-2">
                        <TabsTrigger value="basic">Основная информация</TabsTrigger>
                        <TabsTrigger value="specifications">Характеристики</TabsTrigger>
                    </TabsList>

                    <form onSubmit={handleSubmit(onSubmit)}>
                        <TabsContent value="basic" className="space-y-4 mt-6">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-8">
                                {/* Левая колонка */}
                                <div className="space-y-4 sm:space-y-6">
                                    <div>
                                        <Label htmlFor="brand-select" className="min-w-[120px] mb-1">Бренд</Label>
                                        <div className="relative">
                                            <SearchableSelect
                                                value={brand || ""}
                                                onValueChange={(value) => setValue("brand", value)}
                                                options={brandOptions}
                                                placeholder="Выберите бренд"
                                                searchPlaceholder="Поиск бренда..."
                                                emptyMessage="Бренд не найден"
                                                className="h-10"
                                            />
                                            
                                        </div>
                                        <input
                                            type="hidden"
                                            {...register("brand")}
                                            autoComplete="organization"
                                        />
                                        {errors.brand && <p className="text-red-500 text-sm">{errors.brand.message}</p>}
                                    </div>

                                    <div>
                                        <Label htmlFor="name" className="min-w-[120px] mb-1">Название запчасти</Label>
                                        <Input
                                            id="name"
                                            {...register("name")}
                                            type="text"
                                            placeholder="Front Brake Disc"
                                            className="h-10"
                                            autoComplete="off"
                                        />
                                        {errors.name && <p className="text-red-500 text-sm">{errors.name.message}</p>}
                                    </div>

                                    <div>
                                        <Label htmlFor="category-select" className="min-w-[120px] mb-1">Категория</Label>
                                        <div className="relative">
                                            <ClearableSelect
    value={category || ""}
    onValueChange={(value) => setValue("category", value)}
    placeholder="Выберите категорию" id="category-select" className="h-10 w-full"
>
    {partCatalog?.part_form_categories.map((item) => (
                                                        <SelectItem key={item.code} value={item.name}>{item.name}</SelectItem>
                                                    ))}
</ClearableSelect>
                                            {catalogError && (
                                                <p className="mt-1 text-sm text-red-500">{catalogError.message}</p>
                                            )}
                                            
                                        </div>
                                        <input
                                            type="hidden"
                                            {...register("category")}
                                            autoComplete="off"
                                        />
                                    </div>

                                    <div>
                                        <Label htmlFor="quantity" className="min-w-[120px] mb-1">Количество</Label>
                                        <Input
                                            id="quantity"
                                            {...register("quantity", { valueAsNumber: true })}
                                            type="number"
                                            className="h-10"
                                            autoComplete="off"
                                        />
                                        {errors.quantity && <p className="text-red-500 text-sm">{errors.quantity.message}</p>}
                                    </div>

                                    <div>
                                        <Label htmlFor="vin" className="min-w-[120px] mb-1">VIN</Label>
                                        <Input
                                            id="vin"
                                            {...register("vin")}
                                            type="text"
                                            placeholder="WVWZZZ1JZ3W386549"
                                            className="h-10"
                                            autoComplete="off"
                                        />
                                    </div>
                                </div>

                                {/* Правая колонка */}
                                <div className="space-y-4 sm:space-y-6">
                                    <div>
                                        <Label htmlFor="model" className="min-w-[120px] mb-1">Модель</Label>
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
                                        <Label htmlFor="location" className="min-w-[120px] mb-1">Местоположение</Label>
                                        <Input
                                            id="location"
                                            {...register("location")}
                                            type="text"
                                            placeholder="Shelf A-12"
                                            className="h-10"
                                            autoComplete="shipping location"
                                        />
                                    </div>

                                    <div>
                                        <Label htmlFor="address" className="min-w-[120px] mb-1">Адрес склада</Label>
                                        <Input
                                            id="address"
                                            {...register("address")}
                                            type="text"
                                            placeholder="Профсоюзная 2/11"
                                            className="h-10"
                                            autoComplete="street-address"
                                        />
                                    </div>

                                    <div>
                                        <Label htmlFor="price" className="min-w-[120px] mb-1">Цена (₽)</Label>
                                        <Input
                                            id="price"
                                            {...register("price")}
                                            type="number"
                                            step="100"
                                            className="h-10"
                                            autoComplete="transaction-amount"
                                        />
                                    </div>

                                    <div>
                                        <Label htmlFor="seller-select" className="min-w-[120px] mb-1">Продавец</Label>
                                        <div className="relative">
                                            <Select
                                                value={sellerId?.toString() || ""}
                                                onValueChange={(value) => setValue("seller_id", value === "" ? undefined : value)}
                                                disabled={loadingUsers}
                                            >
                                                <SelectTrigger id="seller-select" className="h-10 w-full">
                                                    <SelectValue placeholder={loadingUsers ? "Загрузка..." : "Выберите продавца"} />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {users.map((user) => (
                                                        <SelectItem key={user.id} value={user.id.toString()}>
                                                            {user.name} ({user.email})
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            
                                        </div>
                                        <input
                                            type="hidden"
                                            id="seller_id"
                                            {...register("seller_id")}
                                            autoComplete="off"
                                        />
                                        {errors.seller_id && <p className="text-red-500 text-sm">{errors.seller_id.message}</p>}
                                    </div>

                                    <div>
                                        <Label htmlFor="condition" className="min-w-[120px] mb-1">Состояние</Label>
                                        <Input
                                            id="condition"
                                            {...register("condition")}
                                            type="text"
                                            placeholder="Б/у или новый"
                                            className="h-10"
                                            autoComplete="off"
                                        />
                                    </div>
                                </div>
                            </div>

                            {/* Секция загрузки фото */}
                            <div className="mt-6 sm:mt-8">
                                <Label htmlFor="photo" className="mb-1">Фото запчасти</Label>
                                <div className="mt-3">
                                    <input
                                        type="file"
                                        id="photo"
                                        name="photo"
                                        accept="image/*"
                                        multiple
                                        onChange={handlePhotoChange}
                                        className="hidden"
                                        autoComplete="off"
                                    />
                                    <label
                                        htmlFor="photo"
                                        className="flex items-center justify-center w-full h-24 sm:h-32 border-2 border-dashed border-gray-300 rounded-lg cursor-pointer hover:border-gray-400 transition-colors"
                                    >
                                        {photoPreviews.length > 0 ? (
                                            <div className="flex flex-wrap gap-2 justify-center">
                                                {photoPreviews.slice(0, 3).map((preview, index) => (
                                                    <img
                                                        key={index}
                                                        src={preview}
                                                        alt={`Preview ${index + 1}`}
                                                        className="w-16 h-16 sm:w-20 sm:h-20 object-cover rounded"
                                                    />
                                                ))}
                                                {photoPreviews.length > 3 && (
                                                    <div className="w-16 h-16 sm:w-20 sm:h-20 bg-gray-200 rounded flex items-center justify-center text-sm text-gray-600">
                                                        +{photoPreviews.length - 3}
                                                    </div>
                                                )}
                                            </div>
                                        ) : (
                                            <div className="text-center">
                                                <Upload className="mx-auto h-6 w-6 sm:h-8 sm:w-8 text-gray-400" />
                                                <p className="mt-2 text-xs sm:text-sm text-gray-500">Нажмите для выбора фото (несколько)</p>
                                            </div>
                                        )}
                                    </label>
                                    {photoFiles.length > 0 && (
                                        <div className="mt-2 flex justify-center gap-2">
                                            <Button
                                                type="button"
                                                onClick={() => {
                                                    setTempImageSrc(photoFiles[0]);
                                                    setShowCropper(true);
                                                }}
                                                variant="outline"
                                                size="sm"
                                            >
                                                Изменить первое фото
                                            </Button>
                                            <Button
                                                type="button"
                                                onClick={() => {
                                                    setPhotoFiles([]);
                                                    setPhotoPreviews(prev => {
                                                        prev.forEach(URL.revokeObjectURL);
                                                        return [];
                                                    });
                                                }}
                                                variant="outline"
                                                size="sm"
                                            >
                                                Очистить
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Раздел заметок во всю ширину */}
                            <div className="mt-6 sm:mt-8">
                                <Label htmlFor="description" className="mb-1">Заметки</Label>
                                <Textarea
                                    id="description"
                                    {...register("description")}
                                    placeholder="Additional information..."
                                    className="min-h-[100px] sm:min-h-[140px] mt-3"
                                    autoComplete="off"
                                />
                                {errors.description && <p className="text-red-500 text-sm">{errors.description.message}</p>}
                            </div>

                            {/* Кнопка «Отправить» внутри формы */}
                            <div className="mt-6 sm:mt-8 flex justify-start">
                                <Button
                                    type="submit"
                                    disabled={loading}
                                    className="px-8 sm:px-12 py-3 sm:py-4 text-sm sm:text-base"
                                >
                                    <Save className="w-4 h-4 sm:w-5 sm:h-5 mr-2 sm:mr-3" />
                                    {loading ? "Сохранение..." : "Сохранить"}
                                </Button>
                            </div>
                        </TabsContent>

                        <TabsContent value="specifications" className="space-y-4 mt-6 max-h-[750px] overflow-y-auto">
                            <div className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <h3 className="text-lg font-medium">Характеристики запчасти</h3>
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    {visibleFields.includes("body_brand") && (
                                        <div>
                                            <Label htmlFor="body_brand" className="mb-1">Марка кузова</Label>
                                            <Input
                                                id="body_brand"
                                                {...register("body_brand")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("engine_brand") && (
                                        <div>
                                            <Label htmlFor="engine_brand" className="mb-1">Марка двигателя</Label>
                                            <Input
                                                id="engine_brand"
                                                {...register("engine_brand")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("car_release_date") && (
                                        <div>
                                            <Label htmlFor="car_release_date" className="mb-1">Дата выпуска автомобиля</Label>
                                            <Input
                                                id="car_release_date"
                                                {...register("car_release_date")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("front_rear") && (
                                        <div>
                                            <Label htmlFor="front_rear-select" className="mb-1">Перед/зад</Label>
                                            <div className="relative">
                                                <ClearableSelect
    value={watch("front_rear") || ""}
    onValueChange={(value) => setValue("front_rear", value)}
    placeholder="Выберите" id="front_rear-select" className="h-10 w-full"
>
    <SelectItem value="F">F (Перед)</SelectItem>
                                                        <SelectItem value="R">R (Зад)</SelectItem>
</ClearableSelect>
                                                
                                            </div>
                                            <input
                                                type="hidden"
                                                {...register("front_rear")}
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("left_right") && (
                                        <div>
                                            <Label htmlFor="left_right-select" className="mb-1">Право/лево</Label>
                                            <div className="relative">
                                                <ClearableSelect
    value={watch("left_right") || ""}
    onValueChange={(value) => setValue("left_right", value)}
    placeholder="Выберите" id="left_right-select" className="h-10 w-full"
>
    <SelectItem value="L">L (Лево)</SelectItem>
                                                        <SelectItem value="R">R (Право)</SelectItem>
</ClearableSelect>
                                                
                                            </div>
                                            <input
                                                type="hidden"
                                                {...register("left_right")}
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("top_bottom") && (
                                        <div>
                                            <Label htmlFor="top_bottom-select" className="mb-1">Верх/низ</Label>
                                            <div className="relative">
                                                <ClearableSelect
    value={watch("top_bottom") || ""}
    onValueChange={(value) => setValue("top_bottom", value)}
    placeholder="Выберите" id="top_bottom-select" className="h-10 w-full"
>
    <SelectItem value="Верх">Верх</SelectItem>
                                                        <SelectItem value="Низ">Низ</SelectItem>
                                                        <SelectItem value="Середина">Середина</SelectItem>
</ClearableSelect>
                                                
                                            </div>
                                            <input
                                                type="hidden"
                                                {...register("top_bottom")}
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("number") && (
                                        <div>
                                            <Label htmlFor="number" className="mb-1">Номер</Label>
                                            <Input
                                                id="number"
                                                {...register("number")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("manufacturer") && (
                                        <div>
                                            <Label htmlFor="manufacturer" className="mb-1">Производитель</Label>
                                            <Input
                                                id="manufacturer"
                                                {...register("manufacturer")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("manufacturer_code") && (
                                        <div>
                                            <Label htmlFor="manufacturer_code" className="mb-1">Код производителя</Label>
                                            <Input
                                                id="manufacturer_code"
                                                {...register("manufacturer_code")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("oem_code") && (
                                        <div>
                                            <Label htmlFor="oem_code" className="mb-1">OEM код</Label>
                                            <Input
                                                id="oem_code"
                                                {...register("oem_code")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("color") && (
                                        <div>
                                            <Label htmlFor="color" className="mb-1">Цвет</Label>
                                            <Input
                                                id="color"
                                                {...register("color")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}


                                    {visibleFields.includes("supplier_code") && (
                                        <div>
                                            <Label htmlFor="supplier_code" className="mb-1">Код поставки</Label>
                                            <Input
                                                id="supplier_code"
                                                {...register("supplier_code")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("defect") && (
                                        <div>
                                            <Label htmlFor="defect" className="mb-1">Дефект</Label>
                                            <Input
                                                id="defect"
                                                {...register("defect")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("transmission") && (
                                        <div>
                                            <Label htmlFor="transmission-select" className="mb-1">Трансмиссия</Label>
                                            <div className="relative">
                                                <ClearableSelect
    value={watch("transmission") || ""}
    onValueChange={(value) => setValue("transmission", value)}
    placeholder="Выберите тип трансмиссии" id="transmission-select" className="h-10 w-full"
>
    {transmissionOptions.map((option) => (
                                                        <SelectItem key={option} value={option}>{option}</SelectItem>
                                                    ))}
</ClearableSelect>
                                                
                                            </div>
                                            <input
                                                type="hidden"
                                                {...register("transmission")}
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("transmission_model") && (
                                        <div>
                                            <Label htmlFor="transmission_model" className="mb-1">Модель трансмиссии</Label>
                                            <Input
                                                id="transmission_model"
                                                {...register("transmission_model")}
                                                type="text"
                                                placeholder="Введите номер трансмиссии"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("drive") && (
                                        <div>
                                            <Label htmlFor="drive-select" className="mb-1">Привод</Label>
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
                                            <input
                                                type="hidden"
                                                {...register("drive")}
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("wear_percentage") && (
                                        <div>
                                            <Label htmlFor="wear_percentage" className="mb-1">Процент износа (%)</Label>
                                            <Input
                                                id="wear_percentage"
                                                {...register("wear_percentage")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("season") && (
                                        <div>
                                            <Label htmlFor="season" className="mb-1">Сезон</Label>
                                            <Input
                                                id="season"
                                                {...register("season")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("diameter") && (
                                        <div>
                                            <Label htmlFor="diameter" className="mb-1">Диаметр</Label>
                                            <Input
                                                id="diameter"
                                                {...register("diameter")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("width") && (
                                        <div>
                                            <Label htmlFor="width" className="mb-1">Ширина</Label>
                                            <Input
                                                id="width"
                                                {...register("width")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("profile") && (
                                        <div>
                                            <Label htmlFor="profile" className="mb-1">Профиль</Label>
                                            <Input
                                                id="profile"
                                                {...register("profile")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("tire_quantity") && (
                                        <div>
                                            <Label htmlFor="tire_quantity" className="mb-1">Количество</Label>
                                            <Input
                                                id="tire_quantity"
                                                {...register("tire_quantity")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("drilling") && (
                                        <div>
                                            <Label htmlFor="drilling" className="mb-1">Сверловка</Label>
                                            <Input
                                                id="drilling"
                                                {...register("drilling")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("offset") && (
                                        <div>
                                            <Label htmlFor="offset" className="mb-1">Вылет</Label>
                                            <Input
                                                id="offset"
                                                {...register("offset")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("center_hole_diameter") && (
                                        <div>
                                            <Label htmlFor="center_hole_diameter" className="mb-1">Диаметр ЦО</Label>
                                            <Input
                                                id="center_hole_diameter"
                                                {...register("center_hole_diameter")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}

                                    {visibleFields.includes("tire_model") && (
                                        <div>
                                            <Label htmlFor="tire_model" className="mb-1">Модель шины</Label>
                                            <Input
                                                id="tire_model"
                                                {...register("tire_model")}
                                                type="text"
                                                className="h-10"
                                                autoComplete="off"
                                            />
                                        </div>
                                    )}
                                </div>
                            </div>
                        </TabsContent>
                    </form>
                </Tabs>
            </div>

            {showCropper && (
                <ImageEditor
                    src={tempImageSrc}
                    onEditComplete={handleCropComplete}
                    onCancel={handleCropCancel}
                    aspect={null} // Free aspect ratio for parts photos
                />
            )}
        </div>
    );
}
