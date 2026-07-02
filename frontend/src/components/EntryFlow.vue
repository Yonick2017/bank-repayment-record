<script setup lang="ts">
import { computed, ref } from 'vue'
import { createRepayment } from '../api/client'
import { CARD_OPTIONS, CURRENCY_OPTIONS } from '../constants'
import type { CardOption, Currency, RepaymentRecord } from '../types'
import { formatDateTime, formatSignedAmount, isValidSignedAmount, toMinutePrecision } from '../utils/format'
import DateTimeWheel from './DateTimeWheel.vue'

const emit = defineEmits<{
  done: [record: RepaymentRecord]
  viewHistory: []
}>()

type EntryStep = 1 | 2 | 3 | 4

const step = ref<EntryStep>(1)
const selectedCard = ref<CardOption | ''>('')
const selectedCurrency = ref<Currency>('RMB')
const amountInput = ref('')
const repaymentTime = ref(toMinutePrecision(new Date()))
const loading = ref(false)
const errorMessage = ref('')
const fieldMessage = ref('')
const createdRecord = ref<RepaymentRecord | null>(null)

const amountPreview = computed(() => {
  if (!isValidSignedAmount(amountInput.value)) {
    return '--'
  }
  return formatSignedAmount(Number(amountInput.value))
})

function goNextFromStepOne() {
  if (!selectedCard.value) {
    fieldMessage.value = '必须选择银行卡'
    return
  }
  fieldMessage.value = ''
  step.value = 2
}

function goNextFromStepTwo() {
  if (!isValidSignedAmount(amountInput.value)) {
    fieldMessage.value = '金额格式必须为数字，且最多两位小数'
    return
  }
  fieldMessage.value = ''
  step.value = 3
}

function goBack() {
  errorMessage.value = ''
  fieldMessage.value = ''
  if (step.value === 2) {
    step.value = 1
  } else if (step.value === 3) {
    step.value = 2
  }
}

async function submitEntry() {
  if (!selectedCard.value || !isValidSignedAmount(amountInput.value)) {
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const payload = {
      card: selectedCard.value,
      currency: selectedCurrency.value,
      amount: Number(amountInput.value),
      repaymentTime: repaymentTime.value,
    }
    const record = await createRepayment(payload)
    createdRecord.value = record
    step.value = 4
    emit('done', record)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '提交失败'
  } finally {
    loading.value = false
  }
}

function resetForAnother() {
  step.value = 1
  selectedCard.value = ''
  selectedCurrency.value = 'RMB'
  amountInput.value = ''
  repaymentTime.value = toMinutePrecision(new Date())
  fieldMessage.value = ''
  errorMessage.value = ''
  createdRecord.value = null
}
</script>

<template>
  <section class="page-shell entry-shell">
    <Transition name="slide-step" mode="out-in">
      <article v-if="step === 1" key="step-1" class="entry-card full-screen-step">
        <h2>步骤 1 / 3：选择银行卡</h2>
        <p class="hint">请选择一张银行卡后继续</p>
        <div class="card-grid">
          <button
            v-for="item in CARD_OPTIONS"
            :key="item"
            type="button"
            class="card-option"
            :class="{ selected: selectedCard === item }"
            @click="selectedCard = item"
          >
            {{ item }}
          </button>
        </div>
        <p v-if="fieldMessage" class="error">{{ fieldMessage }}</p>
        <div class="actions">
          <button type="button" class="primary" @click="goNextFromStepOne">下一步</button>
        </div>
      </article>

      <article v-else-if="step === 2" key="step-2" class="entry-card">
        <h2>步骤 2 / 3：币种与金额</h2>
        <label class="field">
          币种
          <select v-model="selectedCurrency">
            <option v-for="item in CURRENCY_OPTIONS" :key="item" :value="item">{{ item }}</option>
          </select>
        </label>
        <label class="field">
          金额
          <input v-model.trim="amountInput" placeholder="例如 -120.00 或 500.00" />
        </label>
        <p class="hint">预览：{{ amountPreview }}</p>
        <p v-if="fieldMessage" class="error">{{ fieldMessage }}</p>
        <div class="actions split">
          <button type="button" class="ghost" @click="goBack">上一步</button>
          <button type="button" class="primary" @click="goNextFromStepTwo">下一步</button>
        </div>
      </article>

      <article v-else-if="step === 3" key="step-3" class="entry-card">
        <h2>步骤 3 / 3：还款时间</h2>
        <p class="hint">拨盘默认当前本地时间，精确到分钟</p>
        <DateTimeWheel v-model="repaymentTime" />
        <p class="hint">已选：{{ formatDateTime(`${repaymentTime}:00`) }}</p>
        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
        <div class="actions split">
          <button type="button" class="ghost" @click="goBack">上一步</button>
          <button type="button" class="primary" :disabled="loading" @click="submitEntry">
            {{ loading ? '提交中...' : '完成' }}
          </button>
        </div>
      </article>

      <article v-else key="step-4" class="entry-card done-card">
        <h2>记录完成</h2>
        <p class="hint">本次提交摘要</p>
        <ul class="summary-list" v-if="createdRecord">
          <li><span>银行卡</span><strong>{{ createdRecord.card }}</strong></li>
          <li><span>币种</span><strong>{{ createdRecord.currency }}</strong></li>
          <li><span>金额</span><strong>{{ formatSignedAmount(createdRecord.amount) }}</strong></li>
          <li><span>时间</span><strong>{{ formatDateTime(createdRecord.repaymentTime) }}</strong></li>
        </ul>
        <div class="actions split">
          <button type="button" class="secondary" @click="resetForAnother">再记一笔</button>
          <button type="button" class="primary" @click="emit('viewHistory')">查看历史</button>
        </div>
      </article>
    </Transition>
  </section>
</template>
