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

export interface CatalogReportBinding {
  source: string
  target: string
  transform?: string
  categories?: string[]
  excluded_categories?: string[]
  default_value?: string
}

export interface CatalogPartTemplate {
  id: string
  name: string
  category: string
  quantity: number
  price: number
  defaults?: Record<string, string>
}

export interface PartCatalog {
  version: string
  attributes: CatalogAttribute[]
  part_form_categories: CatalogPartFormCategory[]
  report_bindings: CatalogReportBinding[]
  parts: CatalogPartTemplate[]
}

export type CatalogPartView = Omit<CatalogPartTemplate, 'defaults'> &
  Record<string, string | number | undefined>

const fetchPartCatalog = async (): Promise<PartCatalog> => {
  const response = await fetch(`${API_BASE_URL}/api/part-catalog`, {
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

export const flattenCatalogParts = (catalog?: PartCatalog): CatalogPartView[] =>
  (catalog?.parts ?? []).map(({ defaults, ...part }) => ({
    ...part,
    ...defaults,
  }))

export const reportBindingCategories = (catalog: PartCatalog | undefined, source: string) =>
  new Set(catalog?.report_bindings.find((binding) => binding.source === source)?.categories ?? [])
