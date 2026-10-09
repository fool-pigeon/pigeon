<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts" module>
	import type { Snippet } from 'svelte';

	class ModalState {
		isOpen = $state(false);
		title = $state<string>();
		content = $state<Snippet>();

		open(snippet: Snippet, title?: string) {
			this.isOpen = true;
			this.content = snippet;
			this.title = title;
		}
		close() {
			this.isOpen = false;
			this.content = undefined;
			this.title = undefined;
		}
	}

	const mState = new ModalState();
	export const modal = {
		open: mState.open.bind(mState),
		close: mState.close.bind(mState),
	};
</script>

<script lang="ts">
	let dialog = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialog) return;
		if (mState.isOpen) {
			dialog.showModal();
		} else {
			dialog.close();
		}
	});
</script>

<dialog bind:this={dialog} class="modal" onclose={() => mState.close()}>
	<div class="modal-box w-auto max-w-lvw max-h-lvh min-h-0 p-0">
		<div class="flex pl-2 w-full h-10 bg-base-300 items-center justify-between">
			<span>
				{mState.title ?? ''}
			</span>
			<form method="dialog">
				<button class="btn btn-sm size-10 btn-ghost hover:btn-error">✕</button>
			</form>
		</div>
		{@render mState.content?.()}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
