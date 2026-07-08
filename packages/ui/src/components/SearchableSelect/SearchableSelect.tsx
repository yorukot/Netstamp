import * as PopoverPrimitive from "@radix-ui/react-popover";
import type { ComponentPropsWithoutRef, KeyboardEvent, ReactNode } from "react";
import { Children, isValidElement, useEffect, useId, useMemo, useRef, useState } from "react";
import styles from "./SearchableSelect.module.css";

export type SearchableSelectSize = "sm" | "md";

export interface SearchableSelectOption {
	value: string;
	label: ReactNode;
	description?: ReactNode;
	searchText?: string;
	disabled?: boolean;
}

export interface SearchableSelectProps extends Omit<ComponentPropsWithoutRef<"button">, "children" | "defaultValue" | "onChange" | "type" | "value"> {
	options: SearchableSelectOption[];
	value: string;
	onValueChange: (value: string) => void;
	placeholder?: ReactNode;
	searchPlaceholder?: string;
	emptyLabel?: ReactNode;
	invalid?: boolean;
	size?: SearchableSelectSize;
	frameClassName?: string;
	menuClassName?: string;
	name?: string;
	form?: string;
	required?: boolean;
}

function textFromReactNode(node: ReactNode): string {
	return Children.toArray(node)
		.map(child => {
			if (typeof child === "string" || typeof child === "number") {
				return String(child);
			}

			if (isValidElement(child)) {
				return textFromReactNode((child.props as { children?: ReactNode }).children);
			}

			return "";
		})
		.join("");
}

function optionSearchText(option: SearchableSelectOption) {
	return [option.searchText, textFromReactNode(option.label), textFromReactNode(option.description)].filter(Boolean).join(" ").toLocaleLowerCase();
}

function enabledVisibleOptions(options: SearchableSelectOption[]) {
	return options.filter(option => !option.disabled);
}

export function SearchableSelect({
	options,
	value,
	onValueChange,
	placeholder = "Select",
	searchPlaceholder = "Search…",
	emptyLabel = "No matches",
	invalid,
	disabled,
	size = "md",
	frameClassName,
	menuClassName,
	className,
	name,
	form,
	required,
	"aria-invalid": ariaInvalidProp,
	"aria-label": ariaLabel,
	"aria-labelledby": ariaLabelledBy,
	...props
}: SearchableSelectProps) {
	const generatedId = useId();
	const listboxId = `${generatedId}-listbox`;
	const inputId = `${generatedId}-search`;
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const [activeValue, setActiveValue] = useState(value);
	const triggerRef = useRef<HTMLButtonElement>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const ariaInvalid = invalid || ariaInvalidProp === true || ariaInvalidProp === "true";
	const selectedOption = options.find(option => option.value === value);
	const normalizedQuery = query.trim().toLocaleLowerCase();
	const visibleOptions = useMemo(() => {
		if (!normalizedQuery) {
			return options;
		}

		return options.filter(option => optionSearchText(option).includes(normalizedQuery));
	}, [normalizedQuery, options]);
	const activeOption = visibleOptions.find(option => option.value === activeValue && !option.disabled) ?? enabledVisibleOptions(visibleOptions)[0];
	const activeOptionIndex = activeOption ? visibleOptions.findIndex(option => option.value === activeOption.value) : -1;
	const triggerClasses = [styles.trigger, styles[size], className].filter(Boolean).join(" ");
	const frameClasses = [styles.frame, frameClassName].filter(Boolean).join(" ");
	const menuClasses = [styles.menu, styles[size], menuClassName].filter(Boolean).join(" ");

	function setSelectOpen(nextOpen: boolean) {
		if (disabled) {
			setOpen(false);
			return;
		}

		setOpen(nextOpen);
	}

	function commitOption(option: SearchableSelectOption | undefined) {
		if (!option || option.disabled || disabled) {
			return;
		}

		onValueChange(option.value);
		setQuery("");
		setOpen(false);
		window.requestAnimationFrame(() => triggerRef.current?.focus());
	}

	function moveActiveOption(direction: 1 | -1) {
		const enabled = enabledVisibleOptions(visibleOptions);

		if (!enabled.length) {
			return;
		}

		const currentIndex = Math.max(
			0,
			enabled.findIndex(option => option.value === activeOption?.value)
		);
		const nextIndex = (currentIndex + direction + enabled.length) % enabled.length;
		setActiveValue(enabled[nextIndex].value);
	}

	function handleInputKeyDown(event: KeyboardEvent<HTMLInputElement>) {
		if (event.key === "ArrowDown") {
			event.preventDefault();
			moveActiveOption(1);
			return;
		}

		if (event.key === "ArrowUp") {
			event.preventDefault();
			moveActiveOption(-1);
			return;
		}

		if (event.key === "Enter") {
			event.preventDefault();
			commitOption(activeOption);
			return;
		}

		if (event.key === "Escape") {
			event.preventDefault();
			setOpen(false);
		}
	}

	useEffect(() => {
		if (!open) {
			return;
		}

		const nextActive = selectedOption && !selectedOption.disabled ? selectedOption : enabledVisibleOptions(visibleOptions)[0];
		setActiveValue(nextActive?.value ?? "");
	}, [open, selectedOption, visibleOptions]);

	useEffect(() => {
		if (!activeOption) {
			setActiveValue("");
		}
	}, [activeOption]);

	return (
		<PopoverPrimitive.Root open={open} onOpenChange={setSelectOpen}>
			<span className={frameClasses} data-open={open || undefined} data-invalid={Boolean(ariaInvalid) || undefined} data-disabled={Boolean(disabled) || undefined}>
				<PopoverPrimitive.Trigger asChild>
					<button
						ref={triggerRef}
						type="button"
						className={triggerClasses}
						disabled={disabled}
						aria-haspopup="listbox"
						aria-expanded={open}
						aria-invalid={ariaInvalid || undefined}
						aria-label={ariaLabel}
						aria-labelledby={ariaLabelledBy}
						data-placeholder={!selectedOption || undefined}
						{...props}
					>
						<span className={styles.triggerValue}>{selectedOption?.label ?? placeholder}</span>
					</button>
				</PopoverPrimitive.Trigger>
			</span>
			{name ? <input type="hidden" name={name} form={form} required={required} disabled={disabled} value={selectedOption?.value ?? ""} /> : null}
			<PopoverPrimitive.Portal>
				<PopoverPrimitive.Content
					className={menuClasses}
					align="start"
					sideOffset={8}
					collisionPadding={8}
					onOpenAutoFocus={event => {
						event.preventDefault();
						window.requestAnimationFrame(() => inputRef.current?.focus());
					}}
					onCloseAutoFocus={event => {
						event.preventDefault();
						triggerRef.current?.focus();
					}}
				>
					<label className={styles.search} htmlFor={inputId}>
						<input
							id={inputId}
							ref={inputRef}
							value={query}
							placeholder={searchPlaceholder}
							role="combobox"
							autoComplete="off"
							aria-autocomplete="list"
							aria-controls={listboxId}
							aria-expanded={open}
							aria-activedescendant={activeOptionIndex >= 0 ? `${generatedId}-option-${activeOptionIndex}` : undefined}
							onChange={event => setQuery(event.currentTarget.value)}
							onKeyDown={handleInputKeyDown}
						/>
					</label>
					<div id={listboxId} className={["ns-scrollbar", styles.listbox].join(" ")} role="listbox" aria-label={ariaLabel || searchPlaceholder}>
						{visibleOptions.length ? (
							visibleOptions.map((option, index) => {
								const selected = option.value === value;
								const active = option.value === activeOption?.value;

								return (
									<button
										id={`${generatedId}-option-${index}`}
										key={option.value}
										type="button"
										className={styles.option}
										role="option"
										aria-selected={selected}
										disabled={option.disabled}
										data-active={active || undefined}
										data-selected={selected || undefined}
										onMouseEnter={() => setActiveValue(option.value)}
										onClick={() => commitOption(option)}
									>
										<span className={styles.optionLabel}>{option.label}</span>
										{option.description ? <span className={styles.optionDescription}>{option.description}</span> : null}
									</button>
								);
							})
						) : (
							<div className={styles.empty}>{emptyLabel}</div>
						)}
					</div>
				</PopoverPrimitive.Content>
			</PopoverPrimitive.Portal>
		</PopoverPrimitive.Root>
	);
}
