<script lang="ts">
	import type { Movie } from '$lib/api/types';
	import { AlertCircle } from 'lucide-svelte';

	interface Props {
		movie: Movie;
		originalMovie: Movie;
		onUpdate: (movie: Movie) => void;
	}

	let { movie, originalMovie, onUpdate }: Props = $props();

	// Create a local editable copy
	let editedMovie = $state({ ...movie });
	let isInitialized = $state(false);

	// Sync editedMovie when movie prop changes
	$effect(() => {
		editedMovie = { ...movie };
		isInitialized = false;
	});

	// Track which fields have been modified
	function isModified(field: keyof Movie): boolean {
		return editedMovie[field] !== originalMovie[field];
	}

	// Update parent when fields change (but not on initial mount)
	$effect(() => {
		// Access editedMovie to track it as a dependency
		const _ = editedMovie;

		if (!isInitialized) {
			isInitialized = true;
			return;
		}

		onUpdate(editedMovie);
	});

	function handleDateChange(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.value) {
			editedMovie.release_date = target.value;
		}
	}

	// Format date for input field
	const formattedDate = $derived(
		editedMovie.release_date
			? new Date(editedMovie.release_date).toISOString().split('T')[0]
			: ''
	);
</script>

<div class="space-y-4">
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
		<!-- ID -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Movie ID
				{#if isModified('id')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.id}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Content ID -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Content ID
				{#if isModified('content_id')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.content_id}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Title -->
		<div class="md:col-span-2">
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Title
				{#if isModified('title')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.title}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Original Title -->
		<div class="md:col-span-2">
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Original Title (Japanese)
				{#if isModified('original_title')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.original_title}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Description -->
		<div class="md:col-span-2">
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Description
				{#if isModified('description')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<textarea
				bind:value={editedMovie.description}
				rows="4"
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			></textarea>
		</div>

		<!-- Release Date -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Release Date
				{#if isModified('release_date')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="date"
				value={formattedDate}
				onchange={handleDateChange}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Runtime -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Runtime (minutes)
				{#if isModified('runtime')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="number"
				bind:value={editedMovie.runtime}
				min="0"
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Director -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Director
				{#if isModified('director')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.director}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Studio -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Studio / Maker
				{#if isModified('studio')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.studio}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Label -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Label
				{#if isModified('label')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.label}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Series -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Series
				{#if isModified('series')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="text"
				bind:value={editedMovie.series}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>

		<!-- Rating -->
		<div>
			<label class="flex items-center gap-2 text-sm font-medium mb-1">
				Rating
				{#if isModified('rating')}
					<AlertCircle class="h-3 w-3 text-orange-600" />
				{/if}
			</label>
			<input
				type="number"
				bind:value={editedMovie.rating}
				min="0"
				max="10"
				step="0.1"
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
			/>
		</div>
	</div>
</div>
