<script lang="ts">
	import SettingsSection from '$lib/components/settings/SettingsSection.svelte';

	interface Props {
		config: any;
		inputClass: string;
	}

	let { config, inputClass }: Props = $props();
</script>

<SettingsSection title="Server Settings" description="Configure API server host and port" defaultExpanded={false}>
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
</SettingsSection>
