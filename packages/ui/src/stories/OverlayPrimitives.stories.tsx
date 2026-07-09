import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import {
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogOverlay,
	AlertDialogPortal,
	AlertDialogRoot,
	AlertDialogTitle,
	AlertDialogTrigger,
	Button,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogOverlay,
	DialogPortal,
	DialogRoot,
	DialogTitle,
	DialogTrigger,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuPortal,
	DropdownMenuRoot,
	DropdownMenuTrigger,
	Panel,
	PopoverAnchor,
	PopoverClose,
	PopoverContent,
	PopoverPortal,
	PopoverRoot,
	PopoverTrigger
} from "../index";

const meta = {
	title: "Components/Overlay primitives",
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const Dialog: Story = {
	render: () => (
		<div className="storybook-canvas">
			<DialogRoot>
				<DialogTrigger asChild>
					<Button type="button">Open dialog</Button>
				</DialogTrigger>
				<DialogPortal>
					<DialogOverlay />
					<DialogContent>
						<DialogTitle className="ns-title storybook-overlay-title">Probe maintenance window</DialogTitle>
						<DialogDescription className="storybook-overlay-description">Confirm the maintenance note before muting probe alerts for the selected route.</DialogDescription>
						<Panel tone="matte" title="Affected probe" className="storybook-overlay-panel">
							<p>tpe-edge-01 / ICMP, DNS, TCP route checks</p>
						</Panel>
						<div className="storybook-overlay-actions">
							<DialogClose asChild>
								<Button type="button" variant="ghost">
									Cancel
								</Button>
							</DialogClose>
							<DialogClose asChild>
								<Button type="button">Apply window</Button>
							</DialogClose>
						</div>
					</DialogContent>
				</DialogPortal>
			</DialogRoot>
		</div>
	)
};

export const AlertDialog: Story = {
	render: () => (
		<div className="storybook-canvas">
			<AlertDialogRoot>
				<AlertDialogTrigger asChild>
					<Button type="button" variant="danger">
						Delete check
					</Button>
				</AlertDialogTrigger>
				<AlertDialogPortal>
					<AlertDialogOverlay />
					<AlertDialogContent>
						<AlertDialogTitle className="ns-title storybook-overlay-title">Delete HTTP check?</AlertDialogTitle>
						<AlertDialogDescription className="storybook-overlay-description">This removes the check configuration and stops new measurements for api.netstamp.local.</AlertDialogDescription>
						<div className="storybook-overlay-actions">
							<AlertDialogCancel asChild>
								<Button type="button" variant="outline">
									Keep check
								</Button>
							</AlertDialogCancel>
							<AlertDialogAction asChild>
								<Button type="button" variant="danger">
									Delete check
								</Button>
							</AlertDialogAction>
						</div>
					</AlertDialogContent>
				</AlertDialogPortal>
			</AlertDialogRoot>
		</div>
	)
};

export const DropdownMenu: Story = {
	render: function Render() {
		const [streamEnabled, setStreamEnabled] = useState(true);
		const [routePinned, setRoutePinned] = useState(false);

		return (
			<div className="storybook-canvas">
				<DropdownMenuRoot>
					<DropdownMenuTrigger asChild>
						<Button type="button" variant="secondary">
							Open menu
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuPortal>
						<DropdownMenuContent className="storybook-menu" align="start" sideOffset={8}>
							<DropdownMenuItem className="storybook-menu-item">Open probe detail</DropdownMenuItem>
							<DropdownMenuItem className="storybook-menu-item">Copy probe ID</DropdownMenuItem>
							<DropdownMenuCheckboxItem className="storybook-menu-item storybook-menu-item--checkbox" checked={streamEnabled} onCheckedChange={setStreamEnabled}>
								Stream results
							</DropdownMenuCheckboxItem>
							<DropdownMenuCheckboxItem className="storybook-menu-item storybook-menu-item--checkbox" checked={routePinned} onCheckedChange={setRoutePinned}>
								Pin route
							</DropdownMenuCheckboxItem>
						</DropdownMenuContent>
					</DropdownMenuPortal>
				</DropdownMenuRoot>
			</div>
		);
	}
};

export const Popover: Story = {
	render: () => (
		<div className="storybook-canvas">
			<PopoverRoot>
				<PopoverAnchor asChild>
					<div className="storybook-popover-anchor">
						<span>probe:tpe-edge-01</span>
						<PopoverTrigger asChild>
							<Button type="button" variant="outline" size="sm">
								Inspect
							</Button>
						</PopoverTrigger>
					</div>
				</PopoverAnchor>
				<PopoverPortal>
					<PopoverContent className="storybook-popover" align="center" sideOffset={10} collisionPadding={12}>
						<strong>Route hash changed</strong>
						<p>The latest traceroute changed from previous run after hop 6.</p>
						<PopoverClose asChild>
							<Button type="button" variant="ghost" size="sm">
								Close
							</Button>
						</PopoverClose>
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>
		</div>
	)
};
