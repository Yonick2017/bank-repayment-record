import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HistoryPage from './HistoryPage.vue'

const fetchHistoryMock = vi.fn()
const fetchStatsMock = vi.fn()
const deleteRepaymentMock = vi.fn()

vi.mock('../api/client', () => ({
  fetchHistory: (filters: unknown) => fetchHistoryMock(filters),
  fetchStats: (filters: unknown) => fetchStatsMock(filters),
  deleteRepayment: (id: string) => deleteRepaymentMock(id),
}))

describe('HistoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('syncs filters for list and stats, and confirms delete', async () => {
    fetchHistoryMock.mockResolvedValue([
      {
        month: '2026-07',
        records: [
          {
            id: 'r-1',
            card: 'BOCHK Visa',
            currency: 'RMB',
            amount: -50,
            repaymentTime: '2026-07-01T09:10:00',
          },
        ],
      },
    ])
    fetchStatsMock.mockResolvedValue({
      monthlyTotals: { RMB: -50, HKD: 0 },
      averageMonthlySpending: { RMB: -50, HKD: 0 },
      formulaLabel: '平均月开销 = sum(monthly_total) / months_with_records',
    })
    deleteRepaymentMock.mockResolvedValue(undefined)

    const wrapper = mount(HistoryPage)
    await flushPromises()

    expect(wrapper.text()).toContain('sum(monthly_total) / months_with_records')
    expect(wrapper.text()).toContain('50.00 CR')

    const selects = wrapper.findAll('select')
    await selects[0].setValue('BOCHK Visa')
    await selects[1].setValue('RMB')
    await flushPromises()

    expect(fetchHistoryMock).toHaveBeenLastCalledWith({ card: 'BOCHK Visa', currency: 'RMB' })
    expect(fetchStatsMock).toHaveBeenLastCalledWith({ card: 'BOCHK Visa', currency: 'RMB' })

    await wrapper.get('.ghost-danger').trigger('click')
    expect(wrapper.text()).toContain('确认删除记录')
    expect(wrapper.text()).toContain('银行卡：BOCHK Visa')
    expect(wrapper.text()).toContain('金额：50.00 CR')

    const dialogButtons = wrapper.findAll('.dialog-actions button')
    await dialogButtons[1].trigger('click')
    await flushPromises()

    expect(deleteRepaymentMock).toHaveBeenCalledWith('r-1')
  })

  it('keeps dialog open and shows error when delete fails', async () => {
    fetchHistoryMock.mockResolvedValue([
      {
        month: '2026-07',
        records: [
          {
            id: 'r-2',
            card: 'HSBC Pulse',
            currency: 'HKD',
            amount: -88,
            repaymentTime: '2026-07-01T09:10:00',
          },
        ],
      },
    ])
    fetchStatsMock.mockResolvedValue({
      monthlyTotals: { RMB: 0, HKD: -88 },
      averageMonthlySpending: { RMB: 0, HKD: -88 },
      formulaLabel: '平均月开销 = sum(monthly_total) / months_with_records',
    })
    deleteRepaymentMock.mockRejectedValue(new Error('failed to delete repayment'))

    const wrapper = mount(HistoryPage)
    await flushPromises()

    await wrapper.get('.ghost-danger').trigger('click')
    const dialogButtons = wrapper.findAll('.dialog-actions button')
    await dialogButtons[1].trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('确认删除记录')
    expect(wrapper.text()).toContain('failed to delete repayment')
    expect(deleteRepaymentMock).toHaveBeenCalledWith('r-2')
  })
})
