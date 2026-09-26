'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import {
  User,
  Briefcase,
  GraduationCap,
  Code,
  FolderGit2,
  Award,
  Globe2,
  Link as LinkIcon,
  Shield,
  Clock,
  Plus,
  Edit2,
  Trash2,
  CheckCircle2,
  AlertTriangle,
  ChevronRight,
  ArrowLeft,
  Save,
  Sliders,
  FileText,
  Sparkles,
} from 'lucide-react';
import type {
  MasterCareerProfile,
  ContactInfo,
  ExperienceItem,
  EducationItem,
  SkillItem,
  ProjectItem,
  CertificateItem,
  LanguageItem,
  ProfileLink,
  UserCareerConsent,
  CareerPrivacySettings,
} from '@social-platform/contracts';

const initialProfile: MasterCareerProfile = {
  id: '',
  user_id: 'user-current',
  contact: {
    full_name: '',
    email: '',
    phone: '',
    location: '',
    headline: '',
    summary: '',
    citizenship: '',
    work_authorization: '',
  },
  links: [],
  experiences: [],
  education: [],
  skills: [],
  projects: [],
  certificates: [],
  languages: [],
  consent: {
    data_processing_consent: true,
    job_matching_consent: true,
    third_party_sharing_consent: false,
    consent_timestamp: '2026-01-01T00:00:00.000Z',
    consent_version: 'v1.0',
  },
  privacy: {
    profile_visibility: 'owner_private',
    allow_recruiter_view: false,
    share_with_workspace_admin: false,
    anonymize_before_job_search: false,
  },
  audit_trail: [],
  completeness: {
    score: 0,
    missing_sections: ['contact', 'experiences', 'education', 'skills'],
    populated_count: 0,
    total_sections_count: 8,
  },
  created_at: '2026-01-01T00:00:00.000Z',
  updated_at: '2026-01-01T00:00:00.000Z',
};

type ActiveSection = 'contact' | 'experiences' | 'education' | 'skills' | 'projects' | 'certificates' | 'languages' | 'links' | 'privacy' | 'audit';

export default function MasterCareerProfilePage() {
  const [profile, setProfile] = useState<MasterCareerProfile>(initialProfile);
  const [activeSection, setActiveSection] = useState<ActiveSection>('contact');
  const [saveStatus, setSaveStatus] = useState<string | null>(null);

  // Edit States
  const [contactForm, setContactForm] = useState<ContactInfo>(profile.contact);
  const [isEditingContact, setIsEditingContact] = useState(false);
  const [mounted, setMounted] = useState(false);

  // Load confirmed/staged profile from localStorage
  useEffect(() => {
    setMounted(true);
    if (typeof window === 'undefined') return;

    // Automatic purge of past browser test session data
    if (!localStorage.getItem('platform_v2_purged')) {
      localStorage.removeItem('candidate_master_profile');
      localStorage.removeItem('candidate_staged_profile');
      localStorage.removeItem('candidate_confirmed_facts');
      localStorage.removeItem('candidate_generated_resumes');
      localStorage.removeItem('tailored_resumes_history');
      localStorage.removeItem('tailor_target_job');
      localStorage.removeItem('career_onboarding_deferred');
      localStorage.removeItem('extracted_resume_cache');
      localStorage.setItem('platform_v2_purged', 'true');
      setProfile(initialProfile);
      setContactForm(initialProfile.contact);
      return;
    }

    const saved = localStorage.getItem('candidate_master_profile') || localStorage.getItem('candidate_staged_profile');
    if (saved) {
      try {
        const data = JSON.parse(saved);
        const contactData: ContactInfo = data.contact || {
          full_name: data.fullName || '',
          email: data.email || '',
          phone: data.phone || '',
          location: data.location || '',
          headline: data.headline || '',
          summary: data.summary || '',
          citizenship: '',
          work_authorization: '',
        };

        const skillsData: SkillItem[] = (data.skills && data.skills.length > 0)
          ? (typeof data.skills[0] === 'string'
              ? (data.skills as string[]).map((s: string, idx: number) => ({
                  id: `sk-${idx + 1}`,
                  name: s,
                  category: 'technical',
                  proficiency: 'advanced',
                  years_of_experience: 1,
                  confirmed: true,
                }))
              : data.skills)
          : [];

        const expData: ExperienceItem[] = (data.experiences && data.experiences.length > 0)
          ? data.experiences.map((exp: any, idx: number) => ({
              id: exp.id || `exp-${idx + 1}`,
              title: exp.title || '',
              company: exp.company || '',
              location: exp.location || '',
              start_date: exp.start_date || '',
              end_date: exp.end_date || '',
              is_current: exp.is_current ?? true,
              description: exp.description || '',
              highlights: exp.highlights || [],
              skills_used: exp.skills || exp.skills_used || [],
              confirmed: true,
            }))
          : [];

        const eduData: EducationItem[] = (data.education && data.education.length > 0)
          ? data.education.map((edu: any, idx: number) => ({
              id: edu.id || `edu-${idx + 1}`,
              institution: edu.institution || '',
              degree: edu.degree || '',
              field_of_study: edu.field_of_study || '',
              start_date: edu.start_date || '',
              end_date: edu.end_date || edu.year || '',
              grade: edu.grade || '',
              highlights: [],
              confirmed: true,
            }))
          : [];

        const certificatesData: CertificateItem[] = (data.certificates && data.certificates.length > 0)
          ? data.certificates.map((cert: any, idx: number) => ({
              id: cert.id || `crt-${idx + 1}`,
              name: cert.name || '',
              issuer: cert.issuer || '',
              issue_date: cert.issue_date || '',
              credential_id: cert.credential_id || '',
              confirmed: true,
            }))
          : [];

        const projectsData: ProjectItem[] = (data.projects && data.projects.length > 0)
          ? data.projects
          : [];

        const linksData: ProfileLink[] = (data.links && data.links.length > 0)
          ? data.links
          : [];

        const auditData = (data.audit_trail && data.audit_trail.length > 0)
          ? data.audit_trail
          : [];

        const updatedProfile: MasterCareerProfile = {
          ...initialProfile,
          id: data.id || `prof-${Date.now()}`,
          contact: contactData,
          skills: skillsData,
          experiences: expData,
          education: eduData,
          certificates: certificatesData,
          projects: projectsData,
          links: linksData,
          audit_trail: auditData,
          completeness: {
            score: (contactData.full_name ? 25 : 0) + (expData.length > 0 ? 25 : 0) + (eduData.length > 0 ? 25 : 0) + (skillsData.length > 0 ? 25 : 0),
            missing_sections: [],
            populated_count: [contactData.full_name, expData.length, eduData.length, skillsData.length].filter(Boolean).length,
            total_sections_count: 8,
          },
          updated_at: new Date().toISOString(),
        };

        setProfile(updatedProfile);
        setContactForm(contactData);
      } catch (err) {
        console.error('Failed to parse candidate_master_profile:', err);
      }
    }
  }, []);

  // Experience modal/add state
  const [showAddExp, setShowAddExp] = useState(false);
  const [expForm, setExpForm] = useState<Partial<ExperienceItem>>({
    title: '',
    company: '',
    location: '',
    is_current: true,
    highlights: [],
    skills_used: [],
    confirmed: true,
  });

  // Skill add state
  const [showAddSkill, setShowAddSkill] = useState(false);
  const [skillForm, setSkillForm] = useState<Partial<SkillItem>>({
    name: '',
    category: 'technical',
    proficiency: 'advanced',
    years_of_experience: 3,
    confirmed: true,
  });

  const handleSaveContact = () => {
    setProfile((prev) => {
      const updated = {
        ...prev,
        contact: contactForm,
        audit_trail: [
          {
            id: `aud_${Date.now()}`,
            field: 'contact',
            old_value: prev.contact.full_name,
            new_value: contactForm.full_name,
            changed_by: 'user-current',
            changed_at: new Date().toISOString(),
            reason: 'User manual profile update',
          },
          ...prev.audit_trail,
        ],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    setIsEditingContact(false);
    showNotice('Contact details updated successfully.');
  };

  // Education modal/add state
  const [showAddEdu, setShowAddEdu] = useState(false);
  const [eduForm, setEduForm] = useState<Partial<EducationItem>>({
    institution: '',
    degree: '',
    field_of_study: '',
    start_date: '',
    end_date: '',
    grade: '',
    confirmed: true,
  });

  const handleAddExperience = (e: React.FormEvent) => {
    e.preventDefault();
    if (!expForm.title || !expForm.company) return;

    const newItem: ExperienceItem = {
      id: `exp_${Date.now()}`,
      title: expForm.title,
      company: expForm.company,
      location: expForm.location || 'Remote',
      is_current: expForm.is_current ?? true,
      highlights: expForm.highlights || [],
      skills_used: expForm.skills_used || [],
      confirmed: true,
    };

    setProfile((prev) => {
      const updated = {
        ...prev,
        experiences: [newItem, ...prev.experiences],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    setShowAddExp(false);
    setExpForm({ title: '', company: '', location: '', is_current: true, highlights: [], skills_used: [], confirmed: true });
    showNotice('Work experience added to master profile.');
  };

  const handleDeleteExperience = (id: string) => {
    setProfile((prev) => {
      const updated = {
        ...prev,
        experiences: prev.experiences.filter((x) => x.id !== id),
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    showNotice('Experience item removed.');
  };

  const handleAddEducation = (e: React.FormEvent) => {
    e.preventDefault();
    if (!eduForm.institution || !eduForm.degree) return;

    const newItem: EducationItem = {
      id: `edu_${Date.now()}`,
      institution: eduForm.institution,
      degree: eduForm.degree,
      field_of_study: eduForm.field_of_study || 'General Studies',
      start_date: eduForm.start_date || '2016-01-01',
      end_date: eduForm.end_date || '2020-01-01',
      grade: eduForm.grade || '',
      highlights: [],
      confirmed: true,
    };

    setProfile((prev) => {
      const updated = {
        ...prev,
        education: [newItem, ...prev.education],
        audit_trail: [
          {
            id: `aud_${Date.now()}`,
            field: 'education',
            old_value: 'Added new degree',
            new_value: `${newItem.degree} from ${newItem.institution}`,
            changed_by: 'user-current',
            changed_at: new Date().toISOString(),
            reason: 'User added education record',
          },
          ...prev.audit_trail,
        ],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    setShowAddEdu(false);
    setEduForm({ institution: '', degree: '', field_of_study: '', start_date: '', end_date: '', grade: '', confirmed: true });
    showNotice('Education record added to master profile.');
  };

  const handleDeleteEducation = (id: string) => {
    setProfile((prev) => {
      const updated = {
        ...prev,
        education: prev.education.filter((x) => x.id !== id),
        audit_trail: [
          {
            id: `aud_${Date.now()}`,
            field: 'education',
            old_value: 'Removed degree record',
            new_value: id,
            changed_by: 'user-current',
            changed_at: new Date().toISOString(),
            reason: 'User deleted education record',
          },
          ...prev.audit_trail,
        ],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    showNotice('Education item removed.');
  };

  const handleAddSkill = (e: React.FormEvent) => {
    e.preventDefault();
    if (!skillForm.name) return;

    const newSkill: SkillItem = {
      id: `sk_${Date.now()}`,
      name: skillForm.name,
      category: skillForm.category || 'technical',
      proficiency: skillForm.proficiency || 'intermediate',
      years_of_experience: skillForm.years_of_experience || 1,
      confirmed: true,
    };

    setProfile((prev) => {
      const updated = {
        ...prev,
        skills: [...prev.skills, newSkill],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    setShowAddSkill(false);
    setSkillForm({ name: '', category: 'technical', proficiency: 'advanced', years_of_experience: 3, confirmed: true });
    showNotice(`Skill "${newSkill.name}" registered.`);
  };

  const handleDeleteSkill = (id: string) => {
    setProfile((prev) => {
      const updated = {
        ...prev,
        skills: prev.skills.filter((s) => s.id !== id),
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    showNotice('Skill removed.');
  };

  // Certificate modal/add state
  const [showAddCert, setShowAddCert] = useState(false);
  const [certForm, setCertForm] = useState<Partial<CertificateItem>>({
    name: '',
    issuer: '',
    issue_date: '',
    credential_id: '',
    confirmed: true,
  });

  const handleAddCertificate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!certForm.name || !certForm.issuer) return;

    const newItem: CertificateItem = {
      id: `crt_${Date.now()}`,
      name: certForm.name,
      issuer: certForm.issuer,
      issue_date: certForm.issue_date || new Date().toISOString().slice(0, 10),
      credential_id: certForm.credential_id || `CERT-${Math.floor(100000 + Math.random() * 900000)}`,
      confirmed: true,
    };

    setProfile((prev) => {
      const updated = {
        ...prev,
        certificates: [newItem, ...prev.certificates],
        audit_trail: [
          {
            id: `aud_${Date.now()}`,
            field: 'certificates',
            old_value: 'Added new certificate',
            new_value: `${newItem.name} by ${newItem.issuer}`,
            changed_by: 'user-current',
            changed_at: new Date().toISOString(),
            reason: 'User added certificate record',
          },
          ...prev.audit_trail,
        ],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    setShowAddCert(false);
    setCertForm({ name: '', issuer: '', issue_date: '', credential_id: '', confirmed: true });
    showNotice('Certificate record added successfully.');
  };

  const handleDeleteCertificate = (id: string) => {
    setProfile((prev) => {
      const updated = {
        ...prev,
        certificates: prev.certificates.filter((x) => x.id !== id),
        audit_trail: [
          {
            id: `aud_${Date.now()}`,
            field: 'certificates',
            old_value: 'Removed certificate',
            new_value: id,
            changed_by: 'user-current',
            changed_at: new Date().toISOString(),
            reason: 'User deleted certificate record',
          },
          ...prev.audit_trail,
        ],
        updated_at: new Date().toISOString(),
      };
      if (typeof window !== 'undefined') {
        localStorage.setItem('candidate_master_profile', JSON.stringify(updated));
      }
      return updated;
    });
    showNotice('Certificate removed.');
  };

  const handleToggleAdminSharing = (checked: boolean) => {
    setProfile((prev) => ({
      ...prev,
      privacy: {
        ...prev.privacy,
        share_with_workspace_admin: checked,
      },
      audit_trail: [
        {
          id: `aud_${Date.now()}`,
          field: 'privacy.share_with_workspace_admin',
          old_value: String(prev.privacy.share_with_workspace_admin),
          new_value: String(checked),
          changed_by: 'user-current',
          changed_at: new Date().toISOString(),
          reason: checked ? 'Explicit user consent granted (REQ-018)' : 'Explicit user revocation (REQ-018)',
        },
        ...prev.audit_trail,
      ],
      updated_at: new Date().toISOString(),
    }));
    showNotice(checked ? 'Workspace Admin access GRANTED for personal profile.' : 'Workspace Admin access REVOKED (REQ-018: Private Personal).');
  };

  const handleToggleConsent = (key: keyof UserCareerConsent, value: boolean) => {
    setProfile((prev) => ({
      ...prev,
      consent: {
        ...prev.consent,
        [key]: value,
        consent_timestamp: new Date().toISOString(),
      },
      updated_at: new Date().toISOString(),
    }));
    showNotice('Consent preferences updated.');
  };

  const showNotice = (msg: string) => {
    setSaveStatus(msg);
    setTimeout(() => setSaveStatus(null), 4000);
  };

  const handleResetToFresh = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('candidate_master_profile');
      localStorage.removeItem('candidate_staged_profile');
      localStorage.removeItem('candidate_confirmed_facts');
      localStorage.removeItem('candidate_generated_resumes');
      localStorage.removeItem('tailored_resumes_history');
      localStorage.removeItem('tailor_target_job');
      localStorage.removeItem('career_onboarding_deferred');
      localStorage.removeItem('extracted_resume_cache');
    }
    setProfile(initialProfile);
    setContactForm(initialProfile.contact);
    showNotice('All uploaded and cached profile data removed. Fresh copy created.');
  };

  return (
    <div className="space-y-8 pb-12">
      {/* Top Breadcrumb & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-4">
        <div className="flex items-center gap-3">
          <Link
            href="/career"
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 transition"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Back to Career Hub</span>
          </Link>
          <span className="text-slate-300 dark:text-slate-700">•</span>
          <span className="text-xs font-medium text-blue-600 dark:text-blue-400">Master Profile (REQ-002, REQ-003)</span>
        </div>

        <div className="flex items-center gap-2.5 flex-wrap">
          <button
            type="button"
            onClick={handleResetToFresh}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-rose-200 dark:border-rose-900/50 bg-rose-50 dark:bg-rose-950/30 hover:bg-rose-100 text-rose-700 dark:text-rose-300 rounded-lg text-xs font-semibold transition cursor-pointer"
            title="Wipe all uploaded and cached data to create a fresh copy"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>Reset to Fresh</span>
          </button>
          <Link
            href="/career/preferences"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg text-xs font-semibold transition cursor-pointer"
          >
            <Sliders className="w-3.5 h-3.5" />
            <span>Preferences</span>
          </Link>
          <Link
            href="/career/resumes"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-indigo-200 dark:border-indigo-800 bg-indigo-50 dark:bg-indigo-950/40 hover:bg-indigo-100 text-indigo-700 dark:text-indigo-300 rounded-lg text-xs font-semibold transition cursor-pointer"
          >
            <FileText className="w-3.5 h-3.5" />
            <span>Master Resumes</span>
          </Link>
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <Clock className="w-3.5 h-3.5" />
            <span suppressHydrationWarning>
              Last Updated: {mounted && profile.updated_at && profile.updated_at !== '2026-01-01T00:00:00.000Z'
                ? new Date(profile.updated_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                : 'Recently'}
            </span>
          </div>
        </div>
      </div>

      {/* Notification Toast */}
      {saveStatus && (
        <div className="p-3.5 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-300 rounded-xl text-xs font-medium flex items-center justify-between animate-fadeIn">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600" />
            <span>{saveStatus}</span>
          </div>
        </div>
      )}

      {/* Profile Overview Card & Completeness Meter */}
      <div className="bg-gradient-to-r from-blue-900/10 via-indigo-900/10 to-slate-900/10 dark:from-blue-950/30 dark:via-indigo-950/30 dark:to-slate-950/30 border border-blue-200/60 dark:border-blue-800/40 rounded-2xl p-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-1.5">
            <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-300">
              <Shield className="w-3 h-3" />
              <span>Canonical User-Confirmed Profile</span>
            </div>
            <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">
              {profile.contact.full_name || 'No Profile Created Yet'}
            </h1>
            <p className="text-sm font-medium text-slate-600 dark:text-slate-300">
              {profile.contact.headline || 'No headline configured — Upload a resume to populate'}
            </p>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {(profile.contact.location || profile.contact.email) ? (
                <>
                  {profile.contact.location}
                  {profile.contact.location && profile.contact.email ? ' • ' : ''}
                  {profile.contact.email}
                </>
              ) : (
                'No contact details registered'
              )}
            </p>
          </div>

          <div className="bg-white/80 dark:bg-slate-900/80 backdrop-blur border border-slate-200 dark:border-slate-800 rounded-xl p-4 min-w-[220px]">
            <div className="flex items-center justify-between text-xs font-semibold mb-1.5">
              <span className="text-slate-600 dark:text-slate-400">Profile Completeness</span>
              <span className="text-blue-600 dark:text-blue-400">{profile.completeness.score.toFixed(0)}%</span>
            </div>
            <div className="w-full bg-slate-100 dark:bg-slate-800 h-2 rounded-full overflow-hidden">
              <div
                className="bg-gradient-to-r from-blue-600 to-indigo-600 h-full rounded-full transition-all duration-500"
                style={{ width: `${profile.completeness.score}%` }}
              />
            </div>
            <div className="text-[11px] text-slate-500 mt-2 flex items-center gap-1">
              <CheckCircle2 className="w-3 h-3 text-emerald-500" />
              <span>{profile.completeness.populated_count} of {profile.completeness.total_sections_count} sections populated</span>
            </div>
          </div>
        </div>
      </div>

      {/* Strict Privacy Isolation Notice (REQ-018) */}
      <div className="p-4 rounded-xl border border-amber-200/80 dark:border-amber-900/50 bg-amber-50/70 dark:bg-amber-950/20 flex items-start gap-3 text-xs text-amber-800 dark:text-amber-300">
        <AlertTriangle className="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
        <div className="space-y-1 leading-relaxed">
          <p className="font-bold">Strict Private Profile Boundary (REQ-018 & AT-011)</p>
          <p>
            Your master career profile, resume drafts, and extraction facts are strictly private to your personal user account.
            Workspace Admins cannot view or access your resume unless you explicitly enable sharing under Privacy Settings.
          </p>
        </div>
      </div>

      {/* Main Content Grid with Side Tabs */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Navigation Sidebar */}
        <div className="lg:col-span-1 space-y-1 bg-white dark:bg-slate-900 p-2 border border-slate-200 dark:border-slate-800 rounded-xl h-fit">
          {[
            { id: 'contact', label: 'Contact & Bio', icon: User, count: profile.contact.email ? '1' : '0' },
            { id: 'experiences', label: 'Work Experience', icon: Briefcase, count: profile.experiences.length },
            { id: 'education', label: 'Education', icon: GraduationCap, count: profile.education.length },
            { id: 'skills', label: 'Skills & Competencies', icon: Code, count: profile.skills.length },
            { id: 'projects', label: 'Projects', icon: FolderGit2, count: profile.projects.length },
            { id: 'certificates', label: 'Certifications', icon: Award, count: profile.certificates.length },
            { id: 'languages', label: 'Languages', icon: Globe2, count: profile.languages.length },
            { id: 'links', label: 'Links & Profiles', icon: LinkIcon, count: profile.links.length },
            { id: 'privacy', label: 'Privacy & Consent', icon: Shield, count: profile.privacy.share_with_workspace_admin ? 'Shared' : 'Private' },
            { id: 'audit', label: 'Field Audit Trail', icon: Clock, count: profile.audit_trail.length },
          ].map((item) => {
            const Icon = item.icon;
            const isActive = activeSection === item.id;
            return (
              <button
                key={item.id}
                type="button"
                onClick={() => setActiveSection(item.id as ActiveSection)}
                className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-xs font-semibold transition cursor-pointer ${
                  isActive
                    ? 'bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 font-bold'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
                }`}
              >
                <div className="flex items-center gap-2.5">
                  <Icon className={`w-4 h-4 ${isActive ? 'text-blue-600 dark:text-blue-400' : 'text-slate-400'}`} />
                  <span>{item.label}</span>
                </div>
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 font-mono">
                  {item.count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Section Detail Panel */}
        <div className="lg:col-span-3 space-y-6">
          {/* 1. Contact & Bio */}
          {activeSection === 'contact' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">Contact & Biographical Details</h2>
                  <p className="text-xs text-slate-500">Core personal facts verified by candidate.</p>
                </div>
                {!isEditingContact && (
                  <button
                    type="button"
                    onClick={() => setIsEditingContact(true)}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-slate-300 dark:border-slate-700 rounded-lg text-xs font-semibold text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-800 transition"
                  >
                    <Edit2 className="w-3.5 h-3.5" />
                    <span>Edit Details</span>
                  </button>
                )}
              </div>

              {isEditingContact ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Full Name</label>
                      <input
                        type="text"
                        value={contactForm.full_name}
                        onChange={(e) => setContactForm({ ...contactForm, full_name: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Email</label>
                      <input
                        type="email"
                        value={contactForm.email}
                        onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Phone</label>
                      <input
                        type="text"
                        value={contactForm.phone || ''}
                        onChange={(e) => setContactForm({ ...contactForm, phone: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Location</label>
                      <input
                        type="text"
                        value={contactForm.location || ''}
                        onChange={(e) => setContactForm({ ...contactForm, location: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Headline</label>
                      <input
                        type="text"
                        value={contactForm.headline || ''}
                        onChange={(e) => setContactForm({ ...contactForm, headline: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Citizenship</label>
                      <input
                        type="text"
                        value={contactForm.citizenship || ''}
                        onChange={(e) => setContactForm({ ...contactForm, citizenship: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Summary / Bio</label>
                    <textarea
                      rows={3}
                      value={contactForm.summary || ''}
                      onChange={(e) => setContactForm({ ...contactForm, summary: e.target.value })}
                      className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-transparent text-slate-900 dark:text-white"
                    />
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={() => setIsEditingContact(false)}
                      className="px-3 py-1.5 text-xs text-slate-500 hover:text-slate-700 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      onClick={handleSaveContact}
                      className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold transition"
                    >
                      <Save className="w-3.5 h-3.5" />
                      <span>Save Changes</span>
                    </button>
                  </div>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
                    <span className="text-slate-400 block mb-0.5">Full Name</span>
                    <span className="font-semibold text-slate-900 dark:text-white">{profile.contact.full_name || '—'}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
                    <span className="text-slate-400 block mb-0.5">Email</span>
                    <span className="font-semibold text-slate-900 dark:text-white">{profile.contact.email || '—'}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
                    <span className="text-slate-400 block mb-0.5">Phone</span>
                    <span className="font-semibold text-slate-900 dark:text-white">{profile.contact.phone || '—'}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
                    <span className="text-slate-400 block mb-0.5">Location</span>
                    <span className="font-semibold text-slate-900 dark:text-white">{profile.contact.location || '—'}</span>
                  </div>
                  <div className="col-span-full p-3 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
                    <span className="text-slate-400 block mb-0.5">Summary</span>
                    <p className="text-slate-700 dark:text-slate-300 leading-relaxed">{profile.contact.summary || 'No summary set.'}</p>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 2. Work Experience */}
          {activeSection === 'experiences' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">Work Experience</h2>
                  <p className="text-xs text-slate-500">Confirmed positions and professional achievements.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setShowAddExp(true)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold transition cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Add Experience</span>
                </button>
              </div>

              {showAddExp && (
                <form onSubmit={handleAddExperience} className="p-4 border border-blue-200 dark:border-blue-900/40 rounded-xl bg-blue-50/40 dark:bg-blue-950/20 space-y-3">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Job Title *</label>
                      <input
                        type="text"
                        required
                        value={expForm.title}
                        onChange={(e) => setExpForm({ ...expForm, title: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Lead Systems Engineer"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Company *</label>
                      <input
                        type="text"
                        required
                        value={expForm.company}
                        onChange={(e) => setExpForm({ ...expForm, company: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. CloudScale Inc"
                      />
                    </div>
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAddExp(false)}
                      className="px-3 py-1.5 text-xs text-slate-500 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold"
                    >
                      Save Position
                    </button>
                  </div>
                </form>
              )}

              {profile.experiences.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <Briefcase className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No work experience added yet.</p>
                  <p className="text-[11px] text-slate-400">Upload your resume in Career Hub or click &quot;Add Experience&quot; to manually record positions.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {profile.experiences.map((exp) => (
                    <div key={exp.id} className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 transition space-y-2">
                      <div className="flex items-start justify-between">
                        <div>
                          <h3 className="text-sm font-bold text-slate-900 dark:text-white">{exp.title}</h3>
                          <p className="text-xs font-medium text-slate-600 dark:text-slate-300">
                            {exp.company} • {exp.location || 'Remote'}
                          </p>
                        </div>
                        <button
                          type="button"
                          onClick={() => handleDeleteExperience(exp.id)}
                          className="p-1 text-slate-400 hover:text-rose-600 transition"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                      {exp.description && (
                        <p className="text-xs text-slate-600 dark:text-slate-400">{exp.description}</p>
                      )}
                      {exp.highlights && exp.highlights.length > 0 && (
                        <ul className="list-disc list-inside text-xs text-slate-600 dark:text-slate-400 space-y-1 pl-1">
                          {exp.highlights.map((h, i) => (
                            <li key={i}>{h}</li>
                          ))}
                        </ul>
                      )}
                      {exp.skills_used && exp.skills_used.length > 0 && (
                        <div className="flex flex-wrap gap-1.5 pt-1">
                          {exp.skills_used.map((s, idx) => (
                            <span key={idx} className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                              {s}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 3. Skills */}
          {activeSection === 'skills' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">Skills & Competencies</h2>
                  <p className="text-xs text-slate-500">Categorized proficiencies with confirmed status.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setShowAddSkill(true)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold transition cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Add Skill</span>
                </button>
              </div>

              {showAddSkill && (
                <form onSubmit={handleAddSkill} className="p-4 border border-blue-200 dark:border-blue-900/40 rounded-xl bg-blue-50/40 dark:bg-blue-950/20 space-y-3">
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Skill Name *</label>
                      <input
                        type="text"
                        required
                        value={skillForm.name}
                        onChange={(e) => setSkillForm({ ...skillForm, name: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Technical SEO"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Category</label>
                      <select
                        value={skillForm.category}
                        onChange={(e) => setSkillForm({ ...skillForm, category: e.target.value as any })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                      >
                        <option value="technical">Technical</option>
                        <option value="tool">Tool / Infrastructure</option>
                        <option value="domain">Domain Knowledge</option>
                        <option value="soft">Soft Skill</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Proficiency</label>
                      <select
                        value={skillForm.proficiency}
                        onChange={(e) => setSkillForm({ ...skillForm, proficiency: e.target.value as any })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                      >
                        <option value="beginner">Beginner</option>
                        <option value="intermediate">Intermediate</option>
                        <option value="advanced">Advanced</option>
                        <option value="expert">Expert</option>
                      </select>
                    </div>
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAddSkill(false)}
                      className="px-3 py-1.5 text-xs text-slate-500 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold"
                    >
                      Save Skill
                    </button>
                  </div>
                </form>
              )}

              {profile.skills.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <Sparkles className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No skills recorded yet.</p>
                  <p className="text-[11px] text-slate-400">Upload your resume in Career Hub or click &quot;Add Skill&quot; to add your proficiencies.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {profile.skills.map((skill) => (
                    <div
                      key={skill.id}
                      className="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-bold text-slate-900 dark:text-white">{skill.name}</span>
                          <span className="px-1.5 py-0.2 rounded text-[9px] font-semibold uppercase tracking-wider bg-blue-50 dark:bg-blue-950/80 text-blue-700 dark:text-blue-300">
                            {skill.proficiency}
                          </span>
                        </div>
                        <span className="text-[11px] text-slate-500 capitalize">{skill.category} • {skill.years_of_experience || 1}+ yrs</span>
                      </div>
                      <button
                        type="button"
                        onClick={() => handleDeleteSkill(skill.id)}
                        className="p-1 text-slate-400 hover:text-rose-600 transition"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 4. Privacy & Consent Settings (REQ-018) */}
          {activeSection === 'privacy' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="border-b border-slate-100 dark:border-slate-800 pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">Privacy & Workspace Isolation Settings</h2>
                <p className="text-xs text-slate-500">Enforce REQ-018 private boundary controls and processing consents.</p>
              </div>

              <div className="space-y-4 text-xs">
                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between">
                  <div className="space-y-1 max-w-md">
                    <span className="font-bold text-slate-900 dark:text-white block">
                      Share Profile with Workspace Admins (REQ-018)
                    </span>
                    <p className="text-slate-500">
                      When disabled (recommended default), workspace administrators CANNOT view, export, or access your career profile or resume drafts.
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={profile.privacy.share_with_workspace_admin}
                    onChange={(e) => handleToggleAdminSharing(e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500 cursor-pointer"
                  />
                </div>

                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between">
                  <div className="space-y-1 max-w-md">
                    <span className="font-bold text-slate-900 dark:text-white block">
                      Data Processing Consent (REQ-023)
                    </span>
                    <p className="text-slate-500">
                      Consent to parsing resume text, indexing skills, and staging fact drafts.
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={profile.consent.data_processing_consent}
                    onChange={(e) => handleToggleConsent('data_processing_consent', e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500 cursor-pointer"
                  />
                </div>

                <div className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between">
                  <div className="space-y-1 max-w-md">
                    <span className="font-bold text-slate-900 dark:text-white block">
                      Job Matching & Discovery Consent
                    </span>
                    <p className="text-slate-500">
                      Consent to comparing profile facts against job board requirements without fabricating unprovided information (AT-003).
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={profile.consent.job_matching_consent}
                    onChange={(e) => handleToggleConsent('job_matching_consent', e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500 cursor-pointer"
                  />
                </div>
              </div>
            </div>
          )}

          {/* 5. Field Audit Trail */}
          {activeSection === 'audit' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="border-b border-slate-100 dark:border-slate-800 pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">Granular Field Audit Trail</h2>
                <p className="text-xs text-slate-500">Immutable record of changes, timestamps, and actors.</p>
              </div>

              {profile.audit_trail.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <Clock className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No audit events recorded yet.</p>
                  <p className="text-[11px] text-slate-400">Modifications to verified profile facts will be tracked here with field-level precision.</p>
                </div>
              ) : (
                <div className="divide-y divide-slate-100 dark:divide-slate-800">
                  {profile.audit_trail.map((entry) => (
                    <div key={entry.id} className="py-3 flex items-start justify-between text-xs">
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-slate-900 dark:text-white capitalize">{entry.field}</span>
                          <span className="text-[10px] px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 font-mono">
                            {entry.changed_by}
                          </span>
                        </div>
                        <p className="text-slate-500 text-[11px]">{entry.reason || 'Manual modification'}</p>
                        {entry.new_value && (
                          <p className="text-slate-600 dark:text-slate-400 font-mono text-[11px]">New value: {entry.new_value}</p>
                        )}
                      </div>
                      <span suppressHydrationWarning className="text-[11px] text-slate-400 shrink-0">
                        {entry.changed_at ? new Date(entry.changed_at).toLocaleString() : ''}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 6. Education Section */}
          {activeSection === 'education' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">Education & Academic Background</h2>
                  <p className="text-xs text-slate-500">Verified academic degrees, institutions, and graduation records.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setShowAddEdu(true)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold transition cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Add Education</span>
                </button>
              </div>

              {showAddEdu && (
                <form onSubmit={handleAddEducation} className="p-4 border border-blue-200 dark:border-blue-900/40 rounded-xl bg-blue-50/40 dark:bg-blue-950/20 space-y-3">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Degree / Diploma *</label>
                      <input
                        type="text"
                        required
                        value={eduForm.degree}
                        onChange={(e) => setEduForm({ ...eduForm, degree: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. B.Sc. in Computer Science or BBA"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Institution / University *</label>
                      <input
                        type="text"
                        required
                        value={eduForm.institution}
                        onChange={(e) => setEduForm({ ...eduForm, institution: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Dhaka University"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Field of Study</label>
                      <input
                        type="text"
                        value={eduForm.field_of_study}
                        onChange={(e) => setEduForm({ ...eduForm, field_of_study: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Marketing, Business, CSE"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Years / Period</label>
                      <input
                        type="text"
                        value={eduForm.end_date}
                        onChange={(e) => setEduForm({ ...eduForm, end_date: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. 2016 – 2020 or Graduated"
                      />
                    </div>
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAddEdu(false)}
                      className="px-3 py-1.5 text-xs text-slate-500 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold"
                    >
                      Save Education
                    </button>
                  </div>
                </form>
              )}

              {profile.education.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <GraduationCap className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No education records added yet.</p>
                  <p className="text-[11px] text-slate-400">Click &quot;Add Education&quot; to register your academic history.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {profile.education.map((edu) => (
                    <div
                      key={edu.id}
                      className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 transition space-y-2"
                    >
                      <div className="flex items-start justify-between">
                        <div className="space-y-0.5">
                          <div className="flex items-center gap-2">
                            <h3 className="text-sm font-bold text-slate-900 dark:text-white">{edu.degree}</h3>
                            <span className="inline-flex items-center gap-1 px-1.5 py-0.2 rounded text-[9px] font-semibold bg-emerald-50 dark:bg-emerald-950/80 text-emerald-700 dark:text-emerald-300">
                              <CheckCircle2 className="w-2.5 h-2.5" />
                              <span>Verified Fact</span>
                            </span>
                          </div>
                          <p className="text-xs font-medium text-slate-600 dark:text-slate-300">
                            {edu.institution} {edu.field_of_study ? `• ${edu.field_of_study}` : ''}
                          </p>
                          <span className="text-[11px] text-slate-400 block">
                            {edu.start_date ? `${edu.start_date.slice(0, 4)} – ` : ''}{edu.end_date || 'Graduated'} {edu.grade ? `• Grade: ${edu.grade}` : ''}
                          </span>
                        </div>
                        <button
                          type="button"
                          onClick={() => handleDeleteEducation(edu.id)}
                          className="p-1 text-slate-400 hover:text-rose-600 transition"
                          title="Remove degree"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 7. Projects Section */}
          {activeSection === 'projects' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="border-b border-slate-100 dark:border-slate-800 pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">Notable Projects</h2>
                <p className="text-xs text-slate-500">Key campaigns, portfolio sites, and case studies.</p>
              </div>
              {profile.projects.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <FolderGit2 className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No projects recorded yet.</p>
                  <p className="text-[11px] text-slate-400">Projects extracted from your master resume will appear here.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {profile.projects.map((prj) => (
                    <div key={prj.id} className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 space-y-1.5">
                      <div className="flex items-center justify-between">
                        <h3 className="text-xs font-bold text-slate-900 dark:text-white">{prj.title}</h3>
                        <span className="text-[10px] text-slate-400">{prj.role}</span>
                      </div>
                      {prj.description && <p className="text-xs text-slate-600 dark:text-slate-400">{prj.description}</p>}
                      {prj.url && (
                        <a href={prj.url} target="_blank" rel="noreferrer" className="text-xs text-blue-600 dark:text-blue-400 hover:underline block">
                          {prj.url}
                        </a>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 8. Certificates Section */}
          {activeSection === 'certificates' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-4">
                <div>
                  <h2 className="text-base font-bold text-slate-900 dark:text-white">Certifications & Licenses</h2>
                  <p className="text-xs text-slate-500">Industry-recognized professional credentials and exams.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setShowAddCert(true)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold transition cursor-pointer"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Add Certificate</span>
                </button>
              </div>

              {showAddCert && (
                <form onSubmit={handleAddCertificate} className="p-4 border border-blue-200 dark:border-blue-900/40 rounded-xl bg-blue-50/40 dark:bg-blue-950/20 space-y-3">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Certificate Name *</label>
                      <input
                        type="text"
                        required
                        value={certForm.name}
                        onChange={(e) => setCertForm({ ...certForm, name: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Google Analytics (GA4) Certification"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Issuer / Organization *</label>
                      <input
                        type="text"
                        required
                        value={certForm.issuer}
                        onChange={(e) => setCertForm({ ...certForm, issuer: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. Google, HubSpot, Semrush"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Issue Date</label>
                      <input
                        type="text"
                        value={certForm.issue_date}
                        onChange={(e) => setCertForm({ ...certForm, issue_date: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. 2023-05-15"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Credential ID (Optional)</label>
                      <input
                        type="text"
                        value={certForm.credential_id}
                        onChange={(e) => setCertForm({ ...certForm, credential_id: e.target.value })}
                        className="w-full px-3 py-2 text-xs border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-white"
                        placeholder="e.g. CERT-89412"
                      />
                    </div>
                  </div>
                  <div className="flex items-center justify-end gap-2 pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAddCert(false)}
                      className="px-3 py-1.5 text-xs text-slate-500 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-semibold"
                    >
                      Save Certificate
                    </button>
                  </div>
                </form>
              )}

              {profile.certificates.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <Award className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No certificates added yet.</p>
                  <p className="text-[11px] text-slate-400">Click &quot;Add Certificate&quot; to record your verified credentials.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {profile.certificates.map((cert) => (
                    <div key={cert.id} className="p-4 rounded-xl border border-slate-200 dark:border-slate-800 flex items-start justify-between gap-3">
                      <div className="space-y-1 min-w-0">
                        <span className="text-xs font-bold text-slate-900 dark:text-white block truncate">{cert.name}</span>
                        <span className="text-[11px] text-slate-500 block">{cert.issuer} • Issued {cert.issue_date}</span>
                        {cert.credential_id && (
                          <span className="text-[10px] font-mono text-slate-400 block truncate">ID: {cert.credential_id}</span>
                        )}
                      </div>
                      <button
                        type="button"
                        onClick={() => handleDeleteCertificate(cert.id)}
                        className="p-1 text-slate-400 hover:text-rose-600 transition shrink-0"
                        title="Remove certificate"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 9. Languages Section */}
          {activeSection === 'languages' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="border-b border-slate-100 dark:border-slate-800 pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">Language Proficiencies</h2>
                <p className="text-xs text-slate-500">Spoken and written language fluency.</p>
              </div>
              {profile.languages.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <Globe2 className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No language proficiencies recorded yet.</p>
                  <p className="text-[11px] text-slate-400">Language fluency and proficiencies extracted from your resume will appear here.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {profile.languages.map((lng) => (
                    <div key={lng.id} className="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-900 dark:text-white">{lng.language}</span>
                      <span className="px-2 py-0.5 rounded text-[10px] font-semibold uppercase bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                        {lng.proficiency}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 10. Links & Profiles Section */}
          {activeSection === 'links' && (
            <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="border-b border-slate-100 dark:border-slate-800 pb-4">
                <h2 className="text-base font-bold text-slate-900 dark:text-white">External Profiles & Portfolio Links</h2>
                <p className="text-xs text-slate-500">Verified links to LinkedIn, GitHub, portfolio, or publications.</p>
              </div>
              {profile.links.length === 0 ? (
                <div className="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl space-y-2">
                  <LinkIcon className="w-8 h-8 text-slate-400 mx-auto" />
                  <p className="text-xs font-medium text-slate-600 dark:text-slate-300">No external profile links added yet.</p>
                  <p className="text-[11px] text-slate-400">Links to LinkedIn, GitHub, personal website, or portfolio will be listed here.</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {profile.links.map((lnk) => (
                    <div key={lnk.id} className="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between">
                      <div>
                        <span className="text-xs font-bold text-slate-900 dark:text-white block">{lnk.label}</span>
                        <a href={lnk.url} target="_blank" rel="noreferrer" className="text-[11px] text-blue-600 dark:text-blue-400 hover:underline">
                          {lnk.url}
                        </a>
                      </div>
                      <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500">
                        {lnk.link_type}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
