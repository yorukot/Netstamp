import type { ComponentPropsWithoutRef } from "react";
import styles from "./FilterGrid.module.css";

export type FilterGridProps = ComponentPropsWithoutRef<"div">;

export function FilterGrid({ className, ...props }: FilterGridProps) {
	return <div className={[styles.root, className].filter(Boolean).join(" ")} {...props} />;
}
