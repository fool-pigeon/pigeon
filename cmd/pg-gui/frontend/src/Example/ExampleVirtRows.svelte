<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

<script lang="ts">
	import VirtRowsWrapper from '@comp/base/VirtualList/VirtRowsWrapper.svelte';
	import { faker } from '@faker-js/faker';
	function randomNumber(min: number, max: number) {
		return faker.number.int({ min, max });
	}
	function listPart() {
		return new Array(1000).fill(true).map(() => faker.lorem.sentence(randomNumber(20, 70)));
	}
	let list = $state<string[]>([]);
</script>

<div class="w-sm h-full min-h-0 flex flex-col">
	<button class="btn btn-sm btn-soft btn-info" onclick={() => list.push(...listPart())}>
		add part
	</button>
	<div class="grow w-full">
		<VirtRowsWrapper {list}>
			{#snippet itemRender(item: string, i: number)}
				<div class={i % 2 === 0 ? 'bg-info/50' : 'bg-accent/50'}>
					<div>{i}</div>
					<div>{item}</div>
				</div>
			{/snippet}
		</VirtRowsWrapper>
	</div>
</div>
