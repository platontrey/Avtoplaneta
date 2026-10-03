/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';

export interface ActiveFilterItem {
  id: string;
  label: string;
  onClear: () => void;
}

interface ActiveFilterBadgesProps {
  filters: ActiveFilterItem[];
}

export const ActiveFilterBadges: React.FC<ActiveFilterBadgesProps> = ({ filters }) => {
  if (filters.length === 0) return null;

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key="active-filters"
        layout
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.3, ease: 'easeOut' }}
        className="flex flex-wrap gap-2"
      >
        <AnimatePresence>
          {filters.map((filter) => (
            <motion.div
              key={filter.id}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              transition={{ duration: 0.2 }}
            >
              <Badge variant="secondary" className="gap-2 bg-transparent border border-gray-300">
                <Checkbox
                  checked={true}
                  onCheckedChange={(checked) => {
                    if (!checked) filter.onClear();
                  }}
                  className="h-3 w-3"
                />
                <span>{filter.label}</span>
              </Badge>
            </motion.div>
          ))}
        </AnimatePresence>
      </motion.div>
    </AnimatePresence>
  );
};
