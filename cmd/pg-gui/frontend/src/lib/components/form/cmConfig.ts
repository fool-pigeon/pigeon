// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { EditorView } from '@codemirror/view';
import { tags as t } from '@lezer/highlight';

// 1. Editor UI themes
export const daisyColorTheme = EditorView.theme({
    // Editor outermost container
    '&': {
        backgroundColor: 'var(--color-base-100)',
        color: 'var(--color-base-content)',
        // fontSize: '0.875rem', // 14px，Consistent with daisyUI default text size
    },
    // rollable container
    '.cm-scroller': {
        // fontFamily: 'ui-monospace, monospace',
        lineHeight: '1.6',
    },
    // 内容区
    '.cm-content': {
        caretColor: 'var(--color-info)',
        // padding: '0.5rem 0',
    },
    // cursor
    '.cm-cursor, .cm-dropCursor': {
        borderLeftColor: 'var(--color-info)',
        borderLeftWidth: '1px',
    },
    // Selection (focus and out of focus)
    '&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground,.cm-selectionBackground, .cm-content ::selection': {
        // backgroundColor: 'var(--color-primary)',
        backgroundColor: 'color-mix(in oklch, var(--color-primary) 70%, transparent)',
        color: "var(--color-primary-content)"
    },
    // Current row highlighted
    '.cm-activeLine': {
        backgroundColor: 'transparent',
        // backgroundColor: 'var(--color-base-200)',
        // backgroundColor: 'color-mix(in oklch, var(--color-primary) 25%, transparent)',
    },
    // 行号栏
    '.cm-gutters': {
        backgroundColor: 'var(--color-base-200)',
        // color: 'var(--color-base-content)', 
        color: 'color-mix(in oklch, var(--color-base-content) 50%, transparent)',
        border: 'none',
        borderRight: '1px solid var(--color-base-300)',
    },
    // current line number
    '.cm-activeLineGutter': {
        backgroundColor: 'var(--color-base-300)',
        color: 'var(--color-base-content)',
    },
    // matching parentheses
    '.cm-matchingBracket': {
        // backgroundColor: 'var(--color-primary)',
        backgroundColor: 'color-mix(in oklch, var(--color-primary) 30%, transparent)',
        outline: '1px solid var(--color-primary)',
    },
    // mismatched parenthesis
    '.cm-nonmatchingBracket': {
        // backgroundColor: 'var(--color-error)',
        backgroundColor: 'color-mix(in oklch, var(--color-error) 30%, transparent)',
        outline: '1px solid var(--color-error)',
    },
    // search matching
    '.cm-searchMatch': {
        // backgroundColor: 'var(--color-warning)',
        backgroundColor: 'color-mix(in oklch, var(--color-warning) 30%, transparent)',
        outline: '1px solid var(--color-warning)',
    },
    // Current search matches
    '.cm-searchMatch.cm-searchMatch-selected': {
        // backgroundColor: 'var(--color-warning)',
        backgroundColor: 'color-mix(in oklch, var(--color-warning) 50%, transparent)',
    },
    // select matching
    '.cm-selectionMatch': {
        // backgroundColor: 'var(--color-accent)',
        backgroundColor: 'color-mix(in oklch, var(--color-accent) 25%, transparent)',
    },
    // collapsed placeholder
    '.cm-foldPlaceholder': {
        backgroundColor: 'var(--color-base-300)',
        color: 'var(--color-base-content)',
        border: 'none',
    },
    // Tooltips (e.g. autocomplete)
    '.cm-tooltip': {
        backgroundColor: 'var(--color-base-100)',
        border: '1px solid var(--color-base-300)',
        color: 'var(--color-base-content)',
    },
    '.cm-tooltip-autocomplete ul li[aria-selected]': {
        backgroundColor: 'var(--color-primary)',
        color: 'var(--color-primary-content)',
    },
}, { dark: false }); // Set to false, so that the &dark rule does not automatically take effect, completely controlled by the daisyUI variable

// 2. syntax highlight style
const daisyHighlightStyle = HighlightStyle.define([
    { tag: t.comment, color: 'var(--color-base-content)', opacity: 0.6 },
    { tag: t.lineComment, color: 'var(--color-base-content)', opacity: 0.6 },
    { tag: t.blockComment, color: 'var(--color-base-content)', opacity: 0.6 },
    { tag: t.keyword, color: 'var(--color-primary)', fontWeight: 'bold' },
    { tag: t.controlKeyword, color: 'var(--color-primary)', fontWeight: 'bold' },
    { tag: t.operatorKeyword, color: 'var(--color-primary)' },
    { tag: t.string, color: 'var(--color-success)' },
    { tag: t.special(t.string), color: 'var(--color-success)' },
    { tag: t.number, color: 'var(--color-accent)' },
    { tag: t.bool, color: 'var(--color-accent)' },
    { tag: t.null, color: 'var(--color-accent)' },
    { tag: t.variableName, color: 'var(--color-base-content)' },
    { tag: t.definition(t.variableName), color: 'var(--color-secondary)' },
    { tag: t.function(t.variableName), color: 'var(--color-secondary)' },
    { tag: t.propertyName, color: 'var(--color-info)' },
    { tag: t.className, color: 'var(--color-secondary)' },
    { tag: t.typeName, color: 'var(--color-secondary)' },
    { tag: t.tagName, color: 'var(--color-primary)' },
    { tag: t.attributeName, color: 'var(--color-info)' },
    { tag: t.operator, color: 'var(--color-base-content)', opacity: 0.8 },
    { tag: t.punctuation, color: 'var(--color-base-content)', opacity: 0.7 },
    { tag: t.bracket, color: 'var(--color-base-content)', opacity: 0.7 },
    { tag: t.brace, color: 'var(--color-base-content)', opacity: 0.7 },
    { tag: t.paren, color: 'var(--color-base-content)', opacity: 0.7 },
    { tag: t.meta, color: 'var(--color-base-content)', opacity: 0.6 },
    { tag: t.heading, color: 'var(--color-primary)', fontWeight: 'bold' },
    { tag: t.link, color: 'var(--color-info)', textDecoration: 'underline' },
    { tag: t.emphasis, fontStyle: 'italic' },
    { tag: t.strong, fontWeight: 'bold' },
    { tag: t.invalid, color: 'var(--color-error)' },
]);

// 3. Combination for Extension
const daisyThemeExtension = [
    daisyColorTheme,
    syntaxHighlighting(daisyHighlightStyle),
];

export { daisyThemeExtension };