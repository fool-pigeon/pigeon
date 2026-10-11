<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconCircleCheck from '@lucide/svelte/icons/circle-check';
	import IconCircleX from '@lucide/svelte/icons/circle-x';

	import ScrollFlex from '@comp/base/ScrollFlex.svelte';
	import CMJson from '@comp/form/CMJson.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import { EditorState } from '@codemirror/state';
	import type { KArr } from '@type/form';
	import CollapseXs from '@comp/base/CollapseXs.svelte';
	import type { HttpMeta, HttpRes, TaskResHttp } from '@services';
	import type { TestRes } from '@internal/request';

	const ViewPart = {
		ResBody: 'Response',
		ResHeader: 'Header',
		Tests: 'Tests',
		History: 'History',
	} as const;
	type ViewPart = (typeof ViewPart)[keyof typeof ViewPart];
	const BodyView = {
		json: 'JSON',
		xml: 'XML',
		html: 'HTML',
		js: 'JavaScript',
		raw: 'RAW',
	} as const;
	type BodyView = (typeof BodyView)[keyof typeof BodyView];

	let { current, history }: { current: TaskResHttp; history: TaskResHttp[] } = $props();
	let viewPart = $state<ViewPart>(ViewPart.ResBody);
	let bodyView = $state<BodyView>(BodyView.json);

	function formatJSON(s: string): string {
		try {
			const formatted = JSON.stringify(JSON.parse(s), null, 2);
			return formatted;
		} catch (error) {
			return s;
		}
	}
</script>

<!-- TODO finish response viewer -->
<div class="w-full h-full min-h-0 flex flex-col">
	<!-- Top Bar -->
	<div
		class="w-full flex justify-between items-center
			border-b border-primary/50 whitespace-nowrap"
	>
		{@render TopBar(current)}
	</div>
	<!-- content viewer -->
	<div class="w-full grow min-h-0 overflow-hidden">
		{let resErr = $derived(current.err)}
		{#if resErr && viewPart !== 'History'}
			<div class="text-error">{JSON.stringify(resErr, null, 2)}</div>
		{:else if viewPart === 'Response'}
			{let body = $derived(current.data.response.body)}
			<!-- TODO:optional inject request URL 
					for example: body.replace('<head>', `<head><base href="${current.URL}">`)
				-->
			{@render BodyViewer(body, bodyView, 'allow-scripts allow-forms  allow-popups')}
		{:else if viewPart === 'Header'}
			{let header = $derived(current.data.response.header)}
			{@render Header(header)}
		{:else if viewPart === 'Tests'}
			{let tests = $derived(current.tests)}
			{@render Tests(tests)}
		{:else if viewPart === 'History'}
			<div class="p-1 pr-0 w-full h-full min-h-0 overflow-auto">
				{let h = $derived(history.toReversed())}
				{@render History(h)}
			</div>
		{/if}
	</div>
</div>

{#snippet TopBar(current: TaskResHttp)}
	<!-- view part select -->
	<ScrollFlex>
		{#snippet content()}
			{#each Object.values(ViewPart) as p (p)}
				<button
					class="p-1 cursor-pointer text-base-content/50
									{p === viewPart ? 'text-primary-content bg-primary' : 'hover:text-base-content'}"
					onclick={() => (viewPart = p)}>{p}</button
				>
			{/each}
		{/snippet}
	</ScrollFlex>
	<div class="shrink-0 overflow-hidden flex justify-end items-center gap-0.5 text-xs">
		<!-- Change response body preview type -->
		{#if viewPart === 'Response'}
			<select
				title="change preview"
				class="appearance-none pt-1 pb-1 border border-base-content/50
                        focus:outline-none text-center [text-align-last:center]
                        cursor-pointer hover:bg-primary hover:text-primary-content"
				bind:value={bodyView}
			>
				{#each Object.values(BodyView) as bv (bv)}
					<option class="appearance-none" value={bv}>{bv}</option>
				{/each}
			</select>
		{/if}
		<!-- current response view status -->
		<div
			class="grid grid-cols-[auto_auto] grid-rows-[auto_auto] gap-px bg-base-content/40
                    [&>div]:text-center [&>div]:bg-base-100 [&>div]:hover:bg-primary [&>div]:hover:text-primary-content"
		>
			{let success = $derived(current.success)}
			{let taskStatus = $derived(success ? 'Pass' : 'Fail')}
			<div title={taskStatus} class={success ? 'text-success' : 'text-error'}>{taskStatus}</div>
			<div title={current.data.spent}>{current.data.spent}</div>
			<div title={current.status}>{current.status}</div>
			<div title={current.data.size}>{current.data.size}</div>
		</div>
	</div>
{/snippet}

{#snippet BodyViewer(body: string, bodyView: BodyView, enableSandBox?: string)}
	{#if bodyView === BodyView.json}
		<CMJson value={formatJSON(body)} readonly />
	{:else if bodyView === BodyView.html}
		<iframe title="response-body" class="w-full h-full" sandbox={enableSandBox} srcdoc={body}>
		</iframe>
	{:else if bodyView === BodyView.xml}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{:else if bodyView === BodyView.raw}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{:else if bodyView === BodyView.js}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{/if}
{/snippet}
{#snippet SimpleBodyViewer(body: string)}
	<div class="w-full flex flex-col">
		{let simpleBodyView = $state(BodyView.json)}
		<select
			title="change preview"
			class="appearance-none w-24 select select-xs focus:outline-none"
			bind:value={simpleBodyView}
		>
			{#each Object.values(BodyView) as bv (bv)}
				<option class="appearance-none" value={bv}>{bv}</option>
			{/each}
		</select>
		<div class="border-t border-base-content/80 w-full h-[50vh]">
			{@render BodyViewer(body, simpleBodyView, '')}
		</div>
	</div>
{/snippet}

{#snippet Header(header: KArr)}
	<div class="w-full max-w-full h-full overflow-auto whitespace-break-spaces break-all">
		<table
			class="w-full table table-xs rounded-none table-fixed table-pin-rows border border-base-content/10"
		>
			<thead>
				<tr class="bg-base-300">
					<th class="max-w-1/4">Key</th>
					<th>Value</th>
				</tr>
			</thead>
			<tbody>
				{#if header}
					{#each Object.entries(header) as h (h[0])}
						<tr class="hover:bg-base-content/10">
							<td>{h[0]}</td>
							<td>{(h[1] ?? []).join(', ')}</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/snippet}

{#snippet Tests(tests: TestRes[] | null)}
	<div class="p-1 w-full max-w-full h-full overflow-auto flex flex-col gap-1">
		{#if !tests}
			<p>no tests</p>
		{:else}
			{#each tests as t}
				{#if t.passed}
					<div class="text-success flex items-center">
						<IconCircleCheck class="size-4 min-w-4" />
						<span class="break-all">Passed: {t.name}</span>
					</div>
				{:else}
					<div class="flex flex-col">
						<div class="text-error flex items-center break-all">
							<IconCircleX class="size-4 min-w-4" />
							<span>Failed: {t.name}</span>
						</div>
						<div
							class="ml-8 text-base-content/50 inline-flex flex-col gap-0
							text-xs font-mono *:whitespace-pre-wrap"
						>
							{#if t.err}
								{#if t.err.message && t.err.snippet}
									<pre>{`${t.err.message}${t.err.snippet}`}</pre>
								{:else}
									<pre>{`unkown err:please check the result of test must be bool`}</pre>
								{/if}
								<pre>{JSON.stringify(t.err, null, 2)}</pre>
							{/if}
						</div>
					</div>
				{/if}
			{/each}
		{/if}
	</div>
{/snippet}

{#snippet History(history: TaskResHttp[])}
	{let openIndex = $state<number>()}
	{#each history as h, i}
		<details
			class="collapse collapse-arrow bg-base-100 border border-base-content/10"
			name="history-accordion-det-1"
			ontoggle={(e) => {
				if (e.currentTarget.open) openIndex = i;
			}}
		>
			<summary class="collapse-title p-1 pr-7 flex">
				<div class="flex gap-1 min-w-0">
					<div class="badge badge-soft {h.success ? 'badge-success' : 'badge-error'}">
						{h.success ? 'Pass' : 'Fail'}
					</div>
					{#if h.err}
						<div title={h.reqInfo} class="truncate min-w-0">{h.reqInfo}</div>
					{:else}
						<div>{h.data.method}</div>
						<div title={h.data.URL} class="truncate min-w-0">{h.data.URL}</div>
					{/if}
				</div>
			</summary>
			{#if openIndex === i}
				<div class="collapse-content p-1 text-sm whitespace-pre-wrap break-all">
					{#if h.err}
						<div class="text-error">{JSON.stringify(h.err, null, 2)}</div>
					{:else}
						<div class="flex flex-col gap-1">
							<div class="pt-2 border-t border-dashed border-base-content/90">
								{`${h.data.method} ${h.data.URL}`}
							</div>
							<div class="w-full flex gap-2 border-b border-base-content/80">
								<div>{h.proto}</div>
								<div>{h.status}</div>
								<div>{h.data.spent}</div>
								<div>{h.data.size}</div>
							</div>
							{const tabs = ['Request', 'Response', 'Raw', 'Tests']}
							{let tab = $state(0)}
							<div class="w-full border-b border-base-content/20">
								<ScrollFlex>
									{#snippet content()}
										{#each tabs as t, i}
											<button
												class="appearance-none p-1 pl-4 pr-4 border-b-4
												cursor-pointer hover:text-primary-content hover:bg-primary
												{tab === i ? 'text-base-content  border-primary' : 'border-transparent text-base-content/70'}"
												onclick={() => (tab = i)}>{t}</button
											>
										{/each}
									{/snippet}
								</ScrollFlex>
							</div>
							{#if tab === 0}
								{@render HistoryHttpMeta(h.data.request)}
							{:else if tab === 1}
								{@render HistoryHttpMeta(h.data.response)}
							{:else if tab === 2}
								<div class="w-full h-full overflow-auto font-mono text-xs *:whitespace-pre-wrap">
									<pre class="border-base-content">{h.data.request.raw}</pre>
									<pre class="border-t border-primary">{h.data.response.raw}</pre>
								</div>
							{:else if tab === 3}
								{@render Tests(current.tests)}
							{/if}
						</div>
					{/if}
				</div>
			{/if}
		</details>
	{/each}
{/snippet}
{#snippet HistoryHttpMeta(meta: HttpMeta)}
	<CollapseXs title={'Header'} checked>
		{#snippet content()}
			{@render Header(meta.header)}
		{/snippet}
	</CollapseXs>
	<CollapseXs title={'Body'} checked>
		{#snippet content()}
			{#if meta.body === ''}
				<div class="text-base-content/50">No Body</div>
			{:else}
				{@render SimpleBodyViewer(meta.body)}
			{/if}
		{/snippet}
	</CollapseXs>
{/snippet}
