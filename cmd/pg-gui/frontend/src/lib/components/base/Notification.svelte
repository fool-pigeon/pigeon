<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	import type { Snippet } from 'svelte';

	export class Notification<T> {
		content = $state<Snippet<[T]>>();
		data = $state<T>();
		manager;
		duration = 0;
		id = crypto.randomUUID();
		timeoutId: ReturnType<typeof setTimeout> | undefined;

		constructor(content: Snippet<[T]>, manager: NotificationState<T>, duration: number, data?: T) {
			this.content = content;
			this.data = data;
			this.duration = duration;
			this.manager = manager;
			this.resume();
		}
		/**
		 * remove from manager
		 */
		remove() {
			this.manager.remove(this.id);
		}
		/**
		 * pause the destory
		 */
		pause() {
			clearTimeout(this.timeoutId);
		}
		/**
		 * continue destory with duration
		 * @param duration
		 */
		resume() {
			this.timeoutId = setTimeout(() => this.remove(), this.duration);
		}
		setDuration(duration: number) {
			this.duration = duration;
		}
	}

	export class NotificationState<T> {
		items = $state<Notification<T>[]>([]);
		add(content: Snippet<[T]>, duration = 3000, data?: T) {
			this.items.push(new Notification(content, this, duration, data));
		}
		remove(id: string) {
			this.items = this.items.filter((n) => n.id !== id);
		}
	}
	export { defaultNotify };
</script>

{#snippet defaultNotify<T>(nfyState: NotificationState<T>)}
	{#each nfyState.items as n (n.id)}
		<div onmouseenter={() => n.pause()} onmouseleave={() => n.resume()} role="alert">
			{@render n.content?.(n.data!)}
		</div>
	{/each}
{/snippet}
