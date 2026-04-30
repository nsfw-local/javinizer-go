<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { createApiTokensQuery } from '$lib/query/queries';
	import { createToken, revokeToken, regenerateToken } from '$lib/api/tokens';
	import type { CreateTokenResponse } from '$lib/types/token';
	import { toastStore } from '$lib/stores/toast';
	import { confirmDialog } from '$lib/stores/dialog.svelte';
	import SettingsSection from '$lib/components/settings/SettingsSection.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Plus, Loader2, Trash2, RefreshCw } from 'lucide-svelte';

	interface Props {
		onTokenDisplay?: (response: CreateTokenResponse) => void;
	}

	let { onTokenDisplay }: Props = $props();

	const queryClient = useQueryClient();

	const tokensQuery = createApiTokensQuery();
	let tokens = $derived(tokensQuery.data?.tokens ?? []);
	let loading = $derived(tokensQuery.isPending);
	let error = $derived<string | null>(tokensQuery.error?.message ?? null);

	let newTokenName = $state('');

	const createTokenMutation = createMutation(() => ({
		mutationFn: (name?: string) => createToken(name),
		onSuccess: (data: CreateTokenResponse) => {
			newTokenName = '';
			toastStore.success('API token created', 3000);
			onTokenDisplay?.(data);
			void queryClient.invalidateQueries({ queryKey: ['api-tokens'] });
		},
		onError: (err: Error) => {
			toastStore.error(err.message || 'Failed to create token', 4000);
		}
	}));

	const revokeTokenMutation = createMutation(() => ({
		mutationFn: (id: string) => revokeToken(id),
		onSuccess: () => {
			toastStore.success('Token revoked', 3000);
			void queryClient.invalidateQueries({ queryKey: ['api-tokens'] });
		},
		onError: (err: Error) => {
			toastStore.error(err.message || 'Failed to revoke token', 4000);
		}
	}));

	const regenerateTokenMutation = createMutation(() => ({
		mutationFn: (id: string) => regenerateToken(id),
		onSuccess: (data: CreateTokenResponse) => {
			toastStore.success('Token regenerated', 3000);
			onTokenDisplay?.(data);
			void queryClient.invalidateQueries({ queryKey: ['api-tokens'] });
		},
		onError: (err: Error) => {
			toastStore.error(err.message || 'Failed to regenerate token', 4000);
		}
	}));

	async function handleCreate() {
		createTokenMutation.mutate(newTokenName.trim() || undefined);
	}

	async function handleRevoke(id: string, name: string) {
		const confirmed = await confirmDialog(
			'Revoke Token',
			`Are you sure you want to revoke the token "${name || id}"? This action cannot be undone.`,
			{ confirmLabel: 'Revoke', variant: 'danger' }
		);
		if (confirmed) {
			revokeTokenMutation.mutate(id);
		}
	}

	async function handleRegenerate(id: string, name: string) {
		const confirmed = await confirmDialog(
			'Regenerate Token',
			`Regenerating the token "${name || id}" will invalidate the current value. Make sure you no longer need it. Continue?`,
			{ confirmLabel: 'Regenerate', variant: 'danger' }
		);
		if (confirmed) {
			regenerateTokenMutation.mutate(id);
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return 'Never';
		try {
			return new Date(dateStr).toLocaleDateString(undefined, {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return dateStr;
		}
	}

	function handleCreateKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			handleCreate();
		}
	}
</script>

<SettingsSection title="API Tokens" description="Manage API tokens for programmatic access to the Javinizer API" defaultExpanded={false}>
	{#if loading}
		<div class="flex items-center justify-center py-8 text-muted-foreground">
			<Loader2 class="h-5 w-5 animate-spin mr-2" />
			Loading...
		</div>
	{:else if error}
		<div class="text-destructive text-sm py-4">
			Failed to load API tokens: {error}
		</div>
	{:else}
		<div class="space-y-4">
			{#if tokens.length === 0}
				<p class="text-sm text-muted-foreground py-4">
					No API tokens configured. Create one below to enable programmatic access.
				</p>
			{:else}
				<div class="relative border border-border rounded-lg overflow-hidden">
					<div class="overflow-x-auto">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b border-border bg-muted/50">
									<th class="text-left py-2 px-3 font-medium text-muted-foreground">Name</th>
									<th class="text-left py-2 px-3 font-medium text-muted-foreground">Prefix</th>
									<th class="text-left py-2 px-3 font-medium text-muted-foreground">Created</th>
									<th class="text-left py-2 px-3 font-medium text-muted-foreground">Last Used</th>
									<th class="text-right py-2 px-3 font-medium text-muted-foreground">Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each tokens as token (token.id)}
									<tr class="border-b border-border/50 hover:bg-accent/30 transition-colors">
										<td class="py-2 px-3">{#if token.name}{token.name}{:else}<span class="text-muted-foreground italic">Unnamed</span>{/if}</td>
										<td class="py-2 px-3 font-mono text-xs">{token.token_prefix}</td>
										<td class="py-2 px-3 text-xs">{formatDate(token.created_at)}</td>
										<td class="py-2 px-3 text-xs">{formatDate(token.last_used_at)}</td>
										<td class="py-2 px-3 text-right">
											<div class="flex items-center justify-end gap-1">
												<button
													type="button"
													class="text-muted-foreground hover:text-foreground transition-colors p-1 rounded"
													title="Regenerate token"
													onclick={() => handleRegenerate(token.id, token.name)}
													disabled={regenerateTokenMutation.isPending}
												>
													<RefreshCw class="h-4 w-4" />
												</button>
												<button
													type="button"
													class="text-muted-foreground hover:text-destructive transition-colors p-1 rounded"
													title="Revoke token"
													onclick={() => handleRevoke(token.id, token.name)}
													disabled={revokeTokenMutation.isPending}
												>
													<Trash2 class="h-4 w-4" />
												</button>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
				<p class="text-xs text-muted-foreground">
					{tokens.length} token{tokens.length !== 1 ? 's' : ''} active
				</p>
			{/if}

			<div class="border-t pt-4">
				<p class="text-xs text-muted-foreground mb-3">Create a new API token:</p>
				<div class="flex items-end gap-2">
					<div class="flex-1">
						<label for="token-name" class="block text-xs font-medium text-muted-foreground mb-1">Name (optional)</label>
						<input
							id="token-name"
							type="text"
							bind:value={newTokenName}
							placeholder="e.g., CI Pipeline"
							onkeydown={handleCreateKeydown}
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
						/>
					</div>
					<Button
						size="sm"
						onclick={handleCreate}
						disabled={createTokenMutation.isPending}
					>
						{#if createTokenMutation.isPending}
							<Loader2 class="h-4 w-4 animate-spin mr-1" />
						{:else}
							<Plus class="h-4 w-4 mr-1" />
						{/if}
						Create Token
					</Button>
				</div>
			</div>

			<p class="text-xs text-muted-foreground">
				Tokens use the <code class="font-mono bg-muted px-1 rounded">jv_</code> prefix. The full token value is shown only once after creation.
			</p>
		</div>
	{/if}
</SettingsSection>
