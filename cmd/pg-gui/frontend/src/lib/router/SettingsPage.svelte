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
	import SettingFieldWrapper from '@comp/form/SettingFieldWrapper.svelte';
	import { Events, WML } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { settings } from '@comp/context/AppSettings.svelte';
	import { MenuEvent } from '@pg/gui';
	import CMTextArea from '@comp/form/CMTextArea.svelte';

	onMount(() => {
		WML.Reload();
		// for debug view
		Events.Emit(MenuEvent.Preferences);
	});
	const Parts = [
		{ label: 'Info', icon: IconInfo, content: SnippetInfo },
		{ label: 'Theme', icon: IconPalette, content: SnippetTheme },
		{ label: 'Interface', icon: IconMonitor, content: SnippetInterface },
	] as const;
	type Part = (typeof Parts)[number]['label'];
	// let activedPart = $state<Part>('Info');
	let activedPart = $state<Part>('Interface');
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
		<div class="grow h-full min-h-0 p-1 overflow-auto bg-base-300">
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
			<div class="text-base-content/60 text-xl">Theme</div>
			<div class="max-w-full flex items-center">
				<div>Theme Mode</div>
				<select
					bind:value={settings.theme.mode}
					class="select select-sm optional:w-auto optional:max-w-full justify-center outline-none"
				>
					<option value={ThemeMode.ThemeAuto}>Auto</option>
					<option value={ThemeMode.ThemeLight}>Light</option>
					<option value={ThemeMode.ThemeDark}>Dark</option>
				</select>
			</div>
		{/snippet}
	</SettingFieldWrapper>
{/snippet}
{#snippet SnippetInterface()}
	<SettingFieldWrapper>
		{#snippet content()}
			<div class="text-base-content/60 text-xl">Editor</div>
			<div class="flex items-center">
				<div class="max-w-full flex flex-col">
					<span>Prevent Wrap Line</span>
					<span class="text-xs text-base-content/60">
						Enabling this option will prevent automatic line breaks in the text editor
					</span>
				</div>
				<label>
					<input
						type="checkbox"
						bind:checked={settings.interface.editor.preventWrapLine}
						class="checkbox"
					/>
				</label>
			</div>
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
