import { defineConfig } from "vitest/config";

export default defineConfig({
	test: {
		globals: true,
		environment: "node",
		include: ["src/**/*.test.ts", "src/**/*.spec.ts"],
		setupFiles: ["./vitest.setup.ts"],
		coverage: {
			reporter: ["text", "json", "html"],
		},
	},
	resolve: {
		alias: [
			{
				find: /^obsidian$/,
				replacement: new URL("./vitest.obsidian.mock.ts", import.meta.url).pathname,
			},
		],
	},
});

