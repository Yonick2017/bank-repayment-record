<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const parsed = computed(() => {
  const next = new Date(props.modelValue)
  return Number.isNaN(next.getTime()) ? new Date() : next
})

const year = ref(parsed.value.getFullYear())
const month = ref(parsed.value.getMonth() + 1)
const day = ref(parsed.value.getDate())
const hour = ref(parsed.value.getHours())
const minute = ref(parsed.value.getMinutes())

const yearOptions = computed(() => {
  const center = parsed.value.getFullYear()
  return Array.from({ length: 11 }, (_, index) => center - 5 + index)
})

const monthOptions = Array.from({ length: 12 }, (_, index) => index + 1)
const hourOptions = Array.from({ length: 24 }, (_, index) => index)
const minuteOptions = Array.from({ length: 60 }, (_, index) => index)

const dayOptions = computed(() => {
  const lastDate = new Date(year.value, month.value, 0).getDate()
  return Array.from({ length: lastDate }, (_, index) => index + 1)
})

watch(parsed, (value) => {
  year.value = value.getFullYear()
  month.value = value.getMonth() + 1
  day.value = value.getDate()
  hour.value = value.getHours()
  minute.value = value.getMinutes()
})

watch([year, month], () => {
  if (!dayOptions.value.includes(day.value)) {
    day.value = dayOptions.value.at(-1) ?? 1
  }
})

watch([year, month, day, hour, minute], () => {
  const normalized = new Date(year.value, month.value - 1, day.value, hour.value, minute.value)
  if (Number.isNaN(normalized.getTime())) {
    return
  }

  const value = [
    String(normalized.getFullYear()),
    String(normalized.getMonth() + 1).padStart(2, '0'),
    String(normalized.getDate()).padStart(2, '0'),
  ].join('-')
  const time = `${String(normalized.getHours()).padStart(2, '0')}:${String(
    normalized.getMinutes(),
  ).padStart(2, '0')}`
  emit('update:modelValue', `${value}T${time}`)
})
</script>

<template>
  <div class="wheel-grid">
    <div class="wheel-column">
      <label>年</label>
      <select v-model.number="year" aria-label="年份">
        <option v-for="item in yearOptions" :key="item" :value="item">{{ item }}</option>
      </select>
    </div>
    <div class="wheel-column">
      <label>月</label>
      <select v-model.number="month" aria-label="月份">
        <option v-for="item in monthOptions" :key="item" :value="item">
          {{ String(item).padStart(2, '0') }}
        </option>
      </select>
    </div>
    <div class="wheel-column">
      <label>日</label>
      <select v-model.number="day" aria-label="日期">
        <option v-for="item in dayOptions" :key="item" :value="item">
          {{ String(item).padStart(2, '0') }}
        </option>
      </select>
    </div>
    <div class="wheel-column">
      <label>时</label>
      <select v-model.number="hour" aria-label="小时">
        <option v-for="item in hourOptions" :key="item" :value="item">
          {{ String(item).padStart(2, '0') }}
        </option>
      </select>
    </div>
    <div class="wheel-column">
      <label>分</label>
      <select v-model.number="minute" aria-label="分钟">
        <option v-for="item in minuteOptions" :key="item" :value="item">
          {{ String(item).padStart(2, '0') }}
        </option>
      </select>
    </div>
  </div>
</template>
