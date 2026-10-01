import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'

import type { SelectOption } from '@/components/ui/searchable-select'
import { API_BASE_URL } from '@/lib/api'

export interface VehicleModel {
  name: string
  slug?: string
  bodies?: string[]
  engines?: string[]
}

export interface VehicleBrand {
  name: string
  slug?: string
  bodies?: string[]
  engines?: string[]
  models: VehicleModel[]
}

// Справочник марок, моделей, кузовов и двигателей приходит с сервера (GET /api/vehicle-catalog)
// и является общим для веба и мобильного приложения. Списка марок в коде клиента
// быть не должно: он живёт в backend/parts-service/vehicles/vehicles.json и
// обновляется командой vehicle-catalog-sync.
export interface VehicleCatalog {
  version: string
  source?: string
  generated_at?: string
  brands: VehicleBrand[]
}

const fetchVehicleCatalog = async (): Promise<VehicleCatalog> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/vehicle-catalog`, {
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

const toOptionsFromStrings = (values?: string[]): SelectOption[] =>
  (values ?? []).map((value) => ({ value, label: value }))

/**
 * Связанные списки «марка → модель → марка кузова / марка двигателя».
 *
 * `brand` — уже выбранная марка.
 * `model` — выбранная модель (опционально).
 * Модели фильтруются по марке без учёта регистра.
 * Кузова и двигатели берутся из модели (если указана и есть свои данные),
 * либо из марки как общий список бренда.
 */
export const useVehicleOptions = (brand?: string, model?: string) => {
  const { data, isLoading, error } = useVehicleCatalog()

  return useMemo(() => {
    const brands = data?.brands ?? []
    const brandNeedle = brand?.trim().toLowerCase() ?? ''
    const selectedBrand = brandNeedle
      ? brands.find((item) => item.name.trim().toLowerCase() === brandNeedle)
      : undefined

    const models = selectedBrand?.models ?? []
    const modelNeedle = model?.trim().toLowerCase() ?? ''
    const selectedModel = modelNeedle
      ? models.find((item) => item.name.trim().toLowerCase() === modelNeedle)
      : undefined

    const bodies =
      selectedModel?.bodies && selectedModel.bodies.length > 0
        ? selectedModel.bodies
        : (selectedBrand?.bodies ?? [])

    const engines =
      selectedModel?.engines && selectedModel.engines.length > 0
        ? selectedModel.engines
        : (selectedBrand?.engines ?? [])

    return {
      isLoading,
      error,
      version: data?.version,
      brandOptions: toOptions(brands),
      modelOptions: toOptions(models),
      bodyOptions: toOptionsFromStrings(bodies),
      engineOptions: toOptionsFromStrings(engines),
      knownBrand: Boolean(selectedBrand),
      knownModel: Boolean(selectedModel),
    }
  }, [data, brand, model, isLoading, error])
}
