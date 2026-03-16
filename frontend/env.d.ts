/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_ENV: 'development' | 'internal' | 'public' | 'production'
  // 在此处继续添加其他 VITE_ 变量的类型声明
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
