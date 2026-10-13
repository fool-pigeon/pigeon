<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts" module>
	import { SettingsService, SettingsServiceEvent, ThemeMode, type Settings } from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';

	export const settings = $state<Settings>({} as Settings);
</script>

<script lang="ts">
	onMount(() => {
		SettingsService.GetSettings().then((v) => {
			Object.assign(settings, v);
		});
		const unSettingsChanged = Events.On(SettingsServiceEvent.SettingChanged, (ev) => {
			const newSettings = ev.data as Settings;
			Object.assign(settings, newSettings);
		});

		return () => {
			unSettingsChanged();
		};
	});
</script>
