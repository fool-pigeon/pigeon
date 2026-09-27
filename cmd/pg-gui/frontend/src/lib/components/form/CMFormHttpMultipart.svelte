<script lang="ts">
	// Icon
	import IconListPlus from '@lucide/svelte/icons/list-plus';
	import IconTrash from '@lucide/svelte/icons/trash';
	import IconUpload from '@lucide/svelte/icons/upload';
	import IconFile from '@lucide/svelte/icons/file';

	import type { MultiField } from '@internal/request/httpx';
	import CMInput from './CMInput.svelte';
	import { PathSelectService } from '@services';
	import { msgAlert } from '@comp/message/AlertNotify.svelte';

	type Multipart = MultiField[] | null;
	type Prop = {
		multipart: Multipart;
	};
	let { multipart: fields = $bindable() }: Prop = $props();
</script>

<table class="w-full table-fixed border border-base-content/10">
	<thead class="w-full">
		<tr class="[&>th]:border [&>th]:border-base-content/10">
			<th class="w-1/5">Name</th>
			<th class="grow">Value</th>
			<th class="max-w-1/5">Content-Type</th>
			<th class="p-0 w-8 has-[button:hover]:bg-info">
				<button
					title="add"
					class="w-full flex items-center justify-center cursor-pointer text-base-content"
					onclick={() => {
						fields ??= [];
						fields.push({
							name: '',
							value: '',
							fileName: '',
							filePath: '',
							contentType: '',
						});
					}}
				>
					<IconListPlus class="size-5" />
				</button>
			</th>
		</tr>
	</thead>
	<tbody class="w-full">
		{#each fields as field, i}
			<tr class="[&>td]:border [&>td]:border-base-content/10">
				<td><input class="w-full" placeholder={'name'} bind:value={field.name} /></td>
				<td>
					<div class="flex w-full">
						{#if field.filePath !== ''}
							<div class="grow">
								<div
									title={field.filePath}
									class="m-0.5 w-full flex items-center text-base-content/70 border border-base-content/80"
								>
									<IconFile class="size-4" />
									<p class="line-clamp-1">{field.fileName}</p>
								</div>
							</div>
							<button
								class="p-0 max-w-5 appearance-none cursor-pointer hover:text-error"
								onclick={() => {
									field.fileName = '';
									field.filePath = '';
								}}
							>
								<IconTrash class="size-5" />
							</button>
						{:else}
							<div class="grow">
								<CMInput bind:value={field.value} placeholder={'value'} />
							</div>
							<button
								class="p-0 max-w-5 appearance-none cursor-pointer text-base-content/80 hover:text-info"
								onclick={() => {
									PathSelectService.SelectFile()
										.then((v) => {
											// add file fields
											field.filePath = v;
											field.fileName = v.split(/[\\/]/).pop() ?? '';
											// clear other fields
											field.value = '';
										})
										.catch(msgAlert.error);
								}}
							>
								<IconUpload class="size-5" />
							</button>
						{/if}
					</div>
				</td>
				<td><input class="w-full" placeholder={'content-type'} bind:value={field.contentType} /></td
				>
				<td class="has-[button:hover]:bg-error">
					<button
						title="delete"
						class="w-full h-full flex items-center justify-center cursor-pointer text-base-content/80 hover:text-base-content"
						onclick={() => {
							fields?.splice(i, 1);
							if (fields && fields.length <= 0) fields = null;
						}}
					>
						<IconTrash class="size-5" />
					</button>
				</td>
			</tr>
		{/each}
	</tbody>
</table>
