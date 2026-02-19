<script lang="ts">
	import SettingsSection from '$lib/components/settings/SettingsSection.svelte';
	import SettingsSubsection from '$lib/components/settings/SettingsSubsection.svelte';
	import FormTextInput from '$lib/components/settings/FormTextInput.svelte';
	import FormToggle from '$lib/components/settings/FormToggle.svelte';

	interface Props {
		config: any;
	}

	let { config }: Props = $props();
</script>

<SettingsSection title="File Operations" description="Control how Javinizer organizes and moves your files" defaultExpanded={false}>
	<FormToggle
		label="Move to folder"
		description="Create a dedicated folder for each movie and move files into it"
		checked={config.output.move_to_folder ?? true}
		onchange={(val) => {
			config.output.move_to_folder = val;
		}}
	/>

	<FormToggle
		label="Rename file"
		description="Rename video files according to the file naming template"
		checked={config.output.rename_file ?? true}
		onchange={(val) => {
			config.output.rename_file = val;
		}}
	/>

	<FormToggle
		label="Rename folder in place"
		description="Rename the parent folder without moving files to a new location"
		checked={config.output.rename_folder_in_place ?? false}
		onchange={(val) => {
			config.output.rename_folder_in_place = val;
		}}
	/>

	<SettingsSubsection title="Subtitle Handling">
		<FormToggle
			label="Move subtitles"
			description="Automatically move subtitle files (.srt, .ass, etc.) with video files"
			checked={config.output.move_subtitles ?? false}
			onchange={(val) => {
				config.output.move_subtitles = val;
			}}
		/>

		<FormTextInput
			label="Subtitle extensions"
			description="Comma-separated list of subtitle file extensions to move with videos"
			value={config.output.subtitle_extensions?.join(', ') ?? '.srt, .ass, .ssa, .sub, .vtt'}
			placeholder=".srt, .ass, .ssa, .sub, .vtt"
			onchange={(val) => {
				config.output.subtitle_extensions = val
					.split(',')
					.map((s) => s.trim())
					.filter((s) => s.length > 0);
			}}
		/>
	</SettingsSubsection>
</SettingsSection>
