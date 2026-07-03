export const apiQueryKeys = {
	system: {
		all: ["system"] as const,
		health: () => [...apiQueryKeys.system.all, "health"] as const,
		root: () => [...apiQueryKeys.system.all, "root"] as const
	},
	auth: {
		all: ["auth"] as const,
		me: () => [...apiQueryKeys.auth.all, "me"] as const
	},
	publicStatus: {
		all: ["public-status"] as const,
		detailRoot: (slug: string) => [...apiQueryKeys.publicStatus.all, "detail", slug] as const,
		detail: (slug: string, filters: object = {}) => [...apiQueryKeys.publicStatus.detailRoot(slug), filters] as const,
		summary: (slug: string) => [...apiQueryKeys.publicStatus.all, "summary", slug] as const,
		elements: (slug: string) => [...apiQueryKeys.publicStatus.all, "elements", slug] as const,
		incidents: (slug: string, filters: object = {}) => [...apiQueryKeys.publicStatus.all, "incidents", slug, filters] as const,
		elementChart: (slug: string, elementId: string, filters: object = {}) => [...apiQueryKeys.publicStatus.all, "element-chart", slug, elementId, filters] as const
	},
	projects: {
		all: ["projects"] as const,
		list: () => [...apiQueryKeys.projects.all, "list"] as const,
		detail: (ref: string) => [...apiQueryKeys.projects.all, "detail", ref] as const,
		assignmentsRoot: (ref: string) => [...apiQueryKeys.projects.detail(ref), "assignments"] as const,
		assignments: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.assignmentsRoot(ref), filters] as const,
		alertsRoot: (ref: string) => [...apiQueryKeys.projects.detail(ref), "alerts"] as const,
		alertRules: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.alertsRoot(ref), "rules", filters] as const,
		alertIncidents: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.alertsRoot(ref), "incidents", filters] as const,
		alertIncidentDetail: (ref: string, incidentId: string) => [...apiQueryKeys.projects.alertsRoot(ref), "incident", incidentId] as const,
		notifications: (ref: string) => [...apiQueryKeys.projects.alertsRoot(ref), "notifications"] as const,
		statusPagesRoot: (ref: string) => [...apiQueryKeys.projects.detail(ref), "status-pages"] as const,
		statusPages: (ref: string) => [...apiQueryKeys.projects.statusPagesRoot(ref), "list"] as const,
		statusPageDetail: (ref: string, pageId: string) => [...apiQueryKeys.projects.statusPagesRoot(ref), "detail", pageId] as const,
		checks: (ref: string) => [...apiQueryKeys.projects.detail(ref), "checks"] as const,
		checkDetail: (ref: string, checkId: string) => [...apiQueryKeys.projects.checks(ref), checkId] as const,
		labels: (ref: string) => [...apiQueryKeys.projects.detail(ref), "labels"] as const,
		latestResults: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "latest", filters] as const,
		members: (ref: string) => [...apiQueryKeys.projects.detail(ref), "members"] as const,
		invites: (ref: string) => [...apiQueryKeys.projects.detail(ref), "invites"] as const,
		currentUserInvites: () => [...apiQueryKeys.projects.all, "current-user-invites"] as const,
		pingInsight: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "ping", "insight", probeId, checkId, filters] as const,
		pingSeries: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "ping", "series", probeId, checkId, filters] as const,
		tcpInsight: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "tcp", "insight", probeId, checkId, filters] as const,
		tcpSeries: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "tcp", "series", probeId, checkId, filters] as const,
		probes: (ref: string) => [...apiQueryKeys.projects.detail(ref), "probes"] as const,
		probeDetail: (ref: string, probeId: string) => [...apiQueryKeys.projects.probes(ref), probeId] as const,
		tracerouteInsight: (ref: string, probeId: string, checkId: string, filters: object = {}) =>
			[...apiQueryKeys.projects.detail(ref), "results", "traceroute", "insight", probeId, checkId, filters] as const,
		tracerouteRuns: (ref: string, probeId: string, checkId: string, filters: object = {}) =>
			[...apiQueryKeys.projects.detail(ref), "results", "traceroute", "runs", probeId, checkId, filters] as const,
		tracerouteTopology: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "traceroute", "topology", filters] as const
	}
};
