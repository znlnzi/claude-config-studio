import React from 'react'
import {createRoot} from 'react-dom/client'
import { loader } from '@monaco-editor/react'
import './style.css'
import App from './App'

// 使用 CDN 加载 Monaco Editor，避免 Vite 打包 worker 的兼容性问题
loader.config({
  paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.0/min/vs' },
})

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
