import type { BaseTaskEnvelope, DocumentTaskPayload, TaskResult } from '@social-platform/contracts';

console.log('[task-worker] Worker initialized. Ready to process scoped document/media tasks.');

export async function processDocumentTask(task: BaseTaskEnvelope<DocumentTaskPayload>): Promise<TaskResult> {
  console.log(`[task-worker] Processing document task ID: ${task.id} for section: ${task.section}`);
  return {
    taskId: task.id,
    success: true,
    data: {
      status: 'processed',
      documentId: task.payload.documentId,
    },
    completedAt: new Date().toISOString(),
  };
}
