import { glob } from "astro/loaders";
import { defineCollection, z } from "astro:content";

const icon = z.enum(["activity", "api", "bolt", "book", "code", "codeBlock", "compass", "cube", "database", "deployment", "key", "map", "route", "server", "shield", "users", "wrench"]);
const navSection = z.enum(["start", "install", "use", "operate", "api", "development", "community"]);

const docs = defineCollection({
	loader: glob({ pattern: "**/*.mdx", base: "./src/content/docs" }),
	schema: z.object({
		title: z.string(),
		description: z.string(),
		icon: icon.default("book"),
		editPath: z.string().optional(),
		navTitle: z.string().optional(),
		navSection: navSection.optional(),
		navOrder: z.number().int().nonnegative().optional(),
		order: z.number().optional(),
		draft: z.boolean().optional(),
		head: z
			.array(
				z.object({
					tag: z.string(),
					content: z.string().optional(),
					attrs: z.record(z.string(), z.string()).optional()
				})
			)
			.optional()
	})
});

export const collections = { docs };
