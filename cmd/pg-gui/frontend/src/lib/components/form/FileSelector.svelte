<script lang="ts">
	import IconFile from '@lucide/svelte/icons/file';
	import IconTrash from '@lucide/svelte/icons/trash';
	import IconUpload from '@lucide/svelte/icons/upload';

	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { PathSelectService } from '@services';

	let { path = $bindable() }: { path: string } = $props();
	let name = $derived(path.split(/[\\/]/).pop() ?? '');
	let hasFile = $derived(path !== '');
</script>

<div
	class="p-1 min-w-0 max-w-48 flex gap-1 justify-between items-center text-base-content/70 border border-base-content/15"
	title={path}
>
	<IconFile class="size-4 min-w-4" />
	<div class="min-w-0 max-w-38">
		{#if hasFile}
			<p class="line-clamp-1">{name}</p>
		{:else}
			<p>no file</p>
		{/if}
	</div>
	<button
		class="p-0 min-w-4 max-w-5 appearance-none cursor-pointer text-base-content/80
        {hasFile ? 'hover:text-error' : 'hover:text-info'}"
		title={hasFile ? 'remove file' : 'add file'}
		onclick={() => {
			if (hasFile) {
				path = '';
				name = '';
				return;
			}
			PathSelectService.SelectFile()
				.then((v) => (path = v))
				.catch(msgAlert.error);
		}}
	>
		{#if hasFile}
			<IconTrash class="size-4" />
		{:else}
			<IconUpload class="size-4" />
		{/if}
	</button>
</div>
