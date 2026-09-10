<script lang="ts">
	import { WML } from "@wailsio/runtime";
	import { onMount } from "svelte";

	onMount(() => {
		// Event Example:
		// Events.On('time', (v: any) => {
		//   const compact = (full.match(/\d{1,2}:\d{2}:\d{2}/) || [full])[0];
		//   time = window.matchMedia('(max-width: 640px)').matches ? compact : full;
		// });
		// Wire up data-wml-openURL links (logos + footer "Docs" link).
		WML.Reload();
	});
</script>

<div class="w-lvw h-lvh bg-success">
	<h1>Hello World</h1>
</div>