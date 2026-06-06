/// <reference path="../../react-dom.d.ts" />

import * as Label from "@radix-ui/react-label";
import type { CSSProperties, ChangeEvent, ComponentPropsWithoutRef, KeyboardEvent as ReactKeyboardEvent, ReactNode } from "react";
import { Children, Fragment, forwardRef, isValidElement, useEffect, useId, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import styles from "./Field.module.css";

export type ControlVariant = "default" | "compact" | "bare";

function booleanAria(value: boolean | undefined) {
	return value ? true : undefined;
}

export interface InputProps extends ComponentPropsWithoutRef<"input"> {
	variant?: ControlVariant;
	invalid?: boolean;
	frameClassName?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input({ variant = "default", invalid, frameClassName, className, "aria-invalid": ariaInvalidProp, ...props }, ref) {
	const ariaInvalid = invalid || ariaInvalidProp === true || ariaInvalidProp === "true";
	const classes = [styles.control, styles[`${variant}Control`], className].filter(Boolean).join(" ");

	if (variant === "bare") {
		return <input ref={ref} className={classes} aria-invalid={booleanAria(ariaInvalid)} {...props} />;
	}

	const frameClasses = ["ns-cut-frame", styles.controlFrame, styles[`${variant}Frame`], frameClassName].filter(Boolean).join(" ");

	return (
		<span className={frameClasses} data-invalid={Boolean(ariaInvalid)}>
			<input ref={ref} className={classes} aria-invalid={booleanAria(ariaInvalid)} {...props} />
		</span>
	);
});

export interface SelectProps extends ComponentPropsWithoutRef<"select"> {
	variant?: Exclude<ControlVariant, "bare">;
	invalid?: boolean;
	frameClassName?: string;
}

interface SelectOption {
	value: string;
	label: ReactNode;
	disabled?: boolean;
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

function valueToString(value: unknown): string | undefined {
	if (value === undefined || value === null) {
		return undefined;
	}

	if (Array.isArray(value)) {
		return value[0] === undefined || value[0] === null ? undefined : String(value[0]);
	}

	return String(value);
}

function collectSelectOptions(children: ReactNode): SelectOption[] {
	const options: SelectOption[] = [];

	function visit(node: ReactNode) {
		Children.forEach(node, child => {
			if (!isValidElement(child)) {
				return;
			}

			if (child.type === Fragment || child.type === "optgroup") {
				visit((child.props as { children?: ReactNode }).children);
				return;
			}

			if (child.type !== "option") {
				return;
			}

			const props = child.props as ComponentPropsWithoutRef<"option">;

			options.push({
				value: valueToString(props.value) ?? textFromReactNode(props.children),
				label: props.children,
				disabled: props.disabled
			});
		});
	}

	visit(children);

	return options;
}

function setNativeSelectValue(select: HTMLSelectElement, value: string) {
	const descriptor = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, "value");

	if (descriptor?.set) {
		descriptor.set.call(select, value);
		return;
	}

	select.value = value;
}

function getSelectMenuStyle(frame: HTMLElement | null, menu?: HTMLElement | null): CSSProperties | undefined {
	if (!frame || typeof window === "undefined") {
		return undefined;
	}

	const rect = frame.getBoundingClientRect();
	const gap = 8;
	const viewportWidth = window.innerWidth;
	const viewportHeight = window.innerHeight;
	const width = Math.min(rect.width, viewportWidth - gap * 2);
	const left = Math.min(Math.max(gap, rect.left), viewportWidth - width - gap);
	const spaceBelow = viewportHeight - rect.bottom - gap;
	const spaceAbove = rect.top - gap;
	const preferredHeight = menu ? Math.min(288, menu.scrollHeight) : 176;
	const openAbove = preferredHeight > spaceBelow && spaceAbove > spaceBelow;
	const availableHeight = openAbove ? spaceAbove : spaceBelow;
	const maxHeight = Math.min(288, Math.max(96, availableHeight));
	const renderedHeight = menu ? Math.min(menu.scrollHeight, maxHeight) : maxHeight;
	const top = openAbove ? Math.max(gap, rect.top - gap - renderedHeight) : Math.min(rect.bottom + gap, viewportHeight - gap - renderedHeight);

	return {
		left,
		maxHeight,
		top,
		width
	};
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(function Select(
	{
		variant = "default",
		invalid,
		frameClassName,
		className,
		children,
		id,
		disabled,
		value,
		defaultValue,
		onChange,
		onKeyDown,
		style,
		tabIndex,
		autoFocus,
		multiple,
		"aria-invalid": ariaInvalidProp,
		"aria-label": ariaLabel,
		"aria-labelledby": ariaLabelledBy,
		"aria-describedby": ariaDescribedBy,
		...props
	},
	ref
) {
	const generatedId = useId();
	const triggerId = id || generatedId;
	const listboxId = `${triggerId}-listbox`;
	const ariaInvalid = invalid || ariaInvalidProp === true || ariaInvalidProp === "true";
	const options = useMemo(() => collectSelectOptions(children), [children]);
	const isControlled = value !== undefined;
	const initialValue = valueToString(value) ?? valueToString(defaultValue) ?? options[0]?.value ?? "";
	const [internalValue, setInternalValue] = useState(initialValue);
	const [open, setOpen] = useState(false);
	const [activeValue, setActiveValue] = useState(initialValue);
	const [menuStyle, setMenuStyle] = useState<CSSProperties>();
	const rootRef = useRef<HTMLSpanElement>(null);
	const frameRef = useRef<HTMLSpanElement>(null);
	const selectRef = useRef<HTMLSelectElement | null>(null);
	const triggerRef = useRef<HTMLButtonElement>(null);
	const menuRef = useRef<HTMLDivElement>(null);
	const selectedValue = isControlled ? (valueToString(value) ?? "") : internalValue;
	const selectedOption = options.find(option => option.value === selectedValue);
	const activeIndex = options.findIndex(option => option.value === activeValue);
	const activeOptionId = open && activeIndex >= 0 ? `${listboxId}-option-${activeIndex}` : undefined;
	const classes = [styles.control, styles.select, styles[`${variant}Control`], className].filter(Boolean).join(" ");
	const frameClasses = ["ns-cut-frame", styles.controlFrame, styles.selectFrame, styles[`${variant}Frame`], frameClassName].filter(Boolean).join(" ");
	const menuClasses = [styles.selectMenu, styles[`${variant}Menu`]].filter(Boolean).join(" ");

	function setSelectNode(node: HTMLSelectElement | null) {
		selectRef.current = node;

		if (typeof ref === "function") {
			ref(node);
			return;
		}

		if (ref) {
			ref.current = node;
		}
	}

	function enabledOptionFrom(value: string, offset: number) {
		const enabledOptions = options.filter(option => !option.disabled);

		if (!enabledOptions.length) {
			return undefined;
		}

		const currentIndex = enabledOptions.findIndex(option => option.value === value);
		const nextIndex = currentIndex === -1 ? (offset > 0 ? 0 : enabledOptions.length - 1) : (currentIndex + offset + enabledOptions.length) % enabledOptions.length;

		return enabledOptions[nextIndex];
	}

	function firstEnabledOption() {
		return options.find(option => !option.disabled);
	}

	function lastEnabledOption() {
		for (let index = options.length - 1; index >= 0; index -= 1) {
			if (!options[index].disabled) {
				return options[index];
			}
		}

		return undefined;
	}

	function openMenu() {
		if (disabled || !options.length) {
			return;
		}

		const selectedEnabled = options.find(option => option.value === selectedValue && !option.disabled);
		setActiveValue((selectedEnabled ?? firstEnabledOption() ?? selectedOption ?? options[0]).value);
		setMenuStyle(getSelectMenuStyle(frameRef.current));
		setOpen(true);
	}

	function closeMenu() {
		setOpen(false);
	}

	function commitValue(nextValue: string) {
		const nextOption = options.find(option => option.value === nextValue);

		if (!nextOption || nextOption.disabled || disabled) {
			return;
		}

		if (nextValue === selectedValue) {
			setActiveValue(nextValue);
			closeMenu();
			triggerRef.current?.focus();
			return;
		}

		const select = selectRef.current;

		if (select) {
			setNativeSelectValue(select, nextValue);
			select.dispatchEvent(new Event("change", { bubbles: true }));
		} else if (!isControlled) {
			setInternalValue(nextValue);
		}

		setActiveValue(nextValue);
		closeMenu();
		triggerRef.current?.focus();
	}

	function handleNativeChange(event: ChangeEvent<HTMLSelectElement>) {
		if (!isControlled) {
			setInternalValue(event.currentTarget.value);
		}

		onChange?.(event);
	}

	function handleTriggerKeyDown(event: ReactKeyboardEvent<HTMLButtonElement>) {
		onKeyDown?.(event as unknown as ReactKeyboardEvent<HTMLSelectElement>);

		if (event.defaultPrevented || disabled) {
			return;
		}

		if (event.key === "ArrowDown" || event.key === "ArrowUp") {
			event.preventDefault();

			if (!open) {
				openMenu();
				return;
			}

			const nextOption = enabledOptionFrom(open ? activeValue : selectedValue, event.key === "ArrowDown" ? 1 : -1);

			if (nextOption) {
				setActiveValue(nextOption.value);
			}

			return;
		}

		if (event.key === "Home" || event.key === "End") {
			event.preventDefault();
			const nextOption = event.key === "Home" ? firstEnabledOption() : lastEnabledOption();

			if (nextOption) {
				setActiveValue(nextOption.value);
			}

			if (!open && nextOption) {
				setOpen(true);
			}

			return;
		}

		if (event.key === "Enter" || event.key === " ") {
			event.preventDefault();

			if (open) {
				commitValue(activeValue);
				return;
			}

			openMenu();
			return;
		}

		if (event.key === "Escape" && open) {
			event.preventDefault();
			closeMenu();
			return;
		}

		if (event.key === "Tab") {
			closeMenu();
		}
	}

	useEffect(() => {
		if (isControlled || !options.length || options.some(option => option.value === internalValue)) {
			return;
		}

		setInternalValue(options[0].value);
	}, [isControlled, internalValue, options]);

	useEffect(() => {
		if (disabled) {
			closeMenu();
		}
	}, [disabled]);

	useEffect(() => {
		if (!autoFocus || disabled) {
			return;
		}

		triggerRef.current?.focus();
	}, [autoFocus, disabled]);

	useLayoutEffect(() => {
		if (!open) {
			return;
		}

		function updateMenuPosition() {
			const nextMenuStyle = getSelectMenuStyle(frameRef.current, menuRef.current);

			if (!nextMenuStyle) {
				return;
			}

			setMenuStyle(nextMenuStyle);
		}

		function handlePointerDown(event: PointerEvent) {
			const target = event.target as Node;

			if (rootRef.current?.contains(target) || menuRef.current?.contains(target)) {
				return;
			}

			closeMenu();
		}

		updateMenuPosition();
		window.addEventListener("resize", updateMenuPosition);
		window.addEventListener("scroll", updateMenuPosition, true);
		window.addEventListener("pointerdown", handlePointerDown);

		return () => {
			window.removeEventListener("resize", updateMenuPosition);
			window.removeEventListener("scroll", updateMenuPosition, true);
			window.removeEventListener("pointerdown", handlePointerDown);
		};
	}, [open]);

	useLayoutEffect(() => {
		const menu = menuRef.current;

		if (!activeOptionId || !menu || typeof document === "undefined") {
			return;
		}

		const activeOption = document.getElementById(activeOptionId);

		if (!activeOption) {
			return;
		}

		const optionTop = activeOption.offsetTop;
		const optionBottom = optionTop + activeOption.offsetHeight;
		const menuBottom = menu.scrollTop + menu.clientHeight;

		if (optionTop < menu.scrollTop) {
			menu.scrollTop = optionTop;
			return;
		}

		if (optionBottom > menuBottom) {
			menu.scrollTop = optionBottom - menu.clientHeight;
		}
	}, [activeOptionId]);

	if (multiple) {
		return (
			<span className={frameClasses} data-invalid={Boolean(ariaInvalid)} data-disabled={Boolean(disabled)}>
				<select
					ref={setSelectNode}
					className={classes}
					aria-invalid={booleanAria(ariaInvalid)}
					disabled={disabled}
					multiple={multiple}
					defaultValue={defaultValue}
					value={value}
					onChange={onChange}
					{...props}
				>
					{children}
				</select>
			</span>
		);
	}

	const menu =
		open && menuStyle && typeof document !== "undefined"
			? createPortal(
					<div ref={menuRef} id={listboxId} className={menuClasses} style={menuStyle} role="listbox" aria-label={ariaLabel} aria-labelledby={ariaLabel ? undefined : triggerId}>
						{options.map((option, index) => {
							const selected = option.value === selectedValue;
							const active = option.value === activeValue;

							return (
								<div
									key={`${option.value}-${index}`}
									id={`${listboxId}-option-${index}`}
									className={styles.selectOption}
									role="option"
									aria-selected={selected}
									data-active={active || undefined}
									data-selected={selected || undefined}
									data-disabled={option.disabled || undefined}
									onMouseDown={event => event.preventDefault()}
									onMouseEnter={() => {
										if (!option.disabled) {
											setActiveValue(option.value);
										}
									}}
									onClick={() => commitValue(option.value)}
								>
									<span className={styles.selectOptionLabel}>{option.label}</span>
								</div>
							);
						})}
					</div>,
					document.body
				)
			: null;

	return (
		<span ref={rootRef} className={styles.selectRoot}>
			<span ref={frameRef} className={frameClasses} data-invalid={Boolean(ariaInvalid)} data-disabled={Boolean(disabled)} data-open={open}>
				<button
					type="button"
					id={triggerId}
					ref={triggerRef}
					className={classes}
					style={style}
					tabIndex={tabIndex}
					disabled={disabled}
					aria-invalid={booleanAria(ariaInvalid)}
					aria-label={ariaLabel}
					aria-labelledby={ariaLabelledBy}
					aria-describedby={ariaDescribedBy}
					aria-haspopup="listbox"
					aria-expanded={open}
					aria-controls={open ? listboxId : undefined}
					aria-activedescendant={activeOptionId}
					onClick={() => {
						if (open) {
							closeMenu();
							return;
						}

						openMenu();
					}}
					onKeyDown={handleTriggerKeyDown}
				>
					<span className={styles.selectValue}>{selectedOption?.label}</span>
				</button>
			</span>
			<select
				ref={setSelectNode}
				className={styles.nativeSelect}
				aria-hidden="true"
				tabIndex={-1}
				disabled={disabled}
				defaultValue={undefined}
				value={selectedValue}
				onChange={handleNativeChange}
				{...props}
			>
				{children}
			</select>
			{menu}
		</span>
	);
});

export interface CheckboxProps extends Omit<ComponentPropsWithoutRef<"input">, "type"> {
	invalid?: boolean;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(function Checkbox({ invalid, className, "aria-invalid": ariaInvalidProp, ...props }, ref) {
	const ariaInvalid = invalid || ariaInvalidProp === true || ariaInvalidProp === "true";
	const classes = [styles.checkbox, className].filter(Boolean).join(" ");

	return <input ref={ref} type="checkbox" className={classes} aria-invalid={booleanAria(ariaInvalid)} {...props} />;
});

export interface FieldLabelProps extends ComponentPropsWithoutRef<"span"> {}

export function FieldLabel({ className, ...props }: FieldLabelProps) {
	return <span className={[styles.label, className].filter(Boolean).join(" ")} {...props} />;
}

interface FieldShellProps {
	id: string;
	label: ReactNode;
	helper?: ReactNode;
	error?: ReactNode;
	children: ReactNode;
}

function FieldShell({ id, label, helper, error, children }: FieldShellProps) {
	return (
		<div className={styles.field}>
			<Label.Root className={styles.label} htmlFor={id}>
				{label}
			</Label.Root>
			{children}
			{error ? <span className={styles.error}>{error}</span> : null}
			{helper && !error ? <span className={styles.helper}>{helper}</span> : null}
		</div>
	);
}

export interface TextFieldProps extends ComponentPropsWithoutRef<"input"> {
	label: ReactNode;
	helper?: ReactNode;
	error?: ReactNode;
}

export function TextField({ label, helper, error, className, ...props }: TextFieldProps) {
	const generatedId = useId();
	const id = props.id || generatedId;

	return (
		<FieldShell id={id} label={label} helper={helper} error={error}>
			<Input id={id} className={className} invalid={Boolean(error)} {...props} />
		</FieldShell>
	);
}

export interface TextAreaFieldProps extends ComponentPropsWithoutRef<"textarea"> {
	label: ReactNode;
	helper?: ReactNode;
	error?: ReactNode;
}

export function TextAreaField({ label, helper, error, className, ...props }: TextAreaFieldProps) {
	const generatedId = useId();
	const id = props.id || generatedId;
	const classes = [styles.control, styles.area, className].filter(Boolean).join(" ");

	return (
		<FieldShell id={id} label={label} helper={helper} error={error}>
			<span className={["ns-cut-frame", styles.controlFrame].join(" ")} data-invalid={Boolean(error)}>
				<textarea id={id} className={classes} aria-invalid={Boolean(error)} {...props} />
			</span>
		</FieldShell>
	);
}

export interface SelectFieldProps extends ComponentPropsWithoutRef<"select"> {
	label: ReactNode;
	helper?: ReactNode;
	error?: ReactNode;
	options: Array<{ value: string; label: string }>;
}

export function SelectField({ label, helper, error, options, className, ...props }: SelectFieldProps) {
	const generatedId = useId();
	const id = props.id || generatedId;

	return (
		<FieldShell id={id} label={label} helper={helper} error={error}>
			<Select id={id} className={className} invalid={Boolean(error)} {...props}>
				{options.map(option => (
					<option key={option.value} value={option.value}>
						{option.label}
					</option>
				))}
			</Select>
		</FieldShell>
	);
}
