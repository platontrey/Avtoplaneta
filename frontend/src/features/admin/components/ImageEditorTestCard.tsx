/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Image } from 'lucide-react';
import ImageEditor from '@/components/ImageEditor';

export const ImageEditorTestCard: React.FC = () => {
  const [showImageEditorTest, setShowImageEditorTest] = useState(false);
  const [testImageSrc, setTestImageSrc] = useState<string>('');

  const handleTestImageEditComplete = (blob: Blob) => {
    console.log('Test image edit completed, blob size:', blob.size);
    alert(`Тестовое изображение обработано! Размер: ${blob.size} байт`);
    setShowImageEditorTest(false);
  };

  const handleTestImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const url = URL.createObjectURL(file);
      setTestImageSrc(url);
      setShowImageEditorTest(true);
    }
  };

  return (
    <>
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center">
            <Image className="w-5 h-5 mr-2" />
            Тест ImageEditor (Pintura)
          </CardTitle>
          <CardDescription>
            Тестирование компонента редактирования изображений
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Label htmlFor="test-image">Выберите тестовое изображение</Label>
              <Input
                id="test-image"
                type="file"
                accept="image/*"
                onChange={handleTestImageSelect}
                className="mt-2"
              />
            </div>
            <p className="text-sm text-gray-600">
              Выберите изображение, чтобы протестировать работу ImageEditor с Pintura.
            </p>
          </div>
        </CardContent>
      </Card>

      {showImageEditorTest && testImageSrc && (
        <ImageEditor
          src={testImageSrc}
          onEditComplete={handleTestImageEditComplete}
          onCancel={() => setShowImageEditorTest(false)}
          aspect={null}
        />
      )}
    </>
  );
};
