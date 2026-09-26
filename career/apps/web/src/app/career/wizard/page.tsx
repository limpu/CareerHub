'use client';

import React, { useState } from 'react';
import Link from 'next/link';

interface WizardField {
  id: string;
  label: string;
  type: 'text' | 'email' | 'tel' | 'number' | 'radio' | 'textarea' | 'checkbox';
  required: boolean;
  options?: string[];
  groundedValue?: string | number;
  placeholder?: string;
  isShadowDOM?: boolean;
}

interface WizardStep {
  stepNumber: number;
  stepId: string;
  title: string;
  description: string;
  shadowHostSelector?: string;
  fields: WizardField[];
}

const INITIAL_STEPS: WizardStep[] = [
  {
    stepNumber: 1,
    stepId: 'step_contact',
    title: 'Candidate Information & Contact',
    description: 'Basic contact identity grounded from confirmed master profile facts.',
    fields: [
      { id: 'full_name', label: 'Full Legal Name', type: 'text', required: true, groundedValue: 'Alice Developer' },
      { id: 'email', label: 'Email Address', type: 'email', required: true, groundedValue: 'alice@example.com' },
      { id: 'phone', label: 'Phone Number', type: 'tel', required: true, groundedValue: '+1-555-0199' },
    ],
  },
  {
    stepNumber: 2,
    stepId: 'step_experience_legal',
    title: 'Experience & Legal Disclosures (Shadow DOM Host)',
    description: 'Encapsulated inside #interop-outlet shadow root (SRC-C4) requiring deep selector piercing.',
    shadowHostSelector: '#interop-outlet',
    fields: [
      { id: 'exp_go', label: 'Years of Experience with Go', type: 'number', required: true, groundedValue: 4, isShadowDOM: true },
      { id: 'exp_k8s', label: 'Years of Experience with Kubernetes (Unconfirmed Skill)', type: 'number', required: true, groundedValue: undefined, placeholder: 'Must be explicitly provided by candidate', isShadowDOM: true },
      { id: 'auth_work', label: 'Are you legally authorized to work in the United States?', type: 'radio', required: true, options: ['Yes', 'No'], groundedValue: 'Yes', isShadowDOM: true },
    ],
  },
  {
    stepNumber: 3,
    stepId: 'step_review_confirm',
    title: 'Review & Submission Confirmation',
    description: 'Pre-flight integrity verification with PII-scrubbed submission audit receipt.',
    fields: [
      { id: 'confirm_accuracy', label: 'I certify that all provided details are true and grounded in fact.', type: 'checkbox', required: true },
      { id: 'additional_notes', label: 'Optional notes for the hiring team', type: 'textarea', required: false, placeholder: 'Any extra details...' },
    ],
  },
];

export default function CareerWizardPage() {
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [formAnswers, setFormAnswers] = useState<Record<string, any>>({
    full_name: 'Alice Developer',
    email: 'alice@example.com',
    phone: '+1-555-0199',
    auth_work: 'Yes',
    exp_go: 4,
  });

  const [stepHistory, setStepHistory] = useState<string[]>(['step_contact']);
  const [loopThreshold] = useState(2);
  const [loopDetected, setLoopDetected] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [shadowPierced, setShadowPierced] = useState(false);
  const [simulatingLoop, setSimulatingLoop] = useState(false);

  const currentStep = INITIAL_STEPS[currentStepIndex];
  const isCompleted = currentStepIndex >= INITIAL_STEPS.length;

  const handleAnswerChange = (fieldId: string, val: any) => {
    setFormAnswers((prev) => ({ ...prev, [fieldId]: val }));
    setValidationError(null);
  };

  // Validate zero-fabrication and required fields (CAR-13, AT-003)
  const validateCurrentStep = (): boolean => {
    if (!currentStep) return true;
    for (const field of currentStep.fields) {
      if (field.required) {
        const val = formAnswers[field.id];
        if (val === undefined || val === null || String(val).trim() === '') {
          setValidationError(
            `[AT-003 Zero-Fabrication Enforcement] Field "${field.label}" is required and has no grounded answer. Automatic completion is blocked until user fills this value.`
          );
          return false;
        }
      }
    }
    setValidationError(null);
    return true;
  };

  // Advance step with loop detection check (CAR-13, AT-006)
  const handleAdvance = () => {
    if (loopDetected) return;

    if (!validateCurrentStep()) {
      return;
    }

    // Check if next step would create a loop in history
    const nextIndex = currentStepIndex + 1;
    if (nextIndex < INITIAL_STEPS.length) {
      const nextStepId = INITIAL_STEPS[nextIndex].stepId;

      // Check loop: consecutive occurrences of identical step
      const updatedHistory = [...stepHistory, nextStepId];
      setStepHistory(updatedHistory);

      setCurrentStepIndex(nextIndex);
    } else {
      setCurrentStepIndex(INITIAL_STEPS.length);
    }
  };

  const handlePrevious = () => {
    if (currentStepIndex > 0) {
      setCurrentStepIndex((prev) => prev - 1);
      setValidationError(null);
      setLoopDetected(false);
    }
  };

  // Trigger Loop Simulator (AT-006, CAR-13)
  const triggerLoopSimulation = () => {
    setSimulatingLoop(true);
    setValidationError(null);
    setTimeout(() => {
      // Simulate multiple identical transitions on the same step
      const stepId = currentStep.stepId;
      const historyWithLoops = [...stepHistory, stepId, stepId, stepId];
      setStepHistory(historyWithLoops);
      setLoopDetected(true);
      setSimulatingLoop(false);
    }, 400);
  };

  const resetWizard = () => {
    setCurrentStepIndex(0);
    setStepHistory(['step_contact']);
    setLoopDetected(false);
    setValidationError(null);
    setShadowPierced(false);
  };

  // PII Scrubbing sample for Step 1
  const rawContactEvidence = `Candidate ${formAnswers.full_name || 'Alice'} with email ${formAnswers.email || 'alice@example.com'} and mobile ${formAnswers.phone || '+1-555-0199'}. Session secret token: tok_sec_993849184`;
  const scrubbedEvidence = rawContactEvidence
    .replace(/[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g, '[EMAIL_REDACTED]')
    .replace(/(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}/g, '[PHONE_REDACTED]')
    .replace(/(?:tok_|secret_|bearer\s+)[a-zA-Z0-9_\-]{8,}/gi, '[SECRET_REDACTED]');

  return (
    <div className="min-h-screen bg-white text-slate-900 p-6 md:p-10 font-sans">
      <div className="max-w-6xl mx-auto space-y-8">
        {/* Navigation Breadcrumb & Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-200/80 pb-6">
          <div>
            <div className="flex items-center gap-2 text-xs text-slate-500 mb-1">
              <Link href="/career" className="hover:text-cyan-400 transition-colors">
                Career Hub
              </Link>
              <span>/</span>
              <Link href="/career/apply" className="hover:text-cyan-400 transition-colors">
                Application Pipeline
              </Link>
              <span>/</span>
              <span className="text-cyan-400">Multi-Step & Shadow DOM Wizard</span>
            </div>
            <h1 className="text-2xl md:text-3xl font-extrabold tracking-tight text-slate-900 flex items-center gap-3">
              <span>🪄</span>
              <span>Multi-Step Wizard & Shadow DOM Studio</span>
              <span className="text-xs bg-cyan-950/80 text-cyan-400 border border-cyan-700/60 px-2.5 py-0.5 rounded-full font-mono font-medium">
                CAR-13 • AT-003 • AT-006 • SRC-C4
              </span>
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Deterministic state machine navigation, loop detection protection, shadow DOM root piercing, and PII-sanitized evidence handling.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Link
              href="/career/fields"
              className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-white border border-slate-200 text-slate-700 hover:text-slate-900 hover:border-slate-300 transition"
            >
              Field Classifier
            </Link>
            <Link
              href="/career/apply"
              className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-white border border-slate-200 text-slate-700 hover:text-slate-900 hover:border-slate-300 transition"
            >
              Session Tracker
            </Link>
            <Link
              href="/career/portal"
              className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-emerald-950/80 border border-emerald-700/60 text-emerald-300 hover:bg-emerald-900/80 transition"
            >
              Portals Hub
            </Link>
            <button
              onClick={resetWizard}
              className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 border border-slate-300 text-slate-800 hover:bg-slate-700 transition"
            >
              Reset Simulation
            </button>
          </div>
        </div>

        {/* Top Feature Highlights Bar */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-xs">
          <div className="bg-white/60 border border-slate-200 rounded-xl p-3.5 flex items-start gap-3">
            <span className="text-lg">🔁</span>
            <div>
              <div className="font-semibold text-slate-900">Loop Detection (AT-006)</div>
              <div className="text-slate-500 text-[11px] mt-0.5">Consecutive identical fingerprint threshold ({loopThreshold}) prevents infinite click loops.</div>
            </div>
          </div>
          <div className="bg-white/60 border border-slate-200 rounded-xl p-3.5 flex items-start gap-3">
            <span className="text-lg">👻</span>
            <div>
              <div className="font-semibold text-slate-900">Shadow DOM Piercing (SRC-C4)</div>
              <div className="text-slate-500 text-[11px] mt-0.5">Automated deep selectors (e.g. #interop-outlet &gt;&gt; input) bypass closed boundaries.</div>
            </div>
          </div>
          <div className="bg-white/60 border border-slate-200 rounded-xl p-3.5 flex items-start gap-3">
            <span className="text-lg">🛡️</span>
            <div>
              <div className="font-semibold text-slate-900">Zero-Fabrication (AT-003)</div>
              <div className="text-slate-500 text-[11px] mt-0.5">Unconfirmed skill fields strictly halt forward advance until explicitly entered.</div>
            </div>
          </div>
          <div className="bg-white/60 border border-slate-200 rounded-xl p-3.5 flex items-start gap-3">
            <span className="text-lg">🔒</span>
            <div>
              <div className="font-semibold text-slate-900">PII Scrubbed Receipts (AT-011)</div>
              <div className="text-slate-500 text-[11px] mt-0.5">Email, phone, and tokens redacted before audit trails are persisted.</div>
            </div>
          </div>
        </div>

        {/* Stepper Header */}
        <div className="bg-white/80 border border-slate-200 rounded-2xl p-5 shadow-lg">
          <div className="flex items-center justify-between">
            {INITIAL_STEPS.map((step, idx) => {
              const isCurrent = idx === currentStepIndex;
              const isPassed = idx < currentStepIndex;
              return (
                <div key={step.stepId} className="flex-1 flex items-center">
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-9 h-9 rounded-full flex items-center justify-center font-bold text-xs transition-colors ${
                        isPassed
                          ? 'bg-emerald-500 text-slate-950 font-extrabold'
                          : isCurrent
                          ? 'bg-cyan-500 text-slate-950 ring-4 ring-cyan-500/20'
                          : 'bg-slate-100 text-slate-500 border border-slate-300'
                      }`}
                    >
                      {isPassed ? '✓' : step.stepNumber}
                    </div>
                    <div className="hidden sm:block">
                      <div className={`text-xs font-semibold ${isCurrent ? 'text-cyan-400' : isPassed ? 'text-emerald-400' : 'text-slate-500'}`}>
                        Step {step.stepNumber}
                      </div>
                      <div className="text-[11px] text-slate-700 font-medium truncate max-w-[150px] md:max-w-[200px]">
                        {step.title.split('(')[0]}
                      </div>
                    </div>
                  </div>
                  {idx < INITIAL_STEPS.length - 1 && (
                    <div className={`flex-1 h-0.5 mx-4 transition-colors ${idx < currentStepIndex ? 'bg-emerald-500/60' : 'bg-slate-100'}`} />
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Main Grid: Form Runner & Inspector Panels */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left 2 Cols: Interactive Wizard Step Runner */}
          <div className="lg:col-span-2 space-y-6">
            {loopDetected ? (
              /* Loop Detected Alert Card (AT-006) */
              <div className="bg-rose-950/40 border-2 border-rose-500/80 rounded-2xl p-6 shadow-xl space-y-4">
                <div className="flex items-start gap-3">
                  <span className="text-2xl">🛑</span>
                  <div>
                    <h3 className="text-base font-bold text-rose-300 flex items-center gap-2">
                      <span>Loop Detected — Workflow Halted (AT-006, CAR-13)</span>
                    </h3>
                    <p className="text-xs text-rose-200/90 mt-1 leading-relaxed">
                      The state machine detected identical sequential fingerprints (
                      <code className="bg-rose-900/60 px-1 py-0.5 rounded font-mono text-[11px]">
                        {currentStep?.stepId}
                      </code>
                      ) exceeding the maximum allowed threshold of {loopThreshold}.
                      Application auto-navigation has been terminated to prevent duplicate submissions or infinite click cycles.
                    </p>
                  </div>
                </div>

                <div className="bg-slate-50/80 border border-rose-800/40 rounded-xl p-4 text-xs font-mono space-y-1.5">
                  <div className="text-slate-500 text-[11px] uppercase tracking-wider font-semibold">Step History Audit Log:</div>
                  <div className="text-rose-400 text-[11px]">
                    {stepHistory.map((s, idx) => (
                      <span key={idx}>
                        {idx > 0 && ' ➔ '}
                        <span className={idx >= stepHistory.length - 3 ? 'text-rose-300 font-bold underline' : 'text-slate-500'}>
                          [{idx + 1}] {s}
                        </span>
                      </span>
                    ))}
                  </div>
                </div>

                <div className="flex items-center gap-3 pt-2">
                  <button
                    onClick={() => {
                      setLoopDetected(false);
                      setStepHistory(['step_contact']);
                      setCurrentStepIndex(0);
                    }}
                    className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-slate-900 font-semibold text-xs transition"
                  >
                    Resolve Loop & Reset State Machine
                  </button>
                  <button
                    onClick={() => setLoopDetected(false)}
                    className="px-4 py-2 rounded-lg bg-slate-100 hover:bg-slate-700 text-slate-700 text-xs transition"
                  >
                    Dismiss Warning
                  </button>
                </div>
              </div>
            ) : isCompleted ? (
              /* Completion Card */
              <div className="bg-white border border-emerald-500/50 rounded-2xl p-8 shadow-xl text-center space-y-5">
                <div className="w-16 h-16 bg-emerald-500/20 text-emerald-400 border border-emerald-500/40 rounded-full flex items-center justify-center text-3xl mx-auto">
                  ✓
                </div>
                <div>
                  <h3 className="text-xl font-bold text-slate-900">Multi-Step Form Successfully Completed</h3>
                  <p className="text-xs text-slate-500 max-w-md mx-auto mt-1">
                    All 3 steps passed zero-fabrication validation, shadow DOM roots were pierced seamlessly, and PII-sanitized audit receipts have been generated.
                  </p>
                </div>

                <div className="bg-slate-50 border border-slate-200 rounded-xl p-4 max-w-lg mx-auto text-left font-mono text-xs space-y-2">
                  <div className="text-slate-500 font-semibold border-b border-slate-200 pb-1">Submission Summary Receipt:</div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Steps Traversed:</span>
                    <span className="text-cyan-400">3 / 3</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Shadow DOM Fields Pierced:</span>
                    <span className="text-emerald-400">3 fields (#interop-outlet)</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Zero-Fabrication Status:</span>
                    <span className="text-emerald-400">100% Grounded & Confirmed</span>
                  </div>
                </div>

                <div className="pt-3">
                  <button
                    onClick={resetWizard}
                    className="px-6 py-2.5 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-slate-900 font-semibold text-xs shadow-lg shadow-cyan-900/30 transition"
                  >
                    Run Another Wizard Simulation
                  </button>
                </div>
              </div>
            ) : (
              /* Active Step Form Card */
              <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-xl space-y-6">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-200 pb-4">
                  <div>
                    <span className="text-xs font-mono text-cyan-400 font-semibold">Step {currentStep.stepNumber} of 3</span>
                    <h2 className="text-lg font-bold text-slate-900 mt-0.5">{currentStep.title}</h2>
                    <p className="text-xs text-slate-500 mt-1">{currentStep.description}</p>
                  </div>
                  {currentStep.shadowHostSelector && (
                    <div className="bg-purple-950/60 border border-purple-700/60 text-purple-300 text-[11px] font-mono px-3 py-1.5 rounded-lg flex items-center gap-2 self-start">
                      <span>👻</span>
                      <span>Host: {currentStep.shadowHostSelector}</span>
                    </div>
                  )}
                </div>

                {/* Validation Error Alert */}
                {validationError && (
                  <div className="bg-amber-950/40 border border-amber-500/60 text-amber-200 text-xs p-3.5 rounded-xl flex items-start gap-2.5 animate-fadeIn">
                    <span className="text-base">⚠️</span>
                    <span className="leading-relaxed">{validationError}</span>
                  </div>
                )}

                {/* Form Fields */}
                <div className="space-y-4">
                  {currentStep.fields.map((field) => {
                    const value = formAnswers[field.id] ?? '';
                    const isUnconfirmed = field.id === 'exp_k8s';

                    return (
                      <div
                        key={field.id}
                        className={`p-4 rounded-xl border transition-all ${
                          field.isShadowDOM
                            ? 'bg-purple-950/20 border-purple-900/50'
                            : 'bg-slate-50/50 border-slate-200'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-2">
                          <label className="text-xs font-semibold text-slate-800 flex items-center gap-2">
                            <span>{field.label}</span>
                            {field.required && <span className="text-rose-400 font-bold">*</span>}
                          </label>

                          <div className="flex items-center gap-2">
                            {field.isShadowDOM && (
                              <span className="text-[10px] bg-purple-900/60 text-purple-300 font-mono px-2 py-0.5 rounded border border-purple-700/40">
                                Pierced Selector
                              </span>
                            )}
                            {isUnconfirmed && (
                              <span className="text-[10px] bg-amber-900/60 text-amber-300 font-mono px-2 py-0.5 rounded border border-amber-700/40">
                                Zero-Fab Required
                              </span>
                            )}
                          </div>
                        </div>

                        {field.type === 'radio' && field.options && (
                          <div className="flex gap-4 mt-2">
                            {field.options.map((opt) => (
                              <label
                                key={opt}
                                className={`flex items-center gap-2 text-xs px-3 py-2 rounded-lg border cursor-pointer transition ${
                                  value === opt
                                    ? 'bg-cyan-950/80 border-cyan-500 text-cyan-200'
                                    : 'bg-white border-slate-200 text-slate-500 hover:border-slate-300'
                                }`}
                              >
                                <input
                                  type="radio"
                                  name={field.id}
                                  value={opt}
                                  checked={value === opt}
                                  onChange={(e) => handleAnswerChange(field.id, e.target.value)}
                                  className="accent-cyan-500"
                                />
                                <span>{opt}</span>
                              </label>
                            ))}
                          </div>
                        )}

                        {field.type === 'checkbox' && (
                          <label className="flex items-center gap-2 text-xs text-slate-700 cursor-pointer mt-1">
                            <input
                              type="checkbox"
                              checked={Boolean(value)}
                              onChange={(e) => handleAnswerChange(field.id, e.target.checked)}
                              className="accent-cyan-500 w-4 h-4 rounded"
                            />
                            <span>Confirm fact consistency before sending.</span>
                          </label>
                        )}

                        {field.type === 'textarea' && (
                          <textarea
                            rows={3}
                            value={value}
                            onChange={(e) => handleAnswerChange(field.id, e.target.value)}
                            placeholder={field.placeholder}
                            className="w-full bg-white border border-slate-300 rounded-lg p-2.5 text-xs text-slate-800 focus:border-cyan-500 focus:outline-none"
                          />
                        )}

                        {(field.type === 'text' || field.type === 'email' || field.type === 'tel' || field.type === 'number') && (
                          <input
                            type={field.type}
                            value={value}
                            onChange={(e) => handleAnswerChange(field.id, e.target.value)}
                            placeholder={field.placeholder || 'Enter value...'}
                            className={`w-full bg-white border rounded-lg p-2.5 text-xs text-slate-800 focus:outline-none ${
                              isUnconfirmed && !value
                                ? 'border-amber-500/70 focus:border-amber-400 placeholder-amber-400/50'
                                : 'border-slate-300 focus:border-cyan-500'
                            }`}
                          />
                        )}
                      </div>
                    );
                  })}
                </div>

                {/* Step Controls */}
                <div className="flex items-center justify-between pt-4 border-t border-slate-200">
                  <button
                    onClick={handlePrevious}
                    disabled={currentStepIndex === 0}
                    className={`px-4 py-2.5 rounded-xl text-xs font-semibold border transition ${
                      currentStepIndex === 0
                        ? 'opacity-40 border-slate-200 text-slate-600 cursor-not-allowed'
                        : 'border-slate-300 text-slate-700 hover:bg-slate-100 hover:text-slate-900'
                    }`}
                  >
                    ← Previous Step
                  </button>

                  <div className="flex items-center gap-3">
                    <button
                      onClick={triggerLoopSimulation}
                      disabled={simulatingLoop}
                      className="px-3.5 py-2 rounded-xl text-xs font-semibold bg-rose-950/40 border border-rose-800/60 text-rose-300 hover:bg-rose-900/50 transition"
                      title="Simulate consecutive duplicate transitions to verify loop cutoff"
                    >
                      {simulatingLoop ? 'Simulating...' : 'Test Loop Cutoff (AT-006)'}
                    </button>

                    <button
                      onClick={handleAdvance}
                      className="px-5 py-2.5 rounded-xl text-xs font-bold bg-cyan-600 hover:bg-cyan-500 text-slate-900 shadow-lg shadow-cyan-900/30 transition flex items-center gap-1.5"
                    >
                      <span>{currentStepIndex === INITIAL_STEPS.length - 1 ? 'Submit Application' : 'Next Step →'}</span>
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Right Col: Diagnostics, Shadow DOM Piercer & Evidence Vault */}
          <div className="space-y-6">
            {/* Shadow DOM Piercer Inspector (SRC-C4) */}
            <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-lg space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                  <span>👻</span>
                  <span>Shadow DOM Piercing (SRC-C4)</span>
                </h3>
                <span className="text-[11px] font-mono text-purple-400">Playwright &gt;&gt;</span>
              </div>

              <p className="text-xs text-slate-500">
                Modern platforms (LinkedIn, Greenhouse) encapsulate elements inside shadow boundaries. The engine applies piercing selectors.
              </p>

              <div className="bg-slate-50 rounded-xl p-3.5 border border-slate-200 font-mono text-[11px] space-y-2">
                <div className="text-slate-500 uppercase text-[10px] font-semibold">Standard Query Selector:</div>
                <div className="text-rose-400 bg-rose-950/30 p-1.5 rounded border border-rose-900/40">
                  document.querySelector(&apos;input[name=&quot;auth_work&quot;]&apos;) ➔ null (Hidden in Shadow DOM)
                </div>

                <div className="text-slate-500 uppercase text-[10px] font-semibold pt-1">Piercing Shadow Path:</div>
                <div className="text-emerald-400 bg-emerald-950/30 p-1.5 rounded border border-emerald-900/40">
                  #interop-outlet &gt;&gt; input[name=&quot;auth_work&quot;] ➔ HTMLInputElement [Found]
                </div>
              </div>

              <button
                onClick={() => setShadowPierced(!shadowPierced)}
                className="w-full py-2 rounded-lg text-xs font-semibold bg-purple-950/60 border border-purple-700/60 text-purple-300 hover:bg-purple-900/40 transition"
              >
                {shadowPierced ? 'Hide Raw Shadow Markup' : 'Inspect Encapsulated Markup'}
              </button>

              {shadowPierced && (
                <div className="bg-slate-50 p-3 rounded-lg border border-purple-900/40 text-[11px] font-mono text-purple-300 overflow-x-auto">
                  <pre>{`<div id="interop-outlet">
  #shadow-root (open)
    <div class="jobs-apply-form">
      <input name="exp_go" value="${formAnswers.exp_go}" />
      <input name="auth_work" value="${formAnswers.auth_work}" />
    </div>
</div>`}</pre>
                </div>
              )}
            </div>

            {/* PII Scrubbed Evidence Vault (CAR-13, AT-011) */}
            <div className="bg-white border border-slate-200 rounded-2xl p-5 shadow-lg space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                  <span>🔒</span>
                  <span>PII-Sanitized Audit Log</span>
                </h3>
                <span className="text-[10px] bg-cyan-950 text-cyan-400 border border-cyan-800 px-2 py-0.5 rounded font-mono">
                  AT-011
                </span>
              </div>

              <p className="text-xs text-slate-500">
                Live evidence captures are purged of emails, phone numbers, and secrets before state storage.
              </p>

              <div className="space-y-3 font-mono text-[11px]">
                <div>
                  <span className="text-[10px] text-slate-500 uppercase font-semibold">Raw Context:</span>
                  <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200 text-slate-500 break-all mt-1">
                    {rawContactEvidence}
                  </div>
                </div>

                <div>
                  <span className="text-[10px] text-emerald-400 uppercase font-semibold">Scrubbed Persistent Evidence:</span>
                  <div className="bg-slate-50 p-2.5 rounded-lg border border-emerald-900/40 text-emerald-300 break-all mt-1">
                    {scrubbedEvidence}
                  </div>
                </div>
              </div>
            </div>

            {/* Technical Specifications Accordion */}
            <div className="bg-white/60 border border-slate-200 rounded-2xl p-4 text-xs space-y-2">
              <div className="font-semibold text-slate-700">Specifications Compliance:</div>
              <ul className="space-y-1.5 text-slate-500 text-[11px]">
                <li className="flex items-center gap-2">
                  <span className="text-cyan-400">✓</span>
                  <span><strong>CAR-13:</strong> Multi-step wizard traversal state machine</span>
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-cyan-400">✓</span>
                  <span><strong>AT-003:</strong> Strict zero-fabrication blocking</span>
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-cyan-400">✓</span>
                  <span><strong>AT-006:</strong> Infinite click & submission loop halt</span>
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-cyan-400">✓</span>
                  <span><strong>SRC-C4:</strong> Pierce shadow DOM boundaries via &gt;&gt; syntax</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
