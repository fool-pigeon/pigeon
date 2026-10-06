<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	// Icons
	import IconInfo from '@lucide/svelte/icons/info';
	import IconCircleCheck from '@lucide/svelte/icons/circle-check';
	import IconTriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import IconCircleX from '@lucide/svelte/icons/circle-x';

	import { NotificationState } from '../base/Notification.svelte';
	import { fly } from 'svelte/transition';
	import { circIn, linear } from 'svelte/easing';
	import { onMount } from 'svelte';
	import { Events } from '@wailsio/runtime';
	import { AlertEvent } from '@services/models';

	// position of alert notification on window
	export type hPosition = 'left' | 'center' | 'right';
	export type vPosition = 'top' | 'center' | 'bottom';
	class Pos {
		h = $state<hPosition>('center');
		v = $state<vPosition>('top');
	}
	const pos = new Pos();

	type Alert = {
		type: number;
		msg: unknown;
	};
	// alert state
	const aState = new NotificationState<Alert>();
	const Alerts = {
		Info: 0,
		Success: 1,
		Warn: 2,
		Error: 3,
	};
	function add(msg: unknown, type: number, duration: number) {
		aState.add(alertNotify, duration, { msg, type });
	}
	const d = 3000;
	export const msgAlert = {
		info: ((msg: unknown, duration = d) => add(msg, Alerts.Info, duration)).bind(aState),
		success: ((msg: unknown, duration = d) => add(msg, Alerts.Success, duration)).bind(aState),
		warn: ((msg: unknown, duration = d) => add(msg, Alerts.Warn, duration)).bind(aState),
		error: ((msg: unknown, duration = d) => add(msg, Alerts.Error, duration)).bind(aState),

		setHPos: ((h: hPosition) => (pos.h = h)).bind(pos),
		setVPos: ((v: vPosition) => (pos.v = v)).bind(pos),
	};

	const alertStyle = {
		[Alerts.Info]: 'alert-info',
		[Alerts.Success]: 'alert-success',
		[Alerts.Warn]: 'alert-warning',
		[Alerts.Error]: 'alert-error',
	};
	const alertIcon = {
		[Alerts.Info]: IconInfo,
		[Alerts.Success]: IconCircleCheck,
		[Alerts.Warn]: IconTriangleAlert,
		[Alerts.Error]: IconCircleX,
	};
</script>

<script lang="ts">
	onMount(() => {
		const unInfo = Events.On(AlertEvent.AlertInfo, (msg) => msgAlert.info(msg));
		const unSuccess = Events.On(AlertEvent.AlertSuccess, (msg) => msgAlert.success(msg));
		const unWarn = Events.On(AlertEvent.AlertWarn, (msg) => msgAlert.warn(msg));
		const unError = Events.On(AlertEvent.AlertError, (msg) => msgAlert.error(msg));
		return () => {
			unInfo();
			unSuccess();
			unWarn();
			unError();
		};
	});
</script>

<div
	class="absolute z-50 flex flex-col gap-2 m-2
    {pos.v === 'top' ? 'top-0' : pos.v === 'center' ? 'top-1/2 -translate-y-1/2' : 'bottom-0'}
    {pos.h === 'left' ? 'left-0' : pos.h === 'center' ? 'left-1/2 -translate-x-1/2' : 'right-0'}"
>
	{#each aState.items as n (n.id)}
		<div
			in:fly={{ y: pos.v === 'bottom' ? 36 : -36, duration: 200, easing: linear }}
			out:fly={{ y: pos.v === 'bottom' ? -36 : 36, duration: 200, easing: circIn }}
			onmouseenter={() => n.pause()}
			onmouseleave={() => n.resume()}
			role="alert"
		>
			{@render n.content?.(n.data!)}
		</div>
	{/each}
</div>
{#snippet alertNotify(alert: Alert)}
	{const Icon = alertIcon[alert.type]}
	<div class="p-2 alert alert-soft {alertStyle[alert.type]}">
		<div class="flex gap-1 items-center">
			<Icon class="size-4" />
			<span>{alert.msg}</span>
		</div>
	</div>
{/snippet}
