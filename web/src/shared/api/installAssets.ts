import { absoluteApiUrl } from "./client";
import type { paths } from "./openapi";

type InstallAssetPath = Extract<keyof paths, "/install/agent.sh" | "/install/uninstall-agent.sh" | "/install/netstamp-agent-linux-amd64" | "/install/netstamp-agent-linux-arm64">;

export const installAssetPaths = {
	agentInstaller: "/install/agent.sh",
	agentUninstaller: "/install/uninstall-agent.sh",
	linuxAmd64Binary: "/install/netstamp-agent-linux-amd64",
	linuxArm64Binary: "/install/netstamp-agent-linux-arm64"
} as const satisfies Record<string, InstallAssetPath>;

export const installAssetUrl = (path: InstallAssetPath) => absoluteApiUrl(path);

export const controllerInstallTarget = () => window.location.origin;

const shellQuote = (value: string) => `'${value.replace(/'/g, `'\\''`)}'`;

export const probeInstallCommand = (input: { probeId: string; probeSecret: string }) => {
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);

	return [`curl -fsSL ${shellQuote(installerUrl)} | sudo sh`, probeServiceInstallCommand(input)].join("\n");
};

export const probeServiceInstallCommand = (input: { probeId: string; probeSecret: string }) => {
	const controllerUrl = controllerInstallTarget();

	return [
		`sudo netstamp-agent service install \\`,
		`  --url ${shellQuote(controllerUrl)} \\`,
		`  --probe-id ${shellQuote(input.probeId)} \\`,
		`  --probe-secret ${shellQuote(input.probeSecret)}`
	].join("\n");
};

export const probeSecretUpdateCommand = (input: { probeId: string; probeSecret: string }) => {
	const controllerUrl = controllerInstallTarget();

	return [
		`sudo netstamp-agent service install \\`,
		`  --url ${shellQuote(controllerUrl)} \\`,
		`  --probe-id ${shellQuote(input.probeId)} \\`,
		`  --probe-secret ${shellQuote(input.probeSecret)} && \\`,
		`sudo systemctl restart netstamp-agent`
	].join("\n");
};

export const probeReinstallCommand = () => {
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);

	return [`curl -fsSL ${shellQuote(installerUrl)} | sudo sh && \\`, `sudo systemctl restart netstamp-agent`].join("\n");
};

export const probeUpgradeCommand = () => {
	const controllerUrl = controllerInstallTarget();

	return [`sudo netstamp-agent update \\`, `  --url ${shellQuote(controllerUrl)}`].join("\n");
};
