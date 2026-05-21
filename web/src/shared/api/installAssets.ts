import { absoluteApiUrl, apiBaseUrl } from "./client";
import type { paths } from "./openapi";

type InstallAssetPath = Extract<keyof paths, "/install/agent.sh" | "/install/uninstall-agent.sh" | "/install/netstamp-agent-linux-amd64">;

export const installAssetPaths = {
	agentInstaller: "/install/agent.sh",
	agentUninstaller: "/install/uninstall-agent.sh",
	linuxAmd64Binary: "/install/netstamp-agent-linux-amd64"
} as const satisfies Record<string, InstallAssetPath>;

export function installAssetUrl(path: InstallAssetPath) {
	return absoluteApiUrl(path);
}

export function controllerInstallTarget() {
	const url = new URL(apiBaseUrl, window.location.origin);
	const match = url.pathname.match(/^(?<prefix>.*)\/api\/(?<version>[^/]+)\/?$/);
	const prefix = match?.groups?.prefix || "";

	url.pathname = prefix || "/";
	url.search = "";
	url.hash = "";

	return {
		apiVersion: match?.groups?.version || "v1",
		controllerUrl: url.toString().replace(/\/$/, "")
	};
}

function shellQuote(value: string) {
	return `'${value.replace(/'/g, `'\\''`)}'`;
}

export function probeInstallCommand(input: { probeId: string; probeSecret: string }) {
	const { apiVersion, controllerUrl } = controllerInstallTarget();
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);

	return [
		`curl -fsSL ${shellQuote(installerUrl)} | sudo sh -s -- \\`,
		`  --controller-url ${shellQuote(controllerUrl)} \\`,
		`  --probe-id ${shellQuote(input.probeId)} \\`,
		`  --probe-secret ${shellQuote(input.probeSecret)} \\`,
		`  --api-version ${shellQuote(apiVersion)}`
	].join("\n");
}
