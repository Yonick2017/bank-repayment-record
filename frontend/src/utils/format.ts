import type { Currency } from '../types'

export const DEFAULT_FORMULA_LABEL =
  '平均月开销 = sum(monthly_total) / months_with_records'

export function formatMonthLabel(month: string): string {
  if (!month) {
    return '--'
  }

  const normalized = month.length >= 7 ? month.slice(0, 7) : month
  return normalized.replace('-', '年') + '月'
}

export function formatDateTime(value: string): string {
  if (!value) {
    return '--'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('zh-HK', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

export function formatSignedAmount(amount: number): string {
  const absolute = Math.abs(amount).toFixed(2)
  return amount < 0 ? `${absolute} CR` : absolute
}

export function formatMoney(currency: Currency, amount: number): string {
  const absolute = Math.abs(amount)
  const locale = currency === 'HKD' ? 'zh-HK' : 'zh-CN'
  const currencyCode = currency === 'HKD' ? 'HKD' : 'CNY'
  const formatted = new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currencyCode,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(absolute)
  return amount < 0 ? `${formatted} CR` : formatted
}

export function toMinutePrecision(value: Date): string {
  const next = new Date(value)
  next.setSeconds(0, 0)
  const year = next.getFullYear()
  const month = String(next.getMonth() + 1).padStart(2, '0')
  const day = String(next.getDate()).padStart(2, '0')
  const hour = String(next.getHours()).padStart(2, '0')
  const minute = String(next.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hour}:${minute}`
}

export function isValidSignedAmount(input: string): boolean {
  return /^-?\d+(\.\d{1,2})?$/.test(input)
}
