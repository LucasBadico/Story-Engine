import { SyncEntityPayload } from "./entitySyncTypes";

type Subscriber = (payload: SyncEntityPayload) => void | Promise<void>;

class ApiUpdateNotifier {
	private subscribers = new Set<Subscriber>();

	subscribe(subscriber: Subscriber): () => void {
		this.subscribers.add(subscriber);
		return () => {
			this.subscribers.delete(subscriber);
		};
	}

	async notify(payload: SyncEntityPayload): Promise<void> {
		for (const subscriber of this.subscribers) {
			try {
				await subscriber(payload);
			} catch (err) {
				console.error("Auto sync subscriber failed", err);
			}
		}
	}
}

export const apiUpdateNotifier = new ApiUpdateNotifier();

