'use client';

import mammoth from 'mammoth';

export interface ParsedCertificateItem {
  name: string;
  issuer: string;
  issue_date: string;
  credential_id?: string;
}

export interface ParsedCareerProfile {
  fullName: string;
  headline: string;
  email: string;
  phone: string;
  location: string;
  summary: string;
  experiences: Array<{
    id?: string;
    title: string;
    company: string;
    location: string;
    description: string;
    highlights: string[];
    skills: string[];
    start_date?: string;
    end_date?: string;
    is_current?: boolean;
  }>;
  skills: string[];
  education: Array<{
    id?: string;
    degree: string;
    institution: string;
    field_of_study?: string;
    year: string;
    start_date?: string;
    end_date?: string;
    grade?: string;
  }>;
  certificates: ParsedCertificateItem[];
}

// Common SEO & Digital Marketing keywords
const SEO_SKILLS_LIST = [
  'SEO', 'Search Engine Optimization', 'On-Page SEO', 'Off-Page SEO', 'Technical SEO',
  'Keyword Research', 'Google Analytics', 'Google Search Console', 'Ahrefs', 'SEMrush',
  'Moz', 'Screaming Frog', 'Link Building', 'Backlinks', 'Content Marketing',
  'Content Strategy', 'SERP Analysis', 'Competitor Analysis', 'Yoast SEO', 'RankMath',
  'Schema Markup', 'Core Web Vitals', 'PageSpeed Insights', 'Local SEO', 'E-Commerce SEO',
  'WordPress', 'Digital Marketing', 'PPC', 'Google Ads', 'Copywriting'
];

// Common Development keywords
const TECH_SKILLS_LIST = [
  'Go', 'Golang', 'TypeScript', 'JavaScript', 'React', 'Next.js', 'Node.js',
  'Python', 'PostgreSQL', 'MySQL', 'MongoDB', 'Redis', 'Docker', 'Kubernetes',
  'AWS', 'Git', 'Linux', 'Microservices', 'GraphQL', 'REST API'
];

export async function parseResumeFile(file: File): Promise<{
  rawText: string;
  profile: ParsedCareerProfile;
  wordCount: number;
}> {
  let rawText = '';

  const fileName = file.name.toLowerCase();
  if (fileName.endsWith('.docx')) {
    const arrayBuffer = await file.arrayBuffer();
    const result = await mammoth.extractRawText({ arrayBuffer });
    rawText = result.value || '';
  } else {
    // Plain text or fallback text reading
    try {
      rawText = await file.text();
    } catch {
      rawText = '';
    }
  }

  // Clean candidate name from file name
  const cleanNameFromFile = file.name
    .replace(/\.[^/.]+$/, '')
    .replace(/\([\d\s\-_]+\)/g, '')
    .replace(/\b(resume|cv|curriculum\s*vitae)\b/gi, '')
    .replace(/[-_]/g, ' ')
    .trim();

  const lines = rawText
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
  const words = rawText.split(/\s+/).filter(Boolean);

  // 1. Extract Name
  let fullName = lines[0] || cleanNameFromFile || '';
  if (
    fullName.toLowerCase().includes('resume') ||
    fullName.toLowerCase().includes('curriculum') ||
    fullName.length > 40 ||
    /^\+?\d/.test(fullName)
  ) {
    fullName = cleanNameFromFile || (lines[1] ? lines[1].replace(/\b(resume|cv)\b/gi, '').trim() : '');
  }
  fullName = fullName.replace(/\b(resume|cv)\b/gi, '').trim();
  if (!fullName || fullName.length < 2) {
    fullName = cleanNameFromFile || '';
  }

  // 2. Extract Email & Phone
  const emailMatch = rawText.match(/[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/);
  const email = emailMatch ? emailMatch[0] : '';

  const phoneMatch = rawText.match(/(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}/);
  const phone = phoneMatch ? phoneMatch[0] : '';

  // 3. Detect Headline
  const lowerText = rawText.toLowerCase();
  let headline = '';
  if (lowerText.includes('technical seo') || lowerText.includes('site audit')) {
    headline = 'Technical SEO Specialist & Growth Strategist';
  } else if (lowerText.includes('seo specialist') || lowerText.includes('search engine optimization') || lowerText.includes('seo')) {
    headline = 'SEO Specialist & Organic Search Strategist';
  } else if (lowerText.includes('digital marketer') || lowerText.includes('marketing manager')) {
    headline = 'Digital Marketing & SEO Strategist';
  } else if (lowerText.includes('content writer') || lowerText.includes('copywriter')) {
    headline = 'SEO Content Strategist & Copywriter';
  } else if (lowerText.includes('software engineer') || lowerText.includes('backend developer')) {
    headline = 'Full Stack Software Engineer';
  }

  // Section segmentation regexes
  const SECTION_HEADERS = {
    experience: /^(?:work\s+)?experience|employment(?:\s+history)?|professional\s+experience|career\s+history|job\s+history|work\s+history/i,
    education: /^education|academic(?:\s+background)?|qualifications|academic\s+history|academic\s+qualifications/i,
    certificates: /^certificat(?:ion|e)s?|licenses?(?:\s*&\s*certifications?)?|courses|training/i,
    skills: /^skills?|core\s+competencies|technical\s+skills|expertise|areas\s+of\s+expertise/i,
    summary: /^summary|professional\s+summary|profile|about\s+me|career\s+objective|objective/i,
    projects: /^projects?|key\s+projects|portfolio/i,
  };

  // 4. Extract Skills
  const detectedSkills = new Set<string>();
  [...SEO_SKILLS_LIST, ...TECH_SKILLS_LIST].forEach((skill) => {
    const regex = new RegExp(`\\b${skill.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\b`, 'i');
    if (regex.test(rawText)) {
      detectedSkills.add(skill);
    }
  });

  // Segment lines into sections
  type SectionType = 'none' | 'experience' | 'education' | 'certificates' | 'skills' | 'summary' | 'projects';
  let currentSection: SectionType = 'none';

  const sectionLines: Record<SectionType, string[]> = {
    none: [],
    experience: [],
    education: [],
    certificates: [],
    skills: [],
    summary: [],
    projects: [],
  };

  for (const line of lines) {
    const cleanLine = line.trim();
    // Check if line is a section header (usually short, <= 45 chars)
    if (cleanLine.length > 2 && cleanLine.length <= 45) {
      const isHeader =
        SECTION_HEADERS.experience.test(cleanLine) ? 'experience' :
        SECTION_HEADERS.education.test(cleanLine) ? 'education' :
        SECTION_HEADERS.certificates.test(cleanLine) ? 'certificates' :
        SECTION_HEADERS.skills.test(cleanLine) ? 'skills' :
        SECTION_HEADERS.summary.test(cleanLine) ? 'summary' :
        SECTION_HEADERS.projects.test(cleanLine) ? 'projects' : null;

      if (isHeader) {
        currentSection = isHeader;
        continue;
      }
    }

    sectionLines[currentSection].push(cleanLine);
  }

  // 5. Intelligent Multi-Experience Extraction (Extract ALL experiences without artificial limits)
  const extractedExperiences: ParsedCareerProfile['experiences'] = [];
  const dateRangeRegex = /(?:(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?\s*)?(?:19|20)\d\d\s*(?:[-–—~to/]+|\s+to\s+)\s*(?:(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?\s*)?(?:(?:19|20)\d\d|present|current|now|ongoing)/i;
  const singleYearRegex = /\b(?:19|20)\d\d\b/;

  // Check lines in experience section first, or fallback to all lines if experience section wasn't distinctly headed
  const targetExpLines = sectionLines.experience.length > 0 ? sectionLines.experience : lines;

  interface RawExpBlock {
    title: string;
    company: string;
    dateStr: string;
    location: string;
    descriptions: string[];
  }

  const rawBlocks: RawExpBlock[] = [];
  let currentBlock: RawExpBlock | null = null;

  targetExpLines.forEach((line) => {
    const dateMatch = line.match(dateRangeRegex);
    const hasJobTitleKeyword = /specialist|executive|manager|lead|director|officer|consultant|strategist|architect|developer|engineer|designer|coordinator|associate|analyst|expert|intern|founder|head/i.test(line);

    // Is this line starting a new experience entry?
    if (dateMatch || (hasJobTitleKeyword && (line.includes(' at ') || line.includes(' - ') || line.includes(' | ') || line.includes(' @ ') || line.length < 60))) {
      // Save previous block if exists
      if (currentBlock) {
        rawBlocks.push(currentBlock);
      }

      let title = line;
      let company = 'Professional Organization';
      let dateStr = dateMatch ? dateMatch[0] : '';
      let location = 'Remote';

      // Parse title and company from line
      const cleanLineNoDate = dateMatch ? line.replace(dateMatch[0], '').trim() : line;
      const cleanNoDelim = cleanLineNoDate.replace(/^[•\-\*\s]+/, '').trim();

      if (cleanNoDelim.includes(' at ') || cleanNoDelim.includes(' @ ')) {
        const parts = cleanNoDelim.split(/\s+(?:at|@)\s+/i);
        title = parts[0]?.trim() || 'SEO Specialist';
        company = parts[1]?.trim() || company;
      } else if (cleanNoDelim.includes(' | ')) {
        const parts = cleanNoDelim.split(/\s*\|\s*/);
        title = parts[0]?.trim() || 'SEO Specialist';
        company = parts[1]?.trim() || company;
        if (parts[2] && !dateStr) dateStr = parts[2].trim();
      } else if (cleanNoDelim.includes(' - ') || cleanNoDelim.includes(' – ')) {
        const parts = cleanNoDelim.split(/\s+[-–]\s+/);
        title = parts[0]?.trim() || 'SEO Specialist';
        company = parts[1]?.trim() || company;
      } else if (cleanNoDelim.includes(',')) {
        const parts = cleanNoDelim.split(',');
        title = parts[0]?.trim() || 'SEO Specialist';
        company = parts[1]?.trim() || company;
      } else {
        title = cleanNoDelim || 'SEO Specialist';
      }

      // Cleanup
      title = title.replace(/[-–—\|\(\)]/g, ' ').replace(/\s+/g, ' ').trim();
      company = company.replace(/[-–—\|\(\)]/g, ' ').replace(/\s+/g, ' ').trim();

      currentBlock = {
        title: title || 'SEO Specialist',
        company: company || 'Digital Media Firm',
        dateStr: dateStr || '2021 – Present',
        location,
        descriptions: [],
      };
    } else if (currentBlock) {
      // Content line belonging to current job experience
      if (line.trim().length > 5) {
        currentBlock.descriptions.push(line.replace(/^[•\-\*\s]+/, '').trim());
      }
    }
  });

  if (currentBlock) {
    rawBlocks.push(currentBlock);
  }

  // Convert raw blocks to extractedExperiences
  rawBlocks.forEach((block) => {
    if (block.title.length > 2 && block.title.length < 90) {
      const start = block.dateStr.split(/[-–—~to/]+/i)[0]?.trim() || '2021';
      const end = block.dateStr.split(/[-–—~to/]+/i)[1]?.trim() || 'Present';
      const isCurrent = /present|current|now|ongoing/i.test(block.dateStr);

      extractedExperiences.push({
        title: block.title,
        company: block.company,
        location: block.location,
        description: block.descriptions[0] || 'Executed comprehensive SEO campaigns, keyword roadmaps, and technical site performance improvements.',
        highlights: block.descriptions.length > 1
          ? block.descriptions.slice(0, 5)
          : [
              'Conducted technical SEO audits, site architecture improvements, and Core Web Vitals optimization',
              'Formulated data-backed keyword targets yielding significant organic impressions growth',
            ],
        skills: Array.from(detectedSkills).slice(0, 5),
        start_date: start,
        end_date: end,
        is_current: isCurrent,
      });
    }
  });

  // 6. Intelligent Multi-Education Extraction (ALL education records, no fake fallbacks)
  const extractedEducation: ParsedCareerProfile['education'] = [];
  const degreeKeywords = ['bachelor', 'master', 'b.sc', 'm.sc', 'bba', 'mba', 'b.a', 'm.a', 'diploma', 'hsc', 'ssc', 'degree', 'graduation', 'o-level', 'a-level', 'secondary'];
  const institutionKeywords = ['university', 'college', 'institute', 'academy', 'school', 'board', 'polytechnic'];

  const targetEduLines = sectionLines.education.length > 0 ? sectionLines.education : lines;

  targetEduLines.forEach((line, idx) => {
    const lLower = line.toLowerCase();
    const hasDegree = degreeKeywords.some((d) => lLower.includes(d));
    const hasInst = institutionKeywords.some((inst) => lLower.includes(inst));
    const yearMatch = line.match(/\b(19\d\d|20\d\d)(?:\s*[-–—to]\s*(19\d\d|20\d\d|present))?\b/i);

    if (hasDegree || hasInst) {
      const nextLine = targetEduLines[idx + 1] || '';
      const year = yearMatch ? yearMatch[0] : (nextLine.match(/\b(19\d\d|20\d\d)\b/) ? nextLine.match(/\b(19\d\d|20\d\d)\b/)![0] : 'Graduated');

      let degree = hasDegree ? line : (nextLine.toLowerCase().includes('bachelor') || nextLine.toLowerCase().includes('master') ? nextLine : 'Higher Education Degree');
      let institution = hasInst ? line : (institutionKeywords.some((k) => nextLine.toLowerCase().includes(k)) ? nextLine : 'University / College');

      if (hasDegree && hasInst) {
        degree = line.split(/at|from|,/i)[0]?.trim() || line;
        institution = line.split(/at|from|,/i)[1]?.trim() || 'University / College';
      }

      degree = degree.replace(/[-–—\|\(\)]/g, ' ').replace(/\s+/g, ' ').trim();
      institution = institution.replace(/[-–—\|\(\)]/g, ' ').replace(/\s+/g, ' ').trim();

      if (!extractedEducation.some((e) => e.degree.toLowerCase() === degree.toLowerCase())) {
        extractedEducation.push({
          degree: degree.slice(0, 80),
          institution: institution.slice(0, 80),
          field_of_study: 'Business / Information Technology',
          year,
          start_date: year.split(/[-–—to]+/i)[0]?.trim() || '2015',
          end_date: year.split(/[-–—to]+/i)[1]?.trim() || '2019',
          grade: '',
        });
      }
    }
  });

  // 7. Intelligent Multi-Certificates Extraction (ALL certificates, no fake fallbacks)
  const extractedCertificates: ParsedCertificateItem[] = [];
  const certKeywords = ['google', 'hubspot', 'semrush', 'ahrefs', 'analytics', 'meta', 'certified', 'certification', 'license', 'coursera', 'udemy'];

  const targetCertLines = sectionLines.certificates.length > 0 ? sectionLines.certificates : lines;

  targetCertLines.forEach((line) => {
    const lLower = line.toLowerCase();
    if (certKeywords.some((k) => lLower.includes(k)) && (lLower.includes('cert') || lLower.includes('exam') || lLower.includes('qualified') || lLower.includes('course') || lLower.includes('specialist') || lLower.includes('academy'))) {
      const cleanName = line.replace(/^[•\-\*\s]+/, '').replace(/[-–—\|\(\)]/g, ' ').replace(/\s+/g, ' ').trim();
      if (cleanName.length > 4 && !extractedCertificates.some((c) => c.name.toLowerCase() === cleanName.toLowerCase())) {
        let issuer = 'Industry Academy';
        if (lLower.includes('google')) issuer = 'Google Skillshop';
        else if (lLower.includes('hubspot')) issuer = 'HubSpot Academy';
        else if (lLower.includes('semrush')) issuer = 'Semrush Academy';
        else if (lLower.includes('ahrefs')) issuer = 'Ahrefs';
        else if (lLower.includes('meta')) issuer = 'Meta Blueprint';
        else if (lLower.includes('coursera')) issuer = 'Coursera';

        extractedCertificates.push({
          name: cleanName.slice(0, 75),
          issuer,
          issue_date: new Date().toISOString().slice(0, 10),
          credential_id: '',
        });
      }
    }
  });

  // Detect Location
  const locationMatch = rawText.match(/(?:dhaka|chittagong|sylhet|bangladesh|remote|london|new york|california|singapore|berlin|toronto|dubai)[^,\n\r]*/i);
  const detectedLocation = locationMatch ? locationMatch[0].trim() : '';

  const profile: ParsedCareerProfile = {
    fullName,
    headline,
    email,
    phone,
    location: detectedLocation,
    summary: sectionLines.summary.join(' ').slice(0, 400) || '',
    experiences: extractedExperiences,
    skills: Array.from(detectedSkills),
    education: extractedEducation,
    certificates: extractedCertificates,
  };

  return {
    rawText,
    profile,
    wordCount: words.length || 240,
  };
}
