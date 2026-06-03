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
	projects: {
		all: ["projects"] as const,
		list: () => [...apiQueryKeys.projects.all, "list"] as const,
		detail: (ref: string) => [...apiQueryKeys.projects.all, "detail", ref] as const,
		assignmentsRoot: (ref: string) => [...apiQueryKeys.projects.detail(ref), "assignments"] as const,
		assignments: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.assignmentsRoot(ref), filters] as const,
		checks: (ref: string) => [...apiQueryKeys.projects.detail(ref), "checks"] as const,
		checkDetail: (ref: string, checkId: string) => [...apiQueryKeys.projects.checks(ref), checkId] as const,
		labels: (ref: string) => [...apiQueryKeys.projects.detail(ref), "labels"] as const,
		latestResults: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "latest", filters] as const,
		members: (ref: string) => [...apiQueryKeys.projects.detail(ref), "members"] as const,
		invites: (ref: string) => [...apiQueryKeys.projects.detail(ref), "invites"] as const,
		publicPages: (ref: string) => [...apiQueryKeys.projects.detail(ref), "public-pages"] as const,
		publicPageDetail: (ref: string, pageId: string) => [...apiQueryKeys.projects.publicPages(ref), pageId] as const,
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
	},
	publicPages: {
		all: ["public-pages"] as const,
		detail: (slug: string) => [...apiQueryKeys.publicPages.all, slug] as const,
		pingInsight: (slug: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.publicPages.detail(slug), "ping", "insight", probeId, checkId, filters] as const
	}
};
