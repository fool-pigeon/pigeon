<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import AppSettings, { settings } from '@comp/context/AppSettings.svelte';
	import Main from '@lib/router/Main.svelte';
	import SettingsPage from '@lib/router/SettingsPage.svelte';
	import { ThemeMode } from '@services';
	import Router from 'svelte-spa-router';

	const routes = {
		'/': Main,
		'/settings': SettingsPage,
	};
	$effect(() => {
		if (settings.theme) {
			const theme = settings.theme;
			switch (theme.mode) {
				case ThemeMode.ThemeAuto:
					document.documentElement.removeAttribute('data-theme');
					break;
				default:
					document.documentElement.dataset.theme = theme.mode;
			}
		}
	});
</script>

<Router {routes} />
<AppSettings />
