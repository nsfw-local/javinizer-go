<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import Navigation from '$lib/components/Navigation.svelte';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
	import { websocketStore } from '$lib/stores/websocket';
	import '../app.css';

	let { children } = $props();

	// Initialize WebSocket connection on mount
	onMount(() => {
		websocketStore.connect();
	});

	// Clean up WebSocket on destroy
	onDestroy(() => {
		websocketStore.disconnect();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="min-h-screen bg-background">
	<Navigation />
	{@render children?.()}
	<ToastContainer />
</div>
