<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchPublicConfig, login } from '../api/client'

const emit = defineEmits<{
  success: []
}>()

const password = ref('')
const submitting = ref(false)
const errorMessage = ref('')
const beianText = ref('')

async function onSubmit() {
  if (!password.value || submitting.value) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    await login(password.value)
    password.value = ''
    emit('success')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    const config = await fetchPublicConfig()
    beianText.value = config.beianText
  } catch {
    beianText.value = ''
  }
})
</script>

<template>
  <section class="page-shell login-page">
    <header class="home-header">
      <h1>还款记录</h1>
      <p>请输入访问密码</p>
    </header>

    <form class="login-card" @submit.prevent="onSubmit">
      <label class="login-field">
        <span>密码</span>
        <input
          v-model="password"
          type="password"
          name="password"
          autocomplete="current-password"
          :disabled="submitting"
          required
        />
      </label>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <button type="submit" class="primary login-submit" :disabled="submitting || !password">
        {{ submitting ? '登录中...' : '登录' }}
      </button>
    </form>

    <footer v-if="beianText" class="site-footer">
      <a
        class="beian-link"
        href="https://beian.miit.gov.cn/"
        target="_blank"
        rel="noopener noreferrer"
      >{{ beianText }}</a>
    </footer>
  </section>
</template>
