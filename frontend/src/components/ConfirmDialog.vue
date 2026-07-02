<script setup lang="ts">
const props = defineProps<{
  open: boolean
  title: string
  confirmText?: string
  cancelText?: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div v-if="props.open" class="dialog-mask" @click.self="emit('cancel')">
    <div class="dialog-card" role="dialog" aria-modal="true" :aria-label="props.title">
      <h3>{{ props.title }}</h3>
      <div class="dialog-content">
        <slot />
      </div>
      <div class="dialog-actions">
        <button type="button" class="ghost" @click="emit('cancel')">
          {{ props.cancelText ?? '取消' }}
        </button>
        <button type="button" class="danger" @click="emit('confirm')">
          {{ props.confirmText ?? '确认删除' }}
        </button>
      </div>
    </div>
  </div>
</template>
