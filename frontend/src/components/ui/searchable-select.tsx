import * as React from "react"
import * as SelectPrimitive from "@radix-ui/react-select"
import { CheckIcon, ChevronDownIcon, ChevronUpIcon } from "lucide-react"

import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"

function Select({
                    ...props
                }: React.ComponentProps<typeof SelectPrimitive.Root>) {
    return <SelectPrimitive.Root data-slot="select" {...props} />
}

function SelectGroup({
                         ...props
                     }: React.ComponentProps<typeof SelectPrimitive.Group>) {
    return <SelectPrimitive.Group data-slot="select-group" {...props} />
}

function SelectValue({
                         ...props
                     }: React.ComponentProps<typeof SelectPrimitive.Value>) {
    return <SelectPrimitive.Value data-slot="select-value" {...props} />
}

function SelectTrigger({
                           className,
                           size = "default",
                           children,
                           ...props
                       }: React.ComponentProps<typeof SelectPrimitive.Trigger> & {
    size?: "sm" | "default"
}) {
    return (
        <SelectPrimitive.Trigger
            data-slot="select-trigger"
            data-size={size}
            className={cn(
                "border-input data-[placeholder]:text-muted-foreground [&_svg:not([class*='text-'])]:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 dark:hover:bg-input/50 flex w-fit items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm whitespace-nowrap shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 data-[size=default]:h-9 data-[size=sm]:h-8 *:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:flex *:data-[slot=select-value]:items-center *:data-[slot=select-value]:gap-2 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
                className
            )}
            {...props}
        >
            {children}
            <SelectPrimitive.Icon asChild>
                <ChevronDownIcon className="size-4 opacity-50" />
            </SelectPrimitive.Icon>
        </SelectPrimitive.Trigger>
    )
}

function SelectContent({
                           className,
                           children,
                           position = "popper",
                           align = "center",
                           ...props
                       }: React.ComponentProps<typeof SelectPrimitive.Content>) {
    return (
        <SelectPrimitive.Portal>
            <SelectPrimitive.Content
                data-slot="select-content"
                translate="no"
                className={cn(
                    "notranslate bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 relative z-50 max-h-(--radix-select-content-available-height) min-w-[8rem] origin-(--radix-select-content-transform-origin) overflow-x-hidden overflow-y-auto rounded-md border shadow-md",
                    position === "popper" &&
                    "data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
                    className
                )}
                position={position}
                align={align}
                {...props}
            >
                <SelectScrollUpButton />
                <SelectPrimitive.Viewport
                    className={cn(
                        "p-1",
                        position === "popper" &&
                        "h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)] scroll-my-1"
                    )}
                >
                    {children}
                </SelectPrimitive.Viewport>
                <SelectScrollDownButton />
            </SelectPrimitive.Content>
        </SelectPrimitive.Portal>
    )
}

function SelectLabel({
                         className,
                         ...props
                     }: React.ComponentProps<typeof SelectPrimitive.Label>) {
    return (
        <SelectPrimitive.Label
            data-slot="select-label"
            className={cn("text-muted-foreground px-2 py-1.5 text-xs", className)}
            {...props}
        />
    )
}

function SelectItem({
                        className,
                        children,
                        value,
                        ...props
                    }: React.ComponentProps<typeof SelectPrimitive.Item>) {
    if (!value) {
        return null;
    }
    return (
        <SelectPrimitive.Item
            data-slot="select-item"
            value={value}
            className={cn(
                "focus:bg-accent focus:text-accent-foreground [&_svg:not([class*='text-'])]:text-muted-foreground relative flex w-full cursor-default items-center gap-2 rounded-sm py-1.5 pr-8 pl-2 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 *:[span]:last:flex *:[span]:last:items-center *:[span]:last:gap-2",
                className
            )}
            {...props}
        >
      <span className="absolute right-2 flex size-3.5 items-center justify-center">
        <SelectPrimitive.ItemIndicator>
          <CheckIcon className="size-4" />
        </SelectPrimitive.ItemIndicator>
      </span>
            <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
        </SelectPrimitive.Item>
    )
}

function SelectSeparator({
                             className,
                             ...props
                         }: React.ComponentProps<typeof SelectPrimitive.Separator>) {
    return (
        <SelectPrimitive.Separator
            data-slot="select-separator"
            className={cn("bg-border pointer-events-none -mx-1 my-1 h-px", className)}
            {...props}
        />
    )
}

function SelectScrollUpButton({
                                  className,
                                  ...props
                              }: React.ComponentProps<typeof SelectPrimitive.ScrollUpButton>) {
    return (
        <SelectPrimitive.ScrollUpButton
            data-slot="select-scroll-up-button"
            className={cn(
                "flex cursor-default items-center justify-center py-1",
                className
            )}
            {...props}
        >
            <ChevronUpIcon className="size-4" />
        </SelectPrimitive.ScrollUpButton>
    )
}

function SelectScrollDownButton({
                                    className,
                                    ...props
                                }: React.ComponentProps<typeof SelectPrimitive.ScrollDownButton>) {
    return (
        <SelectPrimitive.ScrollDownButton
            data-slot="select-scroll-down-button"
            className={cn(
                "flex cursor-default items-center justify-center py-1",
                className
            )}
            {...props}
        >
            <ChevronDownIcon className="size-4" />
        </SelectPrimitive.ScrollDownButton>
    )
}

export interface SelectOption {
    value: string
    label: string
}

interface SearchableSelectProps {
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

function SearchableSelect({
    value,
    onValueChange,
    options,
    placeholder = "Выберите значение...",
    searchPlaceholder = "Поиск...",
    emptyMessage = "Ничего не найдено.",
    disabled = false,
    className,
    clearable = true,
    allowCustom = false,
}: SearchableSelectProps) {
    const [search, setSearch] = React.useState("")
    const [open, setOpen] = React.useState(false)
    const inputRef = React.useRef<HTMLInputElement>(null)

    const filteredOptions = options
        .filter((option) => Boolean(option.value))
        .filter((option) =>
            option.label.toLowerCase().includes(search.toLowerCase())
        )

    const selectedOption = options.find((option) => option.value === value)
    const trimmedSearch = search.trim()
    const hasExactMatch = options.some(
        (o) => o.value.toLowerCase() === trimmedSearch.toLowerCase() || o.label.toLowerCase() === trimmedSearch.toLowerCase()
    )
    const canAddCustom = allowCustom && trimmedSearch !== "" && !hasExactMatch

    const handleSelectCustom = (customVal: string) => {
        onValueChange(customVal)
        setSearch("")
        setOpen(false)
    }

    React.useEffect(() => {
        if (open && inputRef.current) {
            inputRef.current.focus()
        }
    }, [open])

    return (
        <div className="relative w-full">
            <Select value={value} onValueChange={onValueChange} disabled={disabled} open={open} onOpenChange={setOpen}>
                <SelectTrigger className={cn("flex h-10 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 text-foreground font-normal whitespace-pre-wrap", className)}>
                    {selectedOption ? selectedOption.label : (value || placeholder)}
                </SelectTrigger>
                <SelectContent>
                    <div className="p-2">
                        <Input
                            ref={inputRef}
                            placeholder={searchPlaceholder}
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            onKeyDown={(e) => {
                                e.stopPropagation()
                                if (e.key === "Enter" && canAddCustom) {
                                    e.preventDefault()
                                    handleSelectCustom(trimmedSearch)
                                }
                            }}
                            className="text-foreground border-input"
                        />
                    </div>
                    {canAddCustom && (
                        <SelectItem key={`custom-${trimmedSearch}`} value={trimmedSearch} className="font-semibold text-blue-600 dark:text-blue-400">
                            + Использовать "{trimmedSearch}"
                        </SelectItem>
                    )}
                    {value && !options.some((o) => o.value === value) && value !== trimmedSearch && (
                        <SelectItem key={`current-custom-${value}`} value={value} className="font-semibold text-blue-600 dark:text-blue-400">
                            {value}
                        </SelectItem>
                    )}
                    {filteredOptions.length === 0 && !canAddCustom && (!value || options.some(o => o.value === value)) ? (
                        <div className="py-2 px-3 text-sm text-muted-foreground">{emptyMessage}</div>
                    ) : (
                        filteredOptions.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                                {option.label}
                            </SelectItem>
                        ))
                    )}
                </SelectContent>
            </Select>
            {clearable && value && (
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
    )
}

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
    SearchableSelect,
}
