<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts">
	// Icons
	import IconSave from '@lucide/svelte/icons/save';

	import CMFromHttpURL from '@comp/form/CMFromHttpURL.svelte';
	import { Method, type ReqData } from '@internal/request/httpx';
	import { MenuEvent } from '@pg/gui';
	import { Events } from '@wailsio/runtime';

	let { req = $bindable(), isSending }: { req: ReqData; isSending: boolean } = $props();
	const METHOD_STYLE_CLASS = {
		[Method.MethodGet]: 'text-req-get hover:bg-req-get/50',
		[Method.MethodPost]: 'text-req-post hover:bg-req-post/50',
		[Method.MethodPut]: 'text-req-put hover:bg-req-put/50',
		[Method.MethodDelete]: 'text-req-delete hover:bg-req-delete/50',
		[Method.MethodHead]: 'text-req-head hover:bg-req-head/50',
		[Method.MethodPatch]: 'text-req-patch hover:bg-req-patch/50',
		[Method.MethodOptions]: 'text-req-options hover:bg-req-options/50',
		[Method.MethodConnect]: 'text-req-connect hover:bg-req-connect/50',
		[Method.MethodTrace]: 'text-req-trace hover:bg-req-trace/50',
	};
</script>

<div class="flex gap-1">
	<div class="p-1 w-full flex gap-1 items-center border border-base-content/50">
		<!-- HTTP Method Selector -->
		<select
			class="h-full appearance-none focus:outline-none text-center [text-align-last:center]
            {METHOD_STYLE_CLASS[req?.method ?? ''] ?? ''}"
			bind:value={req.method}
		>
			{#each Object.values(Method) as m (m)}
				<option value={m}>{m}</option>
			{/each}
		</select>
		<CMFromHttpURL bind:uri={req.url} bind:query={req.query} bind:path={req.path} />
		<!-- Save file btn -->
		<button
			title="Save"
			class="pl-1 pr-1 h-full opacity-50 text-base-content hover:opacity-100 hover:bg-base-content/10"
			onclick={() => Events.Emit(MenuEvent.Save)}
		>
			<IconSave class="size-5" />
		</button>
	</div>
	{#if isSending}
		<button
			class="h-full pt-0 pb-0 btn btn-error"
			onclick={() => Events.Emit(MenuEvent.CancelRequest)}
		>
			Cancel
		</button>
	{:else}
		<button
			class="h-full pt-0 pb-0 btn btn-primary"
			onclick={() => Events.Emit(MenuEvent.RunRequest)}
		>
			Send
		</button>
	{/if}
</div>
