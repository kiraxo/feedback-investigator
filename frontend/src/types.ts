export type FeedbackPeriod = 'CURRENT' | 'PREVIOUS'

export type Topic =
  | 'DELIVERY_DELAY'
  | 'DAMAGED_PACKAGING'
  | 'PRINT_QUALITY'
  | 'ADHESION'
  | 'ARTWORK_PROCESS'
  | 'OTHER'
  | 'POSITIVE'
  | 'NEUTRAL'

export type Priority =
  | 'LOW'
  | 'MEDIUM'
  | 'HIGH'
  | 'CRITICAL'

export interface FeedbackInput {
  feedbackId: string
  feedbackText: string
  product: string
  source: string
  occurredAt: string
  period: FeedbackPeriod
}

export interface Classification {
  topic: Topic
  severity: number
  evidence: string
  requiresInvestigation: boolean
  evidenceValid: boolean
}

export interface FeedbackResult extends FeedbackInput {
  classification: Classification
}

export interface SupportingEvidence {
  feedbackId: string
  text: string
}

export interface InvestigationMetrics {
  currentIssueMessages: number
  currentProductMessages: number
  currentSharePercentage: number
  previousIssueMessages: number
  previousProductMessages: number
  previousSharePercentage: number | null
  changePercentagePoints: number | null
  comparisonStatus: string
}

export interface InvestigationTask {
  id: string
  taskKey: string
  title: string
  product: string
  topic: Topic
  priority: Priority
  status: string
  reason: string
  metrics: InvestigationMetrics
  supportingEvidence: SupportingEvidence[]
  recommendedAction: string
  urgent: boolean
  urgentReason: string
  createdAt: string
}

export interface AgentRun {
  id: string
  status: string
  mode: string
  provider: string
  feedbackCount: number
  results: FeedbackResult[]
  tasks: InvestigationTask[]
  startedAt: string
  completedAt: string | null
  errorMessage: string | null
}

export interface Health {
  status: string
  service: string
  version: string
}