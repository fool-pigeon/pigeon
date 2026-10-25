<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { DataEvent } from '@lib/enum/events';
	import { assertPath } from '@lib/utils/asserts';
	import { ReqCtxService, type FileMeta } from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	type ReqCtx = {
		root?: string;
		confFile?: FileMeta;
	};
	export const reqCtx = $state<ReqCtx>({});
</script>

<script lang="ts">
	let cancelled = $state(0);
	onMount(() => {
		const unregister = Events.On(DataEvent.UpdateReqCtx, (evt) => {
			const path = evt.data as string;
			if (reqCtx.root && assertPath.contains(reqCtx.root, path)) {
				return;
			}
			const cancel = ++cancelled;
			ReqCtxService.SetRootFromPath(path)
				.then((v) => {
					if (cancel !== cancelled) return;
					if (reqCtx.root && assertPath.contains(reqCtx.root, path)) {
						return;
					}
					reqCtx.root = v;
				})
				.catch(msgAlert.error);
		});
		return () => {
			unregister();
		};
	});
	$effect(() => {
		const confFile = reqCtx.confFile;
		if (confFile) {
			ReqCtxService.SetRootConfMeta(confFile)
				.then(() => msgAlert.success(`current conf ${confFile.name}`))
				.catch(msgAlert.error);
		}
	});
</script>
