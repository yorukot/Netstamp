import type { ComponentPropsWithoutRef } from "react";
import styles from "./ActionRow.module.css";

export type ActionRowProps = ComponentPropsWithoutRef<"div">;

export function ActionRow({ className, ...props }: ActionRowProps) {
	return <div className={[styles.root, className].filter(Boolean).join(" ")} {...props} />;
}
