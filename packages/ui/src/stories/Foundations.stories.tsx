import type { Meta, StoryObj } from "@storybook/react-vite";
import { CodePreview, KeyValueRow, SpecLabel, Surface } from "../index";

const defaultTokens = [
	["Canvas", "--ns-bg", "#000000"],
	["Surface", "--ns-surface", "#070707"],
	["Raised", "--ns-surface-raised", "#131518"],
	["Deep", "--ns-surface-deep", "#000000"],
	["Text", "--ns-text", "#f8fafc"],
	["Muted text", "--ns-text-muted", "#d8dce8"],
	["Primary orange", "--ns-primary", "#fb923c"],
	["Secondary blue", "--ns-secondary", "#38bdf8"],
	["Success", "--ns-success", "#34c77b"],
	["Warning", "--ns-warning", "#facc15"],
	["Critical", "--ns-critical", "#ff6b63"],
	["Border", "--ns-border", "#2d3035"]
] as const;

const typeRows = [
	["Display", "--ns-font-display", "Route titles, hero statements, large values"],
	["Sans", "--ns-font-sans", "Body, labels, controls, navigation, table cells"],
	["Mono", "--ns-font-mono", "Code, paths, timestamps, methods, technical metadata"]
] as const;

const meta = {
	title: "Foundations/Tokens",
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const ColorRoles: Story = {
	render: () => (
		<div className="storybook-page">
			<section className="storybook-section">
				<div className="storybook-section-header">
					<SpecLabel tone="primary">Foundations</SpecLabel>
					<h2>Color roles</h2>
					<p>Use orange for primary interaction, blue for secondary/reference data, and semantic colors only for state.</p>
				</div>
				<div className="storybook-token-grid">
					{defaultTokens.map(([label, token, value]) => (
						<Surface key={token} tone="glass" frameSize="lg" padding="sm" className="storybook-token-card">
							<span className="storybook-swatch" style={{ background: `var(${token})` }} />
							<strong>{label}</strong>
							<code>{token}</code>
							<small>{value}</small>
						</Surface>
					))}
				</div>
			</section>
		</div>
	)
};

export const Typography: Story = {
	render: () => (
		<div className="storybook-page">
			<section className="storybook-section">
				<div className="storybook-section-header">
					<SpecLabel tone="secondary">Typography</SpecLabel>
					<h2>Font roles</h2>
					<p>The existing Netstamp fonts stay in place. The system changes how strictly each face is assigned.</p>
				</div>
				<Surface tone="glass" frameSize="lg" padding="lg">
					<div className="storybook-type-specimen">
						<div>
							<SpecLabel>Display</SpecLabel>
							<strong className="storybook-type-display">Network state at controller scale.</strong>
						</div>
						<div>
							<SpecLabel>Sans</SpecLabel>
							<p className="storybook-type-sans">Dense operational interfaces use sans text for labels, buttons, tables, and prose.</p>
						</div>
						<div>
							<SpecLabel>Mono</SpecLabel>
							<code className="storybook-type-mono">GET /api/v1/projects/:projectRef/checks</code>
						</div>
					</div>
				</Surface>
				<div className="storybook-demo">
					{typeRows.map(([label, token, usage]) => (
						<KeyValueRow key={token} label={label} value={token} meta={usage} />
					))}
				</div>
			</section>
		</div>
	)
};

export const ImplementationRules: Story = {
	render: () => (
		<div className="storybook-page">
			<section className="storybook-section">
				<div className="storybook-section-header">
					<SpecLabel tone="primary">Rules</SpecLabel>
					<h2>Spec-docs constraints</h2>
					<p>These rules keep the product, docs, API explorer, and Storybook on the same visual contract.</p>
				</div>
				<div className="storybook-grid">
					<Surface tone="glass" frameSize="lg" padding="md">
						<KeyValueRow label="radius" value="0" meta="core controls" tone="primary" />
						<KeyValueRow label="shadow" value="none" meta="use borders" />
						<KeyValueRow label="gradient" value="none" meta="flat tokens" />
					</Surface>
					<Surface tone="glass" frameSize="lg" padding="md">
						<KeyValueRow label="primary" value="orange" meta="actions" tone="primary" />
						<KeyValueRow label="secondary" value="blue" meta="links/data" tone="secondary" />
						<KeyValueRow label="state" value="semantic" meta="health only" tone="success" />
					</Surface>
				</div>
				<CodePreview title="focus contract" meta="css">
					{`:focus-visible {
	outline: var(--ns-focus-outline);
	outline-offset: var(--ns-focus-outline-offset);
}`}
				</CodePreview>
			</section>
		</div>
	)
};
