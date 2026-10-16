<!--
 Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->
<!-- @component
Theme Preview

Color palette, components, typography for daisyUI themes.
-->
<script lang="ts">
	// Color groups: primary tones / functional / base
	const groups = [
		['primary', 'secondary', 'accent', 'neutral'],
		['info', 'success', 'warning', 'error'],
		['base-100', 'base-200', 'base-300', 'base-content'],
	] as const;

	// Read live CSS variables; re-read on data-theme change
	let values = $state<Record<string, string>>({});
	const read = () => {
		const s = getComputedStyle(document.documentElement);
		values = Object.fromEntries(
			groups.flat().map((c) => [c, s.getPropertyValue(`--color-${c}`).trim()]),
		);
	};
	$effect(() => {
		read();
		const mo = new MutationObserver(read);
		mo.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
		return () => mo.disconnect();
	});

	const bg = (c: string) => `var(--color-${c})`;
	const fg = (c: string) =>
		c === 'base-content'
			? 'var(--color-base-100)'
			: c.startsWith('base-')
				? 'var(--color-base-content)'
				: `var(--color-${c}-content)`;

	// [classes, label] — class combo is kept as the title tooltip.
	// Label = default => style short name, otherwise the color name.
	type Item = [cls: string, label: string];

	const btnColored: Item[] = [
		['btn btn-primary', 'primary'],
		['btn btn-secondary', 'secondary'],
		['btn btn-info', 'info'],
		['btn btn-success', 'success'],
		['btn btn-warning', 'warning'],
		['btn btn-error', 'error'],
	];
	const btnSoftColored: Item[] = [
		['btn btn-soft btn-primary', 'primary'],
		['btn btn-soft btn-secondary', 'secondary'],
		['btn btn-soft btn-info', 'info'],
		['btn btn-soft btn-success', 'success'],
		['btn btn-soft btn-warning', 'warning'],
		['btn btn-soft btn-error', 'error'],
	];
	const btnOutlineColored: Item[] = [
		['btn btn-outline btn-primary', 'primary'],
		['btn btn-outline btn-secondary', 'secondary'],
		['btn btn-outline btn-info', 'info'],
		['btn btn-outline btn-success', 'success'],
		['btn btn-outline btn-warning', 'warning'],
		['btn btn-outline btn-error', 'error'],
	];

	const badgeColored: Item[] = [
		['badge badge-primary', 'primary'],
		['badge badge-secondary', 'secondary'],
		['badge badge-info', 'info'],
		['badge badge-success', 'success'],
		['badge badge-warning', 'warning'],
		['badge badge-error', 'error'],
	];
	const badgeSoftColored: Item[] = [
		['badge badge-soft badge-primary', 'primary'],
		['badge badge-soft badge-secondary', 'secondary'],
		['badge badge-soft badge-info', 'info'],
		['badge badge-soft badge-success', 'success'],
		['badge badge-soft badge-warning', 'warning'],
		['badge badge-soft badge-error', 'error'],
	];

	const checkboxColored: Item[] = [
		['checkbox checkbox-primary', 'primary'],
		['checkbox checkbox-secondary', 'secondary'],
		['checkbox checkbox-info', 'info'],
		['checkbox checkbox-success', 'success'],
		['checkbox checkbox-warning', 'warning'],
		['checkbox checkbox-error', 'error'],
	];

	const toggleColored: Item[] = [
		['toggle toggle-primary', 'primary'],
		['toggle toggle-secondary', 'secondary'],
		['toggle toggle-info', 'info'],
		['toggle toggle-success', 'success'],
		['toggle toggle-warning', 'warning'],
		['toggle toggle-error', 'error'],
	];

	const rows: { tag: string; items: Item[] }[] = [
		// btn: default / soft / outline
		{ tag: 'button', items: [['btn', 'btn'], ...btnColored] },
		{ tag: 'button', items: [['btn btn-soft', 'soft'], ...btnSoftColored] },
		{ tag: 'button', items: [['btn btn-outline', 'outline'], ...btnOutlineColored] },
		// badge: default / soft
		{ tag: 'badge', items: [['badge', 'badge'], ...badgeColored] },
		{ tag: 'badge', items: [['badge badge-soft', 'soft'], ...badgeSoftColored] },
		// checkbox / toggle: default
		{ tag: 'checkbox', items: [['checkbox', 'checkbox'], ...checkboxColored] },
		{ tag: 'toggle', items: [['toggle', 'toggle'], ...toggleColored] },
		// range: default + primary
		{
			tag: 'range',
			items: [
				['range', 'range'],
				['range range-primary', 'primary'],
			],
		},
	];
</script>

<div class="p-4 space-y-2 text-base-content">
	<section>
		<h2 class="text-lg font-bold mb-2">Colors</h2>
		<div>
			{#each groups as group}
				<div class="flex">
					{#each group as c}
						<div
							class="group flex-1 h-9 flex items-center justify-center cursor-default overflow-hidden"
							style="background:{bg(c)};color:{fg(c)}"
						>
							<span class="text-[9px] font-mono leading-none px-1 truncate">
								<span class="group-hover:hidden">{values[c] || '…'}</span>
								<span class="hidden group-hover:inline font-bold">{c}</span>
							</span>
						</div>
					{/each}
				</div>
			{/each}
		</div>
	</section>

	<section>
		<h2 class="text-lg font-bold mb-2">Components</h2>
		<div class="space-y-0.5">
			{#each rows as { tag, items }}
				<div class="flex flex-wrap items-center gap-0.5">
					{#each items as [cls, label]}
						{#if tag === 'button'}
							<button class="{cls} btn-xs" title={cls}>{label}</button>
						{:else if tag === 'badge'}
							<span class="{cls} badge-xs" title={cls}>{label}</span>
						{:else if tag === 'range'}
							<input type="range" class="{cls} range-xs" title={cls} value="50" />
						{:else if tag === 'toggle'}
							<input type="checkbox" class="{cls} toggle-xs" title={cls} checked />
						{:else}
							<input type="checkbox" class="{cls} checkbox-xs" title={cls} checked />
						{/if}
					{/each}
				</div>
			{/each}
		</div>
	</section>

	<section>
		<h2 class="text-lg font-bold mb-2">Typography</h2>
		<div class="prose prose-sm max-w-none">
			<h1>Heading 1</h1>
			<h2>Heading 2</h2>
			<h3>Heading 3</h3>
			<p>Body text with <a href="#top">link</a>, <strong>bold</strong>, <em>italic</em>.</p>
			<ul>
				<li>Item one</li>
				<li>Item two</li>
			</ul>
			<blockquote><p>Blockquote.</p></blockquote>
		</div>
	</section>
</div>
