<svelte:options
	customElement={{
		tag: 'code-editor',
		shadow: 'none',
		props: {
			app: { reflect: true, type: 'Object' }
		}
	}}
/>

<!-- Courtesy of https://github.com/ala-garbaa-pro/svelte-5-monaco-editor-two-way-binding -->
<script lang="ts">
	import loader from '@monaco-editor/loader';
	import * as Monaco from 'monaco-editor/esm/vs/editor/editor.api';
	import { onDestroy, onMount } from 'svelte';
	import { risor } from './risor-language';

	let editor: Monaco.editor.IStandaloneCodeEditor;
	let monaco: typeof Monaco;
	let editorContainer: HTMLElement;

	// Define props with Svelte 5 syntax
	interface Props {
		value: string;
		language?: string;
	}
	let { value = $bindable(), language = ''}: Props = $props();

	function getTheme(): string {
		return localStorage.getItem('THEME_PREFERENCE_KEY') == 'dark' ? 'vs-dark' : 'vs'
	}

	onMount(() => {
		(async () => {
			const monacoEditor = await import('monaco-editor');
			loader.config({ monaco: monacoEditor.default });

			monaco = await loader.init();
			monaco.languages.register({ id: 'risor' });
			monaco.languages.setMonarchTokensProvider('risor', risor);
			// Your monaco instance is ready, let's display some code!
			editor = monaco.editor.create(editorContainer, {
				value,
				language,
				theme: getTheme(),
				automaticLayout: true,
				overviewRulerLanes: 0,
				overviewRulerBorder: false,
				wordWrap: 'on',
			});

			editor.onDidChangeModelContent((e) => {
				if (e.isFlush) {
					// true if setValue call
					//console.log('setValue call');
					/* editor.setValue(value); */
				} else {
					// console.log('user input');
					const updatedValue = editor?.getValue() ?? ' ';
					value = updatedValue;
				}
			});
		})();
	});

	$effect(() => {
		if (value) {
			if (editor) {
				// check if the editor is focused
				if (editor.hasWidgetFocus()) {
					// let the user edit with no interference
					// console.log(editor.getOptions())
				} else {
					if (editor?.getValue() ?? ' ' !== value) {
						editor?.setValue(value);
					}
				}
			}
		}
		if (value === '') {
			editor?.setValue(' ');
		}
	});

	onDestroy(() => {
		monaco?.editor.getModels().forEach((model) => model.dispose());
		editor?.dispose();
	});
</script>

<div class="code-editor" bind:this={editorContainer}></div>
