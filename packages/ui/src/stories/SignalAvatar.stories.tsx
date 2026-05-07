import netstampMark from "@netstamp/brand/assets/netstamp-mark-light.svg";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { SignalAvatar, type SignalAvatarSize } from "../index";

const sizes: SignalAvatarSize[] = ["sm", "md", "lg"];

const meta = {
	title: "Components/SignalAvatar",
	component: SignalAvatar,
	args: {
		alt: "Netstamp mark",
		size: "md",
		src: netstampMark
	},
	argTypes: {
		size: { control: "inline-radio", options: sizes },
		src: { control: "text" }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof SignalAvatar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<SignalAvatar {...args} />
		</div>
	)
};

export const Sizes: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-row">
				{sizes.map(size => (
					<SignalAvatar key={size} src={netstampMark} alt="Netstamp mark" size={size} />
				))}
			</div>
		</div>
	)
};
