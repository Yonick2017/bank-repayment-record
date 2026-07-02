import type {
  Currency,
  HistoryFilters,
  HomeSummary,
  MonthlyGroup,
  RepaymentRecord,
  StatsSummary,
} from '../types'
import { DEFAULT_FORMULA_LABEL } from '../utils/format'

const API_BASE = '/api'

interface HistoryResponse {
  months?: HistoryMonthResponse[]
}

interface HistoryMonthResponse {
  month: string
  records: RepaymentItemResponse[]
}

interface RepaymentItemResponse {
  id: number | string
  cardName: string
  currency: Currency
  amount: string | number
  repaymentAt: string
}

interface MonthlyStatsResponse {
  currencies?: Array<{
    currency: Currency
    monthlyTotals?: Array<{
      month: string
      total: string | number
    }>
    averageMonthlyRepayment?: string | number
  }>
}

interface CurrentMonthStatsResponse {
  month?: string
  totals?: Partial<Record<Currency, string | number>>
}

interface CreateRepaymentResponse {
  data?: RepaymentItemResponse
}

function emptyCurrencyRecord(): Record<Currency, number> {
  return { RMB: 0, HKD: 0 }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...init,
  })

  if (!response.ok) {
    let reason = `API request failed: ${response.status}`
    try {
      const payload = (await response.json()) as { error?: string }
      if (payload.error) {
        reason = payload.error
      }
    } catch {
      // Ignore JSON parsing failures for non-JSON error responses.
    }
    throw new Error(reason)
  }

  if (response.status === 204) {
    return {} as T
  }

  return (await response.json()) as T
}

function buildQuery(filters: HistoryFilters): string {
  const params = new URLSearchParams()
  if (filters.card) {
    params.set('cardName', filters.card)
  }
  if (filters.currency) {
    params.set('currency', filters.currency)
  }
  const query = params.toString()
  return query ? `?${query}` : ''
}

function toNumber(value: string | number | undefined): number {
  if (value === undefined) {
    return 0
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function mapRepaymentRecord(item: RepaymentItemResponse): RepaymentRecord {
  return {
    id: String(item.id),
    card: item.cardName as RepaymentRecord['card'],
    currency: item.currency,
    amount: toNumber(item.amount),
    repaymentTime: item.repaymentAt,
  }
}

function normalizeMonthlyStats(response?: MonthlyStatsResponse): StatsSummary {
  const monthlyTotals = emptyCurrencyRecord()
  const averageMonthlySpending = emptyCurrencyRecord()

  for (const item of response?.currencies ?? []) {
    let total = 0
    for (const monthItem of item.monthlyTotals ?? []) {
      total += toNumber(monthItem.total)
    }
    monthlyTotals[item.currency] = total
    averageMonthlySpending[item.currency] = toNumber(item.averageMonthlyRepayment)
  }

  return {
    monthlyTotals,
    averageMonthlySpending,
    formulaLabel: DEFAULT_FORMULA_LABEL,
  }
}

export async function fetchHomeSummary(): Promise<HomeSummary> {
  const response = await request<CurrentMonthStatsResponse>('/stats/current-month')
  const now = new Date()
  const fallbackMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  return {
    currentMonth: response.month ?? fallbackMonth,
    monthlyTotals: {
      RMB: toNumber(response.totals?.RMB),
      HKD: toNumber(response.totals?.HKD),
    },
  }
}

export async function createRepayment(input: {
  card: string
  currency: Currency
  amount: number
  repaymentTime: string
}): Promise<RepaymentRecord> {
  const response = await request<CreateRepaymentResponse>('/repayments', {
    method: 'POST',
    body: JSON.stringify({
      cardName: input.card,
      currency: input.currency,
      amount: input.amount.toFixed(2),
      repaymentAt: input.repaymentTime,
    }),
  })
  if (!response.data) {
    throw new Error('Missing create repayment response data')
  }
  return mapRepaymentRecord(response.data)
}

export async function fetchHistory(filters: HistoryFilters): Promise<MonthlyGroup[]> {
  const response = await request<HistoryResponse>(`/repayments/history${buildQuery(filters)}`)
  return (response.months ?? []).map((item) => ({
    month: item.month,
    records: item.records.map(mapRepaymentRecord),
  }))
}

export async function fetchStats(filters: HistoryFilters): Promise<StatsSummary> {
  const response = await request<MonthlyStatsResponse>(`/stats/monthly${buildQuery(filters)}`)
  return normalizeMonthlyStats(response)
}

export async function deleteRepayment(id: string): Promise<void> {
  await request(`/repayments/${id}`, {
    method: 'DELETE',
  })
}
