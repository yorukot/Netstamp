import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge, CodeBlock, DataTable, FieldLabel, GlobalFooter, Input, KeyValueRow, MetricTile, Panel, SelectField, SpecLabel, TextAreaField, TextField, type DataColumn } from "../index";

interface ProbeRow {
	probe: string;
	region: string;
	status: "healthy" | "degraded" | "failed";
	p95: string;
	loss: string;
}

const rows: ProbeRow[] = [
	{ probe: "ams-edge-01", region: "eu-west", status: "healthy", p95: "42ms", loss: "0.00%" },
	{ probe: "tpe-lab-02", region: "ap-east", status: "degraded", p95: "118ms", loss: "0.12%" },
	{ probe: "sfo-core-03", region: "us-west", status: "healthy", p95: "64ms", loss: "0.01%" },
	{ probe: "sin-route-04", region: "ap-south", status: "failed", p95: "--", loss: "100%" }
];

const statusTone: Record<ProbeRow["status"], "success" | "warning" | "critical"> = {
	degraded: "warning",
	failed: "critical",
	healthy: "success"
};

const columns: DataColumn<ProbeRow>[] = [
	{ key: "probe", label: "Probe", sortable: true },
	{ key: "region", label: "Region" },
	{
		key: "status",
		label: "Status",
		render: row => <Badge tone={statusTone[row.status]}>{row.status}</Badge>
	},
	{ key: "p95", label: "p95" },
	{ key: "loss", label: "Loss" }
];

const storyShortcuts = [
	{ label: "Primitives", href: "./?path=/story/components-button--playground" },
	{ label: "Forms", href: "./?path=/story/forms-textfield--default" },
	{ label: "Data", href: "./?path=/story/components-datatable--default" },
	{ label: "Patterns", href: "./?path=/story/patterns-operational-workspace--dashboard-spec" }
] as const;

const componentGroups = [
	["Actions", "Button, IconButton, ActionRow, DisclosureToggle"],
	["Forms", "TextField, TextAreaField, SelectField, SearchableSelect, Checkbox"],
	["Surfaces", "Panel, Surface, MetricCard, MetricTile, SpecCard"],
	["Data", "DataTable, KeyValueRow, Badge, SpecLabel, CodeBlock"],
	["Layout", "PageShell, Drawer, Dialog, Tabs, SegmentedControl, GlobalFooter"]
] as const;

const meta = {
	title: "Overview",
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const Overview: Story = {
	render: () => (
		<div className="storybook-page storybook-dashboard-page">
			<header className="storybook-dashboard-hero">
				<div className="storybook-dashboard-hero-copy">
					<img src={netstampLogo} alt="Netstamp" className="storybook-logo" />
					<SpecLabel tone="primary">Shared UI system</SpecLabel>
					<h1>Dashboard-grade primitives for Netstamp.</h1>
					<p>The library should feel like the product dashboard: dark by default, dense, technical, mostly borderless, and strict about when a frame is allowed to appear.</p>
					<nav className="storybook-shortcuts" aria-label="Story shortcuts">
						{storyShortcuts.map(shortcut => (
							<a key={shortcut.label} href={shortcut.href} target="_top">
								{shortcut.label}
							</a>
						))}
					</nav>
				</div>
			</header>

			<section className="storybook-dashboard-layout" aria-label="Design system overview">
				<Panel className="storybook-dashboard-overview" title="Surface contract" padded={false} bodyClassName="storybook-dashboard-metrics">
					<MetricTile label="Default borders" value="off" detail="quiet" tone="muted" description="Neutral cards and tags use tone first." />
					<MetricTile label="Section frame" value="dashed" detail="panel" tone="accent" description="Dashboard sections keep the dashed outer boundary." />
					<MetricTile label="State marks" value="explicit" detail="semantic" tone="success" description="Only meaningful state gets colored markers." />
					<MetricTile label="Focus" value="visible" detail="keyboard" tone="neutral" description="Keyboard focus stays clear without mouse outlines." />
				</Panel>

				<Panel title="Component map" bodyClassName="storybook-key-list">
					{componentGroups.map(([label, value]) => (
						<KeyValueRow key={label} label={label} value={value} />
					))}
				</Panel>

				<Panel className="storybook-dashboard-data" title="Telemetry table" summary="Badges, rows, and headers use fills and dividers instead of every cell carrying a frame." padded={false}>
					<DataTable columns={columns} rows={rows} ariaLabel="Probe latency examples" getRowKey={row => row.probe} density="compact" minWidth="42rem" />
				</Panel>

				<Panel title="Control stack" summary="Inputs still keep a quiet one-pixel affordance; hover and keyboard focus change color, not thickness.">
					<div className="storybook-form-grid">
						<TextField label="Probe name" defaultValue="tpe-lab-02" helper="Visible in route and probe tables." />
						<SelectField
							label="Interval"
							defaultValue="30s"
							options={[
								{ value: "10s", label: "10 seconds" },
								{ value: "30s", label: "30 seconds" },
								{ value: "60s", label: "60 seconds" }
							]}
						/>
						<TextAreaField label="Notes" defaultValue="Measure DNS and ICMP from this probe." />
						<label className="storybook-compact-field">
							<FieldLabel>Compact token</FieldLabel>
							<Input variant="compact" defaultValue="trace-window" />
						</label>
					</div>
				</Panel>

				<Panel title="Operator surfaces" summary="Deep surfaces are separated by tone and header structure, not decorative frames.">
					<div className="storybook-operator-grid">
						<CodeBlock title="component rule" meta="css">
							{`badge {
	border: 0;
	background: var(--ns-primary-muted);
}`}
						</CodeBlock>
						<CodeBlock title="netstamp probe" meta="shell">
							netstamp probe run --check ping --target 1.1.1.1
						</CodeBlock>
					</div>
				</Panel>
			</section>

			<GlobalFooter />
		</div>
	)
};
