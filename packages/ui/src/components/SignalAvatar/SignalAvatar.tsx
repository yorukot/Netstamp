import type { ComponentPropsWithoutRef, ImgHTMLAttributes } from "react";
import styles from "./SignalAvatar.module.css";

export type SignalAvatarSize = "sm" | "md" | "lg";

export interface SignalAvatarProps extends ComponentPropsWithoutRef<"span"> {
	src: string;
	alt?: string;
	size?: SignalAvatarSize;
	referrerPolicy?: ImgHTMLAttributes<HTMLImageElement>["referrerPolicy"];
}

export function SignalAvatar({ src, alt = "", size = "md", referrerPolicy, className, ...props }: SignalAvatarProps) {
	const classes = ["ns-cut-frame", styles.avatar, styles[size], className].filter(Boolean).join(" ");

	return (
		<span className={classes} {...props}>
			<img src={src} alt={alt} referrerPolicy={referrerPolicy} />
		</span>
	);
}
