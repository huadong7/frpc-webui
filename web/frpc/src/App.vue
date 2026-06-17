<template>
  <div id="app">
    <div class="cyber-background" :class="{ 'dark-mode': isDark }">
      <div class="grid-pattern"></div>
      <div class="glow-orb orb-1"></div>
      <div class="glow-orb orb-2"></div>
      <div class="scan-line"></div>
      <div class="mouse-glow" ref="mouseGlowRef"></div>
      <div class="color-spots">
        <div class="color-spot spot-1"></div>
        <div class="color-spot spot-2"></div>
        <div class="color-spot spot-3"></div>
        <div class="color-spot spot-4"></div>
      </div>
    </div>

    <header class="header">
      <div class="header-content">
        <div class="brand-section">
          <button v-if="isMobile" class="hamburger-btn" @click="toggleSidebar" aria-label="Toggle menu">
            <span class="hamburger-icon">&#9776;</span>
          </button>
          <div class="logo-wrapper">
            <LogoIcon class="logo-icon" />
          </div>
          <span class="divider">/</span>
          <span class="brand-name cyber-glow-text">frp</span>
          <span class="badge cyber-glow-text">Client</span>
        </div>

        <div class="header-controls">
          <a
            class="github-link"
            href="https://github.com/fatedier/frp"
            target="_blank"
            aria-label="GitHub"
          >
            <GitHubIcon class="github-icon" />
          </a>
          <el-switch
            v-model="isDark"
            inline-prompt
            :active-icon="Moon"
            :inactive-icon="Sunny"
            class="theme-switch"
          />
        </div>
      </div>
    </header>

    <div class="layout">
      <!-- Mobile overlay -->
      <div
        v-if="isMobile && sidebarOpen"
        class="sidebar-overlay"
        @click="closeSidebar"
      />

      <aside class="sidebar" :class="{ 'mobile-open': isMobile && sidebarOpen }">
        <nav class="sidebar-nav">
          <router-link
            to="/"
            class="sidebar-link"
            :class="{ active: route.path === '/' }"
            @click="closeSidebar"
          >
            Dashboard
          </router-link>
          <router-link
            to="/profiles"
            class="sidebar-link"
            :class="{ active: route.path.startsWith('/profiles') }"
            @click="closeSidebar"
          >
            Profiles
          </router-link>
          <router-link
            to="/proxies"
            class="sidebar-link"
            :class="{ active: route.path.startsWith('/proxies') }"
            @click="closeSidebar"
          >
            Proxies
          </router-link>
          <router-link
            to="/visitors"
            class="sidebar-link"
            :class="{ active: route.path.startsWith('/visitors') }"
            @click="closeSidebar"
          >
            Visitors
          </router-link>
        </nav>
      </aside>

      <main id="content">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useDark } from '@vueuse/core'
import { Moon, Sunny } from '@element-plus/icons-vue'
import GitHubIcon from './assets/icons/github.svg?component'
import LogoIcon from './assets/icons/logo.svg?component'
import { useResponsive } from './composables/useResponsive'
import { useCyberEffects } from './composables/useCyberEffects'

const route = useRoute()
const isDark = useDark()
const { isMobile } = useResponsive()
const { mouseGlowRef } = useCyberEffects(isDark)

defineExpose({ mouseGlowRef })

const sidebarOpen = ref(false)

const toggleSidebar = () => {
  sidebarOpen.value = !sidebarOpen.value
}

const closeSidebar = () => {
  sidebarOpen.value = false
}

watch(() => route.path, () => {
  if (isMobile.value) {
    closeSidebar()
  }
})
</script>

<style lang="scss">
body {
  margin: 0;
  font-family: ui-sans-serif, -apple-system, system-ui, Segoe UI, Helvetica,
    Arial, sans-serif;
}

*,
:after,
:before {
  box-sizing: border-box;
  -webkit-tap-highlight-color: transparent;
}

html, body {
  height: 100%;
  overflow: hidden;
}

#app {
  height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background-color: $color-bg-secondary;
  position: relative;
}

// Background effects
.cyber-background {
  position: fixed;
  inset: 0;
  z-index: -1;
  overflow: hidden;
  pointer-events: none;
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.03) 1px, transparent 1px);
  background-size: 40px 40px;
  opacity: 0;
  transition: opacity 0.5s ease;
  animation: grid-scroll 30s linear infinite;
}

.cyber-background.dark-mode .grid-pattern {
  opacity: 1;
}

.glow-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0;
  transition: opacity 0.5s ease;

  &.orb-1 {
    width: 400px;
    height: 400px;
    background: radial-gradient(circle, rgba(0, 212, 255, 0.12), transparent);
    top: -100px;
    right: -100px;
    animation: float-1 15s ease-in-out infinite;
  }

  &.orb-2 {
    width: 300px;
    height: 300px;
    background: radial-gradient(circle, rgba(168, 85, 247, 0.08), transparent);
    bottom: -50px;
    left: -50px;
    animation: float-2 20s ease-in-out infinite;
  }
}

.cyber-background.dark-mode .glow-orb {
  opacity: 0.7;
}

.scan-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg,
    transparent,
    rgba(0, 212, 255, 0.4),
    transparent
  );
  opacity: 0;
  transition: opacity 0.5s ease;
  animation: scan 8s linear infinite;
}

.cyber-background.dark-mode .scan-line {
  opacity: 0.25;
}

// Mouse glow effect
.mouse-glow {
  position: fixed;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  background: radial-gradient(
    circle,
    rgba(0, 212, 255, 0.06) 0%,
    rgba(168, 85, 247, 0.03) 40%,
    transparent 70%
  );
  pointer-events: none;
  transform: translate(-50%, -50%);
  opacity: 0;
  transition: opacity 0.4s ease;
  z-index: 0;
  will-change: transform;
}

.cyber-background.dark-mode .mouse-glow {
  opacity: 1;
}

.mouse-glow.active {
  opacity: 1;
}

// Color spots
.color-spots {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.color-spot {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0;
  transition: opacity 0.8s ease;
  will-change: transform;
}

.cyber-background.dark-mode .color-spot {
  opacity: 0.5;
}

.spot-1 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(0, 212, 255, 0.08), transparent);
  top: 10%;
  left: 60%;
  animation: spot-drift-1 25s ease-in-out infinite;
}

.spot-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(168, 85, 247, 0.06), transparent);
  top: 50%;
  left: 20%;
  animation: spot-drift-2 30s ease-in-out infinite;
}

.spot-3 {
  width: 350px;
  height: 350px;
  background: radial-gradient(circle, rgba(0, 255, 136, 0.05), transparent);
  top: 70%;
  left: 75%;
  animation: spot-drift-3 22s ease-in-out infinite;
}

.spot-4 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(255, 170, 0, 0.04), transparent);
  top: 20%;
  left: 10%;
  animation: spot-drift-4 28s ease-in-out infinite;
}

@keyframes spot-drift-1 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(-80px, 60px) scale(1.1); }
  66% { transform: translate(40px, -40px) scale(0.9); }
}

@keyframes spot-drift-2 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(60px, -80px) scale(0.95); }
  66% { transform: translate(-50px, 30px) scale(1.05); }
}

@keyframes spot-drift-3 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(-60px, -50px) scale(1.08); }
}

@keyframes spot-drift-4 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(70px, 40px) scale(0.92); }
}

@keyframes grid-scroll {
  from { background-position: 0 0; }
  to { background-position: 40px 40px; }
}

@keyframes float-1 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-30px, 30px); }
}

@keyframes float-2 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, -20px); }
}

@keyframes scan {
  0% { top: -2px; }
  100% { top: 100%; }
}

// Header
.header {
  flex-shrink: 0;
  background: $glass-bg;
  backdrop-filter: $glass-blur;
  -webkit-backdrop-filter: $glass-blur;
  height: $header-height;
  position: relative;
  z-index: 10;

  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg,
      transparent,
      $accent-cyan,
      $accent-purple,
      transparent
    );
    opacity: 0.4;
  }
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 $spacing-xl;
}

.brand-section {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.logo-wrapper {
  display: flex;
  align-items: center;
}

.logo-icon {
  width: 28px;
  height: 28px;
}

.divider {
  color: rgba(0, 212, 255, 0.3);
  font-size: 22px;
  font-weight: 200;
}

.brand-name {
  @include gradient-text;
  font-weight: $font-weight-semibold;
  font-size: $font-size-xl;
  letter-spacing: -0.5px;
}

// Cyber glow text effect
.cyber-glow-text {
  position: relative;
  transition: text-shadow 0.3s ease, color 0.3s ease;
  cursor: default;

  &::before {
    content: '';
    position: absolute;
    inset: -8px;
    background: radial-gradient(
      ellipse at center,
      rgba(0, 212, 255, 0.08),
      transparent 70%
    );
    border-radius: 8px;
    opacity: 0;
    transition: opacity 0.3s ease;
    pointer-events: none;
  }

  &:hover {
    text-shadow: 0 0 12px rgba(0, 212, 255, 0.5),
                 0 0 24px rgba(0, 212, 255, 0.2);

    &::before {
      opacity: 1;
    }
  }
}

.badge {
  font-size: $font-size-xs;
  font-weight: $font-weight-medium;
  color: $accent-cyan;
  background: var(--color-accent-cyan-light);
  border: 1px solid rgba(0, 212, 255, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.github-link {
  @include flex-center;
  width: 28px;
  height: 28px;
  border-radius: $radius-sm;
  color: $color-text-secondary;
  transition: all $transition-fast;

  &:hover {
    background: $color-bg-hover;
    color: $accent-cyan;
  }
}

.github-icon {
  width: 18px;
  height: 18px;
}

.theme-switch {
  --el-switch-on-color: #00d4ff;
  --el-switch-off-color: #1e2035;
  --el-switch-border-color: var(--glass-border);
}

html:not(.dark) .theme-switch {
  --el-switch-on-color: #0088cc;
  --el-switch-off-color: #e4e7ed;
}

.theme-switch .el-switch__core .el-switch__inner .el-icon {
  color: #909399 !important;
}

// Layout
.layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: $sidebar-width;
  flex-shrink: 0;
  background: $glass-bg;
  backdrop-filter: $glass-blur;
  -webkit-backdrop-filter: $glass-blur;
  border-right: 1px solid $glass-border;
  padding: $spacing-lg $spacing-md;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: 1px;
    background: linear-gradient(180deg,
      transparent,
      $accent-cyan,
      transparent
    );
    opacity: 0.25;
  }
}

.sidebar-nav {
  @include flex-column;
  gap: 2px;
}

.sidebar-link {
  display: block;
  text-decoration: none;
  font-size: $font-size-lg;
  color: $color-text-secondary;
  padding: 10px $spacing-md;
  border-radius: $radius-md;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 0;
    background: $accent-cyan;
    border-radius: 0 2px 2px 0;
    transition: height 0.2s ease;
  }

  &:hover {
    color: $color-text-primary;
    background: rgba(0, 212, 255, 0.04);

    &::before {
      height: 50%;
    }
  }

  &.active {
    color: $accent-cyan;
    background: rgba(0, 212, 255, 0.08);
    font-weight: $font-weight-medium;

    &::before {
      height: 55%;
      box-shadow: 0 0 8px rgba(0, 212, 255, 0.5);
    }
  }
}

// Hamburger button (mobile only)
.hamburger-btn {
  @include flex-center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: $radius-sm;
  background: transparent;
  cursor: pointer;
  padding: 0;
  transition: background $transition-fast;

  &:hover {
    background: $color-bg-hover;
  }
}

.hamburger-icon {
  font-size: 20px;
  line-height: 1;
  color: $color-text-primary;
}

// Mobile overlay
.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 99;
}

#content {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  background: $color-bg-primary;
  position: relative;

  &::before {
    content: '';
    position: fixed;
    inset: 0;
    background: radial-gradient(ellipse at top right,
      rgba(0, 212, 255, 0.02),
      transparent 60%
    );
    pointer-events: none;
    z-index: 0;
  }
}

// Page transition
.page-enter-active,
.page-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

// Common page styles
.page-title {
  font-size: $font-size-xl + 2px;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

.page-subtitle {
  font-size: $font-size-md;
  color: $color-text-muted;
  margin: $spacing-sm 0 0;
}

.icon-btn {
  @include flex-center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: $radius-sm;
  background: transparent;
  color: $color-text-muted;
  cursor: pointer;
  transition: all $transition-fast;

  &:hover {
    background: $color-bg-hover;
    color: $accent-cyan;
  }
}

.search-input {
  width: 200px;

  .el-input__wrapper {
    border-radius: 10px;
    background: $color-bg-tertiary;
    border: 1px solid var(--glass-border);
    box-shadow: none !important;

    &:hover {
      border-color: rgba(0, 212, 255, 0.2);
    }

    &.is-focus {
      border-color: rgba(0, 212, 255, 0.4);
      box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.08) !important;
    }
  }

  .el-input__inner {
    color: $color-text-primary;
  }

  .el-input__prefix {
    color: $color-text-muted;
  }

  @include mobile {
    flex: 1;
    width: auto;
  }
}

// Element Plus global overrides
.el-button {
  font-weight: $font-weight-medium;
}

.el-tag {
  font-weight: $font-weight-medium;
}

.el-switch {
  --el-switch-on-color: #00d4ff;
  --el-switch-off-color: #dcdfe6;
}

html:not(.dark) .el-switch {
  --el-switch-on-color: #0088cc;
  --el-switch-off-color: #dcdfe6;
}

html.dark .el-switch {
  --el-switch-on-color: #00d4ff;
  --el-switch-off-color: #1e2035;
}

.el-radio {
  --el-radio-text-color: var(--color-text-primary) !important;
  --el-radio-input-border-color-hover: rgba(0, 212, 255, 0.3) !important;
  --el-color-primary: var(--color-primary) !important;
}

.el-form-item {
  margin-bottom: 16px;
}

.el-loading-mask {
  border-radius: $radius-md;
}

// Select overrides
.el-select__wrapper {
  border-radius: $radius-md !important;
  box-shadow: 0 0 0 1px var(--glass-border) inset !important;
  transition: all $transition-fast;

  &:hover {
    box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.2) inset !important;
  }

  &.is-focused {
    box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.4) inset, 0 0 8px rgba(0, 212, 255, 0.1) !important;
  }
}

.el-select-dropdown {
  border-radius: $radius-md !important;
  border: 1px solid var(--glass-border) !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2) !important;
  padding: 4px !important;
}

html.dark .el-select-dropdown {
  background: rgba(15, 16, 25, 0.92) !important;
  backdrop-filter: blur(20px) !important;
  -webkit-backdrop-filter: blur(20px) !important;
  border-color: rgba(0, 212, 255, 0.1) !important;
}

.el-select-dropdown__item {
  border-radius: $radius-sm;
  margin: 2px 0;
  transition: background $transition-fast;

  &.is-selected {
    color: $accent-cyan;
    font-weight: $font-weight-medium;
  }
}

// Input overrides
.el-input__wrapper {
  border-radius: $radius-md !important;
  box-shadow: 0 0 0 1px var(--glass-border) inset !important;
  transition: all $transition-fast;

  &:hover {
    box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.2) inset !important;
  }

  &.is-focus {
    box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.4) inset, 0 0 8px rgba(0, 212, 255, 0.1) !important;
  }
}

// Status pill (shared)
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  padding: 4px 12px;
  border-radius: 20px;
  text-transform: capitalize;
  transition: all 0.2s ease;

  &.running {
    background: rgba(0, 255, 136, 0.1);
    color: #00ff88;
    border: 1px solid rgba(0, 255, 136, 0.2);

    .status-dot {
      animation: dot-pulse 2s ease-in-out infinite;
    }
  }

  &.error {
    background: rgba(255, 68, 102, 0.1);
    color: #ff4466;
    border: 1px solid rgba(255, 68, 102, 0.2);
  }

  &.waiting {
    background: rgba(255, 170, 0, 0.1);
    color: #ffaa00;
    border: 1px solid rgba(255, 170, 0, 0.2);
  }

  &.disabled {
    background: $color-bg-muted;
    color: $color-text-light;
    border: 1px solid var(--glass-border);
  }

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 6px currentColor;
  }
}

@keyframes dot-pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.3); opacity: 0.7; }
}

// Mobile
@include mobile {
  .header-content {
    padding: 0 $spacing-lg;
  }

  .sidebar {
    position: fixed;
    top: $header-height;
    left: 0;
    bottom: 0;
    z-index: 100;
    background: $glass-bg;
    backdrop-filter: $glass-blur;
    -webkit-backdrop-filter: $glass-blur;
    transform: translateX(-100%);
    transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    border-right: 1px solid $glass-border;

    &.mobile-open {
      transform: translateX(0);
    }
  }

  .sidebar-nav {
    flex-direction: column;
    gap: 2px;
  }

  #content {
    width: 100%;
  }

  // Select dropdown overflow prevention
  .el-select-dropdown {
    max-width: calc(100vw - 32px);
  }
}
</style>
