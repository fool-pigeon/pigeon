<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconInfo from '@lucide/svelte/icons/info';
	import IconPalette from '@lucide/svelte/icons/palette';
	import IconMonitor from '@lucide/svelte/icons/monitor';
	import IconFolderOpenDot from '@lucide/svelte/icons/folder-open-dot';

	import { PathSelectService, SettingsService, ThemeMode } from '@services';
	import SettingFieldWrapper, { FieldHeader } from '@comp/form/SettingFieldWrapper.svelte';
	import { Events, WML } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { settings } from '@comp/context/AppSettings.svelte';
	import { MenuEvent } from '@cmd-gui/gui';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import ThemePreviewPage from '@comp/pages/ThemePreviewPage.svelte';

	onMount(() => {
		WML.Reload();
	});
	const LightThemes = [
		'light',
		'cupcake',
		'bumblebee',
		'emerald',
		'corporate',
		'retro',
		'garden',
		'lofi',
		'pastel',
		'fantasy',
		'wireframe',
		'cmyk',
		'autumn',
		'acid',
		'lemonade',
		'winter',
		'caramellatte',
		'silk',
		'nord',
		'valentine',
	] as const;
	const DarkThemes = [
		'dark',
		'synthwave',
		'cyberpunk',
		'halloween',
		'forest',
		'aqua',
		'black',
		'luxury',
		'dracula',
		'business',
		'night',
		'coffee',
		'dim',
		'sunset',
		'abyss',
	] as const;
	const Parts = [
		{ label: 'Info', icon: IconInfo, content: SnippetInfo },
		{ label: 'Theme', icon: IconPalette, content: SnippetTheme },
		{ label: 'Interface', icon: IconMonitor, content: SnippetInterface },
	] as const;
	type Part = (typeof Parts)[number]['label'];
	let activedPart = $state<Part>('Info');
	// for debug view
	// Events.Emit(MenuEvent.Preferences);
	// activedPart = 'Theme';
</script>

<div class="w-lvw h-lvh bg-base-100">
	<div class="w-full h-full flex">
		<div class="shrink-0 h-full max-h-full p-0 flex flex-col border-r border-base-content/20">
			<button
				class="w-full btn btn-soft btn-success"
				onclick={() => SettingsService.SetSettings(settings)}>Save</button
			>
			<div class="grow min-h-0 menu p-0 overflow-auto">
				{#each Parts as p (p.label)}
					{const PartIcon = p.icon}
					<li>
						<button
							class="flex gap-0.5 items-center
								{activedPart === p.label ? 'text-primary-content bg-primary/80' : ''}"
							onclick={() => (activedPart = p.label)}
						>
							<PartIcon class="size-4" />
							<span>{p.label}</span>
						</button>
					</li>
				{/each}
			</div>
		</div>
		<div class="grow h-full min-h-0 p-1 overflow-auto space-y-6 text-sm bg-base-300">
			<!-- Check settings loaded -->
			{#if settings.theme}
				{#each Parts as p}
					{#if p.label === activedPart}
						{@render p.content()}
					{/if}
				{/each}
			{/if}
		</div>
	</div>
</div>

{#snippet PartTitle(title: string)}
	<div class="text-base-content/60 text-[1rem]">{title}</div>
{/snippet}
{#snippet SnippetInfo()}
	{#await SettingsService.GetPathsInfo() then pathInfo}
		<SettingFieldWrapper>
			{#snippet content()}
				<div class="text-base-content/60 text-xl">App Paths</div>
				{#each Object.entries(pathInfo) as p}
					<div class="flex items-center">
						<div>{p[0]} path:</div>
						<div class="max-w-full flex gap-0.5 justify-end items-center">
							<div class="truncate text-sm text-base-content/60">{p[1]}</div>
							<button
								title="reveal in file manager"
								class="cursor-pointer text-base-content/50 hover:text-base-content hover:bg-base-content/20"
								onclick={() => PathSelectService.ShowInFileManager(p[1])}
							>
								<IconFolderOpenDot class="size-3.5 " />
							</button>
						</div>
					</div>
				{/each}
			{/snippet}
		</SettingFieldWrapper>
	{/await}
{/snippet}
{#snippet SnippetTheme()}
	<SettingFieldWrapper>
		{#snippet content()}
			{@render PartTitle('Theme')}
			<!-- Theme color schema -->
			<div
				class="max-w-full flex flex-col gap-1
				*:flex *:justify-between *:items-center @container *:@max-xs:flex-col
				[&_select]:select [&_select]:select-sm [&_select]:optional:w-auto [&_select]:optional:max-w-full [&_select]:justify-center"
			>
				<div>
					<div>Theme Mode</div>
					<select bind:value={settings.theme.mode}>
						<option value={ThemeMode.ThemeAuto}>Auto</option>
						<option value={ThemeMode.ThemeLight}>Light</option>
						<option value={ThemeMode.ThemeDark}>Dark</option>
					</select>
				</div>
				<div>
					<div>Light Theme</div>
					<select bind:value={settings.theme.lightTheme}>
						{#each LightThemes as t}
							<option value={t}>{t}</option>
						{/each}
					</select>
				</div>
				<div>
					<div>Dark Theme</div>
					<select bind:value={settings.theme.darkTheme}>
						{#each DarkThemes as t}
							<option value={t}>{t}</option>
						{/each}
					</select>
				</div>
			</div>
			<div
				class="max-w-full flex flex-col gap-1
				*:flex *:justify-between *:items-center @container *:@max-xs:flex-col
				[&>label]:cursor-pointer
				[&>:first-child>.toggle]:toggle-warning
				[&_input[type=range]]:range [&_input[type=range]]:range-xs"
			>
				<label>
					<span>Force Custom Radius</span>
					<input type="checkbox" bind:checked={settings.theme.forceEnableRadius} class="toggle" />
				</label>
				{#if settings.theme.forceEnableRadius}
					<div>
						{@render FieldHeader('Boxes Radius', 'card,modal,alert')}
						<div>
							<input
								type="range"
								bind:value={settings.theme.radiusBoxes}
								min="0"
								max="2"
								step="0.5"
							/>
						</div>
					</div>
					<div>
						{@render FieldHeader('Fields Radius', 'button,input,select,tab')}
						<div>
							<input
								type="range"
								bind:value={settings.theme.radiusFields}
								min="0"
								max="2"
								step="0.5"
							/>
						</div>
					</div>
					<div>
						{@render FieldHeader('Selectors Radius', 'checkbox,toggle,badge')}
						<div>
							<input
								type="range"
								bind:value={settings.theme.radiusSelectors}
								min="0"
								max="2"
								step="0.5"
							/>
						</div>
					</div>
				{/if}
			</div>
			<div
				class="max-w-full flex flex-col gap-1
				*:flex *:justify-between *:items-center @container *:@max-xs:flex-col
				[&>label]:cursor-pointer
				[&>:first-child>.toggle]:toggle-warning"
			>
				<label>
					<span>Force Custom Effect</span>
					<input type="checkbox" bind:checked={settings.theme.forceEnableEffect} class="toggle" />
				</label>
				{#if settings.theme.forceEnableEffect}
					<label>
						{@render FieldHeader('Depth Effect', '3D depth on fields & selectors')}
						<input type="checkbox" bind:checked={settings.theme.enableDepthEffect} class="toggle" />
					</label>
				{/if}
			</div>
		{/snippet}
	</SettingFieldWrapper>
	<SettingFieldWrapper>
		{#snippet content()}
			{@render PartTitle('Theme Preview')}
			<div>
				<div class="w-full p-1 border border-dashed border-base-content/20">
					<div><ThemePreviewPage /></div>
				</div>
			</div>
		{/snippet}
	</SettingFieldWrapper>
{/snippet}
{#snippet SnippetInterface()}
	<SettingFieldWrapper>
		{#snippet content()}
			<div class="text-base-content/60 text-xl">Editor</div>
			<label class="cursor-pointer items-center">
				{@render FieldHeader(
					'Prevent Wrap Line',
					'Enabling this option will prevent automatic line breaks in the text editor',
				)}
				<input
					type="checkbox"
					bind:checked={settings.interface.editor.preventWrapLine}
					class="toggle"
				/>
			</label>
			<div>
				<div class="flex flex-col">
					<div class="text-base-content/60">Preview Text Editor</div>
					<div class="max-w-full w-xs">
						<CMTextArea
							value={'111111111111111111111111111111111111111111111111111111111111111111111111111\n22222222222222222222222222222222222222222222222222222222222222222'}
							readonly
						/>
					</div>
				</div>
			</div>
		{/snippet}
	</SettingFieldWrapper>
{/snippet}
