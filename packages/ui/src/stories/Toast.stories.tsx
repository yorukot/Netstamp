import type { Meta, StoryObj } from "@storybook/react-vite";
import { Toast, ToastClose, ToastDescription, ToastTitle, ToastViewport, type ToastTone } from "../index";

const examples: Array<{ tone: ToastTone; title: string; message: string }> = [
	{ tone: "success", title: "Probe saved", message: "tpe-edge-01 is now assigned to the ICMP and DNS checks." },
	{ tone: "critical", title: "Request failed", message: "The controller rejected the update because the probe secret has rotated." },
	{ tone: "warning", title: "Route degraded", message: "Packet loss crossed the warning threshold for ap-east." },
	{ tone: "neutral", title: "Export queued", message: "A project export is being prepared in the background." }
];

const meta = {
	title: "Components/Toast",
	component: Toast,
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Toast>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Tones: Story = {
	render: () => (
		<div className="storybook-canvas">
			<ToastViewport className="storybook-toast-viewport">
				{examples.map(example => (
					<Toast key={example.tone} tone={example.tone}>
						<div>
							<ToastTitle>{example.title}</ToastTitle>
							<ToastDescription>{example.message}</ToastDescription>
						</div>
						<ToastClose />
					</Toast>
				))}
			</ToastViewport>
		</div>
	)
};

export const Stacked: Story = {
	render: () => (
		<div className="storybook-canvas">
			<ToastViewport className="storybook-toast-viewport">
				<Toast tone="critical">
					<div>
						<ToastTitle>Notification failed</ToastTitle>
						<ToastDescription>Webhook delivery returned 500 for incident netstamp-alert-418.</ToastDescription>
					</div>
					<ToastClose />
				</Toast>
				<Toast tone="success">
					<div>
						<ToastTitle>Check duplicated</ToastTitle>
						<ToastDescription>The new draft is open. Save it to add the check to this project.</ToastDescription>
					</div>
					<ToastClose />
				</Toast>
			</ToastViewport>
		</div>
	)
};

export const LongMessage: Story = {
	render: () => (
		<div className="storybook-canvas">
			<ToastViewport className="storybook-toast-viewport">
				<Toast tone="success">
					<div>
						<ToastTitle>Status page updated</ToastTitle>
						<ToastDescription>Public status page settings were saved, including chart range, compact display mode, and the selected probe assignment groups.</ToastDescription>
					</div>
					<ToastClose />
				</Toast>
			</ToastViewport>
		</div>
	)
};
