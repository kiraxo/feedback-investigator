import type {
  AgentRun,
  FeedbackInput,
  Health,
} from './types'

interface GraphQLError {
  message: string
}

interface GraphQLResponse<T> {
  data?: T
  errors?: GraphQLError[]
}

async function graphqlRequest<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const response = await fetch('/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query,
      variables,
    }),
  })

  if (!response.ok) {
    throw new Error(
      `GraphQL request failed with HTTP ${response.status}`,
    )
  }

  const payload =
    (await response.json()) as GraphQLResponse<T>

  if (payload.errors && payload.errors.length > 0) {
    throw new Error(
      payload.errors.map((error) => error.message).join('; '),
    )
  }

  if (!payload.data) {
    throw new Error('GraphQL response did not contain data')
  }

  return payload.data
}

const healthQuery = `
  query Health {
    health {
      status
      service
      version
    }
  }
`

const runInvestigationMutation = `
  mutation RunInvestigation(
    $input: RunInvestigationInput!
  ) {
    runInvestigation(input: $input) {
      id
      status
      mode
      provider
      feedbackCount
      startedAt
      completedAt
      errorMessage
      results {
        feedbackId
        feedbackText
        product
        source
        occurredAt
        period
        classification {
          topic
          severity
          evidence
          requiresInvestigation
          evidenceValid
        }
      }
      tasks {
        id
        taskKey
        title
        product
        topic
        priority
        status
        reason
        urgent
        urgentReason
        createdAt
        metrics {
          currentIssueMessages
          currentProductMessages
          currentSharePercentage
          previousIssueMessages
          previousProductMessages
          previousSharePercentage
          changePercentagePoints
          comparisonStatus
        }
        supportingEvidence {
          feedbackId
          text
        }
        recommendedAction
      }
    }
  }
`

export async function getHealth(): Promise<Health> {
  const data = await graphqlRequest<{
    health: Health
  }>(healthQuery)

  return data.health
}

export async function runInvestigation(
  feedback: FeedbackInput[],
): Promise<AgentRun> {
  const data = await graphqlRequest<{
    runInvestigation: AgentRun
  }>(
    runInvestigationMutation,
    {
      input: {
        mode: 'DEMO',
        provider: 'DEMO',
        feedback,
      },
    },
  )

  return data.runInvestigation
}