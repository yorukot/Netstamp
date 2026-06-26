import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import netstampMark from "@netstamp/brand/assets/netstamp-mark-light.svg";
import type { Meta, StoryObj } from "@storybook/react-vite";
import {
	Badge,
	Button,
	Checkbox,
	DataTable,
	FieldLabel,
	GlobalFooter,
	Input,
	MetricCard,
	PageShell,
	Panel,
	SelectField,
	SignalAvatar,
	Surface,
	Terminal,
	TextAreaField,
	TextField,
	type DataColumn
} from "../index";

interface ProbeRow {
	probe: string;
	status: string;
	latency: string;
}

const rows: ProbeRow[] = [
	{ probe: "ams-edge-01", status: "online", latency: "42ms" },
	{ probe: "tpe-lab-02", status: "degraded", latency: "118ms" },
	{ probe: "sfo-core-03", status: "online", latency: "64ms" }
];

const columns: DataColumn<ProbeRow>[] = [
	{ key: "probe", label: "Probe" },
	{ key: "status", label: "Status" },
	{ key: "latency", label: "Latency" }
];

const storyShortcuts = [
	{ label: "Buttons", href: "./?path=/story/components-button--playground" },
	{ label: "Surfaces", href: "./?path=/story/components-surface--playground" },
	{ label: "Fields", href: "./?path=/story/forms-textfield--default" },
	{ label: "Data", href: "./?path=/story/components-datatable--default" }
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
		<div className="storybook-page">
			<header className="storybook-hero storybook-hero--entrance">
				<div className="storybook-hero-copy">
					<img src={netstampLogo} alt="Netstamp" className="storybook-logo" />
					<p className="storybook-kicker">Netstamp UI storyboard</p>
					<h1>Shared UI, in one pass.</h1>
					<span className="storybook-hero-lede">A compact map of the @netstamp/ui primitives used across the app, docs, API explorer, and Storybook.</span>
					<ul className="storybook-hero-points">
						<li>Token roles: orange primary actions, blue reference data, semantic health states.</li>
						<li>Surface rules: square frames, flat panels, no decorative gradients or shadows.</li>
						<li>Core patterns: controls, forms, telemetry cards, tables, terminals, and shell framing.</li>
					</ul>
					<nav className="storybook-shortcuts" aria-label="Story shortcuts">
						{storyShortcuts.map(shortcut => (
							<a key={shortcut.label} href={shortcut.href} target="_top">
								{shortcut.label}
							</a>
						))}
					</nav>
				</div>
			</header>

			<section className="storybook-section">
				<div className="storybook-section-header">
					<span>Actions</span>
					<h2>Buttons and badges</h2>
				</div>
				<div className="storybook-inline-grid">
					<Button>Primary</Button>
					<Button variant="secondary">Secondary</Button>
					<Button variant="outline">Outline</Button>
					<Button variant="ghost">Ghost</Button>
					<Button variant="danger">Danger</Button>
					<Badge tone="accent">Accent</Badge>
					<Badge tone="success">Success</Badge>
					<Badge tone="warning">Warning</Badge>
					<Badge tone="critical">Critical</Badge>
				</div>
			</section>

			<section className="storybook-section">
				<div className="storybook-section-header">
					<span>Surfaces</span>
					<h2>Panels, cards, and shells</h2>
				</div>
				<div className="storybook-card-grid">
					<MetricCard label="p95 latency" value="42ms" detail="healthy" tone="success" />
					<MetricCard label="packet loss" value="0.08%" detail="watch" tone="warning" />
					<Surface tone="accent" frameSize="lg" padding="lg">
						<strong>Accent surface</strong>
						<p>Used for high-intensity blocks and landing page visual anchors.</p>
					</Surface>
				</div>
				<Panel title="Operational summary">
					<p>Panel composes Surface with a header and action slot.</p>
				</Panel>
			</section>

			<section className="storybook-section">
				<div className="storybook-section-header">
					<span>Forms</span>
					<h2>Fields</h2>
				</div>
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
					<label className="storybook-checkbox-row">
						<Checkbox defaultChecked />
						<FieldLabel>Enable alerts</FieldLabel>
					</label>
					<label className="storybook-compact-field">
						<FieldLabel>Compact input</FieldLabel>
						<Input variant="compact" defaultValue="trace-window" />
					</label>
				</div>
			</section>

			<section className="storybook-section">
				<div className="storybook-section-header">
					<span>Data</span>
					<h2>Table and terminal</h2>
				</div>
				<DataTable columns={columns} rows={rows} ariaLabel="Probe latency examples" getRowKey={row => row.probe} />
				<Terminal title="netstamp probe" meta="dry run">
					netstamp probe run --check ping --target 1.1.1.1
				</Terminal>
			</section>

			<section className="storybook-section">
				<div className="storybook-section-header">
					<span>Identity</span>
					<h2>Avatars and page shell</h2>
				</div>
				<div className="storybook-identity-row">
					<SignalAvatar src={netstampMark} alt="Netstamp mark" size="sm" />
					<SignalAvatar src={netstampMark} alt="Netstamp mark" size="md" />
					<SignalAvatar src={netstampMark} alt="Netstamp mark" size="lg" />
				</div>
				<PageShell as="div" variant="constellation" className="storybook-shell-preview">
					<strong>PageShell preview</strong>
					<p>Background utilities match the service app visual language.</p>
				</PageShell>
			</section>

			<GlobalFooter />
		</div>
	)
};
