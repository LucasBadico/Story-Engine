import esbuild from "esbuild";
import process from "process";

const isProd = process.argv[2] === "production";

const ctx = await esbuild.context({
	bundle: true,
	entryPoints: ["src/main.ts"],
	external: ["obsidian"],
	format: "cjs",
	target: "es2018",
	logLevel: "info",
	sourcemap: isProd ? false : "inline",
	treeShaking: true,
	outfile: "main.js",
});

if (isProd) {
	await ctx.rebuild();
	process.exit(0);
} else {
	await ctx.watch();
}

