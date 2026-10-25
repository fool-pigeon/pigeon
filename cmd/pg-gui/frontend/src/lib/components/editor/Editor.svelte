<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	// Icons
	import IconCircle from '@lucide/svelte/icons/circle';
	import IconX from '@lucide/svelte/icons/x';
	import IconFileTypeCorner from '@lucide/svelte/icons/file-type-corner';
	import IconForm from '@lucide/svelte/icons/form';
	import IconLayoutCol from '@lucide/svelte/icons/columns-2';
	import IconLayoutRow from '@lucide/svelte/icons/rows-2';
	import IconEllipsis from '@lucide/svelte/icons/ellipsis';
	import IconChevronRight from '@lucide/svelte/icons/chevron-right';
	import IconSaveAll from '@lucide/svelte/icons/save-all';
	import IconTriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import IconKeyboard from '@lucide/svelte/icons/keyboard';
	import IconInfo from '@lucide/svelte/icons/info';

	import { FileService, FsEvent, type OpenedState } from '@services';
	import { fileExplorer, fsItemDisplay, type FSItem } from '../navigation/FileExplorer.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import { onMount, untrack } from 'svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { modal } from '../base/Modal.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@cmd-gui/gui';
	import { DataEvent } from '@lib/enum/events';
	import { ModeContains } from '@lib/utils/asserts';
	import { NodeMode } from '@cmd-gui/service';
	import FileView from './FileView.svelte';
	import ReqView from './ReqView.svelte';

	export const EditorEvent = {
		Reload: 'editor:reload',
	} as const;
	export type EditorEvent = (typeof EditorEvent)[keyof typeof EditorEvent];
	export type OpenedFS = {
		fs: FSItem;
		saved: boolean;
	};
	class EditorState {
		// state of opened files
		ofs = $state<OpenedFS[]>([]);
		// current fs index
		index = $state(-1);
		closeIndex = $state(-1);
		reload() {
			FileService.GetOpenedPaths()
				.then((val) => {
					this.ofs = openedPathsToOFS(val);
					this.selectCurrent(this.index);
				})
				.catch(msgAlert.error);
		}
		Close(i: number) {
			if (i < 0 || i >= this.ofs.length) {
				return;
			}
			this.closeIndex = i;
			if (this.ofs[i].saved) {
				this.close();
				return;
			}
			modal.open(UnsavedChange, 'Unsaved Change');
		}
		close() {
			const i = this.closeIndex;
			const of = this.ofs[i];
			FileService.Close(of.fs.id)
				.then(() => {
					const cur = this.index;
					this.ofs.splice(i, 1);
					if (i === cur) {
						this.selectCurrent(cur > 0 ? cur - 1 : cur);
					} else if (i < cur) {
						this.selectCurrent(cur - 1);
					}
				})
				.catch(msgAlert.error);
		}
		includesPath(path: string): boolean {
			return this.ofs.some((v) => v.fs.id === path);
		}
		findIndex(path: string): number {
			return this.ofs.findIndex((v) => v.fs.id === path);
		}
		selectCurrent(index: number) {
			if (index < 0 || index >= this.ofs.length) {
				fileExplorer.setActivedFS(undefined);
				this.index = -1;
				return;
			}
			this.index = index;
			const of = this.ofs[index];
			fileExplorer.expandToItem(of.fs);
			fileExplorer.setSelectedFS(of.fs);
		}
	}
	function openedPathsToOFS(val: OpenedState[] | null): OpenedFS[] {
		const ofs = [];
		for (const v of val ?? []) {
			const fs = fileExplorer.find(v.path);
			if (!fs) {
				FileService.Close(v.path);
				continue;
			}
			ofs.push({ fs, saved: v.saved });
		}
		return ofs;
	}
	const eState = new EditorState();
	let isRawMode = $state(false);
	let isFlexRow = $state(true);
	let activedFS = $derived(fileExplorer.getActivedFS());
	let routes = $derived(fileExplorer.findRoutes(activedFS?.id));
</script>

<script lang="ts">
	onMount(() => {
		const unLoad = Events.Once(DataEvent.FileExplorerLoaded, () => eState.reload());
		const unReload = Events.On(EditorEvent.Reload, () => eState.reload());
		const unClose = Events.On(MenuEvent.CloseEditor, () => eState.Close(eState.index));
		const unUpdateOpenedPaths = Events.On(FsEvent.FsUpdatePaths, (v) => {
			const data = v.data as OpenedState[] | null;
			eState.ofs = openedPathsToOFS(data);
			eState.selectCurrent(eState.index);
		});
		return () => {
			unLoad();
			unReload();
			unClose();
			unUpdateOpenedPaths();
			// FileService.CloseAll();
		};
	});
	$effect(() => {
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		if (untrack(() => eState.includesPath(activedFS.id))) {
			const i = untrack(() => eState.findIndex(activedFS.id));
			if (i >= 0 && eState.index !== i) {
				eState.selectCurrent(i);
			}
			return;
		}
		FileService.Open(activedFS.id)
			.then(() => {
				eState.ofs.push({ fs: activedFS, saved: true });
				eState.selectCurrent(eState.ofs.length - 1);
			})
			.catch(msgAlert.error);
	});
</script>

<div class="w-full h-full flex flex-col">
	{#if eState.ofs.length <= 0}
		{@render Tips()}
	{:else}
		<div class="select-none">
			<!-- top bar nav -->
			{@render TopBar()}
			<!-- breadcrumb -->
			{@render Breadcrumb()}
		</div>
		<!-- content -->
		{#if eState.index < 0 || eState.index >= eState.ofs.length}
			{@render Tips()}
		{:else}
			<div class="grow min-h-0">
				{let mode = $derived(eState.ofs[eState.index].fs.additional ?? 0)}
				{#key eState.index}
					{#if isRawMode}
						<!-- the cur should not be bindable -->
						<FileView openedFS={eState.ofs[eState.index]} />
					{:else if ModeContains(mode, NodeMode.Request)}
						<!-- the cur should not be bindable -->
						<ReqView openedFS={eState.ofs[eState.index]} bind:isFlexRow />
					{:else}
						<div class="w-full h-1/3 flex items-center justify-center text-2xl text-info">
							<div class="flex gap-1 items-center">
								<IconInfo class="size-6" />the file not support form editor yet.
							</div>
						</div>
					{/if}
				{/key}
			</div>
		{/if}
	{/if}
</div>
{#snippet Tips()}
	<div class="w-full h-full flex flex-col justify-center items-center">
		<IconKeyboard class="size-32 text-base-content/50" strokeWidth={5} nonScalingStroke />
		<div class="inline-grid grid-cols-[auto_auto] gap-x-2 gap-y-0.5">
			<span class="text-right">New Collection</span>
			<span class="text-left">
				<kbd class="kbd">⌘/Ctrl</kbd> + <kbd class="kbd">Shift</kbd> + <kbd class="kbd">N</kbd>
			</span>
			<span class="text-right">Open Collection</span>
			<span class="text-left">
				<kbd class="kbd">⌘/Ctrl</kbd> + <kbd class="kbd">Shift</kbd> + <kbd class="kbd">O</kbd>
			</span>
			<span class="text-right">Preferences</span>
			<span class="text-left">
				<kbd class="kbd">⌘/Ctrl</kbd> + <kbd class="kbd">,</kbd>
			</span>
			<span class="text-right">Quit</span>
			<span class="text-left">
				<kbd class="kbd">⌘/Ctrl</kbd> + <kbd class="kbd">Q</kbd>
			</span>
		</div>
	</div>
{/snippet}
{#snippet TopBar()}
	<div class="flex gap-1 justify-between">
		<!-- top bar file list -->
		<ScrollFlex>
			{#snippet content()}
				{#each eState.ofs as of, i (of.fs.id)}
					{let isActived = $derived(activedFS && activedFS.id === of.fs.id)}
					{let textStyle = $derived(isActived ? 'text-primary-content' : 'text-base-content/60')}
					<div
						role="button"
						tabindex="0"
						class="group p-1 flex gap-0.5 items-center cursor-pointer
						hover:text-primary hover:bg-base-300
						{textStyle} {isActived ? 'bg-primary' : ''}"
						onkeydown={() => {}}
						onclick={() => eState.selectCurrent(i)}
					>
						{@render fsItemDisplay(of.fs)}
						<button
							class="p-0 size-5 btn btn-sm btn-ghost hover:btn-error {textStyle} group-hover:text-primary"
							onclick={(e) => {
								e.stopPropagation();
								eState.Close(i);
							}}
						>
							{#if of.saved}
								<IconX class="size-5" />
							{:else}
								<IconCircle class="p-1 size-5 group-hover:hidden fill-current border-none" />
								<IconX class="size-5 hidden group-hover:inline" />
							{/if}
						</button>
					</div>
				{/each}
			{/snippet}
		</ScrollFlex>
		<!-- top function btns -->
		<div class="flex border border-primary/40">
			<!-- <button
				title={'Save all'}
				class="size-8 p-0 flex justify-center items-center cursor-pointer hover:bg-base-content/10"
				onclick={() => {
					FileService.SaveAll()
						.then(() => {
							eState.reload();
							msgAlert.success('Saved all');
						})
						.catch(msgAlert.error);
				}}
			>
				<IconSaveAll class="size-5" />
			</button> -->
			<label
				title={isRawMode ? 'To form mode' : 'To file mode'}
				class="size-8 swap swap-rotate hover:bg-base-content/10"
			>
				<input type="checkbox" bind:checked={isRawMode} />
				<div class="swap-on">
					<IconForm class="size-5" />
				</div>
				<div class="swap-off">
					<IconFileTypeCorner class="size-5" />
				</div>
			</label>
			<label
				title={isFlexRow ? 'To column layout' : 'To row layout'}
				class="size-8 swap swap-rotate hover:bg-base-content/10"
			>
				<input type="checkbox" bind:checked={isFlexRow} />
				<div class="swap-on">
					<IconLayoutRow class="size-5" />
				</div>
				<div class="swap-off">
					<IconLayoutCol class="size-5" />
				</div>
			</label>
			<!-- TODO:more function btn -->
			<!-- <button
				title={'more'}
				class="size-8 p-0 flex justify-center items-center cursor-pointer hover:bg-base-content/10"
				onclick={() => {}}
			>
				<IconEllipsis class="size-5" />
			</button> -->
		</div>
	</div>
{/snippet}

{#snippet Breadcrumb()}
	<div class="pt-1 pl-1 border-b border-primary/50">
		<ScrollFlex>
			{#snippet content()}
				{#each routes as r, i (r.id)}
					<div class="flex items-baseline">
						<button
							class="hover:underline hover:cursor-pointer"
							onclick={() => {
								fileExplorer.expandToItem(r);
								fileExplorer.setSelectedFS(r);
							}}
						>
							{@render fsItemDisplay(r)}
						</button>
						{#if i < routes.length - 1}
							<IconChevronRight class="size-4 text-base-content/70" />
						{/if}
					</div>
				{/each}
			{/snippet}
		</ScrollFlex>
	</div>
{/snippet}

{#snippet UnsavedChange()}
	<div class="p-2 flex flex-col">
		<div class="p-1 flex items-center">
			<IconTriangleAlert class="size-8 text-warning" />
			<span class="text-2xl">Wait a moment..</span>
		</div>
		<div>You have unsaved changes in this file.</div>
		<div class="mt-4 flex justify-between">
			<button
				class="rounded-none btn btn-error p-2"
				onclick={() => {
					eState.close();
					modal.close();
				}}>Not Save</button
			>
			<div class="flex gap-1">
				<button
					class="rounded-none btn btn-ghost p-2 hover:bg-base-content/30"
					onclick={() => modal.close()}>Cancel</button
				>
				<button
					class="rounded-none btn btn-info p-2 hover:btn-success"
					onclick={() => {
						const unregister = Events.Once(DataEvent.EditorSaved, () => {
							eState.close();
							modal.close();
							unregister();
						});
						Events.Emit(MenuEvent.Save);
					}}>Save</button
				>
			</div>
		</div>
	</div>
{/snippet}
