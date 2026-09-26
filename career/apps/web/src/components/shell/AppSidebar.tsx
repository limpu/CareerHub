'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Briefcase,
  Users,
  FileText,
  Sliders,
  Sparkles,
  BarChart3,
  Compass,
  Bookmark,
  Send,
  SlidersHorizontal,
  Wand2,
  ExternalLink,
  ClipboardCheck,
  Calendar,
  UserCheck,
  RotateCcw,
  Shield,
  FileSpreadsheet,
  ChevronDown,
  ChevronRight,
  Menu,
  X,
  ScanLine,
  TrendingUp,
  UserPlus,
  MessageSquare,
  PenTool,
  Mic,
  CalendarDays,
  Target,
  Megaphone,
  Radio,
  Clock,
  Key,
  Database
} from 'lucide-react';

interface SubMenuItem {
  href: string;
  label: string;
  badge?: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface MenuGroup {
  id: string;
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  accentColor: string;
  items: SubMenuItem[];
}

const CAREER_MENU_GROUPS: MenuGroup[] = [
  {
    id: 'profile_prep',
    title: 'Profile & Resume',
    icon: FileText,
    accentColor: 'text-blue-600 dark:text-blue-400',
    items: [
      { href: '/career', label: 'Career Overview', icon: Briefcase },
      { href: '/career/profile', label: 'Master Profile', badge: 'CAR-02', icon: UserCheck },
      { href: '/career/preferences', label: 'Preferences & Exclusions', badge: 'CAR-03', icon: Sliders },
      { href: '/career/resumes', label: 'Master Resumes (PDF/DOCX)', badge: 'CAR-04', icon: FileText },
      { href: '/career/tailor', label: 'Tailor per Job', badge: 'CAR-05', icon: Sparkles },
      { href: '/career/feedback', label: 'Feedback & Skills Demand', badge: 'CAR-06', icon: BarChart3 },
      { href: '/career/consistency', label: 'Tri-Source Consistency', badge: 'CAR-24', icon: RotateCcw }
    ]
  },
  {
    id: 'job_discovery',
    title: 'Job Search & Discovery',
    icon: Compass,
    accentColor: 'text-sky-600 dark:text-sky-400',
    items: [
      { href: '/career/discovery', label: 'Multi-Board Discovery', badge: 'CAR-07', icon: Compass },
      { href: '/career/saved', label: 'Saved Jobs & Dedupe', badge: 'CAR-10', icon: Bookmark },
      { href: '/career/leads', label: 'Recruiter Leads & Posts', badge: 'CAR-20', icon: UserPlus }
    ]
  },
  {
    id: 'apply_workflow',
    title: 'Apply & Interviewing',
    icon: Send,
    accentColor: 'text-purple-600 dark:text-purple-400',
    items: [
      { href: '/career/apply', label: 'Review & Apply Gateway', badge: 'CAR-11', icon: Send },
      { href: '/career/fields', label: 'Field Recognition Studio', badge: 'CAR-12', icon: SlidersHorizontal },
      { href: '/career/wizard', label: 'Multi-Step Wizard Studio', badge: 'CAR-13', icon: Wand2 },
      { href: '/career/portal', label: 'External & Indeed Portals', badge: 'CAR-14', icon: ExternalLink },
      { href: '/career/interviews', label: 'Interview Prep & Funnel', badge: 'CAR-23', icon: Mic }
    ]
  },
  {
    id: 'management_reports',
    title: 'Tracking & Governance',
    icon: ClipboardCheck,
    accentColor: 'text-emerald-600 dark:text-emerald-400',
    items: [
      { href: '/career/status', label: 'Status, Proof & Sheets Sync', badge: 'CAR-15', icon: ClipboardCheck },
      { href: '/career/reports', label: 'Daily Reports & Reminders', badge: 'CAR-19', icon: Calendar },
      { href: '/career/runs', label: 'Run Controls & Recovery', badge: 'CAR-21', icon: RotateCcw },
      { href: '/career/privacy', label: 'PII Controls & Redaction', badge: 'CAR-22', icon: Shield }
    ]
  }
];

const LINKEDIN_MENU_GROUPS: MenuGroup[] = [
  {
    id: 'li_profile',
    title: 'Profile & Capability',
    icon: ScanLine,
    accentColor: 'text-sky-600 dark:text-sky-400',
    items: [
      { href: '/linkedin', label: 'Capabilities Inspector', badge: 'LI-01', icon: Shield },
      { href: '/linkedin#sec-profile-import', label: 'Profile Scan / Import', badge: 'LI-02', icon: ScanLine },
      { href: '/linkedin#sec-profile-opt', label: 'Profile Optimization', badge: 'LI-03', icon: TrendingUp },
      { href: '/linkedin#sec-records-vault', label: 'Entities Vault (People/Jobs)', badge: 'LI-04', icon: Database }
    ]
  },
  {
    id: 'li_networking',
    title: 'Networking & Outreach',
    icon: Users,
    accentColor: 'text-indigo-600 dark:text-indigo-400',
    items: [
      { href: '/linkedin#sec-discovery-engine', label: 'Boolean People Discovery', badge: 'LI-05', icon: Compass },
      { href: '/linkedin#sec-recruiter-workspace', label: 'Recruiter Lead Pipeline', badge: 'LI-06', icon: UserPlus },
      { href: '/linkedin#sec-connection-queue', label: 'Connection Queue & Notes', badge: 'LI-07', icon: Send },
      { href: '/linkedin#sec-company-follow', label: 'Company Follow Planning', badge: 'LI-08', icon: Target },
      { href: '/linkedin#sec-inbox-triage', label: 'Inbox Conversation Triage', badge: 'LI-09', icon: MessageSquare }
    ]
  },
  {
    id: 'li_content',
    title: 'Content & Stories',
    icon: PenTool,
    accentColor: 'text-purple-600 dark:text-purple-400',
    items: [
      { href: '/linkedin#sec-comment-sweep', label: 'Comment Drafting Sweep', badge: 'LI-10', icon: MessageSquare },
      { href: '/linkedin#sec-post-writing', label: 'Post Writing & Hooks', badge: 'LI-11', icon: PenTool },
      { href: '/linkedin#sec-humanizer', label: 'Voice & AI Humanizer', badge: 'LI-12', icon: Wand2 },
      { href: '/linkedin#sec-story-bank', label: 'Story Bank & Interviewer', badge: 'LI-13', icon: Mic },
      { href: '/linkedin#sec-content-calendar', label: 'Content Plan & Calendar', badge: 'LI-14', icon: CalendarDays }
    ]
  },
  {
    id: 'li_analytics_safety',
    title: 'Analytics & Safety',
    icon: BarChart3,
    accentColor: 'text-emerald-600 dark:text-emerald-400',
    items: [
      { href: '/linkedin#sec-analytics', label: 'Engagement & ICP Analytics', badge: 'LI-15', icon: BarChart3 },
      { href: '/linkedin#sec-advocacy', label: 'Employee Advocacy & Anti-Pod', badge: 'LI-16', icon: Megaphone },
      { href: '/linkedin#sec-provider-fallback', label: 'Provider Fallback Doctor', badge: 'LI-17', icon: Radio },
      { href: '/linkedin#sec-safety-limits', label: 'Limits & Safety Gatekeeper', badge: 'LI-18', icon: Clock },
      { href: '/linkedin#sec-session-tools', label: 'Session Lifecycle & Tools', badge: 'LI-19', icon: Key },
      { href: '/linkedin#sec-relationship-export', label: 'CSV/Sheets CRM Exports', badge: 'LI-20', icon: FileSpreadsheet }
    ]
  }
];

export function AppSidebar() {
  const pathname = usePathname();
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({});
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  // Close mobile sidebar upon route change
  useEffect(() => {
    setIsMobileOpen(false);
  }, [pathname]);

  const isLinkedIn = pathname.startsWith('/linkedin');

  const currentGroups = isLinkedIn ? LINKEDIN_MENU_GROUPS : CAREER_MENU_GROUPS;

  const toggleGroup = (groupId: string) => {
    setCollapsedGroups((prev) => ({
      ...prev,
      [groupId]: !prev[groupId]
    }));
  };

  return (
    <>
      {/* Mobile Sidebar Toggle Button */}
      <div className="lg:hidden fixed bottom-4 right-4 z-40">
        <button
          type="button"
          onClick={() => setIsMobileOpen(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-full shadow-lg font-medium text-xs focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 cursor-pointer"
        >
          <Menu className="w-4 h-4" />
          <span>Menu</span>
        </button>
      </div>

      {/* Mobile Backdrop */}
      {isMobileOpen && (
        <div
          onClick={() => setIsMobileOpen(false)}
          className="lg:hidden fixed inset-0 bg-slate-900/50 backdrop-blur-xs z-40 transition-opacity"
        />
      )}

      {/* Sidebar Container */}
      <aside
        className={`
          fixed top-16 bottom-0 left-0 z-40 w-72 bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800
          flex flex-col transition-transform duration-300 ease-in-out
          lg:translate-x-0 ${isMobileOpen ? 'translate-x-0 shadow-2xl' : '-translate-x-full'}
        `}
      >
        {/* Top Product / Mode Indicator */}
        <div className="p-4 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/50 dark:bg-slate-900/50 flex items-center justify-between">
          <div className="flex items-center space-x-2.5">
            <div className={`w-8 h-8 rounded-lg flex items-center justify-center font-bold text-white shadow-xs ${
              isLinkedIn
                ? 'bg-gradient-to-br from-sky-600 to-blue-700'
                : 'bg-gradient-to-br from-indigo-600 to-blue-600'
            }`}>
              {isLinkedIn ? <Users className="w-4 h-4" /> : <Briefcase className="w-4 h-4" />}
            </div>
            <div>
              <div className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                Active Module
              </div>
              <div className="text-xs font-bold text-slate-900 dark:text-white truncate">
                {isLinkedIn ? 'LinkedIn Platform' : 'Career Management'}
              </div>
            </div>
          </div>

          <button
            type="button"
            onClick={() => setIsMobileOpen(false)}
            className="lg:hidden p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-md"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Navigation Groups List */}
        <div className="flex-1 overflow-y-auto px-3 py-4 space-y-5">
          {/* Quick Hub Switcher */}
          <div className="grid grid-cols-2 gap-1.5 p-1 bg-slate-100 dark:bg-slate-800 rounded-lg text-xs font-semibold">
            <Link
              href="/career"
              className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-md transition ${
                !isLinkedIn
                  ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
              }`}
            >
              <Briefcase className="w-3.5 h-3.5" />
              <span>Career</span>
            </Link>
            <Link
              href="/linkedin"
              className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-md transition ${
                isLinkedIn
                  ? 'bg-white dark:bg-slate-900 text-sky-600 dark:text-sky-400 shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
              }`}
            >
              <Users className="w-3.5 h-3.5" />
              <span>LinkedIn</span>
            </Link>
          </div>

          {/* Grouped Menus */}
          {currentGroups.map((group) => {
            const isCollapsed = !!collapsedGroups[group.id];
            const GroupIcon = group.icon;

            return (
              <div key={group.id} className="space-y-1">
                {/* Group Header Toggle */}
                <button
                  type="button"
                  onClick={() => toggleGroup(group.id)}
                  className="w-full flex items-center justify-between px-2 py-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 rounded-md transition cursor-pointer"
                >
                  <div className="flex items-center space-x-1.5">
                    <GroupIcon className={`w-3.5 h-3.5 ${group.accentColor}`} />
                    <span>{group.title}</span>
                  </div>
                  {isCollapsed ? (
                    <ChevronRight className="w-3.5 h-3.5 text-slate-400" />
                  ) : (
                    <ChevronDown className="w-3.5 h-3.5 text-slate-400" />
                  )}
                </button>

                {/* Sub-menu Items */}
                {!isCollapsed && (
                  <div className="space-y-0.5 pl-1.5">
                    {group.items.map((item) => {
                      const isActive = pathname === item.href || (item.href !== '/career' && item.href !== '/linkedin' && pathname.startsWith(item.href));
                      const ItemIcon = item.icon;

                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          className={`flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs font-medium transition group ${
                            isActive
                              ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 font-semibold'
                              : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800/60 hover:text-slate-900 dark:hover:text-white'
                          }`}
                        >
                          <div className="flex items-center space-x-2 truncate">
                            <ItemIcon className={`w-3.5 h-3.5 shrink-0 ${isActive ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-400 group-hover:text-slate-600'}`} />
                            <span className="truncate">{item.label}</span>
                          </div>

                          {item.badge && (
                            <span className={`text-[10px] font-mono px-1.5 py-0.2 rounded border ${
                              isActive
                                ? 'bg-indigo-100 dark:bg-indigo-900/60 text-indigo-800 dark:text-indigo-200 border-indigo-200 dark:border-indigo-800'
                                : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-700'
                            }`}>
                              {item.badge}
                            </span>
                          )}
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Sidebar Footer */}
        <div className="p-3.5 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 text-[11px] text-slate-500">
          <div className="flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-300">Deterministic Mode</span>
            <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-100 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300">
              Safe Active
            </span>
          </div>
        </div>
      </aside>
    </>
  );
}
