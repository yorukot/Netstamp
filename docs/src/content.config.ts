import { glob } from "astro/loaders";
import { defineCollection, z } from "astro:content";

const icon = z.enum(["activity", "api", "bolt", "book", "code", "compass", "cube", "database", "deployment", "key", "map", "route", "server", "shield", "terminal", "users", "wrench"]);

const docs = defineCollection({
	loader: glob({ pattern: "**/*.mdx", base: "./src/content/docs" }),
	schema: z.object({
		title: z.string(),
		description: z.string(),
		icon: icon.default("book"),
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
