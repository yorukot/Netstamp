import type { ComponentPropsWithoutRef } from "react";
import styles from "./ResponsiveGrid.module.css";

interface ResponsiveGridProps extends ComponentPropsWithoutRef<"div"> {
	columns?: "two" | "three";
	collapseAt?: "md" | "lg";
}

export function ResponsiveGrid({ columns = "two", collapseAt = "md", className, ...props }: ResponsiveGridProps) {
	return <div className={[styles.grid, styles[columns], styles[collapseAt], className].filter(Boolean).join(" ")} {...props} />;
}
