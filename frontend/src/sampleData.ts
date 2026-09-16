import type { FeedbackInput } from './types'

export const sampleFeedback: FeedbackInput[] = [
  {
    feedbackId: 'FB-001',
    feedbackText:
      'My sticker order arrived four days late, and I needed it for an event this weekend.',
    product: 'Die-cut stickers',
    source: 'Demo review',
    occurredAt: '2026-09-10',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-002',
    feedbackText:
      'The shipping box was crushed and several stickers were bent when they arrived.',
    product: 'Die-cut stickers',
    source: 'Demo ticket',
    occurredAt: '2026-09-11',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-003',
    feedbackText:
      'The colors look excellent and the stickers arrived on time.',
    product: 'Die-cut stickers',
    source: 'Demo review',
    occurredAt: '2026-09-11',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-007',
    feedbackText:
      'The sticker package arrived two days late and missed our customer event.',
    product: 'Die-cut stickers',
    source: 'Demo ticket',
    occurredAt: '2026-09-13',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-008',
    feedbackText:
      'The sticker mailer was crushed and some stickers were damaged inside.',
    product: 'Die-cut stickers',
    source: 'Demo review',
    occurredAt: '2026-09-13',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-011',
    feedbackText:
      'The sticker order arrived one day later than promised.',
    product: 'Die-cut stickers',
    source: 'Demo review',
    occurredAt: '2026-09-14',
    period: 'CURRENT',
  },
  {
    feedbackId: 'FB-009',
    feedbackText:
      'The sticker order arrived one day late.',
    product: 'Die-cut stickers',
    source: 'Demo review',
    occurredAt: '2026-09-06',
    period: 'PREVIOUS',
  },
  {
    feedbackId: 'FB-010',
    feedbackText:
      'The sticker package was damaged at the corner when it arrived.',
    product: 'Die-cut stickers',
    source: 'Demo ticket',
    occurredAt: '2026-09-07',
    period: 'PREVIOUS',
  },
]