import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import * as path from "path";

/**
 * https://vitejs.dev/config/
 * @type {import('vite').UserConfig}
 */
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    server: {
        proxy: {
            '/api/v1/': {
                changeOrigin: true,
                target: "http://localhost:3000/"
            },
        },
    },
});