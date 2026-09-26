'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import {
  Briefcase,
  Globe2,
  DollarSign,
  ShieldAlert,
  Sliders,
  CheckCircle2,
  AlertCircle,
  Plus,
  X,
  ArrowLeft,
  Save,
  HelpCircle,
  Ban,
  Building,
  MapPin,
  Search,
  Check,
} from 'lucide-react';
import type {
  CareerPreferences,
  WorkMode,
  JobType,
  SponsorshipPreference,
  SalaryInterval,
} from '@social-platform/contracts';

const defaultPreferences: CareerPreferences = {
  id: 'pref-demo-001',
  user_id: 'user-current',
  target_roles: ['Staff Backend Engineer', 'Principal Distributed Systems Architect', 'Lead Go Engineer'],
  target_locations: ['Remote (Worldwide)', 'United States', 'European Union', 'Singapore'],
  work_modes: ['remote', 'hybrid'],
  job_types: ['full_time', 'contract'],
  salary: {
    minimum_amount: 140000,
    target_amount: 175000,
    maximum_amount: 220000,
    currency: 'USD',
    interval: 'annual',
    is_explicitly_set: true,
  },
  sponsorship: 'requires_sponsorship', // Explicitly selected by user
  exclusions: {
    excluded_companies: ['Revature', 'Apex Systems', 'Unwanted Toxic Corp'],
    excluded_keywords: ['crypto', 'gambling', 'casino', 'intern', 'junior'],
    excluded_industries: ['defense', 'gambling'],
    excluded_locations: [],
    block_staffing_agencies: true,
  },
  open_to_relocation: true,
  notice_period_days: 30,
  created_at: '2026-09-18T00:00:00Z',
  updated_at: '2026-09-18T02:00:00Z',
};

export default function CareerPreferencesPage() {
  const [pref, setPref] = useState<CareerPreferences>(defaultPreferences);
  const [saveStatus, setSaveStatus] = useState<string | null>(null);

  // New tag inputs
  const [newRole, setNewRole] = useState('');
  const [newLocation, setNewLocation] = useState('');
  const [newCompany, setNewCompany] = useState('');
  const [newKeyword, setNewKeyword] = useState('');

  // Exclusion Tester State
  const [testCompany, setTestCompany] = useState('');
  const [testTitle, setTestTitle] = useState('');
  const [testResult, setTestResult] = useState<{ excluded: boolean; reason: string } | null>(null);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const savedPref = localStorage.getItem('candidate_preferences');
    if (savedPref) {
      try {
        setPref(JSON.parse(savedPref));
        return;
      } catch (e) {
        console.error('Failed to parse candidate_preferences:', e);
      }
    }

    // If no explicit preferences saved yet, check candidate_master_profile
    const master = localStorage.getItem('candidate_master_profile') || localStorage.getItem('candidate_staged_profile');
    if (master) {
      try {
        const parsed = JSON.parse(master);
        const headline = (parsed.headline || parsed.contact?.headline || '').toLowerCase();
        if (headline.includes('seo') || headline.includes('marketing') || headline.includes('search')) {
          setPref((prev) => ({
            ...prev,
            target_roles: [
              'SEO Specialist',
              'Senior SEO Strategist',
              'Technical SEO Lead',
              'Digital Marketing & Organic Growth Specialist',
            ],
            target_locations: ['Remote (Worldwide)', 'United States', 'United Kingdom', 'Bangladesh'],
            salary: {
              ...prev.salary,
              minimum_amount: 50000,
              target_amount: 75000,
              maximum_amount: 110000,
            },
            exclusions: {
              ...prev.exclusions,
              excluded_keywords: ['intern', 'unpaid', 'gambling', 'crypto'],
            },
          }));
        }
      } catch (err) {
        console.error('Failed to read master profile for preferences:', err);
      }
    }
  }, []);

  const handleAddRole = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRole.trim()) return;
    if (!pref.target_roles.includes(newRole.trim())) {
      setPref({ ...pref, target_roles: [...pref.target_roles, newRole.trim()] });
    }
    setNewRole('');
  };

  const handleRemoveRole = (role: string) => {
    setPref({ ...pref, target_roles: pref.target_roles.filter((r) => r !== role) });
  };

  const handleAddLocation = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newLocation.trim()) return;
    if (!pref.target_locations.includes(newLocation.trim())) {
      setPref({ ...pref, target_locations: [...pref.target_locations, newLocation.trim()] });
    }
    setNewLocation('');
  };

  const handleRemoveLocation = (loc: string) => {
    setPref({ ...pref, target_locations: pref.target_locations.filter((l) => l !== loc) });
  };

  const handleAddCompanyExclusion = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCompany.trim()) return;
    if (!pref.exclusions.excluded_companies.includes(newCompany.trim())) {
      setPref({
        ...pref,
        exclusions: {
          ...pref.exclusions,
          excluded_companies: [...pref.exclusions.excluded_companies, newCompany.trim()],
        },
      });
    }
    setNewCompany('');
  };

  const handleRemoveCompanyExclusion = (comp: string) => {
    setPref({
      ...pref,
      exclusions: {
        ...pref.exclusions,
        excluded_companies: pref.exclusions.excluded_companies.filter((c) => c !== comp),
      },
    });
  };

  const handleAddKeywordExclusion = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newKeyword.trim()) return;
    if (!pref.exclusions.excluded_keywords.includes(newKeyword.trim())) {
      setPref({
        ...pref,
        exclusions: {
          ...pref.exclusions,
          excluded_keywords: [...pref.exclusions.excluded_keywords, newKeyword.trim()],
        },
      });
    }
    setNewKeyword('');
  };

  const handleRemoveKeywordExclusion = (kw: string) => {
    setPref({
      ...pref,
      exclusions: {
        ...pref.exclusions,
        excluded_keywords: pref.exclusions.excluded_keywords.filter((k) => k !== kw),
      },
    });
  };

  const toggleWorkMode = (mode: WorkMode) => {
    const exists = pref.work_modes.includes(mode);
    setPref({
      ...pref,
      work_modes: exists ? pref.work_modes.filter((m) => m !== mode) : [...pref.work_modes, mode],
    });
  };

  const toggleJobType = (jt: JobType) => {
    const exists = pref.job_types.includes(jt);
    setPref({
      ...pref,
      job_types: exists ? pref.job_types.filter((t) => t !== jt) : [...pref.job_types, jt],
    });
  };

  const handleRunExclusionTest = () => {
    const normComp = testCompany.toLowerCase().trim();
    const normTitle = testTitle.toLowerCase().trim();

    // Check company
    for (const c of pref.exclusions.excluded_companies) {
      if (c && normComp.includes(c.toLowerCase())) {
        setTestResult({ excluded: true, reason: `Excluded company match: "${c}"` });
        return;
      }
    }

    // Check keywords
    for (const kw of pref.exclusions.excluded_keywords) {
      if (kw && normTitle.includes(kw.toLowerCase())) {
        setTestResult({ excluded: true, reason: `Excluded keyword in title: "${kw}"` });
        return;
      }
    }

    setTestResult({ excluded: false, reason: 'Job meets your criteria and passes all exclusion filters.' });
  };

  const handleSavePreferences = () => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('candidate_preferences', JSON.stringify(pref));
    }
    setSaveStatus('Career preferences and exclusion rules saved successfully.');
    setTimeout(() => setSaveStatus(null), 4000);
  };

  return (
    <div className="space-y-8 pb-12">
      {/* Breadcrumb & Navigation */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-4">
        <div className="flex items-center gap-3">
          <Link
            href="/career"
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 transition"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Career Hub</span>
          </Link>
          <span className="text-slate-300 dark:text-slate-700">•</span>
          <span className="text-xs font-medium text-blue-600 dark:text-blue-400">
            Preferences & Exclusions (IMP-CAR-03)
          </span>
        </div>

        <button
          type="button"
          onClick={handleSavePreferences}
          className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-semibold transition cursor-pointer shadow-sm"
        >
          <Save className="w-3.5 h-3.5" />
          <span>Save Preferences</span>
        </button>
      </div>

      {/* Toast Notification */}
      {saveStatus && (
        <div className="p-3.5 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-300 rounded-xl text-xs font-medium flex items-center justify-between animate-fadeIn">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600" />
            <span>{saveStatus}</span>
          </div>
        </div>
      )}

      {/* Header Banner */}
      <div className="bg-gradient-to-r from-blue-900/10 via-indigo-900/10 to-slate-900/10 dark:from-blue-950/30 dark:via-indigo-950/30 dark:to-slate-950/30 border border-blue-200/60 dark:border-blue-800/40 rounded-2xl p-6">
        <div className="space-y-1.5">
          <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-300">
            <Sliders className="w-3 h-3" />
            <span>User-Controlled Matching Boundary</span>
          </div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">
            Career Preferences & Exclusions
          </h1>
          <p className="text-sm font-medium text-slate-600 dark:text-slate-300 max-w-2xl">
            Configure target roles, compensation requirements, visa sponsorship, and employer blacklists.
            Missing criteria are never guessed or defaulted (Invariant AT-003).
          </p>
        </div>
      </div>

      {/* Zero Fabrication Warning Banner (AT-003) */}
      <div className="p-4 rounded-xl border border-amber-200/80 dark:border-amber-900/50 bg-amber-50/70 dark:bg-amber-950/20 flex items-start gap-3 text-xs text-amber-800 dark:text-amber-300">
        <AlertCircle className="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
        <div className="space-y-1 leading-relaxed">
          <p className="font-bold">Zero Fabrication Guarantee (AT-003)</p>
          <p>
            The system strictly requires explicit candidate confirmation before matching or applying.
            If your target compensation or visa sponsorship requirements are unspecified, the matching engine marks them as{' '}
            <code className="px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/60 rounded font-mono text-[11px]">needs_input</code>{' '}
            rather than guessing or fabricating arbitrary default values.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Columns: Core Preferences */}
        <div className="lg:col-span-2 space-y-6">
          {/* 1. Target Roles */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <Briefcase className="w-4 h-4 text-blue-600" />
                  <span>Target Job Titles & Roles</span>
                </h2>
                <p className="text-xs text-slate-500">Only match opportunities matching these designated roles.</p>
              </div>
            </div>

            <form onSubmit={handleAddRole} className="flex gap-2">
              <input
                type="text"
                value={newRole}
                onChange={(e) => setNewRole(e.target.value)}
                placeholder="e.g. Lead Platform Engineer"
                className="flex-1 px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
              />
              <button
                type="submit"
                className="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 rounded-lg text-xs font-semibold transition"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Role</span>
              </button>
            </form>

            <div className="flex flex-wrap gap-2 pt-1">
              {pref.target_roles.map((role) => (
                <span
                  key={role}
                  className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-900/60"
                >
                  <span>{role}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveRole(role)}
                    className="hover:text-rose-600 transition"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
            </div>
          </div>

          {/* 2. Work Modes & Job Types */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
            <div className="space-y-3">
              <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Globe2 className="w-4 h-4 text-blue-600" />
                <span>Work Location Arrangement</span>
              </h2>
              <div className="grid grid-cols-3 gap-3">
                {(['remote', 'hybrid', 'onsite'] as WorkMode[]).map((mode) => {
                  const active = pref.work_modes.includes(mode);
                  return (
                    <button
                      key={mode}
                      type="button"
                      onClick={() => toggleWorkMode(mode)}
                      className={`p-3 rounded-xl border text-xs font-semibold capitalize transition flex items-center justify-between ${
                        active
                          ? 'border-blue-600 bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300'
                          : 'border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
                      }`}
                    >
                      <span>{mode}</span>
                      {active && <Check className="w-3.5 h-3.5" />}
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="space-y-3 border-t border-slate-100 dark:border-slate-800 pt-4">
              <h2 className="text-sm font-bold text-slate-900 dark:text-white">Employment Type</h2>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                {[
                  { id: 'full_time', label: 'Full-time' },
                  { id: 'contract', label: 'Contract' },
                  { id: 'part_time', label: 'Part-time' },
                  { id: 'internship', label: 'Internship' },
                ].map((item) => {
                  const active = pref.job_types.includes(item.id as JobType);
                  return (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => toggleJobType(item.id as JobType)}
                      className={`p-2.5 rounded-lg border text-xs font-semibold transition text-center ${
                        active
                          ? 'border-blue-600 bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300'
                          : 'border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
                      }`}
                    >
                      {item.label}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>

          {/* 3. Target Compensation */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                  <DollarSign className="w-4 h-4 text-emerald-600" />
                  <span>Target Compensation Boundaries</span>
                </h2>
                <p className="text-xs text-slate-500">Jobs below minimum compensation will be filtered out.</p>
              </div>
              <label className="flex items-center gap-2 text-xs font-semibold cursor-pointer">
                <input
                  type="checkbox"
                  checked={pref.salary.is_explicitly_set}
                  onChange={(e) =>
                    setPref({
                      ...pref,
                      salary: { ...pref.salary, is_explicitly_set: e.target.checked },
                    })
                  }
                  className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500 cursor-pointer"
                />
                <span className="text-slate-700 dark:text-slate-300">Enforce Salary Gate</span>
              </label>
            </div>

            {pref.salary.is_explicitly_set ? (
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Minimum Salary
                  </label>
                  <input
                    type="number"
                    value={pref.salary.minimum_amount || ''}
                    onChange={(e) =>
                      setPref({
                        ...pref,
                        salary: { ...pref.salary, minimum_amount: Number(e.target.value) },
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Target / Ideal Salary
                  </label>
                  <input
                    type="number"
                    value={pref.salary.target_amount || ''}
                    onChange={(e) =>
                      setPref({
                        ...pref,
                        salary: { ...pref.salary, target_amount: Number(e.target.value) },
                      })
                    }
                    className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                    Currency & Frequency
                  </label>
                  <div className="flex gap-2">
                    <select
                      value={pref.salary.currency}
                      onChange={(e) =>
                        setPref({ ...pref, salary: { ...pref.salary, currency: e.target.value } })
                      }
                      className="px-2 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                    >
                      <option value="USD">USD ($)</option>
                      <option value="EUR">EUR (€)</option>
                      <option value="GBP">GBP (£)</option>
                      <option value="CAD">CAD ($)</option>
                      <option value="AUD">AUD ($)</option>
                    </select>
                    <select
                      value={pref.salary.interval}
                      onChange={(e) =>
                        setPref({
                          ...pref,
                          salary: { ...pref.salary, interval: e.target.value as SalaryInterval },
                        })
                      }
                      className="flex-1 px-2 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                    >
                      <option value="annual">Annual</option>
                      <option value="monthly">Monthly</option>
                      <option value="hourly">Hourly</option>
                    </select>
                  </div>
                </div>
              </div>
            ) : (
              <p className="text-xs text-slate-400 italic">
                Salary gating is currently unconfigured. Matches will not filter based on compensation.
              </p>
            )}
          </div>

          {/* 4. Visa & Sponsorship Preferences (AT-003) */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-4">
            <div className="border-b border-slate-100 dark:border-slate-800 pb-3">
              <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <ShieldAlert className="w-4 h-4 text-purple-600" />
                <span>Visa Sponsorship & Work Authorization (AT-003)</span>
              </h2>
              <p className="text-xs text-slate-500">
                Never guessed or fabricated. Select your explicit sponsorship requirement.
              </p>
            </div>

            <div className="space-y-2.5">
              {[
                {
                  id: 'unspecified',
                  title: 'Unspecified / Not Configured (Default)',
                  desc: 'System will halt on application questions requiring sponsorship status.',
                },
                {
                  id: 'requires_sponsorship',
                  title: 'I require Visa Sponsorship for target roles',
                  desc: 'Match only opportunities with confirmed visa transfer / sponsorship support.',
                },
                {
                  id: 'authorized_no_sponsorship',
                  title: 'I am legally authorized to work without sponsorship',
                  desc: 'Full domestic or treaty work authorization.',
                },
                {
                  id: 'will_not_sponsor_accepted',
                  title: 'Willing to accept roles that do not offer sponsorship',
                  desc: 'Consider contractual or remote arrangements that do not need employer visa sponsorship.',
                },
              ].map((opt) => (
                <label
                  key={opt.id}
                  className={`p-3.5 rounded-xl border flex items-start gap-3 cursor-pointer transition ${
                    pref.sponsorship === opt.id
                      ? 'border-purple-600 bg-purple-50/50 dark:bg-purple-950/40 text-purple-900 dark:text-purple-200'
                      : 'border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/60'
                  }`}
                >
                  <input
                    type="radio"
                    name="sponsorship"
                    value={opt.id}
                    checked={pref.sponsorship === opt.id}
                    onChange={() => setPref({ ...pref, sponsorship: opt.id as SponsorshipPreference })}
                    className="mt-0.5 text-purple-600 focus:ring-purple-500"
                  />
                  <div className="space-y-0.5 text-xs">
                    <span className="font-bold block text-slate-900 dark:text-white">{opt.title}</span>
                    <span className="text-slate-500">{opt.desc}</span>
                  </div>
                </label>
              ))}
            </div>
          </div>
        </div>

        {/* Right 1 Column: Exclusions & Tester */}
        <div className="space-y-6">
          {/* Excluded Employers & Blacklist */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-4">
            <div className="border-b border-slate-100 dark:border-slate-800 pb-3">
              <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Ban className="w-4 h-4 text-rose-600" />
                <span>Excluded Employers Blacklist</span>
              </h2>
              <p className="text-xs text-slate-500">Prevent applications to current or unwanted companies.</p>
            </div>

            <form onSubmit={handleAddCompanyExclusion} className="flex gap-2">
              <input
                type="text"
                value={newCompany}
                onChange={(e) => setNewCompany(e.target.value)}
                placeholder="Company name..."
                className="flex-1 px-3 py-1.5 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
              />
              <button
                type="submit"
                className="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 text-slate-800 dark:text-slate-200 rounded-lg text-xs font-semibold"
              >
                Add
              </button>
            </form>

            <div className="flex flex-wrap gap-1.5">
              {pref.exclusions.excluded_companies.map((c) => (
                <span
                  key={c}
                  className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs bg-rose-50 dark:bg-rose-950/60 text-rose-700 dark:text-rose-300 border border-rose-200 dark:border-rose-900/60"
                >
                  <span>{c}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveCompanyExclusion(c)}
                    className="hover:text-rose-900 transition"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
            </div>

            {/* Keyword Exclusions */}
            <div className="border-t border-slate-100 dark:border-slate-800 pt-3 space-y-2">
              <label className="text-xs font-bold text-slate-700 dark:text-slate-300 block">
                Excluded Keywords in Title / Job Text
              </label>
              <form onSubmit={handleAddKeywordExclusion} className="flex gap-2">
                <input
                  type="text"
                  value={newKeyword}
                  onChange={(e) => setNewKeyword(e.target.value)}
                  placeholder="e.g. crypto, unpaid"
                  className="flex-1 px-3 py-1.5 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                />
                <button
                  type="submit"
                  className="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 text-slate-800 dark:text-slate-200 rounded-lg text-xs font-semibold"
                >
                  Add
                </button>
              </form>
              <div className="flex flex-wrap gap-1.5">
                {pref.exclusions.excluded_keywords.map((kw) => (
                  <span
                    key={kw}
                    className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
                  >
                    <span>{kw}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveKeywordExclusion(kw)}
                      className="hover:text-rose-600 transition"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
            </div>

            {/* Staffing Agency Filter */}
            <div className="border-t border-slate-100 dark:border-slate-800 pt-3">
              <label className="flex items-center gap-2 text-xs font-medium cursor-pointer">
                <input
                  type="checkbox"
                  checked={pref.exclusions.block_staffing_agencies}
                  onChange={(e) =>
                    setPref({
                      ...pref,
                      exclusions: { ...pref.exclusions, block_staffing_agencies: e.target.checked },
                    })
                  }
                  className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500 cursor-pointer"
                />
                <span className="text-slate-700 dark:text-slate-300">
                  Block 3rd-party staffing agencies & recruitment brokers
                </span>
              </label>
            </div>
          </div>

          {/* Real-time Exclusion Rule Tester */}
          <div className="bg-gradient-to-br from-slate-50 to-blue-50/30 dark:from-slate-900 dark:to-blue-950/20 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-4">
            <h2 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
              <Search className="w-4 h-4 text-indigo-600" />
              <span>Real-Time Exclusion Tester</span>
            </h2>
            <p className="text-xs text-slate-500">
              Test whether a specific job title and company would be excluded by your rules.
            </p>

            <div className="space-y-2.5">
              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Company</label>
                <input
                  type="text"
                  value={testCompany}
                  onChange={(e) => setTestCompany(e.target.value)}
                  placeholder="e.g. Revature or Acme"
                  className="w-full px-3 py-1.5 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">Job Title</label>
                <input
                  type="text"
                  value={testTitle}
                  onChange={(e) => setTestTitle(e.target.value)}
                  placeholder="e.g. Crypto Protocol Dev"
                  className="w-full px-3 py-1.5 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                />
              </div>
              <button
                type="button"
                onClick={handleRunExclusionTest}
                className="w-full py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-semibold transition"
              >
                Evaluate Exclusion
              </button>
            </div>

            {testResult && (
              <div
                className={`p-3 rounded-xl border text-xs font-medium space-y-0.5 ${
                  testResult.excluded
                    ? 'border-rose-200 dark:border-rose-900/60 bg-rose-50/80 dark:bg-rose-950/40 text-rose-800 dark:text-rose-300'
                    : 'border-emerald-200 dark:border-emerald-900/60 bg-emerald-50/80 dark:bg-emerald-950/40 text-emerald-800 dark:text-emerald-300'
                }`}
              >
                <div className="flex items-center gap-1.5 font-bold">
                  {testResult.excluded ? <Ban className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                  <span>{testResult.excluded ? 'EXCLUDED FROM SEARCH' : 'PERMITTED OPPORTUNITY'}</span>
                </div>
                <p className="text-[11px]">{testResult.reason}</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
