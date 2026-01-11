const ACCENTS_REGEX = /[\u0300-\u036f]/g;
const NON_ALNUM = /[^a-z0-9\s-]/gi;
const SPACES = /\s+/g;
const MULTIPLE_DASH = /-+/g;
const EDGE_DASH = /^-+|-+$/g;

export function slugify(value: string, fallback = "untitled"): string {
	const normalized = value
		.normalize("NFKD")
		.toLowerCase()
		.replace(ACCENTS_REGEX, "")
		.replace(NON_ALNUM, "")
		.trim()
		.replace(SPACES, "-")
		.replace(MULTIPLE_DASH, "-")
		.replace(EDGE_DASH, "");
	return normalized || fallback;
}

