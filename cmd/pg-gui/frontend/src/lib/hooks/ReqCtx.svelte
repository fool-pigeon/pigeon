<script lang="ts" module>
	import { msgAlert } from '@comp/message/AlertNotify.svelte';
	import { DataEvent } from '@lib/enum/events';
	import { assertPath } from '@lib/utils/asserts';
	import { ReqCtxService } from '@services';
	import { Events } from '@wailsio/runtime';
	import { onMount } from 'svelte';
	type ReqCtx = {
		root?: string;
		envFileName?: string;
	};
	let ctx = $state<ReqCtx>({});
	export const reqCtx = {
		root: () => ctx.root,
		envName: () => ctx.envFileName,
	};
</script>

<script lang="ts">
	let cancelled = $state(0);
	onMount(() => {
		const unregister = Events.On(DataEvent.UpdateReqCtx, (evt) => {
			const path = evt.data as string;
			if (ctx.root && assertPath.contains(ctx.root, path)) {
				return;
			}
			const cancel = ++cancelled;
			ReqCtxService.InitFromPath(path)
				.then((v) => {
					if (cancel !== cancelled) return;
					if (ctx.root && assertPath.contains(ctx.root, path)) {
						return;
					}
					ctx.root = v;
				})
				.catch(msgAlert.error);
		});
		return () => {
			unregister();
		};
	});
</script>
