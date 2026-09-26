// TypeScript SDK Client for Career / LinkedIn / Social Platform API
// Generated to match OpenAPI 3.1 contract (api/openapi.yaml)

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
}

export interface SectionMeta {
  id: "career" | "linkedin" | "social";
  name: string;
  description: string;
}

export interface User {
  id: string;
  email: string;
  fullName: string;
  platformRole: "super_admin" | "operator" | "standard";
  createdAt: string;
}

export interface ProfileFact {
  id: string;
  userId: string;
  category: "experience" | "education" | "skill" | "project" | "certification";
  title: string;
  organization?: string;
  startDate?: string;
  endDate?: string;
  isCurrent?: boolean;
  details?: string[];
  confirmed: boolean;
  createdAt: string;
}

export interface ResumeVersion {
  id: string;
  userId: string;
  versionNumber: number;
  title: string;
  templateId: string;
  isMaster: boolean;
  createdAt: string;
}

export interface JobApplication {
  id: string;
  userId: string;
  jobTitle: string;
  companyName: string;
  jobUrl: string;
  status: "draft" | "applied" | "interviewing" | "rejected" | "offer";
  confirmationType: "provider_confirmed" | "user_confirmed" | "unconfirmed";
  appliedAt?: string;
  createdAt: string;
}

export interface SocialPost {
  id: string;
  workspaceId: string;
  authorId: string;
  title: string;
  content: string;
  targetPlatforms: Array<"linkedin" | "twitter" | "facebook" | "instagram" | "youtube">;
  status: "draft" | "pending_approval" | "scheduled" | "published" | "failed";
  scheduledAt?: string;
  createdAt: string;
}

export class PlatformClient {
  private baseUrl: string;
  private token?: string;

  constructor(options: { baseUrl?: string; token?: string } = {}) {
    this.baseUrl = options.baseUrl || "http://localhost:8080";
    this.token = options.token;
  }

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const errorBody = await response.text();
      throw new Error(`API Error [${response.status}]: ${errorBody}`);
    }

    return response.json() as Promise<T>;
  }

  async getHealth(): Promise<HealthResponse> {
    return this.request<HealthResponse>("/healthz");
  }

  async getSections(): Promise<SectionMeta[]> {
    return this.request<SectionMeta[]>("/api/v1/sections");
  }

  async getCareerProfile(): Promise<{ userId: string; facts: ProfileFact[] }> {
    return this.request<{ userId: string; facts: ProfileFact[] }>("/api/v1/career/profile");
  }

  async listResumes(): Promise<ResumeVersion[]> {
    return this.request<ResumeVersion[]>("/api/v1/career/resumes");
  }

  async listApplications(): Promise<JobApplication[]> {
    return this.request<JobApplication[]>("/api/v1/career/applications");
  }

  async listSocialPosts(workspaceId: string): Promise<SocialPost[]> {
    return this.request<SocialPost[]>(`/api/v1/social/posts?workspaceId=${encodeURIComponent(workspaceId)}`);
  }
}