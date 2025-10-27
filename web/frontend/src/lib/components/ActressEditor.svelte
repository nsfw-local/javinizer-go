<script lang="ts">
	import type { Movie, Actress } from '$lib/api/types';
	import Button from './ui/Button.svelte';
	import Card from './ui/Card.svelte';
	import { Plus, Edit, Trash2, X, Save } from 'lucide-svelte';

	interface Props {
		movie: Movie;
		onUpdate: (movie: Movie) => void;
	}

	let { movie, onUpdate }: Props = $props();

	let actresses = $state<Actress[]>(movie.actresses || []);
	let showEditModal = $state(false);
	let editingIndex = $state<number | null>(null);
	let editingActress = $state<Actress>({
		first_name: '',
		last_name: '',
		japanese_name: '',
		thumb_url: ''
	});

	// Sync actresses when movie prop changes
	$effect(() => {
		actresses = movie.actresses || [];
	});

	// Helper to get full name
	function getFullName(actress: Actress): string {
		if (actress.last_name && actress.first_name) {
			return `${actress.last_name} ${actress.first_name}`;
		}
		if (actress.first_name) {
			return actress.first_name;
		}
		return actress.japanese_name || 'Unknown';
	}

	// Update parent when actresses change
	$effect(() => {
		onUpdate({ ...movie, actresses });
	});

	function openAddActress() {
		editingIndex = null;
		editingActress = {
			first_name: '',
			last_name: '',
			japanese_name: '',
			thumb_url: ''
		};
		showEditModal = true;
	}

	function openEditActress(index: number) {
		editingIndex = index;
		editingActress = { ...actresses[index] };
		showEditModal = true;
	}

	function saveActress() {
		if (!editingActress.first_name?.trim() && !editingActress.japanese_name?.trim()) {
			alert('At least first name or Japanese name is required');
			return;
		}

		if (editingIndex !== null) {
			// Edit existing
			actresses[editingIndex] = editingActress;
		} else {
			// Add new
			actresses = [...actresses, editingActress];
		}

		showEditModal = false;
	}

	function removeActress(index: number) {
		if (confirm('Remove this actress?')) {
			actresses = actresses.filter((_, i) => i !== index);
		}
	}

	function cancelEdit() {
		showEditModal = false;
	}
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h3 class="text-lg font-semibold">Actresses ({actresses.length})</h3>
		<Button onclick={openAddActress} size="sm">
			{#snippet children()}
				<Plus class="h-4 w-4 mr-2" />
				Add Actress
			{/snippet}
		</Button>
	</div>

	{#if actresses.length === 0}
		<div class="text-center py-8 text-muted-foreground border-2 border-dashed rounded-lg">
			<p>No actresses added</p>
			<Button onclick={openAddActress} size="sm" class="mt-2">
				{#snippet children()}
					<Plus class="h-4 w-4 mr-2" />
					Add First Actress
				{/snippet}
			</Button>
		</div>
	{:else}
		<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
			{#each actresses as actress, index}
				<Card class="p-3 hover:shadow-md transition-shadow">
					<div class="space-y-2">
						{#if actress.thumb_url}
							<img
								src={actress.thumb_url}
								alt={getFullName(actress)}
								class="w-full aspect-[2/3] object-cover rounded"
								onerror={(e) => {
									e.currentTarget.src =
										'https://via.placeholder.com/200x300?text=No+Image';
								}}
							/>
						{:else}
							<div
								class="w-full aspect-[2/3] bg-accent rounded flex items-center justify-center text-xs text-muted-foreground"
							>
								No Image
							</div>
						{/if}

						<div class="space-y-1">
							<p class="font-medium text-sm truncate" title={getFullName(actress)}>
								{getFullName(actress)}
							</p>
							{#if actress.japanese_name}
								<p class="text-xs text-muted-foreground truncate" title={actress.japanese_name}>
									{actress.japanese_name}
								</p>
							{/if}
						</div>

						<div class="flex gap-1">
							<Button variant="outline" size="sm" onclick={() => openEditActress(index)} class="flex-1">
								{#snippet children()}
									<Edit class="h-3 w-3" />
								{/snippet}
							</Button>
							<Button
								variant="outline"
								size="sm"
								onclick={() => removeActress(index)}
								class="flex-1 text-destructive hover:bg-destructive/10"
							>
								{#snippet children()}
									<Trash2 class="h-3 w-3" />
								{/snippet}
							</Button>
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Edit/Add Actress Modal -->
{#if showEditModal}
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
		<Card class="w-full max-w-2xl flex flex-col max-h-[90vh]">
			<!-- Header -->
			<div class="p-6 border-b flex items-center justify-between">
				<h2 class="text-xl font-bold">
					{editingIndex !== null ? 'Edit Actress' : 'Add Actress'}
				</h2>
				<Button variant="ghost" size="icon" onclick={cancelEdit}>
					{#snippet children()}
						<X class="h-4 w-4" />
					{/snippet}
				</Button>
			</div>

			<!-- Body -->
			<div class="flex-1 overflow-auto p-6 space-y-4">
				<div class="grid md:grid-cols-2 gap-6">
					<!-- Left: Form -->
					<div class="space-y-4">
						<div>
							<label class="text-sm font-medium mb-1 block">First Name</label>
							<input
								type="text"
								bind:value={editingActress.first_name}
								placeholder="e.g., Yume"
								class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
							/>
						</div>

						<div>
							<label class="text-sm font-medium mb-1 block">Last Name</label>
							<input
								type="text"
								bind:value={editingActress.last_name}
								placeholder="e.g., Nishimiya"
								class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
							/>
						</div>

						<div>
							<label class="text-sm font-medium mb-1 block">Japanese Name</label>
							<input
								type="text"
								bind:value={editingActress.japanese_name}
								placeholder="e.g., 西宮ゆめ"
								class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all"
							/>
						</div>

						<div>
							<label class="text-sm font-medium mb-1 block">Thumbnail URL</label>
							<input
								type="url"
								bind:value={editingActress.thumb_url}
								placeholder="https://..."
								class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
							/>
						</div>
					</div>

					<!-- Right: Preview -->
					<div>
						<label class="text-sm font-medium mb-1 block">Preview</label>
						<Card class="p-3">
							{#if editingActress.thumb_url}
								<img
									src={editingActress.thumb_url}
									alt={getFullName(editingActress) || 'Preview'}
									class="w-full aspect-[2/3] object-cover rounded mb-2"
									onerror={(e) => {
										e.currentTarget.src =
											'https://via.placeholder.com/200x300?text=Invalid+URL';
									}}
								/>
							{:else}
								<div
									class="w-full aspect-[2/3] bg-accent rounded flex items-center justify-center text-sm text-muted-foreground mb-2"
								>
									No Thumbnail
								</div>
							{/if}
							<p class="font-medium text-sm truncate">
								{getFullName(editingActress) || 'Name'}
							</p>
							{#if editingActress.japanese_name}
								<p class="text-xs text-muted-foreground truncate">
									{editingActress.japanese_name}
								</p>
							{/if}
						</Card>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="p-6 border-t flex items-center justify-end gap-3">
				<Button variant="outline" onclick={cancelEdit}>
					{#snippet children()}Cancel{/snippet}
				</Button>
				<Button onclick={saveActress}>
					{#snippet children()}
						<Save class="h-4 w-4 mr-2" />
						{editingIndex !== null ? 'Save Changes' : 'Add Actress'}
					{/snippet}
				</Button>
			</div>
		</Card>
	</div>
{/if}
