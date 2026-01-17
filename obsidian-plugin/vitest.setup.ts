import { vi } from "vitest";

vi.mock("obsidian", () => {
	class TFile {
		constructor(public path: string) {
			this.name = path.split("/").pop() ?? path;
			this.extension = this.name.includes(".") ? this.name.split(".").pop() ?? "" : "";
			const nameWithoutExt = this.name.replace(/\.[^/.]+$/, "");
			this.basename = nameWithoutExt;
		}
		name: string;
		extension: string;
		basename: string;
		stat = {
			ctime: 0,
			mtime: 0,
			size: 0,
		};
	}

	class TFolder {
		constructor(public path: string) {
			this.name = path.split("/").pop() ?? path;
		}
		name: string;
		children: any[] = [];
	}

	class Vault {
		private files: TFile[] = [];
		
		getAbstractFileByPath(_path: string): TFile | TFolder | null {
			return null;
		}
		
		getMarkdownFiles(): TFile[] {
			return this.files;
		}
		
		async read(_file: TFile | string): Promise<string> {
			return "";
		}
		async modify(_file: TFile, _content: string): Promise<void> {}
		async create(_path: string, _content: string): Promise<TFile> {
			return new TFile(_path);
		}
		async createFolder(_path: string): Promise<void> {}
		async rename(_file: TFile, _newPath: string): Promise<void> {}
	}

	class MetadataCache {
		getFirstLinkpathDest(_linkpath: string, _sourcePath: string): TFile | null {
			return null;
		}
	}

	return {
		TFile,
		TFolder,
		Vault,
		MetadataCache,
		Notice: class {
			constructor(_message: string) {}
		},
		normalizePath: (path: string) => path,
	};
});

