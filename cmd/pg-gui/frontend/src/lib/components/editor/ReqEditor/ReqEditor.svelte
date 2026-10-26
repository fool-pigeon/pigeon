<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	// Icons
	import IconUser from '@lucide/svelte/icons/user';
	import IconShield from '@lucide/svelte/icons/shield';
	import IconTag from '@lucide/svelte/icons/tag';

	import ScrollFlex from '@comp/base/ScrollFlex.svelte';
	import CMFormHttpMultipart from '@comp/form/CMFormHttpMultipart.svelte';
	import CMFormKArr from '@comp/form/CMFormKArr.svelte';
	import CMFormKV from '@comp/form/CMFormKV.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import FileSelector from '@comp/form/FileSelector.svelte';
	import { AuthKind, BodyKind, ProtocolKind, type Request } from '@pg/protocol';
	import PwdInput from '@comp/form/PwdInput.svelte';
	import TextInput from '@comp/form/TextInput.svelte';

	const Part = {
		Desc: 'Desc',
		Auth: 'Auth',
		Path: 'Path',
		Query: 'Query',
		Header: 'Header',
		Settings: 'Settings',
		Tests: 'Tests',
		Payload: 'Payload',
		PreReq: 'PreRequest',
		// TODO: for more,auth,retry,other setting
	} as const;
	type Part = (typeof Part)[keyof typeof Part];
	const RESTParts = [
		Part.Payload,
		Part.Path,
		Part.Query,
		Part.Header,
		Part.Auth,
		Part.Tests,
		Part.PreReq,
		Part.Settings,
		Part.Desc,
	] as const;
	const GRPCParts = [Part.Payload, Part.Header, Part.Auth, Part.Settings, Part.Desc] as const;
	const WSParts = [
		Part.Payload,
		Part.Query,
		Part.Header,
		Part.Auth,
		Part.Settings,
		Part.Desc,
	] as const;

	type Prop = { temp: Request | null };
	let { temp = $bindable() }: Prop = $props();
</script>

<div class="w-full h-full flex flex-col gap-1">
	{#if !temp}
		<div class="text-error">temp value is null:{temp}</div>
	{:else}
		{let curPart = $state<Part>(Part.Payload)}
		{let parts = $derived.by(() => {
			switch (temp.protocol) {
				case ProtocolKind.ProtocolGRPC:
					return GRPCParts;
				case ProtocolKind.ProtocolWS:
					return WSParts;
				default:
					return RESTParts;
			}
		})}
		<!-- edit part select -->
		<div role="tablist" class="w-full border-b border-primary/50">
			<ScrollFlex>
				{#snippet content()}
					{#each parts as p (p)}
						<button
							class="p-1 cursor-pointer text-base-content/50
							{p === curPart ? 'text-primary-content bg-primary' : 'hover:text-base-content'}"
							onclick={() => (curPart = p)}
						>
							{#if p === 'Payload'}
								{#if temp.protocol === ProtocolKind.ProtocolGRPC || temp.protocol === ProtocolKind.ProtocolWS}
									Message
								{:else}
									<!-- TODO selector for body kind -->
									Body
								{/if}
							{:else if p === 'Auth'}
								{const authKindName = {
									[AuthKind.AuthNone]: 'No Auth',
									[AuthKind.AuthInherited]: 'Inherited',
									[AuthKind.AuthBasic]: 'Basic',
									[AuthKind.AuthDigest]: 'Digest',
									[AuthKind.AuthToken]: 'Token',
								} as const}
								<select bind:value={temp.auth.kind} class="select">
									{#each Object.values(AuthKind) as ak}
										<option value={ak}>{authKindName[ak]}</option>
									{/each}
								</select>
							{:else if p === 'Header' && temp.protocol === ProtocolKind.ProtocolGRPC}
								Metadata
							{:else}
								{p}
							{/if}
						</button>
					{/each}
				{/snippet}
			</ScrollFlex>
		</div>
		<!-- content -->
		<!-- TODO :binding form -->

		{#if curPart === Part.Query}
			<CMFormKArr bind:value={temp.query} />
		{:else if curPart === Part.Desc}
			<CMTextArea bind:value={temp.desc} />
		{:else if curPart === Part.Header}
			<CMFormKArr bind:value={temp.header} />
		{:else if curPart === Part.PreReq}
			<CMTextArea bind:value={temp.preReqExpr} />
		{:else if curPart === Part.Auth}
			{let kind = $derived(temp.auth.kind)}
			{#if kind === AuthKind.AuthNone}
				<div class="w-full h-full flex items-center justify-center">No Authenication</div>
			{:else if kind === AuthKind.AuthInherited}
				<div class="w-full h-full flex items-center justify-center">Inherited</div>
			{:else if kind === AuthKind.AuthBasic || kind === AuthKind.AuthDigest}
				<TextInput bind:val={temp.auth.creds.username} placeholder="username">
					{#snippet iconSnippet()}
						<IconUser class="shrink-0 size-4" />
					{/snippet}
				</TextInput>
				<PwdInput bind:pwd={temp.auth.creds.password} />
			{:else if kind === AuthKind.AuthToken}
				<PwdInput bind:pwd={temp.auth.token.value} placeholder="token" />
				<TextInput bind:val={temp.auth.token.scheme} placeholder="scheme">
					{#snippet iconSnippet()}
						<IconShield class="shrink-0 size-4" />
					{/snippet}
				</TextInput>
				<TextInput bind:val={temp.auth.token.header} placeholder="header">
					{#snippet iconSnippet()}
						<IconTag class="shrink-0 size-4" />
					{/snippet}
				</TextInput>
			{/if}
		{:else if temp.protocol === ProtocolKind.ProtocolREST || temp.protocol === ProtocolKind.ProtocolGQL}
			{#if !temp.confREST}
				<div class="text-error">nil</div>
			{:else if curPart === Part.Tests}
				<!-- TODO: pretty Tests form bindings -->
				<CMFormKV bind:value={temp.confREST.tests} kHeader="name" vHeader="script" />
			{:else if curPart === Part.Settings}
				<!-- TODO Settings page -->
				<div></div>
			{:else if curPart === Part.Payload}
				<select
					class="w-48 select select-sm select-primary focus:outline-none text-center [text-align-last:center]"
					bind:value={temp.payloadREST.kind}
				>
					{#each Object.values(BodyKind) as b (b)}
						<option value={b}>{b}</option>
					{/each}
				</select>
			{/if}
		{/if}
	{/if}
</div>
