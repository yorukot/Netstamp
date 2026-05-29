import { classNames } from "@/shared/utils/classNames";
import type { ComponentPropsWithoutRef } from "react";
import styles from "./FilterGrid.module.css";

type FilterGridProps = ComponentPropsWithoutRef<"div">;

export function FilterGrid({ className, ...props }: FilterGridProps) {
	return <div className={classNames(styles.root, className)} {...props} />;
}
