<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { deleteRepayment, fetchHistory, fetchStats } from '../api/client'
import { CARD_OPTIONS, CURRENCY_OPTIONS } from '../constants'
import type { HistoryFilters, MonthlyGroup, RepaymentRecord, StatsSummary } from '../types'
import { DEFAULT_FORMULA_LABEL, formatDateTime, formatMonthLabel, formatMoney, formatSignedAmount } from '../utils/format'
import ConfirmDialog from './ConfirmDialog.vue'

const emit = defineEmits<{
  backHome: []
}>()

const filters = ref<HistoryFilters>({ card: '', currency: '' })
const months = ref<MonthlyGroup[]>([])
const expanded = ref<Record<string, boolean>>({})
const stats = ref<StatsSummary>({
  monthlyTotals: { RMB: 0, HKD: 0 },
  averageMonthlySpending: { RMB: 0, HKD: 0 },
  formulaLabel: DEFAULT_FORMULA_LABEL,
})

const loading = ref(false)
const errorMessage = ref('')
const deleteTarget = ref<RepaymentRecord | null>(null)
const deleteErrorMessage = ref('')

function isExpanded(month: string): boolean {
  return expanded.value[month] ?? true
}

function toggleMonth(month: string) {
  expanded.value[month] = !isExpanded(month)
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [historyData, statsData] = await Promise.all([
      fetchHistory(filters.value),
      fetchStats(filters.value),
    ])
    months.value = historyData
    stats.value = statsData
    for (const item of historyData) {
      if (expanded.value[item.month] === undefined) {
        expanded.value[item.month] = true
      }
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function onFiltersChange() {
  void loadData()
}

function openDeleteDialog(record: RepaymentRecord) {
  deleteErrorMessage.value = ''
  deleteTarget.value = record
}

function closeDeleteDialog() {
  deleteTarget.value = null
  deleteErrorMessage.value = ''
}

async function confirmDelete() {
  if (!deleteTarget.value) {
    return
  }
  const current = deleteTarget.value
  deleteErrorMessage.value = ''
  try {
    await deleteRepayment(current.id)
    closeDeleteDialog()
    await loadData()
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : '删除失败'
  }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="page-shell">
    <div class="history-head">
      <h2>历史记录</h2>
      <button type="button" class="ghost" @click="emit('backHome')">返回首页</button>
    </div>

    <article class="summary-card stats-card">
      <h3>统计摘要</h3>
      <p class="formula">{{ stats.formulaLabel }}</p>
      <div class="summary-grid">
        <div class="summary-item">
          <span class="label">RMB 月总额</span>
          <strong>{{ formatMoney('RMB', stats.monthlyTotals.RMB) }}</strong>
        </div>
        <div class="summary-item">
          <span class="label">RMB 平均月开销</span>
          <strong>{{ formatMoney('RMB', stats.averageMonthlySpending.RMB) }}</strong>
        </div>
        <div class="summary-item">
          <span class="label">HKD 月总额</span>
          <strong>{{ formatMoney('HKD', stats.monthlyTotals.HKD) }}</strong>
        </div>
        <div class="summary-item">
          <span class="label">HKD 平均月开销</span>
          <strong>{{ formatMoney('HKD', stats.averageMonthlySpending.HKD) }}</strong>
        </div>
      </div>
    </article>

    <article class="filter-card">
      <label class="field">
        银行卡筛选
        <select v-model="filters.card" @change="onFiltersChange">
          <option value="">全部</option>
          <option v-for="item in CARD_OPTIONS" :key="item" :value="item">{{ item }}</option>
        </select>
      </label>
      <label class="field">
        币种筛选
        <select v-model="filters.currency" @change="onFiltersChange">
          <option value="">全部</option>
          <option v-for="item in CURRENCY_OPTIONS" :key="item" :value="item">{{ item }}</option>
        </select>
      </label>
    </article>

    <p v-if="loading" class="hint">加载中...</p>
    <p v-else-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-else class="month-list">
      <article v-for="monthItem in months" :key="monthItem.month" class="month-card">
        <button type="button" class="month-toggle" @click="toggleMonth(monthItem.month)">
          <strong>{{ formatMonthLabel(monthItem.month) }}</strong>
          <span>{{ isExpanded(monthItem.month) ? '收起' : '展开' }}</span>
        </button>

        <div v-if="isExpanded(monthItem.month)" class="record-list">
          <article v-for="record in monthItem.records" :key="record.id" class="record-card">
            <div class="record-row">
              <span>银行卡</span><strong>{{ record.card }}</strong>
            </div>
            <div class="record-row">
              <span>币种</span><strong>{{ record.currency }}</strong>
            </div>
            <div class="record-row">
              <span>金额</span><strong>{{ formatSignedAmount(record.amount) }}</strong>
            </div>
            <div class="record-row">
              <span>时间</span><strong>{{ formatDateTime(record.repaymentTime) }}</strong>
            </div>
            <button type="button" class="danger ghost-danger" @click="openDeleteDialog(record)">
              删除
            </button>
          </article>
        </div>
      </article>
      <p v-if="months.length === 0" class="hint">暂无记录</p>
    </div>

    <ConfirmDialog
      :open="Boolean(deleteTarget)"
      title="确认删除记录"
      confirm-text="确认删除"
      cancel-text="取消"
      @confirm="confirmDelete"
      @cancel="closeDeleteDialog"
    >
      <p v-if="deleteTarget">银行卡：{{ deleteTarget.card }}</p>
      <p v-if="deleteTarget">金额：{{ formatSignedAmount(deleteTarget.amount) }}</p>
      <p v-if="deleteTarget">时间：{{ formatDateTime(deleteTarget.repaymentTime) }}</p>
      <p v-if="deleteErrorMessage" class="error">{{ deleteErrorMessage }}</p>
    </ConfirmDialog>
  </section>
</template>
