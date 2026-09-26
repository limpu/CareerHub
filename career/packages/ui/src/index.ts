export const BRAND_NAME = 'Career & Social Platform';

export interface SectionMeta {
  id: 'career' | 'linkedin' | 'social';
  title: string;
  description: string;
}

export const PLATFORM_SECTIONS: SectionMeta[] = [
  {
    id: 'career',
    title: 'Career',
    description: 'Build your profile, craft ATS-friendly resumes, discover matching jobs and track applications.',
  },
  {
    id: 'linkedin',
    title: 'LinkedIn',
    description: 'Optimize your profile, draft authentic relationship messages and manage allowed connections.',
  },
  {
    id: 'social',
    title: 'Social',
    description: 'Schedule content, research trends, collaborate on campaigns and analyze engagement.',
  },
];
