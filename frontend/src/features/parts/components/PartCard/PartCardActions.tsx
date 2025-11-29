/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { motion } from "framer-motion";
import { Edit, Trash2 } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import type { Part } from "../../types";

interface PartCardActionsProps {
  part: Part;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function PartCardActions({ part, onEdit, onDelete }: PartCardActionsProps) {
  return (
    <div className="flex space-x-1">
      {/* Edit button */}
      <motion.div
        className="opacity-100 transition-opacity p-1 rounded hover:bg-accent"
        onClick={(e) => {
          console.log('Edit button clicked for part:', part.id);
          onEdit?.();
          e.stopPropagation();
        }}
        style={{ cursor: 'pointer' }}
        whileHover={{ scale: 1.1, rotate: 10 }}
        whileTap={{ scale: 0.9 }}
        transition={{ duration: 0.2 }}
      >
        <Edit className="h-4 w-4" />
      </motion.div>

      {/* Delete confirmation dialog */}
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <motion.div
            className="opacity-100 transition-opacity p-1 rounded hover:bg-accent"
            onClick={(e) => e.stopPropagation()}
            whileHover={{ scale: 1.1, rotate: -10 }}
            whileTap={{ scale: 0.9 }}
            transition={{ duration: 0.2 }}
          >
            <Trash2 className={`h-4 w-4 text-destructive hover:text-destructive/90`} />
          </motion.div>
        </AlertDialogTrigger>
        <AlertDialogContent onClick={(e) => e.stopPropagation()}>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the
              part and remove its data from our inventory.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => onDelete?.()}>
              Continue
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}