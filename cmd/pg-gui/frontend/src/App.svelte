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
	function resolveThemeMode(mode: ThemeMode): ThemeMode {
		if (mode !== ThemeMode.ThemeAuto) return mode;
		const mql = window.matchMedia('(prefers-color-scheme: dark)');
		return mql.matches ? ThemeMode.ThemeDark : ThemeMode.ThemeLight;
	}
	$effect(() => {
		if (settings.theme) {
			const theme = settings.theme;
			const mode = resolveThemeMode(theme.mode);
			document.documentElement.dataset.theme =
				mode === ThemeMode.ThemeLight ? theme.lightTheme : theme.darkTheme;

			const r = document.documentElement.style;
			if (theme.forceEnableRadius) {
				r.setProperty('--radius-box', `${theme.radiusBoxes}rem`);
				r.setProperty('--radius-field', `${theme.radiusFields}rem`);
				r.setProperty('--radius-selector', `${theme.radiusSelectors}rem`);
			} else {
				r.removeProperty('--radius-box');
				r.removeProperty('--radius-field');
				r.removeProperty('--radius-selector');
			}

			if (theme.forceEnableEffect) {
				if (theme.enableDepthEffect) r.setProperty('--depth', '1');
				else r.setProperty('--depth', '0');
			} else {
				r.removeProperty('--depth');
			}
		}
	});
</script>

<Router {routes} />
<AppSettings />
