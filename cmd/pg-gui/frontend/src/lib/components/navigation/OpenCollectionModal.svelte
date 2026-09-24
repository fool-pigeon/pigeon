<script lang="ts">
	// Icons
	import IconLoaderCircle from '@lucide/svelte/icons/loader-circle';

	import { MenuEvent } from '@bindings/gui';
	import { FileExploreService, PathSelectService } from '@bindings/service';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	import { modal } from '../base/Modal.svelte';
	import { msgAlert } from '../message/AlertNotify.svelte';
	import Loading from '../base/Loading.svelte';
	type Prop = { afterLoadFunc: (rootPaths: string[]) => void };
	let { afterLoadFunc }: Prop = $props();

	let searchPath = $state('');
	let rootPaths = $state<string[]>([]);
	onMount(() => {
		const unregister = Events.On(MenuEvent.OpenCollection, () => {
			PathSelectService.SelectFolder()
				.then((val) => {
					rootPaths = [];
					searchPath = val;
					modal.open(openCollection, 'Open Collection');
				})
				.catch(msgAlert.error);
		});
		return () => {
			unregister();
		};
	});
</script>

{#snippet openCollection()}
	{#await FileExploreService.ParseRootPaths(searchPath)}
		<div class="w-full h-30">
			<Loading />
		</div>
	{:then paths}
		{#if paths !== null && paths.length > 0}
			<div class="p-4 min-w-sm">
				<div class="flex items-end justify-between">
					<div class="flex items-baseline gap-1">
						<span>Collections</span>
						<div class="badge badge-soft badge-info">{paths.length}</div>
					</div>
					<div>
						<button
							class="btn btn-soft btn-sm font-normal
                            {rootPaths.length === paths.length ? 'btn-error' : 'btn-primary'}"
							onclick={() => {
								if (rootPaths.length === paths.length) {
									rootPaths = [];
								} else {
									rootPaths = paths;
								}
							}}>{rootPaths.length === paths.length ? 'DropAll' : 'SelectAll'}</button
						>
					</div>
				</div>
				<div class="p-2 mt-2 flex items-center border border-base-content/30">
					{#each paths as path}
						<label class="label">
							<input
								type="checkbox"
								class="checkbox checkbox-sm checkbox-primary"
								bind:group={rootPaths}
								value={path}
							/>
							{path}
						</label>
					{/each}
				</div>
				<div class="mt-10 flex gap-1 justify-between items-center">
					<div>
						<span class="text-primary">{rootPaths.length} </span>
						<span class="text-base-content/50">of {paths.length} selected</span>
					</div>
					<div>
						<button class="btn btn-ghost font-normal" onclick={() => modal.close()}>Cancel</button>
						<button
							class="btn btn-primary font-normal"
							onclick={() => {
								// called afterLoadFunc and close modal
								afterLoadFunc(rootPaths);
								modal.close();
							}}>Load</button
						>
					</div>
				</div>
			</div>
		{:else}
			<div class="text-warning h-10 text-center">
				Not Found Collection in {searchPath}
			</div>
		{/if}
	{:catch err}
		<div class="text-error h-10">cause-error: {err}</div>
	{/await}
{/snippet}
