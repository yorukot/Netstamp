/// <reference path="../../react-dom.d.ts" />

import * as Label from "@radix-ui/react-label";
import * as RadixSelect from "@radix-ui/react-select";
import type { ChangeEvent, ComponentPropsWithoutRef, KeyboardEvent as ReactKeyboardEvent, ReactNode } from "react";
import { Children, Fragment, forwardRef, isValidElement, useEffect, useId, useMemo, useRef, useState } from "react";
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
	nativeLabel: string;
	disabled?: boolean;
	radixValue: string;
}

const emptySelectValuePrefix = "__netstamp-ui-empty-select-value-";

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

function collectSelectOptions(children: ReactNode): Omit<SelectOption, "radixValue">[] {
	const options: Omit<SelectOption, "radixValue">[] = [];

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
				nativeLabel: textFromReactNode(props.children),
				disabled: props.disabled
			});
		});
	}

	visit(children);

	return options;
}

function selectOptionRadixValue(value: string, index: number) {
	return value === "" ? `${emptySelectValuePrefix}${index}__` : value;
}

function setNativeSelectValue(select: HTMLSelectElement, value: string) {
	const descriptor = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, "value");

	if (descriptor?.set) {
		descriptor.set.call(select, value);
		return;
	}

	select.value = value;
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
	const ariaInvalid = invalid || ariaInvalidProp === true || ariaInvalidProp === "true";
	const options = useMemo(
		() =>
			collectSelectOptions(children).map((option, index) => ({
				...option,
				radixValue: selectOptionRadixValue(option.value, index)
			})),
		[children]
	);
	const isControlled = value !== undefined;
	const initialValue = valueToString(value) ?? valueToString(defaultValue) ?? options[0]?.value ?? "";
	const [internalValue, setInternalValue] = useState(initialValue);
	const [open, setOpen] = useState(false);
	const selectRef = useRef<HTMLSelectElement | null>(null);
	const triggerRef = useRef<HTMLButtonElement>(null);
	const selectedValue = isControlled ? (valueToString(value) ?? "") : internalValue;
	const selectedOption = options.find(option => option.value === selectedValue);
	const radixSelectedValue = selectedOption?.radixValue ?? "";
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

	function commitValue(nextRadixValue: string) {
		const nextOption = options.find(option => option.radixValue === nextRadixValue);

		if (!nextOption || nextOption.disabled || disabled) {
			return;
		}

		const nextValue = nextOption.value;

		if (nextValue === selectedValue) {
			setOpen(false);
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

		setOpen(false);
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
	}

	useEffect(() => {
		if (isControlled || !options.length || options.some(option => option.value === internalValue)) {
			return;
		}

		setInternalValue(options[0].value);
	}, [isControlled, internalValue, options]);

	useEffect(() => {
		if (disabled) {
			setOpen(false);
		}
	}, [disabled]);

	useEffect(() => {
		if (!autoFocus || disabled) {
			return;
		}

		triggerRef.current?.focus();
	}, [autoFocus, disabled]);

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

	return (
		<RadixSelect.Root value={radixSelectedValue} open={open} onOpenChange={setOpen} onValueChange={commitValue} disabled={disabled}>
			<span className={styles.selectRoot}>
				<span className={frameClasses} data-invalid={Boolean(ariaInvalid)} data-disabled={Boolean(disabled)} data-open={open}>
					<RadixSelect.Trigger
						id={triggerId}
						ref={triggerRef}
						className={classes}
						style={style}
						tabIndex={tabIndex}
						aria-invalid={booleanAria(ariaInvalid)}
						aria-label={ariaLabel}
						aria-labelledby={ariaLabelledBy}
						aria-describedby={ariaDescribedBy}
						onKeyDown={handleTriggerKeyDown}
					>
						<span className={styles.selectValue}>{selectedOption?.label}</span>
					</RadixSelect.Trigger>
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
					{options.map(option => (
						<option key={option.radixValue} value={option.value} disabled={option.disabled}>
							{option.nativeLabel}
						</option>
					))}
				</select>
				<RadixSelect.Portal>
					<RadixSelect.Content className={menuClasses} position="popper" sideOffset={8} collisionPadding={8}>
						<RadixSelect.Viewport className={styles.selectViewport}>
							{options.map((option, index) => (
								<RadixSelect.Item key={`${option.value}-${index}`} className={styles.selectOption} value={option.radixValue} disabled={option.disabled}>
									<RadixSelect.ItemText asChild>
										<span className={styles.selectOptionLabel}>{option.label}</span>
									</RadixSelect.ItemText>
								</RadixSelect.Item>
							))}
						</RadixSelect.Viewport>
					</RadixSelect.Content>
				</RadixSelect.Portal>
			</span>
		</RadixSelect.Root>
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

export interface SelectFieldOption {
	value: string;
	label: string;
	disabled?: boolean;
}

export interface SelectFieldProps extends ComponentPropsWithoutRef<"select"> {
	label: ReactNode;
	helper?: ReactNode;
	error?: ReactNode;
	options: SelectFieldOption[];
}

export function SelectField({ label, helper, error, options, className, ...props }: SelectFieldProps) {
	const generatedId = useId();
	const id = props.id || generatedId;

	return (
		<FieldShell id={id} label={label} helper={helper} error={error}>
			<Select id={id} className={className} invalid={Boolean(error)} {...props}>
				{options.map(option => (
					<option key={option.value} value={option.value} disabled={option.disabled}>
						{option.label}
					</option>
				))}
			</Select>
		</FieldShell>
	);
}
