import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import EntryFlow from './EntryFlow.vue'

const createRepaymentMock = vi.fn()

vi.mock('../api/client', () => ({
  createRepayment: (payload: unknown) => createRepaymentMock(payload),
}))

describe('EntryFlow', () => {
  it('validates fields, submits, and shows completion', async () => {
    createRepaymentMock.mockResolvedValue({
      id: '1',
      card: 'HSBC Pulse',
      currency: 'HKD',
      amount: -120.5,
      repaymentTime: '2026-07-01T12:30:00',
    })

    const wrapper = mount(EntryFlow)

    await wrapper.get('button.primary').trigger('click')
    expect(wrapper.text()).toContain('必须选择银行卡')

    const cardButton = wrapper
      .findAll('.card-option')
      .find((node) => node.text().includes('HSBC Pulse'))
    expect(cardButton).toBeTruthy()
    await cardButton!.trigger('click')
    await wrapper.get('button.primary').trigger('click')
    expect(wrapper.text()).toContain('步骤 2 / 3')

    await wrapper.get('select').setValue('HKD')
    await wrapper.get('input').setValue('-120.50')
    await wrapper.get('button.primary').trigger('click')
    expect(wrapper.text()).toContain('步骤 3 / 3')

    await wrapper.get('button.primary').trigger('click')
    await flushPromises()

    expect(createRepaymentMock).toHaveBeenCalledTimes(1)
    const payload = createRepaymentMock.mock.calls[0][0] as { repaymentTime: string }
    expect(payload.repaymentTime).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/)
    expect(wrapper.text()).toContain('记录完成')
    expect(wrapper.text()).toContain('120.50 CR')
    expect(wrapper.text()).toContain('再记一笔')
    expect(wrapper.text()).toContain('查看历史')
  })
})
