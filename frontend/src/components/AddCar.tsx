/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Link } from 'react-router-dom';
import { ArrowLeft, FileText, Plus } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function AddCar() {
  return (
    <div className="max-w-4xl mx-auto p-4 sm:p-8">
      <Link
        to="/inventory"
        className="flex items-center gap-1 text-gray-600 hover:text-black mb-6 text-sm sm:text-base"
      >
        <ArrowLeft size={16} />
        Назад к инвентарю
      </Link>

      <h2 className="text-2xl sm:text-3xl font-semibold mb-4">Добавить автомобиль</h2>
      <p className="text-gray-500 mb-8 text-base sm:text-lg">
        Выберите действие для автомобиля.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="cursor-pointer hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="w-5 h-5" />
              Дефектная ведомость
            </CardTitle>
            <CardDescription>
              Создать дефектную ведомость для автомобиля
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild className="w-full">
              <Link to="/defect-report">
                Создать дефектную ведомость
              </Link>
            </Button>
          </CardContent>
        </Card>

        <Card className="cursor-pointer hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Plus className="w-5 h-5" />
              Добавить запчасть
            </CardTitle>
            <CardDescription>
              Добавить новую запчасть в инвентарь
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant="outline" className="w-full">
              <Link to="/add-part">
                Добавить запчасть
              </Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}