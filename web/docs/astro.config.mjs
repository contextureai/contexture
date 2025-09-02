	// @ts-check
	import { defineConfig } from 'astro/config';
	import starlight from '@astrojs/starlight';
	import starlightContextualMenu from 'starlight-contextual-menu';
	import starlightLlmsTxt from 'starlight-llms-txt';
	import mermaid from 'astro-mermaid';

	import tailwindcss from '@tailwindcss/vite';

	// https://astro.build/config
	export default defineConfig({
	site: 'https://docs.contexture.sh',

	integrations: [
		mermaid({
			theme: 'dark',
			autoTheme: true
		}),
		starlight({
			title: 'Contexture',
			description: 'Contexture is a CLI tool for managing AI coding assistant context.',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/contextureai/contexture' }],
			plugins: [
				starlightContextualMenu({
					actions: ["copy", "view", "chatgpt", "claude"]
				}),
				starlightLlmsTxt(),
			],
			head: [
				{
					tag: 'meta',
					attrs: {
						'color-scheme': 'light dark',
					}
				}
			],
			customCss: [
				'./src/styles/global.css',
				'@fontsource/geist-sans/400.css',
				'@fontsource-variable/geist-mono/wght.css'
			],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Quick Start', slug: 'getting-started/quick-start' },
					],
				},
				{
					label: 'Core Concepts',
					items: [
						{ label: 'Overview', slug: 'core-concepts/overview' },
						{ label: 'Rules', slug: 'core-concepts/rules' },
						{ label: 'Projects', slug: 'core-concepts/projects' },
						{ label: 'Variables', slug: 'core-concepts/variables' },
						{ label: 'Formats', slug: 'core-concepts/formats' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ 
							label: 'Commands',
							items: [
								{ label: 'init', slug: 'reference/commands/init' },
								{ label: 'build', slug: 'reference/commands/build' },
								{ label: 'config', slug: 'reference/commands/config' },
								{ label: 'rules add', slug: 'reference/commands/rules-add' },
								{ label: 'rules list', slug: 'reference/commands/rules-list' },
								{ label: 'rules remove', slug: 'reference/commands/rules-remove' },
								{ label: 'rules update', slug: 'reference/commands/rules-update' },
							],
						},
						{
							label: 'Rules',
							items: [
								{ label: 'References', slug: 'reference/rules/rule-references' },
								{ label: 'Structure', slug: 'reference/rules/rule-structure' },
							],
						},
						{
							label: 'Configuration',
							items: [
								{ label: 'Configuration File', slug: 'reference/configuration/config-file' },
							],
						}
					],
				}
			],
		}),
		],
		vite: {
			plugins: [tailwindcss()],
		},
	});