<!-- @package -->
<script lang="ts">
	let { showModal = $bindable(false), content } = $props();
	let dialog = $state<HTMLDialogElement>();
</script>

<dialog
	bind:this={dialog}
	open={showModal}
	class="modal"
	onclose={() => (showModal = false)}
	onclick={(e) => {
		if (e.target === dialog) dialog.close();
	}}
>
	<div class="modal-box">
		{@render content?.()}
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">Cancel</button>
			</form>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>
