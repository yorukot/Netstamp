import type { ComponentPropsWithoutRef, ElementType } from "react";
import styles from "./BodyCopy.module.css";

interface BodyCopyOwnProps {
	as?: ElementType;
	className?: string;
}

export type BodyCopyProps<T extends ElementType = "p"> = BodyCopyOwnProps & Omit<ComponentPropsWithoutRef<T>, keyof BodyCopyOwnProps>;

export function BodyCopy<T extends ElementType = "p">({ as, className, ...props }: BodyCopyProps<T>) {
	const Comp = as || "p";

	return <Comp className={[styles.copy, className].filter(Boolean).join(" ")} {...props} />;
}
