import { classNames } from "@/shared/utils/classNames";
import { useTranslation } from "react-i18next";
import styles from "./NewProbeDrawer.module.css";

interface ProbeWizardStep {
	number: string;
	title: string;
	copy: string;
}

interface ProbeWizardTimelineProps {
	steps: ProbeWizardStep[];
	currentStep: number;
}

export function ProbeWizardTimeline({ steps, currentStep }: ProbeWizardTimelineProps) {
	const { t } = useTranslation("probes");
	return (
		<ol className={styles.stepTimeline} aria-label={t("progressAria")}>
			{steps.map((step, index) => (
				<li className={classNames(styles.stepItem, index === currentStep && styles.stepActive, index < currentStep && styles.stepComplete)} key={step.number}>
					<span>{step.number}</span>
					<strong>{step.title}</strong>
					<small>{step.copy}</small>
				</li>
			))}
		</ol>
	);
}
