<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import ScrollFlex from '@comp/base/ScrollFlex.svelte';
	import CMFormHttpMultipart from '@comp/form/CMFormHttpMultipart.svelte';
	import CMFormKArr from '@comp/form/CMFormKArr.svelte';
	import CMFormKV from '@comp/form/CMFormKV.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import FileSelector from '@comp/form/FileSelector.svelte';
	import { EditPart, BodyPart } from './HttpView.svelte';

	type Prop = { temp: ReqData | null; bodyPart: BodyPart };
	let { temp = $bindable(), bodyPart = $bindable() }: Prop = $props();
</script>

<div class="w-full h-full flex flex-col gap-1">
	{let reqEditPart = $state<EditPart>(EditPart.Path)}
	<!-- edit part select -->
	<div role="tablist" class="w-full border-b border-primary/50">
		<ScrollFlex>
			{#snippet content()}
				{#each Object.values(EditPart) as p (p)}
					<button
						class="p-1 cursor-pointer text-base-content/50
							{p === reqEditPart ? 'text-primary-content bg-primary' : 'hover:text-base-content'}"
						onclick={() => (reqEditPart = p)}>{p}</button
					>
				{/each}
			{/snippet}
		</ScrollFlex>
	</div>
	<!-- content -->
	<!-- TODO :binding form -->
	{#if !temp}
		<div class="text-error">temp value is null:{temp}</div>
	{:else if reqEditPart === EditPart.Path}
		<CMFormKV bind:value={temp.path} kReadOnly />
	{:else if reqEditPart === EditPart.Query}
		<CMFormKArr bind:value={temp.query} />
	{:else if reqEditPart === EditPart.Body}
		<select
			class="w-48 select select-sm select-primary focus:outline-none text-center [text-align-last:center]"
			bind:value={bodyPart}
		>
			{#each Object.values(BodyPart) as b (b)}
				<option value={b}>{b}</option>
			{/each}
		</select>
		<!-- BUG: No support of Form for GET or DELETE due to resty/v3 -->
		{#if bodyPart === BodyPart.Raw}
			<!-- TODO: different raw type such as json,xml,text -->
			<CMTextArea bind:value={temp.body.raw} />
		{:else if bodyPart === BodyPart.Multipart}
			<CMFormHttpMultipart bind:multipart={temp.body.multipart} />
		{:else if bodyPart === BodyPart.UrlEncoded}
			<CMFormKArr bind:value={temp.body.form} />
		{:else if bodyPart === BodyPart.Binary}
			<FileSelector bind:path={temp.body.binary} />
		{/if}
	{:else if reqEditPart === EditPart.Header}
		<CMFormKArr bind:value={temp.header} />
	{:else if reqEditPart === EditPart.Tests}
		<!-- TODO: pretty Tests form bindings -->
		<CMFormKV bind:value={temp.tests} kHeader="name" vHeader="script" />
	{:else if reqEditPart === EditPart.Desc}
		<CMTextArea bind:value={temp.desc} />
	{:else if reqEditPart === EditPart.PreReq}
		<CMTextArea bind:value={temp.preReq} />
	{/if}
</div>
