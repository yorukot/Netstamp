import { Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";

export const RouteLoadingSpinner = () => {
	const { t } = useTranslation("navigation");

	return <Spinner label={t("loadingRoute")} layout="page" size="lg" />;
};
