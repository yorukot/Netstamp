import type { Meta, StoryObj } from "@storybook/react-vite";
import { Surface, type SurfaceCut, type SurfacePadding, type SurfaceProps, type SurfaceTone } from "../index";

const tones: SurfaceTone[] = ["glass", "matte", "deep", "flat", "accent", "danger"];
const cuts: SurfaceCut[] = ["xs", "sm", "md", "lg"];
const paddings: SurfacePadding[] = ["none", "sm", "md", "lg"];

const meta = {
	title: "Components/Surface",
	args: {
		as: "div",
		children: (
			<>
				<strong>Surface content</strong>
				<p>Base primitive for cut-frame blocks, cards, and panels.</p>
			</>
		),
		cut: "md",
		padding: "md",
		tone: "glass"
	},
	argTypes: {
		as: { control: false },
		children: { control: false },
		className: { control: false },
		cut: { control: "inline-radio", options: cuts },
		padding: { control: "inline-radio", options: paddings },
		tone: { control: "select", options: tones }
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<Surface {...args} />
			</div>
		</div>
	),
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<SurfaceProps<"div">>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Tones: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				{tones.map(tone => (
					<Surface key={tone} tone={tone} cut="lg" padding="md">
						<strong>{tone}</strong>
						<p>Surface tone sample.</p>
					</Surface>
				))}
			</div>
		</div>
	)
};

export const Cuts: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid storybook-grid--compact">
				{cuts.map(cut => (
					<Surface key={cut} cut={cut} padding="md">
						<strong>{cut}</strong>
					</Surface>
				))}
			</div>
		</div>
	)
};

export const Padding: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid storybook-grid--compact">
				{paddings.map(padding => (
					<Surface key={padding} padding={padding}>
						<strong>{padding}</strong>
					</Surface>
				))}
			</div>
		</div>
	)
};
