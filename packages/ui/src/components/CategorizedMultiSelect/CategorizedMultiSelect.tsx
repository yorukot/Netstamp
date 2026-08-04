import { CaretDownIcon } from "@phosphor-icons/react/dist/csr/CaretDown";
import { MagnifyingGlassIcon } from "@phosphor-icons/react/dist/csr/MagnifyingGlass";
import { useEffect, useId, useMemo, useRef, useState, type ReactNode } from "react";
import { Button } from "../Button/Button";
import { PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger } from "../Dialog/Dialog";
import { Checkbox } from "../Field/Field";
import styles from "./CategorizedMultiSelect.module.css";

export interface CategorizedMultiSelectItem {
	value: string;
	label: ReactNode;
	description?: ReactNode;
	searchText: string;
	selectionValues: string[];
}

export interface CategorizedMultiSelectCategory {
	value: string;
	label: ReactNode;
	items: CategorizedMultiSelectItem[];
}

export interface CategorizedMultiSelectProps {
	label: ReactNode;
	placeholder: ReactNode;
	valueLabel: ReactNode;
	categories: CategorizedMultiSelectCategory[];
	selectedValues: string[];
	disabled?: boolean;
	searchPlaceholder: string;
	selectAllLabel: ReactNode;
	clearVisibleLabel: ReactNode;
	emptyLabel: ReactNode;
	optionsAriaLabel: string;
	categoriesAriaLabel: string;
	selectItemAriaLabel: (item: CategorizedMultiSelectItem) => string;
	onValueChange: (values: string[]) => void;
}

const uniqueValues = (values: string[]) => Array.from(new Set(values));

const SelectionCheckbox = ({ checked, mixed, label, onChange }: { checked: boolean; mixed: boolean; label: string; onChange: () => void }) => {
	const ref = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (ref.current) {
			ref.current.indeterminate = mixed;
		}
	}, [mixed]);

	return <Checkbox ref={ref} checked={checked} aria-checked={mixed ? "mixed" : checked} aria-label={label} onChange={onChange} />;
};

export const CategorizedMultiSelect = ({
	label,
	placeholder,
	valueLabel,
	categories,
	selectedValues,
	disabled,
	searchPlaceholder,
	selectAllLabel,
	clearVisibleLabel,
	emptyLabel,
	optionsAriaLabel,
	categoriesAriaLabel,
	selectItemAriaLabel,
	onValueChange
}: CategorizedMultiSelectProps) => {
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const [requestedCategory, setRequestedCategory] = useState(categories[0]?.value ?? "");
	const triggerRef = useRef<HTMLButtonElement>(null);
	const searchRef = useRef<HTMLInputElement>(null);
	const categoryId = useId();
	const panelId = useId();
	const selectedSet = useMemo(() => new Set(selectedValues), [selectedValues]);
	const activeCategory = categories.find(category => category.value === requestedCategory) ?? categories[0];
	const normalizedQuery = query.trim().toLocaleLowerCase();
	const visibleItems = useMemo(() => activeCategory?.items.filter(item => !normalizedQuery || item.searchText.includes(normalizedQuery)) ?? [], [activeCategory, normalizedQuery]);
	const visibleSelectionValues = useMemo(() => uniqueValues(visibleItems.flatMap(item => item.selectionValues)), [visibleItems]);
	const allVisibleSelected = visibleSelectionValues.length > 0 && visibleSelectionValues.every(value => selectedSet.has(value));

	const setPickerOpen = (nextOpen: boolean) => {
		if (disabled) {
			setOpen(false);
			return;
		}

		setOpen(nextOpen);
		if (!nextOpen) {
			setQuery("");
		}
	};

	const toggleItem = (item: CategorizedMultiSelectItem) => {
		const itemSelected = item.selectionValues.every(value => selectedSet.has(value));
		if (itemSelected) {
			const valuesToRemove = new Set(item.selectionValues);
			onValueChange(selectedValues.filter(value => !valuesToRemove.has(value)));
			return;
		}

		onValueChange(uniqueValues([...selectedValues, ...item.selectionValues]));
	};

	const toggleVisibleItems = () => {
		if (!visibleSelectionValues.length) {
			return;
		}

		if (allVisibleSelected) {
			const valuesToRemove = new Set(visibleSelectionValues);
			onValueChange(selectedValues.filter(value => !valuesToRemove.has(value)));
			return;
		}

		onValueChange(uniqueValues([...selectedValues, ...visibleSelectionValues]));
	};

	return (
		<div className={styles.field}>
			<span className={styles.label}>{label}</span>
			<PopoverRoot open={open} onOpenChange={setPickerOpen}>
				<PopoverTrigger asChild>
					<button ref={triggerRef} type="button" className={styles.trigger} disabled={disabled} aria-haspopup="dialog" data-placeholder={!selectedValues.length || undefined}>
						<span>{selectedValues.length ? valueLabel : placeholder}</span>
						<CaretDownIcon className={styles.icon} size="1rem" weight="bold" aria-hidden="true" focusable="false" />
					</button>
				</PopoverTrigger>
				<PopoverPortal>
					<PopoverContent
						className={styles.popover}
						role="dialog"
						aria-label={optionsAriaLabel}
						align="start"
						sideOffset={8}
						collisionPadding={8}
						onOpenAutoFocus={event => {
							event.preventDefault();
							window.requestAnimationFrame(() => searchRef.current?.focus());
						}}
						onCloseAutoFocus={event => {
							event.preventDefault();
							triggerRef.current?.focus();
						}}
					>
						<div className={styles.header}>
							<label className={styles.search}>
								<MagnifyingGlassIcon size="1rem" aria-hidden="true" focusable="false" />
								<input ref={searchRef} type="search" value={query} placeholder={searchPlaceholder} aria-label={searchPlaceholder} onChange={event => setQuery(event.currentTarget.value)} />
							</label>
							<Button type="button" variant="ghost" size="sm" disabled={!visibleSelectionValues.length} onClick={toggleVisibleItems}>
								{allVisibleSelected ? clearVisibleLabel : selectAllLabel}
							</Button>
						</div>
						<div className={styles.body}>
							<div className={styles.categories} role="tablist" aria-label={categoriesAriaLabel}>
								{categories.map(category => {
									const selected = category.value === activeCategory?.value;

									return (
										<button
											id={`${categoryId}-${category.value}`}
											type="button"
											role="tab"
											aria-selected={selected}
											aria-controls={selected ? panelId : undefined}
											data-selected={selected || undefined}
											onClick={() => {
												setRequestedCategory(category.value);
												setQuery("");
											}}
											key={category.value}
										>
											<span>{category.label}</span>
											<small>{category.items.length}</small>
										</button>
									);
								})}
							</div>
							<div id={panelId} className={styles.optionPane} role="tabpanel" aria-labelledby={`${categoryId}-${activeCategory?.value}`}>
								<div className={["ns-scrollbar", styles.optionList].join(" ")} role="group" aria-label={typeof activeCategory?.label === "string" ? activeCategory.label : categoriesAriaLabel}>
									{visibleItems.length ? (
										visibleItems.map(item => {
											const selectedCount = item.selectionValues.filter(value => selectedSet.has(value)).length;
											const checked = item.selectionValues.length > 0 && selectedCount === item.selectionValues.length;
											const mixed = selectedCount > 0 && !checked;

											return (
												<label className={styles.option} data-selected={checked || mixed || undefined} key={item.value}>
													<SelectionCheckbox checked={checked} mixed={mixed} label={selectItemAriaLabel(item)} onChange={() => toggleItem(item)} />
													<span>
														<strong>{item.label}</strong>
														{item.description ? <small>{item.description}</small> : null}
													</span>
												</label>
											);
										})
									) : (
										<div className={styles.empty}>{emptyLabel}</div>
									)}
								</div>
							</div>
						</div>
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>
		</div>
	);
};
