<script>
	import TestTubeDiagonal from "@lucide/svelte/icons/test-tube-diagonal";
	import Modal from "./Modal.svelte";

	let showModal = $state(false);
</script>

<Modal bind:showModal>
	{#snippet content()}
		Hello
	{/snippet}
</Modal>
<div class="absolute bottom-2 w-full">
	<div class="drawer">
		<input id="my-drawer-1" type="checkbox" class="drawer-toggle" />
		<div class="drawer-content">
			<div class="w-full flex justify-center">
				<label for="my-drawer-1" class="btn">
					<TestTubeDiagonal />
				</label>
			</div>
		</div>
		<div class="drawer-side">
			<label for="my-drawer-1" aria-label="close sidebar" class="drawer-overlay"
			></label>
			<ul class="menu bg-base-200 min-h-full w-80 p-4">
				<li>
					<button onclick={() => (showModal = true)}>ExampleClick</button>
				</li>
			</ul>
		</div>
	</div>
</div>
