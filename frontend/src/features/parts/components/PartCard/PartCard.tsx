/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { motion } from "framer-motion";
import type { Part } from "../../types";
import { PartCardHeader } from "./PartCardHeader.tsx";
import { PartCardContent } from "./PartCardContent.tsx";
import { PartCardActions } from "./PartCardActions.tsx";

interface PartCardProps {
  part: Part;
  isLoading?: boolean;
  onEdit?: () => void;
  onDelete?: () => void;
}

function PartCard({ part, isLoading = false, onEdit, onDelete }: PartCardProps) {
  if (isLoading) {
    return <PartCardSkeleton />;
  }

  return (
    <motion.div
      className="border rounded-lg mb-2 relative group hover:bg-accent hover:border-accent-foreground/20 transition-colors"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: "easeOut" }}
      whileHover={{ scale: 1.02 }}
      layout
    >
      <AccordionItem value={`part-${part.id}`}>
        <AccordionTrigger className="flex items-center justify-between w-full hover:bg-accent hover:text-accent-foreground pr-4">
          <PartCardHeader part={part} />
          <PartCardActions part={part} onEdit={onEdit} onDelete={onDelete} />
        </AccordionTrigger>

        <motion.div
          initial={false}
          animate={{ height: "auto", opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.3, ease: "easeInOut" }}
        >
          <AccordionContent className="hover:bg-accent/50 border-t border-gray-200 pt-3">
            <PartCardContent part={part} />
          </AccordionContent>
        </motion.div>
      </AccordionItem>
    </motion.div>
  );
}

function PartCardSkeleton() {
  return (
    <motion.div
      className="border rounded-lg mb-2 p-4"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <div className="flex items-center space-x-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.1, duration: 0.3 }}
        >
          <div className="h-4 w-48 bg-gray-200 rounded" />
        </motion.div>
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2, duration: 0.3 }}
        >
          <div className="h-8 w-8 bg-gray-200 rounded" />
        </motion.div>
      </div>
      <motion.div
        className="mt-4 space-y-2"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.3, duration: 0.4 }}
      >
        {[32, 40, 36, 28, 30, 34, 26].map((width, index) => (
          <motion.div
            key={index}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.4 + index * 0.1, duration: 0.3 }}
          >
            <div className={`h-4 w-${width} bg-gray-200 rounded`} />
          </motion.div>
        ))}
      </motion.div>
    </motion.div>
  );
}

export { PartCard };
