/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useQuery } from '@tanstack/react-query';
import { statisticsApi } from '@/lib/api';

// Query keys
export const statisticsKeys = {
  all: ['statistics'] as const,
};

// Get statistics
export function useStatistics() {
  return useQuery({
    queryKey: statisticsKeys.all,
    queryFn: statisticsApi.get,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
