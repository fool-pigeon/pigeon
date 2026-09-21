<script lang="ts" module>
	type RawFile = {
		// original file content
		initValue: string;
		// current file content
		value: string;
		// meta of file
		fs: FSItem;
	};
	class RawFileState {
		// raw files
		rfs = $state<RawFile[]>([]);
		// current selected raw file
		curRf = $state<RawFile>();

		/**
		 * return true if has the raw file whose fs.id equals path
		 * @param path
		 */
		includesPath(path: string): boolean {
			return this.rfs.some((r) => r.fs.id === path);
		}
		/**
		 * get raw file from rfs whose fs.id equals path
		 * @param path
		 */
		findRaw(path: string): RawFile | undefined {
			return this.rfs.filter((r) => r.fs.id === path).at(0);
		}
		/**
		 * select rf if rf not undefined
		 * @param rf
		 */
		selectCurrent(rf: RawFile | undefined) {
			this.curRf = rf;
			if (!rf) {
				fileExplorer.setActivedFS(undefined);
				return;
			}
			fileExplorer.expandToItem(rf.fs);
			fileExplorer.setSelectedFS(rf.fs);
		}
		saveRawFile(rf: RawFile) {
			WriteFile(rf.fs.id, rf.value)
				.then(() => (rf.initValue = rf.value))
				.catch(console.error);
		}
	}
	const rState = new RawFileState();
	let isRawFileMode = $state(false);
	let isLayoutRow = $state(true);
	let activedFS = $derived(fileExplorer.getActivedFS());
	let routes = $derived(fileExplorer.findRoutes(activedFS?.id));
</script>

<script lang="ts">
	import {
		fileExplorer,
		fsItemDisplay as fsDisplay,
		type FSItem,
	} from '../navigation/FileExplorer.svelte';
	import ScrollFlex from '../base/ScrollFlex.svelte';
	import FileView from './FileView.svelte';
	import { ReadFile, WriteFile } from '@bindings/service/fileservice';
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
	import { untrack } from 'svelte';

	$effect(() => {
		if (!activedFS || activedFS.type === 'group') {
			return;
		}
		if (untrack(() => rState.includesPath(activedFS.id))) {
			return;
		}
		ReadFile(activedFS.id)
			.then((val) => {
				const rf: RawFile = { initValue: val, value: val, fs: activedFS };
				rState.rfs.push(rf);
				rState.selectCurrent(rf);
			})
			.catch(console.error);
	});
</script>

<div class="w-full h-full flex flex-col">
	{#if !activedFS || activedFS.type === 'group'}
		<!-- TODO: user tips in editor -->
		<div class="w-full h-full flex justify-center items-center text-2xl">
			<div>Some tips like shortcuts and description here</div>
		</div>
	{:else}
		<div class="select-none">
			<!-- top bar nav -->
			<div class="flex gap-1 justify-between">
				<ScrollFlex>
					{#snippet content()}
						{#each rState.rfs as r, i (r.fs.id)}
							<div
								role="button"
								tabindex="0"
								class="p-1 flex gap-0.5 items-center cursor-pointer
									text-base-content/60 hover:text-primary hover:bg-base-300
									{activedFS.id === r.fs.id ? 'bg-primary/30' : ''}"
								onkeydown={() => {}}
								onclick={() => rState.selectCurrent(r)}
							>
								{@render fsDisplay(r.fs)}
								<button
									class="p-0 size-5 btn btn-ghost btn-sm text-base-content/30 hover:btn-error group"
									onclick={(e) => {
										e.stopPropagation();
										if (activedFS.id === r.fs.id) {
											rState.selectCurrent(i > 0 ? rState.rfs[i - 1] : rState.rfs[i + 1]);
										}
										rState.rfs.splice(i, 1);
									}}
								>
									{#if r.initValue === r.value}
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
							for (const rf of rState.rfs) {
								rState.saveRawFile(rf);
							}
						}}
					>
						<IconSaveAll class="size-5" />
					</button>
					<label
						title={isRawFileMode ? 'To form mode' : 'To file mode'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isRawFileMode} />
						<div class="swap-on">
							<IconForm class="size-5" />
						</div>
						<div class="swap-off">
							<IconFileTypeCorner class="size-5" />
						</div>
					</label>
					<label
						title={isLayoutRow ? 'To column layout' : 'To row layout'}
						class="size-8 swap swap-rotate hover:bg-base-content/10"
					>
						<input type="checkbox" bind:checked={isLayoutRow} />
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
									{@render fsDisplay(r)}
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
		{#if !rState.curRf}{:else}
			<div class="grow">
				{#if isRawFileMode}
					<FileView bind:value={rState.curRf.value} />
				{:else}
					<div>Form Mode</div>
					<div>{JSON.stringify(rState.rfs, null, 2)}</div>
					<div>{JSON.stringify(rState.curRf, null, 2)}</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>
