import type { Meta, StoryObj } from "@storybook/react-vite";
import { Surface, type SurfaceFrameSize, type SurfacePadding, type SurfaceProps, type SurfaceTone } from "../index";

const tones: SurfaceTone[] = ["glass", "matte", "deep", "flat", "accent", "danger"];
const frameSizes: SurfaceFrameSize[] = ["xs", "sm", "md", "lg"];
const paddings: SurfacePadding[] = ["none", "sm", "md", "lg"];

const meta = {
	title: "Components/Surface",
	args: {
		as: "div",
		children: (
			<>
				<strong>Surface content</strong>
				<p>Base primitive for square dashboard frames, cards, and panels.</p>
			</>
		),
		frameSize: "md",
		padding: "md",
		tone: "glass"
	},
	argTypes: {
		as: { control: false },
		children: { control: false },
		className: { control: false },
		frameSize: { control: "inline-radio", options: frameSizes },
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
					<Surface key={tone} tone={tone} frameSize="lg" padding="md">
						<strong>{tone}</strong>
						<p>Surface tone sample.</p>
					</Surface>
				))}
			</div>
		</div>
	)
};

export const FrameSizes: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid storybook-grid--compact">
				{frameSizes.map(frameSize => (
					<Surface key={frameSize} frameSize={frameSize} padding="md">
						<strong>{frameSize}</strong>
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
