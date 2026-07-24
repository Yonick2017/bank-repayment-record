<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchAuthMe, fetchHomeSummary, logout, UnauthorizedError } from './api/client'
import EntryFlow from './components/EntryFlow.vue'
import HistoryPage from './components/HistoryPage.vue'
import HomePage from './components/HomePage.vue'
import LoginPage from './components/LoginPage.vue'
import type { HomeSummary } from './types'

type ViewMode = 'home' | 'entry' | 'history'
type AuthState = 'loading' | 'anonymous' | 'authenticated'

const authState = ref<AuthState>('loading')
const authError = ref('')
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
    if (error instanceof UnauthorizedError) {
      authState.value = 'anonymous'
      return
    }
    homeError.value = error instanceof Error ? error.message : '首页统计加载失败'
  } finally {
    homeLoading.value = false
  }
}

async function checkAuth() {
  authState.value = 'loading'
  authError.value = ''
  try {
    const ok = await fetchAuthMe()
    if (!ok) {
      authState.value = 'anonymous'
      return
    }
    authState.value = 'authenticated'
    view.value = 'home'
    await loadHomeSummary()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : '鉴权状态检查失败'
    authState.value = 'anonymous'
  }
}

async function onLoginSuccess() {
  authState.value = 'authenticated'
  view.value = 'home'
  await loadHomeSummary()
}

async function onLogout() {
  try {
    await logout()
  } catch {
    // Still clear local authenticated view if network/API fails.
  }
  authState.value = 'anonymous'
  view.value = 'home'
  homeError.value = ''
}

function onCreateDone() {
  void loadHomeSummary()
}

onMounted(() => {
  void checkAuth()
})
</script>

<template>
  <main class="app-shell">
    <p v-if="authState === 'loading'" class="hint auth-loading">正在检查登录状态...</p>

    <template v-else-if="authState === 'anonymous'">
      <p v-if="authError" class="error">{{ authError }}</p>
      <LoginPage @success="onLoginSuccess" />
    </template>

    <template v-else>
      <HomePage
        v-if="view === 'home'"
        :loading="homeLoading"
        :error-message="homeError"
        :summary="homeSummary"
        @record="view = 'entry'"
        @history="view = 'history'"
        @refresh="loadHomeSummary"
        @logout="onLogout"
      />

      <EntryFlow
        v-else-if="view === 'entry'"
        @done="onCreateDone"
        @view-history="view = 'history'"
      />

      <HistoryPage v-else @back-home="view = 'home'" />
    </template>
  </main>
</template>
