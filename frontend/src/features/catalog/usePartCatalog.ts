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

export type DefectReportValues = Record<string, string | number | null | undefined>

// Разворачиваем каталог на клиенте для совместимости с уже запущенными версиями
// parts-service. Сервер также умеет это делать, но selectedParts гарантирует, что
// характеристики шаблонов не потеряются между формой и очередью обработки.
export const expandDefectReportParts = (
  catalog: PartCatalog,
  values: DefectReportValues,
  supplierCode = Date.now().toString(),
) => catalog.parts.map((template) => {
  const specifications: Record<string, string> = {
    ...(template.defaults ?? {}),
    supplier_code: supplierCode,
  }

  for (const binding of catalog.report_bindings) {
    const included = !binding.categories?.length || binding.categories.includes(template.category)
    const excluded = binding.excluded_categories?.includes(template.category) ?? false
    if (!included || excluded) continue

    const rawValue = values[binding.source]
    const value = rawValue == null || String(rawValue).trim() === ''
      ? binding.default_value
      : String(rawValue).trim()
    if (value) specifications[binding.target] = value
  }

  return {
    name: template.name,
    category: template.category,
    description: '',
    quantity: template.quantity,
    price: template.price,
    ...specifications,
  }
})
