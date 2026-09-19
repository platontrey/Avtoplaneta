import { useQuery } from '@tanstack/react-query'

import { API_BASE_URL } from '@/lib/api'

export interface CatalogAttribute {
  code: string
  label: string
  input_type: 'text' | 'number' | 'select'
  options?: string[]
  sort_order: number
}

export interface CatalogPartFormCategory {
  code: string
  name: string
  attributes: string[]
}

// Клиенты получают только метаданные формы. Шаблоны и report_bindings
// остаются серверной деталью имплементации.
export interface PartCatalog {
  version: string
  attributes: CatalogAttribute[]
  part_form_categories: CatalogPartFormCategory[]
}

const fetchPartCatalog = async (): Promise<PartCatalog> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/part-catalog`, {
    credentials: 'include',
  })
  if (!response.ok) {
    throw new Error(`Не удалось загрузить каталог запчастей: ${response.status}`)
  }
  return response.json() as Promise<PartCatalog>
}

export const usePartCatalog = () =>
  useQuery({
    queryKey: ['part-catalog'],
    queryFn: fetchPartCatalog,
    staleTime: 5 * 60 * 1000,
    gcTime: 24 * 60 * 60 * 1000,
  })
