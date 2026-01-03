export interface UnsplashPhoto {
	id: string;
	width: number;
	height: number;
	urls: {
		raw: string;
		full: string;
		regular: string;
		small: string;
		thumb: string;
	};
	alt_description: string | null;
	description: string | null;
	user: {
		name: string;
		username: string;
	};
	links: {
		html: string;
	};
}

export interface UnsplashSearchResponse {
	total: number;
	total_pages: number;
	results: UnsplashPhoto[];
}

export class UnsplashClient {
	private readonly accessKey: string;
	private readonly secretKey: string;
	private readonly apiUrl = "https://api.unsplash.com";

	constructor(accessKey?: string, secretKey?: string) {
		// Unsplash requires an access key (Application ID) for API access
		// Secret Key is optional but may be needed for some operations
		// You can get both keys from https://unsplash.com/developers
		this.accessKey = accessKey || "YOUR_UNSPLASH_ACCESS_KEY";
		this.secretKey = secretKey || "";
	}

	async searchImages(query: string, page: number = 1, perPage: number = 20): Promise<UnsplashSearchResponse> {
		const url = `${this.apiUrl}/search/photos?query=${encodeURIComponent(query)}&page=${page}&per_page=${perPage}`;
		
		const headers: HeadersInit = {
			"Accept-Version": "v1",
		};

		// Unsplash API requires Authorization header with access key
		if (!this.accessKey || this.accessKey === "YOUR_UNSPLASH_ACCESS_KEY") {
			throw new Error("Unsplash access key is required. Please configure it in plugin settings.");
		}

		headers["Authorization"] = `Client-ID ${this.accessKey}`;

		const response = await fetch(url, { headers });

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Unsplash API error: ${response.status} ${response.statusText}. ${errorText}`);
		}

		return response.json() as Promise<UnsplashSearchResponse>;
	}

	getImageUrl(photo: UnsplashPhoto, size: "raw" | "full" | "regular" | "small" | "thumb" = "regular"): string {
		return photo.urls[size];
	}

	getAttributionUrl(photo: UnsplashPhoto): string {
		return photo.links.html;
	}

	getAttributionText(photo: UnsplashPhoto): string {
		return `Photo by ${photo.user.name} on Unsplash`;
	}
}

