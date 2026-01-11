export class TFile {
	path: string;
	name: string;
	extension: string;
	constructor(path: string) {
		this.path = path;
		this.name = path.split("/").pop() ?? path;
		this.extension = this.name.includes(".") ? this.name.split(".").pop() ?? "" : "";
	}
}

export class TFolder {
	path: string;
	name: string;
	children: Array<TFile | TFolder> = [];
	constructor(path: string) {
		this.path = path;
		this.name = path.split("/").pop() ?? path;
	}
}

export class Vault {
	private files = new Map<string, string>();

	getAbstractFileByPath(path: string): TFile | TFolder | null {
		if (this.files.has(path)) {
			return new TFile(path);
		}
		return null;
	}

	async read(path: TFile | string): Promise<string> {
		const key = typeof path === "string" ? path : path.path;
		return this.files.get(key) ?? "";
	}

	async modify(path: TFile, data: string): Promise<void> {
		this.files.set(path.path, data);
	}

	async create(path: string, data: string): Promise<TFile> {
		this.files.set(path, data);
		return new TFile(path);
	}

	async createFolder(_path: string): Promise<void> {
		return;
	}

	async rename(file: TFile, newPath: string): Promise<void> {
		const data = this.files.get(file.path);
		if (data !== undefined) {
			this.files.delete(file.path);
			this.files.set(newPath, data);
		}
	}
}

export class Notice {
	constructor(_message: string) {}
}

export const normalizePath = (path: string) => path;

