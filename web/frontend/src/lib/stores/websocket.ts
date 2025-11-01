import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { ProgressMessage } from '$lib/api/types';

// Build WebSocket URL dynamically from browser location
// This works for both local dev (localhost:8080) and Docker (any host)
// Converts http -> ws and https -> wss automatically
function getWebSocketURL(): string {
	if (!browser) {
		// During SSR, return a placeholder (won't be used)
		return 'ws://localhost:8080/ws/progress';
	}
	// Replace http/https with ws/wss and append the WebSocket path
	return location.origin.replace(/^http/, 'ws') + '/ws/progress';
}

interface WebSocketState {
	connected: boolean;
	messages: ProgressMessage[];
	error?: string;
}

function createWebSocketStore() {
	const { subscribe, set, update } = writable<WebSocketState>({
		connected: false,
		messages: []
	});

	let ws: WebSocket | null = null;
	let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

	function connect() {
		if (!browser) {
			console.warn('WebSocket connection attempted during SSR, skipping');
			return;
		}

		if (ws?.readyState === WebSocket.OPEN) {
			return;
		}

		const wsUrl = getWebSocketURL();

		try {
			ws = new WebSocket(wsUrl);

			ws.onopen = () => {
				console.log('WebSocket connected to', wsUrl);
				update((state) => ({ ...state, connected: true, error: undefined }));
			};

			ws.onclose = () => {
				console.log('WebSocket disconnected');
				update((state) => ({ ...state, connected: false }));

				// Attempt to reconnect after 3 seconds
				reconnectTimeout = setTimeout(() => {
					console.log('Attempting to reconnect...');
					connect();
				}, 3000);
			};

			ws.onerror = (error) => {
				console.error('WebSocket error:', error);
				update((state) => ({ ...state, error: 'WebSocket connection error' }));
			};

			ws.onmessage = (event) => {
				try {
					const message: ProgressMessage = JSON.parse(event.data);
					update((state) => ({
						...state,
						messages: [...state.messages, message]
					}));
				} catch (error) {
					console.error('Failed to parse WebSocket message:', error);
				}
			};
		} catch (error) {
			console.error('Failed to create WebSocket:', error);
			update((state) => ({ ...state, error: 'Failed to create WebSocket connection' }));
		}
	}

	function disconnect() {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout);
			reconnectTimeout = null;
		}

		if (ws) {
			ws.close();
			ws = null;
		}

		set({ connected: false, messages: [] });
	}

	function clearMessages() {
		update((state) => ({ ...state, messages: [] }));
	}

	return {
		subscribe,
		connect,
		disconnect,
		clearMessages
	};
}

export const websocketStore = createWebSocketStore();
