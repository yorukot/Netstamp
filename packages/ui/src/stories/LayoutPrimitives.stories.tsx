import { XIcon } from "@phosphor-icons/react/dist/csr/X";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { ActionRow, BodyCopy, Button, FilterGrid, IconButton, Input, Select, Spinner } from "../index";

const meta: Meta = {
	title: "Components/Layout Primitives"
};

export default meta;

type Story = StoryObj;

export const Overview: Story = {
	render: () => (
		<div style={{ display: "grid", gap: "1.5rem", maxWidth: "48rem" }}>
			<BodyCopy>Reusable layout primitives for dense app screens.</BodyCopy>
			<FilterGrid>
				<Input aria-label="Search" placeholder="Search checks" />
				<Select aria-label="Status" defaultValue="all">
					<option value="all">All statuses</option>
					<option value="active">Active</option>
				</Select>
				<Button variant="secondary">
					<span aria-hidden="true">/</span>
					Apply
				</Button>
			</FilterGrid>
			<ActionRow>
				<Button>Save</Button>
				<Button variant="ghost">Cancel</Button>
				<IconButton aria-label="Delete" danger>
					<XIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
				</IconButton>
			</ActionRow>
			<Spinner label="Loading panel" layout="compact" size="lg" />
		</div>
	)
};
