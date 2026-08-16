import { execFileSync } from "node:child_process";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import { fixtureResponse, ids } from "./fixtures.mjs";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const baseUrl = (process.env.DOCS_SCREENSHOT_BASE_URL || "https://app.netstamp.dev").replace(/\/$/, "");
const docsOutputRoot = path.join(repositoryRoot, "docs/src/assets/screenshots");
const homepageOutputRoot = path.join(repositoryRoot, "docs/src/assets/homepage-highlights");
const artifactRoot = path.join(repositoryRoot, "output/playwright/docs-screenshots");
const auditOnly = process.argv.includes("--audit");
const catalogOnly = process.argv.includes("--catalog");
const firstProjectOnly = process.argv.includes("--only=create-project");
const fixedNow = Date.parse("2026-08-13T03:20:00.000Z");
const allowedReadMethods = new Set(["GET", "HEAD", "OPTIONS"]);
const leakedWrites = [];
const unhandledApiRequests = [];
const capturedFiles = [];

const ensureDirectories = async () => {
	await Promise.all([mkdir(docsOutputRoot, { recursive: true }), mkdir(homepageOutputRoot, { recursive: true }), mkdir(artifactRoot, { recursive: true })]);
};

const installDeterministicBrowserState = async context => {
	await context.addInitScript(
		({ currentTime }) => {
			const NativeDate = Date;
			class FixedDate extends NativeDate {
				constructor(...args) {
					super(...(args.length ? args : [currentTime]));
				}

				static now() {
					return currentTime;
				}
			}
			window.Date = FixedDate;
			window.localStorage.setItem("netstamp:locale", "en");
			window.localStorage.setItem("netstamp:theme", "light");
			window.localStorage.setItem("netstamp:selected-project-ref", "docs-demo");
			window.localStorage.setItem("netstamp:sidebar-collapsed", "false");
		},
		{ currentTime: fixedNow }
	);
};

const installApiFixtures = async context => {
	await context.route("**/*", async route => {
		const request = route.request();
		const requestUrl = new URL(request.url());
		const method = request.method().toUpperCase();
		if (requestUrl.pathname === "/cdn-cgi/rum") {
			await route.abort("blockedbyclient");
			return;
		}

		if (requestUrl.pathname.startsWith("/api/v1/")) {
			const response = fixtureResponse(request.url(), method);
			if (response) {
				await route.fulfill({
					status: response.status,
					contentType: "application/json",
					headers: { "Cache-Control": "no-store" },
					body: JSON.stringify(response.body)
				});
				return;
			}

			if (!allowedReadMethods.has(method)) {
				leakedWrites.push(`${method} ${requestUrl.pathname}`);
				await route.abort("blockedbyclient");
				return;
			}

			unhandledApiRequests.push(`${method} ${requestUrl.pathname}`);
			await route.fulfill({ status: 501, contentType: "application/problem+json", body: JSON.stringify({ title: "Missing screenshot fixture", status: 501 }) });
			return;
		}

		if (!allowedReadMethods.has(method)) {
			leakedWrites.push(`${method} ${requestUrl.origin}${requestUrl.pathname}`);
			await route.abort("blockedbyclient");
			return;
		}

		await route.continue();
	});
};

const settle = async page => {
	await page.waitForLoadState("domcontentloaded");
	await page
		.locator("h1")
		.first()
		.waitFor({ state: "visible", timeout: 15000 })
		.catch(() => undefined);
	await page.waitForTimeout(900);
	await page.evaluate(() => document.fonts.ready);
	await page.addStyleTag({
		content: `
			*, *::before, *::after {
				animation-duration: 0s !important;
				animation-delay: 0s !important;
				transition-duration: 0s !important;
				caret-color: transparent !important;
			}
			::-webkit-scrollbar { width: 0 !important; height: 0 !important; }
		`
	});
};

const navigate = async (page, routePath) => {
	await page.goto(`${baseUrl}${routePath}`, { waitUntil: "domcontentloaded", timeout: 30000 });
	await settle(page);
};

const toWebP = async (pngPath, webpPath, width, height) => {
	execFileSync("cwebp", ["-quiet", "-mt", "-q", "88", "-metadata", "none", "-resize", String(width), String(height), pngPath, "-o", webpPath]);
	await rm(pngPath);
};

const captureViewport = async (page, outputPath, { width = 1440, height = 1000, homepage = false } = {}) => {
	await page.setViewportSize({ width, height });
	await page.waitForTimeout(250);
	const pngPath = path.join(artifactRoot, `${path.basename(outputPath, path.extname(outputPath))}.png`);
	await page.screenshot({ path: pngPath, fullPage: false });
	await mkdir(path.dirname(outputPath), { recursive: true });
	await toWebP(pngPath, outputPath, homepage ? 1200 : width * 2, homepage ? 780 : height * 2);
	capturedFiles.push(path.relative(repositoryRoot, outputPath));
};

const captureElement = async (page, outputPath, selector = "main", { maxHeight = 1050 } = {}) => {
	const locator = page.locator(selector).first();
	await locator.waitFor({ state: "visible", timeout: 15000 });
	const box = await locator.boundingBox();
	if (!box) throw new Error(`Cannot capture ${selector}: no bounding box`);
	const viewport = page.viewportSize();
	const clip = {
		x: Math.max(0, box.x),
		y: Math.max(0, box.y),
		width: Math.min(box.width, viewport.width - Math.max(0, box.x)),
		height: Math.min(box.height, maxHeight, viewport.height - Math.max(0, box.y))
	};
	const pngPath = path.join(artifactRoot, `${path.basename(outputPath, path.extname(outputPath))}.png`);
	await page.screenshot({ path: pngPath, clip });
	await mkdir(path.dirname(outputPath), { recursive: true });
	await toWebP(pngPath, outputPath, Math.round(clip.width * 2), Math.round(clip.height * 2));
	capturedFiles.push(path.relative(repositoryRoot, outputPath));
};

const captureLocator = async (page, relativePath, locator) => {
	await page.setViewportSize({ width: 1440, height: 1000 });
	await locator.waitFor({ state: "visible", timeout: 15000 });
	await locator.scrollIntoViewIfNeeded();
	await page.waitForTimeout(250);
	const box = await locator.boundingBox();
	if (!box) throw new Error(`Cannot capture ${relativePath}: no bounding box`);
	const outputPath = path.join(docsOutputRoot, relativePath);
	const pngPath = path.join(artifactRoot, `${path.basename(outputPath, path.extname(outputPath))}.png`);
	await locator.screenshot({ path: pngPath });
	await mkdir(path.dirname(outputPath), { recursive: true });
	await toWebP(pngPath, outputPath, Math.round(box.width * 2), Math.round(box.height * 2));
	capturedFiles.push(path.relative(repositoryRoot, outputPath));
};

const clickFirst = async (page, candidates) => {
	for (const candidate of candidates) {
		const locator = typeof candidate === "string" ? page.getByRole("button", { name: candidate, exact: true }) : candidate;
		if ((await locator.count()) && (await locator.first().isVisible())) {
			await locator.first().click();
			await page.waitForTimeout(250);
			return true;
		}
	}
	return false;
};

const fillFirst = async (page, labels, value) => {
	for (const labelText of labels) {
		const field = page.getByLabel(labelText, { exact: false });
		if ((await field.count()) && (await field.first().isVisible())) {
			await field.first().fill(value);
			return true;
		}
	}
	return false;
};

const chooseFirst = async (page, labels, optionName) => {
	for (const labelText of labels) {
		const field = page.getByLabel(labelText, { exact: false });
		if ((await field.count()) && (await field.first().isVisible())) {
			await field.first().click();
			const option = page.getByRole("option", { name: optionName, exact: true });
			if (await option.count()) {
				await option.first().click();
				await page.waitForTimeout(150);
				return true;
			}
			await page.keyboard.press("Escape");
		}
	}
	return false;
};

const scrollToText = async (page, text, { exact = true } = {}) => {
	const candidates = [page.getByRole("heading", { name: text, exact }), page.getByText(text, { exact })];
	for (const candidate of candidates) {
		if (await candidate.count()) {
			await candidate.first().scrollIntoViewIfNeeded();
			await page.waitForTimeout(250);
			return true;
		}
	}
	return false;
};

const scrollTextToTop = async (page, text, { exact = true } = {}) => {
	const candidates = [page.getByRole("heading", { name: text, exact }), page.getByText(text, { exact })];
	for (const candidate of candidates) {
		if (await candidate.count()) {
			await candidate.first().evaluate(element => element.scrollIntoView({ block: "start" }));
			await page.evaluate(() => window.scrollBy(0, -24));
			await page.waitForTimeout(250);
			return true;
		}
	}
	return false;
};

const captureDoc = async (page, relativePath, options) => {
	await page.setViewportSize({ width: 1440, height: 1000 });
	await page.waitForTimeout(200);
	await captureElement(page, path.join(docsOutputRoot, relativePath), "main", options);
};

const captureHomepage = async (page, filename) => {
	await captureViewport(page, path.join(homepageOutputRoot, filename), { width: 1200, height: 780, homepage: true });
};

const openNewCheck = async (page, type) => {
	await navigate(page, "/projects/docs-demo/checks");
	if (!(await clickFirst(page, ["New check"]))) throw new Error("New check button was not found.");
	await page
		.getByRole("heading", { name: "Check", exact: true })
		.waitFor({ state: "visible", timeout: 5000 })
		.catch(() => undefined);
	if (type && type !== "Ping") await chooseFirst(page, ["Check type"], type);
	return page;
};

const fillCheckIdentity = async (page, { name, target }) => {
	await fillFirst(page, ["Check name"], name);
	await fillFirst(page, ["Target"], target);
};

const captureFirstProject = async page => {
	const projectsListRoute = url => url.pathname === "/api/v1/projects";
	const emptyProjectsHandler = async route => {
		if (route.request().method() !== "GET") {
			await route.fallback();
			return;
		}

		await route.fulfill({
			status: 200,
			contentType: "application/json",
			headers: { "Cache-Control": "no-store" },
			body: JSON.stringify({ projects: [] })
		});
	};

	await page.route(projectsListRoute, emptyProjectsHandler);
	try {
		await navigate(page, "/onboarding");
		await page.getByRole("heading", { name: "Create your first project", exact: true }).waitFor({ state: "visible", timeout: 8000 });
		await fillFirst(page, ["Project name"], "Home Lab");
		await captureViewport(page, path.join(docsOutputRoot, "getting-started/create-project.webp"), { width: 1200, height: 1000 });
	} finally {
		await page.unroute(projectsListRoute, emptyProjectsHandler);
	}
};

const audit = async page => {
	const auditRoute = process.env.DOCS_SCREENSHOT_AUDIT_ROUTE || "/projects/docs-demo/dashboard";
	await navigate(page, auditRoute);
	await page.screenshot({ path: path.join(artifactRoot, "audit-dashboard.png"), fullPage: false });
	const report = {
		route: auditRoute,
		url: page.url(),
		title: await page.title(),
		headings: await page.getByRole("heading").allTextContents(),
		buttons: await page.getByRole("button").allTextContents(),
		links: await page.getByRole("link").allTextContents(),
		mainBox: await page
			.locator("main")
			.first()
			.boundingBox()
			.catch(() => null),
		bodyText: (await page.locator("body").innerText()).slice(0, 6000),
		unhandledApiRequests
	};
	await writeFile(path.join(artifactRoot, "audit.json"), `${JSON.stringify(report, null, 2)}\n`);
	console.log(JSON.stringify(report, null, 2));
};

const catalog = async page => {
	const routePaths = [
		"/projects/docs-demo/probes",
		"/projects/docs-demo/probes/new",
		"/projects/docs-demo/checks",
		"/projects/docs-demo/labels",
		"/projects/docs-demo/insight",
		"/projects/docs-demo/alerts?tab=incidents",
		"/projects/docs-demo/alerts?tab=rules",
		"/projects/docs-demo/alerts?tab=notifications",
		"/projects/docs-demo/status-pages",
		`/projects/docs-demo/status-pages/${ids.page}/edit`,
		"/projects/docs-demo/members",
		"/settings",
		"/admin",
		"/status/docs-demo"
	];
	const report = {};
	for (const routePath of routePaths) {
		await navigate(page, routePath);
		report[routePath] = {
			url: page.url(),
			headings: await page.getByRole("heading").allTextContents(),
			buttons: await page.getByRole("button").allTextContents(),
			links: await page.getByRole("link").allTextContents(),
			bodyText: (await page.locator("body").innerText()).slice(0, 10000)
		};
	}
	await writeFile(path.join(artifactRoot, "catalog.json"), `${JSON.stringify(report, null, 2)}\n`);
	console.log(`Cataloged ${Object.keys(report).length} routes.`);
};

const captureAll = async page => {
	// Quick Start
	await captureFirstProject(page);

	await navigate(page, "/projects/docs-demo/probes/new");
	await fillFirst(page, ["Probe name"], "taipei-office-02");
	const manualCoordinates = page.getByText("Manual coordinates", { exact: true });
	if (await manualCoordinates.count()) await manualCoordinates.last().click();
	await fillFirst(page, ["Location name"], "Taipei, Taiwan");
	await fillFirst(page, ["Latitude"], "25.0330");
	await fillFirst(page, ["Longitude"], "121.5654");
	await captureDoc(page, "getting-started/probe-location.webp");
	if (!(await clickFirst(page, ["Continue to install"]))) throw new Error("Probe wizard could not continue to the install step.");
	await page.getByText("Install the probe", { exact: true }).waitFor({ state: "visible", timeout: 8000 });
	await captureDoc(page, "getting-started/probe-install.webp");
	await page
		.getByText("Heartbeat received", { exact: true })
		.waitFor({ state: "visible", timeout: 8000 })
		.catch(() => page.waitForTimeout(2500));
	await captureDoc(page, "getting-started/probe-heartbeat.webp");

	await openNewCheck(page, "Ping");
	await fillCheckIdentity(page, { name: "Cloudflare DNS", target: "1.1.1.1" });
	await captureDoc(page, "getting-started/create-ping-check.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.ping}`);
	await captureDoc(page, "getting-started/first-result.webp");

	// Projects and members
	await navigate(page, "/projects/docs-demo/members");
	await captureDoc(page, "guides/projects/invite-member.webp");
	await scrollToText(page, "Member access");
	await captureDoc(page, "guides/projects/member-access.webp");

	// Probes
	await navigate(page, "/projects/docs-demo/probes");
	await captureDoc(page, "guides/probes/status-list.webp");
	await navigate(page, `/projects/docs-demo/probes/${ids.probes.tokyo}`);
	await captureDoc(page, "guides/probes/detail.webp");

	// Labels and assignments
	await navigate(page, "/projects/docs-demo/labels");
	await captureDoc(page, "guides/labels/registry.webp");
	await navigate(page, `/projects/docs-demo/checks/${ids.checks.ping}`);
	await scrollToText(page, "Probe selector");
	await clickFirst(page, ["Preview selector"]);
	await captureDoc(page, "guides/labels/selector-preview.webp");

	// Checks
	await navigate(page, "/projects/docs-demo/checks");
	await captureDoc(page, "guides/checks/list.webp");
	await openNewCheck(page, "Ping");
	await fillCheckIdentity(page, { name: "DNS latency", target: "1.1.1.1" });
	await scrollToText(page, "Ping config");
	await captureDoc(page, "guides/checks/ping-form.webp");
	await openNewCheck(page, "TCP");
	await fillCheckIdentity(page, { name: "HTTPS port", target: "example.com" });
	await scrollToText(page, "TCP config");
	await captureDoc(page, "guides/checks/tcp-form.webp");
	await openNewCheck(page, "HTTP / HTTPS");
	await fillCheckIdentity(page, { name: "HTTPS health", target: "https://example.com/health" });
	await scrollToText(page, "HTTP config");
	await captureDoc(page, "guides/checks/http-form.webp");
	await openNewCheck(page, "Traceroute");
	await fillCheckIdentity(page, { name: "Public route", target: "1.1.1.1" });
	await scrollToText(page, "Traceroute config");
	await captureDoc(page, "guides/checks/traceroute-form.webp");

	// Result insight
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.ping}`);
	await captureDoc(page, "guides/insight/scope-controls.webp");
	await navigate(page, "/projects/docs-demo/insight");
	if (!(await clickFirst(page, ["Select probes"]))) throw new Error("Insight probe picker could not be opened.");
	await captureDoc(page, "guides/insight/probe-picker.webp", { maxHeight: 650 });
	await page.keyboard.press("Escape");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}`);
	if (!(await clickFirst(page, ["Select checks"]))) throw new Error("Insight check picker could not be opened.");
	await captureDoc(page, "guides/insight/check-picker.webp", { maxHeight: 700 });
	await page.keyboard.press("Escape");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.ping}`);
	if (!(await clickFirst(page, ["Last 24 hours"]))) throw new Error("Insight time picker could not be opened.");
	await captureDoc(page, "guides/insight/time-popover.webp", { maxHeight: 800 });
	await page.keyboard.press("Escape");
	const multiPingScope = `/projects/docs-demo/insight?probeId=${ids.probes.taipei}&probeId=${ids.probes.tokyo}&probeId=${ids.probes.frankfurt}&checkId=${ids.checks.ping}`;
	await navigate(page, multiPingScope);
	await captureDoc(page, "guides/insight/multi-series.webp", { maxHeight: 750 });
	await scrollTextToTop(page, "Ping series (3 assignments)");
	await captureDoc(page, "guides/insight/ping-comparison.webp", { maxHeight: 850 });
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.ping}`);
	await scrollToText(page, "Tokyo Cloud → 1.1.1.1");
	await captureDoc(page, "guides/insight/ping-series.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.tcp}`);
	await captureDoc(page, "guides/insight/tcp-series.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.http}`);
	await captureDoc(page, "guides/insight/http-series.webp");
	await scrollToText(page, "Request phases");
	await captureDoc(page, "guides/insight/http-timing.webp");
	await navigate(page, `/projects/docs-demo/insight?checkId=${ids.checks.http}`);
	await scrollToText(page, "TLS certificate inventory (3)");
	await captureDoc(page, "guides/insight/tls-inventory.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.taipei}&checkId=${ids.checks.traceroute}`);
	await scrollToText(page, "Aggregated route graph");
	await captureDoc(page, "guides/insight/traceroute-graph.webp");
	await captureLocator(page, "guides/insight/traceroute-hop-table.webp", page.getByRole("table").first());
	const runTimeline = page.getByText("Run timeline", { exact: true }).locator("..").locator("..");
	await captureLocator(page, "guides/insight/traceroute-run-timeline.webp", runTimeline);

	// Alerts and incidents
	await navigate(page, "/projects/docs-demo/alerts?tab=incidents");
	await captureDoc(page, "guides/alerts/incident-list.webp");
	await navigate(page, "/projects/docs-demo/alerts?tab=rules");
	if (!(await clickFirst(page, ["Create rule"]))) throw new Error("Create rule button was not found.");
	await scrollToText(page, "Condition");
	await captureDoc(page, "guides/alerts/rule-condition.webp");
	await scrollToText(page, "Notify");
	await captureDoc(page, "guides/alerts/rule-notify.webp");
	await navigate(page, "/projects/docs-demo/alerts?tab=rules");
	await captureDoc(page, "guides/alerts/rule-list.webp");
	await navigate(page, `/projects/docs-demo/alerts/incident/${ids.incidents.open}`);
	await captureDoc(page, "guides/alerts/incident-detail.webp");

	// Notification destinations
	await navigate(page, "/projects/docs-demo/alerts?tab=notifications");
	if (!(await clickFirst(page, ["Add notification"]))) throw new Error("Add notification button was not found.");
	await captureDoc(page, "guides/notifications/type-picker.webp");
	await navigate(page, "/projects/docs-demo/alerts?tab=notifications");
	await captureDoc(page, "guides/notifications/list.webp");

	// Status pages
	await navigate(page, "/projects/docs-demo/status-pages");
	await captureDoc(page, "guides/status-pages/list.webp");
	await navigate(page, `/projects/docs-demo/status-pages/${ids.page}/edit`);
	await captureDoc(page, "guides/status-pages/builder.webp");
	await scrollToText(page, "Public data");
	await captureDoc(page, "guides/status-pages/public-data.webp");
	const editStatusBlock = page.getByRole("button", { name: "Edit HTTPS API", exact: true });
	if (await editStatusBlock.count()) {
		await editStatusBlock.click();
		await page.waitForTimeout(250);
	}
	await captureDoc(page, "guides/status-pages/block-editor.webp");
	await navigate(page, "/status/docs-demo");
	await captureDoc(page, "guides/status-pages/public-page.webp");

	// Account settings
	await navigate(page, "/settings");
	await scrollToText(page, "Active sessions");
	await captureDoc(page, "guides/account/sessions.webp");
	await scrollToText(page, "API tokens");
	if (!(await clickFirst(page, ["Create token"]))) throw new Error("Create token button was not found.");
	await captureDoc(page, "guides/account/token-form.webp");
	await fillFirst(page, ["Token name"], "Docs API");
	for (const scopeLabel of ["Read projects", "Read results"]) {
		const checkbox = page.getByLabel(scopeLabel, { exact: true });
		if ((await checkbox.count()) && !(await checkbox.isChecked())) await checkbox.check();
	}
	await page.getByRole("button", { name: "Create token", exact: true }).last().click();
	await page.getByText("Copy your API token", { exact: true }).waitFor({ state: "visible", timeout: 8000 });
	await captureDoc(page, "guides/account/token-created.webp");

	// System administration
	await navigate(page, "/admin");
	await scrollToText(page, "Instance access");
	await captureDoc(page, "guides/admin/access-policy.webp");
	await scrollToText(page, "SMTP delivery");
	await captureDoc(page, "guides/admin/smtp.webp");
	await scrollToText(page, "Authentication providers");
	await captureDoc(page, "guides/admin/auth-providers.webp");
	await scrollToText(page, "User management");
	await captureDoc(page, "guides/admin/users.webp");

	// Homepage product showcase, exported at the frame's exact 1200x780 slot.
	await navigate(page, "/projects/docs-demo/dashboard");
	await captureHomepage(page, "overview.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.ping}`);
	await captureHomepage(page, "ping-insight.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.http}`);
	await captureHomepage(page, "http-insight.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.tokyo}&checkId=${ids.checks.tcp}`);
	await captureHomepage(page, "tcp-insight.webp");
	await navigate(page, `/projects/docs-demo/insight?probeId=${ids.probes.taipei}&checkId=${ids.checks.traceroute}`);
	await captureHomepage(page, "traceroute-insight.webp");
	await navigate(page, `/projects/docs-demo/insight?checkId=${ids.checks.http}`);
	await scrollToText(page, "TLS certificate inventory (3)");
	await captureHomepage(page, "tls-inventory.webp");
	await navigate(page, `/projects/docs-demo/alerts/incident/${ids.incidents.open}`);
	await captureHomepage(page, "incident-detail.webp");
	await navigate(page, "/status/docs-demo");
	await captureHomepage(page, "public-status.webp");
};

await ensureDirectories();
const browser = await chromium.launch({ channel: "chrome", headless: true });
const context = await browser.newContext({
	viewport: { width: 1440, height: 1000 },
	deviceScaleFactor: 2,
	colorScheme: "light",
	locale: "en-US",
	timezoneId: "Asia/Taipei",
	reducedMotion: "reduce"
});

try {
	await installDeterministicBrowserState(context);
	await installApiFixtures(context);
	const page = await context.newPage();
	page.on("console", message => {
		if (message.type() === "error") console.error(`[browser] ${message.text()}`);
	});
	page.on("pageerror", error => console.error(`[page] ${error.message}`));

	if (auditOnly) await audit(page);
	else if (catalogOnly) await catalog(page);
	else if (firstProjectOnly) await captureFirstProject(page);
	else await captureAll(page);

	if (leakedWrites.length) throw new Error(`Blocked unmocked write requests:\n${leakedWrites.join("\n")}`);
	if (unhandledApiRequests.length) throw new Error(`Missing read fixtures:\n${[...new Set(unhandledApiRequests)].join("\n")}`);

	if (!auditOnly && !catalogOnly) {
		const fixtureText = await readFile(path.join(repositoryRoot, "scripts/docs-screenshots/fixtures.mjs"), "utf8");
		const forbiddenFragments = (process.env.DOCS_SCREENSHOT_FORBIDDEN_FRAGMENTS || "")
			.split(",")
			.map(value => value.trim())
			.filter(Boolean);
		for (const fragment of forbiddenFragments) {
			if (fixtureText.includes(fragment)) throw new Error("A forbidden fragment was found in screenshot fixtures.");
		}
		await writeFile(path.join(artifactRoot, "manifest.json"), `${JSON.stringify({ baseUrl, capturedFiles }, null, 2)}\n`);
	}
} finally {
	await context.close();
	await browser.close();
}
