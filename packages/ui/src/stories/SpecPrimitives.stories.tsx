import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, CodeBlock, DisclosureToggle, EmptyState, KeyValueRow, MetricTile, SelectableRow, SpecCard, SpecLabel } from "../index";

const meta = {
	title: "Components/Spec primitives",
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const LabelsRowsAndToggles: Story = {
	render: () => (
		<div className="storybook-canvas storybook-canvas--top">
			<div className="storybook-demo storybook-demo--wide">
				<div className="storybook-section-header">
					<SpecLabel tone="primary">Spec primitives</SpecLabel>
					<h2>Rows, labels, and disclosure controls</h2>
					<p>Use these for docs navigation, API endpoint lists, search results, and dense operation rows.</p>
				</div>
				<div className="storybook-row">
					<SpecLabel tone="primary">Primary</SpecLabel>
					<SpecLabel tone="secondary">Secondary</SpecLabel>
					<SpecLabel tone="success">Healthy</SpecLabel>
					<SpecLabel tone="warning">Degraded</SpecLabel>
					<SpecLabel tone="critical">Failed</SpecLabel>
				</div>
				<div className="storybook-spec-list">
					<SelectableRow active tone="primary" title="GET /api/v1/projects/{projectRef}/checks" description="List checks and latest state for a project." meta="GET" />
					<SelectableRow title="POST /api/v1/probes" description="Register a probe with controller-issued credentials." meta="POST" />
					<SelectableRow disabled title="DELETE /api/v1/status-pages/{id}" description="Disabled until status-page ownership is confirmed." meta="DELETE" />
				</div>
				<div className="storybook-row">
					<DisclosureToggle open={false} label="Open section" />
					<DisclosureToggle open label="Close section" />
					<DisclosureToggle open={false} label="Open large section" size="md" />
				</div>
			</div>
		</div>
	)
};

export const CardsMetricsAndRows: Story = {
	render: () => (
		<div className="storybook-canvas storybook-canvas--top">
			<div className="storybook-demo storybook-demo--wide">
				<div className="storybook-card-grid">
					<SpecCard eyebrow="Guide" title="Probe operations" description="Register agents, rotate tokens, and verify heartbeat freshness." meta="docs/reference" active />
					<SpecCard eyebrow="API" title="Checks endpoint" description="Create ping, DNS, TCP, and traceroute checks from automation." actions={<Button size="sm">Open API</Button>} />
					<SpecCard eyebrow="State" title="Incident routing" description="Status pages and alert rules share the same state vocabulary." tone="matte" />
				</div>
				<div className="storybook-grid">
					<MetricTile label="active probes" value="38" detail="online" tone="success" trend="+4" />
					<MetricTile label="p95 latency" value="82ms" description="Across regional probes" detail="watch" tone="warning" />
					<MetricTile label="route changes" value="7" description="Last 24 hours" detail="changed" tone="accent" />
				</div>
				<div>
					<KeyValueRow label="controller" value="us-east-1 / controller-02" meta="primary" tone="primary" />
					<KeyValueRow label="heartbeat" value="2026-06-26 18:42:10 UTC" meta="fresh" tone="success" />
					<KeyValueRow label="route hash" value="b94c.22f9.changed" meta="diff" tone="warning" />
				</div>
				<EmptyState title="No checks configured" description="Create a check to start collecting probe measurements." action={<Button size="sm">Create check</Button>} />
			</div>
		</div>
	)
};

export const CodeBlocks: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<CodeBlock title="install probe" meta="shell">
					{`curl -fsSL https://netstamp.dev/install.sh | sh
netstamp probe join --controller https://api.netstamp.dev --token ns_probe_***`}
				</CodeBlock>
				<CodeBlock title="api request" meta="curl">
					{`curl -H "Authorization: Bearer $NETSTAMP_TOKEN" \\
  https://api.netstamp.dev/v1/projects/edge/checks`}
				</CodeBlock>
			</div>
		</div>
	)
};
