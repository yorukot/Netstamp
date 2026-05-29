import js from "@eslint/js";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import tseslint from "typescript-eslint";

export default defineConfig([
	globalIgnores(["dist", "node_modules"]),
	{
		files: ["**/*.{js,jsx,ts,tsx}"],
		extends: [js.configs.recommended, ...tseslint.configs.recommended, reactHooks.configs.flat.recommended, reactRefresh.configs.vite],
		languageOptions: {
			globals: globals.browser,
			parserOptions: { ecmaFeatures: { jsx: true } }
		},
		rules: {
			"no-restricted-globals": [
				"error",
				{ name: "alert", message: "Use pushToast, pushErrorToast, or useAlertDialog instead of browser alert()." },
				{ name: "confirm", message: "Use useConfirm() instead of browser confirm()." },
				{ name: "prompt", message: "Use usePromptDialog() instead of browser prompt()." }
			],
			"no-restricted-properties": [
				"error",
				{ object: "globalThis", property: "alert", message: "Use pushToast, pushErrorToast, or useAlertDialog instead." },
				{ object: "globalThis", property: "confirm", message: "Use useConfirm() instead." },
				{ object: "globalThis", property: "prompt", message: "Use usePromptDialog() instead." },
				{ object: "self", property: "alert", message: "Use pushToast, pushErrorToast, or useAlertDialog instead." },
				{ object: "self", property: "confirm", message: "Use useConfirm() instead." },
				{ object: "self", property: "prompt", message: "Use usePromptDialog() instead." },
				{ object: "window", property: "alert", message: "Use pushToast, pushErrorToast, or useAlertDialog instead." },
				{ object: "window", property: "confirm", message: "Use useConfirm() instead." },
				{ object: "window", property: "prompt", message: "Use usePromptDialog() instead." }
			]
		}
	}
]);
