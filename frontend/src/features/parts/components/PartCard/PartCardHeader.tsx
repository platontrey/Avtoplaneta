/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { motion, AnimatePresence } from "framer-motion";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { LazyLoadImage } from "react-lazy-load-image-component";
import "react-lazy-load-image-component/src/effects/blur.css";
import type { Part } from "../../types";
import { API_BASE_URL } from '@/lib/api';

interface PartCardHeaderProps {
  part: Part;
}

export function PartCardHeader({ part }: PartCardHeaderProps) {
  return (
    <div className="flex items-center space-x-3 flex-1">
      <AnimatePresence>
        {part.photo && (
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            transition={{ duration: 0.3 }}
          >
            <Dialog>
              <DialogTrigger asChild>
                <motion.div
                  className="cursor-pointer"
                  whileHover={{ scale: 1.1, rotate: 5 }}
                  whileTap={{ scale: 0.95 }}
                  transition={{ duration: 0.2 }}
                >
                  <LazyLoadImage
                    src={`${API_BASE_URL}${part.photo}?t=${Date.now()}`}
                    alt={part.name}
                    className="w-12 h-12 object-cover rounded border"
                    effect="blur"
                    placeholderSrc="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHJlY3Qgd2lkdGg9IjQ4IiBoZWlnaHQ9IjQ4IiBmaWxsPSIjRjNGNEY2Ii8+Cjx0ZXh0IHg9IjI0IiB5PSIyNCIgZm9udC1mYW1pbHk9IkFyaWFsLCBzYW5zLXNlcmlmIiBmb250LXNpemU9IjEwIiBmaWxsPSIjOUI5QkE0IiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBkeT0iMC4zZW0iPkxvYWRpbmcuLi48L3RleHQ+Cjwvc3ZnPg=="
                  />
                </motion.div>
              </DialogTrigger>
              <DialogContent className="max-w-4xl">
                <LazyLoadImage
                  src={`${API_BASE_URL}${part.photo}?t=${Date.now()}`}
                  alt={part.name}
                  className="w-full h-auto max-h-[80vh] object-contain"
                  effect="blur"
                  placeholderSrc="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxMDAiIGhlaWdodD0iMTAwIiBmaWxsPSIjRjNGNEY2Ii8+Cjx0ZXh0IHg9IjUwIiB5PSI1MCIgZm9udC1mYW1pbHk9IkFyaWFsLCBzYW5zLXNlcmlmIiBmb250LXNpemU9IjEyIiBmaWxsPSIjOUI5QkE0IiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBkeT0iMC4zZW0iPkxvYWRpbmcuLi48L3RleHQ+Cjwvc3ZnPg=="
                />
              </DialogContent>
            </Dialog>
          </motion.div>
        )}
      </AnimatePresence>
      <div className="flex flex-col">
        <span>{part.name || 'Unnamed Part'}</span>
        {part.category && (
          <span className="text-sm text-muted-foreground">{part.category}</span>
        )}
      </div>
    </div>
  );
}