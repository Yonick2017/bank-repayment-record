export type Currency = 'RMB' | 'HKD'
export type CardOption =
  | 'BOCHK Visa'
  | 'BOCHK Mastercard'
  | 'HSBC Visa Gold'
  | 'HSBC Pulse'

export interface RepaymentRecord {
  id: string
  card: CardOption
  currency: Currency
  amount: number
  repaymentTime: string
}

export interface MonthlyGroup {
  month: string
  records: RepaymentRecord[]
}

export interface StatsSummary {
  monthlyTotals: Record<Currency, number>
  averageMonthlySpending: Record<Currency, number>
  formulaLabel: string
}

export interface HomeSummary {
  currentMonth: string
  monthlyTotals: Record<Currency, number>
}

export interface HistoryFilters {
  card: CardOption | ''
  currency: Currency | ''
}
