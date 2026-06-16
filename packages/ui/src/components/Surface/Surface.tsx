import type { ComponentPropsWithoutRef, ElementType } from "react";
import styles from "./Surface.module.css";

export type SurfaceTone = "glass" | "matte" | "deep" | "flat" | "accent" | "danger";
export type SurfaceFrameSize = "xs" | "sm" | "md" | "lg";
export type SurfacePadding = "none" | "sm" | "md" | "lg";

interface SurfaceOwnProps {
	as?: ElementType;
	tone?: SurfaceTone;
	frameSize?: SurfaceFrameSize;
	padding?: SurfacePadding;
	className?: string;
}

export type SurfaceProps<T extends ElementType = "div"> = SurfaceOwnProps & Omit<ComponentPropsWithoutRef<T>, keyof SurfaceOwnProps>;

const frameSizeClasses: Record<SurfaceFrameSize, string> = {
	xs: styles.frameXs,
	sm: styles.frameSm,
	md: styles.frameMd,
	lg: styles.frameLg
};

const paddingClasses: Record<SurfacePadding, string> = {
	none: styles.paddingNone,
	sm: styles.paddingSm,
	md: styles.paddingMd,
	lg: styles.paddingLg
};

export function Surface<T extends ElementType = "div">({ as, tone = "glass", frameSize = "md", padding = "md", className, ...props }: SurfaceProps<T>) {
	const Comp = as || "div";
	const classes = ["ns-frame", styles.surface, styles[tone], frameSizeClasses[frameSize], paddingClasses[padding], className].filter(Boolean).join(" ");

	return <Comp className={classes} {...props} />;
}
