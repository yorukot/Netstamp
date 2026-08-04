import type { CheckDefinition, CheckType } from "@/features/checks/data/checks";
import { displayInsightTimeRange } from "@/features/insight/insightTime";
import type { InsightRefreshInterval, InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import type { Probe } from "@/features/probes/data/probes";
import { classNames } from "@/shared/utils/classNames";
import { relativeTimeOptions, relativeTimeRangeDurations } from "@/shared/utils/timeRanges";
import { Button, CategorizedMultiSelect, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, SelectField, TextField, type CategorizedMultiSelectCategory } from "@netstamp/ui";
import { ArrowClockwiseIcon } from "@phosphor-icons/react/dist/csr/ArrowClockwise";
import { CaretDownIcon } from "@phosphor-icons/react/dist/csr/CaretDown";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import styles from "./InsightControls.module.css";

interface InsightScopeMultiSelectProps {
	label: string;
	placeholder: string;
	pluralNoun: string;
	categories: CategorizedMultiSelectCategory[];
	selectedValues: string[];
	disabled?: boolean;
	onChange: (values: string[]) => void;
}

const timeRangeDurations: Record<InsightRelativeRange, number> = relativeTimeRangeDurations;
const relativeTimeKeys = {
	"15m": "controls.last15m",
	"1h": "controls.last1h",
	"6h": "controls.last6h",
	"24h": "controls.last24h",
	"7d": "controls.last7d",
	"30d": "controls.last30d"
} as const;

function normalizeSearch(values: string[]) {
	return values.join(" ").toLocaleLowerCase();
}

const InsightScopeMultiSelect = ({ label, placeholder, pluralNoun, categories, selectedValues, disabled, onChange }: InsightScopeMultiSelectProps) => {
	const { t } = useTranslation("insight");
	const primaryItems = categories[0]?.items ?? [];
	const selectedItem = selectedValues.length === 1 ? primaryItems.find(item => item.selectionValues.length === 1 && item.selectionValues[0] === selectedValues[0]) : undefined;
	const summary = selectedItem?.label ?? (selectedValues.length ? t("controls.selected", { count: selectedValues.length }) : placeholder);

	return (
		<CategorizedMultiSelect
			label={label}
			placeholder={placeholder}
			valueLabel={summary}
			categories={categories}
			selectedValues={selectedValues}
			disabled={disabled}
			searchPlaceholder={t("controls.search", { items: pluralNoun })}
			selectAllLabel={t("controls.selectAll")}
			clearVisibleLabel={t("controls.clearVisible")}
			emptyLabel={t("controls.noMatch", { items: pluralNoun })}
			optionsAriaLabel={t("controls.options", { label })}
			categoriesAriaLabel={t("controls.grouping", { label })}
			selectItemAriaLabel={item => t("controls.selectItem", { item: item.label })}
			onValueChange={onChange}
		/>
	);
};

export function ProbeScopeSelect({ probes, selectedValues, disabled, onChange }: { probes: Probe[]; selectedValues: string[]; disabled?: boolean; onChange: (values: string[]) => void }) {
	const { t } = useTranslation(["insight", "probes"]);
	const categories = useMemo<CategorizedMultiSelectCategory[]>(() => {
		const probeItems = probes.map(probe => ({
			value: probe.id,
			label: probe.name,
			description: `${probe.location} · ${t(`probes:status.${probe.status.toLowerCase() as "online" | "draining" | "offline"}`)}`,
			searchText: normalizeSearch([probe.name, probe.location, probe.status, probe.provider, ...probe.labelTokens]),
			selectionValues: [probe.id]
		}));
		const probesByLabel = new Map<string, Set<string>>();

		for (const probe of probes) {
			for (const label of probe.labelTokens) {
				const probeIds = probesByLabel.get(label) ?? new Set<string>();
				probeIds.add(probe.id);
				probesByLabel.set(label, probeIds);
			}
		}

		const labelItems = Array.from(probesByLabel, ([label, probeIdSet]) => ({
			value: label,
			label,
			description: t("controls.probeCount", { count: probeIdSet.size }),
			searchText: label.toLocaleLowerCase(),
			selectionValues: Array.from(probeIdSet)
		})).sort((a, b) => a.label.localeCompare(b.label));

		return [
			{ value: "probe", label: t("controls.byProbe"), items: probeItems },
			{ value: "label", label: t("controls.byLabel"), items: labelItems }
		];
	}, [probes, t]);

	return (
		<InsightScopeMultiSelect
			label={t("controls.selectProbe")}
			placeholder={t("controls.selectProbes")}
			pluralNoun={t("controls.probesNoun")}
			categories={categories}
			selectedValues={selectedValues}
			disabled={disabled}
			onChange={onChange}
		/>
	);
}

export function CheckScopeSelect({ checks, selectedValues, disabled, onChange }: { checks: CheckDefinition[]; selectedValues: string[]; disabled?: boolean; onChange: (values: string[]) => void }) {
	const { t } = useTranslation("insight");
	const categories = useMemo<CategorizedMultiSelectCategory[]>(() => {
		const checkTypeCategories: Array<{ value: string; label: string; type?: CheckType }> = [
			{ value: "all", label: t("controls.allTypes") },
			{ value: "ping", label: "Ping", type: "Ping" },
			{ value: "tcp", label: "TCP", type: "TCP" },
			{ value: "traceroute", label: "Traceroute", type: "Traceroute" },
			{ value: "http", label: "HTTP / HTTPS", type: "HTTP" }
		];
		const checkItems = checks.map(check => ({
			value: check.id,
			label: check.name,
			description: `${check.type} · ${check.target}`,
			searchText: normalizeSearch([check.name, check.target, check.description, check.type]),
			selectionValues: [check.id],
			type: check.type
		}));

		return checkTypeCategories.map(category => ({
			value: category.value,
			label: category.label,
			items: category.type ? checkItems.filter(item => item.type === category.type) : checkItems
		}));
	}, [checks, t]);

	return (
		<InsightScopeMultiSelect
			label={t("controls.selectCheck")}
			placeholder={t("controls.selectChecks")}
			pluralNoun={t("controls.checksNoun")}
			categories={categories}
			selectedValues={selectedValues}
			disabled={disabled}
			onChange={onChange}
		/>
	);
}

function formatDateTimeLocal(value: number) {
	const date = new Date(value);
	const offsetMs = date.getTimezoneOffset() * 60 * 1000;

	return new Date(value - offsetMs).toISOString().slice(0, 16);
}

function parseDateTimeLocal(value: string) {
	const parsed = new Date(value).getTime();

	return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

export function InsightTimeControl({
	timeMode,
	timeRange,
	timeWindow,
	refresh,
	className,
	onApplyRelative,
	onApplyAbsolute,
	onRefresh,
	onRefreshChange
}: {
	timeMode: InsightTimeMode;
	timeRange: InsightRelativeRange;
	timeWindow: TimeWindow;
	refresh: InsightRefreshInterval;
	className?: string;
	onApplyRelative: (range: InsightRelativeRange) => void;
	onApplyAbsolute: (timeWindow: TimeWindow) => void;
	onRefresh: () => void;
	onRefreshChange: (refresh: InsightRefreshInterval) => void;
}) {
	const { t } = useTranslation("insight");
	const timeOptions: Array<{ value: InsightRelativeRange; label: string }> = relativeTimeOptions.map(option => ({ value: option.value, label: t(relativeTimeKeys[option.value]) }));
	const refreshOptions: Array<{ value: InsightRefreshInterval; label: string }> = [
		{ value: "off", label: t("controls.off") },
		{ value: "10s", label: "10s" },
		{ value: "30s", label: "30s" },
		{ value: "1m", label: "1m" },
		{ value: "5m", label: "5m" }
	];
	const [open, setOpen] = useState(false);
	const timeWindowKey = `${timeWindow.from}:${timeWindow.to}`;
	const initialAbsoluteDraft = {
		key: timeWindowKey,
		from: formatDateTimeLocal(timeWindow.from),
		to: formatDateTimeLocal(timeWindow.to)
	};
	const [absoluteDraft, setAbsoluteDraft] = useState(initialAbsoluteDraft);
	const activeAbsoluteDraft = absoluteDraft.key === timeWindowKey ? absoluteDraft : initialAbsoluteDraft;
	const absoluteFrom = activeAbsoluteDraft.from;
	const absoluteTo = activeAbsoluteDraft.to;
	const timeButtonId = useId();

	function updateAbsoluteDraft(field: "from" | "to", value: string) {
		const nextDraft = { ...activeAbsoluteDraft, [field]: value };
		const nextFrom = parseDateTimeLocal(nextDraft.from);
		const nextTo = parseDateTimeLocal(nextDraft.to);

		setAbsoluteDraft(nextDraft);
		if (nextFrom !== null && nextTo !== null && nextFrom < nextTo) {
			onApplyAbsolute({ from: nextFrom, to: nextTo });
		}
	}

	function applyNow() {
		if (timeMode === "relative") {
			onApplyRelative(timeRange);
			return;
		}

		const duration = Math.max(timeWindow.to - timeWindow.from, timeRangeDurations["15m"]);
		const to = Date.now();
		onApplyAbsolute({ from: to - duration, to });
	}

	return (
		<div className={classNames(styles.timeControlRoot, className)}>
			<PopoverRoot open={open} onOpenChange={setOpen}>
				<div className={styles.timeField}>
					<span className={styles.controlLabel}>{t("controls.time")}</span>
					<PopoverTrigger asChild>
						<button id={timeButtonId} type="button" className={styles.timeTrigger} aria-controls={`${timeButtonId}-panel`}>
							<span>{displayInsightTimeRange(timeMode, timeRange, timeWindow)}</span>
							<CaretDownIcon className={styles.controlIcon} size="1rem" weight="bold" aria-hidden="true" focusable="false" />
						</button>
					</PopoverTrigger>
				</div>
				<PopoverPortal>
					<PopoverContent
						id={`${timeButtonId}-panel`}
						className={classNames("ns-scrollbar", styles.timePopover)}
						role="dialog"
						aria-labelledby={timeButtonId}
						align="start"
						sideOffset={8}
						collisionPadding={8}
					>
						<section className={styles.timeSection}>
							<h4>{t("controls.relative")}</h4>
							<div className={styles.timePresetGrid}>
								{timeOptions.map(option => (
									<button
										type="button"
										className={styles.timePreset}
										data-selected={timeMode === "relative" && timeRange === option.value}
										onClick={() => {
											onApplyRelative(option.value);
											setOpen(false);
										}}
										key={option.value}
									>
										{option.label}
									</button>
								))}
							</div>
						</section>
						<section className={styles.timeSection}>
							<h4>{t("controls.absolute")}</h4>
							<div className={styles.absoluteGrid}>
								<TextField label={t("controls.from")} type="datetime-local" value={absoluteFrom} onChange={event => updateAbsoluteDraft("from", event.currentTarget.value)} />
								<TextField label={t("controls.to")} type="datetime-local" value={absoluteTo} onChange={event => updateAbsoluteDraft("to", event.currentTarget.value)} />
							</div>
							<div className={styles.timeActions}>
								<Button type="button" variant="outline" size="sm" onClick={applyNow}>
									{t("controls.now")}
								</Button>
							</div>
						</section>
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>
			<SelectField label={t("controls.refresh")} value={refresh} options={refreshOptions} onChange={event => onRefreshChange(event.currentTarget.value as InsightRefreshInterval)} />
			<Button type="button" variant="outline" className={styles.refreshButton} aria-label={t("controls.refreshInsight")} title={t("controls.refreshInsight")} onClick={onRefresh}>
				<ArrowClockwiseIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
			</Button>
		</div>
	);
}
