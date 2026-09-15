# Feedback Investigator

An evidence-based customer feedback investigation workflow built with n8n and GPT-5 mini.

The workflow classifies customer feedback, validates the exact supporting evidence, detects recurring product and delivery problems, compares two periods, creates prioritized investigation tasks, identifies urgent cases, and exports the result as CSV.

## Workflow

```text
Sample Feedback
→ GPT-5 mini Classification
→ Structured Output Parser
→ Evidence Validation
→ Aggregate Feedback
→ Compare Periods
→ Create Investigation Tasks
→ Urgent Task Detection
→ Slack / Gmail Notification
→ CSV Report
