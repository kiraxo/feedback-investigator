import {
  useEffect,
  useMemo,
  useState,
} from 'react'
import './App.css'
import {
  getHealth,
  runInvestigation,
} from './api'
import { sampleFeedback } from './sampleData'
import type {
  AgentRun,
  FeedbackInput,
  Health,
  InvestigationTask,
} from './types'

function formatLabel(value: string): string {
  return value
    .toLowerCase()
    .split('_')
    .map(
      (part) =>
        part.charAt(0).toUpperCase() + part.slice(1),
    )
    .join(' ')
}

function formatTime(value: string | null): string {
  if (!value) {
    return 'Not completed'
  }

  return new Date(value).toLocaleString()
}

function formatChange(value: number | null): string {
  if (value === null) {
    return 'No baseline'
  }

  if (value > 0) {
    return `+${value.toFixed(1)} pp`
  }

  return `${value.toFixed(1)} pp`
}

function TaskCard({
  task,
}: {
  task: InvestigationTask
}) {
  return (
    <article className="task-card">
      <div className="task-card__header">
        <div>
          <div className="task-card__eyebrow">
            {formatLabel(task.topic)}
          </div>

          <h3>{task.title}</h3>
        </div>

        <div className="badge-group">
          <span
            className={`badge badge--${task.priority.toLowerCase()}`}
          >
            {task.priority}
          </span>

          {task.urgent && (
            <span className="badge badge--urgent">
              Urgent
            </span>
          )}
        </div>
      </div>

      <p className="task-reason">{task.reason}</p>

      <div className="metric-grid">
        <div className="metric">
          <span>Current reports</span>
          <strong>
            {task.metrics.currentIssueMessages}
          </strong>
        </div>

        <div className="metric">
          <span>Current share</span>
          <strong>
            {task.metrics.currentSharePercentage.toFixed(1)}%
          </strong>
        </div>

        <div className="metric">
          <span>Previous share</span>
          <strong>
            {task.metrics.previousSharePercentage === null
              ? 'No baseline'
              : `${task.metrics.previousSharePercentage.toFixed(1)}%`}
          </strong>
        </div>

        <div className="metric">
          <span>Change</span>
          <strong>
            {formatChange(
              task.metrics.changePercentagePoints,
            )}
          </strong>
        </div>
      </div>

      <div className="task-section">
        <h4>Supporting evidence</h4>

        <ul className="evidence-list">
          {task.supportingEvidence.map((evidence) => (
            <li key={evidence.feedbackId}>
              <span>{evidence.feedbackId}</span>
              <p>{evidence.text}</p>
            </li>
          ))}
        </ul>
      </div>

      <div className="recommendation">
        <strong>Recommended action</strong>
        <p>{task.recommendedAction}</p>
      </div>
    </article>
  )
}

function App() {
  const [health, setHealth] =
    useState<Health | null>(null)
  const [healthError, setHealthError] =
    useState(false)
  const [inputText, setInputText] = useState(
    JSON.stringify(sampleFeedback, null, 2),
  )
  const [run, setRun] = useState<AgentRun | null>(
    null,
  )
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(
    null,
  )

  useEffect(() => {
    getHealth()
      .then((result) => {
        setHealth(result)
        setHealthError(false)
      })
      .catch(() => {
        setHealth(null)
        setHealthError(true)
      })
  }, [])

  const inputCount = useMemo(() => {
    try {
      const parsed = JSON.parse(inputText)
      return Array.isArray(parsed) ? parsed.length : 0
    } catch {
      return 0
    }
  }, [inputText])

  const issueCount =
    run?.results.filter(
      (result) =>
        result.classification.requiresInvestigation,
    ).length ?? 0

  const urgentTaskCount =
    run?.tasks.filter((task) => task.urgent).length ??
    0

  async function handleRun(): Promise<void> {
    setError(null)
    setRunning(true)

    try {
      const parsed = JSON.parse(
        inputText,
      ) as FeedbackInput[]

      if (!Array.isArray(parsed) || parsed.length === 0) {
        throw new Error(
          'Input must be a non-empty JSON array.',
        )
      }

      const result = await runInvestigation(parsed)
      setRun(result)
    } catch (caughtError) {
      setRun(null)

      setError(
        caughtError instanceof Error
          ? caughtError.message
          : 'The investigation failed.',
      )
    } finally {
      setRunning(false)
    }
  }

  function resetSample(): void {
    setInputText(
      JSON.stringify(sampleFeedback, null, 2),
    )
    setRun(null)
    setError(null)
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-mark">FI</div>

          <div>
            <strong>Feedback Investigator</strong>
            <span>Evidence-based agent platform</span>
          </div>
        </div>

        <div
          className={`service-status ${
            healthError
              ? 'service-status--offline'
              : ''
          }`}
        >
          <span className="status-dot" />

          {health
            ? `API ${health.status} · v${health.version}`
            : healthError
              ? 'API unavailable'
              : 'Checking API'}
        </div>
      </header>

      <main>
        <section className="hero">
          <div>
            <span className="hero__label">
              Go · GraphQL · PostgreSQL · TypeScript
            </span>

            <h1>
              Turn customer feedback into
              investigation-ready work.
            </h1>

            <p>
              Classify feedback, validate exact evidence,
              detect recurring issues, compare periods,
              prioritize tasks, and preserve every run.
            </p>
          </div>

          <div className="hero__note">
            <span>Demo mode</span>
            <strong>No API key required</strong>
            <p>
              The deterministic provider makes the complete
              workflow immediately testable.
            </p>
          </div>
        </section>

        <section className="summary-grid">
          <article>
            <span>Input records</span>
            <strong>{inputCount}</strong>
          </article>

          <article>
            <span>Issues detected</span>
            <strong>{issueCount}</strong>
          </article>

          <article>
            <span>Tasks created</span>
            <strong>{run?.tasks.length ?? 0}</strong>
          </article>

          <article>
            <span>Urgent tasks</span>
            <strong>{urgentTaskCount}</strong>
          </article>
        </section>

        <section className="workspace">
          <article className="panel input-panel">
            <div className="panel__header">
              <div>
                <span className="section-label">
                  Input
                </span>
                <h2>Feedback dataset</h2>
              </div>

              <button
                className="button button--secondary"
                onClick={resetSample}
                type="button"
              >
                Reset sample
              </button>
            </div>

            <p className="panel__description">
              Edit the JSON or run the included current and
              previous-period sample.
            </p>

            <textarea
              aria-label="Feedback dataset JSON"
              className="json-editor"
              onChange={(event) =>
                setInputText(event.target.value)
              }
              spellCheck={false}
              value={inputText}
            />

            {error && (
              <div className="error-message">
                <strong>Investigation failed</strong>
                <span>{error}</span>
              </div>
            )}

            <button
              className="button button--primary"
              disabled={running || healthError}
              onClick={handleRun}
              type="button"
            >
              {running
                ? 'Running investigation…'
                : 'Run investigation'}
            </button>
          </article>

          <article className="panel results-panel">
            <div className="panel__header">
              <div>
                <span className="section-label">
                  Output
                </span>
                <h2>Agent run</h2>
              </div>

              {run &&run && (
                <span className="badge badge--completed">
                  {run.status}
                </span>
              )}
            </div>

            {!run ? (
              <div className="empty-state">
                <div className="empty-state__mark">
                  01
                </div>
                <h3>Ready to investigate</h3>
                <p>
                  Run the sample dataset to generate
                  classifications and prioritized tasks.
                </p>
              </div>
            ) : (
              <>
                <div className="run-metadata">
                  <div>
                    <span>Run ID</span>
                    <code>{run.id}</code>
                  </div>

                  <div>
                    <span>Provider</span>
                    <strong>{run.provider}</strong>
                  </div>

                  <div>
                    <span>Completed</span>
                    <strong>
                      {formatTime(run.completedAt)}
                    </strong>
                  </div>
                </div>

                <div className="classification-list">
                  <h3>Classifications</h3>

                  {run.results.map((result) => (
                    <div
                      className="classification-row"
                      key={result.feedbackId}
                    >
                      <div>
                        <strong>
                          {result.feedbackId}
                        </strong>
                        <span>{result.product}</span>
                      </div>

                      <span className="topic-pill">
                        {formatLabel(
                          result.classification.topic,
                        )}
                      </span>

                      <span
                        className={`severity severity--${result.classification.severity}`}
                      >
                        Severity{' '}
                        {result.classification.severity}
                      </span>

                      <span
                        className={
                          result.classification
                            .evidenceValid
                            ? 'validation validation--valid'
                            : 'validation'
                        }
                      >
                        {result.classification.evidenceValid
                          ? 'Evidence valid'
                          : 'Evidence invalid'}
                      </span>
                    </div>
                  ))}
                </div>
              </>
            )}
          </article>
        </section>

        {run && (
          <section className="tasks-section">
            <div className="section-heading">
              <div>
                <span className="section-label">
                  Investigation queue
                </span>
                <h2>Prioritized tasks</h2>
              </div>

              <p>
                {run.tasks.length} recurring issue
                {run.tasks.length === 1 ? '' : 's'} found
              </p>
            </div>

            {run.tasks.length === 0 ? (
              <div className="no-tasks">
                No recurring issue reached the task
                threshold.
              </div>
            ) : (
              <div className="task-list">
                {run.tasks.map((task) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                  />
                ))}
              </div>
            )}
          </section>
        )}
      </main>

      <footer>
        <span>
          Feedback Investigator · Portfolio demonstration
        </span>
        <span>
          Deterministic demo + persistent PostgreSQL runs
        </span>
      </footer>
    </div>
  )
}

export default App