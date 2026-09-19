import { getAuthHeaders } from "@/lib/csrf";
import { API_BASE_URL } from "@/lib/api";

export interface DefectReportPayload {
  brand: string;
  model: string;
  year: number;
  car_release_period?: string;
  vin?: string;
  mileage: number;
  engine_brand?: string;
  body_brand?: string;
  interior_color?: string;
  body_color?: string;
  transmission?: string;
  transmission_model?: string;
  drive?: string;
  description?: string;
  catalog_version?: string;
}

export interface DefectReportPreviewPart {
  name: string;
  category: string;
  description: string;
  quantity: number;
  price: number;
  [key: string]: string | number | undefined;
}

export interface DefectReportPreviewResponse {
  catalog_version: string;
  total: number;
  parts: DefectReportPreviewPart[];
}

const parseError = async (response: Response) => {
  const text = await response.text().catch(() => "");
  return text || `Ошибка сервера: ${response.status}`;
};

export const previewDefectReport = async (
  payload: DefectReportPayload,
  signal?: AbortSignal,
): Promise<DefectReportPreviewResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/v1/defect-reports/preview`, {
    method: "POST",
    headers: getAuthHeaders(),
    credentials: "include",
    body: JSON.stringify(payload),
    signal,
  });
  if (!response.ok) throw new Error(await parseError(response));
  return response.json() as Promise<DefectReportPreviewResponse>;
};

export const createDefectReport = async (payload: DefectReportPayload) => {
  const response = await fetch(`${API_BASE_URL}/api/v1/defect-reports`, {
    method: "POST",
    headers: getAuthHeaders(),
    credentials: "include",
    body: JSON.stringify(payload),
  });
  if (!response.ok) throw new Error(await parseError(response));
  return response.json() as Promise<{ message?: string }>;
};
