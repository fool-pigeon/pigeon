<script lang="ts" module>
	export type OpenedFS = {
		fs: FSItem;
		saved: boolean;
	};
	class EditorState {
		// state of opened files
		ofs = $state<OpenedFS[]>([]);
		// current fs index
		index = $state(-1);
		reload() {
			FileService.GetOpenedState()
				.then((val) => {
					this.ofs = [];
					for (const v of val ?? []) {
						const fs = fileExplorer.find(v.path);
						if (!fs) {
							FileService.Close(v.path);
							continue;
						}
						this.ofs.push({ fs, saved: v.saved });
					}
				})
				.catch(msgAlert.error);
		}
		close(i: number) {
			const of = this.ofs[i];
			FileService.Close(of.fs.id)
				.then(() => {
					this.ofs.splice(i, 1);
					if (i===this.index) {
						this.selectCurrent(i > 0 ? i - 1 : i + 1);
					}else if (i<this.index){
						this.selectCurrent(this.index-1)
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
	const eState = new EditorState();
	let isRawMode = $state(false);
	let isFlexRow = $state(true);
	let activedFS = $derived(fileExplorer.getActivedFS());
	let routes = $derived(fileExplorer.findRoutes(activedFS?.id));
</script>

<script lang="ts">
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

	import { FileService } from '@services';
	import { fileExplorer, fsItemDisplay, type FSItem } from '../navigation/FileExplorer.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import { onDestroy, onMount, untrack } from 'svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import { modal } from '../base/Modal.svelte';
	import { Events } from '@wailsio/runtime';
	import { MenuEvent } from '@bindings/gui';
	import { DataEvent } from '@lib/enum/events';
	import { ModeContains } from '@lib/utils/asserts';
	import { NodeMode } from '@bindings/service';
	import FileView from './FileView.svelte';
	import HttpReqView from './HttpReqView.svelte';
	onMount(() => {
		Events.Once(DataEvent.FileExplorerLoaded, () => eState.reload());
	});
	onDestroy(() => FileService.CloseAll());
	$effect(() => {
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		if (untrack(() => eState.includesPath(activedFS.id))) {
			const i = untrack(()=>eState.findIndex(activedFS.id));
			if (i >= 0 && eState.index !== i) {
				eState.selectCurrent(i);
			}
			return;
		}
		FileService.Open(activedFS.id)
			.then(() => {
				eState.ofs.push({ fs: activedFS, saved: true });
				eState.selectCurrent(eState.ofs.length-1);
			})
			.catch(msgAlert.error);
	});
</script>

<div class="w-full h-full flex flex-col">
	{#if eState.ofs.length <= 0}
		<!-- TODO: user tips in editor -->
		<div class="w-full h-full flex justify-center items-center text-2xl">
			<div>Some tips like shortcuts and description here</div>
		</div>
	{:else}
		<div class="select-none">
			<!-- top bar nav -->
			<div class="flex gap-1 justify-between">
				<!-- top bar file list -->
				<ScrollFlex>
					{#snippet content()}
						{#each eState.ofs as of, i (of.fs.id)}
							<div
								role="button"
								tabindex="0"
								class="p-1 flex gap-0.5 items-center cursor-pointer
									text-base-content/60 hover:text-primary hover:bg-base-300
									{activedFS && activedFS.id === of.fs.id ? 'bg-primary/30' : ''}"
								onkeydown={() => {}}
								onclick={() => eState.selectCurrent(i)}
							>
								{@render fsItemDisplay(of.fs)}
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
													eState.close(i);
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
														Events.Emit(MenuEvent.Save);
														// FIXME: should close after saved
													}}>Save</button
												>
											</div>
										</div>
									</div>
								{/snippet}
								<button
									class="p-0 size-5 btn btn-ghost btn-sm text-base-content/30 hover:btn-error group"
									onclick={(e) => {
										e.stopPropagation();
										if (of.saved) {
											eState.close(i);
											return;
										}
										modal.open(UnsavedChange, 'Unsaved Change');
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
				<!-- TODO: complete functions for top btns -->
				<div class="flex border border-primary/40">
					<button
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
					</button>
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
					<button
						title={'more'}
						class="size-8 p-0 flex justify-center items-center cursor-pointer hover:bg-base-content/10"
						onclick={() => {}}
					>
						<IconEllipsis class="size-5" />
					</button>
				</div>
			</div>
			<!-- breadcrumb -->
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
		</div>
		<!-- content -->
		<!-- TODO: fill content of editor -->
		{#if eState.index < 0 || eState.index > eState.ofs.length}{:else}
		<!-- FOR TEST -->
		<div>Form Mode</div>
		<div>{JSON.stringify(eState.index, null, 2)}</div>
		<div>{JSON.stringify(eState.ofs[eState.index], null, 2)}</div>
		<div class="grow">
			{let mode=$derived(eState.ofs[eState.index].fs.additional)}
			{#key eState.index}
					{#if isRawMode}
						<FileView bind:cur={eState.ofs[eState.index]} />
						{:else if !mode}
						<div>Unkown File</div>
						{:else if ModeContains(mode, NodeMode.HttpReq)}
						<HttpReqView bind:cur={eState.ofs[eState.index]} />
					{/if}
				{/key}
			</div>
		{/if}
	{/if}
</div>
