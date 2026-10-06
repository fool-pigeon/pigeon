<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import { modal } from '@lib/components/base/Modal.svelte';
</script>

<button class="btn btn-accent btn-outline" onclick={() => modal.open(content)}>toggle-modal</button>
{#snippet content()}
	<h1>Hello,ExampleModal</h1>
{/snippet}
