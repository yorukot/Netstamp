/// <reference types="astro/client" />

import type { HTMLAttributes } from "react";

type PhosphorIconProps = HTMLAttributes<HTMLElement> & {
	color?: string;
	mirrored?: boolean | string;
	size?: number | string;
	weight?: "thin" | "light" | "regular" | "bold" | "fill" | "duotone";
};

declare module "react" {
	namespace JSX {
		interface IntrinsicElements {
			[tagName: `ph-${string}`]: PhosphorIconProps;
		}
	}
}
