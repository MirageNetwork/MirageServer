import { resolve } from 'path'
import { defineConfig } from 'vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import vue from '@vitejs/plugin-vue'

const INVALID_CHAR_REGEX = /[\x00-\x1F\x7F<>*#"{}|^[\]`;?:&=+$,]/g;
const DRIVE_LETTER_REGEX = /^[a-z]:/i;

// https://vitejs.dev/config/
export default defineConfig({
  base: '/cockpit/',
  mode: 'production',
  plugins: [
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
    vue(),
  ],
  build: {
    /*
    watch:{

    },
    */
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
      },
      output:{
        // https://github.com/rollup/rollup/blob/master/src/utils/sanitizeFileName.ts
        sanitizeFileName(fileName) {
          const match = DRIVE_LETTER_REGEX.exec(fileName);
          const driveLetter = match ? match[0] : "";
          return (
            driveLetter +
            fileName.slice(driveLetter.length).replace(INVALID_CHAR_REGEX, "")
          );
        },
      }
    },
    emptyOutDir: true,
    // 指定输出路径，默认'dist'
    outDir: '../controller/cockpit_html/',
    // 指定生成静态资源的存放路径(相对于build.outDir)
    assetsDir: 'assets',
    // 小于此阈值的导入或引用资源将内联为base64编码，设置为0可禁用此项。默认4096（4kb）
    assetsInlineLimit: '4096',
    // 启用/禁用CSS代码拆分，如果禁用，整个项目的所有CSS将被提取到一个CSS文件中,默认true
    cssCodeSplit: true,
    // 构建后是否生成source map文件，默认false
    sourcemap: false,
    // 为true时，会生成manifest.json文件，用于后端集成
    manifest: false
  }
})
