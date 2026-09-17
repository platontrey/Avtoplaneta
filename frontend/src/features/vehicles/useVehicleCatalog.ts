import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'

import type { SelectOption } from '@/components/ui/searchable-select'
import { API_BASE_URL } from '@/lib/api'

export interface VehicleModel {
  name: string
  slug?: string
}

export interface VehicleBrand {
  name: string
  slug?: string
  models: VehicleModel[]
}

// Справочник марок и моделей приходит с сервера (GET /api/vehicle-catalog) и
// является общим для веба и мобильного приложения. Списка марок в коде клиента
// быть не должно: он живёт в backend/parts-service/vehicles/vehicles.json и
// обновляется командой vehicle-catalog-sync.
export interface VehicleCatalog {
  version: string
  source?: string
  generated_at?: string
  brands: VehicleBrand[]
}

const fetchVehicleCatalog = async (): Promise<VehicleCatalog> => {
  const response = await fetch(`${API_BASE_URL}/api/vehicle-catalog`, {
    credentials: 'include',
  })
  if (!response.ok) {
    throw new Error(`Не удалось загрузить справочник марок: ${response.status}`)
  }
  return response.json() as Promise<VehicleCatalog>
}

export const useVehicleCatalog = () =>
  useQuery({
    queryKey: ['vehicle-catalog'],
    queryFn: fetchVehicleCatalog,
    staleTime: 5 * 60 * 1000,
    gcTime: 24 * 60 * 60 * 1000,
  })

const toOptions = (values: { name: string }[]): SelectOption[] =>
  values.map((value) => ({ value: value.name, label: value.name }))

/**
 * Связанные списки «марка → модель».
 *
 * `brand` — уже выбранная марка; модели фильтруются по ней без учёта регистра,
 * потому что в ранее заведённых запчастях марка лежит свободным текстом.
 * Если марки нет в справочнике или у неё ещё нет моделей, `modelOptions` пуст —
 * поле модели в этом случае должно остаться обычным вводом, чтобы редкие
 * машины по-прежнему можно было завести.
 */
export const useVehicleOptions = (brand?: string) => {
  const { data, isLoading, error } = useVehicleCatalog()

  return useMemo(() => {
    const brands = data?.brands ?? []
    const needle = brand?.trim().toLowerCase() ?? ''
    const selected = needle
      ? brands.find((item) => item.name.trim().toLowerCase() === needle)
      : undefined

    return {
      isLoading,
      error,
      version: data?.version,
      brandOptions: toOptions(brands),
      modelOptions: toOptions(selected?.models ?? []),
      knownBrand: Boolean(selected),
    }
  }, [data, brand, isLoading, error])
}
