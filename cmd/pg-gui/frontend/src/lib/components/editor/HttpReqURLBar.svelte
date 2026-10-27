<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts">
	// Icons
	import IconSave from '@lucide/svelte/icons/save';

	import CMFormHttpURL from '@comp/form/CMFormHttpURL.svelte';
	import { MenuEvent } from '@cmd-gui/gui';
	import { Events } from '@wailsio/runtime';
	import { Method, ProtocolKind, type Request } from '@pg/protocol';

	let { req = $bindable(), isSending }: { req: Request; isSending: boolean } = $props();
	const METHOD_STYLE_CLASS = {
		[Method.$zero]: '',
		[Method.MethodGet]: 'text-req-get hover:not-[after]:bg-req-get/20',
		[Method.MethodPost]: 'text-req-post hover:bg-req-post/20',
		[Method.MethodPut]: 'text-req-put hover:bg-req-put/20',
		[Method.MethodDelete]: 'text-req-delete hover:bg-req-delete/20',
		[Method.MethodHead]: 'text-req-head hover:bg-req-head/20',
		[Method.MethodPatch]: 'text-req-patch hover:bg-req-patch/20',
		[Method.MethodOptions]: 'text-req-options hover:bg-req-options/20',
		[Method.MethodConnect]: 'text-req-connect hover:bg-req-connect/20',
		[Method.MethodTrace]: 'text-req-trace hover:bg-req-trace/20',
	};
</script>

<div class="flex gap-1">
	<div class="p-1 w-full flex gap-1 items-center border border-base-content/50">
		{#if (req.protocol === ProtocolKind.ProtocolREST || req.protocol === ProtocolKind.ProtocolGQL) && req.confREST}
			<!-- HTTP Method Selector -->
			<select
				class="m-0 p-0 pl-2 pr-2 optional:w-auto h-full select justify-center outline-none bg-none border-none shadow-none
				*:text-base-content *:bg-base-100 [&::picker(select)]:bg-base-100 *:hover:bg-base-content/20
				{METHOD_STYLE_CLASS[req.method]}"
				bind:value={req.method}
			>
				{#each Object.values(Method).filter((v) => v !== Method.$zero) as m (m)}
					<option value={m}>{m}</option>
				{/each}
			</select>
			<CMFormHttpURL bind:uri={req.url} bind:query={req.query} bind:path={req.confREST.path} />
		{:else}
			<CMFormHttpURL bind:uri={req.url} bind:query={req.query} />
		{/if}
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
