import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  createRepayment,
  fetchAuthMe,
  fetchHistory,
  fetchHomeSummary,
  fetchStats,
  login,
  logout,
  sha256Hex,
  UnauthorizedError,
} from './client'

const fetchMock = vi.fn()
vi.stubGlobal('fetch', fetchMock)

afterEach(() => {
  fetchMock.mockReset()
})

function mockJsonResponse(payload: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => payload,
  } as Response
}

describe('api client contract mapping', () => {
  it('uses backend payload keys for create repayment', async () => {
    fetchMock.mockResolvedValue(
      mockJsonResponse({
        data: {
          id: 3,
          cardName: 'BOCHK Visa',
          currency: 'RMB',
          amount: '-12.30',
          repaymentAt: '2026-07-02T10:45:00+08:00',
        },
      }),
    )

    const result = await createRepayment({
      card: 'BOCHK Visa',
      currency: 'RMB',
      amount: -12.3,
      repaymentTime: '2026-07-02T10:45',
    })

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/repayments',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    )
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string)
    expect(body).toEqual({
      cardName: 'BOCHK Visa',
      currency: 'RMB',
      amount: '-12.30',
      repaymentAt: '2026-07-02T10:45',
    })
    expect(result).toMatchObject({
      id: '3',
      card: 'BOCHK Visa',
      amount: -12.3,
      repaymentTime: '2026-07-02T10:45:00+08:00',
    })
  })

  it('uses stats/history endpoints and cardName filter', async () => {
    fetchMock
      .mockResolvedValueOnce(
        mockJsonResponse({
          month: '2026-07',
          totals: {
            RMB: '10.00',
            HKD: '-5.50',
          },
        }),
      )
      .mockResolvedValueOnce(
        mockJsonResponse({
          months: [
            {
              month: '2026-07',
              records: [
                {
                  id: 10,
                  cardName: 'HSBC Pulse',
                  currency: 'HKD',
                  amount: '-5.50',
                  repaymentAt: '2026-07-01T09:10:00+08:00',
                },
              ],
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        mockJsonResponse({
          currencies: [
            {
              currency: 'RMB',
              monthlyTotals: [{ month: '2026-07', total: '20.00' }],
              averageMonthlyRepayment: '20.00',
            },
            {
              currency: 'HKD',
              monthlyTotals: [{ month: '2026-07', total: '-11.00' }],
              averageMonthlyRepayment: '-11.00',
            },
          ],
        }),
      )

    const home = await fetchHomeSummary()
    const history = await fetchHistory({ card: 'HSBC Pulse', currency: 'HKD' })
    const stats = await fetchStats({ card: 'HSBC Pulse', currency: 'HKD' })

    expect(fetchMock.mock.calls[0][0]).toBe('/api/stats/current-month')
    expect(fetchMock.mock.calls[1][0]).toBe('/api/repayments/history?cardName=HSBC+Pulse&currency=HKD')
    expect(fetchMock.mock.calls[2][0]).toBe('/api/stats/monthly?cardName=HSBC+Pulse&currency=HKD')

    expect(home).toEqual({
      currentMonth: '2026-07',
      monthlyTotals: { RMB: 10, HKD: -5.5 },
    })
    expect(history[0].records[0]).toMatchObject({
      id: '10',
      card: 'HSBC Pulse',
      amount: -5.5,
    })
    expect(stats).toMatchObject({
      monthlyTotals: { RMB: 20, HKD: -11 },
      averageMonthlySpending: { RMB: 20, HKD: -11 },
    })
    expect(fetchMock.mock.calls[0][1]).toEqual(
      expect.objectContaining({
        credentials: 'include',
      }),
    )
  })

  it('hashes password before login and supports auth helpers', async () => {
    const expectedHash = await sha256Hex('change-me')
    fetchMock
      .mockResolvedValueOnce(mockJsonResponse({ status: 'ok' }))
      .mockResolvedValueOnce(mockJsonResponse({ status: 'ok' }))
      .mockResolvedValueOnce(mockJsonResponse({ error: 'unauthorized' }, 401))
      .mockResolvedValueOnce(mockJsonResponse({ status: 'ok' }))

    await login('change-me')
    expect(fetchMock.mock.calls[0][0]).toBe('/api/auth/login')
    expect(fetchMock.mock.calls[0][1]).toEqual(
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    )
    expect(JSON.parse(fetchMock.mock.calls[0][1].body as string)).toEqual({
      passwordHash: expectedHash,
    })

    await expect(fetchAuthMe()).resolves.toBe(true)
    await expect(fetchAuthMe()).resolves.toBe(false)

    await logout()
    expect(fetchMock.mock.calls[3][0]).toBe('/api/auth/logout')
  })

  it('throws UnauthorizedError on 401 for business APIs', async () => {
    fetchMock.mockResolvedValue(mockJsonResponse({ error: 'unauthorized' }, 401))
    await expect(fetchHomeSummary()).rejects.toBeInstanceOf(UnauthorizedError)
  })
})
