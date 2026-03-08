# Frontend

空 React 项目（Vite + TypeScript），用于 Rocket-Leaf 桌面端界面。

## 开发

```sh
npm install
npm run dev
```

## 构建

```sh
npm run build
```

## 说明

- 与 Wails 集成：保留 `@wailsio/runtime` 与 `bindings/`（由 Wails 生成），便于后续接入 Go 后端。
- 可在 `src/App.tsx` 及新建的组件中自由搭建界面。
