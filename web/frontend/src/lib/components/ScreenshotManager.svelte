<script lang="ts">
	import type { Movie } from '$lib/api/types';
	import Button from './ui/Button.svelte';
	import Card from './ui/Card.svelte';
	import { Plus, Trash2, Image as ImageIcon, Play, ZoomIn, ZoomOut, X, ChevronLeft, ChevronRight, RotateCcw } from 'lucide-svelte';

	interface Props {
		movie: Movie;
		onUpdate: (movie: Movie) => void;
	}

	let { movie, onUpdate }: Props = $props();

	let screenshots = $state<string[]>(movie.screenshot_urls || []);
	let posterUrl = $state(movie.poster_url || '');
	let coverUrl = $state(movie.cover_url || '');
	let trailerUrl = $state(movie.trailer_url || '');
	let newScreenshotUrl = $state('');

	// Screenshot viewer modal state
	let showViewer = $state(false);
	let currentIndex = $state(0);
	let zoomLevel = $state(100);

	// Trailer modal state
	let showTrailerModal = $state(false);

	// Sync state when movie prop changes
	$effect(() => {
		screenshots = movie.screenshot_urls || [];
		posterUrl = movie.poster_url || '';
		coverUrl = movie.cover_url || '';
		trailerUrl = movie.trailer_url || '';
	});

	// Update parent when data changes
	$effect(() => {
		onUpdate({
			...movie,
			screenshot_urls: screenshots,
			poster_url: posterUrl,
			cover_url: coverUrl,
			trailer_url: trailerUrl
		});
	});

	function addScreenshot() {
		if (newScreenshotUrl.trim()) {
			screenshots = [...screenshots, newScreenshotUrl.trim()];
			newScreenshotUrl = '';
		}
	}

	function removeScreenshot(index: number) {
		screenshots = screenshots.filter((_, i) => i !== index);
	}

	function handleKeyPress(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			addScreenshot();
		}
	}

	// Screenshot viewer functions
	function openViewer(index: number) {
		currentIndex = index;
		zoomLevel = 100;
		showViewer = true;
	}

	function closeViewer() {
		showViewer = false;
	}

	function nextImage() {
		if (currentIndex < screenshots.length - 1) {
			currentIndex++;
			zoomLevel = 100;
		}
	}

	function prevImage() {
		if (currentIndex > 0) {
			currentIndex--;
			zoomLevel = 100;
		}
	}

	function zoomIn() {
		zoomLevel = Math.min(zoomLevel + 25, 300);
	}

	function zoomOut() {
		zoomLevel = Math.max(zoomLevel - 25, 50);
	}

	function resetZoom() {
		zoomLevel = 100;
	}

	// Keyboard navigation
	$effect(() => {
		if (!showViewer) return;

		function handleKeyDown(e: KeyboardEvent) {
			switch (e.key) {
				case 'Escape':
					closeViewer();
					break;
				case 'ArrowLeft':
					prevImage();
					break;
				case 'ArrowRight':
					nextImage();
					break;
				case '+':
				case '=':
					zoomIn();
					break;
				case '-':
					zoomOut();
					break;
				case '0':
					resetZoom();
					break;
			}
		}

		window.addEventListener('keydown', handleKeyDown);
		return () => window.removeEventListener('keydown', handleKeyDown);
	});
</script>

<div class="space-y-6">
	<!-- Poster Image -->
	<div>
		<h3 class="text-lg font-semibold mb-3">Poster Image</h3>
		<div class="space-y-3">
			<div>
				<label class="text-sm font-medium mb-1 block">Poster URL</label>
				<input
					type="url"
					bind:value={posterUrl}
					placeholder="https://..."
					class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
				/>
			</div>
			<div>
				<label class="text-sm font-medium mb-1 block">Preview (Cropped)</label>
				{#if posterUrl}
					<!-- Crop to show only right 47.2% of image (removes promotional text on left) -->
					<div class="w-full max-w-xs aspect-[2/3] overflow-hidden rounded border relative">
						<img
							src={posterUrl}
							alt="Poster"
							class="absolute h-full"
							style="right: 0; width: auto; min-width: 211.8%; object-fit: cover; object-position: right center;"
							onerror={(e) => {
								e.currentTarget.src = 'https://via.placeholder.com/300x450?text=Invalid+URL';
							}}
						/>
					</div>
				{:else}
					<div
						class="w-full max-w-xs aspect-[2/3] bg-accent rounded border flex items-center justify-center text-muted-foreground"
					>
						<div class="text-center">
							<ImageIcon class="h-12 w-12 mx-auto mb-2 opacity-50" />
							<p class="text-sm">No poster</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Cover/Fanart Image -->
	<div>
		<h3 class="text-lg font-semibold mb-3">Cover/Fanart Image</h3>
		<div class="space-y-3">
			<div>
				<label class="text-sm font-medium mb-1 block">Cover URL</label>
				<input
					type="url"
					bind:value={coverUrl}
					placeholder="https://..."
					class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
				/>
			</div>
			<div>
				<label class="text-sm font-medium mb-1 block">Preview</label>
				{#if coverUrl}
					<img
						src={coverUrl}
						alt="Cover"
						class="w-full rounded border"
						onerror={(e) => {
							e.currentTarget.src = 'https://via.placeholder.com/400x225?text=Invalid+URL';
						}}
					/>
				{:else}
					<div
						class="w-full h-48 bg-accent rounded border flex items-center justify-center text-muted-foreground"
					>
						<div class="text-center">
							<ImageIcon class="h-12 w-12 mx-auto mb-2 opacity-50" />
							<p class="text-sm">No cover image</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Trailer -->
	<div>
		<h3 class="text-lg font-semibold mb-3">Trailer</h3>
		<div class="space-y-3">
			<div>
				<label class="text-sm font-medium mb-1 block">Trailer URL</label>
				<input
					type="url"
					bind:value={trailerUrl}
					placeholder="https://..."
					class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
				/>
			</div>
			<div>
				<label class="text-sm font-medium mb-1 block">Preview</label>
				{#if trailerUrl}
					<button
						onclick={() => (showTrailerModal = true)}
						class="block w-full h-48 bg-accent rounded border flex items-center justify-center text-primary hover:bg-accent/80 transition-colors cursor-pointer"
					>
						<div class="text-center">
							<Play class="h-12 w-12 mx-auto mb-2" />
							<p class="text-sm font-medium">Play Trailer</p>
						</div>
					</button>
				{:else}
					<div
						class="w-full h-48 bg-accent rounded border flex items-center justify-center text-muted-foreground"
					>
						<div class="text-center">
							<Play class="h-12 w-12 mx-auto mb-2 opacity-50" />
							<p class="text-sm">No trailer</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Screenshots -->
	<div>
		<div class="flex items-center justify-between mb-3">
			<h3 class="text-lg font-semibold">Screenshots ({screenshots.length})</h3>
		</div>

		<!-- Add Screenshot Form -->
		<div class="flex gap-2 mb-4">
			<input
				type="url"
				bind:value={newScreenshotUrl}
				onkeypress={handleKeyPress}
				placeholder="Enter screenshot URL and press Enter or click Add"
				class="flex-1 px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
			/>
			<Button onclick={addScreenshot} disabled={!newScreenshotUrl.trim()}>
				{#snippet children()}
					<Plus class="h-4 w-4 mr-2" />
					Add
				{/snippet}
			</Button>
		</div>

		<!-- Screenshots Grid -->
		{#if screenshots.length === 0}
			<div class="text-center py-8 text-muted-foreground border-2 border-dashed rounded-lg">
				<ImageIcon class="h-12 w-12 mx-auto mb-2 opacity-50" />
				<p>No screenshots added</p>
				<p class="text-xs mt-1">Add screenshot URLs above</p>
			</div>
		{:else}
			<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
				{#each screenshots as url, index}
					<Card class="p-2 group relative">
						<button
							onclick={() => openViewer(index)}
							class="w-full cursor-pointer hover:opacity-80 transition-opacity"
						>
							<img
								src={url}
								alt="Screenshot {index + 1}"
								class="w-full aspect-video object-cover rounded"
								onerror={(e) => {
									e.currentTarget.src = 'https://via.placeholder.com/400x225?text=Invalid+URL';
								}}
							/>
						</button>
						<div class="mt-2 flex items-center justify-between gap-2">
							<p class="text-xs text-muted-foreground truncate flex-1" title={url}>
								Screenshot {index + 1}
							</p>
							<Button
								variant="ghost"
								size="sm"
								onclick={() => removeScreenshot(index)}
								class="text-destructive hover:bg-destructive/10 p-1 h-auto"
							>
								{#snippet children()}
									<Trash2 class="h-3 w-3" />
								{/snippet}
							</Button>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Screenshot Viewer Modal -->
{#if showViewer}
	<div
		class="fixed inset-0 z-50 bg-black/90 flex items-center justify-center"
		onclick={closeViewer}
	>
		<div class="relative w-full h-full flex items-center justify-center p-4" onclick={(e) => e.stopPropagation()}>
			<!-- Close Button -->
			<button
				onclick={closeViewer}
				class="absolute top-4 right-4 z-10 p-2 bg-black/50 hover:bg-black/70 rounded-full text-white transition-colors"
				title="Close (Esc)"
			>
				<X class="h-6 w-6" />
			</button>

			<!-- Image Counter -->
			<div class="absolute top-4 left-4 z-10 px-3 py-2 bg-black/50 rounded text-white text-sm">
				{currentIndex + 1} / {screenshots.length}
			</div>

			<!-- Zoom Controls -->
			<div class="absolute top-4 left-1/2 -translate-x-1/2 z-10 flex items-center gap-2 bg-black/50 rounded px-3 py-2">
				<button
					onclick={zoomOut}
					disabled={zoomLevel <= 50}
					class="p-1 text-white hover:bg-white/10 rounded disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
					title="Zoom Out (-)"
				>
					<ZoomOut class="h-5 w-5" />
				</button>
				<button
					onclick={resetZoom}
					class="px-2 py-1 text-white hover:bg-white/10 rounded text-sm transition-colors"
					title="Reset Zoom (0)"
				>
					{zoomLevel}%
				</button>
				<button
					onclick={zoomIn}
					disabled={zoomLevel >= 300}
					class="p-1 text-white hover:bg-white/10 rounded disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
					title="Zoom In (+)"
				>
					<ZoomIn class="h-5 w-5" />
				</button>
			</div>

			<!-- Previous Button -->
			{#if currentIndex > 0}
				<button
					onclick={prevImage}
					class="absolute left-4 top-1/2 -translate-y-1/2 p-3 bg-black/50 hover:bg-black/70 rounded-full text-white transition-colors"
					title="Previous (←)"
				>
					<ChevronLeft class="h-8 w-8" />
				</button>
			{/if}

			<!-- Next Button -->
			{#if currentIndex < screenshots.length - 1}
				<button
					onclick={nextImage}
					class="absolute right-4 top-1/2 -translate-y-1/2 p-3 bg-black/50 hover:bg-black/70 rounded-full text-white transition-colors"
					title="Next (→)"
				>
					<ChevronRight class="h-8 w-8" />
				</button>
			{/if}

			<!-- Image -->
			<div class="overflow-auto max-w-full max-h-full">
				<img
					src={screenshots[currentIndex]}
					alt="Screenshot {currentIndex + 1}"
					style="width: {zoomLevel}%; height: auto; max-width: none;"
					class="block mx-auto"
				/>
			</div>
		</div>
	</div>
{/if}

<!-- Trailer Modal -->
{#if showTrailerModal && trailerUrl}
	<div class="fixed inset-0 bg-black/90 z-50 flex items-center justify-center p-4" onclick={() => (showTrailerModal = false)}>
		<div class="relative w-full max-w-4xl" onclick={(e) => e.stopPropagation()}>
			<!-- Close Button -->
			<button
				onclick={() => (showTrailerModal = false)}
				class="absolute -top-12 right-0 p-2 bg-black/50 hover:bg-black/70 rounded-full text-white transition-colors"
				title="Close (Esc)"
			>
				<X class="h-6 w-6" />
			</button>

			<!-- Video Player -->
			<video controls class="w-full rounded" src={trailerUrl}>
				Your browser does not support the video tag.
			</video>
		</div>
	</div>
{/if}
