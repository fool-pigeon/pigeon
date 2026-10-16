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
			// set daisyui default theme
			if (v.theme.lightTheme === '') v.theme.lightTheme = 'light';
			if (v.theme.darkTheme === '') v.theme.darkTheme = 'dark';

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
