import React from "react";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface FormRowProps {
  label: string;
  htmlFor?: string;
  children: React.ReactNode;
  className?: string;
  labelClassName?: string;
  contentClassName?: string;
}

export default function FormRow({
  label,
  htmlFor,
  children,
  className,
  labelClassName,
  contentClassName,
}: FormRowProps) {
  return (
    <div className={cn("grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4", className)}>
      <Label htmlFor={htmlFor} className={cn("sm:text-right text-sm", labelClassName)}>
        {label}
      </Label>
      <div className={cn("sm:col-span-3", contentClassName)}>
        {children}
      </div>
    </div>
  );
}
