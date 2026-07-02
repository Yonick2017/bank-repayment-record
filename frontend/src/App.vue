<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchHomeSummary } from './api/client'
import EntryFlow from './components/EntryFlow.vue'
import HistoryPage from './components/HistoryPage.vue'
import HomePage from './components/HomePage.vue'
import type { HomeSummary } from './types'

type ViewMode = 'home' | 'entry' | 'history'

const view = ref<ViewMode>('home')
const homeLoading = ref(false)
const homeError = ref('')
const homeSummary = ref<HomeSummary>({
  currentMonth: new Date().toISOString().slice(0, 7),
  monthlyTotals: { RMB: 0, HKD: 0 },
})

async function loadHomeSummary() {
  homeLoading.value = true
  homeError.value = ''
  try {
    homeSummary.value = await fetchHomeSummary()
  } catch (error) {
    homeError.value = error instanceof Error ? error.message : '首页统计加载失败'
  } finally {
    homeLoading.value = false
  }
}

function onCreateDone() {
  void loadHomeSummary()
}

onMounted(() => {
  void loadHomeSummary()
})
</script>

<template>
  <main class="app-shell">
    <HomePage
      v-if="view === 'home'"
      :loading="homeLoading"
      :error-message="homeError"
      :summary="homeSummary"
      @record="view = 'entry'"
      @history="view = 'history'"
      @refresh="loadHomeSummary"
    />

    <EntryFlow
      v-else-if="view === 'entry'"
      @done="onCreateDone"
      @view-history="view = 'history'"
    />

    <HistoryPage v-else @back-home="view = 'home'" />
  </main>
</template>
