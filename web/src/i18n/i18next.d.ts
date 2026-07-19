import "i18next";
import type { AppResources } from "./resources";

declare module "i18next" {
	interface CustomTypeOptions {
		defaultNS: "common";
		resources: AppResources;
		returnNull: false;
	}
}
