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
	import IconChevronDown from '@lucide/svelte/icons/chevron-down';
	import IconChevronUp from '@lucide/svelte/icons/chevron-up';

	import ScrollFlex from '@comp/base/ScrollFlex.svelte';
	import CMFormHttpMultipart from '@comp/form/CMFormHttpMultipart.svelte';
	import CMFormKArr from '@comp/form/CMFormKArr.svelte';
	import CMFormKV from '@comp/form/CMFormKV.svelte';
	import CMTextArea from '@comp/form/CMTextArea.svelte';
	import FileSelector from '@comp/form/FileSelector.svelte';
	import { AuthKind, BodyKind, ProtocolKind, type Request } from '@pg/protocol';
	import PwdInput from '@comp/form/PwdInput.svelte';
	import TextInput from '@comp/form/TextInput.svelte';
	import CMJson from '@comp/form/CMJson.svelte';
	import SettingFieldWrapper, { FieldHeader } from '@comp/form/SettingFieldWrapper.svelte';

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
	let curPart = $state<Part>(Part.Payload);
</script>

<div class="w-full h-full flex flex-col gap-1">
	{#if !temp}
		<div class="text-error">temp value is null:{temp}</div>
	{:else}
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
						<div
							class="[&>button]:btn [&>button]:outline-none [&>button]:border-none
							[&>button]:p-1 [&>button]:pl-4 [&>button]:pr-4
							[&>button]:h-auto
							{p === curPart
								? '[&>button]:btn-primary'
								: '[&>button]:btn-ghost [&>button]:text-base-content/50 [&>button]:hover:text-base-content'}"
						>
							{#if p === Part.Payload}
								{#if temp.protocol === ProtocolKind.ProtocolGRPC || temp.protocol === ProtocolKind.ProtocolWS}
									<button onclick={() => (curPart = p)}>Message</button>
								{:else}
									{@render RESTBodyKindSelector()}
								{/if}
							{:else if p === Part.Auth}
								{@render AuthKindSelector()}
							{:else if p === Part.Header && temp.protocol === ProtocolKind.ProtocolGRPC}
								<button onclick={() => (curPart = p)}>Metadata</button>
							{:else}
								<button onclick={() => (curPart = p)}>{p}</button>
							{/if}
						</div>
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
			{:else if curPart === Part.Path}
				<CMFormKV bind:value={temp.confREST.path} />
			{:else if curPart === Part.Tests}
				<!-- TODO: pretty Tests form bindings -->
				<CMFormKV bind:value={temp.confREST.tests} kHeader="name" vHeader="script" />
			{:else if curPart === Part.Settings}
				<div class="w-full h-full overflow-auto space-y-6 text-sm">
					<SettingFieldWrapper>
						{#snippet content()}
							<div>
								{@render FieldHeader('timeout', "request's timeout string,use unit ns|ms|s...")}
								<input
									class="max-w-full w-auto input input-sm"
									bind:value={temp.timeout}
									type="text"
									placeholder="0ns"
								/>
							</div>
							<div>
								{@render FieldHeader('response body limit', 'if <=0,means no limit')}
								<input
									class="max-w-full w-auto input input-sm"
									bind:value={temp.confREST!.responseBodyLimit}
									type="text"
								/>
							</div>

							<span class="text-base-content/60 text-[1rem]">Retry</span>
							<div>
								{@render FieldHeader('count', 'the retry max count')}
								<input
									class="max-w-full w-auto input input-sm"
									bind:value={temp.confREST!.retry.count}
									type="text"
								/>
							</div>
							<div>
								{@render FieldHeader('wait time', 'time of waiting single retry')}
								<input
									class="max-w-full w-auto input input-sm"
									bind:value={temp.confREST!.retry.waitTime}
									type="text"
									placeholder="0ns"
								/>
							</div>
							<div>
								{@render FieldHeader('max wait time', 'max time of waiting single retry')}
								<input
									class="max-w-full w-auto input input-sm"
									bind:value={temp.confREST!.retry.maxWaitTime}
									type="text"
									placeholder="0ns"
								/>
							</div>
							<label class="cursor-pointer items-center">
								{@render FieldHeader(
									'allow non idempotent',
									'allow retry method of non idempotent,such as POST',
								)}
								<input
									class="toggle toggle-sm"
									bind:checked={temp.confREST!.retry.isAllowNonIdempotent}
									type="checkbox"
								/>
							</label>
							<label class="cursor-pointer items-center">
								<span class="text-base-content/70">unescape query</span>
								<input
									class="toggle toggle-sm"
									bind:checked={temp.confREST!.unescapeQuery}
									type="checkbox"
								/>
							</label>
							<label class="cursor-pointer items-center">
								<span class="text-base-content/70">unescape JSON</span>
								<input
									class="toggle toggle-sm"
									bind:checked={temp.confREST!.unescapeJSON}
									type="checkbox"
								/>
							</label>
						{/snippet}
					</SettingFieldWrapper>
				</div>
			{:else if curPart === Part.Payload}
				{let bkind = $derived(temp.payloadREST.kind)}
				{#if bkind === BodyKind.BodyURLEncoded}
					<CMFormKArr bind:value={temp.payloadREST.urlEncoded} />
				{:else if bkind === BodyKind.BodyMultipart}
					<CMFormHttpMultipart bind:multipart={temp.payloadREST.multipart} />
				{:else if bkind === BodyKind.BodyGQL}
					<CMTextArea bind:value={temp.payloadREST.raw} />
				{:else if bkind === BodyKind.BodyJSON}
					<CMJson bind:value={temp.payloadREST.raw} />
				{:else if bkind === BodyKind.BodyXML}
					<CMTextArea bind:value={temp.payloadREST.raw} />
				{:else if bkind === BodyKind.BodyRaw}
					<CMTextArea bind:value={temp.payloadREST.raw} />
				{:else if bkind === BodyKind.BodyBinary}
					<FileSelector bind:path={temp.payloadREST.binaryPath} />
				{:else if bkind === BodyKind.BodyNone}
					<div class="w-full h-full flex justify-center items-center">No Body</div>
				{/if}
			{/if}
		{/if}
	{/if}
</div>
{#snippet RESTBodyKindSelector()}
	{const bKindName = {
		[BodyKind.BodyURLEncoded]: 'URL Encoded',
		[BodyKind.BodyMultipart]: 'Multipart',
		[BodyKind.BodyGQL]: 'GraphQL',
		[BodyKind.BodyJSON]: 'JSON',
		[BodyKind.BodyXML]: 'XML',
		[BodyKind.BodyRaw]: 'Raw',
		[BodyKind.BodyBinary]: 'Binary File',
		[BodyKind.BodyNone]: 'No body',
	}}
	<button
		popovertarget="rest-body-kind-popover-1"
		style="anchor-name:--rest-body-kind-popover-1"
		onclick={(e) => {
			if (curPart !== Part.Payload) {
				e.preventDefault();
				curPart = Part.Payload;
				return;
			}
		}}
	>
		<div class="flex items-center">
			<span class="text-nowrap">{bKindName[temp!.payloadREST.kind]}</span>
			<IconChevronDown class="size-4" />
		</div>
	</button>
	<ul
		class="dropdown menu max-h-[60vh] bg-base-100 overflow-auto"
		onwheel={(e) => e.stopPropagation()}
		popover
		id="rest-body-kind-popover-1"
		style="position-anchor:--rest-body-kind-popover-1"
	>
		{#snippet btn(k: BodyKind)}
			<li>
				<button
					popovertarget="rest-body-kind-popover-1"
					popovertargetaction="hide"
					onclick={() => (temp!.payloadREST.kind = k)}
				>
					{bKindName[k]}
				</button>
			</li>
		{/snippet}
		{#snippet divider(s: string)}
			<div class="divider divider-start m-0 p-0">{s}</div>
		{/snippet}
		{@render divider('Form')}
		{@render btn(BodyKind.BodyURLEncoded)}
		{@render btn(BodyKind.BodyMultipart)}
		{@render divider('Text')}
		{@render btn(BodyKind.BodyGQL)}
		{@render btn(BodyKind.BodyJSON)}
		{@render btn(BodyKind.BodyXML)}
		{@render btn(BodyKind.BodyRaw)}
		{@render divider('Other')}
		{@render btn(BodyKind.BodyBinary)}
		{@render btn(BodyKind.BodyNone)}
	</ul>
{/snippet}
{#snippet AuthKindSelector()}
	{const authKindName = {
		[AuthKind.AuthNone]: 'No Auth',
		[AuthKind.AuthInherited]: 'Inherited',
		[AuthKind.AuthBasic]: 'Basic',
		[AuthKind.AuthDigest]: 'Digest',
		[AuthKind.AuthToken]: 'Token',
	} as const}
	<button
		popovertarget="req-auth-kind-popover-1"
		style="anchor-name:--req-auth-kind-popover-1"
		onclick={(e) => {
			if (curPart !== Part.Auth) {
				e.preventDefault();
				curPart = Part.Auth;
				return;
			}
		}}
	>
		<div class="flex items-center">
			<span class="text-nowrap">{authKindName[temp!.auth.kind]}</span>
			<IconChevronDown class="size-4" />
		</div>
	</button>

	<ul
		class="dropdown menu max-h-[60vh] bg-base-100 overflow-auto"
		onwheel={(e) => e.stopPropagation()}
		popover
		id="req-auth-kind-popover-1"
		style="position-anchor:--req-auth-kind-popover-1"
	>
		{#snippet btn(k: AuthKind)}
			<li>
				<button
					popovertarget="req-auth-kind-popover-1"
					popovertargetaction="hide"
					onclick={() => (temp!.auth.kind = k)}
				>
					{authKindName[k]}
				</button>
			</li>
		{/snippet}
		{@render btn(AuthKind.AuthBasic)}
		{@render btn(AuthKind.AuthDigest)}
		{@render btn(AuthKind.AuthToken)}
		<div class="divider m-0 p-0 h-1"></div>
		{@render btn(AuthKind.AuthInherited)}
		{@render btn(AuthKind.AuthNone)}
	</ul>
{/snippet}
