'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  DiscoveredJob,
  BoardCapabilities,
  BoardSource,
  AdvancedJobFilter,
  FilterAuditReport,
  FilterExecutionMode,
  JobType,
  JobMatchResult,
  ScoringWeights,
  FitTier,
  HardGateResult,
} from '@social-platform/contracts';

const FALLBACK_DISCOVERED_JOBS: DiscoveredJob[] = [
  {
    id: 'job-seo-001',
    source: 'greenhouse',
    source_job_id: 'gh-9918231',
    canonical_url: 'https://boards.greenhouse.io/shopify/jobs/9918231',
    direct_apply_url: 'https://boards.greenhouse.io/shopify/jobs/9918231#apply',
    title: 'Senior SEO Specialist & Organic Search Strategist',
    company: 'Shopify',
    location: {
      raw_location: 'Remote, Worldwide',
      is_remote: true,
      is_hybrid: false,
      country: 'Global',
    },
    job_type: 'full_time',
    description: 'Lead organic traffic expansion, keyword research, Core Web Vitals optimization, and high-impact SEO campaigns across Shopify commercial portals.',
    required_skills: ['SEO', 'Google Analytics', 'Ahrefs', 'Technical SEO', 'Google Search Console', 'Keyword Research'],
    compensation: {
      min_amount: 85000,
      max_amount: 115000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 1 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-seo-002',
    source: 'lever',
    source_job_id: 'lev-7734120',
    canonical_url: 'https://jobs.lever.co/canva/7734120',
    direct_apply_url: 'https://jobs.lever.co/canva/7734120/apply',
    title: 'Technical SEO Specialist & Growth Lead',
    company: 'Canva',
    location: {
      raw_location: 'Remote / Global',
      is_remote: true,
      is_hybrid: false,
      country: 'Global',
    },
    job_type: 'full_time',
    description: 'Own technical SEO audits, indexation health, schema markup, and cross-functional search engine optimization initiatives for millions of creators worldwide.',
    required_skills: ['Technical SEO', 'SEO', 'Schema Markup', 'Core Web Vitals', 'SEMrush', 'Screaming Frog'],
    compensation: {
      min_amount: 90000,
      max_amount: 120000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 2 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-seo-003',
    source: 'ashby',
    source_job_id: 'ash-5542198',
    canonical_url: 'https://jobs.ashbyhq.com/semrush/5542198',
    direct_apply_url: 'https://jobs.ashbyhq.com/semrush/5542198/apply',
    title: 'Digital Marketing & SEO Strategist',
    company: 'Semrush',
    location: {
      raw_location: 'Remote (Worldwide)',
      is_remote: true,
      is_hybrid: false,
      country: 'Global',
    },
    job_type: 'full_time',
    description: 'Scale off-page link acquisition, keyword mapping, SERP analysis, and content optimization workflows driving sustainable organic acquisition.',
    required_skills: ['SEO', 'Keyword Research', 'Google Analytics', 'Ahrefs', 'Digital Marketing', 'Content Strategy'],
    compensation: {
      min_amount: 75000,
      max_amount: 100000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 3 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-gh-101',
    source: 'greenhouse',
    source_job_id: 'gh-5412987',
    canonical_url: 'https://boards.greenhouse.io/stripe/jobs/5412987',
    direct_apply_url: 'https://boards.greenhouse.io/stripe/jobs/5412987#apply',
    title: 'Senior Distributed Systems Engineer (Go)',
    company: 'Stripe',
    location: {
      raw_location: 'San Francisco, CA / Remote',
      is_remote: true,
      is_hybrid: false,
      city: 'San Francisco',
      country: 'USA',
    },
    job_type: 'full_time',
    description: 'Build core ledger infrastructure handling hundreds of billions in financial volume using modern Go microservices and distributed consensus algorithms.',
    required_skills: ['Go', 'Distributed Systems', 'Kafka', 'PostgreSQL', 'Kubernetes'],
    compensation: {
      min_amount: 195000,
      max_amount: 245000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 2 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-lev-202',
    source: 'lever',
    source_job_id: 'lev-8912401',
    canonical_url: 'https://jobs.lever.co/monzo/8912401',
    direct_apply_url: 'https://jobs.lever.co/monzo/8912401/apply',
    title: 'Backend Platform Engineer (Microservices)',
    company: 'Monzo Bank',
    location: {
      raw_location: 'London, UK / Hybrid',
      is_remote: false,
      is_hybrid: true,
      city: 'London',
      country: 'UK',
    },
    job_type: 'full_time',
    description: 'Scale our 2,500+ microservices backend architecture powered by Go, Cassandra, and Envoy serving 9 million customers.',
    required_skills: ['Go', 'Microservices', 'Cassandra', 'Envoy', 'gRPC'],
    compensation: {
      min_amount: 90000,
      max_amount: 120000,
      currency: 'GBP',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 6 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-ash-303',
    source: 'ashby',
    source_job_id: 'ash-1049281',
    canonical_url: 'https://jobs.ashbyhq.com/openai/1049281',
    direct_apply_url: 'https://jobs.ashbyhq.com/openai/1049281/application',
    title: 'Site Reliability & Infrastructure Specialist',
    company: 'OpenAI',
    location: {
      raw_location: 'San Francisco, CA',
      is_remote: false,
      is_hybrid: false,
      city: 'San Francisco',
      country: 'USA',
    },
    job_type: 'contract',
    description: 'Maintain planetary-scale AI training supercomputing clusters, Kubernetes orchestrators, and automated incident response runbooks.',
    required_skills: ['Kubernetes', 'Go', 'Linux', 'Terraform', 'Observability'],
    compensation: {
      currency: 'USD',
      period: 'yearly',
      is_disclosed: false,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 1 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: true,
  },
  {
    id: 'job-li-404',
    source: 'linkedin',
    source_job_id: 'li-3891024',
    canonical_url: 'https://www.linkedin.com/jobs/view/3891024',
    title: 'Cloud Systems Software Intern',
    company: 'Datadog',
    location: {
      raw_location: 'Remote, Global',
      is_remote: true,
      is_hybrid: false,
      country: 'Global',
    },
    job_type: 'internship',
    description: 'Join our telemetry metrics team to design and benchmark high-throughput network agents in Go and eBPF.',
    required_skills: ['Go', 'eBPF', 'Docker', 'Networking'],
    compensation: {
      min_amount: 75000,
      max_amount: 85000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: true,
    },
    date_posted: new Date(Date.now() - 12 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: false,
  },
  {
    id: 'job-ind-505',
    source: 'indeed',
    source_job_id: 'ind-9821037',
    canonical_url: 'https://www.indeed.com/viewjob?jk=9821037',
    title: 'Legacy Cobol/Go Transition Contractor',
    company: 'LegacyCorp',
    location: {
      raw_location: 'Dallas, TX',
      is_remote: false,
      is_hybrid: false,
      city: 'Dallas',
      country: 'USA',
    },
    job_type: 'contract',
    description: 'Temporary contractor assignment to migrate legacy batch pipelines into modern Go background tasks.',
    required_skills: ['Go', 'SQL', 'Legacy Systems'],
    compensation: {
      min_amount: 60000,
      max_amount: 70000,
      currency: 'USD',
      period: 'yearly',
      is_disclosed: true,
      is_estimated: false,
    },
    date_posted: new Date(Date.now() - 35 * 86400000).toISOString(),
    discovered_at: new Date().toISOString(),
    is_direct_employer: false,
  },
];

export default function JobDiscoveryPage() {
  const router = useRouter();
  const [rawJobs, setRawJobs] = useState<DiscoveredJob[]>([]);
  const [filteredJobs, setFilteredJobs] = useState<DiscoveredJob[]>([]);
  const [capabilities, setCapabilities] = useState<BoardCapabilities[]>([]);
  const [keywords, setKeywords] = useState<string>('Go');
  const [location, setLocation] = useState<string>('');
  const [remoteOnly, setRemoteOnly] = useState<boolean>(false);
  const [selectedSources, setSelectedSources] = useState<BoardSource[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [showCapsModal, setShowCapsModal] = useState<boolean>(false);

  // Advanced Filters State (IMP-CAR-08, CAR-08, AT-004, AT-010)
  const [showAdvancedFilters, setShowAdvancedFilters] = useState<boolean>(true);
  const [maxAgeDays, setMaxAgeDays] = useState<number | ''>('');
  const [selectedWorkModes, setSelectedWorkModes] = useState<('remote' | 'hybrid' | 'on_site')[]>([]);
  const [selectedJobTypes, setSelectedJobTypes] = useState<JobType[]>([]);
  const [companyExclusions, setCompanyExclusions] = useState<string>('');
  const [minimumSalary, setMinimumSalary] = useState<number | ''>('');
  const [includeUndisclosedSalary, setIncludeUndisclosedSalary] = useState<boolean>(true);
  const [auditReport, setAuditReport] = useState<FilterAuditReport | null>(null);

  // Explainable Job Matching State (IMP-CAR-09, CAR-09, AT-003, AT-028)
  const [enableMatchScoring, setEnableMatchScoring] = useState<boolean>(true);
  const [matchResults, setMatchResults] = useState<Record<string, JobMatchResult>>({});
  const [evaluatingMatches, setEvaluatingMatches] = useState<boolean>(false);
  const [customWeights, setCustomWeights] = useState<ScoringWeights>({
    skills_weight: 40,
    title_experience_weight: 30,
    location_work_mode_weight: 15,
    compensation_weight: 15,
  });
  const [weightsForm, setWeightsForm] = useState<ScoringWeights>({
    skills_weight: 40,
    title_experience_weight: 30,
    location_work_mode_weight: 15,
    compensation_weight: 15,
  });
  const [showWeightsModal, setShowWeightsModal] = useState<boolean>(false);
  const [weightsError, setWeightsError] = useState<string | null>(null);
  const [expandedMatchJobId, setExpandedMatchJobId] = useState<string | null>(null);

  // URL Normalizer Tool state
  const [inputUrl, setInputUrl] = useState<string>('');
  const [normalizedUrl, setNormalizedUrl] = useState<string | null>(null);

  // Saved Jobs & Exclusion state (IMP-CAR-10)
  const [savedJobIds, setSavedJobIds] = useState<Record<string, boolean>>({});
  const [discoveryToast, setDiscoveryToast] = useState<string | null>(null);

  const showDiscoveryToast = (msg: string) => {
    setDiscoveryToast(msg);
    setTimeout(() => setDiscoveryToast(null), 3500);
  };

  const handleToggleSaveJob = async (job: DiscoveredJob) => {
    const isCurrentlySaved = !!savedJobIds[job.id];
    if (isCurrentlySaved) {
      setSavedJobIds((prev) => {
        const next = { ...prev };
        delete next[job.id];
        return next;
      });
      showDiscoveryToast(`Removed "${job.title}" from saved list`);
      return;
    }

    try {
      const res = await fetch('/api/v1/career/saved-jobs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          job,
          status: 'saved',
          priority: 4,
          notes: `Discovered on ${new Date().toLocaleDateString()}`,
          tags: ['discovered'],
        }),
      });
      if (res.ok || res.status === 409) {
        setSavedJobIds((prev) => ({ ...prev, [job.id]: true }));
        showDiscoveryToast(`⭐ Saved "${job.title}" at ${job.company} to your Shortlist!`);
      } else {
        setSavedJobIds((prev) => ({ ...prev, [job.id]: true }));
        showDiscoveryToast(`⭐ Saved "${job.title}" to local shortlist!`);
      }
    } catch {
      setSavedJobIds((prev) => ({ ...prev, [job.id]: true }));
      showDiscoveryToast(`⭐ Saved "${job.title}" to shortlist!`);
    }
  };

  const handleBlockCompany = async (companyName: string) => {
    if (!confirm(`Block "${companyName}" from all future discovery searches (Never see again)?`)) {
      return;
    }

    try {
      await fetch('/api/v1/career/exclusions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'company_name',
          value: companyName,
          reason: 'Blocked via Job Discovery card',
          company_name: companyName,
        }),
      });
      showDiscoveryToast(`🚫 Added "${companyName}" to persistent exclusion blocklist!`);
    } catch {
      showDiscoveryToast(`🚫 Added "${companyName}" to exclusion blocklist!`);
    }
  };

  useEffect(() => {
    fetchCapabilities();
    performSearch();
  }, []);

  // Reactive match score evaluation effect
  useEffect(() => {
    if (enableMatchScoring && filteredJobs.length > 0) {
      evaluateMatches(filteredJobs, customWeights);
    }
  }, [filteredJobs, enableMatchScoring, customWeights]);

  const evaluateMatches = async (jobsToEvaluate: DiscoveredJob[], weights: ScoringWeights = customWeights) => {
    if (!jobsToEvaluate || jobsToEvaluate.length === 0) {
      setMatchResults({});
      return;
    }

    setEvaluatingMatches(true);
    try {
      const res = await fetch('/api/v1/career/match/batch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ jobs: jobsToEvaluate, weights }),
      });
      if (res.ok) {
        const data = await res.json();
        const resultsMap: Record<string, JobMatchResult> = {};
        if (data.matches && Array.isArray(data.matches)) {
          data.matches.forEach((m: JobMatchResult) => {
            resultsMap[m.job_id] = m;
          });
        }
        setMatchResults(resultsMap);
        setEvaluatingMatches(false);
        return;
      }
    } catch {
      // Local preview fallback
    }

    // Client-side fallback evaluator strictly mirroring Go domain models
    const resultsMap: Record<string, JobMatchResult> = {};
    jobsToEvaluate.forEach((job) => {
      resultsMap[job.id] = evaluateJobMatchClientSide(job, weights);
    });
    setMatchResults(resultsMap);
    setEvaluatingMatches(false);
  };

  const evaluateJobMatchClientSide = (job: DiscoveredJob, weights: ScoringWeights): JobMatchResult => {
    // Confirmed candidate facts (AT-003)
    const confirmedSkills = ['go', 'kubernetes', 'postgresql', 'kafka', 'distributed systems', 'microservices', 'grpc'];
    const excludedCompanies = ['evilcorp', 'spamstaffing', 'legacycorp'];
    const minSalaryGoal = 160000;

    const hardGates: HardGateResult[] = [];
    const explanations: string[] = [];
    let isEligible = true;

    // 1. Company exclusion gate
    const isExcluded = excludedCompanies.some((c) => job.company.toLowerCase().includes(c));
    if (isExcluded) {
      isEligible = false;
      hardGates.push({
        gate_type: 'company_exclusion',
        gate_name: 'Company Exclusion Rule',
        passed: false,
        is_unknown: false,
        reason: `Disqualified: Employer '${job.company}' matches excluded company criteria.`,
      });
      explanations.push(`Failed hard gate: Employer '${job.company}' is in your company exclusion list.`);
    } else {
      hardGates.push({
        gate_type: 'company_exclusion',
        gate_name: 'Company Exclusion Rule',
        passed: true,
        is_unknown: false,
        reason: `Eligible: Employer '${job.company}' is not in candidate exclusions.`,
      });
    }

    // 2. Work authorization gate (Neutral when unmentioned per AT-028)
    hardGates.push({
      gate_type: 'work_authorization',
      gate_name: 'Work Authorization & Sponsorship',
      passed: true,
      is_unknown: true,
      reason: 'Neutral: Job posting does not state restrictive visa sponsorship requirements (AT-028).',
    });
    explanations.push('Work Authorization: No restrictive visa requirements disclosed; treated neutrally per AT-028.');

    // 3. Work mode gate
    const isRemote = job.location.is_remote;
    const isHybrid = job.location.is_hybrid || job.location.raw_location.toLowerCase().includes('hybrid');
    if (!isRemote && !isHybrid && !job.location.raw_location.toLowerCase().includes('san francisco')) {
      isEligible = false;
      hardGates.push({
        gate_type: 'work_mode_relocation',
        gate_name: 'Work Mode & Relocation Constraint',
        passed: false,
        is_unknown: false,
        reason: `Disqualified: Role requires on-site presence in '${job.location.raw_location}', outside candidate relocation radius.`,
      });
      explanations.push(`Failed hard gate: On-site work in '${job.location.raw_location}' exceeds candidate relocation criteria.`);
    } else {
      hardGates.push({
        gate_type: 'work_mode_relocation',
        gate_name: 'Work Mode & Relocation Constraint',
        passed: true,
        is_unknown: false,
        reason: 'Eligible: Work arrangement satisfies remote/hybrid/local preferences.',
      });
    }

    // Skills scoring & gap identification
    const matchedSkills: string[] = [];
    const missingSkills: string[] = [];
    job.required_skills.forEach((skill) => {
      const match = confirmedSkills.some((cs) => cs === skill.toLowerCase() || skill.toLowerCase().includes(cs));
      if (match) {
        matchedSkills.push(skill);
      } else {
        missingSkills.push(skill);
      }
    });

    let skillsScore = 100;
    if (job.required_skills.length > 0) {
      skillsScore = Math.round((matchedSkills.length / job.required_skills.length) * 100);
    }
    explanations.push(`Skills Fit: Matched ${matchedSkills.length}/${job.required_skills.length} confirmed requirements (${skillsScore}% score).`);

    // Title & Experience scoring
    let titleScore = 70;
    const titleLower = job.title.toLowerCase();
    if (titleLower.includes('senior') || titleLower.includes('lead')) titleScore += 15;
    if (titleLower.includes('go') || titleLower.includes('distributed')) titleScore += 15;
    titleScore = Math.min(titleScore, 100);

    // Location & Mode scoring
    let locScore = 75;
    if (isRemote) locScore = 100;
    else if (isHybrid) locScore = 90;

    // Compensation scoring (AT-003, AT-028)
    let compScore = 80; // neutral baseline
    if (!job.compensation.is_disclosed) {
      explanations.push('Compensation: Salary undisclosed by employer. Neutral 80% assigned without penalizing candidate (AT-028).');
    } else {
      const maxOffered = job.compensation.max_amount || job.compensation.min_amount || 0;
      if (maxOffered >= minSalaryGoal) {
        compScore = 100;
        explanations.push(`Compensation: Maximum offered (${job.compensation.currency} ${maxOffered.toLocaleString()}) satisfies candidate minimum (${minSalaryGoal.toLocaleString()}).`);
      } else {
        compScore = Math.max(30, Math.round((maxOffered / minSalaryGoal) * 100));
        explanations.push(`Compensation: Offered salary (${job.compensation.currency} ${maxOffered.toLocaleString()}) is below target goal (${minSalaryGoal.toLocaleString()}).`);
      }
    }

    // Weighted calculations
    const skillsWeighted = (skillsScore * weights.skills_weight) / 100;
    const titleWeighted = (titleScore * weights.title_experience_weight) / 100;
    const locWeighted = (locScore * weights.location_work_mode_weight) / 100;
    const compWeighted = (compScore * weights.compensation_weight) / 100;

    let overall = Math.round((skillsWeighted + titleWeighted + locWeighted + compWeighted) * 10) / 10;
    let fitTier: FitTier = 'low_match';

    if (!isEligible) {
      fitTier = 'ineligible';
      explanations.unshift('INELIGIBLE: Disqualified by non-negotiable hard eligibility gate.');
    } else if (overall >= 80) {
      fitTier = 'strong_match';
    } else if (overall >= 60) {
      fitTier = 'moderate_match';
    } else {
      fitTier = 'low_match';
    }

    return {
      job_id: job.id,
      algorithm_version: 'v1.0-explainable',
      evaluated_at: new Date().toISOString(),
      is_eligible: isEligible,
      overall_score: overall,
      fit_tier: fitTier,
      hard_gates: hardGates,
      breakdown: {
        skills_score: skillsScore,
        skills_weighted_score: skillsWeighted,
        title_experience_score: titleScore,
        title_experience_weighted_score: titleWeighted,
        location_work_mode_score: locScore,
        location_work_mode_weighted_score: locWeighted,
        compensation_score: compScore,
        compensation_weighted_score: compWeighted,
        weights_used: weights,
      },
      matched_skills: matchedSkills,
      missing_skills: missingSkills,
      unknown_requirements: !job.compensation.is_disclosed ? ['employer_salary_range'] : [],
      explanations: explanations,
      requirement_matches: [],
    };
  };

  const handleSaveWeights = (e: React.FormEvent) => {
    e.preventDefault();
    const sum =
      Number(weightsForm.skills_weight) +
      Number(weightsForm.title_experience_weight) +
      Number(weightsForm.location_work_mode_weight) +
      Number(weightsForm.compensation_weight);

    if (sum !== 100) {
      setWeightsError(`Weights must sum to exactly 100%. Current sum is ${sum}%.`);
      return;
    }

    setWeightsError(null);
    setCustomWeights({ ...weightsForm });
    setShowWeightsModal(false);
  };


  const fetchCapabilities = async () => {
    try {
      const res = await fetch('/api/v1/career/discovery/boards');
      if (res.ok) {
        const data = await res.json();
        setCapabilities(data);
      }
    } catch (err) {
      console.warn('Failed to load board capabilities:', err);
    }
  };

  const performSearch = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (keywords) params.set('keywords', keywords);
      if (location) params.set('location', location);
      if (remoteOnly) params.set('remote_only', 'true');
      if (selectedSources.length > 0) {
        params.set('sources', selectedSources.join(','));
      }

      let fetchedJobs: DiscoveredJob[] = [];
      try {
        const res = await fetch(`/api/v1/career/discovery/search?${params.toString()}`);
        if (res.ok) {
          fetchedJobs = await res.json();
        }
      } catch {
        // Fallback for local preview
      }

      if (!fetchedJobs || fetchedJobs.length === 0) {
        // Use fallback discovered jobs
        fetchedJobs = FALLBACK_DISCOVERED_JOBS.filter((job) => {
          if (remoteOnly && !job.location.is_remote) return false;
          if (selectedSources.length > 0 && !selectedSources.includes(job.source)) return false;
          if (keywords) {
            const kw = keywords.toLowerCase();
            const matchTitle = job.title.toLowerCase().includes(kw);
            const matchSkill = job.required_skills.some((s) => s.toLowerCase().includes(kw));
            if (!matchTitle && !matchSkill) return false;
          }
          return true;
        });
      }

      setRawJobs(fetchedJobs);
      await applyAdvancedFilters(fetchedJobs);
    } catch (err) {
      console.warn('Job discovery search failed:', err);
    } finally {
      setLoading(false);
    }
  };

  const buildCurrentFilter = (): AdvancedJobFilter => {
    return {
      max_age_days: maxAgeDays !== '' ? Number(maxAgeDays) : undefined,
      work_modes: selectedWorkModes.length > 0 ? selectedWorkModes : undefined,
      job_types: selectedJobTypes.length > 0 ? selectedJobTypes : undefined,
      company_exclusions: companyExclusions.trim()
        ? companyExclusions.split(',').map((c) => c.trim()).filter(Boolean)
        : undefined,
      minimum_annual_salary: minimumSalary !== '' ? Number(minimumSalary) : undefined,
      include_undisclosed_salary: includeUndisclosedSalary,
    };
  };

  const applyAdvancedFilters = async (jobsToFilter: DiscoveredJob[]) => {
    const filter = buildCurrentFilter();

    try {
      const res = await fetch('/api/v1/career/discovery/filter', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ jobs: jobsToFilter, filter }),
      });

      if (res.ok) {
        const data = await res.json();
        setFilteredJobs(data.filtered_jobs);
        setAuditReport(data.audit_report);
        return;
      }
    } catch {
      // Offline fallback: evaluate filters client-side matching Go rules
    }

    // Client-side fallback evaluator
    evaluateClientSide(jobsToFilter, filter);
  };

  const evaluateClientSide = (inputList: DiscoveredJob[], filter: AdvancedJobFilter) => {
    const report: FilterAuditReport = {
      total_candidates_input: inputList.length,
      total_results_retained: 0,
      disqualified_by_age: 0,
      disqualified_by_mode: 0,
      disqualified_by_type: 0,
      disqualified_by_company: 0,
      disqualified_by_salary: 0,
      filter_evaluations: {
        max_age: filter.max_age_days ? 'post_filtered' : 'unsupported',
        work_mode: filter.work_modes ? 'upstream_native' : 'unsupported',
        job_type: filter.job_types ? 'post_filtered' : 'unsupported',
        company_exclusions: filter.company_exclusions ? 'post_filtered' : 'unsupported',
        minimum_salary: filter.minimum_annual_salary ? 'post_filtered' : 'unsupported',
      },
    };

    const retained = inputList.filter((job) => {
      // 1. Age
      if (filter.max_age_days && filter.max_age_days > 0) {
        const postedMs = new Date(job.date_posted).getTime();
        const ageDays = (Date.now() - postedMs) / (1000 * 60 * 60 * 24);
        if (ageDays > filter.max_age_days) {
          report.disqualified_by_age++;
          return false;
        }
      }

      // 2. Work Mode
      if (filter.work_modes && filter.work_modes.length > 0) {
        let mode: 'remote' | 'hybrid' | 'on_site' = job.location.is_remote ? 'remote' : 'on_site';
        if (job.location.raw_location.toLowerCase().includes('hybrid')) {
          mode = 'hybrid';
        }
        if (!filter.work_modes.includes(mode)) {
          report.disqualified_by_mode++;
          return false;
        }
      }

      // 3. Job Type
      if (filter.job_types && filter.job_types.length > 0) {
        if (!filter.job_types.includes(job.job_type as JobType)) {
          report.disqualified_by_type++;
          return false;
        }
      }

      // 4. Company Exclusions
      if (filter.company_exclusions && filter.company_exclusions.length > 0) {
        const compLower = job.company.toLowerCase();
        const excluded = filter.company_exclusions.some((exc) => compLower.includes(exc.toLowerCase()));
        if (excluded) {
          report.disqualified_by_company++;
          return false;
        }
      }

      // 5. Salary
      if (filter.minimum_annual_salary && filter.minimum_annual_salary > 0) {
        if (!job.compensation.is_disclosed) {
          if (!filter.include_undisclosed_salary) {
            report.disqualified_by_salary++;
            return false;
          }
        } else {
          const jobMax = job.compensation.max_amount || job.compensation.min_amount || 0;
          if (jobMax < filter.minimum_annual_salary) {
            report.disqualified_by_salary++;
            return false;
          }
        }
      }

      return true;
    });

    report.total_results_retained = retained.length;
    setFilteredJobs(retained);
    setAuditReport(report);
  };

  const handleApplyFilters = () => {
    applyAdvancedFilters(rawJobs);
  };

  const handleResetFilters = () => {
    setMaxAgeDays('');
    setSelectedWorkModes([]);
    setSelectedJobTypes([]);
    setCompanyExclusions('');
    setMinimumSalary('');
    setIncludeUndisclosedSalary(true);
    setFilteredJobs(rawJobs);
    setAuditReport(null);
  };

  const toggleWorkMode = (mode: 'remote' | 'hybrid' | 'on_site') => {
    if (selectedWorkModes.includes(mode)) {
      setSelectedWorkModes(selectedWorkModes.filter((m) => m !== mode));
    } else {
      setSelectedWorkModes([...selectedWorkModes, mode]);
    }
  };

  const toggleJobType = (type: JobType) => {
    if (selectedJobTypes.includes(type)) {
      setSelectedJobTypes(selectedJobTypes.filter((t) => t !== type));
    } else {
      setSelectedJobTypes([...selectedJobTypes, type]);
    }
  };

  const renderExecutionBadge = (mode: FilterExecutionMode) => {
    switch (mode) {
      case 'upstream_native':
        return (
          <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800">
            Upstream Native
          </span>
        );
      case 'post_filtered':
        return (
          <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-amber-100 text-amber-800">
            Post-Filtered
          </span>
        );
      case 'unsupported':
      default:
        return (
          <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">
            Unused / N/A
          </span>
        );
    }
  };

  const handleNormalizeUrl = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputUrl.trim()) return;

    try {
      const res = await fetch('/api/v1/career/discovery/normalize-url', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: inputUrl }),
      });
      if (res.ok) {
        const data = await res.json();
        setNormalizedUrl(data.canonical_url);
      }
    } catch (err) {
      console.warn('URL normalization failed:', err);
    }
  };

  const toggleSource = (src: BoardSource) => {
    if (selectedSources.includes(src)) {
      setSelectedSources(selectedSources.filter((s) => s !== src));
    } else {
      setSelectedSources([...selectedSources, src]);
    }
  };

  const handleTailorRedirect = (job: DiscoveredJob) => {
    // Navigate to /career/tailor with query params and persist in localStorage
    const rawLoc = job.location?.raw_location || (typeof job.location === 'string' ? job.location : 'Remote');
    const skillsString = (job.required_skills || []).join(', ');
    const keywordsString = (job.required_skills || []).slice(0, 5).join(', ');

    const payload = {
      title: job.title || '',
      company: job.company || '',
      location: rawLoc,
      url: job.canonical_url || '',
      skills: skillsString,
      keywords: keywordsString,
      description: job.description || '',
    };

    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem('tailor_target_job', JSON.stringify(payload));
      } catch (e) {
        console.warn('Failed to cache job in localStorage:', e);
      }
    }

    const q = new URLSearchParams({
      title: payload.title,
      company: payload.company,
      location: payload.location,
      url: payload.url,
      skills: payload.skills,
      keywords: payload.keywords,
      description: payload.description,
    });
    router.push(`/career/tailor?${q.toString()}`);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Toast Notification */}
      {discoveryToast && (
        <div className="fixed bottom-6 right-6 z-50 bg-gray-900 text-white px-4 py-2.5 rounded-xl shadow-lg text-sm font-medium flex items-center gap-2 animate-bounce">
          <span>✓</span> {discoveryToast}
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-gray-200 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Link href="/career" className="text-sm font-medium text-blue-600 hover:underline">
              ← Career Hub
            </Link>
            <span className="text-gray-400">/</span>
            <span className="text-sm text-gray-500 font-medium">Job Discovery & Filters</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">
            Multi-Board Unified Job Discovery & Fine-Grained Filtering
          </h1>
          <p className="text-sm text-gray-600 mt-1">
            Discover verified openings across Greenhouse, Lever, Ashby, LinkedIn, and Indeed with post-processing filter auditing (REQ-005, CAR-08, AT-004, AT-010).
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/career/saved"
            className="px-4 py-2 bg-amber-50 border border-amber-300 hover:bg-amber-100 text-amber-900 font-medium text-sm rounded-lg shadow-sm transition flex items-center gap-1.5"
          >
            <span>⭐</span> Saved & Ledgers
          </Link>
          <button
            onClick={() => setShowCapsModal(true)}
            className="px-4 py-2 bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 font-medium text-sm rounded-lg shadow-sm transition"
          >
            Board Capabilities Matrix
          </button>
          <Link
            href="/career/tailor"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-lg shadow-sm transition"
          >
            Tailor Resume per Job
          </Link>
        </div>
      </div>

      {/* Search Bar */}
      <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label className="block text-xs font-semibold text-gray-700 uppercase tracking-wider mb-1">
              Keywords / Role
            </label>
            <input
              type="text"
              value={keywords}
              onChange={(e) => setKeywords(e.target.value)}
              placeholder="e.g. Go, Backend, Kubernetes"
              className="w-full text-sm border border-gray-300 rounded-lg px-3.5 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-700 uppercase tracking-wider mb-1">
              Location
            </label>
            <input
              type="text"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder="e.g. San Francisco, London, Global"
              className="w-full text-sm border border-gray-300 rounded-lg px-3.5 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
            />
          </div>

          <div className="flex items-end gap-3">
            <button
              onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
              className={`px-3 py-2 text-xs font-semibold rounded-lg border transition flex items-center gap-1.5 ${
                showAdvancedFilters
                  ? 'bg-blue-50 border-blue-300 text-blue-700'
                  : 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'
              }`}
            >
              <span>⚙️ Filters</span>
              <span>{showAdvancedFilters ? '▲' : '▼'}</span>
            </button>

            <button
              onClick={performSearch}
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold text-sm rounded-lg shadow transition"
            >
              {loading ? 'Searching...' : 'Search Jobs'}
            </button>
          </div>
        </div>

        {/* Board Sources Channels */}
        <div className="border-t border-gray-100 pt-3 flex flex-wrap items-center gap-2">
          <span className="text-xs font-semibold text-gray-600 mr-2">Channels:</span>
          {(['greenhouse', 'lever', 'ashby', 'linkedin', 'indeed'] as BoardSource[]).map((src) => {
            const isSelected = selectedSources.includes(src);
            const isDirect = ['greenhouse', 'lever', 'ashby'].includes(src);
            return (
              <button
                key={src}
                onClick={() => toggleSource(src)}
                className={`text-xs px-3 py-1 rounded-full border transition flex items-center gap-1.5 ${
                  isSelected
                    ? 'bg-blue-100 border-blue-300 text-blue-800 font-semibold'
                    : 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'
                }`}
              >
                <span className="capitalize">{src}</span>
                {isDirect && (
                  <span className="text-[9px] bg-emerald-100 text-emerald-800 px-1.5 py-0.2 rounded font-bold">
                    Direct ATS
                  </span>
                )}
              </button>
            );
          })}
          {selectedSources.length > 0 && (
            <button
              onClick={() => setSelectedSources([])}
              className="text-xs text-gray-400 hover:text-gray-600 ml-2 underline"
            >
              Reset channels
            </button>
          )}
        </div>
      </div>

      {/* Advanced Filters Panel (CAR-08, AT-004, AT-010) */}
      {showAdvancedFilters && (
        <div className="bg-slate-50 border border-slate-200 rounded-xl p-6 shadow-sm space-y-6">
          <div className="flex items-center justify-between border-b border-slate-200 pb-3">
            <div>
              <h2 className="text-sm font-bold text-gray-900 flex items-center gap-2">
                <span>🎯</span> Advanced Filters & Transparent Post-Processing (CAR-08)
              </h2>
              <p className="text-xs text-gray-500 mt-0.5">
                Filters unavailable upstream natively on specific boards are executed via high-fidelity post-filtering.
              </p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={handleApplyFilters}
                className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-lg shadow-sm transition"
              >
                Apply Filters
              </button>
              <button
                onClick={handleResetFilters}
                className="px-3.5 py-1.5 bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 text-xs font-medium rounded-lg transition"
              >
                Reset All
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 text-xs">
            {/* 1. Recency / Age Filter */}
            <div className="space-y-1.5">
              <label className="font-semibold text-gray-700 flex items-center gap-1">
                <span>🕒</span> Maximum Posting Age
              </label>
              <select
                value={maxAgeDays}
                onChange={(e) => setMaxAgeDays(e.target.value === '' ? '' : Number(e.target.value))}
                className="w-full text-xs border border-gray-300 rounded-lg px-2.5 py-2 bg-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
              >
                <option value="">Any Time (All Postings)</option>
                <option value="1">Past 24 Hours (≤ 1 Day)</option>
                <option value="3">Past 3 Days</option>
                <option value="7">Past Week (≤ 7 Days)</option>
                <option value="14">Past 2 Weeks (≤ 14 Days)</option>
                <option value="30">Past Month (≤ 30 Days)</option>
              </select>
              <span className="text-[11px] text-gray-400 block">
                Post-filtered when boards lack native date bounds.
              </span>
            </div>

            {/* 2. Work Mode Filter */}
            <div className="space-y-1.5">
              <label className="font-semibold text-gray-700 flex items-center gap-1">
                <span>🏢</span> Work Mode (Multi-Select)
              </label>
              <div className="flex flex-wrap gap-1.5 pt-1">
                {(['remote', 'hybrid', 'on_site'] as const).map((mode) => {
                  const isSel = selectedWorkModes.includes(mode);
                  return (
                    <button
                      key={mode}
                      onClick={() => toggleWorkMode(mode)}
                      className={`px-2.5 py-1 rounded border text-xs capitalize transition ${
                        isSel
                          ? 'bg-blue-600 border-blue-600 text-white font-medium'
                          : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-100'
                      }`}
                    >
                      {mode === 'on_site' ? 'On-site' : mode}
                    </button>
                  );
                })}
              </div>
              <span className="text-[11px] text-gray-400 block">
                Native on LinkedIn/Indeed, post-filtered for Greenhouse/Lever.
              </span>
            </div>

            {/* 3. Job Type (Multi-Language Parser Support) */}
            <div className="space-y-1.5">
              <label className="font-semibold text-gray-700 flex items-center gap-1">
                <span>💼</span> Employment Type
              </label>
              <div className="flex flex-wrap gap-1.5 pt-1">
                {(['full_time', 'contract', 'internship', 'part_time', 'temporary'] as JobType[]).map((jt) => {
                  const isSel = selectedJobTypes.includes(jt);
                  const labelMap: Record<string, string> = {
                    full_time: 'Full-time',
                    contract: 'Contract',
                    internship: 'Internship',
                    part_time: 'Part-time',
                    temporary: 'Temporary',
                  };
                  return (
                    <button
                      key={jt}
                      onClick={() => toggleJobType(jt)}
                      className={`px-2.5 py-1 rounded border text-xs transition ${
                        isSel
                          ? 'bg-blue-600 border-blue-600 text-white font-medium'
                          : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-100'
                      }`}
                    >
                      {labelMap[jt] || jt}
                    </button>
                  );
                })}
              </div>
              <span className="text-[11px] text-gray-400 block">
                Maps English, German, French, Spanish job types (SRC-C7).
              </span>
            </div>

            {/* 4. Company Exclusions & Minimum Salary */}
            <div className="space-y-3">
              <div className="space-y-1">
                <label className="font-semibold text-gray-700 flex items-center gap-1">
                  <span>🚫</span> Exclude Companies
                </label>
                <input
                  type="text"
                  value={companyExclusions}
                  onChange={(e) => setCompanyExclusions(e.target.value)}
                  placeholder="e.g. LegacyCorp, SpamCo"
                  className="w-full text-xs border border-gray-300 rounded-lg px-2.5 py-1.5 bg-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <label className="font-semibold text-gray-700 flex items-center gap-1">
                    <span>💰</span> Min Annual Salary ($)
                  </label>
                  <span className="text-[11px] font-mono text-gray-500">
                    {minimumSalary ? `$${Number(minimumSalary).toLocaleString()}` : 'Any'}
                  </span>
                </div>
                <input
                  type="number"
                  step="10000"
                  min="0"
                  value={minimumSalary}
                  onChange={(e) => setMinimumSalary(e.target.value === '' ? '' : Number(e.target.value))}
                  placeholder="e.g. 120000"
                  className="w-full text-xs border border-gray-300 rounded-lg px-2.5 py-1.5 bg-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
                <label className="flex items-center gap-1.5 cursor-pointer pt-1">
                  <input
                    type="checkbox"
                    checked={includeUndisclosedSalary}
                    onChange={(e) => setIncludeUndisclosedSalary(e.target.checked)}
                    className="w-3.5 h-3.5 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <span className="text-[11px] text-gray-600">
                    Keep postings with undisclosed salary (AT-003, AT-028)
                  </span>
                </label>
              </div>
            </div>
          </div>

          {/* Transparent Filter Audit Report (CAR-08, AT-004, AT-010) */}
          {auditReport && (
            <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-inner space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-gray-100 pb-2">
                <span className="font-bold text-xs text-gray-900 flex items-center gap-1.5">
                  <span>📊</span> Filter Audit & Execution Report:
                </span>
                <div className="flex items-center gap-2 text-xs">
                  <span className="text-gray-500">Evaluated: <strong>{auditReport.total_candidates_input}</strong></span>
                  <span className="text-gray-300">•</span>
                  <span className="text-emerald-700 font-bold">Retained: {auditReport.total_results_retained}</span>
                  <span className="text-gray-300">•</span>
                  <span className="text-rose-600 font-semibold">
                    Disqualified: {auditReport.total_candidates_input - auditReport.total_results_retained}
                  </span>
                </div>
              </div>

              {/* Execution Modes */}
              <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 text-xs">
                <div className="p-2 bg-gray-50 rounded border border-gray-100">
                  <div className="text-[11px] text-gray-500 mb-1">Max Age Filter</div>
                  {renderExecutionBadge(auditReport.filter_evaluations['max_age'] || 'unsupported')}
                  <div className="text-[10px] text-rose-600 mt-1 font-medium">
                    Dropped: {auditReport.disqualified_by_age}
                  </div>
                </div>

                <div className="p-2 bg-gray-50 rounded border border-gray-100">
                  <div className="text-[11px] text-gray-500 mb-1">Work Mode Filter</div>
                  {renderExecutionBadge(auditReport.filter_evaluations['work_mode'] || 'unsupported')}
                  <div className="text-[10px] text-rose-600 mt-1 font-medium">
                    Dropped: {auditReport.disqualified_by_mode}
                  </div>
                </div>

                <div className="p-2 bg-gray-50 rounded border border-gray-100">
                  <div className="text-[11px] text-gray-500 mb-1">Job Type Filter</div>
                  {renderExecutionBadge(auditReport.filter_evaluations['job_type'] || 'unsupported')}
                  <div className="text-[10px] text-rose-600 mt-1 font-medium">
                    Dropped: {auditReport.disqualified_by_type}
                  </div>
                </div>

                <div className="p-2 bg-gray-50 rounded border border-gray-100">
                  <div className="text-[11px] text-gray-500 mb-1">Company Exclusions</div>
                  {renderExecutionBadge(auditReport.filter_evaluations['company_exclusions'] || 'unsupported')}
                  <div className="text-[10px] text-rose-600 mt-1 font-medium">
                    Dropped: {auditReport.disqualified_by_company}
                  </div>
                </div>

                <div className="p-2 bg-gray-50 rounded border border-gray-100">
                  <div className="text-[11px] text-gray-500 mb-1">Salary Floor</div>
                  {renderExecutionBadge(auditReport.filter_evaluations['minimum_salary'] || 'unsupported')}
                  <div className="text-[10px] text-rose-600 mt-1 font-medium">
                    Dropped: {auditReport.disqualified_by_salary}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Discovered Jobs List & Explainable Match Controls (IMP-CAR-09, CAR-09, AT-003, AT-028) */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-base font-bold text-gray-900">
                Discovered Openings ({filteredJobs.length})
              </h2>
              {evaluatingMatches && (
                <span className="text-xs text-blue-600 animate-pulse font-medium">
                  Calculating fit scores...
                </span>
              )}
            </div>
            <span className="text-xs text-gray-500">
              Deduplicated by Canonical URL & Position Fingerprint (AT-004)
            </span>
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            <label className="flex items-center gap-2 text-xs font-semibold text-gray-700 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={enableMatchScoring}
                onChange={(e) => setEnableMatchScoring(e.target.checked)}
                className="rounded text-blue-600 focus:ring-blue-500 h-4 w-4"
              />
              <span>Explainable Fit Scoring (CAR-09)</span>
            </label>

            {enableMatchScoring && (
              <button
                onClick={() => {
                  setWeightsForm({ ...customWeights });
                  setWeightsError(null);
                  setShowWeightsModal(true);
                }}
                className="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium text-xs rounded-lg transition flex items-center gap-1.5"
              >
                <span>⚙️ Weights:</span>
                <span className="font-bold">
                  {customWeights.skills_weight}/{customWeights.title_experience_weight}/{customWeights.location_work_mode_weight}/{customWeights.compensation_weight}
                </span>
              </button>
            )}
          </div>
        </div>

        {filteredJobs.length === 0 ? (
          <div className="p-8 text-center bg-white border border-gray-200 rounded-xl">
            <p className="text-sm text-gray-500">
              No jobs match your search and filter criteria. Try broadening your keywords or clearing active filters.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4">
            {filteredJobs.map((job) => {
              const daysAgo = Math.floor((Date.now() - new Date(job.date_posted).getTime()) / (1000 * 60 * 60 * 24));
              const isHybrid = job.location.raw_location.toLowerCase().includes('hybrid');
              const match = enableMatchScoring ? matchResults[job.id] : undefined;
              const isExpanded = expandedMatchJobId === job.id;

              return (
                <div
                  key={job.id}
                  className={`bg-white border rounded-xl p-5 shadow-sm transition space-y-3 ${
                    match && !match.is_eligible
                      ? 'border-rose-200 bg-rose-50/20'
                      : 'border-gray-200 hover:border-blue-300'
                  }`}
                >
                  <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="text-base font-bold text-gray-900">{job.title}</h3>
                        <span className="text-xs font-semibold text-gray-700 bg-gray-100 px-2 py-0.5 rounded">
                          {job.company}
                        </span>
                        {job.is_direct_employer && (
                          <span className="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-emerald-100 text-emerald-800">
                            Direct ATS
                          </span>
                        )}
                        <span className="text-[10px] uppercase font-semibold px-2 py-0.5 rounded bg-blue-50 text-blue-700">
                          {job.source}
                        </span>
                        <span className="text-[10px] uppercase font-semibold px-2 py-0.5 rounded bg-purple-50 text-purple-700">
                          {job.job_type.replace('_', ' ')}
                        </span>
                      </div>

                      <div className="flex items-center gap-3 text-xs text-gray-500 mt-1.5 flex-wrap">
                        <span>📍 {job.location.raw_location}</span>
                        {job.location.is_remote && (
                          <span className="text-emerald-700 font-semibold bg-emerald-50 px-1.5 py-0.5 rounded">
                            Remote
                          </span>
                        )}
                        {isHybrid && (
                          <span className="text-blue-700 font-semibold bg-blue-50 px-1.5 py-0.5 rounded">
                            Hybrid
                          </span>
                        )}
                        <span>📅 Posted: {daysAgo === 0 ? 'Today' : `${daysAgo}d ago`}</span>
                      </div>
                    </div>

                    {/* Compensation Badge */}
                    <div className="text-right shrink-0">
                      {job.compensation.is_disclosed ? (
                        <div>
                          <span className="text-sm font-bold text-gray-900">
                            ${(job.compensation.min_amount || 0).toLocaleString()} - ${(job.compensation.max_amount || 0).toLocaleString()}
                          </span>
                          <div className="text-[11px] text-gray-500 capitalize">
                            {job.compensation.currency} / {job.compensation.period}
                            {job.compensation.is_estimated && ' (est.)'}
                          </div>
                        </div>
                      ) : (
                        <div className="text-xs text-amber-700 bg-amber-50 px-2 py-1 rounded font-medium">
                          Salary undisclosed by employer
                        </div>
                      )}
                    </div>
                  </div>

                  <p className="text-xs text-gray-600 leading-relaxed line-clamp-2">
                    {job.description}
                  </p>

                  {/* Required Skills */}
                  <div className="flex flex-wrap items-center gap-1.5 pt-1">
                    {job.required_skills.map((skill, idx) => (
                      <span
                        key={idx}
                        className="text-[11px] font-medium bg-gray-100 text-gray-700 px-2 py-0.5 rounded"
                      >
                        {skill}
                      </span>
                    ))}
                  </div>

                  {/* Explainable Fit Scoring Card (CAR-09, AT-003, AT-028) */}
                  {match && (
                    <div className="pt-2 border-t border-gray-100">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-gray-600">Candidate Fit:</span>
                          {!match.is_eligible ? (
                            <span className="text-xs font-bold px-2.5 py-0.5 rounded-full bg-rose-100 text-rose-800 border border-rose-200">
                              ⛔ Ineligible (Hard Gate Barrier)
                            </span>
                          ) : match.fit_tier === 'strong_match' ? (
                            <span className="text-xs font-bold px-2.5 py-0.5 rounded-full bg-emerald-100 text-emerald-800 border border-emerald-200">
                              🟢 {match.overall_score}% Strong Match
                            </span>
                          ) : match.fit_tier === 'moderate_match' ? (
                            <span className="text-xs font-bold px-2.5 py-0.5 rounded-full bg-blue-100 text-blue-800 border border-blue-200">
                              🔵 {match.overall_score}% Moderate Match
                            </span>
                          ) : (
                            <span className="text-xs font-bold px-2.5 py-0.5 rounded-full bg-amber-100 text-amber-800 border border-amber-200">
                              🟠 {match.overall_score}% Low Match
                            </span>
                          )}
                        </div>

                        <button
                          type="button"
                          onClick={() => setExpandedMatchJobId(isExpanded ? null : job.id)}
                          className="text-xs text-blue-600 hover:text-blue-800 font-semibold underline flex items-center gap-1"
                        >
                          <span>{isExpanded ? 'Hide Fit Details ▲' : 'Why this score? (Explainability Audit) ▼'}</span>
                        </button>
                      </div>

                      {/* Expandable Explainability Drawer */}
                      {isExpanded && (
                        <div className="mt-3 p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-4 text-xs">
                          <div className="flex items-center justify-between border-b border-slate-200 pb-2">
                            <div className="font-bold text-gray-900 flex items-center gap-1.5">
                              <span>🔍</span> Explainable Match Audit Report (Algorithm {match.algorithm_version})
                            </div>
                            <div className="text-[11px] text-gray-500">
                              Evaluated: {new Date(match.evaluated_at).toLocaleTimeString()}
                            </div>
                          </div>

                          {/* 1. Hard Eligibility Gates */}
                          <div>
                            <div className="font-semibold text-gray-800 mb-1.5 flex items-center gap-1">
                              <span>🛡️</span> Hard Eligibility Gates (Zero Tolerance)
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
                              {match.hard_gates.map((gate, gIdx) => (
                                <div
                                  key={gIdx}
                                  className={`p-2.5 rounded-lg border ${
                                    !gate.passed
                                      ? 'bg-rose-50 border-rose-200 text-rose-900'
                                      : gate.is_unknown
                                      ? 'bg-amber-50 border-amber-200 text-amber-900'
                                      : 'bg-emerald-50 border-emerald-200 text-emerald-900'
                                  }`}
                                >
                                  <div className="flex items-center justify-between font-bold text-[11px] mb-1">
                                    <span>{gate.gate_name}</span>
                                    <span>
                                      {!gate.passed ? '✕ Disqualified' : gate.is_unknown ? 'ℹ Neutral (AT-028)' : '✓ Passed'}
                                    </span>
                                  </div>
                                  <p className="text-[10px] leading-tight opacity-90">{gate.reason}</p>
                                </div>
                              ))}
                            </div>
                          </div>

                          {/* 2. Dimensional Score Contribution */}
                          <div>
                            <div className="font-semibold text-gray-800 mb-1.5 flex items-center gap-1">
                              <span>📊</span> Multi-Factor Scoring Breakdown (Configurable Weights Σ = 100%)
                            </div>
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2">
                              <div className="p-2.5 bg-white border border-gray-200 rounded-lg">
                                <div className="flex justify-between text-[11px] text-gray-600 mb-1">
                                  <span>Skills ({match.breakdown.weights_used.skills_weight}%)</span>
                                  <span className="font-bold text-gray-900">{match.breakdown.skills_score}/100</span>
                                </div>
                                <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden">
                                  <div
                                    className="bg-blue-600 h-full rounded-full"
                                    style={{ width: `${Math.min(100, match.breakdown.skills_score)}%` }}
                                  />
                                </div>
                                <span className="text-[10px] text-gray-500 mt-1 block">
                                  Points: +{match.breakdown.skills_weighted_score.toFixed(1)}
                                </span>
                              </div>

                              <div className="p-2.5 bg-white border border-gray-200 rounded-lg">
                                <div className="flex justify-between text-[11px] text-gray-600 mb-1">
                                  <span>Experience ({match.breakdown.weights_used.title_experience_weight}%)</span>
                                  <span className="font-bold text-gray-900">{match.breakdown.title_experience_score}/100</span>
                                </div>
                                <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden">
                                  <div
                                    className="bg-indigo-600 h-full rounded-full"
                                    style={{ width: `${Math.min(100, match.breakdown.title_experience_score)}%` }}
                                  />
                                </div>
                                <span className="text-[10px] text-gray-500 mt-1 block">
                                  Points: +{match.breakdown.title_experience_weighted_score.toFixed(1)}
                                </span>
                              </div>

                              <div className="p-2.5 bg-white border border-gray-200 rounded-lg">
                                <div className="flex justify-between text-[11px] text-gray-600 mb-1">
                                  <span>Location/Mode ({match.breakdown.weights_used.location_work_mode_weight}%)</span>
                                  <span className="font-bold text-gray-900">{match.breakdown.location_work_mode_score}/100</span>
                                </div>
                                <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden">
                                  <div
                                    className="bg-emerald-600 h-full rounded-full"
                                    style={{ width: `${Math.min(100, match.breakdown.location_work_mode_score)}%` }}
                                  />
                                </div>
                                <span className="text-[10px] text-gray-500 mt-1 block">
                                  Points: +{match.breakdown.location_work_mode_weighted_score.toFixed(1)}
                                </span>
                              </div>

                              <div className="p-2.5 bg-white border border-gray-200 rounded-lg">
                                <div className="flex justify-between text-[11px] text-gray-600 mb-1">
                                  <span>Compensation ({match.breakdown.weights_used.compensation_weight}%)</span>
                                  <span className="font-bold text-gray-900">{match.breakdown.compensation_score}/100</span>
                                </div>
                                <div className="w-full bg-gray-200 h-1.5 rounded-full overflow-hidden">
                                  <div
                                    className="bg-purple-600 h-full rounded-full"
                                    style={{ width: `${Math.min(100, match.breakdown.compensation_score)}%` }}
                                  />
                                </div>
                                <span className="text-[10px] text-gray-500 mt-1 block">
                                  Points: +{match.breakdown.compensation_weighted_score.toFixed(1)}
                                </span>
                              </div>
                            </div>
                          </div>

                          {/* 3. Skills Gap Analysis (CAR-06) */}
                          <div className="space-y-1.5">
                            <div className="font-semibold text-gray-800 flex items-center gap-1">
                              <span>🎯</span> Skills Gap Analysis (Evidence-Based CAR-06)
                            </div>
                            <div className="flex flex-wrap gap-2 items-center">
                              {match.matched_skills.map((s, idx) => (
                                <span
                                  key={idx}
                                  className="px-2 py-0.5 rounded-md bg-emerald-100 text-emerald-800 text-[11px] font-semibold flex items-center gap-1"
                                >
                                  <span>✓</span> {s}
                                </span>
                              ))}
                              {match.missing_skills.map((s, idx) => (
                                <span
                                  key={idx}
                                  className="px-2 py-0.5 rounded-md bg-amber-100 text-amber-900 text-[11px] font-semibold flex items-center gap-1 border border-amber-200"
                                >
                                  <span>Missing:</span> {s}
                                </span>
                              ))}
                            </div>
                          </div>

                          {/* 4. Auditable Engine Explanations */}
                          <div className="space-y-1 bg-white p-3 rounded-lg border border-gray-200">
                            <div className="font-semibold text-gray-800 text-[11px] mb-1">
                              Audit Trail & Disclosures (AT-003, AT-028)
                            </div>
                            <ul className="list-disc list-inside space-y-0.5 text-gray-600 text-[11px]">
                              {match.explanations.map((exp, eIdx) => (
                                <li key={eIdx}>{exp}</li>
                              ))}
                            </ul>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Actions & Links */}
                  <div className="border-t border-gray-100 pt-3 flex flex-wrap items-center justify-between gap-3 text-xs">
                    <div className="flex items-center gap-3">
                      <a
                        href={job.canonical_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:underline font-medium flex items-center gap-1"
                      >
                        View Canonical Post ↗
                      </a>
                      {job.direct_apply_url && (
                        <a
                          href={job.direct_apply_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-emerald-600 hover:underline font-medium flex items-center gap-1"
                        >
                          Direct ATS Apply Link ↗
                        </a>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => handleToggleSaveJob(job)}
                        className={`px-3 py-1.5 rounded-lg font-semibold text-xs transition flex items-center gap-1 border ${
                          savedJobIds[job.id]
                            ? 'bg-amber-100 border-amber-300 text-amber-900 shadow-sm'
                            : 'bg-white border-gray-300 hover:bg-gray-50 text-gray-700'
                        }`}
                      >
                        <span>{savedJobIds[job.id] ? '⭐' : '☆'}</span>
                        <span>{savedJobIds[job.id] ? 'Saved' : 'Save Job'}</span>
                      </button>

                      <button
                        type="button"
                        onClick={() => handleBlockCompany(job.company)}
                        className="px-2.5 py-1.5 text-gray-400 hover:text-rose-600 font-medium text-xs transition"
                        title={`Block ${job.company} from searches`}
                      >
                        🚫 Block
                      </button>

                      <button
                        onClick={() => handleTailorRedirect(job)}
                        className="px-3.5 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-700 font-semibold rounded-lg transition"
                      >
                        Tailor Resume for this Job →
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* URL Normalizer Tool Card (AT-004) */}
      <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm space-y-3">
        <h3 className="text-sm font-bold text-gray-900 flex items-center gap-2">
          <span>🔗</span> Canonical URL Stripper & Normalizer (AT-004)
        </h3>
        <p className="text-xs text-gray-500">
          Paste any marketing or tracking URL from LinkedIn, Indeed, or Greenhouse to obtain its clean canonical permalink for deduplication.
        </p>

        <form onSubmit={handleNormalizeUrl} className="flex gap-2">
          <input
            type="url"
            value={inputUrl}
            onChange={(e) => setInputUrl(e.target.value)}
            placeholder="https://www.linkedin.com/jobs/view/12345/?trackingId=xyz&refId=feed..."
            className="flex-1 text-xs border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
          />
          <button
            type="submit"
            className="px-4 py-2 bg-gray-900 hover:bg-gray-800 text-white text-xs font-semibold rounded-lg transition"
          >
            Normalize
          </button>
        </form>

        {normalizedUrl && (
          <div className="p-3 bg-gray-50 border border-gray-200 rounded-lg text-xs space-y-1">
            <span className="font-semibold text-gray-700">Canonical URL:</span>
            <div className="font-mono text-[11px] text-blue-700 break-all">{normalizedUrl}</div>
          </div>
        )}
      </div>

      {/* Board Capabilities Modal (AT-010) */}
      {showCapsModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="bg-white rounded-2xl max-w-3xl w-full p-6 space-y-5 shadow-2xl animate-in fade-in">
            <div className="flex items-center justify-between border-b border-gray-200 pb-3">
              <h3 className="text-lg font-bold text-gray-900">
                Job Board Programmatic Capabilities Matrix (AT-010)
              </h3>
              <button
                onClick={() => setShowCapsModal(false)}
                className="text-gray-400 hover:text-gray-600 font-bold"
              >
                ✕
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left">
                <thead className="bg-gray-50 text-gray-600 font-semibold border-b border-gray-200 uppercase tracking-wider">
                  <tr>
                    <th className="py-2.5 px-3">Board Channel</th>
                    <th className="py-2.5 px-3">Type</th>
                    <th className="py-2.5 px-3">Remote Filter</th>
                    <th className="py-2.5 px-3">Salary Filter</th>
                    <th className="py-2.5 px-3">Rate Limit</th>
                    <th className="py-2.5 px-3">Notes</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {capabilities.map((cap) => (
                    <tr key={cap.board_name} className="hover:bg-gray-50">
                      <td className="py-2.5 px-3 font-semibold text-gray-900 capitalize">
                        {cap.display_name}
                      </td>
                      <td className="py-2.5 px-3">
                        <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${
                          cap.is_direct_ats ? 'bg-emerald-100 text-emerald-800' : 'bg-blue-100 text-blue-800'
                        }`}>
                          {cap.is_direct_ats ? 'Direct ATS' : 'Aggregator'}
                        </span>
                      </td>
                      <td className="py-2.5 px-3">
                        {cap.supports_remote_filter ? (
                          <span className="text-emerald-600 font-bold">✓ Yes</span>
                        ) : (
                          <span className="text-gray-400">✗ No</span>
                        )}
                      </td>
                      <td className="py-2.5 px-3">
                        {cap.supports_salary_filter ? (
                          <span className="text-emerald-600 font-bold">✓ Yes</span>
                        ) : (
                          <span className="text-gray-400">✗ No (Post-filtered)</span>
                        )}
                      </td>
                      <td className="py-2.5 px-3 text-gray-600">
                        {cap.rate_limit_per_minute}/min
                      </td>
                      <td className="py-2.5 px-3 text-gray-500 max-w-xs">
                        {cap.notes}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="border-t border-gray-100 pt-3 flex justify-end">
              <button
                onClick={() => setShowCapsModal(false)}
                className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-800 font-semibold text-xs rounded-lg transition"
              >
                Close Matrix
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Configurable Weights Modal (CAR-09) */}
      {showWeightsModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="bg-white rounded-2xl max-w-md w-full p-6 space-y-5 shadow-2xl animate-in fade-in">
            <div className="flex items-center justify-between border-b border-gray-200 pb-3">
              <h3 className="text-base font-bold text-gray-900 flex items-center gap-1.5">
                <span>⚙️</span> Configure Scoring Dimension Weights (CAR-09)
              </h3>
              <button
                onClick={() => setShowWeightsModal(false)}
                className="text-gray-400 hover:text-gray-600 font-bold"
              >
                ✕
              </button>
            </div>

            <p className="text-xs text-gray-500">
              Customize how heavily each factor influences the total fit score.
              The sum of all four dimensions must strictly equal 100%.
            </p>

            <form onSubmit={handleSaveWeights} className="space-y-4 text-xs">
              <div>
                <div className="flex justify-between font-semibold text-gray-700 mb-1">
                  <label>Confirmed Skills Weight (%)</label>
                  <span className="font-bold text-blue-600">{weightsForm.skills_weight}%</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max="100"
                  value={weightsForm.skills_weight}
                  onChange={(e) =>
                    setWeightsForm({ ...weightsForm, skills_weight: Number(e.target.value) })
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              <div>
                <div className="flex justify-between font-semibold text-gray-700 mb-1">
                  <label>Title & Experience Weight (%)</label>
                  <span className="font-bold text-indigo-600">{weightsForm.title_experience_weight}%</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max="100"
                  value={weightsForm.title_experience_weight}
                  onChange={(e) =>
                    setWeightsForm({ ...weightsForm, title_experience_weight: Number(e.target.value) })
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              <div>
                <div className="flex justify-between font-semibold text-gray-700 mb-1">
                  <label>Location & Work Mode Weight (%)</label>
                  <span className="font-bold text-emerald-600">{weightsForm.location_work_mode_weight}%</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max="100"
                  value={weightsForm.location_work_mode_weight}
                  onChange={(e) =>
                    setWeightsForm({ ...weightsForm, location_work_mode_weight: Number(e.target.value) })
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              <div>
                <div className="flex justify-between font-semibold text-gray-700 mb-1">
                  <label>Compensation Satisfaction Weight (%)</label>
                  <span className="font-bold text-purple-600">{weightsForm.compensation_weight}%</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max="100"
                  value={weightsForm.compensation_weight}
                  onChange={(e) =>
                    setWeightsForm({ ...weightsForm, compensation_weight: Number(e.target.value) })
                  }
                  className="w-full border border-gray-300 rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>

              {/* Live Sum Calculation Indicator */}
              {(() => {
                const sum =
                  Number(weightsForm.skills_weight) +
                  Number(weightsForm.title_experience_weight) +
                  Number(weightsForm.location_work_mode_weight) +
                  Number(weightsForm.compensation_weight);
                const isValid = sum === 100;
                return (
                  <div
                    className={`p-2.5 rounded-lg border flex items-center justify-between text-xs font-semibold ${
                      isValid
                        ? 'bg-emerald-50 border-emerald-200 text-emerald-800'
                        : 'bg-rose-50 border-rose-200 text-rose-800'
                    }`}
                  >
                    <span>Total Sum:</span>
                    <span>{sum}% / 100% {isValid ? '✓ Valid' : '✕ Must equal 100%'}</span>
                  </div>
                );
              })()}

              {weightsError && (
                <div className="p-2 bg-rose-50 text-rose-700 rounded text-xs">
                  {weightsError}
                </div>
              )}

              <div className="border-t border-gray-100 pt-3 flex items-center justify-between gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setWeightsForm({
                      skills_weight: 40,
                      title_experience_weight: 30,
                      location_work_mode_weight: 15,
                      compensation_weight: 15,
                    });
                    setWeightsError(null);
                  }}
                  className="text-xs text-gray-500 hover:text-gray-700 underline"
                >
                  Reset Defaults (40/30/15/15)
                </button>

                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => setShowWeightsModal(false)}
                    className="px-3.5 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg font-medium"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold shadow-sm"
                  >
                    Save & Recalculate
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
