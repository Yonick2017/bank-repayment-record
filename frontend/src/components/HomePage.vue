<script setup lang="ts">
import type { HomeSummary } from '../types'
import { formatMonthLabel, formatMoney } from '../utils/format'

const props = defineProps<{
  loading: boolean
  errorMessage: string
  summary: HomeSummary
}>()

const emit = defineEmits<{
  record: []
  history: []
  refresh: []
  logout: []
}>()
</script>

<template>
  <section class="page-shell">
    <header class="home-header">
      <div class="home-header-row">
        <div>
          <h1>还款记录</h1>
          <p>本地账务记录与统计</p>
        </div>
        <button type="button" class="ghost" @click="emit('logout')">登出</button>
      </div>
    </header>

    <article class="summary-card">
      <div class="summary-head">
        <h2>{{ formatMonthLabel(props.summary.currentMonth) }}已还款总额</h2>
        <button type="button" class="ghost" @click="emit('refresh')">刷新</button>
      </div>
      <p v-if="props.loading" class="hint">加载中...</p>
      <p v-else-if="props.errorMessage" class="error">{{ props.errorMessage }}</p>
      <div v-else class="summary-grid">
        <div class="summary-item">
          <span class="label">RMB</span>
          <strong>{{ formatMoney('RMB', props.summary.monthlyTotals.RMB) }}</strong>
        </div>
        <div class="summary-item">
          <span class="label">HKD</span>
          <strong>{{ formatMoney('HKD', props.summary.monthlyTotals.HKD) }}</strong>
        </div>
      </div>
    </article>

    <div class="home-actions">
      <button type="button" class="primary" @click="emit('record')">记录还款</button>
      <button type="button" class="secondary" @click="emit('history')">查看历史记录</button>
    </div>

    <footer class="site-footer">
      <a
        class="beian-link"
        href="https://beian.miit.gov.cn/"
        target="_blank"
        rel="noopener noreferrer"
      >备案号</a>
    </footer>
  </section>
</template>
