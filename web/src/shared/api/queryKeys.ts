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
		measurements: (ref: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "measurements", filters] as const,
		members: (ref: string) => [...apiQueryKeys.projects.detail(ref), "members"] as const,
		pingSeries: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "ping", "series", probeId, checkId, filters] as const,
		probes: (ref: string) => [...apiQueryKeys.projects.detail(ref), "probes"] as const,
		probeDetail: (ref: string, probeId: string) => [...apiQueryKeys.projects.probes(ref), probeId] as const,
		tracerouteRuns: (ref: string, probeId: string, checkId: string, filters: object = {}) => [...apiQueryKeys.projects.detail(ref), "results", "traceroute", "runs", probeId, checkId, filters] as const
	}
};
