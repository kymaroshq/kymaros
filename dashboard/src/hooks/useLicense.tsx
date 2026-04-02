import { createContext, useContext, useMemo } from 'react';
import { useLicense as useLicenseApi } from './useKymarosData';
import type { LicenseResponse, KymarosTier, FeatureFlags } from '../types/license';

interface LicenseContextValue {
  tier: KymarosTier;
  features: FeatureFlags;
  isTrialing: boolean;
  trialDaysLeft: number | undefined;
  isExpired: boolean;
  isCommunity: boolean;
  isTeam: boolean;
  isEnterprise: boolean;
  loading: boolean;
  canAccess: (requiredTier: KymarosTier) => boolean;
}

const defaultFeatures: FeatureFlags = {
  maxHistoryDays: 7,
  compliancePage: false,
  pdfExport: false,
  csvExport: false,
  multiBackup: false,
  multiCluster: false,
  sso: false,
  rtoAnalytics: false,
  regressionAlerts: false,
  timeline: false,
  scoreBreakdown: false,
};

const defaultContext: LicenseContextValue = {
  tier: 'community',
  features: defaultFeatures,
  isTrialing: false,
  trialDaysLeft: undefined,
  isExpired: false,
  isCommunity: true,
  isTeam: false,
  isEnterprise: false,
  loading: true,
  canAccess: () => false,
};

const LicenseContext = createContext<LicenseContextValue>(defaultContext);

const tierOrder: Record<KymarosTier, number> = {
  community: 0,
  team: 1,
  enterprise: 2,
};

export function LicenseProvider({ children }: { children: React.ReactNode }) {
  const license = useLicenseApi();

  const value = useMemo<LicenseContextValue>(() => {
    const data: LicenseResponse | null = license.data;
    const tier: KymarosTier = data?.tier ?? 'community';
    const features = data?.features ?? defaultFeatures;

    return {
      tier,
      features,
      isTrialing: data?.isTrialing ?? false,
      trialDaysLeft: data?.trialDaysLeft,
      isExpired: data?.isExpired ?? false,
      isCommunity: tier === 'community',
      isTeam: tier === 'team',
      isEnterprise: tier === 'enterprise',
      loading: license.loading,
      canAccess: (requiredTier: KymarosTier) =>
        tierOrder[tier] >= tierOrder[requiredTier],
    };
  }, [license.data, license.loading]);

  return (
    <LicenseContext.Provider value={value}>
      {children}
    </LicenseContext.Provider>
  );
}

export function useLicenseContext() {
  return useContext(LicenseContext);
}
