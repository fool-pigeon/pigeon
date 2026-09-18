<script lang="ts">
	import { modal } from '@lib/components/base/Modal.svelte';
</script>

<button class="btn btn-accent btn-outline" onclick={() => modal.open(content)}>toggle-modal</button>
{#snippet content()}
	<h1>Hello,ExampleModal</h1>
{/snippet}
