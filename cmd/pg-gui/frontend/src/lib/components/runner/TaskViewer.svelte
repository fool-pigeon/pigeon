<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<script lang="ts">
	// Icons
	import IconCircleCheck from '@lucide/svelte/icons/circle-check';
	import IconCircleX from '@lucide/svelte/icons/circle-x';

	import { NodeMode, type TaskResHttp } from '@services';
	import type { TaskResWrapper } from './TaskRunner.svelte';
	import { ModeContains } from '@lib/utils/asserts';
	import HttpResViewer from '@comp/editor/HttpResViewer.svelte';
	import { nodeModeIcon } from '@comp/navigation/FileExplorer.svelte';
	import DynamicJoin, { NewPanel } from '@comp/base/DynamicJoin.svelte';

	type Prop = { wrappers: TaskResWrapper[] | undefined };
	let { wrappers }: Prop = $props();
	let current = $state<number>(-1);
</script>

<DynamicJoin panels={[NewPanel(ResList, { value: 30, unit: '%' }), NewPanel(ResDetail)]} />
{#snippet ResList()}
	<div class="min-w-4 w-full h-full min-h-0 p-2 flex flex-col gap-0">
		{#each wrappers as w, i}
			{@render DisplayTaskRes(w, () => (current = i))}
		{/each}
	</div>
{/snippet}
{#snippet ResDetail()}
	{#if current >= 0 && wrappers}
		{#key current}
			{@render TaskResDetail(wrappers[current])}
		{/key}
	{/if}
{/snippet}

{#snippet DisplayTaskRes(w: TaskResWrapper, on: () => void)}
	<div class="w-full text-sm">
		{const mode = w.mode}
		{const res = w.res}
		<button
			title={res.meta.path}
			class="w-full p-1 flex gap-0.5 items-center cursor-pointer hover:bg-base-content/15
			{res.success ? 'text-success' : 'text-error'}"
			onclick={on}
		>
			<div class="flex items-center *:shrink-0">
				<span>{@render nodeModeIcon(mode)}</span>
				<span class="*:size-3.5">
					{#if res.err || !res.success}
						<IconCircleX class="size-4" />
					{:else}
						<IconCircleCheck class="size-4" />
					{/if}
				</span>
			</div>
			<span class="min-w-0 text-sm truncate">{res.meta.path}</span>
		</button>
		<div class="pl-4 **:shrink-0">
			{#each res.tests as t}
				<div class="flex items-center text-xs {t.passed ? 'text-success' : 'text-error'}">
					<span class="*:size-3">
						{#if t.passed}
							<IconCircleCheck class="size-4" />
						{:else}
							<IconCircleX class="size-4" />
						{/if}
					</span>
					<span class="truncate">{`${t.passed ? 'pass' : 'fail'}: ${t.name}`}</span>
				</div>
			{/each}
		</div>
	</div>
{/snippet}

{#snippet TaskResDetail(w: TaskResWrapper)}
	{const mode = w.mode}
	{const res = w.res}
	<div class="w-full h-full min-h-0">
		<svelte:boundary>
			<!-- TODO: Viewer for other type -->
			{#snippet failed(err, reset)}
				{const resStr = JSON.stringify(res, null, 2)}
				<div class="w-full h-full min-h-0 overflow-auto text-error text-xs">
					<span class="break-all">{`lake attr:${err}`}</span>
					<div class="divider m-0"></div>
					<span class="whitespace-pre-wrap">{resStr}</span>
				</div>
			{/snippet}
			{#if ModeContains(mode, NodeMode.HttpReq)}
				{const resHttp = res as TaskResHttp}
				{const history = $derived([resHttp])}
				<HttpResViewer current={resHttp} {history} />
			{:else}
				{const resStr = JSON.stringify(res, null, 2)}
				<div class="w-full h-full min-h-0 overflow-auto">
					<div class="text-warning">not implements the viewer yet.</div>
					<div>mode:{mode}</div>
					<div>{resStr}</div>
				</div>
			{/if}
		</svelte:boundary>
	</div>
{/snippet}
