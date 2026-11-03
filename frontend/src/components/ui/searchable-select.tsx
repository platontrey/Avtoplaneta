/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import * as React from "react"
import { CheckIcon, ChevronDownIcon, PlusIcon, X } from "lucide-react"

import { cn } from "@/lib/utils"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"

export {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectScrollDownButton,
  SelectScrollUpButton,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export interface SelectOption {
  value: string
  label: string
}

export interface SearchableSelectProps {
  value?: string
  onValueChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  searchPlaceholder?: string
  emptyMessage?: string
  disabled?: boolean
  className?: string
  clearable?: boolean
  allowCustom?: boolean
}

export function SearchableSelect({
  value,
  onValueChange,
  options = [],
  placeholder = "Выберите значение...",
  searchPlaceholder = "Поиск...",
  emptyMessage = "Ничего не найдено.",
  disabled = false,
  className,
  clearable = true,
  allowCustom = false,
}: SearchableSelectProps) {
  const [open, setOpen] = React.useState(false)
  const [search, setSearch] = React.useState("")

  const trimmedSearch = search.trim()

  // Быстрая фильтрация опций в памяти по названию и значению
  const filteredOptions = React.useMemo(() => {
    if (!options || options.length === 0) return []
    const s = trimmedSearch.toLowerCase()
    const valid = options.filter((option) => Boolean(option && option.value))
    if (!s) return valid
    return valid.filter(
      (option) =>
        option.label.toLowerCase().includes(s) ||
        option.value.toLowerCase().includes(s)
    )
  }, [options, trimmedSearch])

  // Ограничиваем рендер в DOM первыми 60 элементами для мгновенного появления дропдауна (60 FPS)
  const displayedOptions = React.useMemo(() => {
    return filteredOptions.slice(0, 60)
  }, [filteredOptions])

  const selectedOption = React.useMemo(() => {
    if (!value) return null
    return options.find((option) => option.value === value) || null
  }, [options, value])

  const hasExactMatch = React.useMemo(() => {
    if (!trimmedSearch) return false
    const s = trimmedSearch.toLowerCase()
    return options.some(
      (o) => o.value.toLowerCase() === s || o.label.toLowerCase() === s
    )
  }, [options, trimmedSearch])

  const canAddCustom = allowCustom && trimmedSearch !== "" && !hasExactMatch

  const handleSelectCustom = (customVal: string) => {
    onValueChange(customVal)
    setSearch("")
    setOpen(false)
  }

  const handleSelectOption = (optVal: string) => {
    onValueChange(optVal)
    setSearch("")
    setOpen(false)
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setSearch("")
    }
    setOpen(nextOpen)
  }

  const isCustomSelected = Boolean(value && !options.some((o) => o.value === value))

  return (
    <div className="relative w-full">
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <button
            type="button"
            role="combobox"
            aria-expanded={open}
            disabled={disabled}
            className={cn(
              "border-input data-[placeholder]:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 dark:hover:bg-input/50 flex h-10 w-full items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 text-foreground font-normal text-left",
              clearable && Boolean(value) ? "pr-14" : "pr-8",
              !selectedOption && !value && "text-muted-foreground",
              className
            )}
          >
            <span className="truncate">
              {selectedOption ? selectedOption.label : (value || placeholder)}
            </span>
            <ChevronDownIcon className="size-4 shrink-0 opacity-50 ml-auto pointer-events-none" />
          </button>
        </PopoverTrigger>

        <PopoverContent
          className="w-[var(--radix-popover-trigger-width)] min-w-[220px] p-0 shadow-lg"
          align="start"
          sideOffset={4}
        >
          <Command shouldFilter={false} className="w-full">
            <CommandInput
              placeholder={searchPlaceholder}
              value={search}
              onValueChange={setSearch}
              onKeyDown={(e) => {
                if (e.key === "Enter" && canAddCustom) {
                  e.preventDefault()
                  e.stopPropagation()
                  handleSelectCustom(trimmedSearch)
                }
              }}
            />
            <CommandList className="max-h-[280px]">
              {canAddCustom && (
                <CommandGroup heading="Новое значение">
                  <CommandItem
                    value={`custom-${trimmedSearch}`}
                    onSelect={() => handleSelectCustom(trimmedSearch)}
                    className="font-medium text-blue-600 dark:text-blue-400 cursor-pointer"
                  >
                    <PlusIcon className="size-4 mr-1 shrink-0" />
                    <span>Использовать &laquo;{trimmedSearch}&raquo;</span>
                  </CommandItem>
                </CommandGroup>
              )}

              {isCustomSelected && value !== trimmedSearch && (
                <CommandGroup heading="Текущее значение">
                  <CommandItem
                    value={`current-${value}`}
                    onSelect={() => {
                      setOpen(false)
                    }}
                    className="font-medium text-blue-600 dark:text-blue-400 cursor-pointer flex items-center justify-between"
                  >
                    <span className="truncate">{value}</span>
                    <CheckIcon className="size-4 text-blue-600 dark:text-blue-400 ml-2 shrink-0" />
                  </CommandItem>
                </CommandGroup>
              )}

              {displayedOptions.length === 0 && !canAddCustom && !isCustomSelected ? (
                <div className="py-6 text-center text-sm text-muted-foreground">
                  {emptyMessage}
                </div>
              ) : (
                <CommandGroup>
                  {displayedOptions.map((option) => {
                    const isSelected = value === option.value
                    return (
                      <CommandItem
                        key={option.value}
                        value={option.value}
                        onSelect={() => handleSelectOption(option.value)}
                        className="cursor-pointer flex items-center justify-between py-2"
                      >
                        <span className="truncate">{option.label}</span>
                        {isSelected && (
                          <CheckIcon className="size-4 text-blue-600 dark:text-blue-400 ml-2 shrink-0" />
                        )}
                      </CommandItem>
                    )
                  })}
                </CommandGroup>
              )}

              {filteredOptions.length > displayedOptions.length && (
                <div className="py-2 px-3 text-[11px] text-muted-foreground text-center border-t bg-muted/30">
                  Показано {displayedOptions.length} из {filteredOptions.length}. Уточните поиск для сужения.
                </div>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {clearable && Boolean(value) && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            e.preventDefault()
            onValueChange("")
          }}
          className="absolute right-8 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground p-1 transition-colors duration-150 z-10"
          title="Очистить"
        >
          <X className="size-4" />
        </button>
      )}
    </div>
  )
}
