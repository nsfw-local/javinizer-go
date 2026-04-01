<script lang="ts">
	import SettingsSection from '$lib/components/settings/SettingsSection.svelte';

	interface Props {
		config: any;
		inputClass: string;
	}

	let { config, inputClass }: Props = $props();

	// Sanitize temp_dir to prevent path traversal
	function sanitizeTempDir(value: string): string {
		// Remove leading/trailing whitespace
		value = value.trim();
		// Remove path traversal patterns
		value = value.replace(/\.\.[\\/]/g, '');
		// Remove leading/trailing path separators
		value = value.replace(/^[\\/]+|[\\/]+$/g, '');
		return value;
	}

	function handleTempDirInput(e: Event) {
		const target = e.target as HTMLInputElement;
		config.system.temp_dir = sanitizeTempDir(target.value);
	}
</script>

<SettingsSection title="Server Settings" description="Configure API server host, port, and system paths" defaultExpanded={false}>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div>
				<label class="block text-sm font-medium mb-2" for="server-host">Host</label>
				<input id="server-host" type="text" bind:value={config.server.host} class={inputClass} placeholder="localhost" />
			</div>
			<div>
				<label class="block text-sm font-medium mb-2" for="server-port">Port</label>
				<input id="server-port" type="number" bind:value={config.server.port} class={inputClass} placeholder="8080" />
			</div>
		</div>
		<div>
			<label class="block text-sm font-medium mb-2" for="system-temp-dir">Temporary Directory</label>
			<input
				id="system-temp-dir"
				type="text"
				value={config.system.temp_dir}
				oninput={handleTempDirInput}
				class={inputClass}
				placeholder="data/temp"
			/>
			<p class="text-xs text-muted-foreground mt-1">Base directory for temporary files (default: data/temp). Cannot contain path traversal patterns.</p>
		</div>
	</div>
</SettingsSection>
