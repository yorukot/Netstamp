/// <reference types="astro/client" />

import type { HTMLAttributes } from "react";

declare global {
	interface ImportMetaEnv {
		readonly PUBLIC_NETSTAMP_APP_BASE_URL?: string;
		readonly PUBLIC_NETSTAMP_UMAMI_WEBSITE_ID?: string;
		readonly PUBLIC_NETSTAMP_UMAMI_SCRIPT_URL?: string;
		readonly PUBLIC_NETSTAMP_TRACKING_CONSENT_MODE?: "regional" | "always" | "immediate";
		readonly PUBLIC_NETSTAMP_TRACKING_CONSENT_COUNTRIES?: string;
	}
}

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
