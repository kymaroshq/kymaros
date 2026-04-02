export type KymarosTier = "community" | "team" | "enterprise";

export interface FeatureFlags {
  maxHistoryDays: number;   // 7, 90, or 0 (unlimited)
  compliancePage: boolean;
  pdfExport: boolean;
  csvExport: boolean;
  multiBackup: boolean;
  multiCluster: boolean;
  sso: boolean;
  rtoAnalytics: boolean;
  regressionAlerts: boolean;
  timeline: boolean;
  scoreBreakdown: boolean;
}

export interface LicenseResponse {
  tier: KymarosTier;
  expiresAt?: string;
  trialEndsAt?: string;
  isExpired: boolean;
  isTrialing: boolean;
  trialDaysLeft?: number;
  features: FeatureFlags;
}
