import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';
import tailwindcss from '@tailwindcss/vite';
import path from "path"

export default defineConfig({
	plugins: [solidPlugin(), tailwindcss()],
	server: {
		port: 3000,
		strictPort: true,
	},
	build: {
		target: 'esnext',
		manifest: true,
		rollupOptions: {
			input: {
				index: 'src/index.tsx'
			},
		},
	},
	resolve: {
		alias: {
			"~": path.resolve(__dirname, "./src")
		}
	}
});
