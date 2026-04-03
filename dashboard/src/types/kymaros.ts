// K8s Object metadata
export interface ObjectMeta {
  name: string;
  namespace: string;
  creationTimestamp: string;
  labels?: Record<string, string>;
}

// === RestoreTest CRD ===
export interface RestoreTest {
  metadata: ObjectMeta;
  spec: {
    backupSource: {
      provider: string;
      backupName: string;
      namespaces: Array<{ name: string; sandboxName?: string }>;
    };
    schedule: { cron: string; timezone?: string };
    sandbox: {
      namespacePrefix: string;
      ttl: string;
      networkIsolation: string;
      resourceQuota?: { cpu: string; memory: string; storage?: string };
    };
    healthChecks?: { policyRef: string; timeout: string };
    sla?: { maxRTO: string; alertOnExceed?: boolean };
    notifications?: {
      onFailure?: Array<{ type: string; channel?: string; webhookSecretRef?: string }>;
      onSuccess?: Array<{ type: string; channel?: string; webhookSecretRef?: string }>;
    };
  };
  status?: {
    phase?: string;
    lastRunAt?: string;
    lastScore?: number;
    lastResult?: string;
    lastReportRef?: string;
    nextRunAt?: string;
    sandboxNamespace?: string;
    restoreID?: string;
  };
}

// === RestoreReport CRD ===
export interface ValidationLevelResult {
  status: string; // "pass" | "fail" | "partial" | "not_tested"
  detail?: string;
  tested?: string[];
  notTested?: string[];
}

export interface RestoreReport {
  metadata: ObjectMeta;
  spec: {
    testRef: string;
  };
  status?: {
    score?: number;
    result?: string;
    startedAt?: string;
    completedAt?: string;
    rto?: {
      measured?: string;
      target?: string;
      withinSLA?: boolean;
    };
    backup?: {
      name?: string;
      age?: string;
      size?: string;
    };
    checks?: Array<{
      name: string;
      status: string;
      duration?: string;
      message?: string;
    }>;
    completeness?: {
      deployments?: string;
      services?: string;
      secrets?: string;
      configMaps?: string;
      pvcs?: string;
    };
    validationLevels?: {
      restoreIntegrity?: ValidationLevelResult;
      completeness?: ValidationLevelResult;
      podStartup?: ValidationLevelResult;
      internalHealth?: ValidationLevelResult;
      crossNamespaceDeps?: ValidationLevelResult;
      rtoCompliance?: ValidationLevelResult;
    };
  };
}

// === API Response types ===
export interface TestResponse {
  name: string;
  namespace: string;
  provider: string;
  schedule: string;
  phase: string;
  lastScore: number;
  lastResult: string;
  lastRunAt: string;
  nextRunAt: string;
  sourceNamespaces: string[];
  rtoTarget: string;
}

export interface SummaryResponse {
  averageScore: number;
  totalTests: number;
  testsLastNight: number;
  totalFailures: number;
  totalPartial: number;
  averageRTO: string;
  namespacesCovered: number;
}

export interface DailySummary {
  date: string;
  score: number;
  tests: number;
  failures: number;
}

export interface Alert {
  timestamp: string;
  testName: string;
  namespace: string;
  score: number;
  prevScore: number;
  result: string;
  message: string;
}

export interface ComplianceResponse {
  framework: string;
  period: string;
  status: string;
  testsExecuted: number;
  averageScore: number;
  namespacesCovered: string;
  daysWithTests: number;
  daysInPeriod: number;
  issuesDetected: number;
  averageRTO: string;
  rtoTarget: string;
  rtoCompliant: boolean;
  dailyData: DailySummary[];
}

export interface UpcomingTest {
  name: string;
  namespace: string;
  nextRunAt: string;
  lastScore: number;
}

export interface CreateTestInput {
  name: string;
  provider: string;
  backupName: string;
  namespaces: string[];
  cron: string;
  timezone: string;
  sandboxPrefix: string;
  ttl: string;
  networkIsolation: string;
  quotaCpu: string;
  quotaMemory: string;
  healthCheckRef: string;
  healthCheckTimeout: string;
  maxRTO: string;
  alertOnExceed: boolean;
}

// === Pod Logs ===
export interface ContainerLog {
  name: string;
  type: 'container' | 'init' | 'previous';
  log: string;
  truncated: boolean;
  totalLines?: number;
}

export interface PodLog {
  podName: string;
  namespace: string;
  phase: string;
  containers: ContainerLog[];
}

export interface EventLog {
  type: string;
  reason: string;
  message: string;
  involvedObject: string;
  lastTimestamp: string;
  count?: number;
}

export interface ReportLogsResponse {
  podLogs: PodLog[] | null;
  events: EventLog[] | null;
}

// Legacy alias
export type ValidationResult = ValidationLevelResult;
