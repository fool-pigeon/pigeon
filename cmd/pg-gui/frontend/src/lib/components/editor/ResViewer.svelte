<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconCircleCheck from '@lucide/svelte/icons/circle-check';
	import IconCircleX from '@lucide/svelte/icons/circle-x';
	import IconMoveDownLeft from '@lucide/svelte/icons/move-down-left';
	import IconGlobe from '@lucide/svelte/icons/globe';
	import IconCircleQuestionMark from '@lucide/svelte/icons/circle-question-mark';

	import ScrollFlex from '@comp/base/ScrollFlex.svelte';
	import CMJson from '@comp/form/CMJson.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import { EditorState } from '@codemirror/state';
	import type { KArr } from '@type/form';
	import { TaskService, type TaskResView } from '@services';
	import { ProtocolKind, TraceKind, type TestRes, type TraceView } from '@pg/protocol';
	import type { ULID } from '@bindings/github.com/oklog/ulid/v2';
	import { onMount } from 'svelte';
	import { TaskSSReg } from '@comp/context/TaskSSManager.svelte';
	import { SvelteMap } from 'svelte/reactivity';

	const ViewPart = {
		Trace: 'Trace',
		ResBody: 'Response',
		ReqBody: 'Request',
		Header: 'Header',
		Tests: 'Tests',
	} as const;
	type ViewPart = (typeof ViewPart)[keyof typeof ViewPart];
	const BodyView = {
		JSON: 'JSON',
		XML: 'XML',
		HTML: 'HTML',
		JS: 'JavaScript',
		RAW: 'RAW',
	} as const;
	type BodyView = (typeof BodyView)[keyof typeof BodyView];

	type Prop = {
		res: TaskResView | null;
		resID?: ULID;
		protocol?: ProtocolKind;
	};
	let { res, resID, protocol }: Prop = $props();
	// request's protocol
	let rp = $derived(protocol ?? res?.request.protocol);
	// result's id for viewing trace
	let rid = $derived(resID ?? res?.id);
	let viewPart = $state<ViewPart>(ViewPart.ResBody);
	let bodyView = $state<BodyView>(BodyView.JSON);

	let traceViews = $state<TraceView[]>([]);
	let traceViewsFiltered = $derived(traceViews.filter(Boolean));

	function formatJSON(s: string): string {
		try {
			const formatted = JSON.stringify(JSON.parse(s), null, 2);
			return formatted;
		} catch (error) {
			return s;
		}
	}
	$effect(() => {
		if (rid) {
			// clear
			traceViews = [];

			const unTrace = TaskSSReg.onTrace((t) => {
				if (t.resID === rid && !traceViews[t.id]) {
					traceViews[t.id] = t;
				}
			});
			TaskService.LoadTraceViewList(rid);
			return () => unTrace();
		}
	});
	const TraceKindIcon = {
		[TraceKind.TraceReceived]: IconMoveDownLeft,
		[TraceKind.TraceNetwork]: IconGlobe,
		[TraceKind.$zero]: IconCircleQuestionMark,
	};
</script>

<!-- TODO finish response viewer -->
<div class="w-full h-full min-h-0 flex flex-col">
	<!-- Top Bar -->
	<div
		class="w-full flex justify-between items-center
			border-b border-primary/50 whitespace-nowrap"
	>
		{@render TopBar(res)}
	</div>
	{#if !res}
		<div class="text-warning">result is nil</div>
	{:else}
		<!-- content viewer -->
		<div class="w-full grow min-h-0 overflow-hidden">
			{#if rp === ProtocolKind.ProtocolREST || rp === ProtocolKind.ProtocolGQL}
				{#if viewPart === ViewPart.ResBody}
					{let body = $derived(res.httpResponse?.body ?? '')}
					<!-- TODO:optional inject request URL 
				for example: body.replace('<head>', `<head><base href="${current.URL}">`)
				-->
					{@render BodyViewer(body, bodyView, 'allow-scripts allow-forms  allow-popups')}
				{:else if viewPart === ViewPart.ReqBody}
					{let body = $derived(res.httpRequest?.body ?? '')}
					{#if body === ''}
						<div class="w-full h-full flex justify-center items-center text-base-content/70">
							No body
						</div>
					{:else}
						{@render BodyViewer(body, bodyView, 'allow-scripts allow-forms  allow-popups')}
					{/if}
				{/if}
			{:else if viewPart === ViewPart.Trace}
				<div class="w-full h-full overflow-auto flex flex-col">
					{#each traceViewsFiltered as t}
						{@const Icon = TraceKindIcon[t.kind]}
						<details class="collapse collapse-arrow" name="response-viewer-trace-accordion-det">
							<summary class="collapse-title flex">
								<Icon class="size-4" />
								<span>{t.msg}</span>
							</summary>
							<div class="collapse-title">
								<!-- TODO pretty trace view display -->
								{JSON.stringify(t.meta, null, 2)}
							</div>
						</details>
					{/each}
				</div>
			{:else if viewPart === ViewPart.Header}
				{let header = $derived(res.httpResponse?.header ?? null)}
				{@render Header(header)}
			{:else if viewPart === ViewPart.Tests}
				{@render Tests(res.tests)}
			{/if}
		</div>
	{/if}
</div>

{#snippet TopBar(res: TaskResView | null)}
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
				class="select select-xs justify-center w-16 p-1 outline-none bg-none"
				bind:value={bodyView}
			>
				{#each Object.values(BodyView) as bv (bv)}
					<option value={bv}>{bv}</option>
				{/each}
			</select>
		{/if}
		<!-- current response view status -->
		{#if res}
			<div
				class="grid grid-cols-[auto_auto] grid-rows-[auto_auto] gap-px bg-base-content/40
				[&>div]:text-center [&>div]:bg-base-100 [&>div]:hover:bg-primary [&>div]:hover:text-primary-content"
			>
				{let success = $derived(res.success)}
				{let taskStatus = $derived(success ? 'Pass' : 'Fail')}
				<div title={taskStatus} class={success ? 'text-success' : 'text-error'}>{taskStatus}</div>
				<div title={res.spent}>{res.spent}</div>
				<div title={res.status}>{res.status}</div>
				<div title={res.contentSize}>{res.contentSize}</div>
			</div>
		{/if}
	</div>
{/snippet}

{#snippet BodyViewer(body: string, bodyView: BodyView, enableSandBox?: string)}
	{#if bodyView === BodyView.JSON}
		<CMJson value={formatJSON(body)} readonly />
	{:else if bodyView === BodyView.HTML}
		<iframe title="response-body" class="w-full h-full" sandbox={enableSandBox} srcdoc={body}>
		</iframe>
	{:else if bodyView === BodyView.XML}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{:else if bodyView === BodyView.RAW}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{:else if bodyView === BodyView.JS}
		<CMTextArea value={body} extensions={[EditorState.readOnly.of(true)]} />
	{/if}
{/snippet}
{#snippet SimpleBodyViewer(body: string)}
	<div class="w-full flex flex-col">
		{let simpleBodyView = $state(BodyView.JSON)}
		<select
			title="change preview"
			class="select select-xs optional:w-auto focus:outline-none"
			bind:value={simpleBodyView}
		>
			{#each Object.values(BodyView) as bv (bv)}
				<option value={bv}>{bv}</option>
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
								{#if t.err}
									<pre>{`${t.err}`}</pre>
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
