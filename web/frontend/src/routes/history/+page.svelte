<script lang="ts">
	import { onMount } from 'svelte';
	import { Calendar, CheckCircle, XCircle, Clock } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';

	// Mock history data - in a real implementation, this would come from the API
	let history = $state<any[]>([]);
	let loading = $state(true);

	onMount(() => {
		// Simulate loading history
		setTimeout(() => {
			history = [
				{
					id: '1',
					timestamp: new Date('2024-01-15T10:30:00'),
					operation: 'Batch Scrape',
					files_processed: 15,
					successful: 13,
					failed: 2,
					status: 'completed'
				},
				{
					id: '2',
					timestamp: new Date('2024-01-14T15:20:00'),
					operation: 'Batch Scrape',
					files_processed: 8,
					successful: 8,
					failed: 0,
					status: 'completed'
				},
				{
					id: '3',
					timestamp: new Date('2024-01-14T09:15:00'),
					operation: 'Batch Scrape',
					files_processed: 20,
					successful: 18,
					failed: 2,
					status: 'completed'
				}
			];
			loading = false;
		}, 500);
	});

	function formatDate(date: Date) {
		return new Intl.DateTimeFormat('en-US', {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(date);
	}
</script>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-4xl mx-auto space-y-6">
		<!-- Header -->
		<div>
			<h1 class="text-3xl font-bold">Operation History</h1>
			<p class="text-muted-foreground mt-1">View past scraping and organization operations</p>
		</div>

		{#if loading}
			<Card class="p-8 text-center">
				<Clock class="h-8 w-8 animate-spin mx-auto mb-2" />
				<p class="text-muted-foreground">Loading history...</p>
			</Card>
		{:else if history.length === 0}
			<Card class="p-8 text-center">
				<p class="text-muted-foreground">No operations in history</p>
			</Card>
		{:else}
			<div class="space-y-3">
				{#each history as item}
					<Card class="p-4 hover:shadow-md transition-shadow">
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<div class="flex items-center gap-3 mb-2">
									{#if item.status === 'completed'}
										<CheckCircle class="h-5 w-5 text-green-500" />
									{:else if item.status === 'failed'}
										<XCircle class="h-5 w-5 text-red-500" />
									{:else}
										<Clock class="h-5 w-5 text-blue-500" />
									{/if}
									<h3 class="font-semibold">{item.operation}</h3>
									<span
										class="px-2 py-0.5 text-xs rounded {item.status === 'completed'
											? 'bg-green-100 text-green-700'
											: item.status === 'failed'
												? 'bg-red-100 text-red-700'
												: 'bg-blue-100 text-blue-700'}"
									>
										{item.status}
									</span>
								</div>

								<div class="flex items-center gap-4 text-sm text-muted-foreground">
									<div class="flex items-center gap-1">
										<Calendar class="h-4 w-4" />
										{formatDate(item.timestamp)}
									</div>
									<div>
										{item.files_processed} file{item.files_processed !== 1 ? 's' : ''}
									</div>
									<div class="text-green-600">{item.successful} successful</div>
									{#if item.failed > 0}
										<div class="text-red-600">{item.failed} failed</div>
									{/if}
								</div>
							</div>
						</div>
					</Card>
				{/each}
			</div>
		{/if}

		<!-- Info -->
		<Card class="p-4 bg-accent/30">
			<p class="text-sm text-muted-foreground">
				<strong>Note:</strong> History tracking will be implemented in a future update. This page currently
				shows mock data for demonstration purposes.
			</p>
		</Card>
	</div>
</div>
