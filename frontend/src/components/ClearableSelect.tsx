import React from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";

interface Option {
    value: string;
    label: string;
}

interface ClearableSelectProps {
    value?: string;
    onValueChange: (val: string) => void;
    options?: Option[];
    placeholder?: string;
    id?: string;
    className?: string;
    children?: React.ReactNode; // Optional children if they prefer to pass SelectItem manually
}

export function ClearableSelect({
    value,
    onValueChange,
    options,
    placeholder = "Выберите значение",
    id,
    className,
    children,
}: ClearableSelectProps) {
    return (
        <div className="relative w-full">
            <Select value={value || ""} onValueChange={onValueChange}>
                <SelectTrigger id={id} className={cn("w-full", className)}>
                    <SelectValue placeholder={placeholder} />
                </SelectTrigger>
                <SelectContent>
                    {children ? children : (
                        options
                            ?.filter((opt) => Boolean(opt.value))
                            .map(opt => (
                                <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                </SelectItem>
                            ))
                    )}
                </SelectContent>
            </Select>
            {value && (
                <button
                    type="button"
                    onClick={(e) => {
                        e.stopPropagation();
                        onValueChange("");
                    }}
                    className="absolute right-8 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1 transition-opacity duration-200 z-10"
                    title="Очистить"
                >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            )}
        </div>
    );
}
