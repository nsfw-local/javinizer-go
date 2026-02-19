<script lang="ts">
	import SettingsSection from '$lib/components/settings/SettingsSection.svelte';

	interface Props {
		config: any;
		inputClass: string;
	}

	let { config, inputClass }: Props = $props();
</script>

<SettingsSection title="Logging Settings" description="Configure logging level, format, and output destination" defaultExpanded={false}>
	<div class="space-y-4">
		<div>
			<label class="block text-sm font-medium mb-2" for="log-level">Log Level</label>
			<select id="log-level" bind:value={config.logging.level} class={inputClass}>
				<option value="debug">Debug</option>
				<option value="info">Info</option>
				<option value="warn">Warning</option>
				<option value="error">Error</option>
			</select>
		</div>

		<div>
			<label class="block text-sm font-medium mb-2" for="log-format">Log Format</label>
			<select id="log-format" bind:value={config.logging.format} class={inputClass}>
				<option value="text">Text</option>
				<option value="json">JSON</option>
			</select>
		</div>

		<div>
			<label class="block text-sm font-medium mb-2" for="log-output">Log Output</label>
			<input
				id="log-output"
				type="text"
				bind:value={config.logging.output}
				class={inputClass}
				placeholder="stdout or file path"
			/>
			<p class="text-xs text-muted-foreground mt-1">
				Use "stdout" for console or provide a file path
			</p>
		</div>
	</div>
</SettingsSection>
