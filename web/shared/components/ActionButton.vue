<template>
  <button
    type="button"
    class="action-button"
    :class="[variant, size, { 'is-loading': loading, 'is-danger': danger }]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <div v-if="loading" class="spinner"></div>
    <span v-if="loading && loadingText">{{ loadingText }}</span>
    <slot v-else />
  </button>
</template>

<script setup lang="ts">
interface Props {
  variant?: 'primary' | 'secondary' | 'outline'
  size?: 'small' | 'medium' | 'large'
  disabled?: boolean
  loading?: boolean
  loadingText?: string
  danger?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'medium',
  disabled: false,
  loading: false,
  loadingText: '',
  danger: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}
</script>

<style scoped lang="scss">
.action-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  border-radius: $radius-md;
  font-weight: $font-weight-medium;
  cursor: pointer;
  transition: all $transition-fast;
  border: 1px solid transparent;
  white-space: nowrap;

  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: spin 0.75s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  &.small {
    padding: 5px $spacing-md;
    font-size: $font-size-sm;
  }

  &.medium {
    padding: $spacing-sm $spacing-lg;
    font-size: $font-size-md;
  }

  &.large {
    padding: 10px $spacing-xl;
    font-size: $font-size-lg;
  }

  &.primary {
    background: linear-gradient(135deg, $accent-cyan, #0099cc);
    border-color: $accent-cyan;
    color: #fff;
    box-shadow: 0 2px 10px rgba(0, 212, 255, 0.25);

    &:hover:not(:disabled) {
      background: linear-gradient(135deg, #33ddff, #00aadd);
      box-shadow: 0 4px 18px rgba(0, 212, 255, 0.35);
      transform: translateY(-1px);
    }

    &:active:not(:disabled) {
      transform: translateY(0);
      box-shadow: 0 2px 10px rgba(0, 212, 255, 0.25);
    }
  }

  &.secondary {
    background: $glass-bg;
    backdrop-filter: blur(8px);
    border-color: $glass-border;
    color: $color-text-primary;

    &:hover:not(:disabled) {
      border-color: rgba(0, 212, 255, 0.25);
      box-shadow: 0 0 8px rgba(0, 212, 255, 0.1);
    }
  }

  &.outline {
    background: transparent;
    border-color: $glass-border;
    color: $color-text-primary;

    &:hover:not(:disabled) {
      background: rgba(0, 212, 255, 0.04);
      border-color: rgba(0, 212, 255, 0.25);
      box-shadow: 0 0 8px rgba(0, 212, 255, 0.08);
    }
  }

  &.is-danger {
    &.primary {
      background: linear-gradient(135deg, $color-danger, #cc3355);
      border-color: $color-danger;

      &:hover:not(:disabled) {
        background: linear-gradient(135deg, #ff5577, $color-danger);
        box-shadow: 0 4px 18px rgba(255, 68, 102, 0.35);
      }
    }

    &.outline, &.secondary {
      color: $color-danger;
      border-color: rgba(255, 68, 102, 0.2);

      &:hover:not(:disabled) {
        border-color: rgba(255, 68, 102, 0.4);
        background: rgba(255, 68, 102, 0.06);
        box-shadow: 0 0 8px rgba(255, 68, 102, 0.1);
      }
    }
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}
</style>
